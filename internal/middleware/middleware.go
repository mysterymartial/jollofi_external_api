package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"

	"jollfi-gaming-api/internal/config"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		cfg, exists := c.Get("config")
		if !exists {
			fmt.Printf("Warning: Config not found in context for logging\n")
			return
		}
		config := cfg.(*config.Config)
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoURI))
		if err == nil {
			defer client.Disconnect(context.Background())
			collection := client.Database(config.MongoDatabase).Collection("logs")
			_, err = collection.InsertOne(context.Background(), bson.M{
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"status":    c.Writer.Status(),
				"time":      time.Since(start),
				"timestamp": time.Now(),
			})
			if err != nil {
				fmt.Printf("Failed to log request: %v\n", err)
			}
		}
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error: " + err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
			})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"status":    "healthy",
			"service":   "jollfi-gaming-api",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}
		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			providedKey = c.Query("api_key")
		}
		if providedKey == "" || providedKey == apiKey {
			c.Next()
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid API key",
		})
		c.Abort()
	}
}

func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute * 3)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > time.Minute*3 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(c *gin.Context) {
		if requestsPerMinute <= 0 {
			c.Next()
			return
		}
		ip := c.ClientIP()
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(requestsPerMinute)/60, requestsPerMinute),
			}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			c.Next()
		}()
		select {
		case <-finished:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"error":   "Request timeout",
			})
			c.Abort()
		}
	}
}

func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1048576) // 1MB limit
		c.Next()
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	var (
		requestCount = make(map[string]int)
		mu           sync.RWMutex
	)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		mu.Lock()
		requestCount[c.Request.Method+" "+c.FullPath()]++
		mu.Unlock()
		duration := time.Since(start)
		if duration > 5*time.Second {
			fmt.Printf("SLOW REQUEST: %s %s took %v\n",
				c.Request.Method, c.FullPath(), duration)
		}
	}
}
