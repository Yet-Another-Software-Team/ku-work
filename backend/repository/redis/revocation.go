package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ku-work/backend/repository"

	"github.com/redis/go-redis/v9"
)

// RedisRevocationRepository is a Redis-backed implementation of repository.JWTRevocationRepository.
type RedisRevocationRepository struct {
	client *redis.Client
}

// NewRedisRevocationRepository constructs a new RedisRevocationRepository.
func NewRedisRevocationRepository(client *redis.Client) *RedisRevocationRepository {
	return &RedisRevocationRepository{client: client}
}

// Ensure RedisRevocationRepository implements the interface.
var _ repository.JWTRevocationRepository = (*RedisRevocationRepository)(nil)

// keyRevokedJWT returns the Redis key for a revoked JTI.
func keyRevokedJWT(jti string) string {
	return fmt.Sprintf("revoked:jwt:%s", jti)
}

// keyRevokedUser returns the Redis key for a user-level revocation marker.
func keyRevokedUser(userID string) string {
	return fmt.Sprintf("revoked:user:%s", userID)
}

// RevokeJWT stores the JTI with TTL equal to time until expiresAt.
func (r *RedisRevocationRepository) RevokeJWT(ctx context.Context, jti string, userID string, expiresAt time.Time) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, nothing to store.
		return nil
	}

	info := repository.RevokedJWTInfo{
		UserID:    userID,
		RevokedAt: time.Now(),
	}
	b, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal revocation info: %w", err)
	}

	if err := r.client.Set(ctx, keyRevokedJWT(jti), b, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store revoked JWT in Redis: %w", err)
	}
	return nil
}

// IsJWTRevoked returns true if a JTI is present in the revocation list.
func (r *RedisRevocationRepository) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	key := keyRevokedJWT(jti)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check JWT revocation status in Redis: %w", err)
	}
	return exists > 0, nil
}

// GetRevokedJWTInfo returns stored metadata for a revoked JWT, or (nil, nil) if not present.
func (r *RedisRevocationRepository) GetRevokedJWTInfo(ctx context.Context, jti string) (*repository.RevokedJWTInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	key := keyRevokedJWT(jti)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get revoked JWT info from Redis: %w", err)
	}

	var info repository.RevokedJWTInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal revocation info: %w", err)
	}
	return &info, nil
}

// RevokeAllUserJWTs creates a user-level revocation marker with the specified TTL.
func (r *RedisRevocationRepository) RevokeAllUserJWTs(ctx context.Context, userID string, ttl time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	key := keyRevokedUser(userID)
	timestamp := time.Now().Format(time.RFC3339)
	if err := r.client.Set(ctx, key, timestamp, ttl).Err(); err != nil {
		return fmt.Errorf("failed to revoke all user JWTs in Redis: %w", err)
	}
	return nil
}

// IsUserJWTsRevoked checks whether a user-level revocation marker exists.
func (r *RedisRevocationRepository) IsUserJWTsRevoked(ctx context.Context, userID string) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}
	key := keyRevokedUser(userID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user JWT revocation status in Redis: %w", err)
	}
	return exists > 0, nil
}

// CleanupExpiredJWTs is a no-op because Redis handles expiration via TTL.
func (r *RedisRevocationRepository) CleanupExpiredJWTs(ctx context.Context) error {
	// Redis TTLs expire automatically; nothing to clean here.
	return nil
}

// GetStats returns simple diagnostics about revoked JTIs (counts keys).
func (r *RedisRevocationRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	var cursor uint64
	var revokedCount int64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "revoked:jwt:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan revoked JWT keys: %w", err)
		}
		revokedCount += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	stats := map[string]interface{}{
		"revoked_jwt_count": revokedCount,
		"timestamp":         time.Now().Format(time.RFC3339),
	}
	return stats, nil
}
