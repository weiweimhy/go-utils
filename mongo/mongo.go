package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/weiweimhy/go-utils/v3/logger"
)

// IMongoClient 定义了数据库操作的标准接口，方便外部 Mock 测试。
type IMongoClient interface {
	// InsertOne 插入单个文档。
	InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error)
	// InsertMany 批量插入文档。
	InsertMany(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error)
	// DeleteOne 删除单个匹配的文档。
	DeleteOne(ctx context.Context, collectionName string, filter interface{}) (*mongo.DeleteResult, error)
	// UpdateOne 更新单个匹配的文档。
	UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
	// FindOne 查询单个文档。
	FindOne(ctx context.Context, collectionName string, filter interface{}, result interface{}) error
	// FindMany 查询多个文档。
	FindMany(ctx context.Context, collectionName string, filter interface{}, results interface{}) error
	// Count 统计文档数量。
	Count(ctx context.Context, collectionName string, filter interface{}) (int64, error)
	// Disconnect 关闭数据库连接。
	Disconnect(ctx context.Context) error
}

// Config 包含 MongoDB 的连接配置。
type Config struct {
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	OPTimeout      time.Duration `yaml:"op_timeout"`
	Uri            string        `yaml:"uri"`
	AuthName       string        `yaml:"auth_name"`
	AuthPass       string        `yaml:"auth_pass"`
	AuthDatabase   string        `yaml:"auth_database"`
	DatabaseName   string        `yaml:"database_name"`
}

// DefaultConfig 返回带有合理默认值的配置
func DefaultConfig() Config {
	return Config{
		ConnectTimeout: 10 * time.Second,
		OPTimeout:      5 * time.Second,
	}
}

type clientImpl struct {
	*mongo.Database
	cfg Config
}

// NewClient 创建并返回满足 IMongoClient 接口的数据库实例。
// 如果 config.ConnectTimeout 或 config.OPTimeout 为零，将使用默认值。
func NewClient(ctx context.Context, config Config) (IMongoClient, error) {
	defer logger.Trace(logger.L(), "mongo.NewClient")()

	// 应用默认值
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.OPTimeout == 0 {
		config.OPTimeout = 5 * time.Second
	}

	opts := options.Client().
		SetConnectTimeout(config.ConnectTimeout).
		ApplyURI(config.Uri)

	if config.AuthName != "" && config.AuthPass != "" {
		opts.SetAuth(options.Credential{
			Username:   config.AuthName,
			Password:   config.AuthPass,
			AuthSource: config.AuthDatabase,
		})
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &clientImpl{
		Database: client.Database(config.DatabaseName),
		cfg:      config,
	}, nil
}

func (c *clientImpl) InsertOne(ctx context.Context, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	defer logger.Trace(logger.L(), "mongo.InsertOne", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).InsertOne(opCtx, document)
}

func (c *clientImpl) InsertMany(ctx context.Context, collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	defer logger.Trace(logger.L(), "mongo.InsertMany", zap.String("collection", collectionName), zap.Int("count", len(documents)))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).InsertMany(opCtx, documents)
}

func (c *clientImpl) DeleteOne(ctx context.Context, collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	defer logger.Trace(logger.L(), "mongo.DeleteOne", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).DeleteOne(opCtx, filter)
}

func (c *clientImpl) UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	defer logger.Trace(logger.L(), "mongo.UpdateOne", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).UpdateOne(opCtx, filter, update)
}

func (c *clientImpl) FindOne(ctx context.Context, collectionName string, filter interface{}, result interface{}) error {
	defer logger.Trace(logger.L(), "mongo.FindOne", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).FindOne(opCtx, filter).Decode(result)
}

func (c *clientImpl) FindMany(ctx context.Context, collectionName string, filter interface{}, results interface{}) error {
	defer logger.Trace(logger.L(), "mongo.FindMany", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	cur, err := c.Collection(collectionName).Find(opCtx, filter)
	if err != nil {
		return err
	}
	return cur.All(opCtx, results)
}

func (c *clientImpl) Count(ctx context.Context, collectionName string, filter interface{}) (int64, error) {
	defer logger.Trace(logger.L(), "mongo.Count", zap.String("collection", collectionName))()
	opCtx, cancel := context.WithTimeout(ctx, c.cfg.OPTimeout)
	defer cancel()
	return c.Collection(collectionName).CountDocuments(opCtx, filter)
}

func (c *clientImpl) Disconnect(ctx context.Context) error {
	return c.Client().Disconnect(ctx)
}
