package handlers

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	"ku-work/backend/services"
)

type StudentHandler struct {
	StudentSvc *services.StudentService
	AccountSvc *services.AccountService
}

func NewStudentHandler(studentSvc *services.StudentService, accountSvc *services.AccountService) *StudentHandler {
	return &StudentHandler{
		StudentSvc: studentSvc,
		AccountSvc: accountSvc,
	}
}

// StudentInfo type removed after refactor to service/repository layered architecture.

// anonymizeStudent removed after refactor; anonymization is handled in the StudentService layer.

// @Summary Register as a student
// @Description Handles the registration process for a user who has already authenticated (e.g., via Google OAuth) to become a student. The registration is submitted for admin approval. This endpoint is protected and requires authentication.
// @Tags Students
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param phone formData string false "Phone number"
// @Param birthDate formData string false "Birth date in RFC3339 format (e.g., 2006-01-02T15:04:05Z)"
// @Param aboutMe formData string false "A short bio or about me section"
// @Param github formData string false "GitHub profile URL"
// @Param linkedIn formData string false "LinkedIn profile URL"
// @Param studentId formData string true "10-digit student ID number"
// @Param major formData string true "Major of study" Enums(Software and Knowledge Engineering, Computer Engineering)
// @Param studentStatus formData string true "Current student status" Enums(Graduated, Current Student)
// @Param photo formData file true "Profile photo"
// @Param statusPhoto formData file true "Document proving student status (e.g., student ID card photo)"
// @Success 200 {object} object{message=string} "ok"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 409 {object} object{error=string} "Conflict: User already registered"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/student/register [post]
func (h *StudentHandler) RegisterHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind request body to input struct
	type StudentRegistrationInput struct {
		Phone             string                `form:"phone" binding:"max=20"`
		BirthDate         string                `form:"birthDate" binding:"required,max=27"`
		AboutMe           string                `form:"aboutMe" binding:"max=16384"`
		GitHub            string                `form:"github" binding:"max=256"`
		LinkedIn          string                `form:"linkedIn" binding:"max=256"`
		StudentID         string                `form:"studentId" binding:"required,len=10"`
		Major             string                `form:"major" binding:"required,oneof='Software and Knowledge Engineering' 'Computer Engineering'"`
		StudentStatus     string                `form:"studentStatus" binding:"required,oneof='Graduated' 'Current Student'"`
		Photo             *multipart.FileHeader `form:"photo" binding:"required"`
		StudentStatusFile *multipart.FileHeader `form:"statusPhoto" binding:"required"`
	}
	input := StudentRegistrationInput{}
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delegate to service which handles validation, file saving, and persistence
	svcInput := services.StudentRegistrationInput{
		Phone:             input.Phone,
		BirthDate:         input.BirthDate,
		AboutMe:           input.AboutMe,
		GitHub:            input.GitHub,
		LinkedIn:          input.LinkedIn,
		StudentID:         input.StudentID,
		Major:             input.Major,
		StudentStatus:     input.StudentStatus,
		Photo:             input.Photo,
		StudentStatusFile: input.StudentStatusFile,
	}
	if err := h.StudentSvc.RegisterStudent(ctx, userId, svcInput); err != nil {
		if err == services.ErrAlreadyRegistered {
			ctx.JSON(http.StatusConflict, gin.H{"error": "user already registered to be student"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// @Summary Edit student profile
// @Description Allows an authenticated student to edit their profile information. Supports partial updates for most fields.
// @Tags Students
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param phone formData string false "New phone number"
// @Param birthDate formData string false "New birth date in RFC3339 format"
// @Param aboutMe formData string false "Updated about me section"
// @Param github formData string false "New GitHub profile URL"
// @Param linkedIn formData string false "New LinkedIn profile URL"
// @Param studentStatus formData string true "Updated student status" Enums(Graduated, Current Student)
// @Param photo formData file false "New profile photo"
// @Success 200 {object} object{message=string} "ok"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /me [patch]
func (h *StudentHandler) EditProfileHandler(ctx *gin.Context) {
	// Get user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data to StudentEditProfileInput struct
	type StudentEditProfileInput struct {
		Phone         *string               `form:"phone" binding:"omitempty,max=20"`
		BirthDate     *string               `form:"birthDate" binding:"omitempty,max=27"`
		AboutMe       *string               `form:"aboutMe" binding:"omitempty,max=16384"`
		GitHub        *string               `form:"github" binding:"omitempty,max=256"`
		LinkedIn      *string               `form:"linkedIn" binding:"omitempty,max=256"`
		StudentStatus string                `form:"studentStatus" binding:"required,oneof='Graduated' 'Current Student'"`
		Photo         *multipart.FileHeader `form:"photo"`
	}
	input := StudentEditProfileInput{}
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map handler input to service input and delegate to AccountService
	payload := services.StudentEditProfileInput{
		Phone:         input.Phone,
		BirthDate:     input.BirthDate,
		AboutMe:       input.AboutMe,
		GitHub:        input.GitHub,
		LinkedIn:      input.LinkedIn,
		StudentStatus: input.StudentStatus,
		Photo:         input.Photo,
	}
	if h.AccountSvc == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "account service not configured"})
		return
	}
	if err := h.AccountSvc.UpdateStudentProfile(ctx, userId, payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update student profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// @Summary Approve or reject a student registration (Admin only)
// @Description Allows an admin to approve or reject a student's registration application based on their user ID.
// @Tags Students
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID of the student to be approved/rejected"
// @Param approval body handlers.StudentHandler.ApproveHandler.StudentRegistrationApprovalInput true "Approval action"
// @Success 200 {object} object{message=string} "ok"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 404 {object} object{error=string} "Not Found: Student not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /students/{id}/approval [post]
func (h *StudentHandler) ApproveHandler(ctx *gin.Context) {
	// Bind input data to struct
	type StudentRegistrationApprovalInput struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason" binding:"max=16384"`
	}
	input := StudentRegistrationApprovalInput{}
	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID := ctx.Param("id")
	actorID := ctx.MustGet("userID").(string)

	// Ensure student exists
	if _, err := h.StudentSvc.GetStudentProfile(ctx.Request.Context(), studentID); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Delegate approve/reject to service
	if err := h.StudentSvc.ApproveOrRejectStudent(ctx.Request.Context(), studentID, input.Approve, actorID, input.Reason); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// @Summary Get student profile(s)
// @Description Fetches student profile information. An admin can retrieve a paginated list of all students and filter by approval status. A regular user will get their own detailed profile. An admin can also specify a user ID to get a specific profile.
// @Tags Students
// @Security BearerAuth
// @Produce json
// @Param id query string false "User ID of a specific student (for admins)"
// @Param offset query int false "Pagination offset (for admin list)"
// @Param limit query int false "Pagination limit (for admin list)" default(64)
// @Param approvalStatus query string false "Filter by approval status (for admin list)" Enums(pending, accepted, rejected)
// @Success 200 {object} object{profile=repo.StudentProfile} "Returns a single student's detailed profile"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /students [get]
func (h *StudentHandler) GetProfileHandler(ctx *gin.Context) {
	// Get userId from context (auth middleware)
	userId := ctx.MustGet("userID").(string)

	// Bind input data from request body
	type GetStudentProfileInput struct {
		UserID         string `form:"id" binding:"max=128"`
		Offset         int    `json:"offset" form:"offset"`
		Limit          int    `json:"limit" form:"limit" binding:"max=64"`
		ApprovalStatus string `json:"approvalStatus" form:"approvalStatus" binding:"max=64"`
		SortBy         string `json:"sortBy" form:"sortBy" binding:"omitempty,oneof='latest' 'oldest' 'name_az' 'name_za'"`
	}
	input := GetStudentProfileInput{
		Limit: 64,
	}
	err := ctx.MustBindWith(&input, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If user ID is provided explicitly, return that profile (admin or owner)
	if input.UserID != "" {
		profile, err := h.StudentSvc.GetStudentProfile(ctx.Request.Context(), input.UserID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"profile": profile,
		})
		return
	}

	// If requester is admin, return list with filters
	role := h.AccountSvc.GetRole(ctx.Request.Context(), userId)
	if role == helper.Admin {
		var statusPtr *model.StudentApprovalStatus
		if input.ApprovalStatus != "" {
			s := model.StudentApprovalStatus(input.ApprovalStatus)
			statusPtr = &s
		}
		items, err := h.StudentSvc.ListStudentProfiles(ctx.Request.Context(), repo.StudentListFilter{
			Offset:         input.Offset,
			Limit:          input.Limit,
			ApprovalStatus: statusPtr,
			SortBy:         input.SortBy,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, items)
		return
	}

	// Default: return the caller's own profile
	profile, err := h.StudentSvc.GetStudentProfile(ctx.Request.Context(), userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}
