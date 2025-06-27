package models

type PayWinner struct {
	RequesterAddress string `bson:"requester_address"`
	AccepterAddress  string `bson:"accepter_address"`
	RequesterScore   uint64 `bson:"requester_score"`
	AccepterScore    uint64 `bson:"accepter_score"`
	Winner           string `bson:"winner"`
	PrizeAmount      uint64 `bson:"prize_amount"`
	APIFee           uint64 `bson:"api_fee"`
	EscrowFee        uint64 `bson:"escrow_fee"`
	TotalStake       uint64 `bson:"total_stake"`
	Timestamp        int64  `bson:"timestamp"`
	TransactionHash  string `bson:"transaction_hash,omitempty"`
	StakeAmount      uint64 `bson:"stake_amount"`
}
