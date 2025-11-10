package handlers

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"ku-work/backend/model"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandlers struct {
	aiService  *services.AIService
	JobService *services.JobService
}

func NewJobHandlers(aiService *services.AIService, jobService *services.JobService) (*JobHandlers, error) {
	return &JobHandlers{
		aiService:  aiService,
		JobService: jobService,
	}, nil
}

// CreateJobInput defines the request body for creating a new job.
type CreateJobInput struct {
	Name                string `json:"name" binding:"required,max=128"`
	Position            string `json:"position" binding:"required,max=128"`
	Duration            string `json:"duration" binding:"required,max=128"`
	Description         string `json:"description" binding:"required,max=16384"`
	Location            string `json:"location" binding:"required,max=128"`
	JobType             string `json:"jobType" binding:"required,oneof='fulltime' 'parttime' 'contract' 'casual' 'internship'"`
	Experience          string `json:"experience" binding:"required,oneof='newgrad' 'junior' 'senior' 'manager' 'internship'"`
	MinSalary           *uint  `json:"minSalary" binding:"required"`
	MaxSalary           *uint  `json:"maxSalary" binding:"required"`
	Open                bool   `json:"open"`
	NotifyOnApplication *bool  `json:"notifyOnApplication"`
}

// EditJobInput defines the request body for editing an existing job.
type EditJobInput struct {
	Name                *string `json:"name" binding:"omitempty,max=128"`
	Position            *string `json:"position" binding:"omitempty,max=128"`
	Duration            *string `json:"duration" binding:"omitempty,max=128"`
	Description         *string `json:"description" binding:"omitempty,max=16384"`
	Location            *string `json:"location" binding:"omitempty,max=128"`
	JobType             *string `json:"jobType" binding:"omitempty,oneof='fulltime' 'parttime' 'contract' 'casual' 'internship'"`
	Experience          *string `json:"experience" binding:"omitempty,oneof='newgrad' 'junior' 'senior' 'manager' 'internship'"`
	MinSalary           *uint   `json:"minSalary" binding:"omitempty"`
	MaxSalary           *uint   `json:"maxSalary" binding:"omitempty"`
	Open                *bool   `json:"open" binding:"omitempty"`
	NotifyOnApplication *bool   `json:"notifyOnApplication" binding:"omitempty"`
}

// ApproveJobInput defines the request body for approving a job.
type ApproveJobInput struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason" binding:"max=16384"`
}

// JobResponse defines the structure for a single job listing in API responses.
type JobResponse struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Name                string    `json:"name"`
	CompanyID           string    `json:"companyId"`
	PhotoID             string    `json:"photoId"`
	BannerID            string    `json:"bannerId"`
	CompanyName         string    `json:"companyName"`
	Position            string    `json:"position"`
	Duration            string    `json:"duration"`
	Description         string    `json:"description"`
	Location            string    `json:"location"`
	JobType             string    `json:"jobType"`
	Experience          string    `json:"experience"`
	MinSalary           uint      `json:"minSalary"`
	MaxSalary           uint      `json:"maxSalary"`
	ApprovalStatus      string    `json:"approvalStatus"`
	IsOpen              bool      `json:"open"`
	Applied             bool      `json:"applied"`
	NotifyOnApplication bool      `json:"notifyOnApplication"`
}

// JobWithStatsResponse extends JobResponse with application statistics.
type JobWithStatsResponse struct {
	JobResponse
	Pending  int64 `json:"pending"`
	Accepted int64 `json:"accepted"`
	Rejected int64 `json:"rejected"`
}

// CreateJobHandler handles job creation using the service layer.
func (h *JobHandlers) CreateJobHandler(ctx *gin.Context) {
	userid := ctx.MustGet("userID").(string)

	input := CreateJobInput{}
	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input data
	if input.MinSalary == nil || input.MaxSalary == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "minSalary and maxSalary are required"})
		return
	}
	if *input.MaxSalary < *input.MinSalary {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "minSalary must be lower than or equal to maxSalary"})
		return
	}

	// Resolve company via service abstraction
	company, err := h.JobService.FindCompanyByUserID(ctx.Request.Context(), userid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if company == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "company record not found"})
		return
	}

	if input.NotifyOnApplication == nil {
		defaultNotify := true
		input.NotifyOnApplication = &defaultNotify
	}

	job := model.Job{
		Name:                input.Name,
		CompanyID:           company.UserID,
		Position:            input.Position,
		Duration:            input.Duration,
		Description:         input.Description,
		Location:            input.Location,
		JobType:             model.JobType(input.JobType),
		Experience:          model.ExperienceType(input.Experience),
		MinSalary:           *input.MinSalary,
		MaxSalary:           *input.MaxSalary,
		ApprovalStatus:      model.JobApprovalPending,
		IsOpen:              input.Open,
		NotifyOnApplication: *input.NotifyOnApplication,
	}

	// Persist via service
	if err := h.JobService.CreateJob(ctx.Request.Context(), &job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger AI auto-approval if configured (best-effort)
	if h.aiService != nil {
		_ = h.aiService.PublishAsyncJobCheck(job.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{"id": job.ID})
}

// FetchJobsHandler delegates job listing retrieval to the JobService.
func (h *JobHandlers) FetchJobsHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)

	type FetchJobsInput struct {
		Limit          uint     `json:"limit" form:"limit" binding:"max=128"`
		Offset         uint     `json:"offset" form:"offset"`
		Location       string   `json:"location" form:"location" binding:"max=128"`
		Keyword        string   `json:"keyword" form:"keyword" binding:"max=256"`
		JobType        []string `json:"jobType" form:"jobType" binding:"max=5,dive,max=32"`
		Experience     []string `json:"experience" form:"experience" binding:"max=5,dive,max=32"`
		MinSalary      uint     `json:"minSalary" form:"minSalary"`
		MaxSalary      uint     `json:"maxSalary" form:"maxSalary"`
		Open           *bool    `json:"open" form:"open"`
		CompanyID      string   `json:"companyId" form:"companyId" binding:"max=64"`
		JobID          *uint    `json:"id" form:"id" binding:"omitempty,max=64"`
		ApprovalStatus *string  `json:"approvalStatus" form:"approvalStatus" binding:"omitempty,oneof=pending accepted rejected"`
	}

	input := FetchJobsInput{
		MinSalary: 0,
		MaxSalary: ^uint(0) >> 1,
		Limit:     32,
		Offset:    0,
	}
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve role using service-level resolver so handlers don't access DB directly.
	role, err := h.JobService.ResolveRole(ctx.Request.Context(), userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build params for service
	params := services.FetchJobsParams{
		Limit:          input.Limit,
		Offset:         input.Offset,
		Location:       input.Location,
		Keyword:        input.Keyword,
		JobType:        input.JobType,
		Experience:     input.Experience,
		MinSalary:      input.MinSalary,
		MaxSalary:      input.MaxSalary,
		Open:           input.Open,
		CompanyID:      input.CompanyID,
		JobID:          input.JobID,
		ApprovalStatus: input.ApprovalStatus,
		Role:           role,
		UserID:         userId,
	}

	jobsResult, totalCount, err := h.JobService.FetchJobs(ctx.Request.Context(), &params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"jobs":  jobsResult,
		"total": totalCount,
	})
}

// EditJobHandler updates a job via the service.
func (h *JobHandlers) EditJobHandler(ctx *gin.Context) {
	userid := ctx.MustGet("userID").(string)

	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	// Retrieve job through service
	job, err := h.JobService.FindJobByID(ctx.Request.Context(), jobId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	var input EditJobInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if job.CompanyID != userid {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	needReapproval := false

	// Update job post with new data
	if input.Name != nil {
		needReapproval = true
		job.Name = *input.Name
	}
	if input.Position != nil {
		needReapproval = true
		job.Position = *input.Position
	}
	if input.Duration != nil {
		needReapproval = true
		job.Duration = *input.Duration
	}
	if input.Description != nil {
		needReapproval = true
		job.Description = *input.Description
	}
	if input.Location != nil {
		needReapproval = true
		job.Location = *input.Location
	}
	if input.JobType != nil {
		needReapproval = true
		job.JobType = model.JobType(*input.JobType)
	}
	if input.Open != nil {
		job.IsOpen = *input.Open
	}
	if input.Experience != nil {
		needReapproval = true
		job.Experience = model.ExperienceType(*input.Experience)
	}
	if input.MinSalary != nil {
		needReapproval = true
		job.MinSalary = *input.MinSalary
	}
	if input.MaxSalary != nil {
		needReapproval = true
		job.MaxSalary = *input.MaxSalary
	}
	// Check for invalid salary range
	if job.MinSalary > job.MaxSalary {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "minSalary cannot exceed maxSalary"})
		return
	}
	if input.NotifyOnApplication != nil {
		job.NotifyOnApplication = *input.NotifyOnApplication
	}

	if needReapproval {
		job.ApprovalStatus = model.JobApprovalPending
	}

	if err := h.JobService.UpdateJob(ctx.Request.Context(), job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "job updated successfully"})
}

// JobApprovalHandler approves or rejects a job by delegating to JobService.
// JobService handles audit creation and email notifications if configured.
func (h *JobHandlers) JobApprovalHandler(ctx *gin.Context) {
	input := ApproveJobInput{}
	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse job ID
	jobIDStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil || jobId64 == 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobID := uint(jobId64)

	actorID := ctx.MustGet("userID").(string)
	if err := h.JobService.ApproveOrRejectJob(ctx.Request.Context(), jobID, input.Approve, actorID, input.Reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// GetJobDetailHandler returns detailed job information via the service.
func (h *JobHandlers) GetJobDetailHandler(ctx *gin.Context) {
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	jobDetail, err := h.JobService.GetJobDetail(ctx.Request.Context(), jobId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Note: 'Applied' status is omitted here to keep handlers strictly using service APIs.
	ctx.JSON(http.StatusOK, jobDetail)
}

// AcceptJobApplication delegates application acceptance/rejection to the JobService.
func (h *JobHandlers) AcceptJobApplication(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)

	type AcceptJobApplicationInput struct {
		ID     uint `json:"id" form:"id"`
		Accept bool `json:"accept" form:"accept"`
	}

	input := AcceptJobApplicationInput{}
	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.JobService.AcceptOrRejectJobApplication(ctx.Request.Context(), userId, input.ID, input.Accept); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
