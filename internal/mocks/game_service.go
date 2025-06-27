package mocks

import (
	"jollfi-gaming-api/internal/data"
	"jollfi-gaming-api/internal/service"
	"unsafe"
)

// GameServiceMock wraps the real service with mock dependencies
type GameServiceMock struct {
	*service.GameService
	MockSuiClient   *MockSuiClient
	MockMongoClient *MockMongoClient
}

// NewGameServiceMock creates a new GameService with mocks
func NewGameServiceMock() *GameServiceMock {
	mockSui := NewMockSuiClient()
	mockMongo := NewMockMongoClient()

	// Create real service with mock clients (using unsafe conversion)
	suiClient := (*data.SuiClient)(unsafe.Pointer(mockSui))
	mongoClient := (*data.MongoClient)(unsafe.Pointer(mockMongo))

	gameService := service.NewGameService(suiClient, mongoClient)

	return &GameServiceMock{
		GameService:     gameService,
		MockSuiClient:   mockSui,
		MockMongoClient: mockMongo,
	}
}
