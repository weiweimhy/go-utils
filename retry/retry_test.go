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
