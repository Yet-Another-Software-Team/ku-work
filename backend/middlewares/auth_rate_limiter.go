package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// checkAuthLimit verifies if the IP is within rate limits for failed authentication attempts.
// It does not increment the counter here.
func checkAuthLimit(redisClient *redis.Client, ip string, minuteLimit, hourLimit int) (bool, string) {
	if redisClient == nil {
		return true, "" // Allow if Redis is unavailable
	}

	ctx := context.Background()
	now := time.Now()
	minuteKey := fmt.Sprintf("auth_ratelimit:%s:minute:%d", ip, now.Unix()/60)
	hourKey := fmt.Sprintf("auth_ratelimit:%s:hour:%d", ip, now.Unix()/3600)

	pipe := redisClient.Pipeline()
	minuteCmd := pipe.Get(ctx, minuteKey)
	hourCmd := pipe.Get(ctx, hourKey)
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return true, "" // Allow on Redis error
	}

	minuteCount, _ := minuteCmd.Int64()
	if minuteCount >= int64(minuteLimit) {
		return false, "Too many failed attempts. Please try again later."
	}

	hourCount, _ := hourCmd.Int64()
	if hourCount >= int64(hourLimit) {
		return false, "Too many failed attempts. Please try again in an hour."
	}

	return true, ""
}

// incrementAuthFailedAttempt increments the failed attempt counters for the given IP.
func incrementAuthFailedAttempt(redisClient *redis.Client, ip string) {
	if redisClient == nil {
		return
	}

	ctx := context.Background()
	now := time.Now()
	minuteKey := fmt.Sprintf("auth_ratelimit:%s:minute:%d", ip, now.Unix()/60)
	hourKey := fmt.Sprintf("auth_ratelimit:%s:hour:%d", ip, now.Unix()/3600)

	pipe := redisClient.Pipeline()
	pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 2*time.Minute)
	pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 2*time.Hour)
	_, _ = pipe.Exec(ctx)
}

// AuthRateLimiter creates a rate limiting middleware for authentication routes.
// It only counts requests that result in a 4xx or 5xx status code.
func AuthRateLimiter(redisClient *redis.Client, minuteLimit, hourLimit int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		// Check if the IP is already blocked
		allowed, message := checkAuthLimit(redisClient, ip, minuteLimit, hourLimit)
		if !allowed {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": message})
			ctx.Abort()
			return
		}

		// Process the request
		ctx.Next()

		// After the request, check the status code.
		// If it's a client or server error, and it's an authentication route, increment the counter.
		status := ctx.Writer.Status()
		if status >= 400 && status < 600 {
			incrementAuthFailedAttempt(redisClient, ip)
		}
	}
}
