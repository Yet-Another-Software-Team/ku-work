package repository

import (
	"ku-work/backend/model"
	"time"
)

// AuditRepository defines the contract for audit persistence operations.
type AuditRepository interface {
	FetchAuditLog(offset, limit uint) ([]model.Audit, error)
	FetchMailLog(offset, limit uint) ([]model.MailLog, error)
	CreateAuditLog(log *model.Audit) error
	CreateOrUpdateMailLog(log *model.MailLog) error
	FindMailToRetry(retryAfter time.Time, maxAge time.Time, maxAttempts int) ([]model.MailLog, error)
}
