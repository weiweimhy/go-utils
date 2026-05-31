package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubscribeAndPublishSync(t *testing.T) {
	bus := NewBus()
	defer bus.Close()

	called := false
	unsubscribe, ok := bus.Subscribe("task.created", func(eventType string, data any) {
		called = true
		if eventType != "task.created" {
			t.Fatalf("unexpected event type: %s", eventType)
		}
		if got := data.(string); got != "payload" {
			t.Fatalf("unexpected payload: %s", got)
		}
	})
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}
	defer unsubscribe()

	if !bus.PublishSync("task.created", "payload") {
		t.Fatal("expected publish sync to succeed")
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	bus := NewBus()
	defer bus.Close()

	var calls atomic.Int32
	unsubscribe, ok := bus.Subscribe("task.created", func(eventType string, data any) {
		calls.Add(1)
	})
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}

	unsubscribe()
	unsubscribe()

	if !bus.PublishSync("task.created", nil) {
		t.Fatal("expected publish sync to succeed")
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("expected 0 calls, got %d", got)
	}
}

func TestPublishAsyncAndCloseDrainsQueue(t *testing.T) {
	bus := NewBus(WithWorkerCount(1), WithBufferSize(4))

	start := make(chan struct{})
	done := make(chan struct{})
	unsubscribe, ok := bus.Subscribe("task.created", func(eventType string, data any) {
		close(start)
		time.Sleep(50 * time.Millisecond)
		close(done)
	})
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}
	defer unsubscribe()

	if !bus.Publish("task.created", nil) {
		t.Fatal("expected async publish to succeed")
	}
	<-start

	var queued atomic.Int32
	unsubscribeQueued, ok := bus.Subscribe("task.created", func(eventType string, data any) {
		queued.Add(1)
	})
	if !ok {
		t.Fatal("expected second subscribe to succeed")
	}
	defer unsubscribeQueued()

	for range 3 {
		if !bus.Publish("task.created", nil) {
			t.Fatal("expected queued publish to succeed")
		}
	}

	bus.Close()
	<-done

	if got := queued.Load(); got != 3 {
		t.Fatalf("expected 3 queued handler calls, got %d", got)
	}
}

func TestCloseRejectsNewWork(t *testing.T) {
	bus := NewBus()
	bus.Close()

	if _, ok := bus.Subscribe("task.created", func(eventType string, data any) {}); ok {
		t.Fatal("subscribe should fail after close")
	}
	if bus.Publish("task.created", nil) {
		t.Fatal("publish should fail after close")
	}
	if bus.PublishSync("task.created", nil) {
		t.Fatal("publish sync should fail after close")
	}
}

func TestRecoveredPanicsInvokePanicHandler(t *testing.T) {
	var recovered atomic.Int32
	bus := NewBus(WithPanicHandler(func(eventType string, data any, panicValue any) {
		recovered.Add(1)
		if eventType != "task.created" {
			t.Fatalf("unexpected event type: %s", eventType)
		}
		if panicValue != "boom" {
			t.Fatalf("unexpected panic value: %v", panicValue)
		}
	}))
	defer bus.Close()

	_, ok := bus.Subscribe("task.created", func(eventType string, data any) {
		panic("boom")
	})
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}

	if !bus.PublishSync("task.created", nil) {
		t.Fatal("expected publish sync to succeed")
	}
	if got := recovered.Load(); got != 1 {
		t.Fatalf("expected 1 recovered panic, got %d", got)
	}
}

func TestClearEventType(t *testing.T) {
	bus := NewBus()
	defer bus.Close()

	if _, ok := bus.Subscribe("task.created", func(eventType string, data any) {}); !ok {
		t.Fatal("expected subscribe to succeed")
	}
	if _, ok := bus.Subscribe("task.updated", func(eventType string, data any) {}); !ok {
		t.Fatal("expected subscribe to succeed")
	}

	bus.ClearEventType("task.created")

	if bus.HasSubscribers("task.created") {
		t.Fatal("expected event type subscribers to be cleared")
	}
	if !bus.HasSubscribers("task.updated") {
		t.Fatal("expected other event type subscribers to remain")
	}
}

func TestPublishConcurrentWithCloseDoesNotPanic(t *testing.T) {
	bus := NewBus(WithWorkerCount(4), WithBufferSize(16))
	_, ok := bus.Subscribe("tick", func(eventType string, data any) {})
	if !ok {
		t.Fatal("expected subscribe to succeed")
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for {
				if !bus.Publish("tick", nil) {
					return
				}
			}
		}()
	}

	close(start)
	time.Sleep(10 * time.Millisecond)
	bus.Close()
	wg.Wait()
}
