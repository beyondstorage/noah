package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/aos-dev/go-storage/v3/types"
	"github.com/aos-dev/go-toolbox/zapcontext"
	protobuf "github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/aos-dev/noah/proto"
)

type Runner struct {
	j *proto.Job

	queue    *nats.EncodedConn
	subject  string
	storages []types.Storager

	wg  *sync.WaitGroup
	log *zap.Logger
}

func NewRunner(a *Agent, j *proto.Job) *Runner {
	return &Runner{
		j:        j,
		queue:    a.queue,
		subject:  a.subject,
		storages: a.storages,
		wg:       &sync.WaitGroup{},
		log:      a.log,
	}
}

func (rn *Runner) Handle(subject, reply string) {
	rn.log.Debug("start handle job", zap.String("id", rn.j.Id))

	ctx := context.Background()

	var fn func(ctx context.Context, msg protobuf.Message) error
	var t protobuf.Message

	switch rn.j.Type {
	case TypeCopyDir:
		t = &proto.CopyDir{}
		fn = rn.HandleCopyDir
	case TypeCopyFile:
		t = &proto.CopyFile{}
		fn = rn.HandleCopyFile
	case TypeCopySingleFile:
		t = &proto.CopySingleFile{}
		fn = rn.HandleCopySingleFile
	case TypeCopyMultipartFile:
		t = &proto.CopyMultipartFile{}
		fn = rn.HandleCopyMultipartFile
	case TypeCopyMultipart:
		t = &proto.CopyMultipart{}
		fn = rn.HandleCopyMultipart
	default:
		panic("not support job type")
	}

	err := protobuf.Unmarshal(rn.j.Content, t)
	if err != nil {
		panic("unmarshal failed")
	}

	err = fn(ctx, t)
	if err != nil {
		// TODO: reply to message sender the response.
		panic(fmt.Errorf("handle task: %s", err))
	}

	err = rn.Finish(ctx, reply)
	if err != nil {
		return
	}
}

func (rn *Runner) Async(ctx context.Context, job *proto.Job) (err error) {
	log := zapcontext.From(ctx)

	log.Debug("start async job",
		zap.String("parent_id", rn.j.Id),
		zap.String("id", job.Id))

	rn.wg.Add(1)
	err = rn.queue.PublishRequest(rn.subject, rn.j.Id, job)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}

	return
}

func (rn *Runner) Await(ctx context.Context) (err error) {
	log := zapcontext.From(ctx)

	log.Debug("start await jobs", zap.String("parent_id", rn.j.Id))
	sub, err := rn.queue.Subscribe(rn.j.Id, rn.awaitHandler)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	rn.wg.Wait()

	log.Debug("finish await jobs", zap.String("parent_id", rn.j.Id))
	return
}

func (rn *Runner) awaitHandler(job *proto.JobReply) {
	defer rn.wg.Done()

	switch job.Status {
	case JobStatusSucceed:
		rn.log.Info("job succeed", zap.String("id", job.Id))
	}
	if job.Status != JobStatusSucceed {
		rn.log.Error("job failed",
			zap.String("id", job.Id),
			zap.String("error", job.Message),
		)
	}
}

func (rn *Runner) Sync(ctx context.Context, job *proto.Job) (err error) {
	log := zapcontext.From(ctx)

	var reply proto.JobReply

	log.Debug("sync job",
		zap.String("id", job.Id))

	err = rn.queue.RequestWithContext(ctx, rn.subject, job, &reply)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}
	if reply.Status != JobStatusSucceed {
		log.Error("job failed", zap.String("error", reply.Message))
		return fmt.Errorf("job failed: %v", reply.Message)
	}
	return
}

func (rn *Runner) Finish(ctx context.Context, reply string) (err error) {
	return rn.queue.Publish(reply, &proto.JobReply{
		Id:      rn.j.Id,
		Status:  JobStatusSucceed,
		Message: "",
	})
}