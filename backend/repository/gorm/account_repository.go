package gormrepo

import (
	"context"
	"time"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

// GormAccountRepository is a GORM-backed implementation of repository.AccountRepository.
type GormAccountRepository struct {
	db *gorm.DB
}

// NewGormAccountRepository constructs a new AccountRepository backed by GORM.
func NewGormAccountRepository(db *gorm.DB) repo.AccountRepository {
	return &GormAccountRepository{db: db}
}

// WithTx returns a new repository instance bound to the provided transaction DB.
func (r *GormAccountRepository) WithTx(tx *gorm.DB) repo.AccountRepository {
	return &GormAccountRepository{db: tx}
}

// FindUserByID returns a user by id, optionally including soft-deleted records.
func (r *GormAccountRepository) FindUserByID(ctx context.Context, id string, includeSoftDeleted bool) (*model.User, error) {
	var user model.User
	q := r.db.WithContext(ctx).Model(&model.User{})
	if includeSoftDeleted {
		q = q.Unscoped()
	}
	if err := q.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// SoftDeleteUserByID performs a soft-delete on the user.
func (r *GormAccountRepository) SoftDeleteUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

// RestoreUserByID clears deleted_at on the user (reactivation).
func (r *GormAccountRepository) RestoreUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// IsUserDeactivated reports whether the user is soft-deleted.
func (r *GormAccountRepository) IsUserDeactivated(ctx context.Context, id string) (bool, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Select("deleted_at").
		Where("id = ?", id).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return user.DeletedAt.Valid, nil
}

// ExistsUsername reports whether any user exists with the given username.
func (r *GormAccountRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUserFields performs partial updates on the user record identified by userID.
func (r *GormAccountRepository) UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

// FindCompanyByUserID loads a company by owner user id.
func (r *GormAccountRepository) FindCompanyByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Company, error) {
	var company model.Company
	q := r.db.WithContext(ctx).Model(&model.Company{})
	if includeSoftDeleted {
		q = q.Unscoped()
	}
	if err := q.Where("user_id = ?", userID).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// RestoreCompanyByUserID clears deleted_at on the company record.
func (r *GormAccountRepository) RestoreCompanyByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

// DisableCompanyJobPosts sets is_open = false for jobs owned by the company user.
func (r *GormAccountRepository) DisableCompanyJobPosts(ctx context.Context, companyUserID string) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Job{}).
		Where("company_id = ? AND is_open = ?", companyUserID, true).
		Update("is_open", false)
	return res.RowsAffected, res.Error
}

// FindStudentByUserID loads a student by owner user id.
func (r *GormAccountRepository) FindStudentByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Student, error) {
	var student model.Student
	q := r.db.WithContext(ctx).Model(&model.Student{})
	if includeSoftDeleted {
		q = q.Unscoped()
	}
	if err := q.Where("user_id = ?", userID).First(&student).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

// RestoreStudentByUserID clears deleted_at on the student record.
func (r *GormAccountRepository) RestoreStudentByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

// FindGoogleOAuthByUserID loads Google OAuth details by user id.
func (r *GormAccountRepository) FindGoogleOAuthByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.GoogleOAuthDetails, error) {
	var oauth model.GoogleOAuthDetails
	q := r.db.WithContext(ctx).Model(&model.GoogleOAuthDetails{})
	if includeSoftDeleted {
		q = q.Unscoped()
	}
	if err := q.Where("user_id = ?", userID).First(&oauth).Error; err != nil {
		return nil, err
	}
	return &oauth, nil
}

// RestoreGoogleOAuthByUserID clears deleted_at on the oauth record.
func (r *GormAccountRepository) RestoreGoogleOAuthByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

// ListSoftDeletedUsersBefore returns users soft-deleted before cutoff.
func (r *GormAccountRepository) ListSoftDeletedUsersBefore(ctx context.Context, cutoff time.Time) ([]model.User, error) {
	var users []model.User
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("deleted_at IS NOT NULL").
		Where("deleted_at < ?", cutoff).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserAnonymized updates PII on the user record.
func (r *GormAccountRepository) UpdateUserAnonymized(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

// UpdateStudentFields performs partial updates on the student record identified by userID.
func (r *GormAccountRepository) UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

// UpdateCompanyFields performs partial updates on the company record identified by userID.
func (r *GormAccountRepository) UpdateCompanyFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

// UpdateOAuthFields performs partial updates on the oauth record identified by userID.
func (r *GormAccountRepository) UpdateOAuthFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

// ListJobApplicationsWithFilesByUserID returns all job applications for the given user with Files preloaded.
func (r *GormAccountRepository) ListJobApplicationsWithFilesByUserID(ctx context.Context, userID string) ([]model.JobApplication, error) {
	var apps []model.JobApplication
	if err := r.db.WithContext(ctx).
		Unscoped().
		Preload("Files").
		Where("user_id = ?", userID).
		Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

// UpdateJobApplicationFields performs partial updates on a job application identified by (jobID, userID).
func (r *GormAccountRepository) UpdateJobApplicationFields(ctx context.Context, jobID uint, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Unscoped().
		Where("job_id = ? AND user_id = ?", jobID, userID).
		Updates(fields).Error
}

// FindFileByID returns a file record by its ID.
func (r *GormAccountRepository) FindFileByID(ctx context.Context, fileID string) (*model.File, error) {
	var f model.File
	if err := r.db.WithContext(ctx).
		Model(&model.File{}).
		Where("id = ?", fileID).
		First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// UnscopedDeleteFileRecord permanently deletes a file record by id.
func (r *GormAccountRepository) UnscopedDeleteFileRecord(ctx context.Context, fileID string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", fileID).
		Delete(&model.File{}).Error
}
