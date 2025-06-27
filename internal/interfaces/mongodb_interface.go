package interfaces

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCursorInterface defines the interfaces for MongoDB cursors
type MongoCursorInterface interface {
	Next(ctx context.Context) bool
	Decode(v interface{}) error
	All(ctx context.Context, results interface{}) error
	Close(ctx context.Context) error
}

// MongoCollectionInterface defines the interfaces for MongoDB collections
type MongoCollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (MongoCursorInterface, error)
	FindOne(ctx context.Context, filter interface{}) interface{}
	UpdateOne(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)
}

// MongoDatabaseInterface defines the interfaces for MongoDB databases
type MongoDatabaseInterface interface {
	Collection(name string) MongoCollectionInterface
}

// MongoClientInterface defines the interfaces for MongoDB clients
type MongoClientInterface interface {
	GetDatabase(name string) MongoDatabaseInterface
	Close()
	Ping(ctx context.Context) error

	// Game operations
	CreateGame(ctx context.Context, game interface{}) (string, error)
	GetGame(ctx context.Context, gameID string) (interface{}, error)
	UpdateGame(ctx context.Context, gameID string, updates interface{}) error
	GetGamesByStatus(ctx context.Context, status string) ([]interface{}, error)
	GetGamesByAddress(ctx context.Context, address string) ([]interface{}, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction interface{}) (string, error)
	GetTransactionsByGameID(ctx context.Context, gameID string) ([]interface{}, error)
}
