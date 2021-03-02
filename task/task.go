package task

import (
	"context"
	"fmt"
	fs "github.com/aos-dev/go-service-fs/v2"
	"github.com/aos-dev/go-storage/v3/types"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"log"
	"time"

	"github.com/aos-dev/noah/proto"
)

const (
	TypeCopyDir uint32 = iota + 1
	TypeCopyFile
)

type Worker struct {
	id   string
	addr string
	node proto.NodeClient

	sub *nats.Subscription
}

func NewWorker(ctx context.Context, addr string) (w *Worker, err error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return
	}

	w = &Worker{
		node: proto.NewNodeClient(conn),
		id:   uuid.New().String(),
		addr: "localhost:7000",
	}
	return
}

func (w *Worker) Connect(ctx context.Context) (err error) {
	reply, err := w.node.Register(ctx, &proto.RegisterRequest{
		Id:   w.id,
		Addr: w.addr,
	})
	if err != nil {
		return
	}

	log.Printf("connect to addr %s subject %s", reply.Addr, reply.Subject)
	queue, err := nats.Connect(reply.Addr)
	if err != nil {
		return
	}
	sub, err := queue.Subscribe(reply.Subject, w.Handle)
	if err != nil {
		return
	}

	w.sub = sub
	return nil
}

func (w *Worker) Handle(msg *nats.Msg) {
	log.Printf("worker got message: %v", msg.Subject)
	task := &proto.Task{}

	err := protobuf.Unmarshal(msg.Data, task)
	if err != nil {
		panic("unmarshal failed")
	}

	a := NewAgent(w, task)
	err = a.Handle()
	if err != nil {
		log.Printf("agent handle: %v", err)
	}
}

type Agent struct {
	w *Worker
	t *proto.Task

	conn     *nats.Conn
	subject  string
	storages []types.Storager
}

func NewAgent(w *Worker, t *proto.Task) *Agent {
	return &Agent{w: w, t: t}
}

func (a *Agent) Handle() (err error) {
	ctx := context.Background()

	reply, err := a.w.node.Upgrade(ctx, &proto.UpgradeRequest{
		NodeId: a.w.id,
		TaskId: a.t.Id,
	})
	if err != nil {
		return fmt.Errorf("node upgrade: %v", err)
	}
	log.Printf("upgrade reply: %s", reply.String())

	a.subject = reply.Subject

	err = a.parseStorage(ctx)
	if err != nil {
		return
	}

	if reply.NodeId == a.w.id {
		return a.handleServer(ctx)
	} else {
		return a.handleClient(ctx, reply.Addr)
	}
}

func (a *Agent) parseStorage(ctx context.Context) (err error) {
	for range a.t.Endpoints {
		a.storages = append(a.storages, &fs.Storage{})
	}
	return
}

func (a *Agent) handleServer(ctx context.Context) (err error) {
	// Setup queue
	srv, err := server.NewServer(&server.Options{
		Host:  "localhost",
		Port:  7000,
		Debug: true, // FIXME: allow used for developing
	})
	if err != nil {
		return
	}

	go func() {
		srv.ConfigureLogger()

		err = server.Run(srv)
		if err != nil {
			log.Printf("server run: %s", err)
		}
	}()

	if !srv.ReadyForConnections(time.Second) {
		panic(fmt.Errorf("server start too slow"))
	}

	conn, err := nats.Connect("localhost:7000")
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	a.conn = conn

	return a.Publish(ctx, a.t.Job)
}

func (a *Agent) handleClient(ctx context.Context, addr string) (err error) {
	var conn *nats.Conn
	// Connect queue
	log.Printf("connect to %s", addr)
	end := time.Now().Add(time.Second)
	for time.Now().Before(end) {
		conn, err = nats.Connect(addr)
		if err != nil {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		break
	}
	a.conn = conn

	// FIXME: we need to handle the returning subscription.
	_, err = a.conn.Subscribe(a.subject, a.handleJob)
	if err != nil {
		return fmt.Errorf("nats subscribe: %w", err)
	}
	return
}

func (a *Agent) handleJob(msg *nats.Msg) {
	job := &proto.Job{}
	err := protobuf.Unmarshal(msg.Data, job)
	if err != nil {
		panic("unmarshal failed")
	}

	log.Printf("got job: %v", job.Id)

	ctx := context.Background()

	var fn func(ctx context.Context, msg protobuf.Message) error
	var t protobuf.Message

	switch job.Type {
	case TypeCopyDir:
		t = &proto.CopyDir{}
		fn = a.HandleCopyDir
	case TypeCopyFile:
		t = &proto.CopyFile{}
		fn = a.HandleCopyFile
	default:
		panic("not support job type")
	}

	err = protobuf.Unmarshal(job.Content, t)
	if err != nil {
		panic("unmarshal failed")
	}

	err = fn(ctx, t)
	if err != nil {
		panic(fmt.Errorf("handle task: %s", err))
	}
}

func (a *Agent) Publish(ctx context.Context, job *proto.Job) (err error) {
	log.Printf("publish job: %s", job.Id)

	content, err := protobuf.Marshal(job)
	if err != nil {
		return err
	}

	err = a.conn.Publish(a.subject, content)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return
}
