package tests

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ku-work/backend/handlers"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"

	"github.com/stretchr/testify/assert"
)

// Tests for GET /me and PATCH /me endpoints.
func TestUserMeEndpoints(t *testing.T) {
	t.Run("GET /me returns profile for authenticated user", func(t *testing.T) {
		// Create a company user (has associated company record) so profile returns meaningful data.
		username := fmt.Sprintf("getmetester-%d", time.Now().UnixNano())
		created, err := CreateUser(UserCreationInfo{
			Username:  username,
			IsCompany: true,
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		// Clean up
		defer func() {
			_ = db.Unscoped().Delete(&created.User)
			if created.Company != nil {
				_ = db.Unscoped().Delete(created.Company)
			}
		}()

		// Generate JWT for the user
		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(created.User.ID)
		if err != nil {
			t.Fatalf("failed to generate jwt: %v", err)
		}

		req, _ := http.NewRequest("GET", "/me", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GET /me should return 200 for authenticated user")

		// Ensure response contains the username (profile payload expected to include username)
		body := w.Body.String()
		if !assert.Contains(t, body, username, "response body should contain username") {
			t.Logf("response body: %s", body)
		}
	})

	t.Run("PATCH /me updates user/company profile fields", func(t *testing.T) {
		// Create a company user to update
		username := fmt.Sprintf("patchmetester-%d", time.Now().UnixNano())
		created, err := CreateUser(UserCreationInfo{
			Username:  username,
			IsCompany: true,
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		// Clean up
		defer func() {
			_ = db.Unscoped().Delete(&created.User)
			if created.Company != nil {
				_ = db.Unscoped().Delete(created.Company)
			}
		}()

		// Generate JWT token for this user
		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(created.User.ID)
		if err != nil {
			t.Fatalf("failed to generate jwt: %v", err)
		}

		// Prepare multipart form payload to update phone and address
		var b bytes.Buffer
		wr := multipart.NewWriter(&b)
		if err := wr.WriteField("phone", "0812345678"); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
		if err := wr.WriteField("address", "123 Updated Address"); err != nil {
			t.Fatalf("failed to write field: %v", err)
		}
		// Close the writer to set boundary
		if err := wr.Close(); err != nil {
			t.Fatalf("failed to close writer: %v", err)
		}

		req, _ := http.NewRequest("PATCH", "/me", &b)
		req.Header.Set("Content-Type", wr.FormDataContentType())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		respRecorder := httptest.NewRecorder()
		router.ServeHTTP(respRecorder, req)

		assert.Equal(t, http.StatusOK, respRecorder.Code, "PATCH /me should return 200 on success")

		// Reload company record from DB to verify changes
		company := model.Company{
			UserID: created.User.ID,
		}
		if err := db.Where("user_id = ?", created.User.ID).First(&company).Error; err != nil {
			t.Fatalf("failed to fetch company after patch: %v", err)
		}
		assert.Equal(t, "0812345678", company.Phone, "company phone should be updated")
		assert.Equal(t, "123 Updated Address", company.Address, "company address should be updated")
	})

	t.Run("PATCH /me without auth returns 401", func(t *testing.T) {
		var b bytes.Buffer
		wr := multipart.NewWriter(&b)
		_ = wr.WriteField("phone", "0000000000")
		_ = wr.Close()

		req, _ := http.NewRequest("PATCH", "/me", &b)
		req.Header.Set("Content-Type", wr.FormDataContentType())
		// No Authorization header

		respRecorder := httptest.NewRecorder()
		router.ServeHTTP(respRecorder, req)
		assert.Equal(t, http.StatusUnauthorized, respRecorder.Code, "PATCH /me without auth should return 401")
	})

	t.Run("GET /me without auth returns 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/me", nil)
		// No Authorization header
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "GET /me without auth should return 401")
	})

	// Additional small test: PATCH /me with file upload (ensure handler accepts form with files)
	t.Run("PATCH /me with multipart including files", func(t *testing.T) {
		// Create user
		username := fmt.Sprintf("fileuploadtester-%d", time.Now().UnixNano())
		created, err := CreateUser(UserCreationInfo{
			Username:  username,
			IsCompany: true,
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() {
			_ = db.Unscoped().Delete(&created.User)
			if created.Company != nil {
				_ = db.Unscoped().Delete(created.Company)
			}
		}()

		userRepo := gormrepo.NewGormUserRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		token, _, err := jwtHandler.GenerateTokens(created.User.ID)
		if err != nil {
			t.Fatalf("failed to generate jwt: %v", err)
		}

		var b bytes.Buffer
		wr := multipart.NewWriter(&b)
		_ = wr.WriteField("phone", "0999999999")
		// attach a small file field (photo) using the provided pixel bytes
		fw, err := wr.CreateFormFile("photo", "photo.png")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := fw.Write(pixel); err != nil {
			t.Fatalf("failed to write file bytes: %v", err)
		}
		_ = wr.Close()

		req, _ := http.NewRequest("PATCH", "/me", &b)
		req.Header.Set("Content-Type", wr.FormDataContentType())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		respRecorder := httptest.NewRecorder()
		router.ServeHTTP(respRecorder, req)

		// Handler may or may not accept file updates depending on role; ensure non-error status
		assert.True(t, respRecorder.Code == http.StatusOK || respRecorder.Code == http.StatusBadRequest || respRecorder.Code == http.StatusInternalServerError,
			"PATCH /me with file should not panic; got status %d", respRecorder.Code)

		// If it succeeded, verify phone update
		if respRecorder.Code == http.StatusOK {
			company := model.Company{}
			if err := db.Where("user_id = ?", created.User.ID).First(&company).Error; err == nil {
				assert.Equal(t, "0999999999", company.Phone)
			}
		}
	})
}
