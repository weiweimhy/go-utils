package auditlog

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/weiweimhy/go-utils/v5/fsutil"
)

// RedactFunc can sanitize a record before it is written.
type RedactFunc func(any) any

// Writer appends JSON Lines audit records to a file.
type Writer struct {
	mu     sync.Mutex
	file   *os.File
	redact RedactFunc
}

// Options controls Writer creation.
type Options struct {
	DirPerm  os.FileMode
	FilePerm os.FileMode
	Redact   RedactFunc
}

// Open opens a JSONL audit writer in append mode.
func Open(path string, opts Options) (*Writer, error) {
	dirPerm := opts.DirPerm
	if dirPerm == 0 {
		dirPerm = fsutil.SecureDirPerm
	}
	filePerm := opts.FilePerm
	if filePerm == 0 {
		filePerm = fsutil.SecureFilePerm
	}
	if err := fsutil.EnsureParentDir(path, dirPerm); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(filePerm); err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Writer{file: file, redact: opts.Redact}, nil
}

// Write appends one JSON record and a trailing newline.
func (w *Writer) Write(record any) error {
	if w == nil {
		return os.ErrClosed
	}
	if w.redact != nil {
		record = w.redact(record)
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return os.ErrClosed
	}
	_, err = w.file.Write(data)
	return err
}

// Sync flushes buffered filesystem state.
func (w *Writer) Sync() error {
	if w == nil {
		return os.ErrClosed
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return os.ErrClosed
	}
	return w.file.Sync()
}

// Close closes the writer.
func (w *Writer) Close() error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
