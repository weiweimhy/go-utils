package task

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/weiweimhy/go-utils/logger"
	"go.uber.org/zap"
)

// WorkerPool 是一个通用的工作池实现
type WorkerPool struct {
	logger.CtxLogger

	tasks  chan Task
	cancel context.CancelFunc
	wait   sync.WaitGroup
}

// TaskGroup 用于分组等待一批任务完成
type TaskGroup struct {
	pool *WorkerPool
	wg   sync.WaitGroup
}

// NewWorkerPool 创建一个新的工作池
// workNumber: 工作协程数量
// buffer: 任务队列缓冲区大小
func NewWorkerPool(workNumber int, buffer int) *WorkerPool {
	tasks := make(chan Task, buffer)

	pctx := context.Background()
	ctx, cancel := context.WithCancel(pctx)
	cl := logger.NewCtxLogger(ctx, zap.String("module", "WorkerPool"))

	workerPool := &WorkerPool{
		CtxLogger: cl,
		tasks:     tasks,
		cancel:    cancel,
		wait:      sync.WaitGroup{},
	}

	for i := 0; i < workNumber; i++ {
		workerPool.wait.Add(1)
		go func(index int) {
			defer workerPool.wait.Done()

			gcl := cl.With(
				zap.String("goroutine_index", strconv.Itoa(index)),
				zap.Uint64("goroutine_id", logger.GetGoroutineID()),
			)
			workerPool.workerLoop(&gcl)
		}(i)
	}

	return workerPool
}

func (w *WorkerPool) workerLoop(cl *logger.CtxLogger) {
	for {
		select {
		case <-w.Ctx.Done():
			cl.Log.Debug("worker loop exit with cancel")
			return
		case task, ok := <-w.tasks:
			if !ok {
				cl.Log.Debug("worker loop exit with chan close")
				return
			}
			task.Execute(w.Ctx)
		}
	}
}

// Submit 提交一个任务到工作池
// 返回 true 表示任务已提交，false 表示工作池已关闭
func (w *WorkerPool) Submit(task Task) bool {
	select {
	case w.tasks <- task:
		return true
	case <-w.Ctx.Done():
		return false
	}
}

// NewGroup 创建一个新的任务组，用于批量提交并等待任务完成
func (w *WorkerPool) NewGroup() *TaskGroup {
	return &TaskGroup{pool: w}
}

// Submit 向任务组提交任务
func (g *TaskGroup) Submit(task Task) bool {
	select {
	case <-g.pool.Ctx.Done():
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
	case <-g.pool.Ctx.Done():
		g.wg.Done()
		return false
	}
}

// Wait 等待任务组中的所有任务完成
func (g *TaskGroup) Wait() {
	g.wg.Wait()
}

// Close 优雅关闭工作池
// timeout: 等待所有任务完成的超时时间
func (w *WorkerPool) Close(timeout time.Duration) {
	close(w.tasks)
	w.cancel()

	done := make(chan struct{})
	go func() {
		w.wait.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.Log.Info("worker pool exit beautifully")
	case <-time.After(timeout):
		w.Log.Warn("worker pool exit within timeout")
	}
}
