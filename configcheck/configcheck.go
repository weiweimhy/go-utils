package configcheck

import (
	"errors"
	"os"
	"runtime"
)

var (
	// ErrNotRegular is returned when the checked path is not a regular file.
	ErrNotRegular = errors.New("configcheck: path is not a regular file")
	// ErrWorldWritable is returned on Unix-like systems for files writable by everyone.
	ErrWorldWritable = errors.New("configcheck: file is world-writable")
)

// Result describes the outcome of a file safety check.
type Result struct {
	Path          string
	Mode          os.FileMode
	Regular       bool
	WorldWritable bool
}

// CheckFile returns basic safety information for a config file.
func CheckFile(path string) (Result, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Path:    path,
		Mode:    info.Mode().Perm(),
		Regular: info.Mode().IsRegular(),
	}
	if !result.Regular {
		return result, ErrNotRegular
	}
	if runtime.GOOS != "windows" && result.Mode&0002 != 0 {
		result.WorldWritable = true
		return result, ErrWorldWritable
	}
	return result, nil
}
