package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoEventuallySucceeds(t *testing.T) {
	var attempts int
	err := Do(context.Background(), func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary")
		}
		return nil
	}, Options{Attempts: 3})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestDoStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, func(ctx context.Context) error {
		t.Fatal("fn should not be called")
		return nil
	}, Options{Attempts: 3, Delay: time.Second})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() error = %v, want context.Canceled", err)
	}
}

func TestNextDelayDoesNotOverflow(t *testing.T) {
	const maxDuration = time.Duration(1<<63 - 1)
	got := nextDelay(maxDuration/2+1, 0)
	if got != maxDuration {
		t.Fatalf("nextDelay() = %v, want %v", got, maxDuration)
	}
	if got := nextDelay(time.Hour, time.Minute); got != time.Minute {
		t.Fatalf("nextDelay() with cap = %v, want %v", got, time.Minute)
	}
}
