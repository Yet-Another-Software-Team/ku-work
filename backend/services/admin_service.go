package services

import (
	"ku-work/backend/model"
	"ku-work/backend/repository"
	"log"
)

// AdminService defines service-level operations for admin functionality.
type AdminService interface {
	// Offset is the number of records to skip and limit is the maximum number to return.

	// FetchAuditLog returns audit entries using pagination parameters.
	FetchAuditLog(offset, limit uint) ([]model.Audit, error)
	
	// FetchMailLog returns mail log entries using pagination parameters.
	FetchMailLog(offset, limit uint) ([]model.MailLog, error)
	
}

// adminService is the default implementation of AdminService.
type adminService struct {
	repo repository.AuditRepository
}

// NewAdminService constructs a new AdminService backed by the provided repository.
func NewAdminService(repo repository.AuditRepository) AdminService {
	if repo == nil {
		log.Fatal("audit repository cannot be nil")
	}
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
	return s.repo.FetchAuditLog(offset, limit)
}

func (s *adminService) FetchMailLog(offset, limit uint) ([]model.MailLog, error) {
	if limit == 0 {
		limit = 32
	}
	if limit > 64 {
		limit = 64
	}
	return s.repo.FetchMailLog(offset, limit)
}
