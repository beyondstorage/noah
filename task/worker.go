package task

import (
	"context"
	"errors"
	"fmt"
	"github.com/aos-dev/go-toolbox/zapcontext"
	"github.com/google/uuid"
	"sync"
	"time"

	"github.com/aos-dev/go-storage/v3/types"
	"github.com/nats-io/nats.go"
	natsproto "github.com/nats-io/nats.go/encoders/protobuf"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

type Worker struct {
	id      string
	subject string

	queue    *nats.EncodedConn
	storages []types.Storager

	ctx    context.Context
	cond   *sync.Cond
	logger *zap.Logger
}

func NewWorker(ctx context.Context, addr, subject string, storages []types.Storager) (*Worker, error) {
	logger := zapcontext.From(ctx)

	a := &Worker{
		id:       uuid.NewString(),
		subject:  subject,
		storages: storages,

		ctx:    ctx,
		cond:   sync.NewCond(&sync.Mutex{}),
		logger: logger,
	}
	a.cond.L.Lock()

	// Connect to queue
	queueConn, err := nats.Connect(addr)
	if err != nil {
		return nil, err
	}
	a.queue, err = nats.NewEncodedConn(queueConn, natsproto.PROTOBUF_ENCODER)
	if err != nil {
		return nil, fmt.Errorf("nats encoded connect: %w", err)
	}

	logger.Info("worker has been setup", zap.String("id", a.id))
	return a, nil
}

func (a *Worker) clockin() {
	a.logger.Info("worker start clockin", zap.String("id", a.id))

	reply := &proto.ClockinReply{}

	for {
		err := a.queue.RequestWithContext(a.ctx, SubjectClockin(a.subject),
			&proto.ClockinRequest{}, reply)
		if err != nil && errors.Is(err, nats.ErrNoResponders) {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		if err != nil {
			a.logger.Error("worker clockin", zap.String("id", a.id), zap.Error(err))
			return
		}
		break
	}
}

func (a *Worker) clockout() {
	a.logger.Info("worker start waiting for clockout", zap.String("id", a.id))

	_, err := a.queue.Subscribe(SubjectClockoutNotify(a.subject),
		func(subject, reply string, req *proto.ClockoutRequest) {
			err := a.queue.Publish(reply, &proto.Acknowledgement{})
			if err != nil {
				a.logger.Error("publish ack", zap.Error(err))
				return
			}

			a.cond.Signal()
		})
	if err != nil {
		return
	}
}

func (a *Worker) Handle(ctx context.Context) (err error) {
	go a.clockout()

	// Worker must setup before clockin.
	_, err = a.queue.QueueSubscribe(a.subject, a.subject,
		func(subject, reply string, job *proto.Job) {
			go func() {
				rn, err := NewRunner(a, job)
				if err != nil {
					a.logger.Error("create new runner", zap.Error(err))
					return
				}
				rn.Handle(subject, reply)
			}()
		})
	if err != nil {
		return fmt.Errorf("nats subscribe: %w", err)
	}

	// TODO: we can clockout directly if the task has been finished.
	a.clockin()

	a.cond.Wait()
	return
}

func HandleAsWorker(ctx context.Context, addr, subject string, storages []types.Storager) (err error) {
	w, err := NewWorker(ctx, addr, subject, storages)
	if err != nil {
		return
	}

	return w.Handle(ctx)
}
