package services

import (
	"context"
	"errors"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	gormrepo "ku-work/backend/repository/gorm"

	"gorm.io/gorm"
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
	Repo repo.CompanyRepository
}

// NewCompanyService creates a CompanyService wired with a GORM-backed repository.
// Keeping the convenience constructor signature compatible with existing wiring.
func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{
		Repo: gormrepo.NewGormCompanyRepository(db),
	}
}

// NewCompanyServiceWithRepo creates a CompanyService with an injected repository implementation.
func NewCompanyServiceWithRepo(r repo.CompanyRepository) *CompanyService {
	return &CompanyService{
		Repo: r,
	}
}

// IsAdmin is a convenience method retained for callers that expect a simple boolean.
// It uses the repository GetRole implementation under the hood but provides a
// backward-compatible signature by running the call with background context and
// swallowing errors (returns false on error).
func (s *CompanyService) IsAdmin(userID string) bool {
	if s == nil || s.Repo == nil {
		return false
	}
	role, err := s.Repo.GetRole(context.Background(), userID)
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
	if s == nil || s.Repo == nil {
		return false, nil
	}
	role, err := s.Repo.GetRole(ctx, userID)
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
	proj, err := s.Repo.GetCompanyProjectionByUserID(ctx, userID)
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
	projs, err := s.Repo.ListCompanyProjections(ctx)
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

// Helper: check if student is registered and approved (used by oauth flow).
// Returns (isRegistered, role).
func isStudentRegisteredAndRole(db *gorm.DB, user model.User) (bool, string, error) {
	var count int64
	if err := db.Model(&model.Student{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		// helper.Role is a named type; convert to string when returning from this function.
		return false, string(helper.Viewer), err
	}
	if count == 0 {
		return false, string(helper.Viewer), nil
	}
	var student model.Student
	if err := db.Model(&student).Where("user_id = ?", user.ID).First(&student).Error; err != nil {
		return true, string(helper.Viewer), err
	}
	if student.ApprovalStatus == model.StudentApprovalAccepted {
		return true, string(helper.Student), nil
	}
	return true, string(helper.Viewer), nil
}

// small helper to compute cookie max age in seconds
func CookieMaxAge() int {
	return int(time.Hour * 24 * 30 / time.Second)
}
