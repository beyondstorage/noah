package schedule

import (
	"context"
	"sync"

	"github.com/Xuanwo/navvy"
)

// TaskFunc will be used create a new task.
type TaskFunc func(navvy.Task) navvy.Task

// Scheduler will schedule tasks.
//go:generate mockgen -package mock -destination ../mock/scheduler.go github.com/qingstor/noah/pkg/schedule Scheduler
type Scheduler interface {
	Sync(ctx context.Context, task navvy.Task)
	Async(ctx context.Context, task navvy.Task)

	Wait()
}

// VoidWorkloader will not be included in Worker Pool.
type VoidWorkloader interface {
	VoidWorkload()
}

// IOWorkloader will be included in Worker Pool.
type IOWorkloader interface {
	IOWorkload()
}

type task struct {
	ctx context.Context

	s *RealScheduler
	t navvy.Task
}

func newTask(ctx context.Context, s *RealScheduler, t navvy.Task) *task {
	return &task{
		ctx: ctx,
		s:   s,
		t:   t,
	}
}

func (t *task) Context() context.Context {
	if t.ctx == nil {
		return context.Background()
	}
	return t.ctx
}

// Run run
func (t *task) Run(ctx context.Context) {
	defer func() {
		t.s.wg.Done()
	}()
	if ctx == nil {
		ctx = t.t.Context()
	}
	t.t.Run(ctx)
}

// RealScheduler will hold the task's sub tasks.
type RealScheduler struct {
	wg   *sync.WaitGroup
	pool *navvy.Pool
}

// NewScheduler will create a new RealScheduler.
func NewScheduler(pool *navvy.Pool) *RealScheduler {
	return &RealScheduler{
		wg:   &sync.WaitGroup{},
		pool: pool,
	}
}

// Sync will return after this task finished.
func (s *RealScheduler) Sync(ctx context.Context, task navvy.Task) {
	s.wg.Add(1)

	defer func() {
		s.wg.Done()
	}()
	task.Run(ctx)
}

// Async will create a new task immediately.
func (s *RealScheduler) Async(ctx context.Context, task navvy.Task) {
	s.wg.Add(1)

	t := newTask(ctx, s, task)
	s.pool.Submit(t)
}

// Wait will wait until a task finished.
func (s *RealScheduler) Wait() {
	s.wg.Wait()
}
