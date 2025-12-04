package localDB

import (
	"strconv"
	"time"

	"invoiceClient/pkg/customUtils"
	"path/filepath"
	"sync"

	"github.com/bytedance/sonic"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LocalDB struct {
	*bbolt.DB
}

var (
	db   *LocalDB
	once sync.Once
)

// DB 返回通过 InitLocalDB 初始化的全局 LocalDB 实例
func DB() *LocalDB {
	return db
}

func InitLocalDB(path string, name string) {
	once.Do(func() {
		err := customUtils.CreateDir(path)

		if err != nil {
			zap.L().Log(
				zapcore.ErrorLevel,
				"create local db dir error",
				zap.Error(err),
				zap.String("path", path),
			)
			return
		}

		path = filepath.Join(path, name)

		options := *bbolt.DefaultOptions
		options.NoSync = false
		options.Timeout = 5 * time.Second

		boltDB, err := bbolt.Open(path, 0664, &options)
		if err != nil {
			zap.L().Log(
				zapcore.ErrorLevel,
				"open local db dir error",
				zap.Error(err),
				zap.String("path", path),
			)
			return
		}

		db = &LocalDB{boltDB}
	})
}

func (db *LocalDB) Set(bucket, key string, value []byte) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := db.getOrCreateBucket(tx, bucket)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (db *LocalDB) SetInt(bucket, key string, value int) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := db.getOrCreateBucket(tx, bucket)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(strconv.Itoa(value)))
	})
}

// SetString 存储字符串
func (db *LocalDB) SetString(bucket, key, value string) error {
	return db.Set(bucket, key, []byte(value))
}

// SetBool 存储布尔值（使用 "1"/"0" 表示）
func (db *LocalDB) SetBool(bucket, key string, value bool) error {
	if value {
		return db.Set(bucket, key, []byte("1"))
	}
	return db.Set(bucket, key, []byte("0"))
}

// SetJSON 将任意结构体 / map 序列化为 JSON 存储
func (db *LocalDB) SetJSON(bucket, key string, v any) error {
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
			data = nil
			return nil
		}
		data = b.Get([]byte(key))
		return nil
	})
	return data, err
}

// GetString 获取字符串
func (db *LocalDB) GetString(bucket, key string) (string, error) {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

// GetInt 获取 int
func (db *LocalDB) GetInt(bucket, key string) (int, error) {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// GetBool 获取布尔值（"1"=true，其它/空=false）
func (db *LocalDB) GetBool(bucket, key string) (bool, error) {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return false, err
	}
	return string(data) == "1", nil
}

// GetJSON 从 JSON 反序列化到 out（out 必须是指针）
func (db *LocalDB) GetJSON(bucket, key string, out any) error {
	data, err := db.Get(bucket, key)
	if err != nil || data == nil {
		return err
	}
	return sonic.Unmarshal(data, out)
}

func (db *LocalDB) Delete(bucket, key string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			// bucket 不存在，视为已删除（幂等）
			return nil
		}
		return b.Delete([]byte(key))
	})
}

// Close 关闭底层 bbolt.DB
func (db *LocalDB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// getOrCreateBucket 内部工具：确保 bucket 存在
func (db *LocalDB) getOrCreateBucket(tx *bbolt.Tx, bucket string) (*bbolt.Bucket, error) {
	b := tx.Bucket([]byte(bucket))
	if b != nil {
		return b, nil
	}
	return tx.CreateBucketIfNotExists([]byte(bucket))
}
