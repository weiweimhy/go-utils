package timeutil

import (
	"context"
	"strconv"
	"time"
)

// Clock abstracts time for tests.
type Clock interface {
	Now() time.Time
	Sleep(ctx context.Context, d time.Duration) error
}

// RealClock uses the standard time package.
type RealClock struct{}

// Now returns the current local time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// Sleep sleeps until d elapses or ctx is cancelled.
func (RealClock) Sleep(ctx context.Context, d time.Duration) error {
	return SleepContext(ctx, d)
}

// NowFunc adapts a function into a current-time provider.
type NowFunc func() time.Time

// Now calls f or time.Now when f is nil.
func (f NowFunc) Now() time.Time {
	if f == nil {
		return time.Now()
	}
	return f()
}

// ParseDurationDefault parses s as a time.Duration or returns fallback on empty/invalid input.
func ParseDurationDefault(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}

// SleepContext sleeps until d elapses or ctx is cancelled.
func SleepContext(ctx context.Context, d time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SinceMillis returns elapsed milliseconds since t.
func SinceMillis(t time.Time) int64 {
	return time.Since(t).Milliseconds()
}

// UnixMillisString returns t as Unix milliseconds in base 10.
func UnixMillisString(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}
