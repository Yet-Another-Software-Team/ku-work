package gormrepo

import (
	"ku-work/backend/model"
	"ku-work/backend/repository"

	"gorm.io/gorm"
)

// GormAuditRepository is a GORM-based implementation of AuditRepository.
type GormAuditRepository struct {
	db *gorm.DB
}

// NewGormAuditRepository constructs a new GormAuditRepository.
func NewGormAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &GormAuditRepository{db: db}
}

// Fetch retrieves audit records from the database ordered by newest first.
func (r *GormAuditRepository) Fetch(offset, limit uint) ([]model.Audit, error) {
	var audits []model.Audit

	// Provide a sensible default if caller passes 0 for limit.
	if limit == 0 {
		limit = 32
	}

	result := r.db.
		Model(&model.Audit{}).
		Offset(int(offset)).
		Limit(int(limit)).
		Order("created_at desc").
		Find(&audits)

	return audits, result.Error
}