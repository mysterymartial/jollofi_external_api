package mocks

import (
	"context"
	"fmt"
	"jollfi-gaming-api/internal/interfaces"
	"sync"
	"time"
)

var _ interfaces.SuiClientInterface = (*MockSuiClient)(nil)

type MockSuiClient struct {
	shouldFail   bool
	responses    map[string]interface{}
	balance      uint64
	coins        []map[string]interface{}
	transactions map[string]interface{}
	mu           sync.RWMutex

	ExternalStakeFunc     func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error)
	ExternalPayWinnerFunc func(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error)
	GetBalanceFunc        func(ctx context.Context) (uint64, error)
	GetCoinsFunc          func(ctx context.Context, coinType string) ([]map[string]interface{}, error)
}

func NewMockSuiClient() *MockSuiClient {
	return &MockSuiClient{
		shouldFail:   false,
		responses:    make(map[string]interface{}),
		balance:      1000000,
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

func (m *MockSuiClient) SetShouldFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = fail
}

func (m *MockSuiClient) SetBalance(balance uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balance = balance
}

func (m *MockSuiClient) SetCoins(coins []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coins = coins
}

func (m *MockSuiClient) AddTransaction(digest string, tx interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions[digest] = tx
}

func (m *MockSuiClient) ExternalStake(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("mock external stake: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return "", fmt.Errorf("mock external stake failed")
	}

	if m.ExternalStakeFunc != nil {
		return m.ExternalStakeFunc(requesterCoinID, accepterCoinID, amount, ctx)
	}

	if requesterCoinID == "" || accepterCoinID == "" {
		return "", fmt.Errorf("both coin IDs are required")
	}
	if amount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	digest := fmt.Sprintf("mock_stake_tx_%s_%s_%d", requesterCoinID, accepterCoinID, amount)

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

func (m *MockSuiClient) ExternalPayWinner(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("mock external pay winner: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return "", fmt.Errorf("mock external pay winner failed")
	}

	if m.ExternalPayWinnerFunc != nil {
		return m.ExternalPayWinnerFunc(requesterAddress, accepterAddress, requesterScore, accepterScore, stakeAmount, ctx)
	}

	if requesterAddress == "" || accepterAddress == "" {
		return "", fmt.Errorf("both addresses are required")
	}
	if stakeAmount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	var winner string
	if requesterScore > accepterScore {
		winner = requesterAddress
	} else if accepterScore > requesterScore {
		winner = accepterAddress
	} else {
		winner = "tie"
	}

	totalStake := stakeAmount * 2
	apiFee := totalStake * 8 / 100
	escrowFee := totalStake * 2 / 100
	prizeAmount := totalStake - apiFee - escrowFee

	digest := fmt.Sprintf("mock_pay_winner_tx_%s_%s_%d", requesterAddress, accepterAddress, stakeAmount)

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
		"timestamp":        time.Now().Unix(),
	})

	return digest, nil
}

func (m *MockSuiClient) ExecuteTransactionBlock(ctx context.Context, txBytes []byte) (string, error) {
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

	digest := fmt.Sprintf("mock_execute_tx_%x", len(txBytes))

	m.AddTransaction(digest, map[string]interface{}{
		"type":    "execute_transaction_block",
		"txBytes": txBytes,
		"status":  "success",
	})

	return digest, nil
}

func (m *MockSuiClient) GetTransactionBlock(ctx context.Context, digest string) (interface{}, error) {
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

	m.mu.RLock()
	defer m.mu.RUnlock()
	if tx, exists := m.transactions[digest]; exists {
		return tx, nil
	}

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

func (m *MockSuiClient) BuildTransactionBlock(ctx context.Context, params interface{}) ([]byte, error) {
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

	mockBytes := []byte(fmt.Sprintf("mock_tx_bytes_%v", params))

	return mockBytes, nil
}

func (m *MockSuiClient) GetBalance(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("mock get balance: %w", err)
	}
	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return 0, fmt.Errorf("mock get balance failed")
	}

	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balance, nil
}

func (m *MockSuiClient) GetCoins(ctx context.Context, coinType string) ([]map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("mock get coins: %w", err)
	}

	if m.shouldFail {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return nil, fmt.Errorf("mock get coins failed")
	}

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

func (m *MockSuiClient) GetMockTransaction(digest string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tx, exists := m.transactions[digest]
	return tx, exists
}

func (m *MockSuiClient) ClearTransactions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions = make(map[string]interface{})
}

func (m *MockSuiClient) GetTransactionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.transactions)
}

func (m *MockSuiClient) SetCustomResponse(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[key] = value
}

func (m *MockSuiClient) GetCustomResponse(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.responses[key]
	return value, exists
}
