package handlers

import (
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
	// probUserId, hasUserId := ctx.Get("userid")
	// if !hasUserId {
	// 	ctx.String(http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }
	// userid := probUserId.(uint)
	userid := 1
	var user model.User
	userQueryResult := h.DB.First(&user, userid)
	if userQueryResult.Error != nil {
		ctx.String(http.StatusInternalServerError, userQueryResult.Error.Error())
		return
	}
	type CreateJobInput struct {
		Name        string `json:"name" binding:"required,max=128"`
		Position    string `json:"position" binding:"required,max=128"`
		Duration    string `json:"duration" binding:"required,max=128"`
		Description string `json:"description" binding:"required,max=16384"`
		Location    string `json:"location" binding:"required,max=128"`
		JobType     string `json:"jobtype" binding:"required,oneof='fulltime' 'parttime' 'contract' 'casual'"`
		Experience  string `json:"experience" binding:"required,oneof='newgrad' 'junior' 'senior' 'manager'"`
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
		CompanyID: user.ID,
		Position: input.Position,
		Duration: input.Duration,
		Description: input.Description,
		Location: input.Location,
		JobType: model.JobType(input.JobType),
		Experience: model.ExperienceType(input.Experience),
		MinSalary: input.MinSalary,
		MaxSalary: input.MaxSalary,
		IsApproved: false,
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
