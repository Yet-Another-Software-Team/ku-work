package middlewares

import (
	"ku-work/backend/model"
	"log"
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
			log.Panic("User ID not found in context")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		
		var admin model.Admin
		// Get higest role possible
		if err := db.First(&admin, "user_id = ?", userID).Error; err != nil {
		   if err == gorm.ErrRecordNotFound {
		        ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the necessary permissions to perform this action."})
		    } else {
		        log.Printf("Database error: %v", err)
		        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred."})
		    }
		    return
		}
		
		ctx.Next()
	}
}
