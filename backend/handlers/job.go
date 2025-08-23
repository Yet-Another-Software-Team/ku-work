package handlers

import (
	"fmt"
	"ku-work/backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandlers struct {
	DB *gorm.DB
}

func NewJobHandlers(db *gorm.DB) *JobHandlers {
	return &JobHandlers{
		DB: db,
	}
}

func (h *JobHandlers) CreateJob(ctx *gin.Context) {
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}
	userid := probUserId.(string)
	type CreateJobInput struct {
		Name        string `json:"name" binding:"required,max=128"`
		Position    string `json:"position" binding:"required,max=128"`
		Duration    string `json:"duration" binding:"required,max=128"`
		Description string `json:"description" binding:"required,max=16384"`
		Location    string `json:"location" binding:"required,max=128"`
		JobType     string `json:"jobtype" binding:"required,oneof='fulltime' 'parttime' 'contract' 'casual' 'internship'"`
		Experience  string `json:"experience" binding:"required,oneof='newgrad' 'junior' 'senior' 'manager' 'internship'"`
		MinSalary   uint   `json:"minsalary" binding:"required"`
		MaxSalary   uint   `json:"maxsalary" binding:"required"`
	}
	input := CreateJobInput{}
	err := ctx.Bind(&input)
	if err != nil || input.MaxSalary < input.MinSalary {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	job := model.Job{
		Name: input.Name,
		CompanyID: userid,
		Position: input.Position,
		Duration: input.Duration,
		Description: input.Description,
		Location:    input.Location,
		JobType:     model.JobType(input.JobType),
		Experience:  model.ExperienceType(input.Experience),
		MinSalary:   input.MinSalary,
		MaxSalary:   input.MaxSalary,
		IsApproved:  false,
	}
	result := h.DB.Create(&job)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": job.ID,
	})
}

func (h *JobHandlers) FetchJobs(ctx *gin.Context) {
	type FetchJobsInput struct {
		Limit      uint     `json:"limit" binding:"max=128"`
		Offset     uint     `json:"offset" binding:"max=128"`
		Location   string   `json:"location" binding:"max=128"`
		Keyword    string   `json:"keyword" binding:"max=256"`
		JobType    []string `json:"jobtype" binding:"max=5,dive,max=32"`
		Experience []string `json:"experience" binding:"max=5,dive,max=32"`
		MinSalary  uint     `json:"minsalary"`
		MaxSalary  uint     `json:"maxsalary"`
	}
	input := FetchJobsInput{
		MinSalary: 0,
		MaxSalary: ^uint(0) >> 1,
		Limit:     32,
		Offset:    0,
	}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	keywordPattern := fmt.Sprintf("%%%s%%", input.Keyword)
	query := h.DB.Model(&model.Job{})
	query = query.Where(h.DB.Where("name ILIKE ?", keywordPattern).Or("description ILIKE ?", keywordPattern))
	query = query.Where("min_salary >= ?", input.MinSalary)
	query = query.Where("max_salary <= ?", input.MaxSalary)
	if len(input.Location) != 0 {
		query = query.Where("location = ?", input.Location)
	}
	if len(input.JobType) != 0 {
		query = query.Where("job_type IN ?", input.JobType)
	}
	if len(input.Experience) != 0 {
		query = query.Where("experience IN ?", input.Experience)
	}
	query = query.Where(&model.Job{IsApproved: true})
	query = query.Offset(int(input.Offset))
	query = query.Limit(int(input.Limit))
	var jobs []model.Job
	result := query.Find(&jobs)
	if result.Error != nil {
		ctx.String(http.StatusInternalServerError, result.Error.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}
