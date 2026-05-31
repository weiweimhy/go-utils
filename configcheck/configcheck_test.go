package configcheck

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("ok"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := CheckFile(path)
	if err != nil {
		t.Fatalf("CheckFile() error = %v", err)
	}
	if !result.Regular {
		t.Fatal("expected regular file")
	}
}

func TestCheckFileWorldWritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not reliable on Windows")
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("ok"), 0666); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := CheckFile(path)
	if !errors.Is(err, ErrWorldWritable) {
		t.Fatalf("CheckFile() error = %v, want ErrWorldWritable", err)
	}
}
