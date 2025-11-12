package services

import (
	"context"
	"fmt"
	"log"

	repo "ku-work/backend/repository"
)

type PermissionService struct {
	identityRepo repo.IdentityRepository
}

func NewPermissionService(identityRepo repo.IdentityRepository) *PermissionService {
	if identityRepo == nil {
		log.Fatal("identity repository cannot be nil")
	}
	return &PermissionService{identityRepo: identityRepo}
}

func (s *PermissionService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.identityRepo == nil {
		return false, fmt.Errorf("permission service or identity repository is not initialized")
	}

	count, err := s.identityRepo.CountAdminByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to query admin permission: %w", err)
	}
	return count > 0, nil
}
