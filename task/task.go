package task

import "context"

// Task represents a unit of work executed with a context.
type Task interface {
	Execute(ctx context.Context)
}

// TaskFunc adapts a function so it can be used as a Task.
type TaskFunc func(ctx context.Context)

// Execute runs f.
func (f TaskFunc) Execute(ctx context.Context) {
	f(ctx)
}

// TaskWithError represents a unit of work that may return an error.
type TaskWithError interface {
	Execute(ctx context.Context) error
}

// TaskFuncWithError adapts a function so it can be used as a TaskWithError.
type TaskFuncWithError func(ctx context.Context) error

// Execute runs f.
func (f TaskFuncWithError) Execute(ctx context.Context) error {
	return f(ctx)
}
