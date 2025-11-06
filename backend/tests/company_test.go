package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ku-work/backend/handlers"
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestCompany(t *testing.T) {
	t.Run("Valid Company Creation", func(t *testing.T) {
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		values := map[string]string{
			"username": username,
			"password": "password123",
			"email":    "company@test.com",
			"website":  "https://www.company.com",
			"phone":    "0987654321",
			"address":  "123 Test St",
			"city":     "Testville",
			"country":  "Testland",
		}

		w := httptest.NewRecorder()
		var b bytes.Buffer
		fw := multipart.NewWriter(&b)

		for key, val := range values {
			fiw, err := fw.CreateFormField(key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(fiw, val); err != nil {
				t.Fatal(err)
			}
		}

		// Add photo file
		fiw, err := fw.CreateFormFile("photo", "photo.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Fatal(err)
		}

		// Add banner file
		fiw, err = fw.CreateFormFile("banner", "banner.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Fatal(err)
		}

		if err := fw.Close(); err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("POST", "/auth/company/register", &b)
		req.Header.Set("Content-Type", fw.FormDataContentType())

		router.ServeHTTP(w, req)

		assert.Equal(t, w.Code, http.StatusOK)

		type Result struct {
			Token    string `json:"token"`
			Username string `json:"username"`
			Role     string `json:"role"`
			UserId   string `json:"userId"`
			Error    string `json:"error"`
		}
		var result Result
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}

		if result.Error != "" {
			t.Fatal(result.Error)
		}

		assert.Equal(t, result.Username, username)
		assert.Equal(t, result.Role, "company")
		if result.Token == "" {
			t.Fatal("token is empty")
		}

		// Verify user and company in DB
		var user model.User
		if err := db.Where("username = ? AND user_type = ?", username, "company").First(&user).Error; err != nil {
			t.Fatal(err)
		}

		var company model.Company
		if err := db.Where("user_id = ?", user.ID).First(&company).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, company.Email, "company@test.com")
		assert.Equal(t, company.Phone, "0987654321")
		assert.Equal(t, company.Address, "123 Test St")
		assert.Equal(t, company.City, "Testville")
		assert.Equal(t, company.Country, "Testland")
		_ = db.Delete(&user)
	})
	t.Run("Duplicate Company Creation", func(t *testing.T) {
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		// Create a user first
		user := model.User{Username: username, UserType: "company"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		values := map[string]string{
			"username": username, // Same username
			"password": "password123",
			"email":    "company@test.com",
			"website":  "https://www.company.com",
			"phone":    "0987654321",
			"address":  "123 Test St",
			"city":     "Testville",
			"country":  "Testland",
		}

		w := httptest.NewRecorder()
		var b bytes.Buffer
		fw := multipart.NewWriter(&b)

		for key, val := range values {
			fiw, err := fw.CreateFormField(key)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(fiw, val); err != nil {
				t.Fatal(err)
			}
		}

		// Add photo file
		fiw, err := fw.CreateFormFile("photo", "photo.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Fatal(err)
		}

		// Add banner file
		fiw, err = fw.CreateFormFile("banner", "banner.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Fatal(err)
		}

		if err := fw.Close(); err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest("POST", "/auth/company/register", &b)
		req.Header.Set("Content-Type", fw.FormDataContentType())

		router.ServeHTTP(w, req)

		assert.Equal(t, w.Code, http.StatusConflict)

		type Result struct {
			Error string `json:"error"`
		}
		var result Result
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, result.Error, "Username already exists")
		_ = db.Delete(&user)
	})
	t.Run("Edit Profile", func(t *testing.T) {
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  username,
			IsCompany: true,
		}); err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&userCreationResult.User)
		})()
		company := userCreationResult.Company
		values := map[string]string{
			"phone":   "0123456789",
			"address": "1234 gay street bangcock thailand",
		}
		w := httptest.NewRecorder()
		var b bytes.Buffer
		fw := multipart.NewWriter(&b)
		for key, val := range values {
			fiw, err := fw.CreateFormField(key)
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := io.WriteString(fiw, val); err != nil {
				t.Error(err)
				return
			}
		}
		if err := fw.Close(); err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("PATCH", "/me", &b)
		if err != nil {
			t.Error(err)
			return
		}
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Content-Type", fw.FormDataContentType())
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		resCompany := model.Company{
			UserID: company.UserID,
		}
		if err := db.First(&resCompany).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, resCompany.Phone, "0123456789")
		assert.Equal(t, resCompany.Address, "1234 gay street bangcock thailand")
	})

	t.Run("Get Profile", func(t *testing.T) {
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  username,
			IsAdmin:   true,
			IsCompany: true,
		}); err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&userCreationResult.User)
		})()
		company := userCreationResult.Company
		company.Phone = "0123456789"
		company.Address = "1234 gay street bangcock thailand"
		if err := db.Save(&company).Error; err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", fmt.Sprintf("/company/%s", company.UserID), strings.NewReader(""))
		if err != nil {
			t.Error(err)
			return
		}
		identityRepo := gormrepo.NewGormIdentityRepository(db)
		refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
		revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
		jwtService := services.NewJWTService(refreshRepo, revocationRepo, identityRepo)
		jwtHandler := handlers.NewJWTHandlers(jwtService)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		type Result struct {
			model.Company
			Name string `json:"name"`
		}
		result := Result{}
		err = json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, result.Address, "1234 gay street bangcock thailand")
		assert.Equal(t, result.Phone, "0123456789")
	})
}
