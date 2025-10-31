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
	FileHandlers *FileHandlers
	aiService    *services.AIService
	emailService *services.EmailService
	JobService   *services.JobService
}

func NewJobHandlers(db *gorm.DB, aiService *services.AIService, jobService *services.JobService) (*JobHandlers, error) {
	return &JobHandlers{
		FileHandlers: NewFileHandlers(db),
		aiService:    aiService,
		JobService:   jobService,
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
	NotifyOnApplication bool      `json:"notifyOnApplication"`
}

// JobWithStatsResponse extends JobResponse with application statistics for company users.
type JobWithStatsResponse struct {
	JobResponse
	Pending  int64 `json:"pending"`
	Accepted int64 `json:"accepted"`
	Rejected int64 `json:"rejected"`
}

// @Summary Create a new job listing
// @Description Allows an authenticated company to create a new job posting. The job will be pending approval by an admin.
// @Tags Jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param job body handlers.CreateJobInput true "Job creation data"
// @Success 200 {object} object{id=uint} "Successfully created job listing"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs [post]
func (h *JobHandlers) CreateJobHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")

	// Return error if user ID is not found
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	input := CreateJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
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
	company, err := h.JobService.FindCompanyByUserID(ctx.Request.Context(), userid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Create Job into database via JobService
	if err := h.JobService.CreateJob(ctx.Request.Context(), &job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// return job ID on success
	ctx.JSON(http.StatusOK, gin.H{
		"id": job.ID,
	})

	// Tell AI to approve it for me (if AI service is available)
	if h.aiService != nil {
		go h.aiService.AutoApproveJob(&job)
	}
}

// @Summary Fetch job listings
// @Description Retrieves a list of job postings with extensive filtering options. Behavior changes based on user role. Companies see their own jobs with application stats. Admins can see all jobs. Others see only open, approved jobs.
// @Tags Jobs
// @Security BearerAuth
// @Produce json
// @Param limit query uint false "Pagination limit" default(32)
// @Param offset query uint false "Pagination offset"
// @Param location query string false "Filter by location"
// @Param keyword query string false "Search keyword for name and description"
// @Param jobType query []string false "Filter by job type(s)"
// @Param experience query []string false "Filter by experience level(s)"
// @Param minSalary query uint false "Minimum salary filter"
// @Param maxSalary query uint false "Maximum salary filter"
// @Param open query bool false "Filter by open status (company only)"
// @Param companyId query string false "Filter by company ID"
// @Param id query uint false "Filter by specific job ID"
// @Param approvalStatus query string false "Filter by approval status (admin/company only)"
// @Success 200 {object} object{jobs=[]JobWithStatsResponse} "List of jobs for a company user (includes stats)"
// @Success 200 {object} object{jobs=[]JobResponse} "List of jobs for a non-company user"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs [get]
func (h *JobHandlers) FetchJobsHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	// List of query parameters for filtering jobs
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

	// Set default values for some fields and bind the input
	input := FetchJobsInput{
		MinSalary: 0,
		MaxSalary: ^uint(0) >> 1,
		Limit:     32,
		Offset:    0,
	}
	err := ctx.ShouldBind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Resolve role via JobService so handlers do not access DB directly.
	role, err := h.JobService.ResolveRole(ctx.Request.Context(), userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build params for JobService
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

// @Summary Edit a job listing
// @Description Allows a company to edit one of their own job postings. Supports partial updates.
// @Tags Jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Job ID"
// @Param job body handlers.EditJobInput true "Job update data"
// @Success 200 {object} object{message=string} "Job updated successfully"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 404 {object} object{error=string} "Not Found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id} [patch]
func (h *JobHandlers) EditJobHandler(ctx *gin.Context) {
	// Get user id from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")
	// Denied access if user id is not found
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	// Convert job id to uint
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	// Get job post from database via JobService
	job, err := h.JobService.FindJobByID(ctx.Request.Context(), jobId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var input EditJobInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Denied access if user is not the owner of job post
	if job.CompanyID != userid {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Update job post with new data
	if input.Name != nil {
		job.Name = *input.Name
	}
	if input.Position != nil {
		job.Position = *input.Position
	}
	if input.Duration != nil {
		job.Duration = *input.Duration
	}
	if input.Description != nil {
		job.Description = *input.Description
	}
	if input.Location != nil {
		job.Location = *input.Location
	}
	if input.JobType != nil {
		job.JobType = model.JobType(*input.JobType)
	}
	if input.Open != nil {
		job.IsOpen = *input.Open
	}
	if input.Experience != nil {
		job.Experience = model.ExperienceType(*input.Experience)
	}
	if input.MinSalary != nil {
		job.MinSalary = *input.MinSalary
	}
	if input.MaxSalary != nil {
		job.MaxSalary = *input.MaxSalary
	}
	if job.MinSalary > job.MaxSalary {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "minSalary cannot exceed maxSalary"})
		return
	}
	if input.NotifyOnApplication != nil {
		job.NotifyOnApplication = *input.NotifyOnApplication
	}

	if err := h.JobService.UpdateJob(ctx.Request.Context(), job); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "job updated successfully"})
}

// @Summary Approve or reject a job listing (Admin only)
// @Description Allows an admin to approve or reject a job posting submitted by a company.
// @Tags Jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Job ID"
// @Param approval body handlers.ApproveJobInput true "Approval action"
// @Success 200 {object} object{message=string} "ok"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 404 {object} object{error=string} "Not Found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/approval [post]
func (h *JobHandlers) JobApprovalHandler(ctx *gin.Context) {
	input := ApproveJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get job ID from URL parameter and validate it
	jobIDStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil || jobId64 == 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobID := uint(jobId64)

	// Fetch job via JobService (handlers should not access DB directly)
	job, err := h.JobService.FindJobByID(ctx.Request.Context(), jobID)
	if err != nil {
		// Prefer to return NotFound for missing records; otherwise surface internal error
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use JobService to perform approval/rejection and create audit
	if err := h.JobService.ApproveOrRejectJob(ctx.Request.Context(), job.ID, input.Approve, ctx.MustGet("userID").(string), input.Reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update in-memory job for sending email context
	if input.Approve {
		job.ApprovalStatus = model.JobApprovalAccepted
	} else {
		job.ApprovalStatus = model.JobApprovalRejected
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})

	// Notification email is handled by JobService.ApproveOrRejectJob when JobService is configured with an EmailService/template.
}

// @Summary Get job details
// @Description Retrieves the detailed information for a single job posting by its ID.
// @Tags Jobs
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Success 200 {object} handlers.JobResponse "Job details retrieved successfully"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid job ID"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id} [get]
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

	ctx.JSON(http.StatusOK, jobDetail)
}

// Accept or reject job applications
func (h *JobHandlers) AcceptJobApplication(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data to struct
	type AcceptJobApplicationInput struct {
		ID     uint `json:"id" form:"id"`
		Accept bool `json:"accept" form:"accept"`
	}

	input := AcceptJobApplicationInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use JobService to update application status
	if err := h.JobService.AcceptOrRejectJobApplication(ctx.Request.Context(), userId, input.ID, input.Accept); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return ok message
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
