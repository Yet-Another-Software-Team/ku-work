package middlewares

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterWithLimits creates a configurable rate limiting middleware using Redis
// minuteLimit: maximum attempts per minute per IP
// hourLimit: maximum attempts per hour per IP
func RateLimiterWithLimits(redisClient *redis.Client, minuteLimit, hourLimit int) gin.HandlerFunc {
	// Create a service that encapsulates Redis rate limiting behaviour.
	rateLimiter := services.NewRedisRateLimiter(redisClient, true)

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		allowed, message, err := rateLimiter.Allow(ctx.Request.Context(), ip, minuteLimit, hourLimit)
		if err != nil {
			// If the limiter returned an error, treat it as internal service error.
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limiting service unavailable"})
			ctx.Abort()
			return
		}

		if !allowed {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": message,
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
