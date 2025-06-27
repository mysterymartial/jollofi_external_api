package service

import (
	"context"
	"fmt"
	"jollfi-gaming-api/internal/config"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jollfi-gaming-api/internal/dto/request"
	"jollfi-gaming-api/internal/dto/response"
	"jollfi-gaming-api/internal/interfaces"
	"jollfi-gaming-api/internal/models"
)

type GameService struct {
	suiClient   interfaces.SuiClientInterface
	mongoClient interfaces.MongoClientInterface
}

// Ensure GameService implements GameServiceInterface at compile time
var _ GameServiceInterface = (*GameService)(nil)

// Constructor now accepts interfaces instead of concrete types
func NewGameService(suiClient interfaces.SuiClientInterface, mongoClient interfaces.MongoClientInterface) *GameService {
	return &GameService{
		suiClient:   suiClient,
		mongoClient: mongoClient,
	}
}

func (s *GameService) StakeGame(req *request.StakeRequest) (*response.StakeResponse, error) {
	// Basic validation only
	if req.RequesterCoinID == "" || req.AccepterCoinID == "" || req.StakeAmount == 0 {
		return &response.StakeResponse{
			Success: false,
			Error:   "Invalid stake request: missing required fields",
		}, fmt.Errorf("invalid stake request")
	}
	if req.RequesterAddress == "" || req.AccepterAddress == "" {
		return &response.StakeResponse{
			Success: false,
			Error:   "Invalid stake request: addresses are required",
		}, fmt.Errorf("addresses are required")
	}

	// Log the fee deduction info
	log.Printf("🔄 Processing stake: Amount per player: %d SUI (10%% fee will be deducted by blockchain)", req.StakeAmount)
	log.Printf("🔄 Requester: %s, Accepter: %s", req.RequesterAddress, req.AccepterAddress)

	// Call blockchain - let it handle everything
	txDigest, err := s.suiClient.ExternalStake(
		req.RequesterCoinID,
		req.AccepterCoinID,
		req.StakeAmount,
		context.Background(),
	)
	if err != nil {
		log.Printf("❌ Blockchain stake failed: %v", err)
		return &response.StakeResponse{
			Success: false,
			Error:   fmt.Sprintf("Blockchain transaction failed: %v", err),
		}, err
	}

	// Create stake record with addresses for history tracking
	stake := models.Stake{
		RequesterCoinID:  req.RequesterCoinID,
		AccepterCoinID:   req.AccepterCoinID,
		RequesterAddress: req.RequesterAddress,
		AccepterAddress:  req.AccepterAddress,
		StakeAmount:      req.StakeAmount,
		Status:           "completed",
		Timestamp:        time.Now().Unix(),
		TransactionHash:  txDigest,
	}

	collection := s.mongoClient.GetDatabase("jollfi_games").Collection("stakes")
	_, err = collection.InsertOne(context.Background(), stake)
	if err != nil {
		log.Printf("⚠️  Database save failed (transaction still succeeded): %v", err)
	}

	log.Printf("✅ Stake transaction successful: TxDigest: %s", txDigest)
	return &response.StakeResponse{
		Success:           true,
		TransactionDigest: txDigest,
		Message:           "Stake successful. 10% fee deducted from each player by blockchain.",
	}, nil
}

func (s *GameService) PayWinner(req *request.PayWinnerRequest) (*response.PayWinnerResponse, error) {
	// Basic validation only
	if req.RequesterAddress == "" || req.AccepterAddress == "" || req.StakeAmount == 0 {
		return &response.PayWinnerResponse{
			Success: false,
			Error:   "Invalid pay winner request: missing required fields",
		}, fmt.Errorf("invalid pay winner request")
	}

	log.Printf("🔄 Processing winner payment: Requester Score: %d, Accepter Score: %d, Original Stake: %d",
		req.RequesterScore, req.AccepterScore, req.StakeAmount)

	// Call blockchain - let it handle all prize logic
	txDigest, err := s.suiClient.ExternalPayWinner(
		req.RequesterAddress,
		req.AccepterAddress,
		req.RequesterScore,
		req.AccepterScore,
		req.StakeAmount,
		context.Background(),
	)
	if err != nil {
		log.Printf("❌ Blockchain pay winner failed: %v", err)
		return &response.PayWinnerResponse{
			Success: false,
			Error:   fmt.Sprintf("Blockchain transaction failed: %v", err),
		}, err
	}

	// Create pay winner record
	payWinner := models.PayWinner{
		RequesterAddress: req.RequesterAddress,
		AccepterAddress:  req.AccepterAddress,
		RequesterScore:   req.RequesterScore,
		AccepterScore:    req.AccepterScore,
		StakeAmount:      req.StakeAmount,
		Timestamp:        time.Now().Unix(),
		TransactionHash:  txDigest,
	}

	collection := s.mongoClient.GetDatabase("jollfi_games").Collection("pay_winners")
	_, err = collection.InsertOne(context.Background(), payWinner)
	if err != nil {
		log.Printf("⚠️  Database save failed (transaction still succeeded): %v", err)
	}

	log.Printf("✅ Pay winner transaction successful: TxDigest: %s", txDigest)
	return &response.PayWinnerResponse{
		Success:           true,
		TransactionDigest: txDigest,
		Message:           "Winner payment processed by blockchain with all fees calculated automatically.",
	}, nil
}

func (s *GameService) GetStakeHistory(address string) (*response.StakeHistoryResponse, error) {
	if address == "" {
		return &response.StakeHistoryResponse{
			Success: false,
			Error:   "Address is required",
		}, fmt.Errorf("address is required")
	}

	collection := s.mongoClient.GetDatabase("jollfi_games").Collection("stakes")
	// Find stakes where user participated (either as requester or accepter)
	filter := bson.M{
		"$or": []bson.M{
			{"requester_address": address},
			{"accepter_address": address},
		},
	}

	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(50) // Latest 50 stakes
	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Printf("❌ Failed to fetch stake history for %s: %v", address, err)
		return &response.StakeHistoryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch stake history: %v", err),
		}, err
	}
	defer cursor.Close(context.Background())

	var stakes []models.Stake
	if err = cursor.All(context.Background(), &stakes); err != nil {
		log.Printf("❌ Failed to decode stakes for %s: %v", address, err)
		return &response.StakeHistoryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode stakes: %v", err),
		}, err
	}

	log.Printf("✅ Retrieved %d stakes for address: %s", len(stakes), address)
	return &response.StakeHistoryResponse{
		Success: true,
		Stakes:  stakes,
		Count:   len(stakes),
	}, nil
}

func (s *GameService) GetGameHistory(address string) (*response.GameHistoryResponse, error) {
	if address == "" {
		return &response.GameHistoryResponse{
			Success: false,
			Error:   "Address is required",
		}, fmt.Errorf("address is required")
	}

	collection := s.mongoClient.GetDatabase("jollfi_games").Collection("pay_winners")
	// Find games where user participated (either as requester or accepter)
	filter := bson.M{
		"$or": []bson.M{
			{"requester_address": address},
			{"accepter_address": address},
		},
	}

	opts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(50) // Latest 50 games
	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		log.Printf("❌ Failed to fetch game history for %s: %v", address, err)
		return &response.GameHistoryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch game history: %v", err),
		}, err
	}
	defer cursor.Close(context.Background())

	var games []models.PayWinner
	if err = cursor.All(context.Background(), &games); err != nil {
		log.Printf("❌ Failed to decode games for %s: %v", address, err)
		return &response.GameHistoryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to decode games: %v", err),
		}, err
	}

	log.Printf("✅ Retrieved %d games for address: %s", len(games), address)
	return &response.GameHistoryResponse{
		Success: true,
		Games:   games,
		Count:   len(games),
	}, nil
}

// test function
func NewTestGameService(suiClient interfaces.SuiClientInterface, mongoClient interfaces.MongoClientInterface, cfg *config.Config) GameServiceInterface {
	return &GameService{
		suiClient:   suiClient,
		mongoClient: mongoClient,
	}
}
