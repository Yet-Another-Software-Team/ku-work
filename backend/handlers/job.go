package handlers

import (
	"fmt"
	"ku-work/backend/helper"
	"ku-work/backend/model"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandlers struct {
	DB           *gorm.DB
	FileHandlers *FileHandlers
}

func NewJobHandlers(db *gorm.DB) *JobHandlers {
	return &JobHandlers{
		DB:           db,
		FileHandlers: NewFileHandlers(db),
	}
}

// Create new job listing
//
// CreateJob creates a new job listing in the database.
func (h *JobHandlers) CreateJobHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")

	// Return error if user ID is not found
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	// Bind input data to struct
	type CreateJobInput struct {
		Name        string `json:"name" binding:"required,max=128"`
		Position    string `json:"position" binding:"required,max=128"`
		Duration    string `json:"duration" binding:"required,max=128"`
		Description string `json:"description" binding:"required,max=16384"`
		Location    string `json:"location" binding:"required,max=128"`
		JobType     string `json:"jobtype" binding:"required,oneof='fulltime' 'parttime' 'contract' 'casual' 'internship'"`
		Experience  string `json:"experience" binding:"required,oneof='newgrad' 'junior' 'senior' 'manager' 'internship'"`
		MinSalary   uint   `json:"minsalary"`
		MaxSalary   uint   `json:"maxsalary"`
		Open        bool   `json:"open"`
	}
	input := CreateJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Validate input data
	if input.MaxSalary < input.MinSalary {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "minsalary must be lower than or equal to maxsalary"})
		return
	}
	company := model.Company{
		UserID: userid,
	}
	if result := h.DB.First(&company); result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	job := model.Job{
		Name:           input.Name,
		CompanyID:      company.UserID,
		Position:       input.Position,
		Duration:       input.Duration,
		Description:    input.Description,
		Location:       input.Location,
		JobType:        model.JobType(input.JobType),
		Experience:     model.ExperienceType(input.Experience),
		MinSalary:      input.MinSalary,
		MaxSalary:      input.MaxSalary,
		ApprovalStatus: model.JobApprovalPending,
		IsOpen:         input.Open,
	}

	// Create Job into database
	if result := h.DB.Create(&job); result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// return job ID on success
	ctx.JSON(http.StatusOK, gin.H{
		"id": job.ID,
	})
}

type JobWithApplicationStatistics struct {
	model.Job
	Pending     int64  `json:"pending"`
	Accepted    int64  `json:"accepted"`
	Rejected    int64  `json:"rejected"`
	CompanyName string `json:"companyName"`
}

// Fetch Jobs from Database.
//
// Allow query parameters for filtering jobs.
func (h *JobHandlers) FetchJobsHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	// List of query parameters for filtering jobs
	type FetchJobsInput struct {
		Limit          uint     `json:"limit" form:"limit" binding:"max=128"`
		Offset         uint     `json:"offset" form:"offset"`
		Location       string   `json:"location" form:"location" binding:"max=128"`
		Keyword        string   `json:"keyword" form:"keyword" binding:"max=256"`
		JobType        []string `json:"jobtype" form:"jobtype" binding:"max=5,dive,max=32"`
		Experience     []string `json:"experience" form:"experience" binding:"max=5,dive,max=32"`
		MinSalary      uint     `json:"minsalary" form:"minsalary"`
		MaxSalary      uint     `json:"maxsalary" form:"maxsalary"`
		Open           *bool    `json:"open" form:"open"`
		CompanyID      string   `json:"companyId" form:"companyId" binding:"max=64"`
		JobID          *uint    `json:"id" form:"id" binding:"omitempty,max=64"`
		ApprovalStatus string   `json:"approvalStatus" form:"approvalStatus" binding:"max=64"`
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

	role := helper.GetRole(userId, h.DB)

	query := h.DB.Model(&model.Job{})

	// Optional id limit
	if input.JobID != nil {
		query = query.Where(&model.Job{
			ID: *input.JobID,
		})
	}

	// If the user is a company, include job applications statistic
	if role == helper.Company {
		query = query.Joins("LEFT JOIN job_applications ON job_applications.job_id = jobs.id")
		query = query.Select("jobs.*, ANY_VALUE(users.username) as company_name, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'pending') AS pending, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'accepted') AS accepted, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'rejected') AS rejected")
	}

	// Filter Job post by keyword
	if input.Keyword != "" {
		keywordPattern := fmt.Sprintf("%%%s%%", input.Keyword)
		query = query.Where(h.DB.Where("name ILIKE ?", keywordPattern).Or("description ILIKE ?", keywordPattern))
	}

	// Filter Job post by salary range
	query = query.Where("min_salary >= ?", input.MinSalary)
	query = query.Where("max_salary <= ?", input.MaxSalary)

	// Company should only see their own jobs
	if role == helper.Company {
		query = query.Where("company_id = ?", userId)
		if input.Open != nil {
			query = query.Where("is_open = ?", *input.Open)
		}
	} else {
		// Non-company users can filter by company ID if provided
		if input.CompanyID != "" {
			query = query.Where("company_id = ?", input.CompanyID)
		}
		query = query.Where("is_open = ?", true)
	}

	// Filter Job post by location
	if len(input.Location) != 0 {
		query = query.Where("location = ?", input.Location)
	}

	// Filter Job post by job type
	if len(input.JobType) != 0 {
		query = query.Where("job_type IN ?", input.JobType)
	}

	// Filter Job post by experience
	if len(input.Experience) != 0 {
		query = query.Where("experience IN ?", input.Experience)
	}

	// Only Admin and Company can see unapproved jobs
	if role == helper.Admin || role == helper.Company {
		// If is admin, or company then consider approval status
		if input.ApprovalStatus != "" {
			query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalStatus(input.ApprovalStatus)})
		}
	} else {
		// Non-admin and non-company users can only see approved jobs
		query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalAccepted})
	}

	if role == helper.Company {
		query = query.Group("jobs.id")
	}

	// Offset and Limit
	query = query.Offset(int(input.Offset))
	query = query.Limit(int(input.Limit)).Preload("Company").Joins("INNER JOIN users ON users.id = jobs.company_id")

	// return Job posts with application statistics
	if role == helper.Company {
		var jobsWithStats []JobWithApplicationStatistics
		result := query.Find(&jobsWithStats)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"jobs": jobsWithStats,
		})
		return
	}

	// return Job posts with company info if not company
	type JobWithCompanyInfo struct {
		model.Job
		CompanyName string `json:"companyName"`
	}
	var jobs []JobWithCompanyInfo
	result := query.Select("jobs.*, users.username as company_name").Find(&jobs)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}

// Edit job post that is owned by the user
//
// Support partial update of job post
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
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	// Get job post from database
	job := &model.Job{
		ID: jobId,
	}
	result := h.DB.First(&job)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": result.Error.Error()})
		return
	}

	// Bind request to struct
	type EditJobInput struct {
		Name        *string `json:"name" binding:"omitempty,max=128"`
		Position    *string `json:"position" binding:"omitempty,max=128"`
		Duration    *string `json:"duration" binding:"omitempty,max=128"`
		Description *string `json:"description" binding:"omitempty,max=16384"`
		Location    *string `json:"location" binding:"omitempty,max=128"`
		JobType     *string `json:"jobtype" binding:"omitempty,oneof='fulltime' 'parttime' 'contract' 'casual' 'internship'"`
		Experience  *string `json:"experience" binding:"omitempty,oneof='newgrad' 'junior' 'senior' 'manager' 'internship'"`
		MinSalary   *uint   `json:"minsalary" binding:"omitempty"`
		MaxSalary   *uint   `json:"maxsalary" binding:"omitempty"`
		Open        *bool   `json:"open" binding:"omitempty"`
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "minsalary cannot exceed maxsalary"})
		return
	}

	result = h.DB.Save(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "job updated successfully"})
}

// Handle approval of a job post using its ID
func (h *JobHandlers) JobApprovalHandler(ctx *gin.Context) {
	type ApproveJobInput struct {
		Approve bool `json:"approve"`
	}
	input := ApproveJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get job ID from URL parameter
	jobID := ctx.Param("id")

	job := model.Job{}
	result := h.DB.First(&job, "id = ?", jobID)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	if input.Approve {
		job.ApprovalStatus = model.JobApprovalAccepted
	} else {
		job.ApprovalStatus = model.JobApprovalRejected
	}
	result = h.DB.Save(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// Handle application of a job post using its ID
//
// Use request body of multipart/form-data
func (h *JobHandlers) GetJobDetailHandler(ctx *gin.Context) {
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	job := &model.Job{}
	if err := h.DB.First(job, jobId).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}
