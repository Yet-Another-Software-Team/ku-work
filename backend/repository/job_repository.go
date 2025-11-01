package repository

import (
	"context"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"
)

// FetchJobsParams contains filtering and pagination options.
type FetchJobsParams struct {
	Limit          uint
	Offset         uint
	Location       string
	Keyword        string
	JobType        []string
	Experience     []string
	MinSalary      uint
	MaxSalary      uint
	Open           *bool
	CompanyID      string
	JobID          *uint
	ApprovalStatus *string
	Role           helper.Role
	UserID         string
}

// JobDetail is a denormalized projection of a job record for service/handler use.
type JobDetail struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Name                string    `json:"name"`
	CompanyID           string    `json:"companyId"`
	PhotoID             string    `json:"photoId"`
	BannerID            string    `json:"bannerId"`
	CompanyName         string    `json:"companyName"`
	Position            string    `json:"position"`
	Duration            string    `json:"duration"`
	Description         string    `json:"description"`
	Location            string    `json:"location"`
	JobType             string    `json:"jobType"`
	Experience          string    `json:"experience"`
	MinSalary           uint      `json:"minSalary"`
	MaxSalary           uint      `json:"maxSalary"`
	ApprovalStatus      string    `json:"approvalStatus"`
	IsOpen              bool      `json:"open"`
	NotifyOnApplication bool      `json:"notifyOnApplication"`
}

// JobRepository defines persistence operations for jobs.
type JobRepository interface {
	CreateJob(ctx context.Context, job *model.Job) error
	FindJobByID(ctx context.Context, id uint) (*model.Job, error)
	UpdateJob(ctx context.Context, job *model.Job) error
	ApproveOrRejectJob(ctx context.Context, jobID uint, approve bool, actorID, reason string) error
	FetchJobs(ctx context.Context, params *FetchJobsParams) (any, int64, error)
	GetJobDetail(ctx context.Context, jobID uint) (*JobDetail, error)
	FindCompanyByUserID(ctx context.Context, userID string) (*model.Company, error)
	AcceptOrRejectJobApplication(ctx context.Context, userId string, appID uint, accept bool) error
	GetRole(ctx context.Context, userID string) (helper.Role, error)
}
