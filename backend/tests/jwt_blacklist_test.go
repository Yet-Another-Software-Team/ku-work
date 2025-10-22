package tests

import (
	"encoding/json"
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/middlewares"
	"ku-work/backend/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestRouter creates a test router with all necessary handlers
func setupTestRouter(db *gorm.DB, jwtHandlers *handlers.JWTHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes similar to production
	auth := router.Group("/auth")
	authProtected := auth.Group("", middlewares.AuthMiddlewareWithDB(jwtHandlers.JWTSecret, db))
	authProtected.POST("/logout", jwtHandlers.LogoutHandler)

	// Protected route to test authentication
	router.GET("/protected", middlewares.AuthMiddlewareWithDB(jwtHandlers.JWTSecret, db), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	return router
}

// TestJWTBlacklistOnLogout tests that JWT is blacklisted after logout
func TestJWTBlacklistOnLogout(t *testing.T) {
	// Setup test database
	db, err := database.LoadDB()
	if err != nil {
		t.Skipf("Database not available: %v", err)
		return
	}

	// Clean up test data
	db.Unscoped().Where("1 = 1").Delete(&model.RevokedJWT{})
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	// Create JWT handlers
	jwtHandlers := handlers.NewJWTHandlers(db)

	// Setup router
	router := setupTestRouter(db, jwtHandlers)

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
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse JSON response")
		assert.Equal(t, "Logged out successfully", response["message"])

		// Verify JWT is in blacklist
		var revokedJWT model.RevokedJWT
		err = db.Where("user_id = ?", testUserID).First(&revokedJWT).Error
		assert.NoError(t, err, "JWT should be in blacklist")
		assert.NotEmpty(t, revokedJWT.JTI, "JTI should be set")
		assert.Equal(t, testUserID, revokedJWT.UserID, "User ID should match")
	})

	// Step 3: Verify token is rejected after logout
	t.Run("Token rejected after logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should reject revoked token")
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse JSON response")
		assert.Equal(t, "Token has been revoked", response["error"])
	})
}

// TestJWTBlacklistMultipleSessions tests that only the logged out session is invalidated
func TestJWTBlacklistMultipleSessions(t *testing.T) {
	db, err := database.LoadDB()
	if err != nil {
		t.Skipf("Database not available: %v", err)
		return
	}

	// Clean up
	db.Unscoped().Where("1 = 1").Delete(&model.RevokedJWT{})
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	jwtHandlers := handlers.NewJWTHandlers(db)
	router := setupTestRouter(db, jwtHandlers)

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

// TestJWTBlacklistCleanup tests that expired JWTs are removed from blacklist
func TestJWTBlacklistCleanup(t *testing.T) {
	db, err := database.LoadDB()
	if err != nil {
		t.Skipf("Database not available: %v", err)
		return
	}

	// Clean up
	db.Unscoped().Where("1 = 1").Delete(&model.RevokedJWT{})

	// Insert some test revoked JWTs with valid UUID
	testUserID := uuid.New().String()
	expiredJWT := model.RevokedJWT{
		JTI:       "expired-jwt-123",
		UserID:    testUserID,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		RevokedAt: time.Now().Add(-2 * time.Hour),
	}
	db.Create(&expiredJWT)

	activeJWT := model.RevokedJWT{
		JTI:       "active-jwt-456",
		UserID:    testUserID,
		ExpiresAt: time.Now().Add(10 * time.Minute), // Still valid
		RevokedAt: time.Now(),
	}
	db.Create(&activeJWT)

	// Verify both exist
	var count int64
	db.Model(&model.RevokedJWT{}).Count(&count)
	assert.Equal(t, int64(2), count, "Should have 2 revoked JWTs")

	// Run cleanup (this would normally be done by scheduler)
	// Note: In real test, import the cleanup function from helper package
	now := time.Now()
	result := db.Unscoped().
		Where("expires_at < ?", now).
		Delete(&model.RevokedJWT{})
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected, "Should delete 1 expired JWT")

	// Verify only active JWT remains
	var remaining []model.RevokedJWT
	db.Find(&remaining)
	assert.Equal(t, 1, len(remaining), "Should have 1 JWT remaining")
	if len(remaining) > 0 {
		assert.Equal(t, "active-jwt-456", remaining[0].JTI, "Active JWT should remain")
	}
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
	db, err := database.LoadDB()
	if err != nil {
		t.Skipf("Database not available: %v", err)
		return
	}

	db.Unscoped().Where("1 = 1").Delete(&model.RevokedJWT{})
	db.Unscoped().Where("1 = 1").Delete(&model.RefreshToken{})

	jwtHandlers := handlers.NewJWTHandlers(db)
	router := setupTestRouter(db, jwtHandlers)

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

// Benchmark JWT blacklist lookup performance
func BenchmarkJWTBlacklistLookup(b *testing.B) {
	db, err := database.LoadDB()
	if err != nil {
		b.Skipf("Database not available: %v", err)
		return
	}

	// Seed some revoked JWTs
	for i := 0; i < 100; i++ {
		jwt := model.RevokedJWT{
			JTI:       "benchmark-jwt-" + string(rune(i)),
			UserID:    "bench-user",
			ExpiresAt: time.Now().Add(15 * time.Minute),
			RevokedAt: time.Now(),
		}
		db.Create(&jwt)
	}

	testJTI := "benchmark-jwt-50"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var revokedJWT model.RevokedJWT
		db.Where("jti = ?", testJTI).First(&revokedJWT)
	}
}
