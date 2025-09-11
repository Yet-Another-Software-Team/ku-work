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
		Address *string               `form:"birthDate" binding:"omitempty,max=27"`
		City    *string               `form:"aboutMe" binding:"omitempty,max=16384"`
		Country *string               `form:"github" binding:"omitempty,max=256"`
		Photo   *multipart.FileHeader `form:"photo"`
		Banner  *multipart.FileHeader `form:"photo"`
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
	err := ctx.MustBindWith(&input, binding.FormMultipart)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company := model.Company{
		UserID: input.ID,
	}
	if err := h.DB.First(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"company": company,
	})
}
