package task

import "context"

// Task is the interface to Run with given context and return an error
type Task interface {
	Run(ctx context.Context) error
}
