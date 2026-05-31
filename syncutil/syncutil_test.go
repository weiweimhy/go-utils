package syncutil

import (
	"context"
	"errors"
	"testing"
)

func TestOnceValueWithError(t *testing.T) {
	var calls int
	once := OnceValueWithError(func() (int, error) {
		calls++
		return 42, nil
	})
	for range 3 {
		got, err := once()
		if err != nil || got != 42 {
			t.Fatalf("once() = %d, %v", got, err)
		}
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(1)
	if err := sem.Acquire(context.Background()); err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if sem.TryAcquire() {
		t.Fatal("TryAcquire() should fail when full")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sem.Acquire(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Acquire() error = %v, want context.Canceled", err)
	}
	sem.Release()
	if !sem.TryAcquire() {
		t.Fatal("TryAcquire() should succeed after release")
	}
	sem.Release()
}
