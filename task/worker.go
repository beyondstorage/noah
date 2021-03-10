package task

import (
	"context"

	"github.com/aos-dev/go-toolbox/zapcontext"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
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

	conn, err := nats.Connect(reply.Addr)
	if err != nil {
		return
	}
	queue, err := nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
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

func (w *Worker) Handle(subject, reply string, task *proto.Task) {
	w.log.Debug("start handle task",
		zap.String("subject", subject),
		zap.String("id", task.Id))

	a := NewAgent(w, task)
	err := a.Handle()
	if err != nil {
		w.log.Error("agent handle", zap.Error(err))
	}
}
