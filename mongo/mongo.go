package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/weiweimhy/go-utils/logger"

	"sync"
)

type Config struct {
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	OPTimeout      time.Duration `yaml:"op_timeout"`
	Uri            string        `yaml:"uri"`
	AuthName       string        `yaml:"auth_name"`
	AuthPass       string        `yaml:"auth_pass"`
	AuthDatabase   string        `yaml:"auth_database"`
	DatabaseName   string        `yaml:"database_name"`
}

type DB struct {
	*mongo.Database
	Config
}

var (
	database *DB
	once     sync.Once
)

func GetMongoDB(config Config) *DB {
	defer logger.Trace(logger.L(), "mongo.GetMongoDB")()

	once.Do(func() {
		opts := options.Client()
		opts.SetConnectTimeout(config.ConnectTimeout)
		opts.ApplyURI(config.Uri)

		if config.AuthName != "" && config.AuthPass != "" {
			credential := options.Credential{
				Username:   config.AuthName,
				Password:   config.AuthPass,
				AuthSource: config.AuthDatabase,
			}
			opts.SetAuth(credential)
		}

		client, err := mongo.Connect(opts)
		if err != nil {
			logger.L().Fatal("failed to connect to MongoDB",
				zap.String("uri", config.Uri),
				zap.Error(err),
			)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		defer cancel()

		err = client.Ping(ctx, nil)
		if err != nil {
			logger.L().Fatal("failed to ping MongoDB",
				zap.String("uri", config.Uri),
				zap.Error(err),
			)
			os.Exit(1)
		}

		database = &DB{
			client.Database(config.DatabaseName),
			config,
		}
	})

	return database
}

func InsertOne[T any](db *DB, collectionName string, document T) (*mongo.InsertOneResult, error) {
	defer logger.Trace(logger.L(), "mongo.InsertOne", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("insert one failed: %w", err)
	}
	return result, nil
}

func InsertMany[T any](db *DB, collectionName string, documents []T) (*mongo.InsertManyResult, error) {
	defer logger.Trace(logger.L(), "mongo.InsertMany", zap.String("collection", collectionName), zap.Int("count", len(documents)))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	if len(documents) == 0 {
		return nil, logger.InvalidParam(logger.L(), "documents is empty", zap.String("param", "documents"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)

	docs := make([]interface{}, len(documents))
	for index, doc := range documents {
		docs[index] = doc
	}
	result, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("insert many failed: %w", err)
	}
	return result, nil
}

func DeleteOne(db *DB, collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	defer logger.Trace(logger.L(), "mongo.DeleteOne", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("delete one failed: %w", err)
	}
	return result, nil
}

func DeleteMany(db *DB, collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	defer logger.Trace(logger.L(), "mongo.DeleteMany", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("delete many failed: %w", err)
	}
	return result, nil
}

func UpdateOne[T any](db *DB, collectionName string, filter interface{}, update T) (*mongo.UpdateResult, error) {
	defer logger.Trace(logger.L(), "mongo.UpdateOne", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("update one failed: %w", err)
	}
	return result, nil
}

func UpdateMany[T any](db *DB, collectionName string, filter interface{}, updates []T) (*mongo.UpdateResult, error) {
	defer logger.Trace(logger.L(), "mongo.UpdateMany", zap.String("collection", collectionName), zap.Int("count", len(updates)))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)

	var updateDocs = make([]interface{}, len(updates))
	for index, update := range updates {
		updateDocs[index] = update
	}
	result, err := collection.UpdateMany(ctx, filter, updateDocs)
	if err != nil {
		return nil, fmt.Errorf("update many failed: %w", err)
	}
	return result, nil
}

func ReplaceOne[T any](db *DB, collectionName string, filter interface{}, update T) (*mongo.UpdateResult, error) {
	defer logger.Trace(logger.L(), "mongo.ReplaceOne", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	result, err := collection.ReplaceOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("replace one failed: %w", err)
	}
	return result, nil
}

func FindOne[T any](db *DB, collectionName string, filter interface{}) (*T, error) {
	defer logger.Trace(logger.L(), "mongo.FindOne", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	singleResult := collection.FindOne(ctx, filter)

	var result T
	err := singleResult.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("find one decode failed: %w", err)
	}

	return &result, nil
}

func FindMany[T any](db *DB, collectionName string, filter interface{}) ([]*T, error) {
	defer logger.Trace(logger.L(), "mongo.FindMany", zap.String("collection", collectionName))()

	if db == nil {
		return nil, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find many failed: %w", err)
	}
	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			logger.L().Warn("failed to close cursor",
				zap.String("collection", collectionName),
				zap.Error(err),
			)
		}
	}(cur, ctx)

	var result []*T
	for cur.Next(ctx) {
		var elem T
		err := cur.Decode(&elem)
		if err != nil {
			return nil, fmt.Errorf("find many decode failed: %w", err)
		}
		result = append(result, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("find many cursor error: %w", err)
	}

	return result, nil
}

func Count(db *DB, collectionName string, filter interface{}) (int64, error) {
	defer logger.Trace(logger.L(), "mongo.Count", zap.String("collection", collectionName))()

	if db == nil {
		return 0, logger.InvalidParam(logger.L(), "database is nil", zap.String("param", "db"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.OPTimeout)
	defer cancel()

	collection := db.Collection(collectionName)
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count documents failed: %w", err)
	}

	return count, nil
}
