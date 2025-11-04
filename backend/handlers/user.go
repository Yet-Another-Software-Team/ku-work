package handlers

import (
	"mime/multipart"
	"net/http"
	"time"

	"ku-work/backend/helper"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// UserHandlers struct for handling user-related operations
type UserHandlers struct {
	gracePeriod int
	UserService *services.AccountService
}

func NewUserHandlers(svc *services.AccountService) *UserHandlers {
	return &UserHandlers{
		gracePeriod: helper.GetGracePeriodDays(),
		UserService: svc,
	}
}

// @Summary Edit user profile
// @Description Edits the profile of the currently authenticated user. This endpoint automatically detects whether the user is a student or a company and accepts the appropriate fields for that role. Supports partial updates.
// @Tags Users
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param phone formData string false "Phone number (For both Student and Company)"
// @Param birthDate formData string false "Birth date in RFC3339 format (Student only)"
// @Param aboutMe formData string false "About me/us section (For both Student and Company)"
// @Param github formData string false "GitHub profile URL (Student only)"
// @Param linkedIn formData string false "LinkedIn profile URL (Student only)"
// @Param studentStatus formData string false "Student status (Student only)" Enums(Graduated, Current Student)
// @Param email formData string false "Company email (Company only)"
// @Param website formData string false "Company website URL (Company only)"
// @Param address formData string false "Company address (Company only)"
// @Param city formData string false "Company city (Company only)"
// @Param country formData string false "Company country (Company only)"
// @Param photo formData file false "Profile photo/logo (For both Student and Company)"
// @Param banner formData file false "Company banner image (Company only)"
// @Success 200 {object} object{message=string} "ok"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /me [patch]
func (h *UserHandlers) EditProfileHandler(ctx *gin.Context) {
	// take user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)
	// Find out user role via injected service
	role := h.UserService.GetRole(ctx.Request.Context(), userId)
	if role == helper.Unknown {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid User"})
		return
	}
	// Perform role-specific actions
	switch role {
	case helper.Company:
		h.editCompanyProfile(ctx, userId)
	case helper.Student:
		h.editStudentProfile(ctx, userId)
	default:
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only Company or Student can edit profile"})
	}
}

// Strategy function to edit company profile
func (h *UserHandlers) editCompanyProfile(ctx *gin.Context, userId string) {
	// Expected form of data from request (ctx), no data is required, partial data is allowed.
	type CompanyEditProfileInput struct {
		Phone    *string               `form:"phone" binding:"omitempty,max=20"`
		Email    *string               `form:"email" binding:"omitempty,email,max=256"`
		Website  *string               `form:"website" binding:"omitempty,url,max=256"`
		Address  *string               `form:"address" binding:"omitempty,max=512"`
		City     *string               `form:"city" binding:"omitempty,max=128"`
		Country  *string               `form:"country" binding:"omitempty,max=128"`
		AboutUs  *string               `form:"about" binding:"omitempty,max=16384"`
		Username *string               `form:"username" binding:"omitempty,max=256"`
		Photo    *multipart.FileHeader `form:"photo" binding:"omitempty"`
		Banner   *multipart.FileHeader `form:"banner" binding:"omitempty"`
	}
	input := CompanyEditProfileInput{}

	// Validate input data
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map handler input to service input
	payload := services.CompanyEditProfileInput{
		Phone:    input.Phone,
		Email:    input.Email,
		Website:  input.Website,
		Address:  input.Address,
		City:     input.City,
		Country:  input.Country,
		AboutUs:  input.AboutUs,
		Username: input.Username,
		Photo:    input.Photo,
		Banner:   input.Banner,
	}

	if err := h.UserService.UpdateCompanyProfile(ctx, userId, payload); err != nil {
		switch err {
		case services.ErrUsernameExists:
			ctx.JSON(http.StatusConflict, gin.H{"error": "username already exist"})
			return
		case services.ErrInvalidWebsite:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid website URL"})
			return
		default:
			if err.Error() == "invalid email address" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update company profile"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// Strategy function to edit student profile
func (h *UserHandlers) editStudentProfile(ctx *gin.Context, userId string) {
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

	// Map handler input to service input
	payload := services.StudentEditProfileInput{
		Phone:         input.Phone,
		BirthDate:     input.BirthDate,
		AboutMe:       input.AboutMe,
		GitHub:        input.GitHub,
		LinkedIn:      input.LinkedIn,
		StudentStatus: input.StudentStatus,
		Photo:         input.Photo,
	}

	if err := h.UserService.UpdateStudentProfile(ctx, userId, payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update student profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

// @Summary Get current user's profile
// @Description Retrieves basic profile information (username, role, and user ID) for the currently authenticated user.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{username=string, role=string, userId=string} "Successfully retrieved user profile"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Router /me [get]
func (h *UserHandlers) GetProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	role := h.UserService.GetRole(ctx.Request.Context(), userId)
	username := h.UserService.GetUsername(ctx.Request.Context(), userId, role)
	ctx.JSON(http.StatusOK, gin.H{
		"username": username,
		"role":     role,
		"userId":   userId,
	})
}

// @Summary Deactivate account
// @Description Soft deletes the user account. The account can be reactivated within the grace period (default 30 days, configurable via ACCOUNT_DELETION_GRACE_PERIOD_DAYS env variable). After the grace period expires, all personal data is automatically anonymized (not deleted) to comply with Thailand's PDPA while retaining data for analytics and compliance.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string,grace_period_days=int,deletion_date=string} "Account deactivated successfully"
// @Failure 400 {object} object{error=string} "Account already deactivated"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /me/deactivate [post]
func (h *UserHandlers) DeactivateAccount(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	// Use injected AccountService (DI)
	if h.UserService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "account service not configured"})
		return
	}
	accountService := h.UserService

	deletionDate, err := accountService.DeactivateAccount(ctx.Request.Context(), userID, h.gracePeriod)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		case services.ErrAlreadyDeactivated:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is already deactivated"})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate account"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":           "Account deactivated successfully",
		"grace_period_days": h.gracePeriod,
		"deletion_date":     deletionDate.Format(time.RFC3339),
	})
}

// @Summary Reactivate account
// @Description Reactivates a deactivated account if within the grace period. Once the grace period expires and data is anonymized, reactivation is not possible. Requires valid authentication token.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string} "Account reactivated successfully"
// @Failure 400 {object} object{error=string} "Account is not deactivated"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Grace period expired, account already anonymized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /me/reactivate [post]
func (h *UserHandlers) ReactivateAccount(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	// Use injected AccountService (DI)
	if h.UserService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "account service not configured"})
		return
	}
	accountService := h.UserService

	err := accountService.ReactivateAccount(ctx.Request.Context(), userID, h.gracePeriod)
	if err != nil {
		switch e := err.(type) {
		case *services.GracePeriodExpiredError:
			ctx.JSON(http.StatusForbidden, gin.H{
				"error":        "Grace period has expired. Account cannot be reactivated.",
				"deleted_at":   e.DeletedAt.Format(time.RFC3339),
				"deadline_was": e.Deadline.Format(time.RFC3339),
			})
			return
		default:
			if err == services.ErrUserNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			if err == services.ErrNotDeactivated {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is not deactivated"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate account"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Account reactivated successfully",
	})
}
