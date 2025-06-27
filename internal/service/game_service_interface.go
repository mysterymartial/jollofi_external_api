package service

import (
	"jollfi-gaming-api/internal/dto/request"
	"jollfi-gaming-api/internal/dto/response"
)

type GameServiceInterface interface {
	StakeGame(req *request.StakeRequest) (*response.StakeResponse, error)
	PayWinner(req *request.PayWinnerRequest) (*response.PayWinnerResponse, error)
	GetStakeHistory(address string) (*response.StakeHistoryResponse, error)
	GetGameHistory(address string) (*response.GameHistoryResponse, error)
} // ✅ Removed the trailing comma
