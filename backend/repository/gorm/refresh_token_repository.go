package gormrepo

import (
	"context"
	"time"

	"ku-work/backend/model"

	"gorm.io/gorm"
)

type GormRefreshTokenRepository struct {
	db *gorm.DB
}

func NewGormRefreshTokenRepository(db *gorm.DB) *GormRefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

func (r *GormRefreshTokenRepository) Create(ctx context.Context, rt *model.RefreshToken) error {
	if rt == nil {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *GormRefreshTokenRepository) CountActiveByUser(ctx context.Context, userID string, now time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Count(&count).Error
	return count, err
}

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

func (r *GormRefreshTokenRepository) RevokeByIDs(ctx context.Context, ids []uint, revokedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id IN ?", ids).
		Update("revoked_at", revokedAt).Error
}

func (r *GormRefreshTokenRepository) RevokeBySelector(ctx context.Context, selector string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_selector = ? AND revoked_at IS NULL", selector).
		Update("revoked_at", revokedAt).Error
}

func (r *GormRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
}

func (r *GormRefreshTokenRepository) UpdateRevokedAt(ctx context.Context, id uint, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", revokedAt).Error
}
