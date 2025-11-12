package handlers

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

// AdminHandlers handles admin-only HTTP endpoints and depends on the service layer.
type AdminHandlers struct {
	adminService services.AdminService
}

// NewAdminHandlers constructs AdminHandlers by wiring default repository and service
func NewAdminHandlers(svc services.AdminService) *AdminHandlers {
	return &AdminHandlers{adminService: svc}
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

	audits, err := h.adminService.FetchAuditLog(input.Offset, input.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the same payload shape as before.
	ctx.JSON(http.StatusOK, audits)
}

// @Summary Get log of email (Admin only)
// @Description Retrieves a list logs related to email i.e, sending success failure, to whom.
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.MailLog "List of all audit log entries"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /admin/emaillog [get]
func (h *AdminHandlers) FetchEmailLog(ctx *gin.Context) {
	type FetchEmailLogInput struct {
		Offset uint `json:"offset" form:"offset"`
		Limit  uint `json:"limit" form:"limit" binding:"max=64"`
	}

	input := FetchEmailLogInput{
		Limit: 32,
	}

	if err := ctx.Bind(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	audits, err := h.adminService.FetchMailLog(input.Offset, input.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the same payload shape as before.
	ctx.JSON(http.StatusOK, audits)
}

// @Summary Get log of email (Admin only)
// @Description Retrieves a list logs related to email i.e, sending success failure, to whom.
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.MailLog "List of all audit log entries"
// @Failure 403 {object} object{error=string} "Forbidden"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /admin/emaillog [get]
func (h *AdminHandlers) FetchEmailLog(ctx *gin.Context) {
	type FetchEmailLogInput struct {
		Offset uint `json:"offset" form:"offset"`
		Limit  uint `json:"limit" form:"limit" binding:"max=64"`
	}
	input := FetchEmailLogInput{
		Limit: 32,
	}
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var emailLogs []model.MailLog
	result := h.DB.Model(&model.MailLog{}).
		Offset(int(input.Offset)).
		Limit(int(input.Limit)).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Find(&emailLogs)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, emailLogs)
}
