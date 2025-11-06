package gormrepo

import (
	"context"
	"time"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormAccountRepository struct {
	db *gorm.DB
}

func NewGormAccountRepository(db *gorm.DB) repo.AccountRepository {
	return &GormAccountRepository{db: db}
}

func (r *GormAccountRepository) WithTx(tx *gorm.DB) repo.AccountRepository {
	return &GormAccountRepository{db: tx}
}

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

func (r *GormAccountRepository) SoftDeleteUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *GormAccountRepository) RestoreUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

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

func (r *GormAccountRepository) UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

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

func (r *GormAccountRepository) RestoreCompanyByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

func (r *GormAccountRepository) DisableCompanyJobPosts(ctx context.Context, companyUserID string) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Job{}).
		Where("company_id = ? AND is_open = ?", companyUserID, true).
		Update("is_open", false)
	return res.RowsAffected, res.Error
}

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

func (r *GormAccountRepository) RestoreStudentByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

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

func (r *GormAccountRepository) RestoreGoogleOAuthByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

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

func (r *GormAccountRepository) UpdateUserAnonymized(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

func (r *GormAccountRepository) UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormAccountRepository) UpdateCompanyFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormAccountRepository) UpdateOAuthFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

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

func (r *GormAccountRepository) UpdateJobApplicationFields(ctx context.Context, jobID uint, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Unscoped().
		Where("job_id = ? AND user_id = ?", jobID, userID).
		Updates(fields).Error
}

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

func (r *GormAccountRepository) UnscopedDeleteFileRecord(ctx context.Context, fileID string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", fileID).
		Delete(&model.File{}).Error
}
