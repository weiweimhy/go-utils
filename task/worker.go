package task

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weiweimhy/go-utils/v3/logger"
	"go.uber.org/zap"
)

// WorkerPool 是一个通用的工作池实现
type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *zap.Logger

	tasks     chan Task
	wait      sync.WaitGroup
	closeOnce sync.Once
	closed    atomic.Bool
}

// TaskGroup 用于分组等待一批任务完成
type TaskGroup struct {
	pool *WorkerPool
	wg   sync.WaitGroup
}

// NewWorkerPool 创建一个新的工作池
// ctx: 父 context，用于控制工作池生命周期
// workNumber: 工作协程数量，<=0 则使用 CPU 核心数
// buffer: 任务队列缓冲区大小，<0 则使用 0
func NewWorkerPool(ctx context.Context, workNumber int, buffer int) *WorkerPool {
	if ctx == nil {
		ctx = context.Background()
	}
	if workNumber <= 0 {
		workNumber = runtime.NumCPU()
	}
	if buffer < 0 {
		buffer = 0
	}

	ctx, cancel := context.WithCancel(ctx)
	log := logger.FromContext(ctx, zap.String("module", "WorkerPool"))

	workerPool := &WorkerPool{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		tasks:  make(chan Task, buffer),
	}

	for i := 0; i < workNumber; i++ {
		workerPool.wait.Add(1)
		go func(index int) {
			defer workerPool.wait.Done()
			workerPool.workerLoop(index)
		}(i)
	}

	return workerPool
}

func (w *WorkerPool) workerLoop(index int) {
	log := w.log.With(
		zap.Int("worker_index", index),
		zap.Uint64("goroutine_id", logger.GetGoroutineID()),
	)

	for {
		select {
		case <-w.ctx.Done():
			log.Debug("worker exit: context cancelled")
			return
		case task, ok := <-w.tasks:
			if !ok {
				log.Debug("worker exit: channel closed")
				return
			}
			w.safeExecute(log, task)
		}
	}
}

// safeExecute 安全执行任务，捕获 panic
func (w *WorkerPool) safeExecute(log *zap.Logger, task Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("task panic recovered",
				zap.Any("panic", r),
				zap.Stack("stack"),
			)
		}
	}()
	task.Execute(w.ctx)
}

// Submit 提交一个任务到工作池
// 返回 true 表示任务已提交，false 表示工作池已关闭
func (w *WorkerPool) Submit(task Task) bool {
	if w.closed.Load() {
		return false
	}

	select {
	case w.tasks <- task:
		return true
	case <-w.ctx.Done():
		return false
	}
}

// SubmitFunc 提交一个函数作为任务
func (w *WorkerPool) SubmitFunc(fn func(ctx context.Context)) bool {
	return w.Submit(TaskFunc(fn))
}

// NewGroup 创建一个新的任务组，用于批量提交并等待任务完成
func (w *WorkerPool) NewGroup() *TaskGroup {
	return &TaskGroup{pool: w}
}

// Context 返回工作池的 context
func (w *WorkerPool) Context() context.Context {
	return w.ctx
}

// IsClosed 检查工作池是否已关闭
func (w *WorkerPool) IsClosed() bool {
	return w.closed.Load()
}

// Submit 向任务组提交任务
func (g *TaskGroup) Submit(task Task) bool {
	if g.pool.closed.Load() {
		return false
	}

	select {
	case <-g.pool.ctx.Done():
		return false
	default:
	}

	g.wg.Add(1)
	wrapped := TaskFunc(func(ctx context.Context) {
		defer g.wg.Done()
		task.Execute(ctx)
	})

	select {
	case g.pool.tasks <- wrapped:
		return true
	case <-g.pool.ctx.Done():
		g.wg.Done()
		return false
	}
}

// SubmitFunc 向任务组提交函数作为任务
func (g *TaskGroup) SubmitFunc(fn func(ctx context.Context)) bool {
	return g.Submit(TaskFunc(fn))
}

// Wait 等待任务组中的所有任务完成
func (g *TaskGroup) Wait() {
	g.wg.Wait()
}

// Close 优雅关闭工作池
// timeout: 等待所有任务完成的超时时间
// 返回 true 表示正常关闭，false 表示超时
func (w *WorkerPool) Close(timeout time.Duration) bool {
	graceful := true

	w.closeOnce.Do(func() {
		w.closed.Store(true)
		close(w.tasks)
		w.cancel()

		done := make(chan struct{})
		go func() {
			w.wait.Wait()
			close(done)
		}()

		select {
		case <-done:
			w.log.Info("worker pool closed gracefully")
		case <-time.After(timeout):
			w.log.Warn("worker pool closed with timeout")
			graceful = false
		}
	})

	return graceful
}
