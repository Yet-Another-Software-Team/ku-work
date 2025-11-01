package repository

import (
	"ku-work/backend/model"

	"gorm.io/gorm"
)

// UserRepository defines database operations related to users and user-related entities.
type UserRepository interface {
	// WithTx returns a repository instance bound to the provided transaction DB.
	// Use this when you need to run multiple repository operations within a single transaction.
	WithTx(tx *gorm.DB) UserRepository

	// ExistsByUsernameAndType checks whether a user exists for the given username and type.
	ExistsByUsernameAndType(username, userType string) (bool, error)

	// CreateUser persists a new user.
	CreateUser(user *model.User) error

	// FindUserByUsernameAndType returns a user by username and type.
	FindUserByUsernameAndType(username, userType string) (*model.User, error)

	// FirstOrCreateUser attempts to find a user matching cond; if none exists it creates out.
	FirstOrCreateUser(out *model.User, cond model.User) error

	// FindUserByID returns a user by its ID.
	FindUserByID(id string) (*model.User, error)

	// Google OAuth related operations
	CreateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error
	UpdateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error
	GetGoogleOAuthDetailsByExternalID(externalID string) (*model.GoogleOAuthDetails, error)

	// Convenience counters used by auth flows
	CountCompanyByUserID(userID string) (int64, error)
	CountAdminByUserID(userID string) (int64, error)
}
