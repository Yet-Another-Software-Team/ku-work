package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// checkLimit verifies if the IP is within rate limits using Redis
func checkLimit(redisClient *redis.Client, ip string, minuteLimit, hourLimit int) (bool, string) {
	if redisClient == nil {
		// Fallback: allow request if Redis is not available
		return true, ""
	}

	ctx := context.Background()
	now := time.Now()
	minuteKey := fmt.Sprintf("ratelimit:%s:minute:%d", ip, now.Unix()/60)
	hourKey := fmt.Sprintf("ratelimit:%s:hour:%d", ip, now.Unix()/3600)

	// Use pipeline for better performance
	pipe := redisClient.Pipeline()

	// Increment minute counter
	minuteIncr := pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 2*time.Minute) // Keep for 2 minutes to avoid race conditions

	// Increment hour counter
	hourIncr := pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 2*time.Hour) // Keep for 2 hours to avoid race conditions

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		// If Redis fails, allow the request (fail open)
		return true, ""
	}

	// Check minute limit
	minuteCount, err := minuteIncr.Result()
	if err != nil {
		return true, ""
	}
	if minuteCount > int64(minuteLimit) {
		return false, "Too many requests. Please try again later."
	}

	// Check hour limit
	hourCount, err := hourIncr.Result()
	if err != nil {
		return true, ""
	}
	if hourCount > int64(hourLimit) {
		return false, "Too many requests. Please try again in an hour."
	}

	return true, ""
}

// RateLimiterWithLimits creates a configurable rate limiting middleware using Redis
// minuteLimit: maximum attempts per minute per IP
// hourLimit: maximum attempts per hour per IP
func RateLimiterWithLimits(redisClient *redis.Client, minuteLimit, hourLimit int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		allowed, message := checkLimit(redisClient, ip, minuteLimit, hourLimit)
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
