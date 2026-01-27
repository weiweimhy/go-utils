package task

import "context"

// Task 定义了可执行任务的接口
type Task interface {
	Execute(ctx context.Context)
}

// TaskFunc 是 Task 接口的函数适配器，方便使用闭包作为任务
type TaskFunc func(ctx context.Context)

// Execute 实现 Task 接口
func (f TaskFunc) Execute(ctx context.Context) {
	f(ctx)
}

// TaskWithError 定义了可返回错误的任务接口
type TaskWithError interface {
	Execute(ctx context.Context) error
}

// TaskFuncWithError 是 TaskWithError 接口的函数适配器
type TaskFuncWithError func(ctx context.Context) error

// Execute 实现 TaskWithError 接口
func (f TaskFuncWithError) Execute(ctx context.Context) error {
	return f(ctx)
}
