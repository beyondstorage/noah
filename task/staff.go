package task

import (
	"context"
	"fmt"
	"github.com/aos-dev/go-storage/v3/types"
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

type Staff struct {
	id   string
	addr string
	cfg  StaffConfig

	logger *zap.Logger
	ctx    context.Context

	grpcClient proto.StaffClient
	queueSrv   *server.Server

	queue *nats.EncodedConn
	sub   *nats.Subscription
}

type StaffConfig struct {
	Host string

	ManagerAddr string
}

func NewStaff(ctx context.Context, cfg StaffConfig) (s *Staff, err error) {
	logger := zapcontext.From(ctx)

	s = &Staff{
		id:  uuid.New().String(),
		cfg: cfg,

		ctx:    ctx,
		logger: logger,
	}

	// Setup NATS server.
	s.queueSrv, err = server.NewServer(&server.Options{
		Host: cfg.Host,
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return
	}

	go func() {
		s.queueSrv.SetLoggerV2(natszap.NewLog(logger), false, false, false)

		s.queueSrv.Start()
	}()

	if !s.queueSrv.ReadyForConnections(10 * time.Second) {
		panic(fmt.Errorf("server start too slow"))
	}
	s.addr = s.queueSrv.ClientURL()
	return
}

// Connect will connect to portal task queue.
func (w *Staff) Connect(ctx context.Context) (err error) {
	logger := w.logger

	// FIXME: we need to use ssl/tls to encrypt our channel.
	grpcConn, err := grpc.DialContext(ctx, w.cfg.ManagerAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_zap.UnaryClientInterceptor(logger)),
	)
	if err != nil {
		return
	}
	w.grpcClient = proto.NewStaffClient(grpcConn)

	reply, err := w.grpcClient.Register(ctx, &proto.RegisterRequest{
		Id:   w.id,
		Addr: w.addr,
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
	w.sub, err = w.queue.Subscribe(reply.Subject,
		func(subject, reply string, task *proto.Task) {
			w.logger.Info("start handle task",
				zap.String("subject", subject),
				zap.String("id", task.Id),
				zap.String("node_id", w.id))

			go w.Handle(reply, task)
		})
	if err != nil {
		return
	}
	return nil
}

// Handle will create a new agent to handle task.
func (w *Staff) Handle(reply string, task *proto.Task) {
	// Parse storage
	storages := make([]types.Storager, 0)
	for _, ep := range task.Endpoints {
		store, err := ep.ParseStorager()
		if err != nil {
			w.logger.Error("xxx")
			return
		}
		storages = append(storages, store)
	}

	// Send upgrade
	electReply, err := w.grpcClient.Elect(w.ctx, &proto.ElectRequest{
		StaffId: w.id,
		TaskId:  task.Id,
	})
	if err != nil {
		w.logger.Error("xxx")
		return
	}

	if electReply.LeaderId == w.id {
		err = HandleAsLeader(w.ctx, electReply.Addr, electReply.Subject, electReply.WorkerIds, task.Job)
	} else {
		err = HandleAsWorker(w.ctx, electReply.Addr, electReply.Subject, storages)
	}

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
