package task

import (
	"context"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"time"

	"github.com/aos-dev/go-toolbox/natszap"
	"github.com/aos-dev/go-toolbox/zapcontext"
	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/aos-dev/noah/proto"
)

type Worker struct {
	id   string
	node proto.NodeClient
	srv  *server.Server

	queue  *nats.EncodedConn
	sub    *nats.Subscription
	logger *zap.Logger
}

type WorkerConfig struct {
	Host string

	PortalAddr string
}

func NewWorker(ctx context.Context, cfg WorkerConfig) (w *Worker, err error) {
	logger := zapcontext.From(ctx)

	// FIXME: we need to use ssl/tls to encrypt our channel.
	conn, err := grpc.DialContext(ctx, cfg.PortalAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_zap.UnaryClientInterceptor(logger)),
	)
	if err != nil {
		return
	}

	srv, err := server.NewServer(&server.Options{
		Host: cfg.Host,
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return
	}

	go func() {
		srv.SetLoggerV2(natszap.NewLog(logger), false, false, false)

		err = server.Run(srv)
		if err != nil {
			logger.Error("nats server run failed", zap.Error(err))
		}
	}()

	if !srv.ReadyForConnections(time.Second) {
		panic(fmt.Errorf("server start too slow"))
	}

	w = &Worker{
		id:   uuid.New().String(),
		node: proto.NewNodeClient(conn),
		srv:  srv,

		logger: logger,
	}
	return
}

func (w *Worker) Connect(ctx context.Context) (err error) {
	logger := w.logger

	reply, err := w.node.Register(ctx, &proto.RegisterRequest{
		Id:   w.id,
		Addr: w.srv.Addr().String(),
	})
	if err != nil {
		return
	}

	logger.Info("connect to task queue",
		zap.String("addr", reply.Addr),
		zap.String("subject", reply.Subject))

	conn, err := nats.Connect(reply.Addr)
	if err != nil {
		return
	}
	w.queue, err = nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return
	}
	w.sub, err = w.queue.Subscribe(reply.Subject, w.Handle)
	if err != nil {
		return
	}
	return nil
}

func (w *Worker) Handle(subject, reply string, task *proto.Task) {
	w.logger.Info("start handle task",
		zap.String("subject", subject),
		zap.String("id", task.Id),
		zap.String("node_id", w.id))

	a := NewAgent(w, task)
	err := a.Handle()

	tr := &proto.TaskReply{Id: task.Id, NodeId: w.id}
	if err == nil {
		tr.Status = JobStatusSucceed
	} else {
		tr.Status = JobStatusFailed
		tr.Message = fmt.Sprintf("task handle: %v", err)
	}

	err = w.queue.Publish(reply, tr)
	if err != nil {
		w.logger.Error("worker publish reply", zap.Error(err))
	}
}
