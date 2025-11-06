package gormrepo

import (
	"ku-work/backend/model"
	"ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormAuditRepository struct {
	db *gorm.DB
}

func NewGormAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &GormAuditRepository{db: db}
}

func (r *GormAuditRepository) FetchAuditLog(offset, limit uint) ([]model.Audit, error) {
	var audits []model.Audit

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

func (r *GormAuditRepository) FetchMailLog(offset uint, limit uint) ([]model.MailLog, error) {
	var mailLogs []model.MailLog

	if limit == 0 {
		limit = 32
	}

	result := r.db.
		Model(&model.MailLog{}).
		Offset(int(offset)).
		Limit(int(limit)).
		Order("created_at desc").
		Find(&mailLogs)

	return mailLogs, result.Error
}
