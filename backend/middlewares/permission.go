package middlewares

import (
	"ku-work/backend/model"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminPermissionMiddleware checks if the user has admin permissions.
// Required to be use with AuthMiddleware
func AdminPermissionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exist := ctx.Get("userID")
		if !exist {
			slog.Error("User ID not found in context in AdminPermissionMiddleware")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		var count int64
		db.Model(&model.Admin{}).Where("user_id = ?", userID).Count(&count)

		if count == 0 {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the necessary permissions to perform this action."})
			return
		}

		ctx.Next()
	}
}
