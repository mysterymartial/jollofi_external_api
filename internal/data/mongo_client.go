package data

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"jollfi-gaming-api/internal/interfaces"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	dbName   string
}

// Game represents the game document structure
type Game struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RequesterAddress  string             `bson:"requester_address" json:"requester_address"`
	AccepterAddress   string             `bson:"accepter_address" json:"accepter_address"`
	RequesterCoinID   string             `bson:"requester_coin_id" json:"requester_coin_id"`
	AccepterCoinID    string             `bson:"accepter_coin_id" json:"accepter_coin_id"`
	StakeAmount       uint64             `bson:"stake_amount" json:"stake_amount"`
	Status            string             `bson:"status" json:"status"`
	RequesterScore    *uint64            `bson:"requester_score,omitempty" json:"requester_score,omitempty"`
	AccepterScore     *uint64            `bson:"accepter_score,omitempty" json:"accepter_score,omitempty"`
	Winner            string             `bson:"winner,omitempty" json:"winner,omitempty"`
	TransactionDigest string             `bson:"transaction_digest,omitempty" json:"transaction_digest,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
	CompletedAt       *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// User represents the user document structure
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Address   string             `bson:"address" json:"address"`
	Username  string             `bson:"username,omitempty" json:"username,omitempty"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
	Balance   uint64             `bson:"balance" json:"balance"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	LastSeen  time.Time          `bson:"last_seen" json:"last_seen"`
}

// Transaction represents the transaction document structure
type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GameID      primitive.ObjectID `bson:"game_id" json:"game_id"`
	Type        string             `bson:"type" json:"type"` // "stake", "payout", "refund"
	FromAddress string             `bson:"from_address" json:"from_address"`
	ToAddress   string             `bson:"to_address,omitempty" json:"to_address,omitempty"`
	Amount      uint64             `bson:"amount" json:"amount"`
	TxDigest    string             `bson:"tx_digest" json:"tx_digest"`
	Status      string             `bson:"status" json:"status"` // "pending", "confirmed", "failed"
	BlockHeight *uint64            `bson:"block_height,omitempty" json:"block_height,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ConfirmedAt *time.Time         `bson:"confirmed_at,omitempty" json:"confirmed_at,omitempty"`
}

func NewMongoClient(mongoURI string, dbName string) *MongoClient {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(100).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetSocketTimeout(30 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test connection
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("✅ Connected to MongoDB")

	mongoClient := &MongoClient{
		client:   client,
		database: client.Database(dbName),
		dbName:   dbName,
	}

	// Create indexes
	if err := mongoClient.createIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	return mongoClient
}

// Interface method implementations

// GetDatabase implements interfaces.MongoClientInterface - THIS FIXES THE ERROR
func (m *MongoClient) GetDatabase(name string) interfaces.MongoDatabaseInterface {
	database := m.client.Database(name)
	return &MongoDatabase{database: database}
}

// Close implements interfaces.MongoClientInterface
func (m *MongoClient) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.client.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}
}

// Ping implements interfaces.MongoClientInterface
func (m *MongoClient) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// Legacy method for backward compatibility
func (m *MongoClient) PingLegacy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Ping(ctx, readpref.Primary())
}

// CreateGame implements interfaces.MongoClientInterface
func (m *MongoClient) CreateGame(ctx context.Context, game interface{}) (string, error) {
	collection := m.database.Collection("games")

	// Convert interface{} to Game struct if needed
	var gameDoc Game
	switch g := game.(type) {
	case Game:
		gameDoc = g
	case *Game:
		gameDoc = *g
	case map[string]interface{}:
		// Convert map to Game struct
		gameDoc = Game{
			RequesterAddress: getStringFromMap(g, "requester_address"),
			AccepterAddress:  getStringFromMap(g, "accepter_address"),
			RequesterCoinID:  getStringFromMap(g, "requester_coin_id"),
			AccepterCoinID:   getStringFromMap(g, "accepter_coin_id"),
			StakeAmount:      getUint64FromMap(g, "stake_amount"),
			Status:           getStringFromMap(g, "status"),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
	default:
		return "", fmt.Errorf("unsupported game type: %T", game)
	}

	// Set timestamps if not already set
	if gameDoc.CreatedAt.IsZero() {
		gameDoc.CreatedAt = time.Now()
	}
	if gameDoc.UpdatedAt.IsZero() {
		gameDoc.UpdatedAt = time.Now()
	}

	result, err := collection.InsertOne(ctx, gameDoc)
	if err != nil {
		return "", fmt.Errorf("failed to create game: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return fmt.Sprintf("%v", result.InsertedID), nil
}

// GetGame implements interfaces.MongoClientInterface
func (m *MongoClient) GetGame(ctx context.Context, gameID string) (interface{}, error) {
	collection := m.database.Collection("games")

	objectID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return nil, fmt.Errorf("invalid game ID format: %v", err)
	}

	var game Game
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game: %v", err)
	}

	return game, nil
}

// UpdateGame implements interfaces.MongoClientInterface
func (m *MongoClient) UpdateGame(ctx context.Context, gameID string, updates interface{}) error {
	collection := m.database.Collection("games")

	objectID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return fmt.Errorf("invalid game ID format: %v", err)
	}

	// Prepare update document
	var updateDoc bson.M
	switch u := updates.(type) {
	case bson.M:
		updateDoc = u
	case map[string]interface{}:
		updateDoc = bson.M(u)
	case Game:
		updateDoc = bson.M{
			"requester_address":  u.RequesterAddress,
			"accepter_address":   u.AccepterAddress,
			"requester_coin_id":  u.RequesterCoinID,
			"accepter_coin_id":   u.AccepterCoinID,
			"stake_amount":       u.StakeAmount,
			"status":             u.Status,
			"requester_score":    u.RequesterScore,
			"accepter_score":     u.AccepterScore,
			"winner":             u.Winner,
			"transaction_digest": u.TransactionDigest,
			"completed_at":       u.CompletedAt,
		}
	default:
		return fmt.Errorf("unsupported update type: %T", updates)
	}

	// Always update the updated_at field
	updateDoc["updated_at"] = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updateDoc}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update game: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("game not found")
	}

	return nil
}

// GetGamesByStatus implements interfaces.MongoClientInterface
func (m *MongoClient) GetGamesByStatus(ctx context.Context, status string) ([]interface{}, error) {
	collection := m.database.Collection("games")

	filter := bson.M{"status": status}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query games by status: %v", err)
	}
	defer cursor.Close(ctx)

	var games []interface{}
	for cursor.Next(ctx) {
		var game Game
		if err := cursor.Decode(&game); err != nil {
			return nil, fmt.Errorf("failed to decode game: %v", err)
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return games, nil
}

// GetGamesByAddress implements interfaces.MongoClientInterface

// GetGamesByAddress implements interfaces.MongoClientInterface
func (m *MongoClient) GetGamesByAddress(ctx context.Context, address string) ([]interface{}, error) {
	collection := m.database.Collection("games")

	filter := bson.M{
		"$or": []bson.M{
			{"requester_address": address},
			{"accepter_address": address},
		},
	}

	// Sort by created_at descending
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query games by address: %v", err)
	}
	defer cursor.Close(ctx)

	var games []interface{}
	for cursor.Next(ctx) {
		var game Game
		if err := cursor.Decode(&game); err != nil {
			return nil, fmt.Errorf("failed to decode game: %v", err)
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return games, nil
}

// CreateTransaction implements interfaces.MongoClientInterface
func (m *MongoClient) CreateTransaction(ctx context.Context, transaction interface{}) (string, error) {
	collection := m.database.Collection("transactions")

	// Convert interface{} to Transaction struct if needed
	var txDoc Transaction
	switch tx := transaction.(type) {
	case Transaction:
		txDoc = tx
	case *Transaction:
		txDoc = *tx
	case map[string]interface{}:
		// Convert map to Transaction struct
		gameID, _ := primitive.ObjectIDFromHex(getStringFromMap(tx, "game_id"))
		txDoc = Transaction{
			GameID:      gameID,
			Type:        getStringFromMap(tx, "type"),
			FromAddress: getStringFromMap(tx, "from_address"),
			ToAddress:   getStringFromMap(tx, "to_address"),
			Amount:      getUint64FromMap(tx, "amount"),
			TxDigest:    getStringFromMap(tx, "tx_digest"),
			Status:      getStringFromMap(tx, "status"),
			CreatedAt:   time.Now(),
		}
	default:
		return "", fmt.Errorf("unsupported transaction type: %T", transaction)
	}

	// Set timestamps if not already set
	if txDoc.CreatedAt.IsZero() {
		txDoc.CreatedAt = time.Now()
	}

	result, err := collection.InsertOne(ctx, txDoc)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return fmt.Sprintf("%v", result.InsertedID), nil
}

// GetTransactionsByGameID implements interfaces.MongoClientInterface
func (m *MongoClient) GetTransactionsByGameID(ctx context.Context, gameID string) ([]interface{}, error) {
	collection := m.database.Collection("transactions")

	objectID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return nil, fmt.Errorf("invalid game ID format: %v", err)
	}

	filter := bson.M{"game_id": objectID}
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer cursor.Close(ctx)

	var transactions []interface{}
	for cursor.Next(ctx) {
		var tx Transaction
		if err := cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %v", err)
		}
		transactions = append(transactions, tx)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return transactions, nil
}

// Additional utility methods

// CreateUser creates a new user document
func (m *MongoClient) CreateUser(ctx context.Context, user User) (string, error) {
	collection := m.database.Collection("users")

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.LastSeen = time.Now()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return fmt.Sprintf("%v", result.InsertedID), nil
}

// GetUser retrieves a user by address
func (m *MongoClient) GetUser(ctx context.Context, address string) (*User, error) {
	collection := m.database.Collection("users")

	var user User
	err := collection.FindOne(ctx, bson.M{"address": address}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (m *MongoClient) UpdateUser(ctx context.Context, address string, updates bson.M) error {
	collection := m.database.Collection("users")

	updates["updated_at"] = time.Now()
	filter := bson.M{"address": address}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateUserLastSeen updates the last seen timestamp for a user
func (m *MongoClient) UpdateUserLastSeen(ctx context.Context, address string) error {
	return m.UpdateUser(ctx, address, bson.M{"last_seen": time.Now()})
}

// GetActiveGames retrieves games that are currently active (not completed)
func (m *MongoClient) GetActiveGames(ctx context.Context) ([]Game, error) {
	collection := m.database.Collection("games")

	filter := bson.M{
		"status": bson.M{
			"$in": []string{"pending", "staked", "in_progress"},
		},
	}

	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query active games: %v", err)
	}
	defer cursor.Close(ctx)

	var games []Game
	for cursor.Next(ctx) {
		var game Game
		if err := cursor.Decode(&game); err != nil {
			return nil, fmt.Errorf("failed to decode game: %v", err)
		}
		games = append(games, game)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return games, nil
}

// GetGameStats retrieves statistics about games
func (m *MongoClient) GetGameStats(ctx context.Context) (map[string]interface{}, error) {
	collection := m.database.Collection("games")

	// Aggregate pipeline to get game statistics
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":         "$status",
				"count":       bson.M{"$sum": 1},
				"total_stake": bson.M{"$sum": "$stake_amount"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate game stats: %v", err)
	}
	defer cursor.Close(ctx)

	stats := make(map[string]interface{})
	var totalGames int64
	var totalStake uint64

	for cursor.Next(ctx) {
		var result struct {
			ID         string `bson:"_id"`
			Count      int64  `bson:"count"`
			TotalStake uint64 `bson:"total_stake"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %v", err)
		}

		stats[result.ID] = map[string]interface{}{
			"count":       result.Count,
			"total_stake": result.TotalStake,
		}

		totalGames += result.Count
		totalStake += result.TotalStake
	}

	stats["total_games"] = totalGames
	stats["total_stake"] = totalStake

	return stats, nil
}

// GetUserStats retrieves statistics for a specific user
func (m *MongoClient) GetUserStats(ctx context.Context, address string) (map[string]interface{}, error) {
	collection := m.database.Collection("games")

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"requester_address": address},
					{"accepter_address": address},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":         "$status",
				"count":       bson.M{"$sum": 1},
				"total_stake": bson.M{"$sum": "$stake_amount"},
				"wins": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$winner", address}},
							1,
							0,
						},
					},
				},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate user stats: %v", err)
	}
	defer cursor.Close(ctx)

	stats := make(map[string]interface{})
	var totalGames int64
	var totalWins int64
	var totalStake uint64

	for cursor.Next(ctx) {
		var result struct {
			ID         string `bson:"_id"`
			Count      int64  `bson:"count"`
			TotalStake uint64 `bson:"total_stake"`
			Wins       int64  `bson:"wins"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode user stats: %v", err)
		}

		stats[result.ID] = map[string]interface{}{
			"count":       result.Count,
			"total_stake": result.TotalStake,
			"wins":        result.Wins,
		}

		totalGames += result.Count
		totalWins += result.Wins
		totalStake += result.TotalStake
	}

	stats["total_games"] = totalGames
	stats["total_wins"] = totalWins
	stats["total_stake"] = totalStake

	// Calculate win rate
	if totalGames > 0 {
		stats["win_rate"] = float64(totalWins) / float64(totalGames)
	} else {
		stats["win_rate"] = 0.0
	}

	return stats, nil
}

// UpdateTransactionStatus updates the status of a transaction
func (m *MongoClient) UpdateTransactionStatus(ctx context.Context, txDigest string, status string, blockHeight *uint64) error {
	collection := m.database.Collection("transactions")

	updates := bson.M{
		"status": status,
	}

	if status == "confirmed" {
		updates["confirmed_at"] = time.Now()
		if blockHeight != nil {
			updates["block_height"] = *blockHeight
		}
	}

	filter := bson.M{"tx_digest": txDigest}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// GetPendingTransactions retrieves transactions with pending status
func (m *MongoClient) GetPendingTransactions(ctx context.Context) ([]Transaction, error) {
	collection := m.database.Collection("transactions")

	filter := bson.M{"status": "pending"}
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending transactions: %v", err)
	}
	defer cursor.Close(ctx)

	var transactions []Transaction
	for cursor.Next(ctx) {
		var tx Transaction
		if err := cursor.Decode(&tx); err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %v", err)
		}
		transactions = append(transactions, tx)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return transactions, nil
}

// Helper method to create database indexes
func (m *MongoClient) createIndexes(ctx context.Context) error {
	// Games collection indexes
	gamesCollection := m.database.Collection("games")
	gamesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"requester_address", 1}},
		},
		{
			Keys: bson.D{{"accepter_address", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
		{
			Keys: bson.D{{"transaction_digest", 1}},
		},
	}

	if _, err := gamesCollection.Indexes().CreateMany(ctx, gamesIndexes); err != nil {
		return fmt.Errorf("failed to create games indexes: %v", err)
	}

	// Users collection indexes
	usersCollection := m.database.Collection("users")
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"address", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	if _, err := usersCollection.Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %v", err)
	}

	// Transactions collection indexes
	txCollection := m.database.Collection("transactions")
	txIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"game_id", 1}},
		},
		{
			Keys: bson.D{{"tx_digest", 1}},
		},
		{
			Keys: bson.D{{"from_address", 1}},
		},
		{
			Keys: bson.D{{"to_address", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	}

	if _, err := txCollection.Indexes().CreateMany(ctx, txIndexes); err != nil {
		return fmt.Errorf("failed to create transactions indexes: %v", err)
	}

	return nil
}

// Helper functions for type conversion
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getUint64FromMap(m map[string]interface{}, key string) uint64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case uint64:
			return v
		case int64:
			if v >= 0 {
				return uint64(v)
			}
		case int:
			if v >= 0 {
				return uint64(v)
			}
		case float64:
			if v >= 0 {
				return uint64(v)
			}
		}
	}
	return 0
}

// Health check method
func (m *MongoClient) HealthCheck(ctx context.Context) error {
	return m.Ping(ctx)
}

// GetCollectionStats returns statistics about collections
func (m *MongoClient) GetCollectionStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	collections := []string{"games", "users", "transactions"}

	for _, collName := range collections {
		collection := m.database.Collection(collName)
		count, err := collection.CountDocuments(ctx, bson.M{})
		if err != nil {
			return nil, fmt.Errorf("failed to count documents in %s: %v", collName, err)
		}
		stats[collName] = count
	}

	return stats, nil
}

// Cleanup old completed games (optional maintenance method)
func (m *MongoClient) CleanupOldGames(ctx context.Context, olderThanDays int) (int64, error) {
	collection := m.database.Collection("games")
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	filter := bson.M{
		"status":       "completed",
		"completed_at": bson.M{"$lt": cutoffDate},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old games: %v", err)
	}

	return result.DeletedCount, nil
}
