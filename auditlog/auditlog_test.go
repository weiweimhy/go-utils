package auditlog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWriterWriteSyncClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit", "events.jsonl")
	writer, err := Open(path, Options{
		Redact: func(record any) any {
			return map[string]any{"event": "login", "token": "[REDACTED]"}
		},
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := writer.Write(map[string]any{"event": "login", "token": "secret"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(raw), "[REDACTED]") || strings.Contains(string(raw), "secret") {
		t.Fatalf("unexpected audit content: %q", string(raw))
	}
}

func TestWriterConcurrentClose(t *testing.T) {
	writer, err := Open(filepath.Join(t.TempDir(), "events.jsonl"), Options{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range 16 {
				_ = writer.Write(map[string]string{"event": "login"})
				_ = writer.Sync()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		_ = writer.Close()
	}()

	close(start)
	wg.Wait()
	if err := writer.Write(map[string]string{"event": "after-close"}); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Write() error = %v, want os.ErrClosed", err)
	}
	if err := writer.Sync(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Sync() error = %v, want os.ErrClosed", err)
	}
}
