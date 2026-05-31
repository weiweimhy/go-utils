package fsutil

import (
	"os"
	"path/filepath"
)

const (
	// DefaultDirPerm is used when a helper creates normal application directories.
	DefaultDirPerm os.FileMode = 0755
	// SecureDirPerm is used by helpers intended for private data.
	SecureDirPerm os.FileMode = 0700
)

// DirExists reports whether the given path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDir creates the given directory path, including any missing parents.
func CreateDir(path string) error {
	return CreateDirWithPerm(path, DefaultDirPerm)
}

// CreateDirWithPerm creates the given directory path with perm, including any missing parents.
func CreateDirWithPerm(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultDirPerm
	}
	return os.MkdirAll(path, perm)
}

// CreateParentDir creates the parent directory for the given file path.
func CreateParentDir(path string) error {
	return CreateParentDirWithPerm(path, DefaultDirPerm)
}

// CreateParentDirWithPerm creates the parent directory for the given file path with perm.
func CreateParentDirWithPerm(path string, perm os.FileMode) error {
	return EnsureParentDir(path, perm)
}

// EnsureParentDir creates the parent directory for path with perm.
func EnsureParentDir(path string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return CreateDirWithPerm(dir, perm)
}
