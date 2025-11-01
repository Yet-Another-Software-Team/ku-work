package repository

import (
	"ku-work/backend/model"
)

// AuditRepository defines the contract for audit persistence operations.
type AuditRepository interface {
	Fetch(offset, limit uint) ([]model.Audit, error)
}
