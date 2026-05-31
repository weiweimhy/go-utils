package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// Options controls retry behavior.
type Options struct {
	Attempts int
	Delay    time.Duration
	MaxDelay time.Duration
	Jitter   bool
}

func (opts Options) withDefaults() Options {
	if opts.Attempts <= 0 {
		opts.Attempts = 1
	}
	return opts
}

// Do calls fn until it succeeds, attempts are exhausted, or ctx is cancelled.
func Do(ctx context.Context, fn func(context.Context) error, opts Options) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if fn == nil {
		return nil
	}
	opts = opts.withDefaults()

	var errs []error
	delay := opts.Delay
	for attempt := 1; attempt <= opts.Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}
		err := fn(ctx)
		if err == nil {
			return nil
		}
		errs = append(errs, err)
		if attempt == opts.Attempts {
			break
		}
		if delay > 0 {
			sleep := withJitter(delay, opts.Jitter)
			if err := sleepContext(ctx, sleep); err != nil {
				return errors.Join(append(errs, err)...)
			}
			delay = nextDelay(delay, opts.MaxDelay)
		}
	}
	return errors.Join(errs...)
}

func nextDelay(delay, maxDelay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	next := delay * 2
	if maxDelay > 0 && next > maxDelay {
		return maxDelay
	}
	return next
}

func withJitter(delay time.Duration, enabled bool) time.Duration {
	if !enabled || delay <= 0 {
		return delay
	}
	half := delay / 2
	return half + time.Duration(rand.Int64N(int64(delay-half)+1))
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
