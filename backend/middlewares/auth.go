package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"ku-work/backend/model"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware creates an authentication middleware that uses an injected JWTRevocationService
// to check for revoked tokens. This allows the caller to provide the revocation implementation
// (Redis-based or another implementation) and makes the middleware easier to test.
func AuthMiddleware(jwtSecret []byte, revocationService *services.JWTRevocationService) gin.HandlerFunc {
	if revocationService == nil {
		log.Fatal("FATAL: JWT revocation service is nil. Authentication requires a revocation service to be provided.")
	}

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
			isRevoked, err := revocationService.IsJWTRevoked(context.Background(), claims.ID)
			if err != nil {
				log.Printf("ERROR: Failed to check JWT revocation status for user %s: %v. Denying request.", claims.UserID, err)
				// Fail closed: deny request if revocation check fails
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service temporarily unavailable"})
				return
			}

			if isRevoked {
				// Token found in blacklist - it was revoked
				log.Printf("SECURITY: Blocked revoked JWT usage. User: %s, JTI: %s, IP: %s",
					claims.UserID, claims.ID, ctx.ClientIP())
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

// AuthMiddlewareWithRedis is a small compatibility helper that constructs a Redis-backed
// JWTRevocationService and delegates to AuthMiddleware. Existing call sites using the
// older signature can keep working; prefer calling AuthMiddleware with an explicit
// revocation service where possible.
func AuthMiddlewareWithRedis(jwtSecret []byte, redisClient *redis.Client) gin.HandlerFunc {
	// Validate the provided client and construct the revocation service.
	if redisClient == nil {
		log.Fatal("FATAL: Redis client is nil. JWT revocation requires Redis to be available.")
	}

	revocationService := services.NewJWTRevocationService(redisClient)

	// Delegate to the injected-service-based middleware so we keep a single auth flow.
	return AuthMiddleware(jwtSecret, revocationService)
}
