package response

import "jollfi-gaming-api/internal/models"

type PayWinnerResponse struct {
	Success           bool   `json:"success"`
	TransactionDigest string `json:"transaction_digest,omitempty"`
	Message           string `json:"message,omitempty"`
	Error             string `json:"error,omitempty"`
}

type GameHistoryResponse struct {
	Success bool               `json:"success"`
	Games   []models.PayWinner `json:"games,omitempty"`
	Count   int                `json:"count"`
	Error   string             `json:"error,omitempty"`
}
