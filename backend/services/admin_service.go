package services

import (
	"ku-work/backend/model"
	"ku-work/backend/repository"
)

// AdminService defines service-level operations for admin functionality.
type AdminService interface {
	// FetchAuditLog returns audit entries using pagination parameters.
	// Offset is the number of records to skip and limit is the maximum number to return.
	FetchAuditLog(offset, limit uint) ([]model.Audit, error)
}

// adminService is the default implementation of AdminService.
type adminService struct {
	repo repository.AuditRepository
}

// NewAdminService constructs a new AdminService backed by the provided repository.
func NewAdminService(repo repository.AuditRepository) AdminService {
	return &adminService{repo: repo}
}

// FetchAuditLog implements AdminService.FetchAuditLog.
func (s *adminService) FetchAuditLog(offset, limit uint) ([]model.Audit, error) {
	if limit == 0 {
		limit = 32
	}
	if limit > 64 {
		limit = 64
	}
	return s.repo.Fetch(offset, limit)
}
