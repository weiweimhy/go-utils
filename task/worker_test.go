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

func TestWorkerPoolCloseCancelsBlockedSubmit(t *testing.T) {
	pool := NewWorkerPool(
		context.Background(),
		WithWorkerCount(1),
		WithBufferSize(1),
	)
	started := make(chan struct{})
	release := make(chan struct{})
	if !pool.SubmitFunc(func(ctx context.Context) {
		close(started)
		<-release
	}) {
		t.Fatal("expected first task to be accepted")
	}
	<-started
	if !pool.SubmitFunc(func(context.Context) {}) {
		t.Fatal("expected queued task to be accepted")
	}

	submitResult := make(chan bool, 1)
	go func() {
		submitResult <- pool.SubmitFunc(func(context.Context) {})
	}()

	if pool.Close(20 * time.Millisecond) {
		t.Fatal("expected close to time out while a task is blocked")
	}
	if submitted := <-submitResult; submitted {
		t.Fatal("blocked submit should fail after close cancels the pool")
	}

	close(release)
	if pool.Close(time.Second) {
		t.Fatal("a pool that timed out must keep reporting a non-graceful close")
	}
}

func TestWorkerPoolCloseFromTaskDoesNotDeadlock(t *testing.T) {
	pool := NewWorkerPool(context.Background(), WithWorkerCount(1))
	closedFromTask := make(chan struct{})
	if !pool.SubmitFunc(func(context.Context) {
		if pool.Close(time.Second) {
			t.Error("Close() from a worker should not claim synchronous completion")
		}
		close(closedFromTask)
	}) {
		t.Fatal("expected task to be accepted")
	}

	select {
	case <-closedFromTask:
	case <-time.After(time.Second):
		t.Fatal("Close() from a task deadlocked")
	}
	if !pool.Close(time.Second) {
		t.Fatal("external Close() should observe graceful completion")
	}
}

func TestTaskGroupWaitSealsSubmissions(t *testing.T) {
	pool := NewWorkerPool(context.Background())
	defer pool.Close(time.Second)
	group := pool.NewGroup()
	group.Wait()
	if group.SubmitFunc(func(context.Context) {}) {
		t.Fatal("group should reject submissions after Wait starts")
	}
}
