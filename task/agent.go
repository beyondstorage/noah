package task

import (
	"context"
	"fmt"

	"github.com/aos-dev/go-storage/v3/types"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

type Agent struct {
	w *Worker
	t *proto.Task

	queue    *nats.EncodedConn
	subject  string // All agent will share the same task subject
	storages []types.Storager

	logger *zap.Logger
}

func NewAgent(w *Worker, t *proto.Task) *Agent {
	return &Agent{
		w: w,
		t: t,

		logger: w.logger,
	}
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
	a.logger.Info("receive upgrade", zap.String("reply", reply.String()))

	a.subject = reply.Subject

	err = a.parseStorage(ctx)
	if err != nil {
		return
	}

	if reply.NodeId == a.w.id {
		return a.handleServer(ctx, reply.Addr)
	} else {
		return a.handleClient(ctx, reply.Addr)
	}
}

func (a *Agent) parseStorage(ctx context.Context) (err error) {
	for _, ep := range a.t.Endpoints {
		store, err := ep.ParseStorager()
		if err != nil {
			return err
		}
		a.storages = append(a.storages, store)
	}
	return
}

func (a *Agent) handleServer(ctx context.Context, addr string) (err error) {
	logger := a.logger

	logger.Info("agent connect to job queue as server", zap.String("addr", addr))

	conn, err := nats.Connect(addr)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	queue, err := nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return fmt.Errorf("nats encoded connect: %w", err)
	}
	a.queue = queue

	// FIXME: we need to maintain task running status instead of job's
	rn := NewRunner(a, a.t.Job)

	return rn.Async(ctx, a.t.Job)
}

func (a *Agent) handleClient(ctx context.Context, addr string) (err error) {
	logger := a.logger

	logger.Info("agent connect to job queue as client",
		zap.String("addr", addr), zap.String("subject", a.subject))

	conn, err := nats.Connect(addr)
	if err != nil {
		return
	}
	queue, err := nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return
	}
	a.queue = queue

	// FIXME: we need to handle the returning subscription.
	_, err = a.queue.QueueSubscribe(a.subject, a.subject, a.handleJob)
	if err != nil {
		return fmt.Errorf("nats subscribe: %w", err)
	}
	return
}

func (a *Agent) handleJob(subject, reply string, job *proto.Job) {
	go NewRunner(a, job).Handle(subject, reply)
}
