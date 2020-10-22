package fault

import (
	"strings"
	"sync"
)

type errList []error

func (l errList) Error() string {
	x := make([]string, 0, len(l))
	for _, v := range l {
		x = append(x, v.Error())
	}
	return strings.Join(x, "\n")
}

// Fault will handle multi error in tasks.
type Fault struct {
	errs errList
	lock sync.RWMutex
}

// New will create a new Fault.
func New() *Fault {
	return &Fault{}
}

// Append will append errors in fault.
func (f *Fault) Append(err ...error) {
	f.lock.Lock()
	f.errs = append(f.errs, err...)
	f.lock.Unlock()
}

// HasError checks whether this fault has error or not.
func (f *Fault) HasError() bool {
	f.lock.RLock()
	defer f.lock.RUnlock()

	return len(f.errs) != 0
}

// Error will print all errors in fault.
func (f *Fault) Error() string {
	f.lock.RLock()
	defer f.lock.RUnlock()

	return f.errs.Error()
}

// Unwrap implements unwrap interface.
func (f *Fault) Unwrap() error {
	f.lock.RLock()
	defer f.lock.RUnlock()

	if len(f.errs) == 0 {
		return nil
	}
	return f.errs[0]
}

// Pop get all errors and clear the error list
func (f *Fault) Pop() error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if len(f.errs) == 0 {
		return nil
	}
	res := make(errList, len(f.errs))
	copy(res, f.errs)
	f.errs = f.errs[:0]
	return res
}
