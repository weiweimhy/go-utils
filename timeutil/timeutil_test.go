package timeutil

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParseDurationDefault(t *testing.T) {
	if got := ParseDurationDefault("2s", time.Second); got != 2*time.Second {
		t.Fatalf("ParseDurationDefault() = %v", got)
	}
	if got := ParseDurationDefault("bad", time.Second); got != time.Second {
		t.Fatalf("ParseDurationDefault() = %v", got)
	}
}

func TestSleepContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := SleepContext(ctx, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SleepContext() error = %v", err)
	}
}
