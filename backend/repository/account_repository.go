package repository

import (
	"context"
	"time"

	"ku-work/backend/model"

	"gorm.io/gorm"
)

// AccountRepository defines database operations required by account-related services.
//
// This interface exists to enable a layered architecture with dependency injection:
// - Handlers invoke service layer methods.
// - Services depend on this repository abstraction.
// - Only repository implementations are allowed to access the database directly.
//
// Provide a GORM-backed implementation under repository/gorm to wire this into the app.
type AccountRepository interface {
	// WithTx returns a repository instance bound to the provided transaction DB.
	// This allows services to coordinate multiple repository operations atomically.
	WithTx(tx *gorm.DB) AccountRepository

	// ---------------------------
	// User operations
	// ---------------------------

	// FindUserByID returns a user by id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindUserByID(ctx context.Context, id string, includeSoftDeleted bool) (*model.User, error)

	// SoftDeleteUserByID performs a soft delete on the user (sets deleted_at).
	SoftDeleteUserByID(ctx context.Context, id string) error

	// RestoreUserByID clears deleted_at on the user (reactivation).
	RestoreUserByID(ctx context.Context, id string) error

	// IsUserDeactivated reports whether the user's account is soft-deleted.
	IsUserDeactivated(ctx context.Context, id string) (bool, error)

	// ExistsUsername reports whether any user exists with the given username.
	ExistsUsername(ctx context.Context, username string) (bool, error)

	// UpdateUserFields performs partial updates on the user record identified by userID.
	UpdateUserFields(ctx context.Context, userID string, fields map[string]any) error

	// ---------------------------
	// Company operations
	// ---------------------------

	// FindCompanyByUserID loads a company by owner user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindCompanyByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Company, error)

	// RestoreCompanyByUserID clears deleted_at on the company record, if present.
	RestoreCompanyByUserID(ctx context.Context, userID string) error

	// DisableCompanyJobPosts sets is_open = false for all jobs belonging to the company.
	// Returns the number of affected rows.
	DisableCompanyJobPosts(ctx context.Context, companyUserID string) (int64, error)

	// ---------------------------
	// Student operations
	// ---------------------------

	// FindStudentByUserID loads a student by owner user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindStudentByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.Student, error)

	// RestoreStudentByUserID clears deleted_at on the student record, if present.
	RestoreStudentByUserID(ctx context.Context, userID string) error

	// ---------------------------
	// OAuth operations
	// ---------------------------

	// FindGoogleOAuthByUserID loads Google OAuth details by user id. If includeSoftDeleted is true, soft-deleted rows are eligible.
	FindGoogleOAuthByUserID(ctx context.Context, userID string, includeSoftDeleted bool) (*model.GoogleOAuthDetails, error)

	// RestoreGoogleOAuthByUserID clears deleted_at on the oauth record, if present.
	RestoreGoogleOAuthByUserID(ctx context.Context, userID string) error

	// ---------------------------
	// Anonymization helpers
	// ---------------------------

	// ListSoftDeletedUsersBefore returns users soft-deleted before the cutoff time.
	ListSoftDeletedUsersBefore(ctx context.Context, cutoff time.Time) ([]model.User, error)

	// UpdateUserAnonymized updates PII on the user record (e.g., username, password_hash).
	// Callers should limit fields to non-sensitive, intended columns.
	UpdateUserAnonymized(ctx context.Context, userID string, fields map[string]any) error

	// UpdateStudentFields performs partial updates on the student record identified by userID.
	UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error

	// UpdateCompanyFields performs partial updates on the company record identified by userID.
	UpdateCompanyFields(ctx context.Context, userID string, fields map[string]any) error

	// UpdateOAuthFields performs partial updates on the oauth record identified by userID.
	UpdateOAuthFields(ctx context.Context, userID string, fields map[string]any) error

	// ---------------------------
	// Files and applications
	// ---------------------------

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
