package middlewares

import (
	"ku-work/backend/services"
	"log"
	"net/http"

	repo "ku-work/backend/repository"

	"github.com/gin-gonic/gin"
)

// AdminPermissionMiddleware checks if the user has admin permissions.
func AdminPermissionMiddleware(identityRepo repo.IdentityRepository) gin.HandlerFunc {
	permSvc := services.NewPermissionService(identityRepo)

	return func(ctx *gin.Context) {
		userIDVal, exist := ctx.Get("userID")
		if !exist {
			log.Println("User ID not found in context")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		userID, ok := userIDVal.(string)
		if !ok {
			log.Println("Invalid userID type in context")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		isAdmin, err := permSvc.IsAdmin(ctx.Request.Context(), userID)
		if err != nil {
			log.Printf("ERROR: failed to check admin permission for user %s: %v", userID, err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if !isAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have the necessary permissions to perform this action."})
			return
		}

		ctx.Next()
	}
}
