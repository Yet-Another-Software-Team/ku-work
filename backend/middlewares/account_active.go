package middlewares

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

// AccountActiveMiddleware blocks requests if the authenticated user's account is deactivated.
// Uses AccountStatusService following layered architecture principles.
func AccountActiveMiddleware(accountStatusService *services.AccountStatusService) gin.HandlerFunc {
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

		// Use service layer instead of direct database access
		isDeactivated, err := accountStatusService.IsAccountDeactivated(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to verify account status"})
			return
		}

		if isDeactivated {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account deactivated"})
			return
		}

		c.Next()
	}
}
