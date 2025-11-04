package services

import (
	"context"
	"fmt"
	"time"

	repo "ku-work/backend/repository"
)

// JWTRevocationService is a thin wrapper around a repository.JWTRevocationRepository.
// This allows the service layer to depend on an abstraction rather than a concrete
// Redis client implementation.
type JWTRevocationService struct {
	jwtRepo repo.JWTRevocationRepository
}

// NewJWTRevocationService creates a new JWTRevocationService backed by the provided repository.
func NewJWTRevocationService(r repo.JWTRevocationRepository) *JWTRevocationService {
	return &JWTRevocationService{jwtRepo: r}
}

// RevokeJWT delegates revocation to the underlying repository implementation.
func (s *JWTRevocationService) RevokeJWT(ctx context.Context, jti string, userID string, expiresAt time.Time) error {
	if s == nil || s.jwtRepo == nil {
		return fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.RevokeJWT(ctx, jti, userID, expiresAt)
}

// IsJWTRevoked delegates to the repository to check revocation status.
func (s *JWTRevocationService) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	if s == nil || s.jwtRepo == nil {
		return false, fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.IsJWTRevoked(ctx, jti)
}

// GetRevokedJWTInfo delegates retrieving metadata about a revoked JWT to the repository.
func (s *JWTRevocationService) GetRevokedJWTInfo(ctx context.Context, jti string) (*repo.RevokedJWTInfo, error) {
	if s == nil || s.jwtRepo == nil {
		return nil, fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.GetRevokedJWTInfo(ctx, jti)
}

// RevokeAllUserJWTs delegates creation of a user-level revocation marker to the repository.
func (s *JWTRevocationService) RevokeAllUserJWTs(ctx context.Context, userID string, ttl time.Duration) error {
	if s == nil || s.jwtRepo == nil {
		return fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.RevokeAllUserJWTs(ctx, userID, ttl)
}

// IsUserJWTsRevoked delegates checking a user-level revocation marker to the repository.
func (s *JWTRevocationService) IsUserJWTsRevoked(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.jwtRepo == nil {
		return false, fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.IsUserJWTsRevoked(ctx, userID)
}

// CleanupExpiredJWTs delegates any cleanup to the repository (often a no-op for TTL-backed stores).
func (s *JWTRevocationService) CleanupExpiredJWTs(ctx context.Context) error {
	if s == nil || s.jwtRepo == nil {
		return fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.CleanupExpiredJWTs(ctx)
}

// GetStats delegates collection of revocation stats to the repository.
func (s *JWTRevocationService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	if s == nil || s.jwtRepo == nil {
		return nil, fmt.Errorf("revocation repository is not initialized")
	}
	return s.jwtRepo.GetStats(ctx)
}
