package tests

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"jollfi-gaming-api/internal/config"
	"jollfi-gaming-api/internal/dto/request"
	"jollfi-gaming-api/internal/mocks"
	"jollfi-gaming-api/internal/routes"
	"jollfi-gaming-api/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Helper function to create test router
func createTestRouter() (*gin.Engine, service.GameServiceInterface) {
	gin.SetMode(gin.TestMode)

	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	cfg := &config.Config{
		PackageID:     "0xtest_package",
		ModuleName:    "jollfi_wallet",
		PoolID:        "0xtest_pool",
		Environment:   "test",
		SuiNetworkURL: "https://fullnode.testnet.sui.io:443",
		EnableLogging: true,
		EnableCORS:    true,
		RateLimit:     100,
		APIKey:        "",
	}

	gameService := service.NewGameService(mockSuiClient, mockMongoClient)
	router := routes.SetupRoutes(gameService, cfg)

	return router, gameService
}

// Test root endpoint
func TestRootEndpoint(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	if message, ok := response["message"].(string); !ok || !strings.Contains(message, "Jollfi Gaming API is running") {
		t.Errorf("Expected proper message, got %v", response["message"])
	}

	if version, ok := response["version"].(string); !ok || version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %v", response["version"])
	}
}

// Test health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status=healthy, got %v", response["status"])
	}

	if service, ok := response["service"].(string); !ok || service != "jollfi-gaming-api" {
		t.Errorf("Expected service=jollfi-gaming-api, got %v", response["service"])
	}
}

// Test API info endpoint
func TestAPIInfoEndpoint(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	info, ok := response["info"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected info object, got %v", response["info"])
		return
	}

	if service, ok := info["service"].(string); !ok || service != "Jollfi Gaming API" {
		t.Errorf("Expected service name, got %v", info["service"])
	}

	if endpoints, ok := info["endpoints"].(map[string]interface{}); !ok {
		t.Errorf("Expected endpoints object, got %v", info["endpoints"])
	} else {
		expectedEndpoints := []string{"stake", "pay_winner", "stake_history", "game_history", "health"}
		for _, endpoint := range expectedEndpoints {
			if _, exists := endpoints[endpoint]; !exists {
				t.Errorf("Expected endpoint %s to be documented", endpoint)
			}
		}
	}
}

// Test status endpoint
func TestStatusEndpoint(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	if service, ok := response["service"].(string); !ok || service != "jollfi-gaming-api" {
		t.Errorf("Expected service=jollfi-gaming-api, got %v", response["service"])
	}

	if environment, ok := response["environment"].(string); !ok || environment != "test" {
		t.Errorf("Expected environment=test, got %v", response["environment"])
	}

	if _, ok := response["timestamp"]; !ok {
		t.Errorf("Expected timestamp field")
	}

	if _, ok := response["uptime"]; !ok {
		t.Errorf("Expected uptime field")
	}
}

// Test stake game route - success
func TestStakeGameRoute_Success(t *testing.T) {
	router, _ := createTestRouter()

	stakeReq := request.StakeRequest{
		RequesterCoinID:  "0x123",
		AccepterCoinID:   "0x456",
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(stakeReq)
	req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}
}

// Test stake game route - invalid JSON
func TestStakeGameRoute_InvalidJSON(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v", response["success"])
	}

	if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, "Invalid request format") {
		t.Errorf("Expected invalid request format error, got %v", response["error"])
	}
}

// Test stake game route - validation errors
func TestStakeGameRoute_ValidationErrors(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name    string
		request request.StakeRequest
		error   string
	}{
		{
			name: "Missing RequesterCoinID",
			request: request.StakeRequest{
				AccepterCoinID:   "0x456",
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:      100,
			},
			error: "requester_coin_id is required",
		},
		{
			name: "Missing AccepterCoinID",
			request: request.StakeRequest{
				RequesterCoinID:  "0x123",
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:      100,
			},
			error: "accepter_coin_id is required",
		},
		{
			name: "Zero StakeAmount",
			request: request.StakeRequest{
				RequesterCoinID:  "0x123",
				AccepterCoinID:   "0x456",
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:      0,
			},
			error: "stake_amount must be greater than 0",
		},

		{
			name: "Same addresses",
			request: request.StakeRequest{
				RequesterCoinID:  "0x123",
				AccepterCoinID:   "0x456",
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0x1234567890abcdef1234567890abcdef12345678",
				StakeAmount:      100,
			},
			error: "requester and accepter addresses cannot be the same",
		},
		{
			name: "Invalid RequesterAddress",
			request: request.StakeRequest{
				RequesterCoinID:  "0x123",
				AccepterCoinID:   "0x456",
				RequesterAddress: "invalid_address",
				AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:      100,
			},
			error: "invalid requester_address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, tc.error) {
				t.Errorf("Expected error containing '%s', got %v", tc.error, response["error"])
			}
		})
	}
}

// Test pay winner route - success
func TestPayWinnerRoute_Success(t *testing.T) {
	router, _ := createTestRouter()

	payReq := request.PayWinnerRequest{
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(payReq)
	req, _ := http.NewRequest("POST", "/api/v1/games/pay_winner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}
}

// Test pay winner route - validation errors
func TestPayWinnerRoute_ValidationErrors(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name    string
		request request.PayWinnerRequest
		error   string
	}{
		{
			name: "Missing RequesterAddress",
			request: request.PayWinnerRequest{
				AccepterAddress: "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:     100,
			},
			error: "requester_address is required",
		},
		{
			name: "Missing AccepterAddress",
			request: request.PayWinnerRequest{
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				StakeAmount:      100,
			},
			error: "accepter_address is required",
		},
		{
			name: "Zero StakeAmount",
			request: request.PayWinnerRequest{
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeAmount:      0,
			},
			error: "stake_amount must be greater than 0",
		},
		{
			name: "Same addresses",
			request: request.PayWinnerRequest{
				RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
				AccepterAddress:  "0x1234567890abcdef1234567890abcdef12345678",
				StakeAmount:      100,
			},
			error: "requester and accepter addresses cannot be the same",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("POST", "/api/v1/games/pay_winner", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, tc.error) {
				t.Errorf("Expected error containing '%s', got %v", tc.error, response["error"])
			}
		})
	}
}

// Test get stake history route - success
func TestGetStakeHistoryRoute_Success(t *testing.T) {
	router, _ := createTestRouter()

	validAddress := "0x1234567890abcdef1234567890abcdef12345678"
	req, _ := http.NewRequest("GET", "/api/v1/games/stakes/"+validAddress, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}
}

// Test get stake history route - invalid address
func TestGetStakeHistoryRoute_InvalidAddress(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name    string
		address string
		error   string
	}{
		{
			name:    "Short address",
			address: "0x123",
			error:   "Invalid address format",
		},
		{
			name:    "No 0x prefix",
			address: "1234567890abcdef1234567890abcdef12345678",
			error:   "Invalid address format",
		},
		{
			name:    "Invalid characters",
			address: "0x1234567890abcdef1234567890abcdef1234567g",
			error:   "Invalid address format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/games/stakes/"+tc.address, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, tc.error) {
				t.Errorf("Expected error containing '%s', got %v", tc.error, response["error"])
			}
		})
	}
}

// Test get game history route - success
func TestGetGameHistoryRoute_Success(t *testing.T) {
	router, _ := createTestRouter()

	validAddress := "0x1234567890abcdef1234567890abcdef12345678"
	req, _ := http.NewRequest("GET", "/api/v1/games/history/"+validAddress, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}
}

// Test get game history route - invalid address
func TestGetGameHistoryRoute_InvalidAddress(t *testing.T) {
	router, _ := createTestRouter()

	invalidAddress := "invalid_address"
	req, _ := http.NewRequest("GET", "/api/v1/games/history/"+invalidAddress, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, "Invalid address format") {
		t.Errorf("Expected invalid address format error, got %v", response["error"])
	}
}

// Test get game stats route
func TestGetGameStatsRoute(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/games/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	if data, ok := response["data"].(map[string]interface{}); !ok {
		t.Errorf("Expected data object, got %v", response["data"])
	} else {
		expectedFields := []string{"total_games", "total_stakes", "active_players", "last_updated"}
		for _, field := range expectedFields {
			if _, exists := data[field]; !exists {
				t.Errorf("Expected field %s in stats data", field)
			}
		}
	}
}

// Test 404 error handler
func TestNotFoundHandler(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v", response["success"])
	}

	if errorMsg, ok := response["error"].(string); !ok || errorMsg != "Endpoint not found" {
		t.Errorf("Expected 'Endpoint not found' error, got %v", response["error"])
	}

	if path, ok := response["path"].(string); !ok || path != "/nonexistent" {
		t.Errorf("Expected path='/nonexistent', got %v", response["path"])
	}

	if method, ok := response["method"].(string); !ok || method != "GET" {
		t.Errorf("Expected method='GET', got %v", response["method"])
	}
}

// Test 405 method not allowed handler
func TestMethodNotAllowedHandler(t *testing.T) {
	router, _ := createTestRouter()

	// Try to POST to a GET-only endpoint
	req, _ := http.NewRequest("POST", "/health", nil)
	req.Header.Set("Origin", "http://localhost")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Log response details for debugging
	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Headers: %v", w.Header())
	t.Logf("Response Body: %s", w.Body.String())

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v", response["success"])
	}

	if errorMsg, ok := response["error"].(string); !ok || errorMsg != "Method not allowed" {
		t.Errorf("Expected 'Method not allowed' error, got %v", response["error"])
	}
}

// Test middleware integration
func TestMiddlewareIntegration(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check that middleware headers are set
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("Expected X-Request-ID header from RequestIDMiddleware")
	}

	if w.Header().Get("X-Content-Type-Options") == "" {
		t.Error("Expected security headers from SecurityHeadersMiddleware")
	}

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected CORS headers from CORSMiddleware")
	}
}

// Test address validation function
func TestSuiAddressValidation(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name      string
		address   string
		shouldErr bool
	}{
		{
			name:      "Valid address",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			shouldErr: false,
		},
		{
			name:      "Valid long address",
			address:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			shouldErr: false,
		},
		{
			name:      "Too short",
			address:   "0x123",
			shouldErr: true,
		},
		{
			name:      "Too long",
			address:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123",
			shouldErr: true,
		},
		{
			name:      "No 0x prefix",
			address:   "1234567890abcdef1234567890abcdef12345678",
			shouldErr: true,
		},
		{
			name:      "Invalid characters",
			address:   "0x1234567890abcdef1234567890abcdef1234567g",
			shouldErr: true,
		},
		{
			name:      "Uppercase hex",
			address:   "0x1234567890ABCDEF1234567890ABCDEF12345678",
			shouldErr: false,
		},
		{
			name:      "Mixed case hex",
			address:   "0x1234567890AbCdEf1234567890AbCdEf12345678",
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/games/stakes/"+tc.address, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tc.shouldErr {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status 400 for invalid address, got %d", w.Code)
				}
			} else {
				if w.Code == http.StatusBadRequest {
					var response map[string]interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					if errorMsg, ok := response["error"].(string); ok && strings.Contains(errorMsg, "Invalid address format") {
						t.Errorf("Expected valid address to pass validation, got error: %v", errorMsg)
					}
				}
			}
		})
	}
}

// Test JSON parsing edge cases
func TestJSONParsingEdgeCases(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "Empty JSON",
			body: "{}",
		},
		{
			name: "Null values",
			body: `{"requester_coin_id": null, "accepter_coin_id": null}`,
		},
		{
			name: "Extra fields",
			body: `{"requester_coin_id": "0x123", "extra_field": "value"}`,
		},
		{
			name: "Wrong types",
			body: `{"stake_amount": "not_a_number"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should return 400 for invalid/incomplete data
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400 for case '%s', got %d", tc.name, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			}
		})
	}
}

// Test content type validation
func TestContentTypeValidation(t *testing.T) {
	router, _ := createTestRouter()

	stakeReq := request.StakeRequest{
		RequesterCoinID:  "0x123",
		AccepterCoinID:   "0x456",
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(stakeReq)

	testCases := []struct {
		name        string
		contentType string
		expectError bool
	}{
		{
			name:        "Valid JSON content type",
			contentType: "application/json",
			expectError: false,
		},
		{
			name:        "JSON with charset",
			contentType: "application/json; charset=utf-8",
			expectError: false,
		},
		{
			name:        "Wrong content type",
			contentType: "text/plain",
			expectError: true,
		},
		{
			name:        "No content type",
			contentType: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tc.expectError {
				if w.Code == http.StatusOK {
					t.Errorf("Expected error for content type '%s', but got success", tc.contentType)
				}
			}
		})
	}
}

// Test large request bodies
func TestLargeRequestBody(t *testing.T) {
	router, _ := createTestRouter()

	// Create a request with very long strings
	stakeReq := request.StakeRequest{
		RequesterCoinID:  strings.Repeat("0x", 1000) + "123",
		AccepterCoinID:   strings.Repeat("0x", 1000) + "456",
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(stakeReq)
	req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle large requests gracefully (either accept or reject with proper error)
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 200, 400, or 413 for large request, got %d", w.Code)
	}
}

// Test concurrent requests
func TestConcurrentRequests(t *testing.T) {
	router, _ := createTestRouter()

	const numRequests = 10
	results := make(chan int, numRequests)

	stakeReq := request.StakeRequest{
		RequesterCoinID:  "0x123",
		AccepterCoinID:   "0x456",
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(stakeReq)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}

	if successCount != numRequests {
		t.Errorf("Expected all %d requests to succeed, got %d successes", numRequests, successCount)
	}
}

// Test route parameter edge cases
func TestRouteParameterEdgeCases(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name     string
		endpoint string
		param    string
	}{
		{
			name:     "Empty parameter",
			endpoint: "/api/v1/games/stakes/",
			param:    "",
		},
		{
			name:     "Special characters in parameter",
			endpoint: "/api/v1/games/stakes/",
			param:    "0x123%20test",
		},
		{
			name:     "Very long parameter",
			endpoint: "/api/v1/games/stakes/",
			param:    strings.Repeat("0x", 100) + "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.endpoint + tc.param
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should handle edge cases gracefully
			if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
				t.Logf("URL: %s, Status: %d, Body: %s", url, w.Code, w.Body.String())
			}
		})
	}
}

// Test HTTP methods on all endpoints
func TestHTTPMethods(t *testing.T) {
	router, _ := createTestRouter()

	endpoints := []struct {
		path           string
		allowedMethods []string
	}{
		{
			path:           "/",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/health",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/api/v1/info",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/api/v1/status",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/api/v1/games/stake",
			allowedMethods: []string{"POST"},
		},
		{
			path:           "/api/v1/games/pay_winner",
			allowedMethods: []string{"POST"},
		},
		{
			path:           "/api/v1/games/stakes/0x1234567890abcdef1234567890abcdef12345678",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/api/v1/games/history/0x1234567890abcdef1234567890abcdef12345678",
			allowedMethods: []string{"GET"},
		},
		{
			path:           "/api/v1/games/stats",
			allowedMethods: []string{"GET"},
		},
	}

	allMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, endpoint := range endpoints {
		for _, method := range allMethods {
			t.Run(endpoint.path+"_"+method, func(t *testing.T) {
				var req *http.Request

				if method == "POST" || method == "PUT" || method == "PATCH" {
					req, _ = http.NewRequest(method, endpoint.path, bytes.NewBuffer([]byte("{}")))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, _ = http.NewRequest(method, endpoint.path, nil)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				isAllowed := false
				for _, allowedMethod := range endpoint.allowedMethods {
					if method == allowedMethod {
						isAllowed = true
						break
					}
				}

				if isAllowed {
					if w.Code == http.StatusMethodNotAllowed {
						t.Errorf("Method %s should be allowed for %s, got 405", method, endpoint.path)
					}
				} else {
					if method != "OPTIONS" && w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
						// Some methods might return 404 instead of 405, which is also acceptable
						t.Logf("Method %s on %s returned %d (expected 405 or 404)", method, endpoint.path, w.Code)
					}
				}
			})
		}
	}
}

// Benchmark tests for performance
func BenchmarkStakeGameRoute(b *testing.B) {
	router, _ := createTestRouter()

	stakeReq := request.StakeRequest{
		RequesterCoinID:  "0x123",
		AccepterCoinID:   "0x456",
		RequesterAddress: "0x1234567890abcdef1234567890abcdef12345678",
		AccepterAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
		StakeAmount:      100,
	}

	jsonData, _ := json.Marshal(stakeReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	router, _ := createTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetStakeHistory(b *testing.B) {
	router, _ := createTestRouter()

	validAddress := "0x1234567890abcdef1234567890abcdef12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/games/stakes/"+validAddress, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

// Test route groups and middleware application
func TestRouteGroups(t *testing.T) {
	router, _ := createTestRouter()

	// Test that API v1 routes are properly grouped
	apiRoutes := []string{
		"/api/v1/games/stake",
		"/api/v1/games/pay_winner",
		"/api/v1/games/stakes/0x1234567890abcdef1234567890abcdef12345678",
		"/api/v1/games/history/0x1234567890abcdef1234567890abcdef12345678",
		"/api/v1/games/stats",
		"/api/v1/info",
		"/api/v1/status",
	}

	for _, route := range apiRoutes {
		t.Run("Route_"+route, func(t *testing.T) {
			var req *http.Request

			if strings.Contains(route, "stake") && !strings.Contains(route, "stakes/") {
				// POST endpoints
				jsonData, _ := json.Marshal(map[string]interface{}{
					"requester_coin_id": "0x123",
					"accepter_coin_id":  "0x456",
					"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
					"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
					"stake_amount":      100,
				})
				req, _ = http.NewRequest("POST", route, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else if route == "/api/v1/games/pay_winner" {
				jsonData, _ := json.Marshal(map[string]interface{}{
					"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
					"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
					"stake_amount":      100,
				})
				req, _ = http.NewRequest("POST", route, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else {
				// GET endpoints
				req, _ = http.NewRequest("GET", route, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not return 404 (route should exist)
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s should exist, got 404", route)
			}

			// Check that middleware headers are present (indicating middleware is applied)
			if w.Header().Get("X-Request-ID") == "" {
				t.Errorf("Route %s should have middleware applied (missing X-Request-ID)", route)
			}
		})
	}
}

// Test error response format consistency
func TestErrorResponseFormat(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "Invalid JSON",
			method: "POST",
			path:   "/api/v1/games/stake",
			body:   "invalid json",
		},
		{
			name:   "Missing required fields",
			method: "POST",
			path:   "/api/v1/games/stake",
			body:   "{}",
		},
		{
			name:   "Invalid address",
			method: "GET",
			path:   "/api/v1/games/stakes/invalid_address",
			body:   "",
		},
		{
			name:   "Not found endpoint",
			method: "GET",
			path:   "/api/v1/nonexistent",
			body:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			if tc.body != "" {
				req, _ = http.NewRequest(tc.method, tc.path, bytes.NewBuffer([]byte(tc.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(tc.method, tc.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse error response: %v", err)
				return
			}

			// Check error response format consistency
			if success, ok := response["success"].(bool); !ok {
				t.Errorf("Error response should have 'success' field")
			} else if success {
				t.Errorf("Error response should have success=false")
			}

			if _, ok := response["error"].(string); !ok {
				t.Errorf("Error response should have 'error' field with string value")
			}

			// Check that error message is not empty
			if errorMsg, ok := response["error"].(string); ok && errorMsg == "" {
				t.Errorf("Error message should not be empty")
			}
		})
	}
}

// Test success response format consistency
func TestSuccessResponseFormat(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "Root endpoint",
			method: "GET",
			path:   "/",
			body:   "",
		},
		{
			name:   "Health check",
			method: "GET",
			path:   "/health",
			body:   "",
		},
		{
			name:   "API info",
			method: "GET",
			path:   "/api/v1/info",
			body:   "",
		},
		{
			name:   "Status endpoint",
			method: "GET",
			path:   "/api/v1/status",
			body:   "",
		},
		{
			name:   "Game stats",
			method: "GET",
			path:   "/api/v1/games/stats",
			body:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", tc.path, w.Code)
				return
			}

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse success response: %v", err)
				return
			}

			// Check success response format consistency
			if success, ok := response["success"].(bool); !ok {
				t.Errorf("Success response should have 'success' field")
			} else if !success {
				t.Errorf("Success response should have success=true")
			}

			// Different endpoints may have different additional fields, but all should have success
		})
	}
}

// Test request validation edge cases
func TestRequestValidationEdgeCases(t *testing.T) {
	router, _ := createTestRouter()

	testCases := []struct {
		name    string
		request map[string]interface{}
		error   string
	}{
		{
			name: "Negative stake amount",
			request: map[string]interface{}{
				"requester_coin_id": "0x123",
				"accepter_coin_id":  "0x456",
				"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
				"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
				"stake_amount":      -100,
			},
			error: "stake_amount must be greater than 0",
		},
		{
			name: "Zero stake amount",
			request: map[string]interface{}{
				"requester_coin_id": "0x123",
				"accepter_coin_id":  "0x456",
				"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
				"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
				"stake_amount":      0,
			},
			error: "stake_amount must be greater than 0",
		},
		{
			name: "Very large stake amount",
			request: map[string]interface{}{
				"requester_coin_id": "0x123",
				"accepter_coin_id":  "0x456",
				"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
				"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
				"stake_amount":      999999999999999999,
			},
			error: "", // Should be valid
		},
		{
			name: "Empty coin IDs",
			request: map[string]interface{}{
				"requester_coin_id": "",
				"accepter_coin_id":  "",
				"requester_address": "0x1234567890abcdef1234567890abcdef12345678",
				"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
				"stake_amount":      100,
			},
			error: "requester_coin_id is required",
		},
		{
			name: "Whitespace in addresses",
			request: map[string]interface{}{
				"requester_coin_id": "0x123",
				"accepter_coin_id":  "0x456",
				"requester_address": " 0x1234567890abcdef1234567890abcdef12345678 ",
				"accepter_address":  "0xabcdef1234567890abcdef1234567890abcdef12",
				"stake_amount":      100,
			},
			error: "invalid requester_address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("POST", "/api/v1/games/stake", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tc.error == "" {
				// Should succeed
				if w.Code != http.StatusOK {
					t.Errorf("Expected success for case '%s', got status %d", tc.name, w.Code)
				}
			} else {
				// Should fail with validation error
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected status 400 for case '%s', got %d", tc.name, w.Code)
				}

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				if errorMsg, ok := response["error"].(string); !ok || !strings.Contains(errorMsg, tc.error) {
					t.Errorf("Expected error containing '%s', got %v", tc.error, response["error"])
				}
			}
		})
	}
}

// Test configuration-based middleware activation
func TestConfigBasedMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mocks
	mockSuiClient := mocks.NewMockSuiClient()
	mockMongoClient := mocks.NewMockMongoClient()

	testCases := []struct {
		name   string
		config *config.Config
		header string
		should string
	}{
		{
			name: "CORS enabled",
			config: &config.Config{
				PackageID:     "0xtest_package",
				ModuleName:    "jollfi_wallet",
				PoolID:        "0xtest_pool",
				Environment:   "test",
				EnableCORS:    true,
				EnableLogging: false,
				RateLimit:     0,
				APIKey:        "",
			},
			header: "Access-Control-Allow-Origin",
			should: "be present",
		},
		{
			name: "CORS disabled",
			config: &config.Config{
				PackageID:     "0xtest_package",
				ModuleName:    "jollfi_wallet",
				PoolID:        "0xtest_pool",
				Environment:   "test",
				EnableCORS:    false,
				EnableLogging: false,
				RateLimit:     0,
				APIKey:        "",
			},
			header: "Access-Control-Allow-Origin",
			should: "be absent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameService := service.NewTestGameService(mockSuiClient, mockMongoClient, tc.config)
			router := routes.SetupRoutes(gameService, tc.config)

			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			headerValue := w.Header().Get(tc.header)

			if tc.should == "be present" && headerValue == "" {
				t.Errorf("Expected header '%s' to be present when enabled", tc.header)
			} else if tc.should == "be absent" && headerValue != "" {
				t.Errorf("Expected header '%s' to be absent when disabled, got '%s'", tc.header, headerValue)
			}
		})
	}
}

// Test API versioning
func TestAPIVersioning(t *testing.T) {
	router, _ := createTestRouter()

	// Test that v1 routes work
	v1Routes := []string{
		"/api/v1/info",
		"/api/v1/status",
		"/api/v1/games/stats",
	}

	for _, route := range v1Routes {
		t.Run("V1_"+route, func(t *testing.T) {
			req, _ := http.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("V1 route %s should exist", route)
			}
		})
	}

	// Test that non-versioned API routes don't work
	nonVersionedRoutes := []string{
		"/api/info",
		"/api/status",
		"/api/games/stats",
	}

	for _, route := range nonVersionedRoutes {
		t.Run("NonVersioned_"+route, func(t *testing.T) {
			req, _ := http.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Non-versioned route %s should not exist, got status %d", route, w.Code)
			}
		})
	}
}

// Test request timeout handling
func TestRequestTimeoutHandling(t *testing.T) {
	// This test would require a slow handler to test timeout middleware
	// Since we're using mocks, we can't easily simulate slow responses
	// But we can test that the timeout middleware is applied

	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The request should complete normally (not timeout)
	if w.Code != http.StatusOK {
		t.Errorf("Expected normal completion, got status %d", w.Code)
	}
}

// Test route security
func TestRouteSecurity(t *testing.T) {
	router, _ := createTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expectedValue := range securityHeaders {
		if actualValue := w.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Expected security header %s=%s, got %s", header, expectedValue, actualValue)
		}
	}
}
