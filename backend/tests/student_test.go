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
	"strings"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestStudent(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		user := model.User{
			Username: fmt.Sprintf("registerstudenttester-%d", time.Now().UnixNano()),
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
		if fiw, err = fw.CreateFormFile("photo", "photo.png"); err != nil {
			t.Error(err)
			return
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Error(err)
			return
		}
		if fiw, err = fw.CreateFormFile("statusPhoto", "photo.png"); err != nil {
			t.Error(err)
			return
		}
		if _, err := fiw.Write(pixel); err != nil {
			t.Error(err)
			return
		}
		if err := fw.Close(); err != nil {
			t.Error(err)
			return
		}
		req, _ := http.NewRequest("POST", "/auth/student/register", &b)
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
		assert.Equal(t, student.ApprovalStatus, model.StudentApprovalPending)
		assert.Equal(t, student.Phone, "0123456789")
		assert.Equal(t, student.AboutMe, "I am a software tester")
		assert.Equal(t, student.GitHub, "localhost")
		assert.Equal(t, student.LinkedIn, "localhost")
		assert.Equal(t, student.StudentID, "6612345678")
		assert.Equal(t, student.Major, "Software and Knowledge Engineering")
		assert.Equal(t, student.StudentStatus, "Graduated")
		_ = db.Delete(&user)
	})
	t.Run("GetCurrentProfile", func(t *testing.T) {
		student, err := CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("getcurrentprofilestudenttester-%d", time.Now().UnixNano()),
			IsStudent: true,
			IsOAuth:   true,
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&student.User)
		})()
		student.OAuth.FirstName = "MyFirstName"
		student.OAuth.LastName = "MyLastName"
		student.OAuth.Email = "testemail@email.test"
		if err := db.Save(&student.OAuth).Error; err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/students", strings.NewReader(""))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(student.User.ID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		type StudentInfo struct {
			model.Student
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			Email     string `json:"email"`
		}
		type Result struct {
			Profile StudentInfo `json:"profile"`
			Error   string      `json:"error"`
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
		assert.Equal(t, result.Profile.UserID, student.Student.UserID)
		assert.Equal(t, result.Profile.FirstName, student.OAuth.FirstName)
		assert.Equal(t, result.Profile.LastName, student.OAuth.LastName)
		assert.Equal(t, result.Profile.Email, student.OAuth.Email)
	})
	t.Run("AdminGetProfiles", func(t *testing.T) {
		var students []model.Student
		for i := 0; i < 5; i += 1 {
			student, err := CreateUser(UserCreationInfo{
				Username:  fmt.Sprintf("admingetprofilesstudenttester-%d-%d", i, time.Now().UnixNano()),
				IsStudent: true,
				IsOAuth: true,
			})
			if err != nil {
				t.Error(err)
				return
			}
			defer (func() {
				_ = db.Delete(&student.User)
			})()
			students = append(students, *student.Student)
		}
		admin, err := CreateUser(UserCreationInfo{
			Username: fmt.Sprintf("admingetprofilesstudenttester-%d", time.Now().UnixNano()),
			IsAdmin:  true,
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&admin.User)
		})()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/students", strings.NewReader(""))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(admin.User.ID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		var result []model.Student
		err = json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Error(err)
			return
		}
		for i, createdStudent := range students {
			found := false
			for _, student := range result {
				if student.UserID == createdStudent.UserID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Student %d not found", i)
			}
		}
	})
	t.Run("Approve", func(t *testing.T) {
		student, err := CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("approvestudenttester-%d", time.Now().UnixNano()),
			IsAdmin:   true,
			IsStudent: true,
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&student.User)
		})()
		student.Student.ApprovalStatus = model.StudentApprovalPending
		if err = db.Save(&student.Student).Error; err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		payload := `{"approve": true}`
		req, _ := http.NewRequest("POST", fmt.Sprintf("/students/%s/approval", student.Student.UserID), strings.NewReader(payload))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(student.User.ID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/json")
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
		approvedStudent := &model.Student{
			UserID: student.User.ID,
		}
		if err = db.Find(&approvedStudent).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, approvedStudent.ApprovalStatus, model.StudentApprovalAccepted)
	})
}
