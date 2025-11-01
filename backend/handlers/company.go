package handlers

import (
	"ku-work/backend/helper"
	"ku-work/backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CompanyHandlers struct {
	Service *services.CompanyService
}

func NewCompanyHandlers(db *gorm.DB) *CompanyHandlers {
	return &CompanyHandlers{
		Service: services.NewCompanyService(db),
	}
}

// anonymizeCompany zeros or replaces personally-identifying fields for deactivated accounts.
func anonymizeCompany(c *services.CompanyResponse) {
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
	company, err := h.Service.GetCompanyByUserID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// anonymize if deactivated
	if helper.IsDeactivated(h.Service.DB, id) {
		anonymizeCompany(&company)
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
	if !h.Service.IsAdmin(userId) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	companies, err := h.Service.ListCompanies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Anonymize deactivated accounts before returning
	for i := range companies {
		if helper.IsDeactivated(h.Service.DB, companies[i].UserID) {
			anonymizeCompany(&companies[i])
		}
	}

	ctx.JSON(http.StatusOK, companies)
}
