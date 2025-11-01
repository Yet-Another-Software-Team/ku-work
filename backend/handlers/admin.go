package handlers

import (
	"net/http"

	gormrepo "ku-work/backend/repository/gorm"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminHandlers handles admin-only HTTP endpoints and depends on the service layer.
type AdminHandlers struct {
	svc services.AdminService
}

// NewAdminHandlers constructs AdminHandlers by wiring default repository and service
func NewAdminHandlers(db *gorm.DB) *AdminHandlers {
	repo := gormrepo.NewGormAuditRepository(db)
	svc := services.NewAdminService(repo)
	return &AdminHandlers{svc: svc}
}

// NewAdminHandlersWithService constructs handlers with an explicit service implementation.
// Use this for unit tests or when you already have a service instance.
func NewAdminHandlersWithService(svc services.AdminService) *AdminHandlers {
	return &AdminHandlers{svc: svc}
}

// @Summary Get an Audit Log (Admin only)
// @Description Retrieves a list of all audit log entries. This endpoint is restricted to admin users.
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.Audit "List of all audit log entries"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /admin/audits [get]
func (h *AdminHandlers) FetchAuditLog(ctx *gin.Context) {
	type FetchAuditLogInput struct {
		Offset uint `json:"offset" form:"offset"`
		Limit  uint `json:"limit" form:"limit" binding:"max=64"`
	}

	input := FetchAuditLogInput{
		Limit: 32,
	}

	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	audits, err := h.svc.FetchAuditLog(input.Offset, input.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the same payload shape as before.
	ctx.JSON(http.StatusOK, audits)
}