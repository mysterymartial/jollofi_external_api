package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"jollfi-gaming-api/internal/dto/request"
	"jollfi-gaming-api/internal/mocks"
	"jollfi-gaming-api/internal/service"
)

func TestGameService_StakeGame_Success(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	mockSuiClient.ExternalStakeFunc = func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
		if requesterCoinID != "0xcoin123" {
			t.Errorf("Expected requester coin ID '0xcoin123', got %s", requesterCoinID)
		}
		if accepterCoinID != "0xcoin456" {
			t.Errorf("Expected accepter coin ID '0xcoin456', got %s", accepterCoinID)
		}
		if amount != 100 {
			t.Errorf("Expected stake amount 100, got %d", amount)
		}
		return "mock_stake_digest", nil
	}

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)
	req := &request.StakeRequest{
		RequesterCoinID:  "0xcoin123",
		AccepterCoinID:   "0xcoin456",
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		StakeAmount:      100,
	}

	resp, err := gameService.StakeGame(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected success=true, got %v", resp.Success)
	}
	if resp.TransactionDigest != "mock_stake_digest" {
		t.Errorf("Expected transaction digest 'mock_stake_digest', got %s", resp.TransactionDigest)
	}
}

func TestGameService_StakeGame_InvalidAmount(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.StakeRequest{
		RequesterCoinID:  "0xcoin123",
		AccepterCoinID:   "0xcoin456",
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		StakeAmount:      0,
	}
	resp, err := gameService.StakeGame(req)
	if err == nil {
		t.Errorf("Expected error for invalid stake amount, got nil")
	}
	if resp.Success {
		t.Errorf("Expected success=false, got %v", resp.Success)
	}
}

func TestGameService_StakeGame_BlockchainFailure(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	mockSuiClient.SetShouldFail(true)

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.StakeRequest{
		RequesterCoinID:  "0xcoin123",
		AccepterCoinID:   "0xcoin456",
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		StakeAmount:      100,
	}

	resp, err := gameService.StakeGame(req)
	if err == nil {
		t.Errorf("Expected error for blockchain failure, got nil")
	}
	if resp.Success {
		t.Errorf("Expected success=false for blockchain failure, got %v", resp.Success)
	}
}

func TestGameService_PayWinner_Success(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	mockSuiClient.ExternalPayWinnerFunc = func(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error) {
		if requesterAddress != "0x123" {
			t.Errorf("Expected requester address '0x123', got %s", requesterAddress)
		}
		return "mock_pay_winner_digest", nil
	}

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.PayWinnerRequest{
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		RequesterScore:   10,
		AccepterScore:    5,
		StakeAmount:      100,
	}

	resp, err := gameService.PayWinner(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected success=true, got %v", resp.Success)
	}
	if resp.TransactionDigest != "mock_pay_winner_digest" {
		t.Errorf("Expected transaction digest 'mock_pay_winner_digest', got %s", resp.TransactionDigest)
	}
}

func TestGameService_PayWinner_InvalidRequest(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.PayWinnerRequest{
		RequesterAddress: "", // Missing address
		AccepterAddress:  "0x456",
		RequesterScore:   10,
		AccepterScore:    5,
		StakeAmount:      100,
	}

	resp, err := gameService.PayWinner(req)
	if err == nil {
		t.Errorf("Expected error for missing address, got nil")
	}

	if resp.Success {
		t.Errorf("Expected success=false for missing address, got %v", resp.Success)
	}
}

func TestGameService_PayWinner_BlockchainFailure(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	mockSuiClient.SetShouldFail(true)

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.PayWinnerRequest{
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		RequesterScore:   10,
		AccepterScore:    5,
		StakeAmount:      100,
	}

	resp, err := gameService.PayWinner(req)
	if err == nil {
		t.Errorf("Expected error for blockchain failure, got nil")
	}
	if resp.Success {
		t.Errorf("Expected success=false for blockchain failure, got %v", resp.Success)
	}
	if !strings.Contains(resp.Error, "Blockchain transaction failed") {
		t.Errorf("Expected error to contain 'Blockchain transaction failed', got %s", resp.Error)
	}
}

func TestGameService_GetStakeHistory_Success(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	resp, err := gameService.GetStakeHistory("0x123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected success=true, got %v", resp.Success)
	}

	if resp.Count != 0 {
		t.Errorf("Expected count 0, got %d", resp.Count)
	}
	if len(resp.Stakes) != 0 {
		t.Errorf("Expected 0 stakes, got %d", len(resp.Stakes))
	}
}

func TestGameService_GetStakeHistory_EmptyAddress(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	resp, err := gameService.GetStakeHistory("")
	if err == nil {
		t.Errorf("Expected error for empty address, got nil")
	}
	if resp.Success {
		t.Errorf("Expected success=false for empty address, got %v", resp.Success)
	}
	if resp.Error != "Address is required" {
		t.Errorf("Expected error 'Address is required', got %s", resp.Error)
	}
}

func TestGameService_GetGameHistory_Success(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	resp, err := gameService.GetGameHistory("0x123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !resp.Success {
		t.Errorf("Expected success=true, got %v", resp.Success)
	}
	if resp.Count != 0 {
		t.Errorf("Expected count 0, got %d", resp.Count)
	}
	if len(resp.Games) != 0 {
		t.Errorf("Expected 0 games, got %d", len(resp.Games))
	}
}

func TestGameService_GetGameHistory_EmptyAddress(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	resp, err := gameService.GetGameHistory("")
	if err == nil {
		t.Errorf("Expected error for empty address, got nil")
	}
	if resp.Success {
		t.Errorf("Expected success=false for empty address, got %v", resp.Success)
	}
	if resp.Error != "Address is required" {
		t.Errorf("Expected error 'Address is required', got %s", resp.Error)
	}
}

func TestGameService_StakeGame_ValidationScenarios(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	testCases := []struct {
		name          string
		req           *request.StakeRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "Missing RequesterCoinID",
			req: &request.StakeRequest{
				AccepterCoinID:   "0xcoin456",
				RequesterAddress: "0x123",
				AccepterAddress:  "0x456",
				StakeAmount:      100,
			},
			expectError:   true,
			errorContains: "missing required fields",
		},
		{
			name: "Missing AccepterCoinID",
			req: &request.StakeRequest{
				RequesterCoinID:  "0xcoin123",
				RequesterAddress: "0x123",
				AccepterAddress:  "0x456",
				StakeAmount:      100,
			},
			expectError:   true,
			errorContains: "missing required fields",
		},
		{
			name: "Missing RequesterAddress",
			req: &request.StakeRequest{
				RequesterCoinID: "0xcoin123",
				AccepterCoinID:  "0xcoin456",
				AccepterAddress: "0x456",
				StakeAmount:     100,
			},
			expectError:   true,
			errorContains: "addresses are required",
		},
		{
			name: "Missing AccepterAddress",
			req: &request.StakeRequest{
				RequesterCoinID:  "0xcoin123",
				AccepterCoinID:   "0xcoin456",
				RequesterAddress: "0x123",
				StakeAmount:      100,
			},
			expectError:   true,
			errorContains: "addresses are required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := gameService.StakeGame(tc.req)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
				if resp.Success {
					t.Errorf("Expected success=false for %s, got %v", tc.name, resp.Success)
				}
				if tc.errorContains != "" && !strings.Contains(resp.Error, tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got %s", tc.errorContains, resp.Error)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got %v", tc.name, err)
				}
				if !resp.Success {
					t.Errorf("Expected success=true for %s, got %v", tc.name, resp.Success)
				}
			}
		})
	}
}

func TestGameService_PayWinner_ValidationScenarios(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()
	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	testCases := []struct {
		name          string
		req           *request.PayWinnerRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "Missing RequesterAddress",
			req: &request.PayWinnerRequest{
				AccepterAddress: "0x456",
				RequesterScore:  10,
				AccepterScore:   5,
				StakeAmount:     100,
			},
			expectError:   true,
			errorContains: "missing required fields",
		},
		{
			name: "Missing AccepterAddress",
			req: &request.PayWinnerRequest{
				RequesterAddress: "0x123",
				RequesterScore:   10,
				AccepterScore:    5,
				StakeAmount:      100,
			},
			expectError:   true,
			errorContains: "missing required fields",
		},
		{
			name: "Zero StakeAmount",
			req: &request.PayWinnerRequest{
				RequesterAddress: "0x123",
				AccepterAddress:  "0x456",
				RequesterScore:   10,
				AccepterScore:    5,
				StakeAmount:      0,
			},
			expectError:   true,
			errorContains: "missing required fields",
		},
		{
			name: "Valid Request",
			req: &request.PayWinnerRequest{
				RequesterAddress: "0x123",
				AccepterAddress:  "0x456",
				RequesterScore:   10,
				AccepterScore:    5,
				StakeAmount:      100,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := gameService.PayWinner(tc.req)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
				if resp.Success {
					t.Errorf("Expected success=false for %s, got %v", tc.name, resp.Success)
				}
				if tc.errorContains != "" && !strings.Contains(resp.Error, tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got %s", tc.errorContains, resp.Error)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got %v", tc.name, err)
				}
				if !resp.Success {
					t.Errorf("Expected success=true for %s, got %v", tc.name, resp.Success)
				}
			}
		})
	}
}

func TestGameService_CustomMockBehavior(t *testing.T) {
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	mockSuiClient.ExternalStakeFunc = func(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
		return fmt.Sprintf("custom_stake_%s_%s_%d", requesterCoinID, accepterCoinID, amount), nil
	}

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)

	req := &request.StakeRequest{
		RequesterCoinID:  "coin1",
		AccepterCoinID:   "coin2",
		RequesterAddress: "0x123",
		AccepterAddress:  "0x456",
		StakeAmount:      500,
	}

	resp, err := gameService.StakeGame(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedDigest := "custom_stake_coin1_coin2_500"
	if resp.TransactionDigest != expectedDigest {
		t.Errorf("Expected transaction digest '%s', got %s", expectedDigest, resp.TransactionDigest)
	}
}
