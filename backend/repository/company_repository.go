package repository

import (
	"context"
	"time"

	"ku-work/backend/model"
)

// CompanyProjection is a denormalized view of a company joined with the owner's username.
// This projection is intended for read-only views returned by repository methods, so
// services/handlers don't need to craft raw SQL joins themselves.
type CompanyProjection struct {
	UserID    string    `json:"userId" gorm:"column:user_id"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`

	Email    string `json:"email" gorm:"column:email"`
	Phone    string `json:"phone" gorm:"column:phone"`
	PhotoID  string `json:"photoId" gorm:"column:photo_id"`
	BannerID string `json:"bannerId" gorm:"column:banner_id"`
	Address  string `json:"address" gorm:"column:address"`
	City     string `json:"city" gorm:"column:city"`
	Country  string `json:"country" gorm:"column:country"`
	Website  string `json:"website" gorm:"column:website"`
	AboutUs  string `json:"about" gorm:"column:about_us"`

	// Owner's username (sourced from users.username)
	Name string `json:"name" gorm:"column:name"`
}

// CompanyRepository defines persistence operations related to company entities.
//
// Implementations (e.g. a GORM-backed repository) should live under the
// repository/gorm package and be the only part of the codebase that performs DB access.
type CompanyRepository interface {
	// FindCompanyByUserID returns the full Company model for the given user id.
	FindCompanyByUserID(ctx context.Context, userID string) (*model.Company, error)

	// GetCompanyProjectionByUserID returns the denormalized projection for a single company.
	GetCompanyProjectionByUserID(ctx context.Context, userID string) (*CompanyProjection, error)

	// ListCompanyProjections returns projections for all companies (owner username included).
	ListCompanyProjections(ctx context.Context) ([]CompanyProjection, error)

	// IsUserDeactivated returns whether the specified user account is soft-deleted.
	IsUserDeactivated(ctx context.Context, userID string) (bool, error)
}
