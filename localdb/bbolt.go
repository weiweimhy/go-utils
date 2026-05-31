package localdb

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"time"

	"github.com/weiweimhy/go-utils/v5/fsutil"
	"go.etcd.io/bbolt"
)

// DB is a thin BoltDB-backed local key-value store.
type DB struct {
	*bbolt.DB
}

// Open opens a BoltDB-backed local key-value store.
func Open(path string, name string) (*DB, error) {
	if err := fsutil.CreateDir(path); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(path, name)
	opts := *bbolt.DefaultOptions
	opts.Timeout = 5 * time.Second

	db, err := bbolt.Open(fullPath, 0664, &opts)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// Set stores a raw byte value in the given bucket and key.
func (db *DB) Set(bucket, key string, value []byte) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

// SetJSONValue marshals v as JSON and stores it in the given bucket and key.
func (db *DB) SetJSONValue(bucket, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.Set(bucket, key, data)
}

// Get loads a raw byte value from the given bucket and key.
func (db *DB) Get(bucket, key string) ([]byte, error) {
	var data []byte
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		val := b.Get([]byte(key))
		if val != nil {
			data = make([]byte, len(val))
			copy(data, val)
		}
		return nil
	})
	return data, err
}

// GetJSONValue loads JSON data from the given bucket and key into out.
func (db *DB) GetJSONValue(bucket, key string, out interface{}) error {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// SetInt stores an int value in decimal string form.
func (db *DB) SetInt(bucket, key string, value int) error {
	return db.Set(bucket, key, []byte(strconv.Itoa(value)))
}

// GetInt loads an int value stored in decimal string form.
func (db *DB) GetInt(bucket, key string) (int, error) {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// Delete removes a key from the given bucket.
func (db *DB) Delete(bucket, key string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}
