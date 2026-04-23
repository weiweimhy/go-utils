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

var workerPoolIDCounter atomic.Uint64

type WorkerPoolOption func(*workerPoolConfig)

type workerPoolConfig struct {
	workerCount int
	bufferSize  int
	name        string
}

func defaultWorkerPoolConfig() workerPoolConfig {
	workerCount := runtime.NumCPU()
	return workerPoolConfig{
		workerCount: workerCount,
		bufferSize:  workerCount,
	}
}

func WithWorkerCount(count int) WorkerPoolOption {
	return func(cfg *workerPoolConfig) {
		if count > 0 {
			cfg.workerCount = count
		}
	}
}

func WithBufferSize(size int) WorkerPoolOption {
	return func(cfg *workerPoolConfig) {
		if size >= 0 {
			cfg.bufferSize = size
		}
	}
}

func WithName(name string) WorkerPoolOption {
	return func(cfg *workerPoolConfig) {
		cfg.name = name
	}
}

// WorkerPool 是一个通用的工作池实现
type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *zap.Logger
	id     uint64

	stateMu   sync.RWMutex
	tasks     chan Task
	wait      sync.WaitGroup
	closeOnce sync.Once
	closed    atomic.Bool
}

// NewWorkerPool 创建一个新的工作池。
// 可通过 WithWorkerCount、WithBufferSize、WithName 等选项覆盖默认配置。
func NewWorkerPool(ctx context.Context, opts ...WorkerPoolOption) *WorkerPool {
	cfg := defaultWorkerPoolConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	poolID := workerPoolIDCounter.Add(1)
	fields := []zap.Field{
		zap.String("module", "WorkerPool"),
		zap.Uint64("pool_id", poolID),
	}
	if cfg.name != "" {
		fields = append(fields, zap.String("pool_name", cfg.name))
	}
	log := logger.FromContext(ctx, fields...)

	workerPool := &WorkerPool{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		id:     poolID,
		tasks:  make(chan Task, cfg.bufferSize),
	}

	for i := 0; i < cfg.workerCount; i++ {
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
	w.stateMu.RLock()
	defer w.stateMu.RUnlock()
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

// Close 优雅关闭工作池
// timeout: 等待所有任务完成的超时时间
// 返回 true 表示正常关闭，false 表示超时
func (w *WorkerPool) Close(timeout time.Duration) bool {
	graceful := true

	w.closeOnce.Do(func() {
		w.stateMu.Lock()
		w.closed.Store(true)
		close(w.tasks)
		w.stateMu.Unlock()

		done := make(chan struct{})
		go func() {
			w.wait.Wait()
			close(done)
		}()

		select {
		case <-done:
			w.cancel()
			w.log.Info("worker pool closed gracefully")
		case <-time.After(timeout):
			w.cancel()
			w.log.Warn("worker pool closed with timeout")
			graceful = false
		}
	})

	return graceful
}

// TaskGroup 用于分组等待一批任务完成
type TaskGroup struct {
	pool *WorkerPool
	wg   sync.WaitGroup
}

// Submit 向任务组提交任务
func (g *TaskGroup) Submit(task Task) bool {
	g.pool.stateMu.RLock()
	defer g.pool.stateMu.RUnlock()
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
