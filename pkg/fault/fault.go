package fault

import (
	"strings"
)

// Fault will handle multi error in tasks.
type Fault struct {
	errs []error
}

// New will create a new Fault.
func New() *Fault {
	return &Fault{}
}

// Append will append errors in fault.
func (f *Fault) Append(err ...error) {
	f.errs = append(f.errs, err...)
}

// HasError checks whether this fault has error or not.
func (f *Fault) HasError() bool {
	return len(f.errs) != 0
}

// Error will print all errors in fault.
func (f *Fault) Error() string {
	x := make([]string, 0)
	for _, v := range f.errs {
		x = append(x, v.Error())
	}
	return strings.Join(x, "\n")
}

// Unwrap implements unwarp interface.
func (f *Fault) Unwrap() error {
	if len(f.errs) == 0 {
		return nil
	}
	return f.errs[0]
}
