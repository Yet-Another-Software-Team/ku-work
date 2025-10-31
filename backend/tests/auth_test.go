package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ku-work/backend/handlers"
	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestAuth covers admin login, company login and token refresh endpoints.
func TestAuth(t *testing.T) {
	t.Run("Admin Login success and failure", func(t *testing.T) {
		// Create an admin user
		username := fmt.Sprintf("admintester-%d", time.Now().UnixNano())
		userRes, err := CreateUser(UserCreationInfo{
			Username: username,
			IsAdmin:  true,
		})
		if err != nil {
			t.Fatalf("failed to create admin user: %v", err)
		}
		defer func() {
			_ = db.Unscoped().Delete(&userRes.User)
			if userRes.Admin != nil {
				_ = db.Unscoped().Delete(userRes.Admin)
			}
		}()

		// Set a known password for login
		password := "StrongPassword123!"
		hashed, err := helper.HashPassword(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}
		if err := db.Model(&model.User{}).Where("id = ?", userRes.User.ID).Update("password_hash", hashed).Error; err != nil {
			t.Fatalf("failed to set password hash: %v", err)
		}

		// Prepare login payload
		loginPayload := map[string]string{
			"username": username,
			"password": password,
		}
		payloadBytes, _ := json.Marshal(loginPayload)

		// Perform login request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/admin/login", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "admin login should succeed")

		var resp struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Role     string `json:"role"`
			UserId   string `json:"userId"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		assert.Empty(t, resp.Error)
		assert.Equal(t, username, resp.Username)
		assert.Equal(t, string(helper.Admin), resp.Role)
		assert.NotEmpty(t, resp.Token)

		// Wrong password should fail
		w2 := httptest.NewRecorder()
		wrongPayload, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "incorrect-password",
		})
		req2, _ := http.NewRequest("POST", "/auth/admin/login", bytes.NewReader(wrongPayload))
		req2.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code, "login with wrong password should be unauthorized")
	})

	t.Run("Company Login success and failure", func(t *testing.T) {
		// Create a company user using the CreateUser helper
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		userRes, err := CreateUser(UserCreationInfo{
			Username:  username,
			IsCompany: true,
		})
		if err != nil {
			t.Fatalf("failed to create company user: %v", err)
		}
		defer func() {
			_ = db.Unscoped().Delete(&userRes.User)
			if userRes.Company != nil {
				_ = db.Unscoped().Delete(userRes.Company)
			}
		}()

		password := "CompanyPass!234"
		hashed, err := helper.HashPassword(password)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}
		if err := db.Model(&model.User{}).Where("id = ?", userRes.User.ID).Update("password_hash", hashed).Error; err != nil {
			t.Fatalf("failed to set password hash: %v", err)
		}

		// Successful login
		payloadBytes, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/company/login", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "company login should succeed")
		var resp struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Role     string `json:"role"`
			UserId   string `json:"userId"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		assert.Empty(t, resp.Error)
		assert.Equal(t, username, resp.Username)
		assert.Equal(t, string(helper.Company), resp.Role)
		assert.NotEmpty(t, resp.Token)

		// Invalid credentials
		w2 := httptest.NewRecorder()
		badBytes, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "bad-password",
		})
		req2, _ := http.NewRequest("POST", "/auth/company/login", bytes.NewReader(badBytes))
		req2.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code, "company login with wrong password should be unauthorized")
	})

	t.Run("Refresh Token success and missing cookie", func(t *testing.T) {
		// Create a user (student/company/any) and generate tokens via JWTHandlers
		userRes, err := CreateUser(UserCreationInfo{
			Username: fmt.Sprintf("refreshtester-%d", time.Now().UnixNano()),
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		defer func() {
			_ = db.Unscoped().Delete(&userRes.User)
		}()

		jwtHandler := handlers.NewJWTHandlers(db, redisClient)
		jwtToken, refreshToken, err := jwtHandler.GenerateTokens(userRes.User.ID)
		if err != nil {
			t.Fatalf("failed to generate tokens: %v", err)
		}

		// Build a local test router that registers only the refresh handler (no rate-limiter)
		localRouter := gin.New()
		// Register the refresh endpoint directly without the RateLimiter middleware used in production routes.
		localRouter.POST("/auth/refresh", jwtHandler.RefreshTokenHandler)

		// Successful refresh: send Authorization header and refresh_token cookie
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/refresh", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		// Set cookie header directly
		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: refreshToken,
			Path:  "/",
		})
		localRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "refresh should succeed with valid cookie and jwt")

		var resp struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Role     string `json:"role"`
			UserId   string `json:"userId"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse refresh response: %v", err)
		}
		assert.Empty(t, resp.Error)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, userRes.User.ID, resp.UserId)

		// Missing cookie should fail
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/auth/refresh", nil)
		req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		// no cookie set
		localRouter.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code, "refresh without cookie should be unauthorized")
	})
}
