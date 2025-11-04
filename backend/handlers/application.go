package handlers

import (
	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	"ku-work/backend/services"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApplicationHandlers struct {
	DB           *gorm.DB
	FileHandlers *FileHandlers
	appService   *services.ApplicationService
}

func NewApplicationHandlers(db *gorm.DB, fileHandlers *FileHandlers, appService *services.ApplicationService) (*ApplicationHandlers, error) {
	return &ApplicationHandlers{
		DB:           db,
		FileHandlers: fileHandlers,
		appService:   appService,
	}, nil
}

// ApplyJobInput defines the structure for the job application form data.
type ApplyJobInput struct {
	AltPhone string                  `form:"phone" binding:"max=20"`
	AltEmail string                  `form:"email" binding:"max=128"`
	Files    []*multipart.FileHeader `form:"files" binding:"max=2,required"`
}

// ShortApplicationDetail defines the response structure including the applicant's name.
type ShortApplicationDetail struct {
	model.JobApplication
	Username  string `json:"username"`
	Major     string `json:"major"`
	StudentID string `json:"studentId"`
	Status    string `json:"status"`
}

type ApplicationWithJobDetails struct {
	model.JobApplication
	JobPosition   string `json:"position"`
	JobName       string `json:"jobName"`
	CompanyName   string `json:"companyName"`
	CompanyLogoID string `json:"photoId"`
	JobType       string `json:"jobType"`
	Experience    string `json:"experience"`
	MinSalary     uint   `json:"minSalary"`
	MaxSalary     uint   `json:"maxSalary"`
	IsOpen        bool   `json:"isOpen"`
}

// FullApplicantDetail defines the response structure for a detailed application view.
type FullApplicantDetail struct {
	model.JobApplication
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   *string   `json:"photoId"`
	BirthDate time.Time `json:"birthDate"`
	AboutMe   string    `json:"aboutMe"`
	GitHub    string    `json:"github"`
	LinkedIn  string    `json:"linkedIn"`
	StudentID string    `json:"studentId"`
	Major     string    `json:"major"`
}

// @Summary Apply to a job
// @Description Creates a new job application. Allows an approved student to apply to an approved job posting by submitting their application with optional alternate contact information and required document files (e.g., resume, cover letter).
// @Tags Job Applications
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path uint true "Job ID"
// @Param Files formData file true "Files to upload (e.g., Resume, Cover Letter). Max 2 files."
// @Param AltPhone formData string false "Alternate phone number"
// @Param AltEmail formData string false "Alternate email address"
// @Success 200 {object} object{message=string} "Successfully created job application"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: Student status not approved"
// @Failure 404 {object} object{error=string} "Not Found: Invalid Job ID"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/apply [post]
func (h *ApplicationHandlers) CreateJobApplicationHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	input := ApplyJobInput{}
	err = ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	params := services.ApplyToJobParams{
		UserID:       userid,
		JobID:        jobId,
		ContactPhone: input.AltPhone,
		ContactEmail: input.AltEmail,
		Files:        input.Files,
		GinCtx:       ctx,
	}

	if err := h.appService.ApplyToJob(ctx.Request.Context(), params); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// @Summary Get applications for a specific job
// @Description Fetches all job applications for a specific job posting. This endpoint is for companies to view applicants. It supports status filtering (pending, accepted, rejected) and pagination.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param status query string false "Filter by status (pending, accepted, rejected)"
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(32)
// @Success 200 {array} handlers.ShortApplicationDetail "List of job applications"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid job ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: User is not authorized to view these applications"
// @Failure 404 {object} object{error=string} "Not Found: Job not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications [get]
func (h *ApplicationHandlers) GetJobApplicationsHandler(ctx *gin.Context) {

	userId := ctx.MustGet("userID").(string)

	// Convert job id to uint from URL parameter
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	// Verify the job exists and check authorization
	job := &model.Job{}
	if err := h.DB.First(job, jobId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Check if user is authorized to view applications for this job
	// Only the company that posted the job or an admin can view its applications
	if job.CompanyID != userId {
		// Check if user is an admin
		admin := model.Admin{}
		result := h.DB.Where("user_id = ?", userId).First(&admin)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		if result.RowsAffected == 0 {
			// User is not an admin and not the company that posted the job
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden: only the company that posted this job or an admin can view its applications"})
			return
		}
	}

	// Parse and validate query parameters
	type FetchJobApplicationsInput struct {
		Status *string `json:"status" form:"status" binding:"omitempty,max=64"`
		Offset uint    `json:"offset" form:"offset"`
		Limit  uint    `json:"limit" form:"limit" binding:"max=64"`
		SortBy string  `json:"sortBy" form:"sortBy" binding:"oneof='latest' 'oldest' 'name_az' 'name_za'"`
	}

	input := FetchJobApplicationsInput{}
	err = ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set default limit if not provided
	if input.Limit == 0 {
		input.Limit = 32
	}

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	params := repo.FetchJobApplicationsParams{
		Status: input.Status,
		Offset: input.Offset,
		Limit:  input.Limit,
		SortBy: input.SortBy,
	}
	rows, err := h.appService.GetApplicationsForJob(ctx.Request.Context(), jobId, &params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, rows)
}

// @Summary Delete job applications for a job
// @Description Delete all job applications for a job, includes filter for only rejected, pending, or accepted jobs.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param pending body handlers.ClearJobApplicationsHandler.ClearJobApplicationsInput true "Whether to include pending job applications or not"
// @Param accepted body handlers.ClearJobApplicationsHandler.ClearJobApplicationsInput true "Whether to include accepted job applications or not"
// @Param rejected body handlers.ClearJobApplicationsHandler.ClearJobApplicationsInput true "Whether to include rejected job applications or not"
// @Success 200 {object} object{message=string} "Success"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid job ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "Not Found: Job not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications [get]
func (h *ApplicationHandlers) ClearJobApplicationsHandler(ctx *gin.Context) {

	userId := ctx.MustGet("userID").(string)

	// Convert job id to uint from URL parameter
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	jobId := uint(jobId64)

	// Parse parameters
	type ClearJobApplicationsInput struct {
		Pending  bool `json:"pending"`
		Rejected bool `json:"rejected"`
		Accepted bool `json:"accepted"`
	}
	var input ClearJobApplicationsInput
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the job exists and check authorization
	job := &model.Job{
		ID:        jobId,
		CompanyID: userId,
	}
	if err := h.DB.Take(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	if _, err := h.appService.ClearJobApplications(ctx.Request.Context(), jobId, input.Pending, input.Rejected, input.Accepted); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// @Summary Get a specific job application
// @Description Retrieves detailed information about a single job application for a specific student, including the applicant's full profile, contact information, and attached files (resume, etc.).
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param email query string true "Student Email"
// @Success 200 {object} handlers.FullApplicantDetail "Detailed job application"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid job ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "Not Found: Job application not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications/{email} [get]
func (h *ApplicationHandlers) GetJobApplicationHandler(ctx *gin.Context) {

	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	email := ctx.Param("email")

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	detail, err := h.appService.GetApplicationByJobAndEmail(ctx.Request.Context(), jobId, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job application not found or student account deactivated"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, detail)
}

// @Summary Get all applications for the current user
// @Description Fetches job applications for the authenticated user. If the user is a company, it returns all applications for all their job postings. If the user is a student, it returns all of their own applications. Supports pagination and status filtering.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (pending, accepted, rejected)"
// @Param sortBy query string false "Sort by (name, date-desc, date-asc)" default(date-desc)
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(32)
// @Success 200 {object} object{applications=[]handlers.ApplicationWithJobDetails,total=int} "List of job applications with total count"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: User is not a company or student"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /applications [get]
func (h *ApplicationHandlers) GetAllJobApplicationsHandler(ctx *gin.Context) {

	userId := ctx.MustGet("userID").(string)

	// Parse and validate query parameters
	type FetchJobApplicationsInput struct {
		Status *string `json:"status" form:"status" binding:"omitempty,oneof=pending accepted rejected"`
		SortBy string  `json:"sortBy" form:"sortBy" binding:"omitempty,oneof=name date-desc date-asc"`
		Offset uint    `json:"offset" form:"offset"`
		Limit  uint    `json:"limit" form:"limit" binding:"max=64"`
	}
	input := FetchJobApplicationsInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set default limit if not provided
	if input.Limit == 0 {
		input.Limit = 32
	}

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	params := repo.FetchAllApplicationsParams{
		Status: input.Status,
		SortBy: input.SortBy,
		Offset: input.Offset,
		Limit:  input.Limit,
	}

	rows, total, err := h.appService.GetAllApplicationsForUser(ctx.Request.Context(), userId, &params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"applications": rows,
		"total":        total,
	})
}

// @Summary Update job application status
// @Description Updates the status of a job application to 'accepted', 'rejected', or 'pending'. This action can only be performed by the company that posted the job.
// @Tags Job Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Job ID"
// @Param studentUserId path string true "Student User ID"
// @Param status body handlers.UpdateJobApplicationStatusHandler.UpdateStatusInput true "New status"
// @Success 200 {object} object{message=string, status=string} "Application status updated successfully"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid ID or input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: User is not authorized to update this application"
// @Failure 404 {object} object{error=string} "Not Found: Job or application not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications/{studentUserId} [patch]
func (h *ApplicationHandlers) UpdateJobApplicationStatusHandler(ctx *gin.Context) {

	userId := ctx.MustGet("userID").(string)

	// Convert job id to uint from URL parameter
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	// Extract student ID from URL parameter
	studentUserId := ctx.Param("studentUserId")

	// Verify the job exists and check authorization
	job := &model.Job{}
	if err := h.DB.First(job, jobId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Check if user is authorized to update applications for this job
	// Only the company that posted the job
	if job.CompanyID != userId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden: only the company that posted this job"})
		return
	}

	// Parse input data
	type UpdateStatusInput struct {
		Status string `json:"status" binding:"required,oneof=accepted rejected pending"`
	}
	input := UpdateStatusInput{}
	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.appService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "application service not configured"})
		return
	}

	// Best-effort collect email details for notification
	var applicantEmail string
	var oauth model.GoogleOAuthDetails
	if err := h.DB.Select("email").Where("user_id = ?", studentUserId).Take(&oauth).Error; err == nil {
		applicantEmail = oauth.Email
	}
	var companyName string
	_ = h.DB.Model(&model.User{ID: job.CompanyID}).Pluck("username", &companyName).Error

	params := services.UpdateStatusParams{
		JobID:                jobId,
		StudentUserID:        studentUserId,
		NewStatus:            model.JobApplicationStatus(input.Status),
		NotifyApplicantEmail: applicantEmail,
		CompanyName:          companyName,
	}
	if err := h.appService.UpdateJobApplicationStatus(ctx.Request.Context(), params); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job application not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "application status updated successfully",
		"status":  model.JobApplicationStatus(input.Status),
	})
}
