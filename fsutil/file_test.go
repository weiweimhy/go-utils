package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFileWithPerm(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fsutil_test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test.file")
	data := []byte("test content")
	perm := os.FileMode(0600)

	err := SaveToFileWithPerm(path, data, perm)
	if err != nil {
		t.Fatalf("SaveToFileWithPerm failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// In Windows, Unix-style permissions might not be fully reflected,
	// but we can at least check if the file was created.
	// We'll skip exact perm check on Windows but create it for cross-platform robustness.
	if info.Mode().Perm() != perm && os.PathSeparator != '\\' {
		t.Errorf("expected perm %o, got %o", perm, info.Mode().Perm())
	}
}

func TestFileExistsReturnsFalseForMissingFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fsutil_exist_test")
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "missing.file")
	if FileExists(path) {
		t.Fatalf("expected missing file to return false")
	}
}
