package handlers

import (
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
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

// Get Company Profile
//
// Use userID to get company profile.
// return company profile according to model.Company.
func (h *CompanyHandlers) GetCompanyProfileHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	// Try to get company info with company name included.
	type CompanyInfo struct {
		model.Company
		Name string `json:"name"`
	}
	var company CompanyInfo
	if err := h.DB.Model(&model.Company{}).Select("companies.*, users.username as name").Joins("INNER JOIN users on users.id = companies.user_id").Where("companies.user_id = ?", id).Take(&company).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, company)
}


// Admin Only to get collection of companies
func (h *CompanyHandlers) GetCompanyListHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userId").(string)
	role := helper.GetRole(userId, h.DB)
	if role != helper.Admin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	var companies []model.Company
	if err := h.DB.Model(&model.Company{}).Find(&companies).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

