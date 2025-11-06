package services

import (
	"context"
	"errors"
	"time"

	repo "ku-work/backend/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// CompanyService provides read operations for company data.
type CompanyService struct {
	companyRepo repo.CompanyRepository
}

// NewCompanyServiceWithRepo creates a CompanyService with an injected repository implementation.
func NewCompanyService(r repo.CompanyRepository) *CompanyService {
	return &CompanyService{
		companyRepo: r,
	}
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
