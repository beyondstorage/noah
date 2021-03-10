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
	subject  string // All runner will share the same task subject
	storages []types.Storager

	wg     *sync.WaitGroup
	logger *zap.Logger
}

func NewRunner(a *Agent, j *proto.Job) *Runner {
	return &Runner{
		j:        j,
		queue:    a.queue,
		subject:  a.subject, // Copy task subject from agent.
		storages: a.storages,
		wg:       &sync.WaitGroup{},
		logger:   a.logger,
	}
}

func (rn *Runner) Handle(subject, reply string) {
	rn.logger.Debug("start handle job", zap.String("id", rn.j.Id))

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

	// Send JobReply after the job has been handled.
	err = rn.Finish(ctx, reply)
	if err != nil {
		return
	}
}

func (rn *Runner) Async(ctx context.Context, job *proto.Job) (err error) {
	logger := zapcontext.From(ctx)

	logger.Debug("start async job",
		zap.String("parent_id", rn.j.Id),
		zap.String("id", job.Id))

	rn.wg.Add(1)
	// Publish new job with the specific reply subject on the task subject.
	// After this job finished, the runner will send a JobReply to the reply subject.
	//
	// For now, we use the job id as the reply subject.
	// This could be changed.
	err = rn.queue.PublishRequest(rn.subject, rn.j.Id, job)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}

	return
}

func (rn *Runner) Await(ctx context.Context) (err error) {
	logger := zapcontext.From(ctx)

	// Wait for all JobReply sending to the reply subject.
	sub, err := rn.queue.Subscribe(rn.j.Id, rn.awaitHandler)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	rn.wg.Wait()

	logger.Debug("finish await jobs", zap.String("parent_id", rn.j.Id))
	return
}

func (rn *Runner) awaitHandler(job *proto.JobReply) {
	defer rn.wg.Done()

	switch job.Status {
	case JobStatusSucceed:
		rn.logger.Info("job succeed", zap.String("id", job.Id))
	}
	if job.Status != JobStatusSucceed {
		rn.logger.Error("job failed",
			zap.String("id", job.Id),
			zap.String("error", job.Message),
		)
	}
}

func (rn *Runner) Sync(ctx context.Context, job *proto.Job) (err error) {
	logger := zapcontext.From(ctx)

	var reply proto.JobReply

	logger.Debug("sync job",
		zap.String("id", job.Id))

	// NATS provides the builtin request-response style API, so that we don't need to
	// care about the reply id.
	err = rn.queue.RequestWithContext(ctx, rn.subject, job, &reply)
	if err != nil {
		return fmt.Errorf("nats request: %w", err)
	}
	if reply.Status != JobStatusSucceed {
		logger.Error("job failed", zap.String("error", reply.Message))
		return fmt.Errorf("job failed: %v", reply.Message)
	}
	return
}

func (rn *Runner) Finish(ctx context.Context, reply string) (err error) {
	return rn.queue.Publish(reply, &proto.JobReply{
		Id:      rn.j.Id, // Make sure JobReply sends to the parent job.
		Status:  JobStatusSucceed,
		Message: "",
	})
}
