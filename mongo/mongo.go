package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Config contains MongoDB connection settings.
type Config struct {
	ConnectTimeout   time.Duration `yaml:"connect_timeout"`
	OperationTimeout time.Duration `yaml:"operation_timeout"`
	URI              string        `yaml:"uri"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	AuthDatabase     string        `yaml:"auth_database"`
	DatabaseName     string        `yaml:"database_name"`
}

// DefaultConfig returns a config with package defaults applied.
func DefaultConfig() Config {
	return Config{
		ConnectTimeout:   10 * time.Second,
		OperationTimeout: 5 * time.Second,
	}
}

// Client wraps a MongoDB database handle with package-level timeout defaults.
type Client struct {
	*driver.Database
	cfg Config
}

// NewClient creates a MongoDB client with package-level timeout defaults.
// Zero ConnectTimeout or OperationTimeout values are replaced with defaults.
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 应用默认值
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.OperationTimeout == 0 {
		config.OperationTimeout = 5 * time.Second
	}

	opts := options.Client().
		SetConnectTimeout(config.ConnectTimeout).
		ApplyURI(config.URI)

	if config.Username != "" && config.Password != "" {
		opts.SetAuth(options.Credential{
			Username:   config.Username,
			Password:   config.Password,
			AuthSource: config.AuthDatabase,
		})
	}

	client, err := driver.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		defer disconnectCancel()
		pingErr := fmt.Errorf("failed to ping MongoDB: %w", err)
		if disconnectErr := client.Disconnect(disconnectCtx); disconnectErr != nil {
			return nil, errors.Join(pingErr, fmt.Errorf("failed to disconnect MongoDB client after ping failure: %w", disconnectErr))
		}
		return nil, pingErr
	}

	return &Client{
		Database: client.Database(config.DatabaseName),
		cfg:      config,
	}, nil
}

// InsertOne inserts a single document into the named collection.
func (c *Client) InsertOne(ctx context.Context, collectionName string, document interface{}) (*driver.InsertOneResult, error) {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).InsertOne(opCtx, document)
}

// InsertMany inserts multiple documents into the named collection.
func (c *Client) InsertMany(ctx context.Context, collectionName string, documents []interface{}) (*driver.InsertManyResult, error) {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).InsertMany(opCtx, documents)
}

// DeleteOne deletes a single document from the named collection.
func (c *Client) DeleteOne(ctx context.Context, collectionName string, filter interface{}) (*driver.DeleteResult, error) {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).DeleteOne(opCtx, filter)
}

// UpdateOne updates a single document in the named collection.
func (c *Client) UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface{}) (*driver.UpdateResult, error) {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).UpdateOne(opCtx, filter, update)
}

// FindOne finds a single document in the named collection and decodes it into result.
func (c *Client) FindOne(ctx context.Context, collectionName string, filter interface{}, result interface{}) error {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).FindOne(opCtx, filter).Decode(result)
}

// FindMany finds multiple documents in the named collection and decodes them into results.
func (c *Client) FindMany(ctx context.Context, collectionName string, filter interface{}, results interface{}) error {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	cur, err := c.Collection(collectionName).Find(opCtx, filter)
	if err != nil {
		return err
	}
	return cur.All(opCtx, results)
}

// Count counts documents in the named collection matching the filter.
func (c *Client) Count(ctx context.Context, collectionName string, filter interface{}) (int64, error) {
	opCtx, cancel := c.operationContext(ctx)
	defer cancel()
	return c.Collection(collectionName).CountDocuments(opCtx, filter)
}

// Disconnect closes the underlying MongoDB client.
func (c *Client) Disconnect(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return c.Client().Disconnect(ctx)
}

func (c *Client) operationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, c.cfg.OperationTimeout)
}
