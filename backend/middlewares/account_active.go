package middlewares

import (
	"net/http"

	"ku-work/backend/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AccountActiveMiddleware blocks requests if the authenticated user's account is deactivated.
func AccountActiveMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, ok := uidVal.(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// helper.IsDeactivated returns true when the user's DeletedAt is valid.
		if helper.IsDeactivated(db, userID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account deactivated"})
			return
		}

		c.Next()
	}
}
