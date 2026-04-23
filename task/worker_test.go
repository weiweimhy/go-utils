package task

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPoolWithOptions(t *testing.T) {
	pool := NewWorkerPool(
		context.Background(),
		WithWorkerCount(2),
		WithBufferSize(4),
		WithName("test-pool"),
	)
	defer pool.Close(time.Second)

	if pool == nil {
		t.Fatal("expected pool")
	}
	if cap(pool.tasks) != 4 {
		t.Fatalf("expected buffer size 4, got %d", cap(pool.tasks))
	}
}

func TestNewWorkerPoolDefaults(t *testing.T) {
	pool := NewWorkerPool(context.Background())
	defer pool.Close(time.Second)

	if pool == nil {
		t.Fatal("expected pool")
	}
	if cap(pool.tasks) != runtime.NumCPU() {
		t.Fatalf("expected default buffer size %d, got %d", runtime.NumCPU(), cap(pool.tasks))
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		if !pool.SubmitFunc(func(ctx context.Context) {}) {
			t.Fatal("expected default pool to accept task submission")
		}
	}
}

func TestWorkerPoolCloseDrainsQueuedTasks(t *testing.T) {
	pool := NewWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithBufferSize(4),
	)

	start := make(chan struct{})
	done := make(chan struct{})
	pool.SubmitFunc(func(ctx context.Context) {
		close(start)
		time.Sleep(50 * time.Millisecond)
		close(done)
	})

	<-start

	var executed atomic.Int32
	for range 3 {
		if !pool.SubmitFunc(func(ctx context.Context) {
			executed.Add(1)
		}) {
			t.Fatal("expected queued submission to succeed")
		}
	}

	if !pool.Close(time.Second) {
		t.Fatal("expected close to drain queued tasks")
	}

	<-done

	if got := executed.Load(); got != 3 {
		t.Fatalf("expected 3 queued tasks to run, got %d", got)
	}
}

func TestWorkerPoolCloseRejectsNewTasks(t *testing.T) {
	pool := NewWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithBufferSize(1),
	)
	if !pool.Close(time.Second) {
		t.Fatal("expected close to succeed")
	}
	if pool.SubmitFunc(func(ctx context.Context) {}) {
		t.Fatal("submit should fail after close")
	}
}
