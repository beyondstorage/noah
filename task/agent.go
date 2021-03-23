package task

import (
	"context"
	"fmt"
	"sync"

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

	wg     *sync.WaitGroup // Control client runners via wait group
	logger *zap.Logger
}

func NewAgent(w *Worker, t *proto.Task) *Agent {
	return &Agent{
		w: w,
		t: t,

		wg:     &sync.WaitGroup{},
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

	a.subject = reply.Subject

	err = a.parseStorage(ctx)
	if err != nil {
		return
	}
	err = a.connect(ctx, reply.Addr)
	if err != nil {
		return
	}

	if reply.NodeId == a.w.id {
		err = a.handleServer(ctx)
	} else {
		err = a.handleClient(ctx)
	}
	if err != nil {
		return
	}

	return nil
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

func (a *Agent) connect(ctx context.Context, addr string) error {
	conn, err := nats.Connect(addr)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	a.queue, err = nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return fmt.Errorf("nats encoded connect: %w", err)
	}
	return nil
}

func (a *Agent) handleServer(ctx context.Context) (err error) {
	logger := a.logger

	logger.Info("agent handle as server", zap.String("subject", a.subject), zap.String("node_id", a.w.id))

	// FIXME: we need to maintain task running status instead of job's
	rn := NewRunner(a, a.t.Job)
	err = rn.Sync(ctx, a.t.Job)
	if err != nil {
		return err
	}

	return a.queue.Drain()
}

func (a *Agent) handleClient(ctx context.Context) (err error) {
	logger := a.logger

	logger.Info("agent handle as client", zap.String("subject", a.subject), zap.String("node_id", a.w.id))

	// FIXME: we need to handle the returning subscription.
	_, err = a.queue.QueueSubscribe(a.subject, a.subject,
		func(subject, reply string, job *proto.Job) {
			a.wg.Add(1)
			go NewRunner(a, job).Handle(subject, reply)
		})
	if err != nil {
		return fmt.Errorf("nats subscribe: %w", err)
	}

	a.wg.Wait()
	return
}
