package services

import (
	"context"
	"fmt"
	"time"

	repo "ku-work/backend/repository"
)

// RateLimiter defines the contract for rate limiting services.
// Implementations return:
//
//	allowed: whether the request should proceed
//	message: user-facing message when not allowed (empty if allowed)
//	err: internal error (infrastructure/config) – callers may treat this as 503 or fail-open
type RateLimiter interface {
	Allow(ctx context.Context, ip string, minuteLimit, hourLimit int) (allowed bool, message string, err error)
}

// RateLimiterService is a repository-backed fixed-window rate limiter.
// It delegates all storage operations to a RateLimitRepository so the service layer
// never directly accesses infrastructure (Redis, etc.).
type RateLimiterService struct {
	repo     repo.RateLimitRepository
	failOpen bool
}

// NewRateLimiterService constructs a RateLimiterService.
//
//	rateRepo: implementation of RateLimitRepository (e.g., Redis)
//	failOpen: if true, requests are allowed when repository errors occur.
func NewRateLimiterService(rateRepo repo.RateLimitRepository, failOpen bool) *RateLimiterService {
	return &RateLimiterService{
		repo:     rateRepo,
		failOpen: failOpen,
	}
}

// Allow applies two fixed windows (per-minute and per-hour) using time-bucketed keys:
//
//	ratelimit:<ip>:minute:<bucket>
//	ratelimit:<ip>:hour:<bucket>
//
// TTLs are set slightly larger than the window size to tolerate minor delays.
func (s *RateLimiterService) Allow(ctx context.Context, ip string, minuteLimit, hourLimit int) (bool, string, error) {
	if minuteLimit <= 0 || hourLimit <= 0 {
		return false, "Invalid rate limit configuration", fmt.Errorf("minute and hour limits must be positive")
	}

	if s == nil || s.repo == nil {
		// If service or repository not wired: fail-open if configured, else error.
		if s != nil && !s.failOpen {
			return false, "", fmt.Errorf("rate limiter repository not initialized")
		}
		return true, "", nil
	}

	now := time.Now()
	minuteKey := fmt.Sprintf("ratelimit:%s:minute:%d", ip, now.Unix()/60)
	hourKey := fmt.Sprintf("ratelimit:%s:hour:%d", ip, now.Unix()/3600)

	items := []repo.RateLimitItem{
		{Key: minuteKey, TTL: 2 * time.Minute},
		{Key: hourKey, TTL: 2 * time.Hour},
	}

	counts, err := s.repo.IncrBatchWithTTL(ctx, items)
	if err != nil {
		if s.failOpen {
			return true, "", nil
		}
		return false, "", fmt.Errorf("rate limit repository error: %w", err)
	}
	if len(counts) != 2 {
		if s.failOpen {
			return true, "", nil
		}
		return false, "", fmt.Errorf("unexpected counters length from repository: %d", len(counts))
	}

	// Enforce limits
	if counts[0] > int64(minuteLimit) {
		return false, "Too many requests. Please try again later.", nil
	}
	if counts[1] > int64(hourLimit) {
		return false, "Too many requests. Please try again in an hour.", nil
	}

	return true, "", nil
}
