package handlers

import (
	"ku-work/backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminHandlers struct {
	DB *gorm.DB
}

func NewAdminHandlers(db *gorm.DB) *AdminHandlers {
	return &AdminHandlers{
		DB: db,
	}
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
	err := ctx.Bind(&input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var auditLogEntry []model.Audit
	result := h.DB.Model(&model.Audit{}).Offset(int(input.Offset)).Limit(int(input.Limit)).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).Find(&auditLogEntry)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	ctx.JSON(http.StatusOK, auditLogEntry)
}
