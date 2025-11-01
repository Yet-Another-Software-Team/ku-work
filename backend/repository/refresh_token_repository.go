package repository

import (
	"context"
	"time"

	"ku-work/backend/model"
)

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	// Create persists a new refresh token record.
	Create(ctx context.Context, rt *model.RefreshToken) error
	CountActiveByUser(ctx context.Context, userID string, now time.Time) (int64, error)
	FindBySelector(ctx context.Context, selector string) (*model.RefreshToken, error)
	FindOldestActiveByUserLimit(ctx context.Context, userID string, limit int) ([]model.RefreshToken, error)
	RevokeByIDs(ctx context.Context, ids []uint, revokedAt time.Time) error
	RevokeBySelector(ctx context.Context, selector string, revokedAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID string, revokedAt time.Time) error
	UpdateRevokedAt(ctx context.Context, id uint, revokedAt time.Time) error
}
