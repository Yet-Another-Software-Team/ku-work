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

func (h *JobHandlers) CreateJob(ctx *gin.Context) {
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
		Open        bool   `json:"open"`
	}
	input := CreateJobInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
		Name:        input.Name,
		CompanyID:   company.UserID,
		Position:    input.Position,
		Duration:    input.Duration,
		Description: input.Description,
		Location:    input.Location,
		JobType:     model.JobType(input.JobType),
		Experience:  model.ExperienceType(input.Experience),
		MinSalary:   input.MinSalary,
		MaxSalary:   input.MaxSalary,
		IsApproved:  false,
		IsOpen:      input.Open,
	}
	if result := h.DB.Create(&job); result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": job.ID,
	})
}

type JobWithApplicationStatistics struct {
	model.Job
	Pending  int64 `json:"pending"`
	Accepted int64 `json:"accepted"`
	Rejected int64 `json:"rejected"`
}

func (h *JobHandlers) FetchJobs(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type FetchJobsInput struct {
		Limit      uint     `json:"limit" form:"limit" binding:"max=128"`
		Offset     uint     `json:"offset" form:"offset"`
		Location   string   `json:"location" form:"location" binding:"max=128"`
		Keyword    string   `json:"keyword" form:"keyword" binding:"max=256"`
		JobType    []string `json:"jobtype" form:"jobtype" binding:"max=5,dive,max=32"`
		Experience []string `json:"experience" form:"experience" binding:"max=5,dive,max=32"`
		MinSalary  uint     `json:"minsalary" form:"minsalary"`
		MaxSalary  uint     `json:"maxsalary" form:"maxsalary"`
		Open       *bool    `json:"open" form:"open"`
		CompanyID  *string  `json:"companyId" form:"companyId" binding:"omitempty,max=64"`
	}
	input := FetchJobsInput{
		MinSalary: 0,
		MaxSalary: ^uint(0) >> 1,
		Limit:     32,
		Offset:    0,
	}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	keywordPattern := fmt.Sprintf("%%%s%%", input.Keyword)
	query := h.DB.Model(&model.Job{})
	if isCompany {
		query = query.Joins("LEFT JOIN job_applications ON job_applications.job_id = jobs.id")
		query = query.Select("jobs.*, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'pending') AS pending, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'accepted') AS accepted, COUNT(job_applications.id) FILTER(WHERE job_applications.Status = 'rejected') AS rejected")
	}
	query = query.Where(h.DB.Where("name ILIKE ?", keywordPattern).Or("description ILIKE ?", keywordPattern))
	query = query.Where("min_salary >= ?", input.MinSalary)
	query = query.Where("max_salary <= ?", input.MaxSalary)
	if input.CompanyID != nil {
		query = query.Where("company_id = ?", *input.CompanyID)
		if input.Open != nil {
			query = query.Where("is_open = ?", *input.Open)
		}
	} else {
		query = query.Where("is_open = ?", true)
	}
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
	if isCompany {
		query = query.Group("jobs.id")
	}
	query = query.Offset(int(input.Offset))
	query = query.Limit(int(input.Limit)).Preload("Company").Preload("Company.User")
	if isCompany {
		var jobsWithStats []JobWithApplicationStatistics
		result = query.Find(&jobsWithStats)
		for _, job := range jobsWithStats {
			println(job.ID)
		}
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"jobs": jobsWithStats,
		})
		return
	}
	var jobs []model.Job
	result = query.Find(&jobs)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"jobs": jobs,
	})
}

func (h *JobHandlers) EditJob(ctx *gin.Context) {
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)
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
	if job.CompanyID != userid {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
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

func (h *JobHandlers) ApproveJob(ctx *gin.Context) {
	type ApproveJobInput struct {
		ID uint `json:"id" binding:"required"`
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
	job.IsApproved = true
	result = h.DB.Save(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (h *JobHandlers) ApplyJob(ctx *gin.Context) {
	probUserId, hasUserId := ctx.Get("userID")
	if !hasUserId {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userid := probUserId.(string)
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
	student := model.Student{
		UserID: userid,
	}
	result := h.DB.First(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if !student.Approved {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "your student status is not approved yet"})
		return
	}
	job := model.Job{
		ID: input.JobID,
	}
	result = h.DB.First(&job)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if !job.IsApproved {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "job is not approved yet"})
		return
	}
	jobApplication := model.JobApplication{
		UserID:   student.UserID,
		JobID:    job.ID,
		AltPhone: input.AltPhone,
		AltEmail: input.AltEmail,
	}
	success := false
	defer (func() {
		if !success {
			for _, file := range jobApplication.Files {
				_ = h.DB.Delete(&file)
			}
		}
	})()
	for _, file := range input.Files {
		fileObject, err := SaveFile(ctx, h.DB, student.UserID, file, model.FileCategoryDocument)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save file %s: %s", file.Filename, err.Error())})
			return
		}
		jobApplication.Files = append(jobApplication.Files, *fileObject)
	}
	if err := h.DB.Create(&jobApplication).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	success = true
	ctx.JSON(http.StatusOK, gin.H{
		"id": jobApplication.ID,
	})
}

func (h *JobHandlers) FetchJobApplications(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type FetchJobApplicationsInput struct {
		JobID  *uint `json:"id" form:"id"`
		Offset uint  `json:"offset" form:"offset"`
		Limit  uint  `json:"limit" form:"limit" binding:"max=64"`
	}
	input := FetchJobApplicationsInput{}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	query := h.DB.Model(&model.JobApplication{})
	if input.JobID != nil {
		query = query.Where(&model.JobApplication{JobID: *input.JobID})
	} else {
		company := model.Company{
			UserID: userId,
		}
		result := h.DB.Limit(1).Find(&company)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		} else if result.RowsAffected != 0 {
			query = query.Joins("INNER JOIN jobs on jobs.id = job_applications.job_id").Where("company_id = ?", userId)
		} else {
			student := model.Student{
				UserID: userId,
			}
			result = h.DB.Limit(1).Find(&student)
			if result.Error != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
				return
			} else if result.RowsAffected != 0 {
				query = query.Where(&model.JobApplication{UserID: userId})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "neither company nor student"})
				return
			}
		}
	}
	var jobApplications []model.JobApplication
	result := query.Offset(int(input.Offset)).Limit(int(input.Limit)).Preload("Files").Find(&jobApplications)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, jobApplications)
}
