package mocks

import (
	"jollfi-gaming-api/internal/data"
	"jollfi-gaming-api/internal/service"
	"unsafe"
)

type GameServiceMock struct {
	*service.GameService
	MockSuiClient   *MockSuiClient
	MockMongoClient *MockMongoClient
}

// this is for future changes i will later use this mock
// instead of using the services directly
func NewGameServiceMock() *GameServiceMock {
	mockSui := NewMockSuiClient()
	mockMongo := NewMockMongoClient()

	suiClient := (*data.SuiClient)(unsafe.Pointer(mockSui))
	mongoClient := (*data.MongoClient)(unsafe.Pointer(mockMongo))

	gameService := service.NewGameService(suiClient, mongoClient)

	return &GameServiceMock{
		GameService:     gameService,
		MockSuiClient:   mockSui,
		MockMongoClient: mockMongo,
	}
}
