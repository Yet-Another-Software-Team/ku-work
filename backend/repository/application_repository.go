package repository

import (
	"context"
	"time"

	"ku-work/backend/model"
)

// FetchJobApplicationsParams holds filters/pagination for a job's applications.
type FetchJobApplicationsParams struct {
	// Optional: filter by status (pending, accepted, rejected)
	Status *string
	// Pagination parameters
	Offset uint
	Limit  uint
	// Sort order: latest, oldest, name_az, name_za
	SortBy string
}

// FetchAllApplicationsParams holds filters/pagination for user-visible applications.
type FetchAllApplicationsParams struct {
	// Optional: filter by status (pending, accepted, rejected)
	Status *string
	// Sort order: name, date-desc, date-asc
	SortBy string
	// Pagination parameters
	Offset uint
	Limit  uint
}

// ShortApplicationDetail projection for listing applications of a job.
type ShortApplicationDetail struct {
	// Base application fields
	CreatedAt    time.Time `json:"createdAt"`
	JobID        uint      `json:"jobId"`
	UserID       string    `json:"userId"`
	ContactPhone string    `json:"phone"`
	ContactEmail string    `json:"email"`
	Status       string    `json:"status"`

	// Denormalized applicant fields
	Username  string `json:"username"`
	Major     string `json:"major"`
	StudentID string `json:"studentId"`

	// Attached files for the application (if loaded by the repository implementation)
	Files []model.File `json:"files" gorm:"-"`
}

type ApplicationWithJobDetails struct {
	// Base application fields
	CreatedAt    time.Time `json:"createdAt"`
	JobID        uint      `json:"jobId"`
	UserID       string    `json:"userId"`
	ContactPhone string    `json:"phone"`
	ContactEmail string    `json:"email"`
	Status       string    `json:"status"`

	// Denormalized job/company fields
	JobPosition   string `json:"position"`
	JobName       string `json:"jobName"`
	CompanyName   string `json:"companyName"`
	CompanyLogoID string `json:"photoId"`
	JobType       string `json:"jobType"`
	Experience    string `json:"experience"`
	MinSalary     uint   `json:"minSalary"`
	MaxSalary     uint   `json:"maxSalary"`
	IsOpen        bool   `json:"isOpen"`

	// Attached files for the application (if loaded by the repository implementation)
	Files []model.File `json:"files" gorm:"-"`
}

// FullApplicantDetail projection for a single application with profile and files.
type FullApplicantDetail struct {
	// Base application fields
	CreatedAt    time.Time    `json:"createdAt"`
	JobID        uint         `json:"jobId"`
	UserID       string       `json:"userId"`
	ContactPhone string       `json:"phone"`
	ContactEmail string       `json:"email"`
	Status       string       `json:"status"`
	Files        []model.File `json:"files" gorm:"-"`

	// Applicant details
	Username  string    `json:"username"`
	PhotoID   *string   `json:"photoId"`
	BirthDate time.Time `json:"birthDate"`
	AboutMe   string    `json:"aboutMe"`
	GitHub    string    `json:"github"`
	LinkedIn  string    `json:"linkedIn"`
	StudentID string    `json:"studentId"`
	Major     string    `json:"major"`
}

// ApplicationRepository defines persistence operations for applications.
type ApplicationRepository interface {
	// CreateApplication persists a new application record. If app.Files contains
	// File records, the implementation should persist the many2many relations.
	CreateApplication(ctx context.Context, app *model.JobApplication) error

	// GetApplicationsForJob returns applications for a job id according to the provided filters.
	// Implementations should return denormalized applicant info and may also populate Files.
	GetApplicationsForJob(ctx context.Context, jobID uint, params *FetchJobApplicationsParams) ([]ShortApplicationDetail, error)

	// ClearJobApplications deletes applications for a job id. The boolean flags indicate which
	// statuses should be included in the deletion. The returned int64 is the number of rows affected.
	ClearJobApplications(ctx context.Context, jobID uint, includePending, includeRejected, includeAccepted bool) (int64, error)

	// GetApplicationByJobAndEmail returns a detailed projection for a single application identified
	// by job id and applicant email. Implementations should exclude deactivated/anonymized users.
	GetApplicationByJobAndEmail(ctx context.Context, jobID uint, email string) (*FullApplicantDetail, error)

	// GetAllApplicationsForUser returns applications for a given user. If the user is a company,
	// it returns applications for all jobs owned by the company; if the user is a student,
	// it returns applications submitted by that student. The second return value is the total count.
	// Implementations should include denormalized job/company info and may also populate Files.
	GetAllApplicationsForUser(ctx context.Context, userID string, params *FetchAllApplicationsParams) ([]ApplicationWithJobDetails, int64, error)

	// UpdateApplicationStatus sets the application status for the given (jobID, studentUserID) pair.
	UpdateApplicationStatus(ctx context.Context, jobID uint, studentUserID string, status model.JobApplicationStatus) error
}
