package tests

import (
	"context"
	"fmt"
	"ku-work/backend/database"
	"ku-work/backend/middlewares"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestRedis creates a Redis container for testing
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Name:         "kuwork-test-redis",
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(redisContainer)
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	// Set environment variables for Redis connection
	_ = os.Setenv("REDIS_HOST", "127.0.0.1")
	_ = os.Setenv("REDIS_PORT", port.Port())
	_ = os.Setenv("REDIS_PASSWORD", "")
	_ = os.Setenv("REDIS_DB", "0")

	// Small delay to ensure Redis is ready
	time.Sleep(100 * time.Millisecond)

	// Load Redis client
	redisClient, err := database.LoadRedis()
	if err != nil {
		_ = testcontainers.TerminateContainer(redisContainer)
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		if redisClient != nil {
			_ = redisClient.FlushDB(context.Background()).Err()
			_ = redisClient.Close()
		}
		_ = testcontainers.TerminateContainer(redisContainer)
	}

	return redisClient, cleanup
}

// TestRateLimiterWithNilRedis tests that rate limiter fails open when Redis is nil
func TestRateLimiterWithNilRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create middleware with nil Redis client
	middleware := middlewares.RateLimiterWithLimits(nil, 5, 20)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 10 requests - all should succeed since Redis is nil (fail open)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d, expected 200 (fail open)", i+1, w.Code)
		}
	}
}

// TestRateLimiterMinuteLimit tests the per-minute rate limit
func TestRateLimiterMinuteLimit(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	// Create middleware with 5 requests per minute limit
	middleware := middlewares.RateLimiterWithLimits(redisClient, 5, 100)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.100:12345"

	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed but got status %d", i+1, w.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = testIP
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6 should be rate limited (429) but got status %d", w.Code)
	}
}

// TestRateLimiterHourLimit tests the per-hour rate limit
func TestRateLimiterHourLimit(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	// Create middleware with high minute limit but low hour limit
	middleware := middlewares.RateLimiterWithLimits(redisClient, 100, 10)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.101:12345"

	// First 10 requests should succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed but got status %d", i+1, w.Code)
		}
	}

	// 11th request should be rate limited by hour limit
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = testIP
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 11 should be rate limited (429) but got status %d", w.Code)
	}
}

// TestRateLimiterDifferentIPs tests that different IPs are tracked separately
func TestRateLimiterDifferentIPs(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 5, 20)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// IP 1: Make 5 requests (should all succeed)
	ip1 := "192.168.1.100:12345"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip1
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP1 Request %d should succeed but got status %d", i+1, w.Code)
		}
	}

	// IP 1: 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip1
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 should be rate limited but got status %d", w.Code)
	}

	// IP 2: Should be able to make requests (different IP)
	ip2 := "192.168.1.200:12345"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip2
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP2 Request %d should succeed but got status %d", i+1, w.Code)
		}
	}
}

// TestRateLimiterRedisKeys tests that Redis keys are created correctly
func TestRateLimiterRedisKeys(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 5, 20)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.150:12345"

	// Make a request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = testIP
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that Redis keys were created
	ctx := context.Background()
	keys, err := redisClient.Keys(ctx, "ratelimit:192.168.1.150:*").Result()
	if err != nil {
		t.Fatalf("Failed to get Redis keys: %v", err)
	}

	if len(keys) == 0 {
		t.Error("Expected Redis keys to be created but none found")
	}

	// Check that we have both minute and hour keys
	hasMinuteKey := false
	hasHourKey := false

	for _, key := range keys {
		if contains(key, ":minute:") {
			hasMinuteKey = true

			// Check value
			val, err := redisClient.Get(ctx, key).Int()
			if err != nil {
				t.Errorf("Failed to get minute key value: %v", err)
			}
			if val != 1 {
				t.Errorf("Expected minute key value to be 1, got %d", val)
			}

			// Check TTL
			ttl, err := redisClient.TTL(ctx, key).Result()
			if err != nil {
				t.Errorf("Failed to get TTL: %v", err)
			}
			if ttl <= 0 || ttl > 2*time.Minute {
				t.Errorf("Expected TTL to be between 0 and 2 minutes, got %v", ttl)
			}
		}

		if contains(key, ":hour:") {
			hasHourKey = true

			// Check TTL
			ttl, err := redisClient.TTL(ctx, key).Result()
			if err != nil {
				t.Errorf("Failed to get TTL: %v", err)
			}
			if ttl <= 0 || ttl > 2*time.Hour {
				t.Errorf("Expected TTL to be between 0 and 2 hours, got %v", ttl)
			}
		}
	}

	if !hasMinuteKey {
		t.Error("Expected minute key to be created")
	}

	if !hasHourKey {
		t.Error("Expected hour key to be created")
	}
}

// TestRateLimiterCounterIncrement tests that counters increment correctly
func TestRateLimiterCounterIncrement(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 10, 50)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.180:12345"
	numRequests := 3

	// Make multiple requests
	for i := 0; i < numRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Check that counters are correct
	ctx := context.Background()
	keys, err := redisClient.Keys(ctx, "ratelimit:192.168.1.180:minute:*").Result()
	if err != nil {
		t.Fatalf("Failed to get Redis keys: %v", err)
	}

	if len(keys) == 0 {
		t.Fatal("Expected minute key to exist")
	}

	val, err := redisClient.Get(ctx, keys[0]).Int()
	if err != nil {
		t.Fatalf("Failed to get counter value: %v", err)
	}

	if val != numRequests {
		t.Errorf("Expected counter to be %d, got %d", numRequests, val)
	}
}

// TestRateLimiterConcurrency tests that rate limiter works correctly with concurrent requests
func TestRateLimiterConcurrency(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 10, 50)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.190:12345"
	numConcurrent := 20

	// Channel to collect results
	results := make(chan int, numConcurrent)

	// Make concurrent requests
	for i := 0; i < numConcurrent; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = testIP
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < numConcurrent; i++ {
		code := <-results
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitedCount++
		}
	}

	// Should have 10 successful and 10 rate limited
	if successCount != 10 {
		t.Errorf("Expected 10 successful requests, got %d", successCount)
	}

	if rateLimitedCount != 10 {
		t.Errorf("Expected 10 rate limited requests, got %d", rateLimitedCount)
	}
}

// TestRateLimiterMultipleEndpoints tests rate limiting across different endpoints
func TestRateLimiterMultipleEndpoints(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Different endpoints with different limits
	router.GET("/strict", middlewares.RateLimiterWithLimits(redisClient, 2, 10), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "strict"})
	})

	router.GET("/lenient", middlewares.RateLimiterWithLimits(redisClient, 10, 50), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "lenient"})
	})

	// Use different IPs for independent testing
	strictIP := "192.168.1.200:12345"
	lenientIP := "192.168.1.201:12345"

	// Test strict endpoint (2 req/min limit)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/strict", nil)
		req.RemoteAddr = strictIP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Strict endpoint request %d should succeed", i+1)
		}
	}

	// 3rd request to strict endpoint should fail
	req := httptest.NewRequest(http.MethodGet, "/strict", nil)
	req.RemoteAddr = strictIP
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Strict endpoint should be rate limited")
	}

	// Lenient endpoint with different IP - should allow 10 requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/lenient", nil)
		req.RemoteAddr = lenientIP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Lenient endpoint request %d should succeed", i+1)
		}
	}

	// 11th request should fail on lenient endpoint
	req = httptest.NewRequest(http.MethodGet, "/lenient", nil)
	req.RemoteAddr = lenientIP
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Lenient endpoint should be rate limited after 10 requests")
	}
}

// TestRateLimiterErrorMessage tests that appropriate error messages are returned
func TestRateLimiterErrorMessage(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 2, 5)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testIP := "192.168.1.210:12345"

	// Exhaust minute limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Next request should return minute limit error
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = testIP
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 status code, got %d", w.Code)
	}

	// Check response body contains error message
	body := w.Body.String()
	if !contains(body, "error") {
		t.Errorf("Expected error message in response body, got: %s", body)
	}
}

// TestRateLimiterRedisConnection tests behavior when Redis connection is available
func TestRateLimiterRedisConnection(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	// Verify Redis is connected
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis should be connected: %v", err)
	}

	if pong != "PONG" {
		t.Errorf("Expected PONG from Redis, got %s", pong)
	}
}

// Benchmark for rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	redisClient, cleanup := setupTestRedis(&testing.T{})
	defer cleanup()

	gin.SetMode(gin.TestMode)

	middleware := middlewares.RateLimiterWithLimits(redisClient, 1000000, 1000000)

	router := gin.New()
	router.GET("/test", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", i%255)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
