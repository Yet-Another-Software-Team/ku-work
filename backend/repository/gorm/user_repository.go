package gormrepo

import (
	"ku-work/backend/model"

	"gorm.io/gorm"
)

// GormUserRepository is a GORM-backed implementation of UserRepository.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) ExistsByUsernameAndType(db *gorm.DB, username, userType string) (bool, error) {
	var count int64
	if err := db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormUserRepository) CreateUser(db *gorm.DB, user *model.User) error {
	if user == nil {
		return gorm.ErrInvalidData
	}
	return db.Create(user).Error
}

func (r *GormUserRepository) FindUserByUsernameAndType(db *gorm.DB, username, userType string) (*model.User, error) {
	var user model.User
	if err := db.Model(&model.User{}).
		Where("username = ? AND user_type = ?", username, userType).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FirstOrCreateUser(db *gorm.DB, out *model.User, cond model.User) error {
	return db.FirstOrCreate(out, cond).Error
}

func (r *GormUserRepository) FindUserByID(db *gorm.DB, id string) (*model.User, error) {
	var user model.User
	if err := db.Model(&model.User{}).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) CreateGoogleOAuthDetails(db *gorm.DB, details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return db.Create(details).Error
}

func (r *GormUserRepository) UpdateGoogleOAuthDetails(db *gorm.DB, details *model.GoogleOAuthDetails) error {
	if details == nil {
		return gorm.ErrInvalidData
	}
	return db.Model(&model.GoogleOAuthDetails{}).Where("user_id = ?", details.UserID).Updates(details).Error
}

func (r *GormUserRepository) GetGoogleOAuthDetailsByExternalID(db *gorm.DB, externalID string) (*model.GoogleOAuthDetails, error) {
	var det model.GoogleOAuthDetails
	if err := db.Model(&model.GoogleOAuthDetails{}).
		Where("external_id = ?", externalID).
		First(&det).Error; err != nil {
		return nil, err
	}
	return &det, nil
}

func (r *GormUserRepository) CountCompanyByUserID(db *gorm.DB, userID string) (int64, error) {
	var count int64
	if err := db.Model(&model.Company{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormUserRepository) CountAdminByUserID(db *gorm.DB, userID string) (int64, error) {
	var count int64
	if err := db.Model(&model.Admin{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
