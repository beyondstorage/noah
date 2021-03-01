package task

import (
	"context"
	"github.com/aos-dev/go-storage/v3/types"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"

	"github.com/aos-dev/noah/proto"
)

const (
	TypeCopyDir uint32 = iota + 1
	TypeCopyFile
)

type Worker struct {
	id   string
	node proto.NodeClient

	sub *nats.Subscription
}

func NewWorker(ctx context.Context, addr string) (w *Worker, err error) {
	conn, err := grpc.DialContext(ctx, addr)
	if err != nil {
		return
	}

	w = &Worker{node: proto.NewNodeClient(conn), id: uuid.New().String()}
	return
}

func (w *Worker) Connect(ctx context.Context) (err error) {
	reply, err := w.node.Register(ctx, &proto.RegisterRequest{Id: w.id})
	if err != nil {
		return
	}

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
	var task *proto.Task

	err := protobuf.Unmarshal(msg.Data, task)
	if err != nil {
		panic("unmarshal failed")
	}

	a := NewAgent(w, task)
	err = a.Handle()
	if err != nil {
		panic("agent handle failed")
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
		return
	}

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
	// TODO: Implement storage parse
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
		srv.Start()
	}()

	conn, err := nats.Connect("localhost:7000")
	if err != nil {
		return
	}
	a.conn = conn

	return a.Publish(ctx, a.t.Job)
}

func (a *Agent) handleClient(ctx context.Context, addr string) (err error) {
	// Connect queue
	conn, err := nats.Connect(addr)
	if err != nil {
		return
	}
	a.conn = conn

	// FIXME: we need to handle the returning subscription.
	_, err = conn.Subscribe(a.subject, a.handleJob)
	if err != nil {
		return
	}
	return
}

func (a *Agent) handleJob(msg *nats.Msg) {
	var job *proto.Job

	err := protobuf.Unmarshal(msg.Data, job)
	if err != nil {
		panic("unmarshal failed")
	}

	ctx := context.Background()

	switch job.Type {
	case TypeCopyDir:
		var t *proto.CopyDir
		err := protobuf.Unmarshal(job.Content, t)
		if err != nil {
			panic("unmarshal failed")
		}
		err = a.HandleCopyDir(ctx, t)
		if err != nil {
			panic("handle copy dir failed")
		}
	}
}

func (a *Agent) Publish(ctx context.Context, job *proto.Job) (err error) {
	content, err := protobuf.Marshal(job)
	if err != nil {
		return err
	}

	err = a.conn.Publish(a.subject, content)
	if err != nil {
		return
	}
	return
}
