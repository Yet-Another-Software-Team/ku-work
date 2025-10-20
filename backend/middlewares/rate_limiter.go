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

// RateLimiter stores rate limit data for IP addresses
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
		now := time.Now()
		entriesToDelete := make([]string, 0)

		// First pass: identify entries to delete (with read lock)
		rl.mu.RLock()
		for ip, entry := range rl.entries {
			entry.mu.Lock()
			// Mark entries that are past both reset times for deletion
			if now.After(entry.minuteResetAt) && now.After(entry.hourResetAt) {
				entriesToDelete = append(entriesToDelete, ip)
			}
			entry.mu.Unlock()
		}
		rl.mu.RUnlock()

		// Second pass: delete marked entries (with write lock)
		if len(entriesToDelete) > 0 {
			rl.mu.Lock()
			for _, ip := range entriesToDelete {
				delete(rl.entries, ip)
			}
			rl.mu.Unlock()
		}
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

// RateLimiterWithLimits creates a configurable rate limiting middleware
// minuteLimit: maximum attempts per minute per IP
// hourLimit: maximum attempts per hour per IP
func RateLimiterWithLimits(minuteLimit, hourLimit int) gin.HandlerFunc {
	limiter := GetRateLimiter()

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()

		allowed, message := limiter.checkLimit(ip, minuteLimit, hourLimit)
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
