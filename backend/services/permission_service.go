// TODO: migrate to layered architecture

package services

import (
	"context"
	"fmt"
	"time"

	"ku-work/backend/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PermissionService struct {
	DB *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{DB: db}
}

func (s *PermissionService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.DB == nil {
		return false, fmt.Errorf("permission service or database is not initialized")
	}

	var count int64
	if err := s.DB.WithContext(ctx).Model(&model.Admin{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to query admin permission: %w", err)
	}

	return count > 0, nil
}

type RedisRateLimiter struct {
	Redis    *redis.Client
	FailOpen bool
}

func NewRedisRateLimiter(redisClient *redis.Client, failOpen bool) *RedisRateLimiter {
	return &RedisRateLimiter{Redis: redisClient, FailOpen: failOpen}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, ip string, minuteLimit, hourLimit int) (bool, string, error) {
	if minuteLimit <= 0 || hourLimit <= 0 {
		return false, "Invalid rate limit configuration", fmt.Errorf("minute and hour limits must be positive")
	}

	if r == nil || r.Redis == nil {
		if r != nil && !r.FailOpen {
			return false, "", fmt.Errorf("redis client is not initialized")
		}
		return true, "", nil
	}

	now := time.Now()
	minuteKey := fmt.Sprintf("ratelimit:%s:minute:%d", ip, now.Unix()/60)
	hourKey := fmt.Sprintf("ratelimit:%s:hour:%d", ip, now.Unix()/3600)

	pipe := r.Redis.Pipeline()

	minuteIncr := pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 2*time.Minute)

	hourIncr := pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 2*time.Hour)

	if _, err := pipe.Exec(ctx); err != nil {
		if r.FailOpen {
			return true, "", nil
		}
		return false, "", fmt.Errorf("redis pipeline exec failed: %w", err)
	}

	minuteCount, err := minuteIncr.Result()
	if err != nil {
		if r.FailOpen {
			return true, "", nil
		}
		return false, "", fmt.Errorf("failed to read minute counter: %w", err)
	}
	if minuteCount > int64(minuteLimit) {
		return false, "Too many requests. Please try again later.", nil
	}

	hourCount, err := hourIncr.Result()
	if err != nil {
		if r.FailOpen {
			return true, "", nil
		}
		return false, "", fmt.Errorf("failed to read hour counter: %w", err)
	}
	if hourCount > int64(hourLimit) {
		return false, "Too many requests. Please try again in an hour.", nil
	}

	return true, "", nil
}
