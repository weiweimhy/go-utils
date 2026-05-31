package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/weiweimhy/go-utils/v5/cryptoutil"
)

const (
	// DefaultFilePerm is used by general-purpose write helpers.
	DefaultFilePerm os.FileMode = 0644
	// SecureFilePerm is used by helpers intended for private data.
	SecureFilePerm os.FileMode = 0600
)

// WriteOptions controls how WriteFile persists data.
type WriteOptions struct {
	DirPerm    os.FileMode
	FilePerm   os.FileMode
	Atomic     bool
	Sync       bool
	CreateDirs bool
}

func (opts WriteOptions) withDefaults() WriteOptions {
	if opts.DirPerm == 0 {
		opts.DirPerm = DefaultDirPerm
	}
	if opts.FilePerm == 0 {
		opts.FilePerm = DefaultFilePerm
	}
	return opts
}

// FileExists reports whether the given path exists and is a file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// SaveToFile 保存数据到文件，默认使用 0644 权限。
// 如果父目录不存在，会自动创建。
func SaveToFile(path string, data []byte) error {
	return WriteFile(path, data, WriteOptions{CreateDirs: true})
}

// SaveToFileWithPerm 保存数据到文件，并指定权限。
func SaveToFileWithPerm(path string, data []byte, perm os.FileMode) error {
	return WriteFile(path, data, WriteOptions{FilePerm: perm, CreateDirs: true})
}

// WriteFile writes data to path according to opts.
func WriteFile(path string, data []byte, opts WriteOptions) error {
	if path == "" {
		return fmt.Errorf("fsutil: path is required")
	}
	opts = opts.withDefaults()

	if opts.CreateDirs {
		if err := EnsureParentDir(path, opts.DirPerm); err != nil {
			return err
		}
	}

	if opts.Atomic {
		return writeFileAtomic(path, data, opts)
	}
	return writeFileDirect(path, data, opts)
}

// SecureWriteFile atomically writes private data using 0700 parent directories and 0600 files.
func SecureWriteFile(path string, data []byte) error {
	return WriteFile(path, data, WriteOptions{
		DirPerm:    SecureDirPerm,
		FilePerm:   SecureFilePerm,
		Atomic:     true,
		Sync:       true,
		CreateDirs: true,
	})
}

func writeFileDirect(path string, data []byte, opts WriteOptions) error {
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(path, flag, opts.FilePerm)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()

	if _, err := file.Write(data); err != nil {
		return err
	}
	if opts.Sync {
		if err := file.Sync(); err != nil {
			return err
		}
	}
	if err := file.Chmod(opts.FilePerm); err != nil {
		return err
	}
	closed = true
	return file.Close()
}

func writeFileAtomic(path string, data []byte, opts WriteOptions) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if dir == "." {
		dir = "."
	}

	tmp, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	keepTemp := false
	defer func() {
		_ = tmp.Close()
		if !keepTemp {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmp.Chmod(opts.FilePerm); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if opts.Sync {
		if err := tmp.Sync(); err != nil {
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	keepTemp = true

	if opts.Sync {
		_ = syncDir(filepath.Dir(path))
	}
	return nil
}

func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func GetFileBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return cryptoutil.Base64FromBytes(data), nil
}

// GetStringFromFile reads a file into a string.
func GetStringFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetNameWithPathWithoutExt
//   - 包含完整路径
//   - 不包含文件后缀
//   - 示例："/home/user/docs/report.pdf" → "/home/user/docs/report"
func GetNameWithPathWithoutExt(fullPath string) string {
	// 1. 取目录 + 文件名（不含后缀）
	dir := filepath.Dir(fullPath)
	base := filepath.Base(fullPath)

	// 2. 去掉后缀
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

	// 3. 拼接：目录 + 文件名（无后缀）
	if dir == "." {
		return nameWithoutExt
	}
	return filepath.Join(dir, nameWithoutExt)
}
