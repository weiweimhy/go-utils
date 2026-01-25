package localDB

import (
	"path/filepath"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/weiweimhy/go-utils/fsutil"
	"go.etcd.io/bbolt"
)

type LocalDB struct {
	*bbolt.DB
}

// Open 打开一个新的 BoltDB 实例。不再使用全局变量存储。
func Open(path string, name string) (*LocalDB, error) {
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

	return &LocalDB{db}, nil
}

func (db *LocalDB) Set(bucket, key string, value []byte) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (db *LocalDB) SetJSON(bucket, key string, v interface{}) error {
	data, err := sonic.Marshal(v)
	if err != nil {
		return err
	}
	return db.Set(bucket, key, data)
}

func (db *LocalDB) Get(bucket, key string) ([]byte, error) {
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

func (db *LocalDB) GetJSON(bucket, key string, out interface{}) error {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return err
	}
	return sonic.Unmarshal(data, out)
}

func (db *LocalDB) SetInt(bucket, key string, value int) error {
	return db.Set(bucket, key, []byte(strconv.Itoa(value)))
}

func (db *LocalDB) GetInt(bucket, key string) (int, error) {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func (db *LocalDB) Delete(bucket, key string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
}
