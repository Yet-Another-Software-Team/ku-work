package repository

import (
	"ku-work/backend/model"
)

// UserRepository defines database operations related to users and user-related entities.
type UserRepository interface {
	// Transaction
	BeginTx() (UserRepository, error)
	CommitTx() error
	RollbackTx() error

	// ExistsByUsernameAndType checks whether a user exists for the given username and type.
	ExistsByUsernameAndType(username, userType string) (bool, error)

	// CreateUser persists a new user.
	CreateUser(user *model.User) error

	// CreateCompany persists a new company.
	CreateCompany(company *model.Company) error

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

	// isStudentRegisteredAndRole checks if a student is registered and their role.
	IsStudentRegisteredAndRole(user model.User) (bool, string, error)
}
