package handlers

import (
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"net/http"
	"time"

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

// CompanyResponse defines the API response for company data, excluding GORM-specific types like DeletedAt.
type CompanyResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   string    `json:"photoId"`
	BannerID  string    `json:"bannerId"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Website   string    `json:"website"`
	AboutUs   string    `json:"about"`
	Name      string    `json:"name"`
}

// @Summary Get a company's profile
// @Description Retrieves the profile of a specific company using their user ID.
// @Tags Companies
// @Security BearerAuth
// @Produce json
// @Param id path string true "Company User ID"
// @Success 200 {object} handlers.CompanyResponse "Company profile retrieved successfully"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /company/{id} [get]
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

// @Summary Get a list of all companies (Admin only)
// @Description Retrieves a list of all registered companies. This endpoint is restricted to admin users.
// @Tags Companies
// @Security BearerAuth
// @Produce json
// @Success 200 {array} handlers.CompanyResponse "List of all companies"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /company [get]
func (h *CompanyHandlers) GetCompanyListHandler(ctx *gin.Context) {
	userId := ctx.MustGet("userID").(string)
	role := helper.GetRole(userId, h.DB)
	if role != helper.Admin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}
	var companies []CompanyResponse
	if err := h.DB.Model(&model.Company{}).
		Select("companies.id, companies.created_at, companies.updated_at, companies.user_id, companies.email, companies.phone, companies.photo_id, companies.banner_id, companies.address, companies.city, companies.country, companies.website, companies.about_us, users.username as name").
		Joins("INNER JOIN users on users.id = companies.user_id").
		Find(&companies).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, companies)
}
