package repository

import (
	"context"
	"time"

	"ku-work/backend/model"
)

// IdentityRepository is a unified repository abstraction for identity-related data.
type IdentityRepository interface {

	/*
	 * Transaction Contol
	 */

	BeginTx() (IdentityRepository, error)
	CommitTx() error
	RollbackTx() error

	/*
	 * User operations
	 */

	// FindUserByID returns a user by id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindUserByID(ctx context.Context, id string, includeSoftDeleted bool) (*model.User, error)

	// ExistsUsername reports whether any user exists with the given username.
	ExistsUsername(ctx context.Context, username string) (bool, error)

	// ExistsByUsernameAndType checks whether a user exists for the given username and type.
	ExistsByUsernameAndType(username, userType string) (bool, error)

	// CreateUser persists a new user.
	CreateUser(user *model.User) error

	// FindUserByUsernameAndType returns a user by username and type.
	FindUserByUsernameAndType(username, userType string) (*model.User, error)

	// FirstOrCreateUser attempts to find a user matching cond; if none exists it creates out.
	FirstOrCreateUser(out *model.User, cond model.User) error

	// UpdateUserFields performs partial updates on the user record identified by userID.
	UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error

	// SoftDeleteUserByID performs a soft delete on the user (sets deleted_at).
	SoftDeleteUserByID(ctx context.Context, id string) error

	// RestoreUserByID clears deleted_at on the user (reactivation).
	RestoreUserByID(ctx context.Context, id string) error

	// IsUserDeactivated reports whether the user's account is soft-deleted.
	IsUserDeactivated(ctx context.Context, id string) (bool, error)

	// UpdateUserAnonymized updates PII on the user record (e.g., username, password_hash).
	// Callers should limit fields to intended columns only.
	UpdateUserAnonymized(ctx context.Context, userID string, fields map[string]any) error

	/*
	 * Company operations
	 */

	// CreateCompany persists a new company profile.
	CreateCompany(company *model.Company) error

	// FindCompanyByUserID loads a company by owner user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindCompanyByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Company, error)

	// RestoreCompanyByUserID clears deleted_at on the company record, if present.
	RestoreCompanyByUserID(ctx context.Context, userID string) error

	// UpdateCompanyFields performs partial updates on the company record identified by userID.
	UpdateCompanyFields(ctx context.Context, userID string, fields map[string]any) error

	// DisableCompanyJobPosts sets is_open = false for all jobs belonging to the company.
	// Returns the number of affected rows.
	DisableCompanyJobPosts(ctx context.Context, companyUserID string) (int64, error)

	// CountCompanyByUserID returns number of company records for a user (used by auth/role checks).
	CountCompanyByUserID(userID string) (int64, error)

	/*
	 * Student operations
	 */

	// FindStudentByUserID loads a student by owner user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindStudentByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Student, error)

	// RestoreStudentByUserID clears deleted_at on the student record, if present.
	RestoreStudentByUserID(ctx context.Context, userID string) error

	// UpdateStudentFields performs partial updates on the student record identified by userID.
	UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error

	// IsStudentRegisteredAndRole checks if a student is registered and returns the effective role string.
	IsStudentRegisteredAndRole(user model.User) (bool, string, error)

	/*
	 * OAuth operations
	 */

	// FindGoogleOAuthByUserID loads Google OAuth details by user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindGoogleOAuthByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.GoogleOAuthDetails, error)

	// RestoreGoogleOAuthByUserID clears deleted_at on the oauth record, if present.
	RestoreGoogleOAuthByUserID(ctx context.Context, userID string) error

	// UpdateOAuthFields performs partial updates on the oauth record identified by userID.
	UpdateOAuthFields(ctx context.Context, userID string, fields map[string]any) error

	// CreateGoogleOAuthDetails persists new Google OAuth details.
	CreateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error

	// UpdateGoogleOAuthDetails updates Google OAuth details for a given user id contained in the struct.
	UpdateGoogleOAuthDetails(details *model.GoogleOAuthDetails) error

	// GetGoogleOAuthDetailsByExternalID resolves OAuth details by provider external id.
	GetGoogleOAuthDetailsByExternalID(externalID string) (*model.GoogleOAuthDetails, error)

	/*
	 * Admin operations
	 */

	// CountAdminByUserID returns number of admin records for a user (used by auth/role checks).
	CountAdminByUserID(userID string) (int64, error)

	/*
	 * User Anonymization operationss
	 */

	// ListSoftDeletedUsersBefore returns users soft-deleted before the cutoff time.
	ListSoftDeletedUsersBefore(ctx context.Context, cutoff time.Time) ([]model.User, error)

	// ListJobApplicationsWithFilesByUserID returns all job applications for the given user (student),
	// with Files preloaded for cleanup/anonymization flows.
	ListJobApplicationsWithFilesByUserID(ctx context.Context, userID string) ([]model.JobApplication, error)

	// UpdateJobApplicationFields performs partial updates on a job application identified by (jobID, userID).
	UpdateJobApplicationFields(ctx context.Context, jobID uint, userID string, fields map[string]any) error

	// FindFileByID returns a file record by id (used to invoke storage hooks before deletion).
	FindFileByID(ctx context.Context, fileID string) (*model.File, error)

	// UnscopedDeleteFileRecord permanently deletes a file record by id.
	UnscopedDeleteFileRecord(ctx context.Context, fileID string) error
}
