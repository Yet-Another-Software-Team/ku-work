package services

import (
	"context"
	"errors"
	"time"

	"ku-work/backend/helper"
	repo "ku-work/backend/repository"
)

var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidWebsite     = errors.New("invalid website url")
)

// CompanyService provides read operations for company data.
//
// It uses the repository abstraction for all DB access related to companies.
// Construct via NewCompanyService (convenience constructor that wires a GORM-backed repository)
// or NewCompanyServiceWithRepo to inject a repository (preferred for testing).
type CompanyService struct {
	companyRepo repo.CompanyRepository
}

// NewCompanyServiceWithRepo creates a CompanyService with an injected repository implementation.
func NewCompanyService(r repo.CompanyRepository) *CompanyService {
	return &CompanyService{
		companyRepo: r,
	}
}

// IsAdmin is a convenience method retained for callers that expect a simple boolean.
// It uses the repository GetRole implementation under the hood but provides a
// backward-compatible signature by running the call with background context and
// swallowing errors (returns false on error).
func (s *CompanyService) IsAdmin(userID string) bool {
	if s == nil || s.companyRepo == nil {
		return false
	}
	role, err := s.companyRepo.GetRole(context.Background(), userID)
	if err != nil {
		// On error, conservatively treat as non-admin.
		return false
	}
	return role == helper.Admin
}

// IsAdminCtx is the preferred form: callers pass a context and receive an error
// if role resolution fails. Handlers/middlewares should switch to this method
// and pass request context.
func (s *CompanyService) IsAdminCtx(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.companyRepo == nil {
		return false, nil
	}
	role, err := s.companyRepo.GetRole(ctx, userID)
	if err != nil {
		return false, err
	}
	return role == helper.Admin, nil
}

// CompanyResponse mirrors the public response shape for company profiles.
type CompanyResponse struct {
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   string    `json:"photoId"`
	BannerID  string    `json:"bannerId"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Website   string    `json:"website"`
	AboutUs   string    `json:"about"`
	Name      string    `json:"name"`
}

// GetCompanyByUserID fetches company profile and the owner's username.
// Accepts a context so callers (handlers/middlewares) don't need direct DB access.
func (s *CompanyService) GetCompanyByUserID(ctx context.Context, userID string) (CompanyResponse, error) {
	proj, err := s.companyRepo.GetCompanyProjectionByUserID(ctx, userID)
	if err != nil {
		return CompanyResponse{}, err
	}
	resp := CompanyResponse{
		CreatedAt: proj.CreatedAt,
		UserID:    proj.UserID,
		Email:     proj.Email,
		Phone:     proj.Phone,
		PhotoID:   proj.PhotoID,
		BannerID:  proj.BannerID,
		Address:   proj.Address,
		City:      proj.City,
		Country:   proj.Country,
		Website:   proj.Website,
		AboutUs:   proj.AboutUs,
		Name:      proj.Name,
	}
	return resp, nil
}

// ListCompanies returns all companies with owner username populated.
// Accepts a context so callers receive anonymized projections from the repository without touching the DB.
func (s *CompanyService) ListCompanies(ctx context.Context) ([]CompanyResponse, error) {
	projs, err := s.companyRepo.ListCompanyProjections(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CompanyResponse, 0, len(projs))
	for _, p := range projs {
		out = append(out, CompanyResponse{
			CreatedAt: p.CreatedAt,
			UserID:    p.UserID,
			Email:     p.Email,
			Phone:     p.Phone,
			PhotoID:   p.PhotoID,
			BannerID:  p.BannerID,
			Address:   p.Address,
			City:      p.City,
			Country:   p.Country,
			Website:   p.Website,
			AboutUs:   p.AboutUs,
			Name:      p.Name,
		})
	}
	return out, nil
}