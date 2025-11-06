package gormrepo

import (
	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// BeginTx starts a new transaction
func (r *GormUserRepository) BeginTx() (repo.UserRepository, error) {
	return NewGormUserRepository(r.db.Begin()), nil
}

// CommitTx commits the current transaction
func (r *GormUserRepository) CommitTx() error {
	return r.db.Commit().Error
}

// RollbackTx rolls back the current transaction
func (r *GormUserRepository) RollbackTx() error {
	return r.db.Rollback().Error
}

func (r *GormUserRepository) ExistsByUsernameAndType(username, userType string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormUserRepository) CreateUser(user *model.User) error {
	if user == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(user).Error
}

func (r *GormUserRepository) CreateCompany(company *model.Company) error {
	if company == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(company).Error
}

func (r *GormUserRepository) FindUserByUsernameAndType(username, userType string) (*model.User, error) {
	var user model.User
	if err := r.db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FirstOrCreateUser(out *model.User, cond model.User) error {
	return r.db.FirstOrCreate(out, cond).Error
}

func (r *GormUserRepository) FindUserByID(id string) (*model.User, error) {
	var user model.User
	if err := r.db.Model(&model.User{}).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) CreateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(details).Error
}

func (r *GormUserRepository) UpdateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Model(&model.GoogleOAuthDetails{}).Where("user_id = ?", details.UserID).Updates(details).Error
}

func (r *GormUserRepository) GetGoogleOAuthDetailsByExternalID(externalID string) (*model.GoogleOAuthDetails, error) {
	var det model.GoogleOAuthDetails
	if err := r.db.Model(&model.GoogleOAuthDetails{}).
		Where("external_id = ?", externalID).
		First(&det).Error; err != nil {
		return nil, err
	}
	return &det, nil
}

func (r *GormUserRepository) CountCompanyByUserID(userID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Company{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormUserRepository) CountAdminByUserID(userID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Admin{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormUserRepository) IsStudentRegisteredAndRole(user model.User) (bool, string, error) {
	var count int64
	if err := r.db.Model(&model.Student{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		// helper.Role is a named type; convert to string when returning from this function.
		return false, string(helper.Viewer), err
	}
	if count == 0 {
		return false, string(helper.Viewer), nil
	}
	var student model.Student
	if err := r.db.Model(&student).Where("user_id = ?", user.ID).First(&student).Error; err != nil {
		return true, string(helper.Viewer), err
	}
	if student.ApprovalStatus == model.StudentApprovalAccepted {
		return true, string(helper.Student), nil
	}
	return true, string(helper.Viewer), nil
}
