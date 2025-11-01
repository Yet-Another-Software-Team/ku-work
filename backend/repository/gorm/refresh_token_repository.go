package gormrepo

import (
	"context"
	"time"

	"ku-work/backend/model"

	"gorm.io/gorm"
)

// GormRefreshTokenRepository is a GORM-backed implementation of the
// repository.RefreshTokenRepository interface.
//
// This adapter encapsulates all DB operations related to refresh tokens so
// service-layer code does not need to depend on *gorm.DB directly.
type GormRefreshTokenRepository struct {
	db *gorm.DB
}

// NewGormRefreshTokenRepository constructs a new GORM-backed refresh-token repository.
func NewGormRefreshTokenRepository(db *gorm.DB) *GormRefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

// Create persists a new refresh token record.
func (r *GormRefreshTokenRepository) Create(ctx context.Context, rt *model.RefreshToken) error {
	if rt == nil {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Create(rt).Error
}

// CountActiveByUser returns the number of active (non-revoked, non-expired) refresh tokens
// for the given user as of the provided reference time.
func (r *GormRefreshTokenRepository) CountActiveByUser(ctx context.Context, userID string, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Count(&count).Error
	return count, err
}

// FindBySelector returns the refresh token record identified by the public selector value.
// If not found, (nil, nil) is returned.
func (r *GormRefreshTokenRepository) FindBySelector(ctx context.Context, selector string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_selector = ?", selector).
		First(&rt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rt, nil
}

// FindOldestActiveByUserLimit returns up to `limit` oldest active (non-revoked) refresh tokens
// for the specified user ordered by creation time ascending.
func (r *GormRefreshTokenRepository) FindOldestActiveByUserLimit(ctx context.Context, userID string, limit int) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("created_at ASC").
		Limit(limit).
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// RevokeByIDs marks the given refresh token IDs as revoked (sets revoked_at).
func (r *GormRefreshTokenRepository) RevokeByIDs(ctx context.Context, ids []uint, revokedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id IN ?", ids).
		Update("revoked_at", revokedAt).Error
}

// RevokeBySelector revokes a single refresh token identified by its selector.
func (r *GormRefreshTokenRepository) RevokeBySelector(ctx context.Context, selector string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_selector = ? AND revoked_at IS NULL", selector).
		Update("revoked_at", revokedAt).Error
}

// RevokeAllForUser revokes all active refresh tokens for the specified user.
func (r *GormRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
}

// UpdateRevokedAt updates the revoked_at timestamp for a specific refresh token ID.
func (r *GormRefreshTokenRepository) UpdateRevokedAt(ctx context.Context, id uint, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", revokedAt).Error
}
