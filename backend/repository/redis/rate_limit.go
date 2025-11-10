package redisrepo

import (
	"context"
	"fmt"
	"time"

	repo "ku-work/backend/repository"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimitRepository is a Redis-backed implementation of repo.RateLimitRepository.
// It uses time-bucketed fixed windows keyed by a caller-provided string and refreshes TTL
// on every increment so sparse traffic does not prematurely evict counters.
type RedisRateLimitRepository struct {
	client *redis.Client
}

// NewRedisRateLimitRepository constructs a new rate limit repository with the given Redis client.
func NewRedisRateLimitRepository(client *redis.Client) *RedisRateLimitRepository {
	return &RedisRateLimitRepository{client: client}
}

// Ensure interface compliance at compile time.
var _ repo.RateLimitRepository = (*RedisRateLimitRepository)(nil)

// IncrWithTTL increments a single key and sets (or refreshes) its expiration.
// Returns the post-increment value of the counter.
func (r *RedisRateLimitRepository) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("redis rate limit repository not initialized")
	}
	if key == "" {
		return 0, fmt.Errorf("key must not be empty")
	}
	if ttl <= 0 {
		return 0, fmt.Errorf("ttl must be positive")
	}

	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis pipeline exec failed for key %s: %w", key, err)
	}

	val, err := incr.Result()
	if err != nil {
		return 0, fmt.Errorf("failed to read incremented value for key %s: %w", key, err)
	}
	return val, nil
}

// IncrBatchWithTTL increments multiple keys atomically (best-effort) using a pipeline,
// setting/refreshing their TTLs. Returns post-increment values in the same order as items.
func (r *RedisRateLimitRepository) IncrBatchWithTTL(ctx context.Context, items []repo.RateLimitItem) ([]int64, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("redis rate limit repository not initialized")
	}
	if len(items) == 0 {
		return []int64{}, nil
	}

	pipe := r.client.TxPipeline()
	incrCmds := make([]*redis.IntCmd, len(items))

	for i, it := range items {
		if it.Key == "" {
			return nil, fmt.Errorf("rate limit item[%d] key must not be empty", i)
		}
		if it.TTL <= 0 {
			return nil, fmt.Errorf("rate limit item[%d] TTL must be positive", i)
		}
		incrCmds[i] = pipe.Incr(ctx, it.Key)
		pipe.Expire(ctx, it.Key, it.TTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("redis batch pipeline exec failed: %w", err)
	}

	results := make([]int64, len(items))
	for i, cmd := range incrCmds {
		v, err := cmd.Result()
		if err != nil {
			return nil, fmt.Errorf("failed to read incremented value for key %s: %w", items[i].Key, err)
		}
		results[i] = v
	}
	return results, nil
}
