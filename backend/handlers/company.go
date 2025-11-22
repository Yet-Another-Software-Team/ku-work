package handlers

import (
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log/slog"
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

// CompanyResponse defines the API response for company data, aligned with model.Company.
type CompanyResponse struct {
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

// anonymizeCompany zeros or replaces personally-identifying fields for deactivated accounts.
func anonymizeCompany(c *CompanyResponse) {
	c.Email = ""
	c.Phone = ""
	c.PhotoID = ""
	c.BannerID = ""
	c.Address = ""
	c.City = ""
	c.Country = ""
	c.Website = ""
	c.AboutUs = ""
	c.Name = "Deactivated Account"
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
	if err := h.DB.Model(&model.Company{}).
		Select("companies.*, users.username as name").
		Joins("INNER JOIN users on users.id = companies.user_id").
		Where("companies.user_id = ?", id).
		Take(&company).Error; err != nil {
		slog.Error("Failed to get company profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get company profile"})
		return
	}

	// Map to response type to avoid leaking GORM internals and to make anonymization straightforward.
	resp := CompanyResponse{
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
		UserID:    company.UserID,
		Email:     company.Email,
		Phone:     company.Phone,
		PhotoID:   company.PhotoID,
		BannerID:  company.BannerID,
		Address:   company.Address,
		City:      company.City,
		Country:   company.Country,
		Website:   company.Website,
		AboutUs:   company.AboutUs,
		Name:      company.Name,
	}

	// If the user is deactivated, anonymize sensitive fields.
	if helper.IsDeactivated(h.DB, id) {
		anonymizeCompany(&resp)
	}

	ctx.JSON(http.StatusOK, resp)
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

	var rawResults []struct {
		CreatedAt time.Time
		UpdatedAt time.Time
		UserID    string
		Email     string
		Phone     string
		PhotoID   string
		BannerID  string
		Address   string
		City      string
		Country   string
		Website   string
		AboutUs   string
		Name      string
	}

	if err := h.DB.Model(&model.Company{}).
		Select("companies.created_at, companies.updated_at, companies.user_id, companies.email, companies.phone, companies.photo_id, companies.banner_id, companies.address, companies.city, companies.country, companies.website, companies.about_us, users.username as name").
		Joins("INNER JOIN users on users.id = companies.user_id").
		Find(&rawResults).Error; err != nil {
		slog.Error("Failed to get company list", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get company list"})
		return
	}

	companies := make([]CompanyResponse, 0, len(rawResults))
	for _, r := range rawResults {
		company := CompanyResponse{
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			UserID:    r.UserID,
			Email:     r.Email,
			Phone:     r.Phone,
			PhotoID:   r.PhotoID,
			BannerID:  r.BannerID,
			Address:   r.Address,
			City:      r.City,
			Country:   r.Country,
			Website:   r.Website,
			AboutUs:   r.AboutUs,
			Name:      r.Name,
		}

		// If the account is deactivated, anonymize the entry.
		if helper.IsDeactivated(h.DB, r.UserID) {
			anonymizeCompany(&company)
		}

		companies = append(companies, company)
	}

	ctx.JSON(http.StatusOK, companies)
}
