package handlers

import (
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type CompanyHandlers struct {
	DB *gorm.DB
}

func NewCompanyHandlers(db *gorm.DB) *CompanyHandlers {
	return &CompanyHandlers{
		DB: db,
	}
}

// Handler function for editing company profile
// 
// Taking user input and updating the company profile in the database.
// Support partial updates. If any field is not provided, it will not be updated.
// 
// Support request with multipart/form-data
func (h *CompanyHandlers) EditProfileHandler(ctx *gin.Context) {
	// take user ID from context (auth middleware)
	userId := ctx.MustGet("userID").(string)
	// Expected form of data from request (ctx), no data is required, partial data is allowed.
	type CompanyEditProfileInput struct {
		Phone   *string               `form:"phone" binding:"omitempty,max=20"`
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
	if input.Photo != nil {
		photo, err := SaveFile(ctx, h.DB, userId, input.Photo, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		company.PhotoID = photo.ID
	}
	
	success := false

	defer (func() {
		if !success && input.Photo != nil {
			_ = h.DB.Delete(&model.File{
				ID: company.PhotoID,
			})
		}
	})()
	if input.Banner != nil {
		banner, err := SaveFile(ctx, h.DB, userId, input.Banner, model.FileCategoryImage)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		company.BannerID = banner.ID
	}
	
	// Remove file that are recently saved if request is not successful
	defer (func() {
		if !success && input.Banner != nil {
			_ = h.DB.Delete(&model.File{
				ID: company.BannerID,
			})
		}
	})()
	
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

// Get Company Profile
// 
// Use userID to get company profile.
// return company profile according to model.Company.
// 
// if no id is provided, use userID from context.
func (h *CompanyHandlers) GetProfileHandler(ctx *gin.Context) {
	type CompanyGetProfileInput struct {
		ID string `form:"id" binding:"max=64"`
	}
	
	input := CompanyGetProfileInput{}
	err := ctx.MustBindWith(&input, binding.Form)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Use current user ID if no ID is provided.
	if input.ID == "" {
		input.ID = ctx.MustGet("userID").(string)
	}
	
	// Try to get company info with company name included.
	type CompanyInfo struct {
		model.Company
		Name string `json:"name"`
	}
	var company CompanyInfo
	if err := h.DB.Model(&model.Company{}).Select("companies.*, users.username as name").Joins("INNER JOIN users on users.id = companies.user_id").Where("companies.user_id = ?", input.ID).Take(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	ctx.JSON(http.StatusOK, company)
}
