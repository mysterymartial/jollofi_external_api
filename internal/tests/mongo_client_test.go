package tests

import (
	"context"
	"jollfi-gaming-api/internal/mocks"
	"testing"
)

func TestMockMongoClient_GameOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()
	ctx := context.Background()

	testGame := map[string]interface{}{
		"requester_address": "addr1",
		"accepter_address":  "addr2",
		"stake_amount":      1000,
		"status":            "pending",
	}

	gameID, err := client.CreateGame(ctx, testGame)
	if err != nil {
		t.Errorf("Expected no error creating game, got %v", err)
	}
	if gameID == "" {
		t.Errorf("Expected game ID, got empty string")
	}

	retrievedGame, err := client.GetGame(ctx, gameID)
	if err != nil {
		t.Errorf("Expected no error getting game, got %v", err)
	}
	if retrievedGame == nil {
		t.Errorf("Expected game, got nil")
	}

	updates := map[string]interface{}{
		"status": "completed",
	}
	err = client.UpdateGame(ctx, gameID, updates)
	if err != nil {
		t.Errorf("Expected no error updating game, got %v", err)
	}

	games, err := client.GetGamesByStatus(ctx, "completed")
	if err != nil {
		t.Errorf("Expected no error getting games by status, got %v", err)
	}
	if len(games) == 0 {
		t.Errorf("Expected at least 1 game, got %d", len(games))
	}
}

func TestMockMongoClient_TransactionOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()
	ctx := context.Background()

	testTx := map[string]interface{}{
		"game_id":      "game123",
		"type":         "stake",
		"from_address": "addr1",
		"amount":       1000,
		"status":       "pending",
	}

	txID, err := client.CreateTransaction(ctx, testTx)
	if err != nil {
		t.Errorf("Expected no error creating transaction, got %v", err)
	}
	if txID == "" {
		t.Errorf("Expected transaction ID, got empty string")
	}

	transactions, err := client.GetTransactionsByGameID(ctx, "game123")
	if err != nil {
		t.Errorf("Expected no error getting transactions, got %v", err)
	}

	if len(transactions) == 0 {
		t.Errorf("Expected at least 1 transaction, got %d", len(transactions))
	}

	err = client.UpdateTransactionStatus(ctx, "tx_digest_123", "confirmed", nil)
	if err != nil {
		t.Errorf("Expected no error updating transaction status, got %v", err)
	}

	pendingTxs, err := client.GetPendingTransactions(ctx)
	if err != nil {
		t.Errorf("Expected no error getting pending transactions, got %v", err)
	}
	if pendingTxs == nil {
		t.Errorf("Expected transactions slice, got nil")
	}
}

func TestMockMongoClient_UserOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()
	ctx := context.Background()

	testUser := map[string]interface{}{
		"address":  "user_addr_123",
		"username": "testuser",
		"balance":  5000,
	}

	userID, err := client.CreateUser(ctx, testUser)
	if err != nil {
		t.Errorf("Expected no error creating user, got %v", err)
	}
	if userID == "" {
		t.Errorf("Expected user ID, got empty string")
	}

	user, err := client.GetUser(ctx, "user_addr_123")
	if err != nil {
		t.Errorf("Expected no error getting user, got %v", err)
	}
	if user == nil {
		t.Errorf("Expected user, got nil")
	}

	updates := map[string]interface{}{
		"balance": 6000,
	}
	err = client.UpdateUser(ctx, "user_addr_123", updates)
	if err != nil {
		t.Errorf("Expected no error updating user, got %v", err)
	}

	err = client.UpdateUserLastSeen(ctx, "user_addr_123")
	if err != nil {
		t.Errorf("Expected no error updating user last seen, got %v", err)
	}
}

func TestMockMongoClient_StatisticsOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()
	ctx := context.Background()

	testGame := map[string]interface{}{
		"requester_address": "addr1",
		"accepter_address":  "addr2",
		"stake_amount":      1000,
		"status":            "active",
	}
	_, err := client.CreateGame(ctx, testGame)
	if err != nil {
		t.Errorf("Expected no error creating test game, got %v", err)
	}

	activeGames, err := client.GetActiveGames(ctx)
	if err != nil {
		t.Errorf("Expected no error getting active games, got %v", err)
	}
	if len(activeGames) == 0 {
		t.Errorf("Expected at least 1 active game, got %d", len(activeGames))
	}

	gameStats, err := client.GetGameStats(ctx)
	if err != nil {
		t.Errorf("Expected no error getting game stats, got %v", err)
	}
	if gameStats == nil {
		t.Errorf("Expected game stats, got nil")
	}
	if totalGames, ok := gameStats["total_games"]; !ok || totalGames.(int) == 0 {
		t.Errorf("Expected total_games > 0, got %v", totalGames)
	}

	userStats, err := client.GetUserStats(ctx, "addr1")
	if err != nil {
		t.Errorf("Expected no error getting user stats, got %v", err)
	}
	if userStats == nil {
		t.Errorf("Expected user stats, got nil")
	}

	collStats, err := client.GetCollectionStats(ctx)
	if err != nil {
		t.Errorf("Expected no error getting collection stats, got %v", err)
	}
	if collStats == nil {
		t.Errorf("Expected collection stats, got nil")
	}
}

func TestMockMongoClient_UtilityOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Expected no error on health check, got %v", err)
	}

	deletedCount, err := client.CleanupOldGames(ctx, 30)
	if err != nil {
		t.Errorf("Expected no error on cleanup, got %v", err)
	}
	if deletedCount < 0 {
		t.Errorf("Expected non-negative deleted count, got %d", deletedCount)
	}
}

func TestMockMongoClient_ErrorHandling(t *testing.T) {
	client := mocks.NewMockMongoClient()
	ctx := context.Background()

	client.Close()

	_, err := client.CreateGame(ctx, map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error on closed client, got nil")
	}

	_, err = client.GetGame(ctx, "game123")
	if err == nil {
		t.Errorf("Expected error on closed client, got nil")
	}

	err = client.UpdateGame(ctx, "game123", map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error on closed client, got nil")
	}

	client2 := mocks.NewMockMongoClient()
	defer client2.Close()

	_, err = client2.GetGame(ctx, "nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent game, got nil")
	}

	err = client2.UpdateGame(ctx, "nonexistent", map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error for non-existent game update, got nil")
	}
}

func TestMockCollection_AdvancedOperations(t *testing.T) {
	client := mocks.NewMockMongoClient()
	defer client.Close()

	db := client.GetDatabase("test_db")
	coll := db.Collection("test_collection")
	ctx := context.Background()

	docs := []interface{}{
		map[string]interface{}{"name": "doc1", "value": 100},
		map[string]interface{}{"name": "doc2", "value": 200},
		map[string]interface{}{"name": "doc3", "value": 300},
	}

	for _, doc := range docs {
		_, err := coll.InsertOne(ctx, doc)
		if err != nil {
			t.Errorf("Expected no error inserting document, got %v", err)
		}
	}
	cursor, err := coll.Find(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected no error on find, got %v", err)
	}

	docCount := 0
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		err := cursor.Decode(&doc)
		if err != nil {
			t.Errorf("Expected no error on decode, got %v", err)
		}
		docCount++
	}
	cursor.Close(ctx)

	if docCount != len(docs) {
		t.Errorf("Expected %d documents, got %d", len(docs), docCount)
	}

	for i := 0; i < len(docs); i++ {
		result, err := coll.DeleteOne(ctx, map[string]interface{}{})
		if err != nil {
			t.Errorf("Expected no error on delete, got %v", err)
		}
		if result.DeletedCount != 1 {
			t.Errorf("Expected 1 deleted document, got %d", result.DeletedCount)
		}
	}

	cursor2, err := coll.Find(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected no error on find empty collection, got %v", err)
	}

	emptyCount := 0
	for cursor2.Next(ctx) {
		emptyCount++
	}
	cursor2.Close(ctx)

	if emptyCount != 0 {
		t.Errorf("Expected 0 documents after deletion, got %d", emptyCount)
	}
}
