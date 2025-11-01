package repository

import (
	"ku-work/backend/model"

	"gorm.io/gorm"
)

// UserRepository defines database operations related to users and user-related entities.
type UserRepository interface {
	// ExistsByUsernameAndType checks whether a user exists for the given username and type.
	ExistsByUsernameAndType(db *gorm.DB, username, userType string) (bool, error)

	// CreateUser persists a new user using the provided DB (tx or regular DB).
	CreateUser(db *gorm.DB, user *model.User) error

	// FindUserByUsernameAndType returns a user by username and type.
	FindUserByUsernameAndType(db *gorm.DB, username, userType string) (*model.User, error)

	// FirstOrCreateUser attempts to find a user matching cond; if none exists it creates out.
	FirstOrCreateUser(db *gorm.DB, out *model.User, cond model.User) error

	// FindUserByID returns a user by its ID.
	FindUserByID(db *gorm.DB, id string) (*model.User, error)

	// Google OAuth related operations
	CreateGoogleOAuthDetails(db *gorm.DB, details *model.GoogleOAuthDetails) error
	UpdateGoogleOAuthDetails(db *gorm.DB, details *model.GoogleOAuthDetails) error
	GetGoogleOAuthDetailsByExternalID(db *gorm.DB, externalID string) (*model.GoogleOAuthDetails, error)

	// Convenience counters used by auth flows
	CountCompanyByUserID(db *gorm.DB, userID string) (int64, error)
	CountAdminByUserID(db *gorm.DB, userID string) (int64, error)
}