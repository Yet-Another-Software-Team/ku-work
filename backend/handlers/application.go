package handlers

import (
	"fmt"
	"ku-work/backend/model"
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
}

func NewApplicationHandlers(db *gorm.DB) *ApplicationHandlers {
	return &ApplicationHandlers{
		DB:           db,
		FileHandlers: NewFileHandlers(db),
	}
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
// @Success 200 {object} object{id=uint} "Successfully created job application"
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

	// Check if user is approved student, denied otherwise
	student := model.Student{
		UserID: userid,
	}
	result := h.DB.First(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if student.ApprovalStatus != model.StudentApprovalAccepted {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "your student status is not approved yet"})
		return
	}
	job := model.Job{
		ID:             jobId,
		ApprovalStatus: model.JobApprovalAccepted,
	}
	result = h.DB.First(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Use the student's profile phone and email as the default contact information if none is provided
	if input.AltPhone == "" {
		input.AltPhone = student.Phone
	}
	if input.AltEmail == "" {
		user := model.User{
			ID: student.UserID,
		}
		result = h.DB.First(&user)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		input.AltEmail = user.Username // The username of OAuth User is already set as their email
	}

	jobApplication := model.JobApplication{
		UserID:       student.UserID,
		JobID:        job.ID,
		ContactPhone: input.AltPhone,
		ContactEmail: input.AltEmail,
		Status:       model.JobApplicationPending,
	}
	success := false
	// If create job application fails remove files
	defer (func() {
		if !success {
			for _, file := range jobApplication.Files {
				_ = h.DB.Delete(&file)
			}
		}
	})()

	// Save all files into database and file system
	for _, file := range input.Files {
		fileObject, err := SaveFile(ctx, h.DB, student.UserID, file, model.FileCategoryDocument)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save file %s: %s", file.Filename, err.Error())})
			return
		}
		jobApplication.Files = append(jobApplication.Files, *fileObject)
	}

	// Create application database object
	if err := h.DB.Create(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	success = true
	ctx.JSON(http.StatusOK, gin.H{
		"id": jobApplication.ID,
	})
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
	// Extract authenticated user ID from context
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

	// Build base query joining with users table to fetch applicant username
	// Filter by the job ID from the URL parameter
	query := h.DB.Model(&model.JobApplication{}).
		Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.user_id = job_applications.user_id").
		Joins("INNER JOIN students ON students.user_id = job_applications.user_id").
		Select("job_applications.*",
			"CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) as username",
			"students.major as major",
			"students.student_id as student_id",
			"job_applications.status as status").
		Where("job_applications.job_id = ?", jobId)

	// Filter by status if provided (for tabs: pending, accepted, rejected)
	if input.Status != nil && *input.Status != "" {
		query = query.Where("job_applications.status = ?", *input.Status)
	}

	// Sort results
	switch input.SortBy {
	case "latest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	case "name_az":
		query = query.Order("username ASC")
	case "name_za":
		query = query.Order("username DESC")
	}

	// Execute query with pagination and preload associated files
	var jobApplications []ShortApplicationDetail
	result := query.Offset(int(input.Offset)).Limit(int(input.Limit)).Preload("Files").Scan(&jobApplications)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobApplications)
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
	// Extract authenticated user ID from context
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

	// Delete job applications
	query := h.DB.Where("job_id = ?", jobId)
	if !input.Accepted {
		query = query.Not("status = ?", model.JobApplicationAccepted)
	}
	if !input.Rejected {
		query = query.Not("status = ?", model.JobApplicationRejected)
	}
	if !input.Pending {
		query = query.Not("status = ?", model.JobApplicationPending)
	}
	if err := query.Delete(&model.JobApplication{}).Error; err != nil {
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
// @Param studentId path string true "Student User ID"
// @Success 200 {object} handlers.FullApplicantDetail "Detailed job application"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid job ID"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "Not Found: Job application not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications/{studentId} [get]
func (h *ApplicationHandlers) GetJobApplicationHandler(ctx *gin.Context) {
	// Extract job ID from URL parameter
	jobIdStr := ctx.Param("id")
	jobId64, err := strconv.ParseUint(jobIdStr, 10, 64)
	if err != nil || jobId64 <= 0 || jobId64 > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}
	jobId := uint(jobId64)

	// Extract student ID from URL parameter
	studentId := ctx.Param("studentId")

	// Query for the specific job application with full student details
	var jobApplication FullApplicantDetail

	// Fetch the main application data, without preloading Files.
	query := h.DB.Model(&model.JobApplication{}).
		Joins("INNER JOIN users ON users.id = job_applications.user_id").
		Joins("INNER JOIN students ON students.user_id = job_applications.user_id").
		Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.user_id = job_applications.user_id").
		Select(`job_applications.*,
		 	CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) as username,
			users.username as email,
			students.phone as phone, students.photo_id as photo_id,
			students.birth_date as birth_date, students.about_me as about_me,
			students.git_hub as github, students.linked_in as linked_in,
			students.student_id as student_id, students.major as major`).
		Where("job_applications.job_id = ? AND students.student_id = ?", jobId, studentId)

	if err := query.First(&jobApplication).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job application not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// explicitly load the "Files" association into the struct.
	if err := h.DB.Model(&jobApplication.JobApplication).Association("Files").Find(&jobApplication.Files); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load application files"})
		return
	}

	ctx.JSON(http.StatusOK, jobApplication)
}

// @Summary Get all applications for the current user
// @Description Fetches job applications for the authenticated user. If the user is a company, it returns all applications for all their job postings. If the user is a student, it returns all of their own applications. Supports pagination.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(32)
// @Success 200 {array} handlers.ShortApplicationDetail "List of job applications"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: User is not a company or student"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /applications [get]
func (h *ApplicationHandlers) GetAllJobApplicationsHandler(ctx *gin.Context) {
	// Extract authenticated user ID from context
	userId := ctx.MustGet("userID").(string)

	// Parse and validate query parameters
	type FetchJobApplicationsInput struct {
		Offset uint `json:"offset" form:"offset"`
		Limit  uint `json:"limit" form:"limit" binding:"max=64"`
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

	// Build base query joining with users table to fetch applicant username
	query := h.DB.Model(&model.JobApplication{}).
		Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.id = job_applications.user_id").
		Joins("INNER JOIN students ON students.user_id = job_applications.user_id").
		Select("job_applications.*",
			"CONCAT(google_o_auth_details.FirstName, ' ', google_o_auth_details.LastName) as username",
			"students.major as major",
			"job_applications.status as status")

	// Determine user role and filter applications accordingly
	// Check if user is a company
	company := model.Company{
		UserID: userId,
	}
	result := h.DB.Limit(1).Find(&company)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected != 0 {
		// User is a company: fetch applications for all their job postings
		query = query.Joins("INNER JOIN jobs on jobs.id = job_applications.job_id").Where("jobs.company_id = ?", userId)
	} else {
		// Check if user is a student
		student := model.Student{
			UserID: userId,
		}
		result = h.DB.Limit(1).Find(&student)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		if result.RowsAffected != 0 {
			// User is a student: fetch only their own applications
			query = query.Where("job_applications.user_id = ?", userId)
		} else {
			// User is neither company nor student: deny access
			ctx.JSON(http.StatusForbidden, gin.H{"error": "user is neither company nor student"})
			return
		}
	}

	// Execute query with pagination and preload associated files
	var jobApplications []ShortApplicationDetail
	result = query.Offset(int(input.Offset)).Limit(int(input.Limit)).Preload("Files").Find(&jobApplications)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusOK, jobApplications)
}

// @Summary Update job application status
// @Description Updates the status of a job application to 'accepted', 'rejected', or 'pending'. This action can only be performed by the company that posted the job.
// @Tags Job Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Job ID"
// @Param studentId path string true "Student User ID"
// @Param status body handlers.UpdateJobApplicationStatusHandler.UpdateStatusInput true "New status"
// @Success 200 {object} object{message=string, status=string} "Application status updated successfully"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid ID or input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden: User is not authorized to update this application"
// @Failure 404 {object} object{error=string} "Not Found: Job or application not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /jobs/{id}/applications/{studentId} [patch]
func (h *ApplicationHandlers) UpdateJobApplicationStatusHandler(ctx *gin.Context) {
	// Extract authenticated user ID from context
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
	studentId := ctx.Param("studentId")

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

	// Find the job application
	jobApplication := &model.JobApplication{}
	if err := h.DB.Model(&model.JobApplication{}).
		Joins("INNER JOIN students ON job_applications.user_id = students.user_id").
		Where("job_id = ? AND students.student_id = ?", jobId, studentId).First(jobApplication).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "job application not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Update the status
	jobApplication.Status = model.JobApplicationStatus(input.Status)
	if err := h.DB.Save(jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "application status updated successfully",
		"status":  jobApplication.Status,
	})
}
