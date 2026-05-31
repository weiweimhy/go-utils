package filemeta

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Metadata describes a file.
type Metadata struct {
	Path      string
	SizeBytes int64
	SHA256    string
	ModTime   time.Time
}

// Options controls metadata collection.
type Options struct {
	ComputeSHA256 bool
}

// Get returns metadata for path. SHA-256 is calculated only when requested.
func Get(path string, opts Options) (Metadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Metadata{}, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return Metadata{}, err
	}
	meta := Metadata{
		Path:      abs,
		SizeBytes: info.Size(),
		ModTime:   info.ModTime(),
	}
	if opts.ComputeSHA256 {
		sum, err := SHA256(path)
		if err != nil {
			return Metadata{}, err
		}
		meta.SHA256 = sum
	}
	return meta, nil
}

// SHA256 returns the SHA-256 checksum of a file as hex.
func SHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
