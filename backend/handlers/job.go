package handlers

import (
	"fmt"
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"

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
func (h *JobHandlers) CreateJob(ctx *gin.Context) {
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
func (h *JobHandlers) FetchJobs(ctx *gin.Context) {
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

	// Check if the user is a company
	isCompany := false
	company := model.Company{
		UserID: userId,
	}
	result := h.DB.Limit(1).Find(&company)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	} else if result.RowsAffected != 0 {
		isCompany = true
	}

	query := h.DB.Model(&model.Job{})

	// Optional id limit
	if input.JobID != nil {
		query = query.Where(&model.Job{
			ID: *input.JobID,
		})
	}

	// If the user is a company, include job applications statistic
	if isCompany {
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

	// Filter Job post by company ID
	if input.CompanyID != "" {
		if input.CompanyID == "self" {
			input.CompanyID = userId
		}
		query = query.Where("company_id = ?", input.CompanyID)
		if input.Open != nil {
			query = query.Where("is_open = ?", *input.Open)
		}
	} else {
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

	// Check if current user is Admin
	result = h.DB.Limit(1).Find(&model.Admin{
		UserID: userId,
	})
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	} else if result.RowsAffected != 0 || (input.CompanyID == "self" || input.CompanyID == userId) {
		// If is admin, or same company then consider approval status
		if input.ApprovalStatus != "" {
			query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalStatus(input.ApprovalStatus)})
		}
	} else {
		query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalAccepted})
	}
	if isCompany {
		query = query.Group("jobs.id")
	}

	// Offset and Limit
	query = query.Offset(int(input.Offset))
	query = query.Limit(int(input.Limit)).Preload("Company").Joins("INNER JOIN users ON users.id = jobs.company_id")

	// return Job posts with application statistics
	if isCompany {
		var jobsWithStats []JobWithApplicationStatistics
		result = query.Find(&jobsWithStats)
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
	result = query.Select("jobs.*, users.username as company_name").Find(&jobs)
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
func (h *JobHandlers) EditJob(ctx *gin.Context) {
	// Get user id from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")
	// Denied access if user id is not found
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	// Bind request to struct
	type EditJobInput struct {
		Name        *string `json:"name" binding:"omitempty,max=128"`
		ID          uint    `json:"id" binding:"required"`
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

	// Get job post from database
	job := &model.Job{
		ID: input.ID,
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
func (h *JobHandlers) ApproveJob(ctx *gin.Context) {
	type ApproveJobInput struct {
		ID      uint `json:"id" binding:"required"`
		Approve bool `json:"approve"`
	}
	input := ApproveJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job := model.Job{
		ID: input.ID,
	}
	result := h.DB.First(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
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
func (h *JobHandlers) ApplyJob(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)

	// Bind request body to input struct
	type ApplyJobInput struct {
		JobID    uint                    `form:"id" binding:"required"`
		AltPhone string                  `form:"phone" binding:"max=20"`
		AltEmail string                  `form:"email" binding:"max=128"`
		Files    []*multipart.FileHeader `form:"files" binding:"max=2,required"`
	}
	input := ApplyJobInput{}
	err := ctx.Bind(&input)
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
		ID:             input.JobID,
		ApprovalStatus: model.JobApprovalAccepted,
	}
	result = h.DB.First(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	jobApplication := model.JobApplication{
		UserID:   student.UserID,
		JobID:    job.ID,
		AltPhone: input.AltPhone,
		AltEmail: input.AltEmail,
		Status:   model.JobApplicationPending,
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

// Fetch job applications from database
func (h *JobHandlers) FetchJobApplications(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data to struct
	type FetchJobApplicationsInput struct {
		ID     *uint  `json:"id" form:"id"`
		JobID  *uint  `json:"jobId" form:"jobId"`
		Offset uint   `json:"offset" form:"offset"`
		Limit  uint   `json:"limit" form:"limit" binding:"max=64"`
		SortBy string `json:"sortBy" form:"sortBy" binding:"oneof='latest' 'oldest' 'name_az' 'name_za'"`
	}
	input := FetchJobApplicationsInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create Job application with applicant name
	type JobApplicationWithApplicantName struct {
		model.JobApplication
		Username string `json:"username"`
	}
	query := h.DB.Model(&model.JobApplication{}).Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.user_id = job_applications.user_id").Select("job_applications.*", "CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) as username")

	// If ID is provided fetch only that job application
	if input.ID != nil {
		var jobApplication JobApplicationWithApplicantName
		if err := query.Take(&jobApplication).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, jobApplication)
		return
	}

	// If JobID is provided fetch job applications for that job
	if input.JobID != nil {
		query = query.Where(&model.JobApplication{JobID: *input.JobID})
	} else {
		// If JobID is not provided, fetch job applications for the user's company or student profile
		company := model.Company{
			UserID: userId,
		}
		result := h.DB.Limit(1).Find(&company)
		if result.Error != nil {
			// Handle error
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		} else if result.RowsAffected != 0 {
			// Case: User is a company
			query = query.Joins("INNER JOIN jobs on jobs.id = job_applications.job_id").Where("company_id = ?", userId)
		} else {
			// Case: User is a student
			student := model.Student{
				UserID: userId,
			}
			result = h.DB.Limit(1).Find(&student)
			if result.Error != nil {
				// Handle error
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			} else if result.RowsAffected != 0 {
				// Case: User is a student
				query = query.Where(&model.JobApplication{UserID: userId})
			} else {
				// Case: User is neither company nor student
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "neither company nor student"})
				return
			}
		}
	}

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

	// Return job application with name and preloaded files
	var jobApplications []JobApplicationWithApplicantName
	result := query.Offset(int(input.Offset)).Limit(int(input.Limit)).Preload("Files").Scan(&jobApplications)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobApplications)
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
	
	// Get Job application
	jobApplication := model.JobApplication{}
	if err := h.DB.Model(&jobApplication).
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Where("jobs.company_id = ?", userId).
		Where("job_applications.id = ?", input.ID).
		Take(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set new status
	if input.Accept {
		jobApplication.Status = model.JobApplicationAccepted
	} else {
		jobApplication.Status = model.JobApplicationRejected
	}

	// Save
	if err := h.DB.Save(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return ok message
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
