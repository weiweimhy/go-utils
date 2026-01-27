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
