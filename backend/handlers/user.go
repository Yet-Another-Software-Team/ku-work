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
	DB *gorm.DB
}

func NewUserHandlers(db *gorm.DB) *UserHandlers {
	return &UserHandlers{
		DB: db,
	}
}

// Handler function for editing user profile
//
// Taking user input and updating the company profile in the database.
// Support partial updates. If any field is not provided, it will not be updated.
//
// Support request with multipart/form-data
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
	if role == helper.Company {
		h.editCompanyProfile(ctx, userId)
	} else if role == helper.Student {
		h.editStudentProfile(ctx, userId)
	} else {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only Company or Student can edit profile"})
	}
}

func (h *UserHandlers) editCompanyProfile(ctx *gin.Context, userId string) {
	// Expected form of data from request (ctx), no data is required, partial data is allowed.
	type CompanyEditProfileInput struct {
		Phone   *string               `form:"phone" binding:"omitempty,max=20"`
		Email   *string               `form:"email" binding:"omitempty,email,max=256"`
		Website *string               `form:"website" binding:"omitempty,url,max=256"`
		Address *string               `form:"address" binding:"omitempty,max=512"`
		City    *string               `form:"city" binding:"omitempty,max=128"`
		Country *string               `form:"country" binding:"omitempty,max=128"`
		Photo   *multipart.FileHeader `form:"photo" binding:"omitempty"`
		Banner  *multipart.FileHeader `form:"banner" binding:"omitempty"`
	}
	input := CompanyEditProfileInput{}

	// Validate input data
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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


// Simple Handler to get profile of current user
func (h *UserHandlers) GetProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)
	role := helper.GetRole(userId, h.DB)
	username := helper.GetUsername(userId, role, h.DB)
	ctx.JSON(http.StatusOK, gin.H{
		"username": username,
		"role": role,
	})
}