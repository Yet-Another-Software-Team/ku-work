package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ku-work/backend/handlers"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestServeFileHandler exercises the ServeFileHandler via the router configured
// in TestMain. It covers serving an existing file, missing file (404) and path
// traversal / invalid identifier (400).
func TestServeFileHandler(t *testing.T) {
	// Ensure files directory exists
	if err := os.MkdirAll("./files", 0755); err != nil {
		t.Fatalf("failed to ensure files directory: %v", err)
	}

	t.Run("ServeExistingFile", func(t *testing.T) {
		// Create a unique file id and write test PNG bytes
		fileID := uuid.New().String()
		filePath := filepath.Join("./files", fileID)

		if err := os.WriteFile(filePath, pixel, 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		// Clean up after test
		defer func() { _ = os.Remove(filePath) }()

		// Create a user and get JWT so the protected route can be accessed
		user := model.User{
			Username: fmt.Sprintf("fileuser-%d", time.Now().UnixNano()),
			UserType: "company",
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() { _ = db.Unscoped().Delete(&user) }()

		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(user.ID)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req, _ := http.NewRequest("GET", "/files/"+fileID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "should return 200 for existing file")
		// Content type for our test PNG should contain "png"
		ct := w.Header().Get("Content-Type")
		assert.Contains(t, ct, "png", "expected png content type")

		// Body should match the file bytes
		assert.Equal(t, pixel, w.Body.Bytes())
	})

	t.Run("NotFoundReturns404", func(t *testing.T) {
		// Use a random id that doesn't exist on disk
		nonExistentID := "nonexistent-" + uuid.New().String()
		// Create a user and token
		user := model.User{
			Username: fmt.Sprintf("nfuser-%d", time.Now().UnixNano()),
			UserType: "company",
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() { _ = db.Unscoped().Delete(&user) }()

		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(user.ID)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req, _ := http.NewRequest("GET", "/files/"+nonExistentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "should return 404 for missing file")
	})

	t.Run("PathTraversalAndInvalidIdentifier", func(t *testing.T) {
		// Create a user and token
		user := model.User{
			Username: fmt.Sprintf("ptuser-%d", time.Now().UnixNano()),
			UserType: "company",
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() { _ = db.Unscoped().Delete(&user) }()

		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(user.ID)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		// Try a fileID containing '..' which should be rejected
		req, _ := http.NewRequest("GET", "/files/..", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "should reject '..' identifier with 400")

		// Try a fileID containing a slash (encoded form is not necessary here because route param will capture a single segment).
		// We request something that includes backslash; server checks for backslash and should return 400.
		req2, _ := http.NewRequest("GET", "/files/%5Cbackslash", nil) // %5C is encoded '\'
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		// The handler should detect invalid identifier (backslash) and return 400
		assert.Equal(t, http.StatusBadRequest, w2.Code, "should reject backslash containing identifier with 400")
	})
}
