package fsutil

import (
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// CreateParentDir creates the parent directory for the given file path.
func CreateParentDir(path string) error {
	return CreateDir(filepath.Dir(path))
}
