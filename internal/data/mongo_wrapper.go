package data

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jollfi-gaming-api/internal/interfaces"
)

// MongoCursor wraps mongo.Cursor to implement MongoCursorInterface
type MongoCursor struct {
	cursor *mongo.Cursor
}

// MongoCollection wraps mongo.Collection to implement MongoCollectionInterface
type MongoCollection struct {
	collection *mongo.Collection
}

// MongoDatabase wraps mongo.Database to implement MongoDatabaseInterface
type MongoDatabase struct {
	database *mongo.Database
}

// MongoCursor implementations
func (mc *MongoCursor) Next(ctx context.Context) bool {
	return mc.cursor.Next(ctx)
}

func (mc *MongoCursor) Decode(v interface{}) error {
	return mc.cursor.Decode(v)
}

func (mc *MongoCursor) All(ctx context.Context, results interface{}) error {
	return mc.cursor.All(ctx, results)
}

func (mc *MongoCursor) Close(ctx context.Context) error {
	return mc.cursor.Close(ctx)
}

// MongoCollection implementations
func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return mc.collection.InsertOne(ctx, document)
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (interfaces.MongoCursorInterface, error) {
	cursor, err := mc.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursor{cursor: cursor}, nil
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{}) interface{} {
	return mc.collection.FindOne(ctx, filter)
}

func (mc *MongoCollection) UpdateOne(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error) {
	return mc.collection.UpdateOne(ctx, filter, update)
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return mc.collection.DeleteOne(ctx, filter)
}

// MongoDatabase implementations
func (md *MongoDatabase) Collection(name string) interfaces.MongoCollectionInterface {
	collection := md.database.Collection(name)
	return &MongoCollection{collection: collection}
}
