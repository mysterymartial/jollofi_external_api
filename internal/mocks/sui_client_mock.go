package mocks

import (
	"context"
	"fmt"
	"jollfi-gaming-api/internal/interfaces"
	"sync"
	"time" // Added for timestamp
)

// Ensure MockSuiClient implements the interface (compile-time check)
var _ interfaces.SuiClientInterface = (*MockSuiClient)(nil)

// MockSuiClient implements a mock for SuiClient
type MockSuiClient struct {
	shouldFail   bool
	responses    map[string]interface{}
	balance      uint64
	coins        []map[string]interface{}
	transactions map[string]interface{}
	mu           sync.RWMutex // Mutex for thread-safety

	// Function overrides for custom behavior
	ExternalStakeFunc     func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error)
	ExternalPayWinnerFunc func(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error)
	GetBalanceFunc        func(ctx context.Context) (uint64, error)
	GetCoinsFunc          func(ctx context.Context, coinType string) ([]map[string]interface{}, error)
}

// NewMockSuiClient creates a new mock Sui client
func NewMockSuiClient() *MockSuiClient {
	return &MockSuiClient{
		shouldFail:   false,
		responses:    make(map[string]interface{}),
		balance:      1000000, // Default 1 SUI
		transactions: make(map[string]interface{}),
		coins: []map[string]interface{}{
			{
				"coinObjectId": "0xmock_coin_1",
				"balance":      "500000",
			},
			{
				"coinObjectId": "0xmock_coin_2",
				"balance":      "300000",
			},
			{
				"coinObjectId": "0xmock_coin_3",
				"balance":      "200000",
			},
		},
	}
}

// SetShouldFail sets whether the mock should simulate failures
func (m *MockSuiClient) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

// SetBalance sets the mock balance
func (m *MockSuiClient) SetBalance(balance uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balance = balance
}

// SetCoins sets the mock coins
func (m *MockSuiClient) SetCoins(coins []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coins = coins
}

// AddTransaction adds a mock transaction
func (m *MockSuiClient) AddTransaction(digest string, tx interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions[digest] = tx
}

// ExternalStake mocks the external stake method
func (m *MockSuiClient) ExternalStake(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("mock external stake: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return "", fmt.Errorf("mock external stake failed")
	}

	// Use custom function if provided
	if m.ExternalStakeFunc != nil {
		return m.ExternalStakeFunc(requesterCoinID, accepterCoinID, amount, ctx)
	}

	// Validate inputs
	if requesterCoinID == "" || accepterCoinID == "" {
		return "", fmt.Errorf("both coin IDs are required")
	}
	if amount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	// Generate mock transaction digest
	digest := fmt.Sprintf("mock_stake_tx_%s_%s_%d", requesterCoinID, accepterCoinID, amount)

	// Store transaction
	m.AddTransaction(digest, map[string]interface{}{
		"type":            "external_stake",
		"requesterCoinID": requesterCoinID,
		"accepterCoinID":  accepterCoinID,
		"amount":          amount,
		"status":          "success",
		"timestamp":       time.Now().Unix(), // Added timestamp
	})

	return digest, nil
}

// ExternalPayWinner mocks the external pay winner method
func (m *MockSuiClient) ExternalPayWinner(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("mock external pay winner: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return "", fmt.Errorf("mock external pay winner failed")
	}

	// Use custom function if provided
	if m.ExternalPayWinnerFunc != nil {
		return m.ExternalPayWinnerFunc(requesterAddress, accepterAddress, requesterScore, accepterScore, stakeAmount, ctx)
	}

	// Validate inputs
	if requesterAddress == "" || accepterAddress == "" {
		return "", fmt.Errorf("both addresses are required")
	}
	if stakeAmount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	// Determine winner
	var winner string
	if requesterScore > accepterScore {
		winner = requesterAddress
	} else if accepterScore > requesterScore {
		winner = accepterAddress
	} else {
		winner = "tie"
	}

	// Calculate prize (simplified)
	totalStake := stakeAmount * 2
	apiFee := totalStake * 8 / 100    // 8% API fee
	escrowFee := totalStake * 2 / 100 // 2% escrow fee
	prizeAmount := totalStake - apiFee - escrowFee

	// Generate mock transaction digest
	digest := fmt.Sprintf("mock_pay_winner_tx_%s_%s_%d", requesterAddress, accepterAddress, stakeAmount)

	// Store transaction
	m.AddTransaction(digest, map[string]interface{}{
		"type":             "external_pay_winner",
		"requesterAddress": requesterAddress,
		"accepterAddress":  accepterAddress,
		"requesterScore":   requesterScore,
		"accepterScore":    accepterScore,
		"winner":           winner,
		"stakeAmount":      stakeAmount,
		"totalStake":       totalStake,
		"prizeAmount":      prizeAmount,
		"apiFee":           apiFee,
		"escrowFee":        escrowFee,
		"status":           "success",
		"timestamp":        time.Now().Unix(), // Added timestamp
	})

	return digest, nil
}

// ExecuteTransactionBlock mocks executing a transaction block
func (m *MockSuiClient) ExecuteTransactionBlock(ctx context.Context, txBytes []byte) (string, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("mock execute transaction block: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return "", fmt.Errorf("mock execute transaction block failed")
	}

	if len(txBytes) == 0 {
		return "", fmt.Errorf("transaction bytes cannot be empty")
	}

	// Generate mock digest based on tx bytes
	digest := fmt.Sprintf("mock_execute_tx_%x", len(txBytes))

	// Store transaction
	m.AddTransaction(digest, map[string]interface{}{
		"type":    "execute_transaction_block",
		"txBytes": txBytes,
		"status":  "success",
	})

	return digest, nil
}

// GetTransactionBlock mocks getting a transaction block
func (m *MockSuiClient) GetTransactionBlock(ctx context.Context, digest string) (interface{}, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("mock get transaction block: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return nil, fmt.Errorf("mock get transaction block failed")
	}

	if digest == "" {
		return nil, fmt.Errorf("transaction digest is required")
	}

	// Check if transaction exists in mock storage
	m.mu.RLock()
	defer m.mu.RUnlock()
	if tx, exists := m.transactions[digest]; exists {
		return tx, nil
	}

	// Return default mock transaction
	return map[string]interface{}{
		"digest": digest,
		"status": "success",
		"effects": map[string]interface{}{
			"status": map[string]interface{}{
				"status": "success",
			},
		},
		"events": []interface{}{},
	}, nil
}

// BuildTransactionBlock mocks building a transaction block
func (m *MockSuiClient) BuildTransactionBlock(ctx context.Context, params interface{}) ([]byte, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("mock build transaction block: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return nil, fmt.Errorf("mock build transaction block failed")
	}

	if params == nil {
		return nil, fmt.Errorf("transaction parameters are required")
	}

	// Generate mock transaction bytes based on params
	mockBytes := []byte(fmt.Sprintf("mock_tx_bytes_%v", params))

	return mockBytes, nil
}

// GetBalance mocks getting balance
func (m *MockSuiClient) GetBalance(ctx context.Context) (uint64, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("mock get balance: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return 0, fmt.Errorf("mock get balance failed")
	}

	// Use custom function if provided
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balance, nil
}

// GetCoins mocks getting coins
func (m *MockSuiClient) GetCoins(ctx context.Context, coinType string) ([]map[string]interface{}, error) {
	// Check for context cancellation or timeout
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("mock get coins: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return nil, fmt.Errorf("mock get coins failed")
	}

	// Use custom function if provided
	if m.GetCoinsFunc != nil {
		return m.GetCoinsFunc(ctx, coinType)
	}

	if coinType == "" {
		return nil, fmt.Errorf("coin type is required")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.coins, nil
}

// Additional helper methods for testing

// GetMockTransaction retrieves a stored mock transaction
func (m *MockSuiClient) GetMockTransaction(digest string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tx, exists := m.transactions[digest]
	return tx, exists
}

// ClearTransactions clears all stored mock transactions
func (m *MockSuiClient) ClearTransactions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions = make(map[string]interface{})
}

// GetTransactionCount returns the number of stored mock transactions
func (m *MockSuiClient) GetTransactionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.transactions)
}

// SetCustomResponse sets a custom response for testing
func (m *MockSuiClient) SetCustomResponse(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[key] = value
}

// GetCustomResponse gets a custom response for testing
func (m *MockSuiClient) GetCustomResponse(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.responses[key]
	return value, exists
}
