package services

import (
	"context"

	"ku-work/backend/repository"
)

// AccountStatusService handles account status-related operations using the repository layer
type AccountStatusService struct {
	identityRepo repository.IdentityRepository
}

// NewAccountStatusService creates a new AccountStatusService
func NewAccountStatusService(identityRepo repository.IdentityRepository) *AccountStatusService {
	return &AccountStatusService{
		identityRepo: identityRepo,
	}
}

// IsAccountDeactivated checks if a user account is deactivated (soft-deleted)
func (s *AccountStatusService) IsAccountDeactivated(ctx context.Context, userID string) (bool, error) {
	return s.identityRepo.IsUserDeactivated(ctx, userID)
}
