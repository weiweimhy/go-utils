package mongo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/weiweimhy/go-utils/v3/logger"
)

// MockClient 实现了 IMongoClient 接口，用于测试。
type MockClient struct {
	IMongoClient
}

func (m *MockClient) InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	return nil, nil
}

func TestIMongoClientInterface(t *testing.T) {
	// 验证 MockClient 是否满足接口
	var _ IMongoClient = (*MockClient)(nil)
}

func TestConfig(t *testing.T) {
	logger.Init()
	cfg := Config{
		Uri:            "mongodb://localhost:27017",
		DatabaseName:   "testdb",
		ConnectTimeout: 5 * time.Second,
		OPTimeout:      5 * time.Second,
	}

	if cfg.Uri != "mongodb://localhost:27017" {
		t.Errorf("expected uri mongodb://localhost:27017, got %s", cfg.Uri)
	}
}
