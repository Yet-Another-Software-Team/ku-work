package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimitEntry tracks attempts for an IP address
type rateLimitEntry struct {
	minuteCount   int
	hourCount     int
	minuteResetAt time.Time
	hourResetAt   time.Time
	mu            sync.Mutex
}
d// RateLimiter stores rate limit data for IP addresses
type RateLimiter struct {
	entries map[string]*rateLimitEntry
	mu      sync.RWMutex
}

var (
	globalRateLimiter *RateLimiter
	once              sync.Once
)

// GetRateLimiter returns the singleton rate limiter instance
func GetRateLimiter() *RateLimiter {
	once.Do(func() {
		globalRateLimiter = &RateLimiter{
			entries: make(map[string]*rateLimitEntry),
		}
		// Start cleanup goroutine
		go globalRateLimiter.cleanup()
	})
	return globalRateLimiter
}

// cleanup removes expired entries every 5 minutes
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			entry.mu.Lock()
			// Remove entries that are past both reset times
			if now.After(entry.minuteResetAt) && now.After(entry.hourResetAt) {
				delete(rl.entries, ip)
			}
			entry.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// checkLimit verifies if the IP is within rate limits
func (rl *RateLimiter) checkLimit(ip string, minuteLimit, hourLimit int) (bool, string) {
	rl.mu.Lock()
	entry, exists := rl.entries[ip]
	if !exists {
		entry = &rateLimitEntry{
			minuteResetAt: time.Now().Add(time.Minute),
			hourResetAt:   time.Now().Add(time.Hour),
		}
		rl.entries[ip] = entry
	}
	rl.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()

	// Reset minute counter if needed
	if now.After(entry.minuteResetAt) {
		entry.minuteCount = 0
		entry.minuteResetAt = now.Add(time.Minute)
	}

	// Reset hour counter if needed
	if now.After(entry.hourResetAt) {
		entry.hourCount = 0
		entry.hourResetAt = now.Add(time.Hour)
	}

	// Check limits
	if entry.minuteCount >= minuteLimit {
		return false, "Too many requests. Please try again later."
	}
	if entry.hourCount >= hourLimit {
		return false, "Too many requests. Please try again in an hour."
	}

	// Increment counters
	entry.minuteCount++
	entry.hourCount++

	return true, ""
}

// RefreshTokenRateLimiter creates a rate limiting middleware for refresh token endpoint
// Limits: 5 attempts per minute per IP, 20 attempts per hour per IP
func RefreshTokenRateLimiter() gin.HandlerFunc {
	limiter := GetRateLimiter()

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		allowed, message := limiter.checkLimit(ip, 5, 20)
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
