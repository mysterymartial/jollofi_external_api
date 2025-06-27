package tests

import (
	"context"
	"fmt"
	"jollfi-gaming-api/internal/mocks"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMockSuiClient_ExternalStake(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test successful stake
	digest, err := client.ExternalStake("0xrequester_coin", "0xaccepter_coin", 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if digest == "" {
		t.Errorf("Expected transaction digest, got empty string")
	}

	// Verify transaction was stored
	tx, exists := client.GetMockTransaction(digest)
	if !exists {
		t.Errorf("Expected transaction to be stored")
	}
	if tx == nil {
		t.Errorf("Expected transaction data, got nil")
	}

	// Test with empty coin IDs
	_, err = client.ExternalStake("", "0xaccepter_coin", 1000, ctx)
	if err == nil {
		t.Errorf("Expected error for empty requester coin ID")
	}

	_, err = client.ExternalStake("0xrequester_coin", "", 1000, ctx)
	if err == nil {
		t.Errorf("Expected error for empty accepter coin ID")
	}

	// Test with zero amount
	_, err = client.ExternalStake("0xrequester_coin", "0xaccepter_coin", 0, ctx)
	if err == nil {
		t.Errorf("Expected error for zero stake amount")
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.ExternalStake("0xrequester_coin", "0xaccepter_coin", 1000, ctx)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}
}

func TestMockSuiClient_ExternalPayWinner(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test successful payment with requester winning
	digest, err := client.ExternalPayWinner("0xrequester", "0xaccepter", 100, 80, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if digest == "" {
		t.Errorf("Expected transaction digest, got empty string")
	}

	// Verify transaction was stored and winner is correct
	tx, exists := client.GetMockTransaction(digest)
	if !exists {
		t.Errorf("Expected transaction to be stored")
	}
	if txMap, ok := tx.(map[string]interface{}); ok {
		if winner := txMap["winner"].(string); winner != "0xrequester" {
			t.Errorf("Expected winner to be 0xrequester, got %s", winner)
		}
	}

	// Test with accepter winning
	digest2, err := client.ExternalPayWinner("0xrequester", "0xaccepter", 70, 90, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tx2, _ := client.GetMockTransaction(digest2)
	if txMap, ok := tx2.(map[string]interface{}); ok {
		if winner := txMap["winner"].(string); winner != "0xaccepter" {
			t.Errorf("Expected winner to be 0xaccepter, got %s", winner)
		}
	}

	// Test tie game
	digest3, err := client.ExternalPayWinner("0xrequester", "0xaccepter", 85, 85, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tx3, _ := client.GetMockTransaction(digest3)
	if txMap, ok := tx3.(map[string]interface{}); ok {
		if winner := txMap["winner"].(string); winner != "tie" {
			t.Errorf("Expected winner to be tie, got %s", winner)
		}
	}

	// Test with empty addresses
	_, err = client.ExternalPayWinner("", "0xaccepter", 100, 80, 1000, ctx)
	if err == nil {
		t.Errorf("Expected error for empty requester address")
	}
	_, err = client.ExternalPayWinner("0xrequester", "", 100, 80, 1000, ctx)
	if err == nil {
		t.Errorf("Expected error for empty accepter address")
	}

	// Test with zero stake amount
	_, err = client.ExternalPayWinner("0xrequester", "0xaccepter", 100, 80, 0, ctx)
	if err == nil {
		t.Errorf("Expected error for zero stake amount")
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.ExternalPayWinner("0xrequester", "0xaccepter", 100, 80, 1000, ctx)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}
}

func TestMockSuiClient_GetBalance(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test default balance
	balance, err := client.GetBalance(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}

	// Test custom balance
	client.SetBalance(5000000)
	balance, err = client.GetBalance(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if balance != 5000000 {
		t.Errorf("Expected balance 5000000, got %d", balance)
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.GetBalance(ctx)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}

	// Test custom function
	client.SetShouldFail(false)
	client.GetBalanceFunc = func(ctx context.Context) (uint64, error) {
		return 9999999, nil
	}
	balance, err = client.GetBalance(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if balance != 9999999 {
		t.Errorf("Expected balance 9999999, got %d", balance)
	}
}

func TestMockSuiClient_GetCoins(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test getting coins
	coins, err := client.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(coins) != 3 {
		t.Errorf("Expected 3 coins, got %d", len(coins))
	}

	// Verify coin structure
	if coins[0]["coinObjectId"] != "0xmock_coin_1" {
		t.Errorf("Expected first coin ID to be 0xmock_coin_1, got %v", coins[0]["coinObjectId"])
	}
	if coins[0]["balance"] != "500000" {
		t.Errorf("Expected first coin balance to be 500000, got %v", coins[0]["balance"])
	}

	// Test with empty coin type
	_, err = client.GetCoins(ctx, "")
	if err == nil {
		t.Errorf("Expected error for empty coin type")
	}

	// Test custom coins
	customCoins := []map[string]interface{}{
		{"coinObjectId": "0xcustom_coin", "balance": "2000000"},
	}
	client.SetCoins(customCoins)
	coins, err = client.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(coins) != 1 {
		t.Errorf("Expected 1 coin, got %d", len(coins))
	}
	if coins[0]["coinObjectId"] != "0xcustom_coin" {
		t.Errorf("Expected custom coin ID, got %v", coins[0]["coinObjectId"])
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.GetCoins(ctx, "0x2::sui::SUI")
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}

	// Test custom function
	client.SetShouldFail(false)
	client.GetCoinsFunc = func(ctx context.Context, coinType string) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"coinObjectId": "0xfunc_coin", "balance": "7777777"},
		}, nil
	}
	coins, err = client.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(coins) != 1 {
		t.Errorf("Expected 1 coin, got %d", len(coins))
	}
	if coins[0]["coinObjectId"] != "0xfunc_coin" {
		t.Errorf("Expected func coin ID, got %v", coins[0]["coinObjectId"])
	}
}

func TestMockSuiClient_ExecuteTransactionBlock(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test successful execution
	txBytes := []byte("mock_transaction_data")
	digest, err := client.ExecuteTransactionBlock(ctx, txBytes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if digest == "" {
		t.Errorf("Expected transaction digest, got empty string")
	}

	// Verify transaction was stored
	tx, exists := client.GetMockTransaction(digest)
	if !exists {
		t.Errorf("Expected transaction to be stored")
	}
	if txMap, ok := tx.(map[string]interface{}); ok {
		if txMap["type"] != "execute_transaction_block" {
			t.Errorf("Expected transaction type to be execute_transaction_block")
		}
	}

	// Test with empty tx bytes
	_, err = client.ExecuteTransactionBlock(ctx, []byte{})
	if err == nil {
		t.Errorf("Expected error for empty transaction bytes")
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.ExecuteTransactionBlock(ctx, txBytes)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}
}

func TestMockSuiClient_BuildTransactionBlock(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test successful build
	params := map[string]interface{}{
		"function":  "test_function",
		"arguments": []interface{}{"arg1", "arg2"},
	}
	txBytes, err := client.BuildTransactionBlock(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(txBytes) == 0 {
		t.Errorf("Expected transaction bytes, got empty")
	}

	// Test with nil params
	_, err = client.BuildTransactionBlock(ctx, nil)
	if err == nil {
		t.Errorf("Expected error for nil parameters")
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.BuildTransactionBlock(ctx, params)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}
}

func TestMockSuiClient_GetTransactionBlock(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// First create a transaction
	digest, _ := client.ExternalStake("0xrequester_coin", "0xaccepter_coin", 1000, ctx)

	// Test getting existing transaction
	tx, err := client.GetTransactionBlock(ctx, digest)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if tx == nil {
		t.Errorf("Expected transaction data, got nil")
	}

	// Test getting non-existent transaction (should return default)
	tx, err = client.GetTransactionBlock(ctx, "0xnonexistent")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if tx == nil {
		t.Errorf("Expected default transaction data, got nil")
	}

	// Test with empty digest
	_, err = client.GetTransactionBlock(ctx, "")
	if err == nil {
		t.Errorf("Expected error for empty digest")
	}

	// Test failure mode
	client.SetShouldFail(true)
	_, err = client.GetTransactionBlock(ctx, digest)
	if err == nil {
		t.Errorf("Expected error when shouldFail is true")
	}
}

func TestMockSuiClient_CustomFunctions(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test custom ExternalStake function
	client.ExternalStakeFunc = func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
		return "custom_stake_digest", nil
	}

	digest, err := client.ExternalStake("0xreq", "0xacc", 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if digest != "custom_stake_digest" {
		t.Errorf("Expected custom_stake_digest, got %s", digest)
	}

	// Test custom ExternalPayWinner function
	client.ExternalPayWinnerFunc = func(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error) {
		return "custom_pay_winner_digest", nil
	}

	digest, err = client.ExternalPayWinner("0xreq", "0xacc", 100, 80, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if digest != "custom_pay_winner_digest" {
		t.Errorf("Expected custom_pay_winner_digest, got %s", digest)
	}
}

func TestMockSuiClient_HelperMethods(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test transaction count
	initialCount := client.GetTransactionCount()
	if initialCount != 0 {
		t.Errorf("Expected initial transaction count to be 0, got %d", initialCount)
	}

	// Add some transactions
	client.ExternalStake("0xreq1", "0xacc1", 1000, ctx)
	client.ExternalStake("0xreq2", "0xacc2", 2000, ctx)

	count := client.GetTransactionCount()
	if count != 2 {
		t.Errorf("Expected transaction count to be 2, got %d", count)
	}

	// Test clear transactions
	client.ClearTransactions()
	count = client.GetTransactionCount()
	if count != 0 {
		t.Errorf("Expected transaction count to be 0 after clear, got %d", count)
	}

	// Test custom responses
	client.SetCustomResponse("test_key", "test_value")
	value, exists := client.GetCustomResponse("test_key")
	if !exists {
		t.Errorf("Expected custom response to exist")
	}
	if value != "test_value" {
		t.Errorf("Expected test_value, got %v", value)
	}

	// Test non-existent custom response
	_, exists = client.GetCustomResponse("nonexistent_key")
	if exists {
		t.Errorf("Expected custom response to not exist")
	}
}

func TestMockSuiClient_EdgeCases(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test with very large amounts
	digest, err := client.ExternalStake("0xreq", "0xacc", 18446744073709551615, ctx) // max uint64
	if err != nil {
		t.Errorf("Expected no error for max uint64, got %v", err)
	}
	if digest == "" {
		t.Errorf("Expected digest for max uint64 amount")
	}

	// Test with special characters in addresses
	digest, err = client.ExternalPayWinner("0x123!@#", "0x456$%^", 100, 80, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error for special characters, got %v", err)
	}
	if digest == "" {
		t.Errorf("Expected digest for special character addresses")
	}

	// Test prize calculation accuracy
	digest, err = client.ExternalPayWinner("0xreq", "0xacc", 100, 80, 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tx, exists := client.GetMockTransaction(digest)
	if !exists {
		t.Errorf("Expected transaction to exist")
	}

	if txMap, ok := tx.(map[string]interface{}); ok {
		totalStake := txMap["totalStake"].(uint64)
		apiFee := txMap["apiFee"].(uint64)
		escrowFee := txMap["escrowFee"].(uint64)
		prizeAmount := txMap["prizeAmount"].(uint64)

		expectedTotal := uint64(2000)   // 1000 * 2
		expectedApiFee := uint64(160)   // 2000 * 8 / 100
		expectedEscrowFee := uint64(40) // 2000 * 2 / 100
		expectedPrize := uint64(1800)   // 2000 - 160 - 40

		if totalStake != expectedTotal {
			t.Errorf("Expected total stake %d, got %d", expectedTotal, totalStake)
		}
		if apiFee != expectedApiFee {
			t.Errorf("Expected API fee %d, got %d", expectedApiFee, apiFee)
		}
		if escrowFee != expectedEscrowFee {
			t.Errorf("Expected escrow fee %d, got %d", expectedEscrowFee, escrowFee)
		}
		if prizeAmount != expectedPrize {
			t.Errorf("Expected prize amount %d, got %d", expectedPrize, prizeAmount)
		}
	}
}

// Benchmark tests
func BenchmarkMockSuiClient_ExternalStake(b *testing.B) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requesterCoinID := fmt.Sprintf("0xreq_%d", i)
		accepterCoinID := fmt.Sprintf("0xacc_%d", i)
		_, err := client.ExternalStake(requesterCoinID, accepterCoinID, 1000, ctx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkMockSuiClient_ExternalPayWinner(b *testing.B) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requesterAddr := fmt.Sprintf("0xreq_%d", i)
		accepterAddr := fmt.Sprintf("0xacc_%d", i)
		_, err := client.ExternalPayWinner(requesterAddr, accepterAddr, 100, 80, 1000, ctx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkMockSuiClient_GetBalance(b *testing.B) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetBalance(ctx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkMockSuiClient_GetCoins(b *testing.B) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetCoins(ctx, "0x2::sui::SUI")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Test suite for concurrent access
func TestMockSuiClient_Concurrent(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Test concurrent ExternalStake operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				requesterCoinID := fmt.Sprintf("0xreq_%d_%d", id, j)
				accepterCoinID := fmt.Sprintf("0xacc_%d_%d", id, j)
				_, err := client.ExternalStake(requesterCoinID, accepterCoinID, 1000, ctx)
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify all transactions were recorded
	expectedCount := numGoroutines * numOperations
	actualCount := client.GetTransactionCount()
	if actualCount != expectedCount {
		t.Errorf("Expected %d transactions, got %d", expectedCount, actualCount)
	}
}

// Test error handling scenarios
func TestMockSuiClient_ErrorHandling(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err := client.ExternalStake("0xreq", "0xacc", 1000, cancelCtx)
	if err == nil {
		t.Errorf("Expected error for cancelled context")
	}

	// Test timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	_, err = client.GetBalance(timeoutCtx)
	if err == nil {
		t.Errorf("Expected error for timeout context")
	}

	// Test custom error function
	client.ExternalStakeFunc = func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
		return "", fmt.Errorf("custom error: insufficient funds")
	}

	_, err = client.ExternalStake("0xreq", "0xacc", 1000, context.Background())
	if err == nil {
		t.Errorf("Expected custom error")
	}
	if !strings.Contains(err.Error(), "insufficient funds") {
		t.Errorf("Expected custom error message, got: %v", err)
	}
}

// Test memory usage and cleanup
func TestMockSuiClient_MemoryManagement(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Create many transactions
	const numTransactions = 1000
	for i := 0; i < numTransactions; i++ {
		requesterCoinID := fmt.Sprintf("0xreq_%d", i)
		accepterCoinID := fmt.Sprintf("0xacc_%d", i)
		_, err := client.ExternalStake(requesterCoinID, accepterCoinID, 1000, ctx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// Verify all transactions exist
	if client.GetTransactionCount() != numTransactions {
		t.Errorf("Expected %d transactions, got %d", numTransactions, client.GetTransactionCount())
	}

	// Clear transactions and verify cleanup
	client.ClearTransactions()
	if client.GetTransactionCount() != 0 {
		t.Errorf("Expected 0 transactions after clear, got %d", client.GetTransactionCount())
	}

	// Verify we can still create new transactions after cleanup
	_, err := client.ExternalStake("0xnew_req", "0xnew_acc", 1000, ctx)
	if err != nil {
		t.Errorf("Expected no error after cleanup, got %v", err)
	}
	if client.GetTransactionCount() != 1 {
		t.Errorf("Expected 1 transaction after cleanup and new transaction, got %d", client.GetTransactionCount())
	}
}

// Test state consistency
func TestMockSuiClient_StateConsistency(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Test balance consistency
	initialBalance, _ := client.GetBalance(ctx)
	client.SetBalance(5000000)
	newBalance, _ := client.GetBalance(ctx)

	if newBalance == initialBalance {
		t.Errorf("Balance should have changed after SetBalance")
	}
	if newBalance != 5000000 {
		t.Errorf("Expected balance 5000000, got %d", newBalance)
	}

	// Test coins consistency
	initialCoins, _ := client.GetCoins(ctx, "0x2::sui::SUI")
	customCoins := []map[string]interface{}{
		{"coinObjectId": "0xcustom1", "balance": "1000000"},
		{"coinObjectId": "0xcustom2", "balance": "2000000"},
	}
	client.SetCoins(customCoins)
	newCoins, _ := client.GetCoins(ctx, "0x2::sui::SUI")

	if len(newCoins) == len(initialCoins) {
		t.Errorf("Coins should have changed after SetCoins")
	}
	if len(newCoins) != 2 {
		t.Errorf("Expected 2 custom coins, got %d", len(newCoins))
	}

	// Test failure state consistency
	client.SetShouldFail(true)

	_, err1 := client.GetBalance(ctx)
	_, err2 := client.GetCoins(ctx, "0x2::sui::SUI")
	_, err3 := client.ExternalStake("0xreq", "0xacc", 1000, ctx)

	if err1 == nil || err2 == nil || err3 == nil {
		t.Errorf("All operations should fail when shouldFail is true")
	}

	// Reset failure state
	client.SetShouldFail(false)
	_, err4 := client.GetBalance(ctx)
	if err4 != nil {
		t.Errorf("Operations should succeed after resetting shouldFail, got error: %v", err4)
	}
}

// Test transaction data integrity
func TestMockSuiClient_TransactionDataIntegrity(t *testing.T) {
	client := mocks.NewMockSuiClient()
	ctx := context.Background()

	// Create a stake transaction
	stakeDigest, err := client.ExternalStake("0xreq_coin", "0xacc_coin", 1500, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify stake transaction data
	stakeTx, exists := client.GetMockTransaction(stakeDigest)
	if !exists {
		t.Fatalf("Stake transaction should exist")
	}

	stakeMap, ok := stakeTx.(map[string]interface{})
	if !ok {
		t.Fatalf("Transaction should be a map")
	}

	expectedStakeFields := []string{"type", "requesterCoinID", "accepterCoinID", "amount", "timestamp"}
	for _, field := range expectedStakeFields {
		if _, exists := stakeMap[field]; !exists {
			t.Errorf("Stake transaction missing field: %s", field)
		}
	}

	if stakeMap["type"] != "external_stake" {
		t.Errorf("Expected type external_stake, got %v", stakeMap["type"])
	}
	if stakeMap["amount"] != uint64(1500) {
		t.Errorf("Expected amount 1500, got %v", stakeMap["amount"])
	}

	// Create a pay winner transaction
	payDigest, err := client.ExternalPayWinner("0xreq_addr", "0xacc_addr", 95, 85, 2000, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify pay winner transaction data
	payTx, exists := client.GetMockTransaction(payDigest)
	if !exists {
		t.Fatalf("Pay winner transaction should exist")
	}

	payMap, ok := payTx.(map[string]interface{})
	if !ok {
		t.Fatalf("Transaction should be a map")
	}

	expectedPayFields := []string{"type", "requesterAddress", "accepterAddress", "requesterScore", "accepterScore", "stakeAmount", "totalStake", "apiFee", "escrowFee", "prizeAmount", "winner", "timestamp"}
	for _, field := range expectedPayFields {
		if _, exists := payMap[field]; !exists {
			t.Errorf("Pay winner transaction missing field: %s", field)
		}
	}

	if payMap["type"] != "external_pay_winner" {
		t.Errorf("Expected type external_pay_winner, got %v", payMap["type"])
	}
	if payMap["requesterScore"] != uint64(95) {
		t.Errorf("Expected requester score 95, got %v", payMap["requesterScore"])
	}
	if payMap["accepterScore"] != uint64(85) {
		t.Errorf("Expected accepter score 85, got %v", payMap["accepterScore"])
	}

	// Verify winner determination
	expectedWinner := "0xreq_addr" // Higher score wins
	if payMap["winner"] != expectedWinner {
		t.Errorf("Expected winner %s, got %v", expectedWinner, payMap["winner"])
	}
}
