package gormrepo

import (
	"context"
	"time"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormIdentityRepository struct {
	db   *gorm.DB
	inTx bool
}

// NewGormIdentityRepository constructs a new IdentityRepository backed by GORM.
func NewGormIdentityRepository(db *gorm.DB) repo.IdentityRepository {
	return &GormIdentityRepository{db: db, inTx: false}
}

// WithTx returns a repository instance bound to the provided transaction DB.
func (r *GormIdentityRepository) WithTx(tx *gorm.DB) repo.IdentityRepository {
	return &GormIdentityRepository{db: tx, inTx: true}
}

// BeginTx starts a new transaction and returns a repository bound to it.
func (r *GormIdentityRepository) BeginTx() (repo.IdentityRepository, error) {
	tx := r.db.Begin()
	return &GormIdentityRepository{db: tx, inTx: true}, nil
}

// CommitTx commits the current transaction.
func (r *GormIdentityRepository) CommitTx() error {
	return r.db.Commit().Error
}

// RollbackTx rolls back the current transaction.
func (r *GormIdentityRepository) RollbackTx() error {
	return r.db.Rollback().Error
}

func (r *GormIdentityRepository) FindUserByID(ctx context.Context, id string, includeSoftDeleted bool) (*model.User, error) {
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

func (r *GormIdentityRepository) ExistsUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormIdentityRepository) ExistsByUsernameAndType(username, userType string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormIdentityRepository) CreateUser(user *model.User) error {
	if user == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(user).Error
}

func (r *GormIdentityRepository) FindUserByUsernameAndType(username, userType string) (*model.User, error) {
	var user model.User
	if err := r.db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormIdentityRepository) FirstOrCreateUser(out *model.User, cond model.User) error {
	return r.db.FirstOrCreate(out, cond).Error
}

func (r *GormIdentityRepository) UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) SoftDeleteUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.User{}).Error
}

func (r *GormIdentityRepository) RestoreUserByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *GormIdentityRepository) IsUserDeactivated(ctx context.Context, id string) (bool, error) {
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

func (r *GormIdentityRepository) UpdateUserAnonymized(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) CreateCompany(company *model.Company) error {
	if company == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(company).Error
}

func (r *GormIdentityRepository) FindCompanyByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Company, error) {
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

func (r *GormIdentityRepository) RestoreCompanyByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

func (r *GormIdentityRepository) UpdateCompanyFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Company{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) DisableCompanyJobPosts(ctx context.Context, companyUserID string) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Job{}).
		Where("company_id = ? AND is_open = ?", companyUserID, true).
		Update("is_open", false)
	return res.RowsAffected, res.Error
}

func (r *GormIdentityRepository) CountCompanyByUserID(userID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Company{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormIdentityRepository) FindStudentByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Student, error) {
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

func (r *GormIdentityRepository) RestoreStudentByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

func (r *GormIdentityRepository) UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) IsStudentRegisteredAndRole(user model.User) (bool, string, error) {
	var count int64
	if err := r.db.Model(&model.Student{}).
		Where("user_id = ?", user.ID).
		Count(&count).Error; err != nil {
		return false, string("viewer"), err
	}
	if count == 0 {
		return false, string("viewer"), nil
	}
	var student model.Student
	if err := r.db.Model(&student).
		Where("user_id = ?", user.ID).
		First(&student).Error; err != nil {
		return true, string("viewer"), err
	}
	if student.ApprovalStatus == model.StudentApprovalAccepted {
		return true, string("student"), nil
	}
	return true, string("viewer"), nil
}

func (r *GormIdentityRepository) FindGoogleOAuthByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.GoogleOAuthDetails, error) {
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

func (r *GormIdentityRepository) RestoreGoogleOAuthByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Update("deleted_at", nil).Error
}

func (r *GormIdentityRepository) UpdateOAuthFields(ctx context.Context, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.GoogleOAuthDetails{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) CreateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(details).Error
}

func (r *GormIdentityRepository) UpdateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Model(&model.GoogleOAuthDetails{}).
		Where("user_id = ?", details.UserID).
		Updates(details).Error
}

func (r *GormIdentityRepository) GetGoogleOAuthDetailsByExternalID(externalID string) (*model.GoogleOAuthDetails, error) {
	var det model.GoogleOAuthDetails
	if err := r.db.Model(&model.GoogleOAuthDetails{}).
		Where("external_id = ?", externalID).
		First(&det).Error; err != nil {
		return nil, err
	}
	return &det, nil
}

func (r *GormIdentityRepository) CountAdminByUserID(userID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Admin{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormIdentityRepository) ListSoftDeletedUsersBefore(ctx context.Context, cutoff time.Time) ([]model.User, error) {
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

func (r *GormIdentityRepository) ListJobApplicationsWithFilesByUserID(ctx context.Context, userID string) ([]model.JobApplication, error) {
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

func (r *GormIdentityRepository) UpdateJobApplicationFields(ctx context.Context, jobID uint, userID string, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Unscoped().
		Where("job_id = ? AND user_id = ?", jobID, userID).
		Updates(fields).Error
}

func (r *GormIdentityRepository) UnscopedDeleteFileRecord(ctx context.Context, fileID string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", fileID).
		Delete(&model.File{}).Error
}
