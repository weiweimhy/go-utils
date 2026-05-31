package syncutil

import (
	"context"
	"sync"
)

// OnceValueWithError runs fn once and caches its value and error.
func OnceValueWithError[T any](fn func() (T, error)) func() (T, error) {
	var (
		once sync.Once
		v    T
		err  error
	)
	return func() (T, error) {
		once.Do(func() {
			v, err = fn()
		})
		return v, err
	}
}

// Semaphore is a small weighted-by-one semaphore.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a semaphore with n slots.
func NewSemaphore(n int) *Semaphore {
	if n <= 0 {
		n = 1
	}
	return &Semaphore{ch: make(chan struct{}, n)}
}

// Acquire waits for one slot or returns ctx.Err().
func (s *Semaphore) Acquire(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire one slot without waiting.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release releases one slot.
func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
		panic("syncutil: release without acquire")
	}
}
