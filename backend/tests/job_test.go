package tests

import (
	"encoding/json"
	"fmt"
	"ku-work/backend/handlers"
	"ku-work/backend/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestJob(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		var user *model.User
		var err error
		if user, err = CreateAdminUser("createjobtester"); err != nil {
			t.Error(err)
			return
		}
		w := httptest.NewRecorder()
		payload := `{
	"name": "testjob1",
	"position": "testposition",
	"duration": "forever",
	"description": "ass",
	"location": "thailand",
	"jobtype": "casual",
	"experience": "internship",
	"minsalary": 1,
	"maxsalary": 2
}`
		req, _ := http.NewRequest("POST", "/job", strings.NewReader(payload))
		jwtHandler := handlers.NewJWTHandlers(db)
		jwtToken, _, err := jwtHandler.GenerateTokens(user.ID)
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
		assert.Equal(t, job.Name, "testjob1")
		assert.Equal(t, job.Position, "testposition")
		assert.Equal(t, job.Duration, "forever")
		assert.Equal(t, job.Description, "ass")
		assert.Equal(t, job.Location, "thailand")
		assert.Equal(t, job.JobType, model.JobType("casual"))
		assert.Equal(t, job.Experience, model.ExperienceType("internship"))
		assert.Equal(t, job.MinSalary, uint(1))
		assert.Equal(t, job.MaxSalary, uint(2))
	})
}