package middlewares

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

// RateLimiterWithLimits creates a configurable rate limiting middleware using Redis
// minuteLimit: maximum attempts per minute per IP
// hourLimit: maximum attempts per hour per IP
func RateLimiterWithLimits(rateLimiter services.RateLimiter, minuteLimit, hourLimit int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		if rateLimiter == nil {
			// If not wired, allow traffic (fail-open) to avoid blocking core flows.
			ctx.Next()
			return
		}
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
