package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"jollfi-gaming-api/internal/config"
	"jollfi-gaming-api/internal/data"
	"jollfi-gaming-api/internal/routes"
	"jollfi-gaming-api/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Validate required configuration
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize dependencies
	suiClient, mongoClient := initializeDependencies(cfg)
	defer mongoClient.Close()

	// Initialize services
	gameService := service.NewGameService(suiClient, mongoClient)

	// Setup routes
	router := routes.SetupRoutes(gameService, cfg)

	// Start server
	startServer(router, cfg)
}

// initializeDependencies initializes Sui and MongoDB clients
func initializeDependencies(cfg *config.Config) (*data.SuiClient, *data.MongoClient) {
	// Initialize Sui client
	suiConfig := &data.Config{
		PackageID:  cfg.PackageID,
		ModuleName: cfg.ModuleName,
		PoolID:     cfg.PoolID,
	}

	suiClient, err := data.NewSuiClient(cfg.SuiNetworkURL, cfg.SuiPrivateKey, suiConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Sui client: %v", err)
	}

	// Test Sui connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := suiClient.HealthCheck(ctx); err != nil {
		log.Fatalf("Sui client health check failed: %v", err)
	}
	log.Println("✅ Sui client connected successfully")

	// Initialize MongoDB client
	mongoClient := data.NewMongoClient(cfg.MongoURI, cfg.GetDatabaseName())

	// Test MongoDB connection - FIX: Add context parameter
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	if err := mongoClient.Ping(ctx2); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("✅ MongoDB connected successfully")

	return suiClient, mongoClient
}

// startServer starts the HTTP server with graceful shutdown
func startServer(router http.Handler, cfg *config.Config) {
	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logServerInfo(cfg)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server shutdown complete")
	}
}

// logServerInfo logs server startup information
func logServerInfo(cfg *config.Config) {
	log.Println("🚀 ================================")
	log.Println("🚀 Jollfi Gaming API Starting...")
	log.Println("🚀 ================================")
	log.Printf("📍 Environment: %s", cfg.Environment)
	log.Printf("🌐 Server Port: %s", cfg.Port)
	log.Printf("🔗 Sui Network: %s", cfg.SuiNetworkURL)
	log.Printf("🗄️  MongoDB Database: %s", cfg.GetDatabaseName())
	log.Printf("📦 Package ID: %s", cfg.PackageID)
	log.Printf("🏊 Pool ID: %s", cfg.PoolID)
	log.Printf("🔧 Module Name: %s", cfg.ModuleName)
	log.Printf("📊 Rate Limit: %d req/min", cfg.RateLimit)
	log.Printf("🔒 CORS Enabled: %t", cfg.EnableCORS)
	log.Printf("📝 Logging Enabled: %t", cfg.EnableLogging)

	if cfg.APIKey != "" {
		log.Println("🔑 API Key Authentication: Enabled")
	} else {
		log.Println("🔑 API Key Authentication: Disabled")
	}

	log.Println("🚀 ================================")
	log.Printf("🚀 Server running on http://localhost:%s", cfg.Port)
	log.Println("🚀 ================================")
	log.Println("📋 Available Endpoints:")
	log.Println("   GET  /health")
	log.Println("   GET  /api/v1/info")
	log.Println("   GET  /api/v1/status")
	log.Println("   POST /api/v1/games/stake")
	log.Println("   POST /api/v1/games/pay_winner")
	log.Println("   GET  /api/v1/games/stakes/:address")
	log.Println("   GET  /api/v1/games/history/:address")
	log.Println("🚀 ================================")
}
