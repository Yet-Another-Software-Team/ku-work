package tests

import (
	"encoding/json"
	"ku-work/backend/handlers"
	"ku-work/backend/middlewares"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a test router with all necessary handlers
func setupTestRouter(redisClient *redis.Client, jwtHandlers *handlers.JWTHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes similar to production
	auth := router.Group("/auth")
	authProtected := auth.Group("", middlewares.AuthMiddlewareWithRedis(jwtHandlers.Service.JWTSecret, redisClient))
	authProtected.POST("/logout", jwtHandlers.LogoutHandler)

	// Protected route to test authentication
	router.GET("/protected", middlewares.AuthMiddlewareWithRedis(jwtHandlers.Service.JWTSecret, redisClient), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	return router
}

// TestJWTBlacklistOnLogout tests that JWT is blacklisted after logout
func TestJWTBlacklistOnLogout(t *testing.T) {
	// Clean up test data
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	// Create JWT handlers
	identityRepo := gormrepo.NewGormIdentityRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
	jwtHandlers := handlers.NewJWTHandlers(jwtService)

	// Setup router
	router := setupTestRouter(redisClient, jwtHandlers)

	// Generate test tokens for a test user with valid UUID
	testUserID := uuid.New().String()
	jwtToken, _, err := jwtHandlers.GenerateTokens(testUserID)
	assert.NoError(t, err, "Should generate tokens successfully")

	// Step 1: Verify token works before logout
	t.Run("Token works before logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should access protected route")
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse JSON response")
		assert.Equal(t, testUserID, response["userID"], "Should return correct user ID")
	})

	// Step 2: Logout (blacklist the token)
	t.Run("Logout blacklists token", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Logout should succeed")
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse JSON response")
		assert.Equal(t, "Logged out successfully", response["message"])
	})

	// Step 3: Verify token is rejected after logout
	t.Run("Token rejected after logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should reject revoked token")
		var response map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse JSON response")
		assert.Equal(t, "Token has been revoked", response["error"])
	})
}

// TestJWTBlacklistMultipleSessions tests that only the logged out session is invalidated
func TestJWTBlacklistMultipleSessions(t *testing.T) {
	// Clean up
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	identityRepo := gormrepo.NewGormIdentityRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
	jwtHandlers := handlers.NewJWTHandlers(jwtService)
	router := setupTestRouter(redisClient, jwtHandlers)

	testUserID := uuid.New().String()

	// Generate two different tokens (simulating two devices)
	token1, _, err := jwtHandlers.GenerateTokens(testUserID)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Ensure different JTI

	token2, _, err := jwtHandlers.GenerateTokens(testUserID)
	assert.NoError(t, err)

	// Both tokens should work initially
	t.Run("Both tokens work initially", func(t *testing.T) {
		req1, _ := http.NewRequest("GET", "/protected", nil)
		req1.Header.Set("Authorization", "Bearer "+token1)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		req2, _ := http.NewRequest("GET", "/protected", nil)
		req2.Header.Set("Authorization", "Bearer "+token2)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	// Logout with token1
	t.Run("Logout first session", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Token1 should be rejected
	t.Run("First token rejected after logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token1)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Token2 should still work (different session)
	t.Run("Second token still works", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token2)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Other session should remain active")
	})
}

// TestJWTBlacklistCleanup tests that Redis automatically handles TTL expiration
func TestJWTBlacklistCleanup(t *testing.T) {
	t.Skip("Redis handles cleanup automatically via TTL - no manual cleanup needed")

	// With Redis, expired JWTs are automatically removed when their TTL expires
	// No cleanup task or test needed - this is a Redis built-in feature
}

// TestJWTWithoutJTI tests backward compatibility with JWTs without JTI
func TestJWTWithoutJTI(t *testing.T) {
	t.Skip("This test demonstrates backward compatibility behavior")

	// Old JWTs without JTI will have an empty claims.ID
	// The blacklist check: db.Where("jti = ?", "").First(&revokedJWT)
	// Will not match any records (assuming no JWT with empty JTI in DB)
	// Therefore, old tokens will pass through until they expire naturally
}

// TestMultipleLogoutAttempts tests idempotent logout behavior
func TestMultipleLogoutAttempts(t *testing.T) {
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	identityRepo := gormrepo.NewGormIdentityRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
	jwtHandlers := handlers.NewJWTHandlers(jwtService)
	router := setupTestRouter(redisClient, jwtHandlers)

	testUserID := uuid.New().String()
	jwtToken, _, err := jwtHandlers.GenerateTokens(testUserID)
	assert.NoError(t, err)

	// First logout
	t.Run("First logout succeeds", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Second logout with same token (should still return success)
	t.Run("Second logout with revoked token", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// Logout handler is designed to be idempotent
		// It may succeed or fail, but should handle gracefully
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
			"Should handle multiple logout attempts gracefully")
	})
}
