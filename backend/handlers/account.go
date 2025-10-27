package handlers

import (
	"ku-work/backend/model"
	"ku-work/backend/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// Find User
	user := model.User{}
	if err := h.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	// Check if already deactivated
	if user.DeletedAt.Valid {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is already deactivated"})
		return
	}

	// Check if user is a company and disable their job posts
	var company model.Company
	if err := h.DB.Where("user_id = ?", userID).First(&company).Error; err == nil {
		// User is a company - disable all their job posts
		if err := services.DisableCompanyJobPosts(h.DB, userID); err != nil {
			log.Printf("Warning: Failed to disable job posts for company %s: %v", userID, err)
			// Don't fail the deactivation, just log the warning
		}
	}

	// Soft Delete User (this will trigger BeforeDelete hooks)
	if err := h.DB.Delete(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate account"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":           "Account deactivated successfully",
		"grace_period_days": h.gracePeriod,
		"deletion_date":     time.Now().Add(time.Duration(h.gracePeriod) * 24 * time.Hour).Format(time.RFC3339),
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

	// Find User including soft deleted records
	user := model.User{}
	if err := h.DB.Unscoped().Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	// Check if account is deactivated
	if !user.DeletedAt.Valid {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account is not deactivated"})
		return
	}

	// Check if within grace period
	gracePeriodDuration := time.Duration(h.gracePeriod) * 24 * time.Hour
	deletionDeadline := user.DeletedAt.Time.Add(gracePeriodDuration)

	if time.Now().After(deletionDeadline) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"error":        "Grace period has expired. Account cannot be reactivated.",
			"deleted_at":   user.DeletedAt.Time.Format(time.RFC3339),
			"deadline_was": deletionDeadline.Format(time.RFC3339),
		})
		return
	}

	// Restore the user account
	if err := h.DB.Model(&user).Unscoped().Update("deleted_at", nil).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reactivate account"})
		return
	}

	// Also restore related student or company records
	// Check if student exists
	var student model.Student
	if err := h.DB.Unscoped().Where("user_id = ?", userID).First(&student).Error; err == nil {
		if student.DeletedAt.Valid {
			h.DB.Model(&student).Unscoped().Update("deleted_at", nil)
		}
	}

	// Check if company exists and re-enable their job posts
	var company model.Company
	if err := h.DB.Unscoped().Where("user_id = ?", userID).First(&company).Error; err == nil {
		if company.DeletedAt.Valid {
			h.DB.Model(&company).Unscoped().Update("deleted_at", nil)
		}
		// Re-enable job posts for reactivated company
		// Note: We set them back to open, but companies may want to review them
		result := h.DB.Model(&model.Job{}).
			Where("company_id = ? AND is_open = ?", userID, false).
			Update("is_open", true)
		if result.Error != nil {
			log.Printf("Warning: Failed to re-enable job posts for company %s: %v", userID, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("Re-enabled %d job posts for company: %s", result.RowsAffected, userID)
		}
	}

	// Check if Google OAuth details exist
	var googleOAuth model.GoogleOAuthDetails
	if err := h.DB.Unscoped().Where("user_id = ?", userID).First(&googleOAuth).Error; err == nil {
		if googleOAuth.DeletedAt.Valid {
			h.DB.Model(&googleOAuth).Unscoped().Update("deleted_at", nil)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Account reactivated successfully",
	})
}

// @Summary Request account deletion
// @Description Marks the account for anonymization. Account will be deactivated immediately and all personal data will be anonymized after the grace period (PDPA compliant). Anonymized data is retained for analytics and compliance purposes but cannot be linked back to the individual.
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string,grace_period_days=int,deletion_date=string} "Account marked for anonymization"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /me/delete [delete]
func (h *UserHandlers) RequestAccountDeletion(ctx *gin.Context) {
	// This is essentially the same as deactivation
	// The scheduler will handle anonymization (not deletion) after grace period
	// This complies with Thailand's PDPA - personal data is anonymized while
	// retaining records for legitimate business purposes (analytics, compliance)
	h.DeactivateAccount(ctx)
}
