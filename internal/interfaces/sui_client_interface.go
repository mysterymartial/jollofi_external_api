package interfaces

import "context"

// SuiClientInterface defines the interfaces for SUI blockchain client
type SuiClientInterface interface {
	ExternalStake(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error)
	ExternalPayWinner(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error)
	ExecuteTransactionBlock(ctx context.Context, txBytes []byte) (string, error)
	GetTransactionBlock(ctx context.Context, digest string) (interface{}, error)
	BuildTransactionBlock(ctx context.Context, params interface{}) ([]byte, error)
	GetBalance(ctx context.Context) (uint64, error)
	GetCoins(ctx context.Context, coinType string) ([]map[string]interface{}, error)
}
