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

func (h *CompanyHandlers) EditProfileHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	type CompanyEditProfileInput struct {
		Phone   *string               `form:"phone" binding:"omitempty,max=20"`
		Address *string               `form:"address" binding:"omitempty,max=512"`
		City    *string               `form:"city" binding:"omitempty,max=128"`
		Country *string               `form:"country" binding:"omitempty,max=128"`
		Photo   *multipart.FileHeader `form:"photo" binding:"omitempty"`
		Banner  *multipart.FileHeader `form:"banner" binding:"omitempty"`
	}
	input := CompanyEditProfileInput{}
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company := model.Company{
		UserID: userId,
	}
	if err := h.DB.First(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	defer (func() {
		if !success && input.Banner != nil {
			_ = h.DB.Delete(&model.File{
				ID: company.BannerID,
			})
		}
	})()
	if err := h.DB.Save(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	success = true
	ctx.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

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
	if input.ID == "" {
		input.ID = ctx.MustGet("userID").(string)
	}
	company := model.Company{
		UserID: input.ID,
	}
	if err := h.DB.First(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, company)
}
