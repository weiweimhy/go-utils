package task

import (
	"context"
	"time"
)

func ExampleNewWorkerPool() {
	pool := NewWorkerPool(
		context.Background(),
		WithWorkerCount(2),
		WithBufferSize(8),
		WithName("example"),
	)
	defer pool.Close(time.Second)

	group := pool.NewGroup()
	group.SubmitFunc(func(ctx context.Context) {})
	group.Wait()
}

func ExampleNewWorkerPool_defaults() {
	pool := NewWorkerPool(context.Background())
	defer pool.Close(time.Second)

	group := pool.NewGroup()
	group.SubmitFunc(func(ctx context.Context) {})
	group.Wait()
}
