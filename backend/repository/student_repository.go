package repository

import (
	"context"

	"ku-work/backend/model"
)

// StudentProfile is a denormalized projection of a student joined with OAuth details.
// It mirrors the fields returned by the original handlers layer without depending on handlers.
type StudentProfile struct {
	model.Student

	FirstName string `json:"firstName" gorm:"column:first_name"`
	LastName  string `json:"lastName" gorm:"column:last_name"`
	Email     string `json:"email" gorm:"column:email"`
	FullName  string `json:"fullName" gorm:"column:fullname"`
}

// StudentListFilter describes server-side filtering, sorting and pagination options
// for listing student profiles.
type StudentListFilter struct {
	// Pagination
	Offset int
	Limit  int

	// Optional filters
	ApprovalStatus *model.StudentApprovalStatus // nil means no filter

	// Sorting option:
	// - "latest"  -> created_at DESC
	// - "oldest"  -> created_at ASC
	// - "name_az" -> fullname ASC
	// - "name_za" -> fullname DESC
	SortBy string
}

// StudentRepository defines database operations related to students.
type StudentRepository interface {
	// ExistsByUserID reports whether a student row already exists for the given user id.
	ExistsByUserID(ctx context.Context, userID string) (bool, error)

	// CreateStudent inserts a new student row.
	CreateStudent(ctx context.Context, s *model.Student) error

	// FindStudentByUserID returns the raw student model by owner user id.
	FindStudentByUserID(ctx context.Context, userID string) (*model.Student, error)

	// UpdateStudentFields performs partial updates on the student record identified by userID.
	// Only updates the provided fields.
	UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error

	// FindStudentProfileByUserID returns a denormalized student profile (student + oauth fields).
	FindStudentProfileByUserID(ctx context.Context, userID string) (*StudentProfile, error)

	// ListStudentProfiles returns a list of denormalized student profiles using the given filter.
	ListStudentProfiles(ctx context.Context, filter StudentListFilter) ([]StudentProfile, error)

	// ApproveOrRejectStudent updates the student's approval status and records an audit entry.
	ApproveOrRejectStudent(ctx context.Context, userID string, approve bool, actorID, reason string) error
}
