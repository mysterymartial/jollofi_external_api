package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Environment   string
	MongoURI      string
	MongoDatabase string
	SuiNetworkURL string
	PackageID     string
	PoolID        string
	ModuleName    string
	SuiPrivateKey string
	JWTSecret     string
	APIKey        string
	LogLevel      string
	EnableCORS    bool
	EnableLogging bool
	RateLimit     int
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}
	return &Config{
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "jollfi_games"),
		SuiNetworkURL: getEnv("SUI_NETWORK_URL", "https://fullnode.testnet.sui.io:443"),
		PackageID:     getEnv("SUI_PACKAGE_ID", ""),
		PoolID:        getEnv("SUI_POOL_ID", ""),
		ModuleName:    getEnv("SUI_MODULE_NAME", "jollfi_wallet"),
		SuiPrivateKey: getEnv("SUI_PRIVATE_KEY", ""),
		JWTSecret:     getEnv("JWT_SECRET", generateRandomSecret()),
		APIKey:        getEnv("API_KEY", ""),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		EnableCORS:    getEnvBool("ENABLE_CORS", true),
		EnableLogging: getEnvBool("ENABLE_LOGGING", true),
		RateLimit:     getEnvInt("RATE_LIMIT", 100),
	}
}

func (c *Config) ValidateConfig() error {
	required := map[string]string{
		"SUI_PRIVATE_KEY": c.SuiPrivateKey,
		"SUI_PACKAGE_ID":  c.PackageID,
		"SUI_POOL_ID":     c.PoolID,
		"MONGODB_URI":     c.MongoURI,
	}
	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "default-secret-1234567890abcdef"
	}
	return hex.EncodeToString(b)
}
