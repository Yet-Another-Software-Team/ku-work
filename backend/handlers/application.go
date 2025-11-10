package handlers

import (
	"math"
	"mime/multipart"
	"net/http"
	"strconv"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

// ApplicationHandlers exposes HTTP handlers for job application related endpoints.
// It performs only input extraction and validation, delegating ALL business logic
// (authorization, existence checks, status rules, email notifications, etc.)
// / to the ApplicationService layer.
type ApplicationHandlers struct {
	appService *services.ApplicationService
}

// NewApplicationHandlers constructs a new handlers wrapper.
// Returns an error if the service dependency is missing.
func NewApplicationHandlers(appService *services.ApplicationService) *ApplicationHandlers {
	if appService == nil {
		panic("missing dependency: ApplicationService")
	}
	return &ApplicationHandlers{appService: appService}
}

// ApplyJobInput represents the multipart form fields for creating a job application.
// Files are required and limited to a maximum of 2 documents (e.g., resume, cover letter).
type ApplyJobInput struct {
	AltPhone string                  `form:"phone" binding:"omitempty,max=20"`
	AltEmail string                  `form:"email" binding:"omitempty,max=128"`
	Files    []*multipart.FileHeader `form:"files" binding:"required,max=2"`
}

// @Summary Apply to a job
// @Description Submit a new application (approved student applying to an approved job). Files must include up to 2 documents.
// @Tags Job Applications
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path uint true "Job ID"
// @Param phone formData string false "Alternate phone number"
// @Param email formData string false "Alternate email address"
// @Param files formData file true "Application files (max 2)"
// @Success 200 {object} object{message=string} "OK"
// @Failure 400 {object} object{error=string} "Invalid input"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Internal error"
// @Router /jobs/{id}/apply [post]
func (h *ApplicationHandlers) CreateJobApplicationHandler(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	jobID, ok := parseUintPath(ctx, "id")
	if !ok {
		return
	}

	var input ApplyJobInput
	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := services.ApplyToJobParams{
		UserID:       userID,
		JobID:        jobID,
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

// @Summary List applications for a job
// @Description Returns applications for a specific job (authorization enforced in service layer).
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param status query string false "Status filter (pending|accepted|rejected)"
// @Param offset query int false "Pagination offset"
// @Param limit query int false "Pagination limit (max 64)"
// @Param sortBy query string false "Sort (latest|oldest|name_az|name_za)"
// @Success 200 {array} repository.ShortApplicationDetail
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /jobs/{id}/applications [get]
func (h *ApplicationHandlers) GetJobApplicationsHandler(ctx *gin.Context) {
	jobID, ok := parseUintPath(ctx, "id")
	if !ok {
		return
	}

	type Query struct {
		Status *string `form:"status" binding:"omitempty,oneof=pending accepted rejected"`
		Offset uint    `form:"offset"`
		Limit  uint    `form:"limit" binding:"omitempty,max=64"`
		SortBy string  `form:"sortBy" binding:"omitempty,oneof=latest oldest name_az name_za"`
	}
	var q Query
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.Limit == 0 {
		q.Limit = 32
	}

	params := repo.FetchJobApplicationsParams{
		Status: q.Status,
		Offset: q.Offset,
		Limit:  q.Limit,
		SortBy: q.SortBy,
	}
	rows, err := h.appService.GetApplicationsForJob(ctx.Request.Context(), jobID, &params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, rows)
}

// @Summary Clear job applications
// @Description Deletes applications for a job filtered by status flags.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param pending body bool false "Include pending"
// @Param rejected body bool false "Include rejected"
// @Param accepted body bool false "Include accepted"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /jobs/{id}/applications [delete]
func (h *ApplicationHandlers) ClearJobApplicationsHandler(ctx *gin.Context) {
	jobID, ok := parseUintPath(ctx, "id")
	if !ok {
		return
	}

	type Body struct {
		Pending  bool `json:"pending"`
		Rejected bool `json:"rejected"`
		Accepted bool `json:"accepted"`
	}
	var b Body
	if err := ctx.ShouldBindJSON(&b); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.appService.ClearJobApplications(ctx.Request.Context(), jobID, b.Pending, b.Rejected, b.Accepted); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// @Summary Get a single job application
// @Description Returns detailed application info for a student email on a job.
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param id path uint true "Job ID"
// @Param email path string true "Student email"
// @Success 200 {object} repository.FullApplicantDetail
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /jobs/{id}/applications/{email} [get]
func (h *ApplicationHandlers) GetJobApplicationHandler(ctx *gin.Context) {
	jobID, ok := parseUintPath(ctx, "id")
	if !ok {
		return
	}
	email := ctx.Param("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	detail, err := h.appService.GetApplicationByJobAndEmail(ctx.Request.Context(), jobID, email)
	if err != nil {
		// Service decides if not found vs internal; here treat nil detail as not found.
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if detail == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}
	ctx.JSON(http.StatusOK, detail)
}

// @Summary List current user's applications
// @Description Returns applications belonging to the authenticated user (company or student).
// @Tags Job Applications
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter (pending|accepted|rejected)"
// @Param sortBy query string false "Sort (name|date-desc|date-asc)"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit (max 64)"
// @Success 200 {object} object{applications=[]repository.ApplicationWithJobDetails,total=int}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /applications [get]
func (h *ApplicationHandlers) GetAllJobApplicationsHandler(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	type Query struct {
		Status *string `form:"status" binding:"omitempty,oneof=pending accepted rejected"`
		SortBy string  `form:"sortBy" binding:"omitempty,oneof=name date-desc date-asc"`
		Offset uint    `form:"offset"`
		Limit  uint    `form:"limit" binding:"omitempty,max=64"`
	}
	var q Query
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.Limit == 0 {
		q.Limit = 32
	}

	params := repo.FetchAllApplicationsParams{
		Status: q.Status,
		SortBy: q.SortBy,
		Offset: q.Offset,
		Limit:  q.Limit,
	}
	rows, total, err := h.appService.GetAllApplicationsForUser(ctx.Request.Context(), userID, &params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"applications": rows, "total": total})
}

// UpdateStatusInput is the body payload for updating an application status.
type UpdateStatusInput struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected pending"`
}

// @Summary Update application status
// @Description Changes the status of a student's application for a job.
// @Tags Job Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path uint true "Job ID"
// @Param studentUserId path string true "Student User ID"
// @Param body body handlers.UpdateStatusInput true "New status"
// @Success 200 {object} object{message=string,status=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /jobs/{id}/applications/{studentUserId} [patch]
func (h *ApplicationHandlers) UpdateJobApplicationStatusHandler(ctx *gin.Context) {
	jobID, ok := parseUintPath(ctx, "id")
	if !ok {
		return
	}
	studentUserID := ctx.Param("studentUserId")
	if studentUserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "studentUserId required"})
		return
	}

	var input UpdateStatusInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := services.UpdateStatusParams{
		JobID:         jobID,
		StudentUserID: studentUserID,
		NewStatus:     model.JobApplicationStatus(input.Status),
		CompanyName:   "", // Service can populate if needed
	}

	if err := h.appService.UpdateJobApplicationStatus(ctx.Request.Context(), params); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "application status updated successfully",
		"status":  input.Status,
	})
}

// parseUintPath is a small helper to extract & validate a uint path parameter.
func parseUintPath(ctx *gin.Context, name string) (uint, bool) {
	raw := ctx.Param(name)
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 || v > math.MaxUint32 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return uint(v), true
}
