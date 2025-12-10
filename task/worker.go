package task

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/weiweimhy/go-utils/v2/logger"
	"go.uber.org/zap"
)

type WorkerPool struct {
	logger.CtxLogger

	tasks  chan Task
	cancel context.CancelFunc
	wait   sync.WaitGroup
}

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

func (w *WorkerPool) Submit(task Task) bool {
	select {
	case w.tasks <- task:
		return true
	case <-w.Ctx.Done():
		return false
	}
}

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
		w.Log.Info("worker pool exit beautifully", zap.Uint64("goroutine_id", logger.GetGoroutineID()))
	case <-time.After(timeout):
		w.Log.Warn("worker pool exit within timeout", zap.Uint64("goroutine_id", logger.GetGoroutineID()))
	}
}
