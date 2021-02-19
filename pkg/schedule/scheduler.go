package schedule

import (
	"sync"

	"github.com/aos-dev/noah/pkg/fault"
)

//go:generate mockgen -package mock -destination ../mock/scheduler.go github.com/aos-dev/noah/pkg/schedule Scheduler

// Scheduler will schedule tasks
type Scheduler interface {
	Add(n int)
	Done()
	Await() error

	AppendFault(err error)
}

// RealScheduler is the struct for task schedule
type RealScheduler struct {
	wg *sync.WaitGroup
	f  *fault.Fault
}

// Add wrap the waitGroup.Add
func (r *RealScheduler) Add(n int) {
	r.wg.Add(n)
}

// Done wrap the waitGroup.Done
func (r *RealScheduler) Done() {
	r.wg.Done()
}

// Await wait the waitGroup and then pop all errors
func (r *RealScheduler) Await() error {
	r.wg.Wait()
	return r.f.Pop()
}

// AppendFault append error into embedded fault
func (r *RealScheduler) AppendFault(err error) {
	r.f.Append(err)
}

// New a real scheduler
func New() *RealScheduler {
	return &RealScheduler{
		wg: &sync.WaitGroup{},
		f:  fault.New(),
	}
}
