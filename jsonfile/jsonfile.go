package jsonfile

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/weiweimhy/go-utils/v6/fsutil"
)

// Options controls JSON file load and save behavior.
type Options struct {
	DirPerm    os.FileMode
	FilePerm   os.FileMode
	Indent     string
	MissingOK  bool
	AtomicSave bool
	Sync       bool
}

func (opts Options) withDefaults() Options {
	if opts.DirPerm == 0 {
		opts.DirPerm = fsutil.DefaultDirPerm
	}
	if opts.FilePerm == 0 {
		opts.FilePerm = fsutil.DefaultFilePerm
	}
	if opts.Indent == "" {
		opts.Indent = "  "
	}
	return opts
}

// Load reads path as JSON into T. If the file is missing and MissingOK is true, fallback is returned.
func Load[T any](path string, fallback T, opts Options) (T, error) {
	opts = opts.withDefaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && opts.MissingOK {
			return fallback, nil
		}
		return fallback, err
	}

	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return fallback, err
	}
	return out, nil
}

// Save marshals value as JSON and writes it to path.
func Save[T any](path string, value T, opts Options) error {
	opts = opts.withDefaults()

	var (
		data []byte
		err  error
	)
	if opts.Indent != "" {
		data, err = json.MarshalIndent(value, "", opts.Indent)
	} else {
		data, err = json.Marshal(value)
	}
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return fsutil.WriteFile(path, data, fsutil.WriteOptions{
		DirPerm:    opts.DirPerm,
		FilePerm:   opts.FilePerm,
		Atomic:     opts.AtomicSave,
		Sync:       opts.Sync,
		CreateDirs: true,
	})
}
