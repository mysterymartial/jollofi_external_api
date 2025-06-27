package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"jollfi-gaming-api/internal/config"
	"jollfi-gaming-api/internal/dto/request"
	"jollfi-gaming-api/internal/dto/response"
	"jollfi-gaming-api/internal/middleware"
	"jollfi-gaming-api/internal/service"
)

// Global variable to track start time
var startTime = time.Now()

// SetupRoutes configures all application routes
// @title Jollfi Gaming API
// @version 1.0
// @description API for gaming on the Sui blockchain
// @host localhost:8080
// @BasePath /api/v1
func SetupRoutes(gameService service.GameServiceInterface, cfg *config.Config) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	})
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	setupMiddleware(r, cfg)
	setupAPIRoutes(r, gameService, cfg)
	setupUtilityRoutes(r, cfg)
	setupErrorHandlers(r)
	return r
}

// setupMiddleware configures all middleware
func setupMiddleware(r *gin.Engine, cfg *config.Config) {
	if cfg.EnableLogging {
		r.Use(middleware.LoggerMiddleware())
	}
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(func(c *gin.Context) {
		log.Printf("SecurityHeadersMiddleware triggered for %s %s", c.Request.Method, c.Request.URL.Path)
		middleware.SecurityHeadersMiddleware()(c)
		c.Next()
	})
	if cfg.EnableCORS {
		r.Use(func(c *gin.Context) {
			log.Printf("CORSMiddleware triggered for %s %s, Origin: %s", c.Request.Method, c.Request.URL.Path, c.Request.Header.Get("Origin"))
			middleware.CORSMiddleware()(c)
			if c.IsAborted() {
				log.Printf("CORSMiddleware aborted request for %s %s", c.Request.Method, c.Request.URL.Path)
			}
			c.Next()
		})
	}
	if cfg.RateLimit > 0 {
		r.Use(middleware.RateLimitMiddleware(cfg.RateLimit))
	}
	if cfg.APIKey != "" {
		r.Use(middleware.APIKeyMiddleware(cfg.APIKey))
	}
	r.Use(middleware.MetricsMiddleware())
}

// setupAPIRoutes configures API routes
func setupAPIRoutes(r *gin.Engine, gameService service.GameServiceInterface, cfg *config.Config) {
	api := r.Group("/api/v1")
	{
		gamesWithValidation := api.Group("/games")
		gamesWithValidation.Use(middleware.ValidationMiddleware())
		{
			gamesWithValidation.POST("/pay_winner", handlePayWinner(gameService))
			gamesWithValidation.GET("/stakes/:address", handleGetStakeHistory(gameService))
			gamesWithValidation.GET("/history/:address", handleGetGameHistory(gameService))
			gamesWithValidation.GET("/stats", handleGetGameStats(gameService))
		}
		gamesWithoutValidation := api.Group("/games")
		{
			gamesWithoutValidation.POST("/stake", handleStakeGame(gameService))
		}
	}
}

// setupUtilityRoutes configures utility and system routes
func setupUtilityRoutes(r *gin.Engine, cfg *config.Config) {
	r.GET("/health", middleware.HealthCheckHandler())
	r.POST("/health", func(c *gin.Context) {
		log.Printf("Explicit POST handler for /health triggered")
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"error":   "Method not allowed",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
	r.POST("/health/", func(c *gin.Context) {
		log.Printf("Explicit POST handler for /health/ triggered")
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"error":   "Method not allowed",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Jollfi Gaming API is running",
			"version": "1.0.0",
			"docs":    "/api/v1/info",
		})
	})
	r.GET("/api/v1/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"info": gin.H{
				"service":     "Jollfi Gaming API",
				"version":     "1.0.0",
				"environment": cfg.Environment,
				"network":     cfg.SuiNetworkURL,
				"module":      cfg.ModuleName,
				"endpoints": gin.H{
					"stake":         "POST /api/v1/games/stake",
					"pay_winner":    "POST /api/v1/games/pay_winner",
					"stake_history": "GET /api/v1/games/stakes/:address",
					"game_history":  "GET /api/v1/games/history/:address",
					"stats":         "GET /api/v1/games/stats",
					"health":        "GET /health",
				},
			},
		})
	})
	r.GET("/api/v1/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"service":     "jollfi-gaming-api",
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"timestamp":   time.Now().Unix(),
			"uptime":      time.Since(startTime).String(),
		})
	})
}

// setupErrorHandlers configures error handling routes
func setupErrorHandlers(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		log.Printf("NoRoute triggered for path: %s, method: %s", c.Request.URL.Path, c.Request.Method)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Endpoint not found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
	r.NoMethod(func(c *gin.Context) {
		log.Printf("NoMethod triggered for path: %s, method: %s", c.Request.URL.Path, c.Request.Method)
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"error":   "Method not allowed",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
}

// Handler functions
type tempStakeRequest struct {
	RequesterCoinID  string `json:"requester_coin_id"`
	AccepterCoinID   string `json:"accepter_coin_id"`
	RequesterAddress string `json:"requester_address"`
	AccepterAddress  string `json:"accepter_address"`
	StakeAmount      int64  `json:"stake_amount"`
}

// @Summary Stake in a game
// @Description Creates a stake on the Sui blockchain
// @Accept json
// @Produce json
// @Param stake body request.StakeRequest true "Stake request"
// @Success 200 {object} response.StakeResponse
// @Failure 400 {object} response.StakeResponse
// @Router /games/stake [post]
func handleStakeGame(gameService service.GameServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tempReq tempStakeRequest
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json;") {
			log.Printf("Invalid Content-Type for %s %s: %s", c.Request.Method, c.Request.URL.Path, contentType)
			c.JSON(http.StatusBadRequest, response.StakeResponse{
				Success: false,
				Error:   "Invalid Content-Type: expected application/json",
			})
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.StakeResponse{
				Success: false,
				Error:   "Invalid request format: failed to read request body",
			})
			return
		}
		if err := json.Unmarshal(body, &tempReq); err != nil {
			c.JSON(http.StatusBadRequest, response.StakeResponse{
				Success: false,
				Error:   "Invalid request format: " + err.Error(),
			})
			return
		}
		if tempReq.StakeAmount < 0 {
			c.JSON(http.StatusBadRequest, response.StakeResponse{
				Success: false,
				Error:   "stake_amount must be greater than 0",
			})
			return
		}
		req := request.StakeRequest{
			RequesterCoinID:  tempReq.RequesterCoinID,
			AccepterCoinID:   tempReq.AccepterCoinID,
			RequesterAddress: tempReq.RequesterAddress,
			AccepterAddress:  tempReq.AccepterAddress,
			StakeAmount:      uint64(tempReq.StakeAmount),
		}
		if err := validateStakeRequest(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.StakeResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		resp, err := gameService.StakeGame(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Pay winner of a game
// @Description Processes winner payment on the Sui blockchain
// @Accept json
// @Produce json
// @Param pay_winner body request.PayWinnerRequest true "Pay winner request"
// @Success 200 {object} response.PayWinnerResponse
// @Failure 400 {object} response.PayWinnerResponse
// @Router /games/pay_winner [post]
func handlePayWinner(gameService service.GameServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.PayWinnerRequest
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json;") {
			log.Printf("Invalid Content-Type for %s %s: %s", c.Request.Method, c.Request.URL.Path, contentType)
			c.JSON(http.StatusBadRequest, response.PayWinnerResponse{
				Success: false,
				Error:   "Invalid Content-Type: expected application/json",
			})
			return
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.PayWinnerResponse{
				Success: false,
				Error:   "Invalid request format: " + err.Error(),
			})
			return
		}
		if err := validatePayWinnerRequest(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.PayWinnerResponse{
				Success: false,
				Error:   "Validation failed: " + err.Error(),
			})
			return
		}
		resp, err := gameService.PayWinner(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Get stake history
// @Description Retrieves stake history for a given address
// @Produce json
// @Param address path string true "Sui address"
// @Success 200 {object} response.StakeHistoryResponse
// @Failure 400 {object} response.StakeHistoryResponse
// @Router /games/stakes/{address} [get]
func handleGetStakeHistory(gameService service.GameServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, response.StakeHistoryResponse{
				Success: false,
				Error:   "Address parameter is required",
			})
			return
		}
		if err := validateSuiAddress(address); err != nil {
			c.JSON(http.StatusBadRequest, response.StakeHistoryResponse{
				Success: false,
				Error:   "Invalid address format: " + err.Error(),
			})
			return
		}
		resp, err := gameService.GetStakeHistory(address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Get game history
// @Description Retrieves game history for a given address
// @Produce json
// @Param address path string true "Sui address"
// @Success 200 {object} response.GameHistoryResponse
// @Failure 400 {object} response.GameHistoryResponse
// @Router /games/history/{address} [get]
func handleGetGameHistory(gameService service.GameServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, response.GameHistoryResponse{
				Success: false,
				Error:   "Address parameter is required",
			})
			return
		}
		if err := validateSuiAddress(address); err != nil {
			c.JSON(http.StatusBadRequest, response.GameHistoryResponse{
				Success: false,
				Error:   "Invalid address format: " + err.Error(),
			})
			return
		}
		resp, err := gameService.GetGameHistory(address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Get game stats
// @Description Retrieves game statistics (placeholder)
// @Produce json
// @Success 200 {object} gin.H
// @Router /games/stats [get]
func handleGetGameStats(gameService service.GameServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Game stats endpoint - to be implemented",
			"data": gin.H{
				"total_games":    0,
				"total_stakes":   0,
				"active_players": 0,
				"last_updated":   time.Now().Unix(),
			},
		})
	}
}

func validateStakeRequest(req *request.StakeRequest) error {
	if req.RequesterCoinID == "" {
		return fmt.Errorf("requester_coin_id is required")
	}
	if req.AccepterCoinID == "" {
		return fmt.Errorf("accepter_coin_id is required")
	}
	if req.RequesterAddress == "" {
		return fmt.Errorf("requester_address is required")
	}
	if req.AccepterAddress == "" {
		return fmt.Errorf("accepter_address is required")
	}
	if req.StakeAmount == 0 {
		return fmt.Errorf("stake_amount must be greater than 0")
	}
	if req.RequesterAddress == req.AccepterAddress {
		return fmt.Errorf("requester and accepter addresses cannot be the same")
	}
	if err := validateSuiAddress(req.RequesterAddress); err != nil {
		return fmt.Errorf("invalid requester_address: %v", err)
	}
	if err := validateSuiAddress(req.AccepterAddress); err != nil {
		return fmt.Errorf("invalid accepter_address: %v", err)
	}
	return nil
}

func validatePayWinnerRequest(req *request.PayWinnerRequest) error {
	if req.RequesterAddress == "" {
		return fmt.Errorf("requester_address is required")
	}
	if req.AccepterAddress == "" {
		return fmt.Errorf("accepter_address is required")
	}
	if req.StakeAmount == 0 {
		return fmt.Errorf("stake_amount must be greater than 0")
	}
	if req.RequesterAddress == req.AccepterAddress {
		return fmt.Errorf("requester and accepter addresses cannot be the same")
	}
	if err := validateSuiAddress(req.RequesterAddress); err != nil {
		return fmt.Errorf("invalid requester_address: %v", err)
	}
	if err := validateSuiAddress(req.AccepterAddress); err != nil {
		return fmt.Errorf("invalid accepter_address: %v", err)
	}
	return nil
}

func validateSuiAddress(address string) error {
	if len(address) < 40 || len(address) > 66 {
		return fmt.Errorf("invalid address length")
	}
	if !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("address must start with 0x")
	}
	hexPart := address[2:]
	for _, char := range hexPart {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return fmt.Errorf("address contains invalid characters")
		}
	}
	return nil
}
