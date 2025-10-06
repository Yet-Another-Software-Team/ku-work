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

func TestJob(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("createjobtester-%d", time.Now().UnixNano()),
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
		w := httptest.NewRecorder()
		jobName := fmt.Sprintf("testjob1-%d", time.Now().UnixNano())
		payload := fmt.Sprintf(`{
	"name": "%s",
	"position": "testposition",
	"duration": "forever",
	"description": "ass",
	"location": "thailand",
	"jobtype": "casual",
	"experience": "internship",
	"minsalary": 1,
	"maxsalary": 2
}`, jobName)
		req, _ := http.NewRequest("POST", "/job", strings.NewReader(payload))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		job := &model.Job{}
		type Result struct {
			ID *uint `json:"id"`
		}
		result := Result{}
		err = json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Error(err)
			return
		}
		if result.ID == nil {
			t.Error("Got null for id")
			return
		}
		if err := db.First(job, &model.Job{ID: *result.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, job.Name, jobName)
		assert.Equal(t, job.Position, "testposition")
		assert.Equal(t, job.Duration, "forever")
		assert.Equal(t, job.Description, "ass")
		assert.Equal(t, job.Location, "thailand")
		assert.Equal(t, job.JobType, model.JobType("casual"))
		assert.Equal(t, job.Experience, model.ExperienceType("internship"))
		assert.Equal(t, job.MinSalary, uint(1))
		assert.Equal(t, job.MaxSalary, uint(2))
	})
	t.Run("Edit", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("editjobtester-%d", time.Now().UnixNano()),
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
		job := model.Job{
			Name:        fmt.Sprintf("nice-job-%d", time.Now().UnixNano()),
			CompanyID:   company.UserID,
			Position:    "software engineer",
			Duration:    "6 months",
			Description: "make software",
			Location:    "bangkok",
			JobType:     model.JobTypeInternship,
			Experience:  model.ExperienceInternship,
			MinSalary:   10,
			MaxSalary:   100,
		}
		err = db.Create(&job).Error
		if err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		payload := fmt.Sprintf(`{
	"id": %d,
	"name": "good job",
	"position": "software tester",
	"description": "test software"
}`, job.ID)
		req, _ := http.NewRequest("PATCH", "/job", strings.NewReader(payload))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
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
		edited_job := &model.Job{}
		if err := db.First(edited_job, &model.Job{ID: job.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, edited_job.Name, "good job")
		assert.Equal(t, edited_job.Position, "software tester")
		assert.Equal(t, edited_job.Duration, "6 months")
		assert.Equal(t, edited_job.Description, "test software")
		assert.Equal(t, edited_job.Location, "bangkok")
		assert.Equal(t, edited_job.JobType, model.JobType("internship"))
		assert.Equal(t, edited_job.Experience, model.ExperienceType("internship"))
		assert.Equal(t, edited_job.MinSalary, uint(10))
		assert.Equal(t, edited_job.MaxSalary, uint(100))
	})
	t.Run("Fetch", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("fetchjobtester-%d", time.Now().UnixNano()),
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
		job := model.Job{
			Name:           fmt.Sprintf("nice-job-%d", time.Now().UnixNano()),
			CompanyID:      company.UserID,
			Position:       "software engineer",
			Duration:       "6 months",
			Description:    "make software",
			Location:       "bangkok",
			JobType:        model.JobTypeInternship,
			Experience:     model.ExperienceInternship,
			MinSalary:      10,
			MaxSalary:      100,
			IsOpen:         true,
			ApprovalStatus: model.JobApprovalAccepted,
		}
		err = db.Create(&job).Error
		if err != nil {
			t.Error(err)
			return
		}
		photoFile := model.File{UserID: company.UserID, FileType: model.FileTypeJPEG, Category: model.FileCategoryImage}
		if err := db.Create(&photoFile).Error; err != nil {
			t.Error(err)
			return
		}
		statusFile := model.File{UserID: company.UserID, FileType: model.FileTypePDF, Category: model.FileCategoryImage}
		if err := db.Create(&statusFile).Error; err != nil {
			t.Error(err)
			return
		}
		student := model.Student{
			UserID:              company.UserID,
			ApprovalStatus:      model.StudentApprovalAccepted,
			PhotoID:             photoFile.ID,
			StudentStatusFileID: statusFile.ID,
		}
		err = db.Create(&student).Error
		if err != nil {
			t.Error(err)
			return
		}
		jobApp := model.JobApplication{
			JobID:  job.ID,
			UserID: company.UserID,
			Status: model.JobApplicationAccepted,
		}
		err = db.Create(&jobApp).Error
		if err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/job?keyword=nice", strings.NewReader(""))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		type JobWithApplicationStatistics struct {
			model.Job
			Pending  int64 `json:"pending"`
			Accepted int64 `json:"accepted"`
			Rejected int64 `json:"rejected"`
		}
		type Result struct {
			Jobs  []JobWithApplicationStatistics `json:"jobs"`
			Error string                         `json:"error"`
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
		assert.Equal(t, len(result.Jobs), 1)
		assert.Equal(t, result.Jobs[0].Accepted, int64(1))
		assert.Equal(t, result.Jobs[0].Position, "software engineer")
	})
	t.Run("Approve", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("approvejobtester-%d", time.Now().UnixNano()),
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
		job := model.Job{
			Name:        fmt.Sprintf("nice-job-%d", time.Now().UnixNano()),
			CompanyID:   company.UserID,
			Position:    "software engineer",
			Duration:    "6 months",
			Description: "make software",
			Location:    "bangkok",
			JobType:     model.JobTypeInternship,
			Experience:  model.ExperienceInternship,
			MinSalary:   10,
			MaxSalary:   100,
		}
		err = db.Create(&job).Error
		if err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		payload := fmt.Sprintf(`{"id": %d,"approve": true}`, job.ID)
		req, _ := http.NewRequest("POST", "/job/approve", strings.NewReader(payload))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
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
		edited_job := &model.Job{}
		if err := db.First(edited_job, &model.Job{ID: job.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, edited_job.ApprovalStatus, model.JobApprovalAccepted)
		assert.Equal(t, edited_job.Name, job.Name)
		assert.Equal(t, edited_job.Position, "software engineer")
		assert.Equal(t, edited_job.Duration, "6 months")
		assert.Equal(t, edited_job.Description, "make software")
		assert.Equal(t, edited_job.Location, "bangkok")
		assert.Equal(t, edited_job.JobType, model.JobType("internship"))
		assert.Equal(t, edited_job.Experience, model.ExperienceType("internship"))
		assert.Equal(t, edited_job.MinSalary, uint(10))
		assert.Equal(t, edited_job.MaxSalary, uint(100))
	})
	t.Run("Apply", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("applyjobtester-%d", time.Now().UnixNano()),
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
		user := model.User{
			Username: fmt.Sprintf("applyjobtester-%d", time.Now().UnixNano()),
		}
		if err := db.Create(&user).Error; err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&user)
		})()
		// Create dummy files for photo and status photo
		photoFile := model.File{UserID: user.ID, FileType: model.FileTypeJPEG, Category: model.FileCategoryImage}
		if err := db.Create(&photoFile).Error; err != nil {
			t.Error(err)
			return
		}
		statusFile := model.File{UserID: user.ID, FileType: model.FileTypePDF, Category: model.FileCategoryImage}
		if err := db.Create(&statusFile).Error; err != nil {
			t.Error(err)
			return
		}
		student := model.Student{
			UserID:              user.ID,
			ApprovalStatus:      model.StudentApprovalAccepted,
			PhotoID:             photoFile.ID,
			StudentStatusFileID: statusFile.ID,
		}
		if result := db.Create(&student); result.Error != nil {
			t.Error(result.Error)
			return
		}
		job := model.Job{
			CompanyID:      company.UserID,
			ApprovalStatus: model.JobApprovalAccepted,
		}
		if result := db.Create(&job); result.Error != nil {
			t.Error(result.Error)
			return
		}
		values := map[string]string{
			"phone": "0123456789",
			"id":    fmt.Sprintf("%d", job.ID),
			"email": "cool@localhost",
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
		if fiw, err = fw.CreateFormFile("files", "cv.pdf"); err != nil {
			t.Error(err)
			return
		}
		if _, err := io.WriteString(fiw, "[cv content]"); err != nil {
			t.Error(err)
			return
		}
		if fiw, err = fw.CreateFormFile("files", "cv2.pdf"); err != nil {
			t.Error(err)
			return
		}
		if _, err := io.WriteString(fiw, "[cv2 content]"); err != nil {
			t.Error(err)
			return
		}
		if err := fw.Close(); err != nil {
			t.Error(err)
			return
		}
		req, _ := http.NewRequest("POST", "/job/apply", &b)
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
			Error         string `json:"error"`
			ApplicationID uint   `json:"id"`
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
		jobApp := model.JobApplication{
			ID: result.ApplicationID,
		}
		if err := db.First(&jobApp).Error; err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, jobApp.AltPhone, "0123456789")
		assert.Equal(t, jobApp.AltEmail, "cool@localhost")
	})
	t.Run("FetchSelf", func(t *testing.T) {
		var err error
		var userCreationResult *UserCreationResult
		if userCreationResult, err = CreateUser(UserCreationInfo{
			Username:  fmt.Sprintf("fetchselfjobtester-%d", time.Now().UnixNano()),
			IsCompany: true,
		}); err != nil {
			t.Error(err)
			return
		}
		defer (func() {
			_ = db.Delete(&userCreationResult.User)
		})()
		company := userCreationResult.Company
		job := model.Job{
			Name:           fmt.Sprintf("nice-self-job-%d", time.Now().UnixNano()),
			CompanyID:      company.UserID,
			Position:       "software engineer",
			Duration:       "6 months",
			Description:    "make software",
			Location:       "bangkok",
			JobType:        model.JobTypeInternship,
			Experience:     model.ExperienceInternship,
			MinSalary:      10,
			MaxSalary:      100,
			IsOpen:         false,
			ApprovalStatus: model.JobApprovalPending,
		}
		err = db.Create(&job).Error
		if err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/job?companyId=self", strings.NewReader(""))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(company.UserID)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		type JobWithApplicationStatistics struct {
			model.Job
			Pending  int64 `json:"pending"`
			Accepted int64 `json:"accepted"`
			Rejected int64 `json:"rejected"`
		}
		type Result struct {
			Jobs  []JobWithApplicationStatistics `json:"jobs"`
			Error string                         `json:"error"`
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
		assert.Equal(t, len(result.Jobs), 1)
		assert.Equal(t, result.Jobs[0].Position, "software engineer")
	})

}
