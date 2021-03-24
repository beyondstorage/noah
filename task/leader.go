package task

import (
	"context"
	"github.com/aos-dev/go-toolbox/zapcontext"
	"github.com/aos-dev/noah/proto"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
	"go.uber.org/zap"
	"sync"
)

type Leader struct {
	id        string
	addr      string
	subject   string
	workerIds []string

	queue *nats.EncodedConn

	logger      *zap.Logger
	workerClock *sync.WaitGroup
}

func NewLeader(ctx context.Context, addr, subject string, workerIds []string) (l *Leader, err error) {
	logger := zapcontext.From(ctx)

	l = &Leader{
		id:        uuid.NewString(),
		addr:      addr,
		subject:   subject,
		workerIds: workerIds,

		logger:      logger,
		workerClock: &sync.WaitGroup{},
	}
	l.workerClock.Add(len(l.workerIds))

	conn, err := nats.Connect(addr)
	if err != nil {
		return
	}
	l.queue, err = nats.NewEncodedConn(conn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return
	}

	logger.Info("leader has been setup", zap.String("id", l.id))
	return
}

func (l *Leader) clockinWorkers() {
	// TODO: we need to unsubscribe after we finished this task.
	_, err := l.queue.Subscribe(SubjectClockin(l.subject),
		func(subject, reply string, arg *proto.ClockinRequest) {
			defer l.workerClock.Done()

			err := l.queue.Publish(reply, &proto.ClockinReply{})
			if err != nil {
				l.logger.Error("publish clockin reply", zap.Error(err))
				return
			}
		})
	if err != nil {
		l.logger.Error("subscribe clockin",
			zap.String("subject", SubjectClockin(l.subject)),
			zap.Error(err),
		)
	}
	return
}

func (l *Leader) clockoutWorkers() {
	_, err := l.queue.Subscribe(SubjectClockout(l.subject),
		func(subject, reply string, arg *proto.Acknowledgement) {
			l.workerClock.Done()
		})
	if err != nil {
		l.logger.Error("subscribe clockout",
			zap.String("subject", SubjectClockin(l.subject)),
			zap.Error(err),
		)
	}

	err = l.queue.PublishRequest(SubjectClockoutNotify(l.subject), SubjectClockout(l.subject), &proto.ClockoutRequest{})
	if err != nil {
		l.logger.Error("publish chock out notify", zap.Error(err))
		return
	}

	l.workerClock.Wait()
	return
}

func (l *Leader) Handle(ctx context.Context, job *proto.Job) (err error) {
	go l.clockinWorkers()
	// All staff are clocked in, we can start work now.
	l.workerClock.Wait()

	reply := &proto.JobReply{}
	err = l.queue.RequestWithContext(ctx, l.subject, job, reply)
	if err != nil {
		return
	}

	// Reset to prepare for staff chock out.
	l.workerClock.Add(len(l.workerIds))
	l.clockoutWorkers()
	return
}

func HandleAsLeader(ctx context.Context, addr, subject string, workerIds []string, job *proto.Job) (err error) {
	l, err := NewLeader(ctx, addr, subject, workerIds)
	if err != nil {
		return
	}

	return l.Handle(ctx, job)
}
