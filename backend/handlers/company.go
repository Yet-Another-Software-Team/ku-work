package handlers

import (
	"log"
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

type CompanyHandlers struct {
	Service *services.CompanyService
}

func NewCompanyHandlers(service *services.CompanyService) *CompanyHandlers {
	return &CompanyHandlers{
		Service: service,
	}
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
	company, err := h.Service.GetCompanyByUserID(ctx.Request.Context(), id)
	if err != nil {
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
	// Use context-aware admin check from service (delegates to repository)
	isAdmin, err := h.Service.IsAdminCtx(ctx.Request.Context(), userId)
	if err != nil {
		log.Printf("ERROR: failed to check admin permission for user %s: %v", userId, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	if !isAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	companies, err := h.Service.ListCompanies(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, companies)
}
