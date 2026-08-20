package streamutil

import (
	"errors"
	"io"
	"math"
	"sync"
)

var (
	// ErrNilReader is returned when a read helper receives a nil reader.
	ErrNilReader = errors.New("streamutil: nil reader")
	// ErrLimitExceeded is returned when input exceeds a configured byte limit.
	ErrLimitExceeded = errors.New("streamutil: byte limit exceeded")
	// ErrNilBuffer is returned when Write is called on a nil LimitedBuffer.
	ErrNilBuffer = errors.New("streamutil: nil limited buffer")
)

// ReadAllLimit reads r completely unless it exceeds maxBytes.
// A non-positive maxBytes disables the limit. The returned data is never
// silently truncated.
func ReadAllLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if r == nil {
		return nil, ErrNilReader
	}
	if maxBytes <= 0 || maxBytes == math.MaxInt64 {
		return io.ReadAll(r)
	}

	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrLimitExceeded
	}
	return data, nil
}

// LimitedBuffer is a concurrency-safe io.Writer that keeps at most MaxBytes
// bytes while recording whether data overflowed its limit. Write always
// accepts the full input so it can safely receive subprocess output.
type LimitedBuffer struct {
	mu       sync.RWMutex
	maxBytes int64
	data     []byte
	overflow bool
}

// NewLimitedBuffer creates a buffer that retains at most maxBytes bytes.
// A negative limit is treated as zero.
func NewLimitedBuffer(maxBytes int64) *LimitedBuffer {
	if maxBytes < 0 {
		maxBytes = 0
	}
	return &LimitedBuffer{maxBytes: maxBytes}
}

// Write implements io.Writer.
func (b *LimitedBuffer) Write(p []byte) (int, error) {
	if b == nil {
		return 0, ErrNilBuffer
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	remaining := b.maxBytes - int64(len(b.data))
	if remaining > 0 {
		keep := int64(len(p))
		if keep > remaining {
			keep = remaining
		}
		b.data = append(b.data, p[:int(keep)]...)
	}
	if int64(len(p)) > remaining {
		b.overflow = true
	}
	return len(p), nil
}

// Bytes returns a copy of the retained bytes.
func (b *LimitedBuffer) Bytes() []byte {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]byte(nil), b.data...)
}

// String returns the retained bytes as a string.
func (b *LimitedBuffer) String() string {
	return string(b.Bytes())
}

// Overflowed reports whether any input was omitted from the retained buffer.
func (b *LimitedBuffer) Overflowed() bool {
	if b == nil {
		return false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.overflow
}

// MaxBytes returns the buffer's retention limit.
func (b *LimitedBuffer) MaxBytes() int64 {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.maxBytes
}
