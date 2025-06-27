package request

type StakeRequest struct {
	RequesterCoinID  string `json:"requester_coin_id" binding:"required"`
	AccepterCoinID   string `json:"accepter_coin_id" binding:"required"`
	RequesterAddress string `json:"requester_address" binding:"required"`
	AccepterAddress  string `json:"accepter_address" binding:"required"`
	StakeAmount      uint64 `json:"stake_amount" binding:"required,min=1"`
}
