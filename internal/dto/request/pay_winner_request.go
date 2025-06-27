package request

type PayWinnerRequest struct {
	RequesterAddress string `bson:"requester_address"`
	AccepterAddress  string `bson:"accepter_address"`
	RequesterScore   uint64 `bson:"requester_score"`
	AccepterScore    uint64 `bson:"accepter_score"`
	StakeAmount      uint64 `bson:"stake_amount"`
	Timestamp        int64  `bson:"timestamp"`
	TransactionHash  string `bson:"transaction_hash,omitempty"`
}
