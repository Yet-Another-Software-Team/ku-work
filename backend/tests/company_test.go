package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

		req, _ := http.NewRequest("POST", "/company/register", &b)
		req.Header.Set("Content-Type", fw.FormDataContentType())

		router.ServeHTTP(w, req)

		assert.Equal(t, w.Code, http.StatusOK)

		type Result struct {
			Token     string `json:"token"`
			Username  string `json:"username"`
			IsStudent bool   `json:"isStudent"`
			IsCompany bool   `json:"isCompany"`
			Error     string `json:"error"`
		}
		var result Result
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}

		if result.Error != "" {
			t.Fatal(result.Error)
		}

		assert.Equal(t, result.Username, username)
		assert.Equal(t, result.IsStudent, false)
		assert.Equal(t, result.IsCompany, true)
		if result.Token == "" {
			t.Fatal("token is empty")
		}

		// Verify user and company in DB
		var user model.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
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
	})

	t.Run("Duplicate Company Creation", func(t *testing.T) {
		username := fmt.Sprintf("companytester-%d", time.Now().UnixNano())
		// Create a user first
		user := model.User{Username: username}
		if err := db.Create(&user).Error; err != nil {
			t.Fatal(err)
		}

		values := map[string]string{
			"username": username, // Same username
			"password": "password123",
			"email":    "company@test.com",
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

		req, _ := http.NewRequest("POST", "/company/register", &b)
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
	})
}
