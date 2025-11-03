package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"ku-work/backend/services"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandlers struct {
	DB                                   *gorm.DB
	FileHandlers                         *FileHandlers
	aiService                            *services.AIService
	emailService                         *services.EmailService
	jobApprovalStatusUpdateEmailTemplate *template.Template
}

func NewJobHandlers(db *gorm.DB, aiService *services.AIService, emailService *services.EmailService) (*JobHandlers, error) {
	jobApprovalStatusUpdateEmailTemplate, err := template.New("job_approval_status_update.tmpl").ParseFiles("email_templates/job_approval_status_update.tmpl")
	if err != nil {
		return nil, err
	}
	return &JobHandlers{
		DB:                                   db,
		FileHandlers:                         NewFileHandlers(db),
		aiService:                            aiService,
		emailService:                         emailService,
		jobApprovalStatusUpdateEmailTemplate: jobApprovalStatusUpdateEmailTemplate,
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

	company := model.Company{UserID: userid}
	if err := h.DB.First(&company).Error; err != nil {
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

	if err := h.DB.Create(&job).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": job.ID})

	go h.aiService.AutoApproveJob(&job)
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

	role := helper.GetRole(userId, h.DB)

	query := h.DB.Model(&model.Job{}).
		Joins("INNER JOIN users ON users.id = jobs.company_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id")

	if input.JobID != nil {
		query = query.Where(&model.Job{ID: *input.JobID})
	}

	if input.Keyword != "" {
		keywords := strings.FieldsSeq(input.Keyword) // Split by whitespace
		for keyword := range keywords {
			keywordPattern := fmt.Sprintf("%%%s%%", keyword)
			searchGroup := h.DB.Where("name ILIKE ?", keywordPattern).
				Or("description ILIKE ?", keywordPattern).
				Or("position ILIKE ?", keywordPattern).
				Or("duration ILIKE ?", keywordPattern).
				Or("users.username ILIKE ?", keywordPattern)

			query = query.Where(searchGroup)
		}
	}

	query = query.Where("min_salary >= ?", input.MinSalary)
	query = query.Where("max_salary <= ?", input.MaxSalary)

	if role == helper.Company {
		query = query.Where("company_id = ?", userId)
	} else if input.CompanyID != "" {
		query = query.Where("company_id = ?", input.CompanyID)
	}

	if (role == helper.Company || role == helper.Admin) && input.Open != nil {
		query = query.Where("is_open = ?", *input.Open)
	} else if role == helper.Viewer || role == helper.Student || role == helper.Unknown {
		query = query.Where("is_open = ?", true)
	}

	if len(input.Location) != 0 {
		query = query.Where("location ILIKE ?", input.Location)
	}

	if len(input.JobType) != 0 {
		query = query.Where("job_type IN ?", input.JobType)
	}

	if len(input.Experience) != 0 {
		query = query.Where("experience IN ?", input.Experience)
	}

	if role == helper.Admin || role == helper.Company {
		if input.ApprovalStatus != nil && *input.ApprovalStatus != "" {
			query = query.Where("approval_status = ?", *input.ApprovalStatus)
		}
	} else {
		query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalAccepted})
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query = query.Offset(int(input.Offset)).Limit(int(input.Limit))

	if role == helper.Company {
		var jobsWithStats []JobWithStatsResponse
		result := query.
			Select("jobs.*, users.username as company_name, companies.photo_id, companies.banner_id, COUNT(CASE WHEN job_applications.status = 'pending' THEN 1 END) AS pending, COUNT(CASE WHEN job_applications.status = 'accepted' THEN 1 END) AS accepted, COUNT(CASE WHEN job_applications.status = 'rejected' THEN 1 END) AS rejected").
			Joins("LEFT JOIN job_applications ON job_applications.job_id = jobs.id").
			Group("jobs.id, users.username, companies.photo_id, companies.banner_id").
			Find(&jobsWithStats)

		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"jobs":  jobsWithStats,
			"total": totalCount,
		})
		return
	}

	// return Job posts with company info if not company (include whether current user has applied)
	var jobs []JobResponse
	result := query.
		Select("jobs.*, users.username as company_name, companies.photo_id, companies.banner_id, EXISTS (SELECT 1 FROM job_applications WHERE job_applications.job_id = jobs.id AND job_applications.user_id = ?) AS applied", userId).
		Find(&jobs)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
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
	userid := ctx.MustGet("userID").(string)
	var b bytes.Buffer
	ctx.Request.Body = io.NopCloser(io.TeeReader(ctx.Request.Body, &b))

	var input EditJobInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	needReapproval := input.Name != nil || input.Position != nil || input.Duration != nil || input.Description != nil || input.Location != nil || input.JobType != nil || input.Experience != nil || input.MinSalary != nil || input.MaxSalary != nil

	if ctx.GetBool("ShouldCF") {
		ctx.Request.Body = io.NopCloser(bytes.NewReader(b.Bytes()))
		ctx.Set("DoCF", needReapproval)
		ctx.Next()
		return
	}

	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	job := &model.Job{ID: jobId}
	if err := h.DB.First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

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

	if err := h.DB.Save(&job).Error; err != nil {
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
	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := ctx.Param("id")

	job := model.Job{}
	if err := h.DB.First(&job, "id = ?", jobID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	tx := h.DB.Begin()

	if input.Approve {
		job.ApprovalStatus = model.JobApprovalAccepted
	} else {
		job.ApprovalStatus = model.JobApprovalRejected
	}

	if err := tx.Save(&job).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Create(&model.Audit{
		ActorID:    ctx.MustGet("userID").(string),
		Action:     string(job.ApprovalStatus),
		Reason:     input.Reason,
		ObjectName: "Job",
		ObjectID:   strconv.FormatUint(uint64(job.ID), 10),
	}).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})

	go func() {
		type Context struct {
			Company model.Company
			User    model.User
			Job     model.Job
			Status  string
			Reason  string
		}
		var context Context
		if err := h.DB.Select("email").Take(&context.Company, "user_id = ?", job.CompanyID).Error; err != nil {
			return
		}
		if err := h.DB.Select("username").Take(&context.User, "id = ?", job.CompanyID).Error; err != nil {
			return
		}
		context.Job = job
		context.Status = string(job.ApprovalStatus)
		context.Reason = input.Reason
		var tpl bytes.Buffer
		if err := h.jobApprovalStatusUpdateEmailTemplate.Execute(&tpl, context); err != nil {
			return
		}
		_ = h.emailService.SendTo(
			context.Company.Email,
			fmt.Sprintf("[KU-Work] Your \"%s - %s\" job has been reviewed", job.Name, job.Position),
			tpl.String(),
		)
	}()
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

	var job JobResponse
	if err := h.DB.Model(&model.Job{}).
		Select("jobs.*, users.username as company_name, companies.photo_id, companies.banner_id").
		Joins("INNER JOIN users ON users.id = jobs.company_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id").
		First(&job, jobId).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	applied := false
	if uidVal, ok := ctx.Get("userID"); ok {
		if uidStr, ok2 := uidVal.(string); ok2 && uidStr != "" {
			var count int64
			if err := h.DB.Model(&model.JobApplication{}).
				Where("job_id = ? AND user_id = ?", jobId, uidStr).
				Count(&count).Error; err == nil {
				applied = count > 0
			}
		}
	}
	job.Applied = applied

	ctx.JSON(http.StatusOK, job)
}

// Accept or reject job applications
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

	jobApplication := model.JobApplication{}
	if err := h.DB.Model(&jobApplication).
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Where("jobs.company_id = ?", userId).
		Where("job_applications.id = ?", input.ID).
		Take(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Accept {
		jobApplication.Status = model.JobApplicationAccepted
	} else {
		jobApplication.Status = model.JobApplicationRejected
	}

	if err := h.DB.Save(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
