package task

import (
	"context"
	"fmt"
	"time"

	fs "github.com/aos-dev/go-service-fs/v2"
	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/go-toolbox/zapcontext"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/aos-dev/noah/proto"
)

type Worker struct {
	id   string
	addr string
	node proto.NodeClient

	sub *nats.Subscription
	log *zap.Logger
}

func NewWorker(ctx context.Context, addr string) (w *Worker, err error) {
	logger := zapcontext.From(ctx)

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return
	}

	w = &Worker{
		node: proto.NewNodeClient(conn),
		id:   uuid.New().String(),
		addr: "localhost:7000",
		log:  logger,
	}
	return
}

func (w *Worker) Connect(ctx context.Context) (err error) {
	log := zapcontext.From(ctx)

	reply, err := w.node.Register(ctx, &proto.RegisterRequest{
		Id:   w.id,
		Addr: w.addr,
	})
	if err != nil {
		return
	}

	log.Info("connect to nats server",
		zap.String("addr", reply.Addr),
		zap.String("subject", reply.Subject))

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
	w.log.Debug("worker got message",
		zap.String("subject", msg.Subject))

	task := &proto.Task{}

	err := protobuf.Unmarshal(msg.Data, task)
	if err != nil {
		panic("unmarshal failed")
	}

	a := NewAgent(w, task)
	err = a.Handle()
	if err != nil {
		w.log.Error("agent handle", zap.Error(err))
	}
}

type Agent struct {
	w *Worker
	t *proto.Task

	conn     *nats.Conn
	subject  string
	storages []types.Storager

	log *zap.Logger
}

func NewAgent(w *Worker, t *proto.Task) *Agent {
	return &Agent{w: w, t: t, log: w.log}
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
	a.log.Info("receive upgrade", zap.String("reply", reply.String()))

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
			a.log.Error("server run", zap.Error(err))
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
	log := zapcontext.From(ctx)

	var conn *nats.Conn
	// Connect queue
	log.Info("agent connect to nats server", zap.String("addr", addr))
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

	a.log.Debug("got job", zap.String("id", job.Id))

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
	case TypeCopySingleFile:
		t = &proto.CopySingleFile{}
		fn = a.HandleCopySingleFile
	case TypeCopyMultipartFile:
		t = &proto.CopyMultipartFile{}
		fn = a.HandleCopyMultipartFile
	case TypeCopyMultipart:
		t = &proto.CopyMultipart{}
		fn = a.HandleCopyMultipart
	default:
		panic("not support job type")
	}

	err = protobuf.Unmarshal(job.Content, t)
	if err != nil {
		panic("unmarshal failed")
	}

	err = fn(ctx, t)
	if err != nil {
		// TODO: reply to message sender the response.
		panic(fmt.Errorf("handle task: %s", err))
	}
}

func (a *Agent) Publish(ctx context.Context, job *proto.Job) (err error) {
	log := zapcontext.From(ctx)

	log.Debug("publish job", zap.String("id", job.Id))

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
