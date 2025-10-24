package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RevokedJWTInfo stores metadata about a revoked JWT token
type RevokedJWTInfo struct {
	UserID    string    `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
}

// JWTRevocationService handles JWT token revocation using Redis
type JWTRevocationService struct {
	redis *redis.Client
}

// NewJWTRevocationService creates a new JWT revocation service
func NewJWTRevocationService(redisClient *redis.Client) *JWTRevocationService {
	return &JWTRevocationService{
		redis: redisClient,
	}
}

// RevokeJWT adds a JWT to the revocation list (blacklist) in Redis
// The key will automatically expire when the JWT expires (TTL)
func (s *JWTRevocationService) RevokeJWT(ctx context.Context, jti string, userID string, expiresAt time.Time) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	// Calculate TTL: time until the JWT expires
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Create revocation info
	info := RevokedJWTInfo{
		UserID:    userID,
		RevokedAt: time.Now(),
	}

	// Serialize to JSON
	infoJSON, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal revocation info: %w", err)
	}

	// Store in Redis with key format: "revoked:jwt:{JTI}"
	key := fmt.Sprintf("revoked:jwt:%s", jti)
	err = s.redis.Set(ctx, key, infoJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store revoked JWT in Redis: %w", err)
	}

	return nil
}

// IsJWTRevoked checks if a JWT is in the revocation list (blacklist)
func (s *JWTRevocationService) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	if s.redis == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("revoked:jwt:%s", jti)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check JWT revocation status in Redis: %w", err)
	}

	return exists > 0, nil
}

// GetRevokedJWTInfo retrieves information about a revoked JWT
func (s *JWTRevocationService) GetRevokedJWTInfo(ctx context.Context, jti string) (*RevokedJWTInfo, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("revoked:jwt:%s", jti)
	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found - token is not revoked
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get revoked JWT info from Redis: %w", err)
	}

	var info RevokedJWTInfo
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal revocation info: %w", err)
	}

	return &info, nil
}

// RevokeAllUserJWTs revokes all JWTs for a specific user by setting a user-level blacklist marker
// This is useful for security events like password changes or account compromise
// Note: This creates a user-level blacklist entry with a reasonable TTL (24 hours)
func (s *JWTRevocationService) RevokeAllUserJWTs(ctx context.Context, userID string, ttl time.Duration) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("revoked:user:%s", userID)
	timestamp := time.Now().Format(time.RFC3339)

	err := s.redis.Set(ctx, key, timestamp, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke all user JWTs in Redis: %w", err)
	}

	return nil
}

// IsUserJWTsRevoked checks if all JWTs for a user have been revoked
func (s *JWTRevocationService) IsUserJWTsRevoked(ctx context.Context, userID string) (bool, error) {
	if s.redis == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	key := fmt.Sprintf("revoked:user:%s", userID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check user JWT revocation status in Redis: %w", err)
	}

	return exists > 0, nil
}

// CleanupExpiredJWTs is a no-op for Redis-based revocation since Redis handles expiration automatically via TTL
// This method is kept for compatibility with the existing scheduler interface
func (s *JWTRevocationService) CleanupExpiredJWTs(ctx context.Context) error {
	// Redis automatically removes expired keys via TTL
	// No cleanup needed!
	return nil
}

// GetStats returns statistics about the revoked JWTs in Redis
func (s *JWTRevocationService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}

	// Count revoked JWT keys
	var cursor uint64
	var revokedCount int64

	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, "revoked:jwt:*", 100).Result()
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
