package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ku-work/backend/handlers"
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestStudent(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		user := model.User{
			Username: "registerstudenttester",
		}
		if result := db.Create(&user); result.Error != nil {
			t.Error(result.Error)
			return
		}
		values := map[string]string{
			"phone":         "0123456789",
			"birthDate":     "2025-09-01T07:21:14.806Z",
			"aboutMe":       "I am a software tester",
			"github":        "localhost",
			"linkedIn":      "localhost",
			"studentId":     "6612345678",
			"major":         "Software and Knowledge Engineering",
			"studentStatus": "Graduated",
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
		var fiw io.Writer
		var err error
		if fiw, err = fw.CreateFormFile("photo", "photo.jpg"); err != nil {
			t.Error(err)
			return
		}
		if _, err := io.WriteString(fiw, "[photo content]"); err != nil {
			t.Error(err)
			return
		}
		if fiw, err = fw.CreateFormFile("statusPhoto", "statusPhoto.jpg"); err != nil {
			t.Error(err)
			return
		}
		if _, err := io.WriteString(fiw, "[statusPhoto content]"); err != nil {
			t.Error(err)
			return
		}
		if err := fw.Close(); err != nil {
			t.Error(err)
			return
		}
		req, _ := http.NewRequest("POST", "/students/register", &b)
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(user.ID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Set("Content-Type", fw.FormDataContentType())
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		type Result struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		result := Result{}
		err = json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Error(err)
			return
		}
		if result.Error != "" {
			t.Error(result.Error)
			return
		}
		if result.Message != "ok" {
			t.Error("Message is not ok")
			return
		}
		student := model.Student{
			UserID: user.ID,
		}
		if err := db.First(&student).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, student.Phone, "0123456789")
		assert.Equal(t, student.AboutMe, "I am a software tester")
		assert.Equal(t, student.GitHub, "localhost")
		assert.Equal(t, student.LinkedIn, "localhost")
		assert.Equal(t, student.StudentID, "6612345678")
		assert.Equal(t, student.Major, "Software and Knowledge Engineering")
		assert.Equal(t, student.StudentStatus, "Graduated")
	})
}
