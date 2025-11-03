package gormrepo

import (
	"context"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

// GormCompanyRepository is a GORM-backed implementation of repo.CompanyRepository.
type GormCompanyRepository struct {
	db *gorm.DB
}

// NewGormCompanyRepository constructs a new CompanyRepository backed by GORM.
func NewGormCompanyRepository(db *gorm.DB) repo.CompanyRepository {
	return &GormCompanyRepository{db: db}
}

// FindCompanyByUserID returns the Company model for the provided user id.
func (r *GormCompanyRepository) FindCompanyByUserID(ctx context.Context, userID string) (*model.Company, error) {
	var company model.Company
	if err := r.db.WithContext(ctx).Where(&model.Company{UserID: userID}).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// anonymizeProjection removes PII from a CompanyProjection and sets the public name.
func anonymizeProjection(p *repo.CompanyProjection) {
	p.Email = ""
	p.Phone = ""
	p.PhotoID = ""
	p.BannerID = ""
	p.Address = ""
	p.City = ""
	p.Country = ""
	p.Website = ""
	p.AboutUs = ""
	p.Name = "Deactivated Account"
}

// GetCompanyProjectionByUserID returns a denormalized projection combining company fields.
func (r *GormCompanyRepository) GetCompanyProjectionByUserID(ctx context.Context, userID string) (*repo.CompanyProjection, error) {
	var proj repo.CompanyProjection
	if err := r.db.WithContext(ctx).Model(&model.Company{}).
		Select("companies.user_id, companies.created_at, companies.updated_at, "+
			"companies.email, companies.phone, companies.photo_id, companies.banner_id, "+
			"companies.address, companies.city, companies.country, companies.website, companies.about_us, "+
			"users.username as name").
		Joins("INNER JOIN users ON users.id = companies.user_id").
		Where("companies.user_id = ?", userID).
		First(&proj).Error; err != nil {
		return nil, err
	}

	deactivated, err := r.IsUserDeactivated(ctx, userID)
	if err != nil {
		return nil, err
	}
	if deactivated {
		anonymizeProjection(&proj)
	}

	return &proj, nil
}

// ListCompanyProjections returns denormalized projections for all companies (with owner's username).
// Any entries for deactivated accounts are anonymized before being returned.
func (r *GormCompanyRepository) ListCompanyProjections(ctx context.Context) ([]repo.CompanyProjection, error) {
	var out []repo.CompanyProjection
	if err := r.db.WithContext(ctx).Model(&model.Company{}).
		Select("companies.user_id, companies.created_at, companies.updated_at, " +
			"companies.email, companies.phone, companies.photo_id, companies.banner_id, " +
			"companies.address, companies.city, companies.country, companies.website, companies.about_us, " +
			"users.username as name").
		Joins("INNER JOIN users ON users.id = companies.user_id").
		Find(&out).Error; err != nil {
		return nil, err
	}

	// Gather deactivated user ids in one query to avoid N+1 checks.
	var deactivatedIDs []string
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("deleted_at IS NOT NULL").
		Pluck("id", &deactivatedIDs).Error; err != nil && err != gorm.ErrRecordNotFound {
		// If this query fails, return error.
		return nil, err
	}
	deactivatedSet := make(map[string]struct{}, len(deactivatedIDs))
	for _, id := range deactivatedIDs {
		deactivatedSet[id] = struct{}{}
	}

	for i := range out {
		if _, ok := deactivatedSet[out[i].UserID]; ok {
			anonymizeProjection(&out[i])
		}
	}

	return out, nil
}

// GetRole resolves the Role for the given user id using repository-backed checks.
func (r *GormCompanyRepository) GetRole(ctx context.Context, userID string) (helper.Role, error) {
	if userID == "" {
		return helper.Unknown, nil
	}

	var count int64

	// Admin
	if err := r.db.WithContext(ctx).Model(&model.Admin{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return helper.Unknown, err
	}
	if count > 0 {
		return helper.Admin, nil
	}

	// Company
	if err := r.db.WithContext(ctx).Model(&model.Company{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return helper.Unknown, err
	}
	if count > 0 {
		return helper.Company, nil
	}

	// Student (accepted)
	if err := r.db.WithContext(ctx).Model(&model.Student{}).
		Where("user_id = ? AND approval_status = ?", userID, model.StudentApprovalAccepted).
		Count(&count).Error; err != nil {
		return helper.Unknown, err
	}
	if count > 0 {
		return helper.Student, nil
	}

	// Google OAuth details -> Viewer
	if err := r.db.WithContext(ctx).Model(&model.GoogleOAuthDetails{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return helper.Unknown, err
	}
	if count > 0 {
		return helper.Viewer, nil
	}

	return helper.Unknown, nil
}

// IsUserDeactivated returns whether the specified user account is soft-deleted.
func (r *GormCompanyRepository) IsUserDeactivated(ctx context.Context, userID string) (bool, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("deleted_at").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return user.DeletedAt.Valid, nil
}
