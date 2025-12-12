package task

import "context"

type Task interface {
	Execute(ctx context.Context)
}

type TaskFunc func(ctx context.Context)

func (f TaskFunc) Execute(ctx context.Context) {
	f(ctx)
}
