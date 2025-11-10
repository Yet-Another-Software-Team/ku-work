package handlers

import (
	"ku-work/backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Deactivate account
// @Description Soft deletes the user account. The account can be reactivated within the grace period (default 30 days, configurable via ACCOUNT_DELETION_GRACE_PERIOD_DAYS env variable). After the grace period expires, all personal data is automatically anonymized (not deleted) to comply with Thailand's PDPA while retaining data for analytics and compliance.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string,grace_period_days=int,deletion_date=string} "Account deactivated successfully"
// @Failure 400 {object} object{error=string} "Account already deactivated"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /me/deactivate [post]
func (h *UserHandlers) DeactivateAccount(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	deletionDate, err := h.UserService.DeactivateAccount(ctx.Request.Context(), userID, h.gracePeriod)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case services.ErrAlreadyDeactivated:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is already deactivated"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate account"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":           "Account deactivated successfully",
		"grace_period_days": h.gracePeriod,
		"deletion_date":     deletionDate.Format(time.RFC3339),
	})
}

// @Summary Reactivate account
// @Description Reactivates a deactivated account if within the grace period. Once the grace period expires and data is anonymized, reactivation is not possible. Requires valid authentication token.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string} "Account reactivated successfully"
// @Failure 400 {object} object{error=string} "Account is not deactivated"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 403 {object} object{error=string} "Grace period expired, account already anonymized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /me/reactivate [post]
func (h *UserHandlers) ReactivateAccount(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	err := h.UserService.ReactivateAccount(ctx.Request.Context(), userID, h.gracePeriod)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case services.ErrNotDeactivated:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is not deactivated"})
		default:
			if gracePeriodErr, ok := err.(*services.GracePeriodExpiredError); ok {
				ctx.JSON(http.StatusForbidden, gin.H{
					"error":        "Grace period has expired. Account cannot be reactivated.",
					"deleted_at":   gracePeriodErr.DeletedAt.Format(time.RFC3339),
					"deadline_was": gracePeriodErr.Deadline.Format(time.RFC3339),
				})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate account"})
			}
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Account reactivated successfully",
	})
}
