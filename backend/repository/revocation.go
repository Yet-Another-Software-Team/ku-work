package repository

import (
	"context"
	"time"
)

// RevokedJWTInfo stores metadata about a revoked JWT token.
type RevokedJWTInfo struct {
	UserID    string    `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
}

// JWTRevocationRepository defines an abstraction for JWT revocation storage.
type JWTRevocationRepository interface {
	RevokeJWT(ctx context.Context, jti string, userID string, expiresAt time.Time) error
	IsJWTRevoked(ctx context.Context, jti string) (bool, error)
	GetRevokedJWTInfo(ctx context.Context, jti string) (*RevokedJWTInfo, error)
	RevokeAllUserJWTs(ctx context.Context, userID string, ttl time.Duration) error
	IsUserJWTsRevoked(ctx context.Context, userID string) (bool, error)
	CleanupExpiredJWTs(ctx context.Context) error
	GetStats(ctx context.Context) (map[string]any, error)
}
