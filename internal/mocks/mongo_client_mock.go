package mocks

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jollfi-gaming-api/internal/interfaces"
)

var _ interfaces.MongoClientInterface = (*MockMongoClient)(nil)
var _ interfaces.MongoDatabaseInterface = (*MockDatabase)(nil)
var _ interfaces.MongoCollectionInterface = (*MockCollection)(nil)
var _ interfaces.MongoCursorInterface = (*MockCursor)(nil)

type MockMongoClient struct {
	databases    map[string]*MockDatabase
	closed       bool
	games        map[string]interface{}
	transactions map[string]interface{}
	users        map[string]interface{}
}

type MockDatabase struct {
	Name        string
	collections map[string]*MockCollection
}

type MockCollection struct {
	Name      string
	documents []interface{}
}

type MockCursor struct {
	documents []interface{}
	position  int
}

// MockCursor implementations
func (c *MockCursor) Next(ctx context.Context) bool {
	if c.position < len(c.documents) {
		return true
	}
	return false
}

func (c *MockCursor) Decode(v interface{}) error {
	if c.position >= len(c.documents) {
		return fmt.Errorf("no more documents")
	}
	// Move to next position after decode
	c.position++
	return nil
}

func (c *MockCursor) All(ctx context.Context, results interface{}) error {
	// For testing, we'll return empty results
	return nil
}

func (c *MockCursor) Close(ctx context.Context) error {
	return nil
}

// NewMockMongoClient creates a new mock MongoDB client
func NewMockMongoClient() *MockMongoClient {
	return &MockMongoClient{
		databases:    make(map[string]*MockDatabase),
		closed:       false,
		games:        make(map[string]interface{}),
		transactions: make(map[string]interface{}),
		users:        make(map[string]interface{}),
	}
}

// Basic MongoDB client methods
func (m *MockMongoClient) GetDatabase(name string) interfaces.MongoDatabaseInterface {
	if m.closed {
		return nil
	}
	if db, exists := m.databases[name]; exists {
		return db
	}
	db := &MockDatabase{
		Name:        name,
		collections: make(map[string]*MockCollection),
	}
	m.databases[name] = db
	return db
}

func (m *MockMongoClient) Close() {
	m.closed = true
	m.databases = make(map[string]*MockDatabase)
}

func (m *MockMongoClient) Ping(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}
	return nil
}

// Game-related methods
func (m *MockMongoClient) CreateGame(ctx context.Context, game interface{}) (string, error) {
	if m.closed {
		return "", fmt.Errorf("client is closed")
	}

	gameID := primitive.NewObjectID().Hex()
	m.games[gameID] = game
	return gameID, nil
}

func (m *MockMongoClient) GetGame(ctx context.Context, gameID string) (interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	if game, exists := m.games[gameID]; exists {
		return game, nil
	}
	return nil, fmt.Errorf("game not found")
}

func (m *MockMongoClient) UpdateGame(ctx context.Context, gameID string, updates interface{}) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	if _, exists := m.games[gameID]; !exists {
		return fmt.Errorf("game not found")
	}

	// In a real mock, you'd merge the updates with existing game
	m.games[gameID] = updates
	return nil
}

func (m *MockMongoClient) GetGamesByStatus(ctx context.Context, status string) ([]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	var games []interface{}
	for _, game := range m.games {
		// In a real implementation, you'd check the status field
		games = append(games, game)
	}
	return games, nil
}

func (m *MockMongoClient) GetGamesByAddress(ctx context.Context, address string) ([]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	var games []interface{}
	for _, game := range m.games {
		// In a real implementation, you'd check address fields
		games = append(games, game)
	}
	return games, nil
}

// Transaction-related methods
func (m *MockMongoClient) CreateTransaction(ctx context.Context, transaction interface{}) (string, error) {
	if m.closed {
		return "", fmt.Errorf("client is closed")
	}

	txID := primitive.NewObjectID().Hex()
	m.transactions[txID] = transaction
	return txID, nil
}

func (m *MockMongoClient) GetTransactionsByGameID(ctx context.Context, gameID string) ([]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	var transactions []interface{}
	for _, tx := range m.transactions {
		// In a real implementation, you'd filter by gameID
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (m *MockMongoClient) UpdateTransactionStatus(ctx context.Context, txDigest string, status string, blockHeight *uint64) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	// Mock implementation - in reality you'd find and update the transaction
	return nil
}

func (m *MockMongoClient) GetPendingTransactions(ctx context.Context) ([]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	var pendingTxs []interface{}
	for _, tx := range m.transactions {
		// In a real implementation, you'd filter by status
		pendingTxs = append(pendingTxs, tx)
	}
	return pendingTxs, nil
}

// User-related methods
func (m *MockMongoClient) CreateUser(ctx context.Context, user interface{}) (string, error) {
	if m.closed {
		return "", fmt.Errorf("client is closed")
	}

	userID := primitive.NewObjectID().Hex()
	m.users[userID] = user
	return userID, nil
}

func (m *MockMongoClient) GetUser(ctx context.Context, address string) (interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	for _, user := range m.users {
		// In a real implementation, you'd check the address field
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockMongoClient) UpdateUser(ctx context.Context, address string, updates interface{}) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	// Mock implementation
	return nil
}

func (m *MockMongoClient) UpdateUserLastSeen(ctx context.Context, address string) error {
	if m.closed {
		return fmt.Errorf("client is closed")
	}

	// Mock implementation
	return nil
}

// Statistics and utility methods
func (m *MockMongoClient) GetActiveGames(ctx context.Context) ([]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	var activeGames []interface{}
	for _, game := range m.games {
		activeGames = append(activeGames, game)
	}
	return activeGames, nil
}

func (m *MockMongoClient) GetGameStats(ctx context.Context) (map[string]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	stats := map[string]interface{}{
		"total_games": len(m.games),
		"total_stake": uint64(0),
		"completed": map[string]interface{}{
			"count":       0,
			"total_stake": uint64(0),
		},
	}
	return stats, nil
}

func (m *MockMongoClient) GetUserStats(ctx context.Context, address string) (map[string]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	stats := map[string]interface{}{
		"total_games": 0,
		"total_wins":  0,
		"total_stake": uint64(0),
		"win_rate":    0.0,
	}
	return stats, nil
}

func (m *MockMongoClient) GetCollectionStats(ctx context.Context) (map[string]interface{}, error) {
	if m.closed {
		return nil, fmt.Errorf("client is closed")
	}

	stats := map[string]interface{}{
		"games":        len(m.games),
		"users":        len(m.users),
		"transactions": len(m.transactions),
	}
	return stats, nil
}

func (m *MockMongoClient) HealthCheck(ctx context.Context) error {
	return m.Ping(ctx)
}

func (m *MockMongoClient) CleanupOldGames(ctx context.Context, olderThanDays int) (int64, error) {
	if m.closed {
		return 0, fmt.Errorf("client is closed")
	}

	// Mock implementation - return 0 deleted
	return 0, nil
}

// MockDatabase implementations
func (d *MockDatabase) Collection(name string) interfaces.MongoCollectionInterface {
	if coll, exists := d.collections[name]; exists {
		return coll
	}
	coll := &MockCollection{
		Name:      name,
		documents: make([]interface{}, 0),
	}
	d.collections[name] = coll
	return coll
}

// MockCollection implementations
func (c *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	c.documents = append(c.documents, document)
	return &mongo.InsertOneResult{
		InsertedID: primitive.NewObjectID(),
	}, nil
}

func (c *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (interfaces.MongoCursorInterface, error) {
	return &MockCursor{
		documents: c.documents,
		position:  0,
	}, nil
}

func (c *MockCollection) FindOne(ctx context.Context, filter interface{}) interface{} {
	if len(c.documents) > 0 {
		return c.documents[0]
	}
	return nil
}

func (c *MockCollection) UpdateOne(ctx context.Context, filter, update interface{}) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
	}, nil
}

func (c *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	if len(c.documents) > 0 {
		c.documents = c.documents[1:]
		return &mongo.DeleteResult{DeletedCount: 1}, nil
	}
	return &mongo.DeleteResult{DeletedCount: 0}, nil
}
