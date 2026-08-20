package streamutil

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestReadAllLimit(t *testing.T) {
	data, err := ReadAllLimit(strings.NewReader("hello"), 5)
	if err != nil || string(data) != "hello" {
		t.Fatalf("ReadAllLimit() = %q, %v", data, err)
	}
	if _, err := ReadAllLimit(strings.NewReader("hello"), 4); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("ReadAllLimit() error = %v, want ErrLimitExceeded", err)
	}
	if _, err := ReadAllLimit(nil, 1); !errors.Is(err, ErrNilReader) {
		t.Fatalf("ReadAllLimit(nil) error = %v, want ErrNilReader", err)
	}
}

func TestLimitedBuffer(t *testing.T) {
	b := NewLimitedBuffer(5)
	for _, input := range []string{"abc", "def"} {
		if n, err := b.Write([]byte(input)); err != nil || n != len(input) {
			t.Fatalf("Write(%q) = %d, %v", input, n, err)
		}
	}
	if got := b.String(); got != "abcde" {
		t.Fatalf("String() = %q, want abcde", got)
	}
	if !b.Overflowed() {
		t.Fatal("Overflowed() = false, want true")
	}
}

func TestLimitedBufferConcurrentWrites(t *testing.T) {
	b := NewLimitedBuffer(64)
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = b.Write([]byte("0123456789"))
		}()
	}
	wg.Wait()
	if got := len(b.Bytes()); got != 64 {
		t.Fatalf("retained bytes = %d, want 64", got)
	}
	if !b.Overflowed() {
		t.Fatal("Overflowed() = false, want true")
	}
}
