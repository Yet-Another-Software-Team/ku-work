package repository

import (
	"ku-work/backend/model"
)

// AuditRepository defines the contract for audit persistence operations.
type AuditRepository interface {
	FetchAuditLog(offset, limit uint) ([]model.Audit, error)
	FetchMailLog(offset, limit uint) ([]model.MailLog, error)
}
