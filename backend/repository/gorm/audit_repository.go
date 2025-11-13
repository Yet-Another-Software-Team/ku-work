package gormrepo

import (
	"ku-work/backend/model"
	"ku-work/backend/repository"
	"time"

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

// CreateAuditLog create entry of autdit log on database from audit entity
func (r *GormAuditRepository) CreateAuditLog(logEntry *model.Audit) error {
	return r.db.Create(logEntry).Error
}

// CreateOrUpdateMailLog create entry of mail log on database from mail log entity or update it if it already exists.
func (r *GormAuditRepository) CreateOrUpdateMailLog(logEntry *model.MailLog) error {
	return r.db.Save(logEntry).Error
}

// Find mails that are eligible for retry
func (r *GormAuditRepository) FindMailToRetry(retryAfter time.Time, maxAge time.Time, maxAttempts int) ([]model.MailLog, error) {
	var failedEmails []model.MailLog
	r.db.Where(
		"status = ? AND updated_at < ? AND created_at > ? AND retry_count < ?",
		model.MailLogStatusTemporaryError,
		retryAfter,
		maxAge,
		maxAttempts,
	).Find(&failedEmails)
	return failedEmails, nil
}
