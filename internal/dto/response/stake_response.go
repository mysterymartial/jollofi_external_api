package response

import "jollfi-gaming-api/internal/models"

type StakeResponse struct {
	Success           bool   `json:"success"`
	TransactionDigest string `json:"transaction_digest,omitempty"`
	Message           string `json:"message,omitempty"`
	Error             string `json:"error,omitempty"`
}

type StakeHistoryResponse struct {
	Success bool           `json:"success"`
	Stakes  []models.Stake `json:"stakes,omitempty"`
	Count   int            `json:"count"`
	Error   string         `json:"error,omitempty"`
}
