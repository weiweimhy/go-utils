package auditlog

import (
	"os"
	"path/filepath"
	"strings"
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
