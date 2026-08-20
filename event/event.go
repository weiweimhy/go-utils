package event

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Handler handles an event payload for the given event type.
type Handler func(eventType string, data any)

// PanicHandler is invoked when a handler panics during dispatch.
type PanicHandler func(eventType string, data any, recovered any)

// Option customizes Bus behavior.
type Option func(*config)

type config struct {
	workerCount  int
	bufferSize   int
	panicHandler PanicHandler
}

func defaultConfig() config {
	workerCount := runtime.NumCPU()
	if workerCount < 2 {
		workerCount = 2
	}

	return config{
		workerCount: workerCount,
		bufferSize:  workerCount * 32,
	}
}

// WithWorkerCount overrides the async dispatcher worker count.
func WithWorkerCount(count int) Option {
	return func(cfg *config) {
		if count > 0 {
			cfg.workerCount = count
		}
	}
}

// WithBufferSize overrides the async queue size.
func WithBufferSize(size int) Option {
	return func(cfg *config) {
		if size >= 0 {
			cfg.bufferSize = size
		}
	}
}

// WithPanicHandler sets an optional panic callback for recovered handler panics.
func WithPanicHandler(handler PanicHandler) Option {
	return func(cfg *config) {
		cfg.panicHandler = handler
	}
}

// Bus is an in-process event bus with optional async dispatch workers.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string]map[uint64]Handler
	queue    chan dispatchTask

	panicHandler PanicHandler

	nextHandlerID atomic.Uint64
	closed        atomic.Bool
	closeOnce     sync.Once
	publishers    sync.WaitGroup
	workers       sync.WaitGroup
}

type dispatchTask struct {
	eventType string
	data      any
	handler   Handler
}

// NewBus creates a new bus instance.
func NewBus(opts ...Option) *Bus {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	b := &Bus{
		handlers:     make(map[string]map[uint64]Handler),
		queue:        make(chan dispatchTask, cfg.bufferSize),
		panicHandler: cfg.panicHandler,
	}

	for i := 0; i < cfg.workerCount; i++ {
		b.workers.Add(1)
		go func() {
			defer b.workers.Done()
			b.runDispatcher()
		}()
	}

	return b
}

// Subscribe registers a handler for eventType.
// It returns an unsubscribe function and whether the subscription succeeded.
func (b *Bus) Subscribe(eventType string, handler Handler) (func(), bool) {
	if handler == nil {
		return func() {}, false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed.Load() {
		return func() {}, false
	}

	id := b.nextHandlerID.Add(1)
	if b.handlers[eventType] == nil {
		b.handlers[eventType] = make(map[uint64]Handler)
	}
	b.handlers[eventType][id] = handler

	var once sync.Once
	return func() {
		once.Do(func() {
			b.unsubscribe(eventType, id)
		})
	}, true
}

func (b *Bus) unsubscribe(eventType string, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventType]
	if handlers == nil {
		return
	}
	delete(handlers, id)
	if len(handlers) == 0 {
		delete(b.handlers, eventType)
	}
}

// Publish dispatches an event asynchronously.
// It returns false when the bus is already closed.
func (b *Bus) Publish(eventType string, data any) bool {
	b.mu.RLock()
	if b.closed.Load() {
		b.mu.RUnlock()
		return false
	}

	registered := b.handlers[eventType]
	handlers := make([]Handler, 0, len(registered))
	for _, handler := range registered {
		handlers = append(handlers, handler)
	}
	b.publishers.Add(1)
	b.mu.RUnlock()
	defer b.publishers.Done()

	for _, handler := range handlers {
		b.queue <- dispatchTask{
			eventType: eventType,
			data:      data,
			handler:   handler,
		}
	}

	return true
}

// PublishSync dispatches an event synchronously in the caller goroutine.
// It returns false when the bus is already closed.
func (b *Bus) PublishSync(eventType string, data any) bool {
	handlers, ok := b.snapshotHandlers(eventType)
	if !ok {
		return false
	}

	for _, handler := range handlers {
		b.safelyInvoke(handler, eventType, data)
	}

	return true
}

func (b *Bus) snapshotHandlers(eventType string) ([]Handler, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed.Load() {
		return nil, false
	}

	registered := b.handlers[eventType]
	if len(registered) == 0 {
		return nil, true
	}

	handlers := make([]Handler, 0, len(registered))
	for _, handler := range registered {
		handlers = append(handlers, handler)
	}
	return handlers, true
}

// HasSubscribers reports whether eventType has at least one subscriber.
func (b *Bus) HasSubscribers(eventType string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.handlers[eventType]) > 0
}

// GetSubscriberCount returns the subscriber count for eventType.
func (b *Bus) GetSubscriberCount(eventType string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.handlers[eventType])
}

// Clear removes all subscriptions across all event types.
func (b *Bus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers = make(map[string]map[uint64]Handler)
}

// ClearEventType removes all subscriptions for eventType.
func (b *Bus) ClearEventType(eventType string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.handlers, eventType)
}

// IsClosed reports whether Close has been called.
func (b *Bus) IsClosed() bool {
	return b.closed.Load()
}

// Close stops accepting new subscriptions and publish requests, then drains queued tasks.
func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		b.mu.Lock()
		b.closed.Store(true)
		b.mu.Unlock()
		b.publishers.Wait()
		close(b.queue)
		b.workers.Wait()
	})
}

func (b *Bus) runDispatcher() {
	for task := range b.queue {
		b.safelyInvoke(task.handler, task.eventType, task.data)
	}
}

func (b *Bus) safelyInvoke(handler Handler, eventType string, data any) {
	defer func() {
		if recovered := recover(); recovered != nil && b.panicHandler != nil {
			b.panicHandler(eventType, data, recovered)
		}
	}()

	handler(eventType, data)
}
