package middlewares

import (
	"log"
	"net/http"
	"strings"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddlewareWithDB creates an authentication middleware with database access for blacklist checking
// This is the OWASP-compliant version that checks for revoked tokens
func AuthMiddlewareWithDB(jwtSecret []byte, db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		// Check if the token format is "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in the format 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return jwtSecret, nil
		})

		// Check for errors in token parsing or validation
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(*model.UserClaims); ok && token.Valid {
			// OWASP Requirement: Check if JWT is blacklisted (revoked after logout)
			// This prevents use of tokens after session termination
			var revokedJWT model.RevokedJWT
			err := db.Where("jti = ?", claims.ID).First(&revokedJWT).Error
			if err == nil {
				// Token found in blacklist - it was revoked
				log.Printf("SECURITY: Blocked revoked JWT usage. User: %s, JTI: %s, IP: %s",
					claims.UserID, claims.ID, ctx.ClientIP())
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				return
			}
			// err == gorm.ErrRecordNotFound means token is not blacklisted, which is good

			ctx.Set("userID", claims.UserID)
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}
}

// AuthMiddleware is kept for backward compatibility but should be replaced with AuthMiddlewareWithDB
// Deprecated: Use AuthMiddlewareWithDB for OWASP-compliant blacklist checking
func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the token from the Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		// Check if the token format is "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in the format 'Bearer <token>'"})
			return
		}

		tokenString := parts[1]

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNotSupported
			}
			return jwtSecret, nil
		})

		// Check for errors in token parsing or validation
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(*model.UserClaims); ok && token.Valid {
			ctx.Set("userID", claims.UserID)
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}
}
