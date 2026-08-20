package task

import (
	"context"
	"log/slog"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var workerPoolIDCounter atomic.Uint64

type WorkerPoolOption func(*workerPoolConfig)

type workerPoolConfig struct {
	workerCount int
	bufferSize  int
	name        string
	log         *slog.Logger
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

// WithLogger sets an optional logger for worker-pool lifecycle events.
// When unset, the worker pool stays silent by default.
func WithLogger(log *slog.Logger) WorkerPoolOption {
	return func(cfg *workerPoolConfig) {
		cfg.log = log
	}
}

// WorkerPool 是一个通用的工作池实现
type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *slog.Logger
	id     uint64

	stateMu      sync.Mutex
	tasks        chan Task
	submitters   sync.WaitGroup
	wait         sync.WaitGroup
	closeOnce    sync.Once
	closeStarted chan struct{}
	closeDone    chan struct{}
	closed       atomic.Bool
	timedOut     atomic.Bool
	workerIDs    sync.Map
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

	parentCtx := ctx
	ctx, cancel := context.WithCancel(parentCtx)
	poolID := workerPoolIDCounter.Add(1)
	log := cfg.log
	if log != nil {
		attrs := []any{
			"module", "WorkerPool",
			"pool_id", poolID,
		}
		if cfg.name != "" {
			attrs = append(attrs, "pool_name", cfg.name)
		}
		log = log.With(attrs...)
	}

	workerPool := &WorkerPool{
		ctx:          ctx,
		cancel:       cancel,
		log:          log,
		id:           poolID,
		tasks:        make(chan Task, cfg.bufferSize),
		closeStarted: make(chan struct{}),
		closeDone:    make(chan struct{}),
	}

	for i := 0; i < cfg.workerCount; i++ {
		workerPool.wait.Add(1)
		go func(index int) {
			defer workerPool.wait.Done()
			workerPool.workerLoop(index)
		}(i)
	}
	go func() {
		select {
		case <-parentCtx.Done():
			workerPool.Close(0)
		case <-workerPool.closeStarted:
		}
	}()

	return workerPool
}

func (w *WorkerPool) workerLoop(index int) {
	workerID := currentGoroutineID()
	if workerID != 0 {
		w.workerIDs.Store(workerID, struct{}{})
		defer w.workerIDs.Delete(workerID)
	}

	log := w.log
	if log != nil {
		log = log.With("worker_index", index)
	}

	for task := range w.tasks {
		w.safeExecute(log, task)
	}
	if log != nil {
		log.DebugContext(context.Background(), "worker exit: channel closed")
	}
}

// safeExecute 安全执行任务，捕获 panic
func (w *WorkerPool) safeExecute(log *slog.Logger, task Task) {
	defer func() {
		if r := recover(); r != nil {
			if log != nil {
				log.ErrorContext(w.ctx, "task panic recovered", "panic", r, "stack", string(debug.Stack()))
			}
		}
	}()
	task.Execute(w.ctx)
}

// Submit 提交一个任务到工作池
// 返回 true 表示任务已提交，false 表示工作池已关闭
func (w *WorkerPool) Submit(task Task) bool {
	w.stateMu.Lock()
	if w.closed.Load() {
		w.stateMu.Unlock()
		return false
	}
	w.submitters.Add(1)
	w.stateMu.Unlock()
	defer w.submitters.Done()

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
// 返回 true 表示调用者观察到正常关闭，false 表示超时或调用者是池内任务。
// 池内任务调用 Close 时只会发起关闭，避免等待自身退出造成死锁。
func (w *WorkerPool) Close(timeout time.Duration) bool {
	w.closeOnce.Do(func() {
		close(w.closeStarted)
		w.stateMu.Lock()
		w.closed.Store(true)
		w.stateMu.Unlock()
		go func() {
			w.submitters.Wait()
			close(w.tasks)
			w.wait.Wait()
			w.cancel()
			close(w.closeDone)
		}()
	})
	if w.isWorker() {
		return false
	}
	if timeout <= 0 {
		select {
		case <-w.closeDone:
			return !w.timedOut.Load()
		default:
			w.timedOut.Store(true)
			w.cancel()
			if w.log != nil {
				w.log.WarnContext(context.Background(), "worker pool closed with timeout")
			}
			return false
		}
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.closeDone:
		return !w.timedOut.Load()
	case <-timer.C:
		w.timedOut.Store(true)
		w.cancel()
		if w.log != nil {
			w.log.WarnContext(context.Background(), "worker pool closed with timeout")
		}
		return false
	}
}

func (w *WorkerPool) isWorker() bool {
	workerID := currentGoroutineID()
	if workerID == 0 {
		return false
	}
	_, ok := w.workerIDs.Load(workerID)
	return ok
}

func currentGoroutineID() uint64 {
	var buffer [64]byte
	n := runtime.Stack(buffer[:], false)
	fields := strings.Fields(string(buffer[:n]))
	if len(fields) < 2 || fields[0] != "goroutine" {
		return 0
	}
	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// TaskGroup 用于分组等待一批任务完成
type TaskGroup struct {
	pool   *WorkerPool
	wg     sync.WaitGroup
	mu     sync.Mutex
	sealed bool
}

// Submit 向任务组提交任务
func (g *TaskGroup) Submit(task Task) bool {
	g.mu.Lock()
	if g.sealed {
		g.mu.Unlock()
		return false
	}
	g.wg.Add(1)
	g.mu.Unlock()

	wrapped := TaskFunc(func(ctx context.Context) {
		defer g.wg.Done()
		task.Execute(ctx)
	})
	if g.pool.Submit(wrapped) {
		return true
	}
	g.wg.Done()
	return false
}

// SubmitFunc 向任务组提交函数作为任务
func (g *TaskGroup) SubmitFunc(fn func(ctx context.Context)) bool {
	return g.Submit(TaskFunc(fn))
}

// Wait 等待任务组中的所有任务完成
func (g *TaskGroup) Wait() {
	g.mu.Lock()
	g.sealed = true
	g.mu.Unlock()
	g.wg.Wait()
}
