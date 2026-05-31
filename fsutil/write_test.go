package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteFileAtomicCreatesParents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.txt")

	err := WriteFile(path, []byte("hello"), WriteOptions{
		DirPerm:    SecureDirPerm,
		FilePerm:   SecureFilePerm,
		Atomic:     true,
		Sync:       true,
		CreateDirs: true,
	})
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("content = %q, want hello", string(got))
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if info.Mode().Perm() != SecureFilePerm {
			t.Fatalf("file mode = %o, want %o", info.Mode().Perm(), SecureFilePerm)
		}
	}
}

func TestSecureWriteFileReplacesExistingAtomically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := SecureWriteFile(path, []byte("new")); err != nil {
		t.Fatalf("SecureWriteFile() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("content = %q, want new", string(got))
	}
}
