package models

type Stake struct {
	ID               string `bson:"_id,omitempty" json:"id,omitempty"`
	RequesterCoinID  string `bson:"requester_coin_id" json:"requester_coin_id"`
	AccepterCoinID   string `bson:"accepter_coin_id" json:"accepter_coin_id"`
	RequesterAddress string `bson:"requester_address" json:"requester_address"`
	AccepterAddress  string `bson:"accepter_address" json:"accepter_address"`
	StakeAmount      uint64 `bson:"stake_amount" json:"stake_amount"`
	Status           string `bson:"status" json:"status"`
	Timestamp        int64  `bson:"timestamp" json:"timestamp"`
	TransactionHash  string `bson:"transaction_hash,omitempty" json:"transaction_hash,omitempty"`
}
