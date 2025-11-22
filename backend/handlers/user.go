package handlers

import (
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserHandlers struct for handling user-related operations
type UserHandlers struct {
	DB          *gorm.DB
	gracePeriod int
}

func NewUserHandlers(db *gorm.DB, gracePeriod int) *UserHandlers {
	return &UserHandlers{
		DB:          db,
		gracePeriod: gracePeriod,
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
	// Find out user role
	role := helper.GetRole(userId, h.DB)
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

	// Update username
	if input.Username != nil {
		var usernameCount int64
		if err := h.DB.Model(&model.User{}).Where(&model.User{
			Username: *input.Username,
		}).Count(&usernameCount).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if usernameCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "username already exist"})
			return
		}
		user := model.User{
			ID: userId,
		}
		if err := h.DB.Take(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user.Username = *input.Username
		if err := h.DB.Save(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Get current company data
	company := model.Company{
		UserID: userId,
	}
	if err := h.DB.First(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update company data if input is provided in request
	if input.Phone != nil {
		company.Phone = *input.Phone
	}
	if input.Address != nil {
		company.Address = *input.Address
	}
	if input.City != nil {
		company.City = *input.City
	}
	if input.Country != nil {
		company.Country = *input.Country
	}
	if input.AboutUs != nil {
		company.AboutUs = *input.AboutUs
	}
	if input.Email != nil {
		if _, err := mail.ParseAddress(*input.Email); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
			return
		}
		company.Email = *input.Email
	}
	if input.Website != nil {
		if _, err := url.Parse(*input.Website); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid website URL"})
			return
		}
		company.Website = *input.Website
	}

	// Handle Photo and Banner Update
	// Remove file that are recently saved if request is not successful
	success := false
	defer (func() {
		if !success && input.Photo != nil {
			_ = h.DB.Delete(&model.File{
				ID: company.PhotoID,
			})
		}
	})()

	defer (func() {
		if !success && input.Banner != nil {
			_ = h.DB.Delete(&model.File{
				ID: company.BannerID,
			})
		}
	})()

	if input.Photo != nil {
		photo, err := SaveFile(ctx, h.DB, userId, input.Photo, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		company.PhotoID = photo.ID
	}

	if input.Banner != nil {
		banner, err := SaveFile(ctx, h.DB, userId, input.Banner, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		company.BannerID = banner.ID
	}

	// Save updated company to database.
	if err := h.DB.Save(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	success = true

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

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

	// Get current data from database.
	student := model.Student{
		UserID: userId,
	}
	if err := h.DB.First(&student).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update student data according to input, maintain same data if not provided
	if input.BirthDate != nil {
		parsedBirthDate, err := time.Parse(time.RFC3339, *input.BirthDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		student.BirthDate = datatypes.Date(parsedBirthDate)
	}

	if input.Phone != nil {
		student.Phone = *input.Phone
	}
	if input.AboutMe != nil {
		student.AboutMe = *input.AboutMe
	}
	if input.GitHub != nil {
		student.GitHub = *input.GitHub
	}
	if input.LinkedIn != nil {
		student.LinkedIn = *input.LinkedIn
	}
	if input.StudentStatus != "" {
		student.StudentStatus = input.StudentStatus
	}
	if input.Photo != nil {
		photo, err := SaveFile(ctx, h.DB, userId, input.Photo, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		student.PhotoID = photo.ID
	}

	// Save data into database
	result := h.DB.Save(&student)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
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
	role := helper.GetRole(userId, h.DB)
	username := helper.GetUsername(userId, role, h.DB)

	isRegistered := true 
	if role == helper.Viewer {
		var count int64
		h.DB.Model(&model.Student{}).Where("user_id = ?", userId).Count(&count)
		if count == 0 {
			isRegistered = false
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"username": username,
		"role":     role,
		"userId":   userId,
		"isRegistered": isRegistered,
	})
}
