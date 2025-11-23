package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"ku-work/backend/model"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware creates an authentication middleware with Redis-based JWT revocation checking
// This is the OWASP-compliant version that checks for revoked tokens using Redis (faster than DB)
// IMPORTANT: Redis client MUST not be nil. This middleware will fail if Redis is unavailable.
func AuthMiddleware(jwtSecret []byte, redisClient *redis.Client) gin.HandlerFunc {
	if redisClient == nil {
		slog.Error("FATAL: Redis client is nil. JWT revocation requires Redis to be available.")
		os.Exit(1)
	}

	revocationService := services.NewJWTRevocationService(redisClient)

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
			// Using Redis for O(1) lookup performance
			isRevoked, err := revocationService.IsJWTRevoked(context.Background(), claims.ID)
			if err != nil {
				slog.Error("Failed to check JWT revocation status, denying request.", "user_id", claims.UserID, "error", err)
				// Fail closed: deny request if Redis check fails
				// This ensures security is maintained even during Redis issues
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service temporarily unavailable"})
				return
			}

			if isRevoked {
				// Token found in blacklist - it was revoked
				slog.Warn("SECURITY: Blocked revoked JWT usage.", "user_id", claims.UserID, "jti", claims.ID, "ip", ctx.ClientIP())
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				return
			}

			ctx.Set("userID", claims.UserID)
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}
}
