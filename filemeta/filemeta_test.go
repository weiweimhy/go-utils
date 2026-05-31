package filemeta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetWithSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	meta, err := Get(path, Options{ComputeSHA256: true})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if meta.SizeBytes != 5 {
		t.Fatalf("SizeBytes = %d, want 5", meta.SizeBytes)
	}
	if meta.SHA256 != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256 = %q", meta.SHA256)
	}
}
