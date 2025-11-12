package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"
	"ku-work/backend/services/infraService"
)

// JobService encapsulates all database operations related to jobs.
type JobService struct {
	jobRepo  repo.JobRepository
	eventBus *infraService.EventBus
}

// NewJobService creates a new JobService instance wired with a JobRepository and EventBus.
func NewJobService(jobRepo repo.JobRepository, eventBus *infraService.EventBus) *JobService {
	if jobRepo == nil {
		log.Fatal("job repository cannot be nil")
	}
	// EventBus is optional; when nil, notification/AI events are simply skipped.
	return &JobService{jobRepo: jobRepo, eventBus: eventBus}
}

// FetchJobsParams contains the filtering & pagination options used when fetching jobs.
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
	// Role & UserID are used to alter query behavior depending on who is requesting
	Role   helper.Role
	UserID string
}

// JobResponse mirrors the fields returned to API consumers for a job.
// This struct is intentionally similar to handlers.JobResponse so that handlers
// can easily return the data from the service.
type JobResponse struct {
	ID                  uint      `json:"id" gorm:"column:id"`
	CreatedAt           time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt           time.Time `json:"updatedAt" gorm:"column:updated_at"`
	Name                string    `json:"name" gorm:"column:name"`
	CompanyID           string    `json:"companyId" gorm:"column:company_id"`
	PhotoID             string    `json:"photoId" gorm:"column:photo_id"`
	BannerID            string    `json:"bannerId" gorm:"column:banner_id"`
	CompanyName         string    `json:"companyName" gorm:"column:company_name"`
	Position            string    `json:"position" gorm:"column:position"`
	Duration            string    `json:"duration" gorm:"column:duration"`
	Description         string    `json:"description" gorm:"column:description"`
	Location            string    `json:"location" gorm:"column:location"`
	JobType             string    `json:"jobType" gorm:"column:job_type"`
	Experience          string    `json:"experience" gorm:"column:experience"`
	MinSalary           uint      `json:"minSalary" gorm:"column:min_salary"`
	MaxSalary           uint      `json:"maxSalary" gorm:"column:max_salary"`
	ApprovalStatus      string    `json:"approvalStatus" gorm:"column:approval_status"`
	IsOpen              bool      `json:"open" gorm:"column:is_open"`
	NotifyOnApplication bool      `json:"notifyOnApplication" gorm:"column:notify_on_application"`
}

// JobWithStatsResponse extends JobResponse with application statistics.
type JobWithStatsResponse struct {
	JobResponse
	Pending  int64 `json:"pending"`
	Accepted int64 `json:"accepted"`
	Rejected int64 `json:"rejected"`
}

// CreateJob inserts a new job record via repository.
func (s *JobService) CreateJob(ctx context.Context, job *model.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	err := s.jobRepo.CreateJob(ctx, job)
	s.eventBus.PublishAIJobCheck(job.ID)
	return err
}

// FindJobByID retrieves a job by ID via repository.
func (s *JobService) FindJobByID(ctx context.Context, id uint) (*model.Job, error) {
	return s.jobRepo.FindJobByID(ctx, id)
}

// UpdateJob persists changes to a job via repository.
func (s *JobService) UpdateJob(ctx context.Context, job *model.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	return s.jobRepo.UpdateJob(ctx, job)
}

// ApproveOrRejectJob updates job approval status and records an audit entry via repository.
// It also publishes an email event via EventBus (if configured).
func (s *JobService) ApproveOrRejectJob(ctx context.Context, jobID uint, approve bool, actorID, reason string) error {
	// delegate persistence to repository (this creates audit as well)
	if err := s.jobRepo.ApproveOrRejectJob(ctx, jobID, approve, actorID, reason); err != nil {
		return err
	}

	// Publish email via EventBus (best-effort)
	if s.eventBus != nil {
		// Fetch denormalized job detail (includes company id and company name)
		jobDetail, err := s.jobRepo.GetJobDetail(ctx, jobID)
		if err == nil && jobDetail != nil {
			// Fetch company record to get email address
			company, err := s.jobRepo.FindCompanyByUserID(ctx, jobDetail.CompanyID)
			if err == nil && company != nil && company.Email != "" {
				status := string(model.JobApprovalRejected)
				if approve {
					status = string(model.JobApprovalAccepted)
				}
				ev := infraService.EmailJobApprovalEvent{
					CompanyEmail:    company.Email,
					CompanyUsername: jobDetail.CompanyName,
					JobName:         jobDetail.Name,
					JobPosition:     jobDetail.Position,
					Status:          status,
					Reason:          reason,
				}
				_ = s.eventBus.PublishEmailJobApproval(ev)
			}
		}
	}

	return nil
}

// CountJobs returns the number of jobs matching the provided criteria by delegating to repository.FetchJobs
// and returning the total count.
func (s *JobService) CountJobs(ctx context.Context, params *FetchJobsParams) (int64, error) {
	// Convert to repository params
	rp := repo.FetchJobsParams{
		Limit:          params.Limit,
		Offset:         params.Offset,
		Location:       params.Location,
		Keyword:        params.Keyword,
		JobType:        params.JobType,
		Experience:     params.Experience,
		MinSalary:      params.MinSalary,
		MaxSalary:      params.MaxSalary,
		Open:           params.Open,
		CompanyID:      params.CompanyID,
		JobID:          params.JobID,
		ApprovalStatus: params.ApprovalStatus,
		Role:           params.Role,
		UserID:         params.UserID,
	}
	_, total, err := s.jobRepo.FetchJobs(ctx, &rp)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// FetchJobs returns a paginated list of jobs according to the provided parameters by delegating to repository.
// The repository returns generic projections; normalize those into service-level DTOs so handlers get stable shapes.
func (s *JobService) FetchJobs(ctx context.Context, params *FetchJobsParams) (any, int64, error) {
	rp := repo.FetchJobsParams{
		Limit:          params.Limit,
		Offset:         params.Offset,
		Location:       params.Location,
		Keyword:        params.Keyword,
		JobType:        params.JobType,
		Experience:     params.Experience,
		MinSalary:      params.MinSalary,
		MaxSalary:      params.MaxSalary,
		Open:           params.Open,
		CompanyID:      params.CompanyID,
		JobID:          params.JobID,
		ApprovalStatus: params.ApprovalStatus,
		Role:           params.Role,
		UserID:         params.UserID,
	}

	// Retrieve raw repository projection
	raw, total, err := s.jobRepo.FetchJobs(ctx, &rp)
	if err != nil {
		return nil, 0, err
	}

	// Normalize the repository result into service DTOs by marshaling then unmarshaling.
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal repository result: %w", err)
	}

	// If requester is a company, return jobs with stats (pending/accepted/rejected)
	if params.Role == helper.Company {
		var items []JobWithStatsResponse
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal company job projection: %w", err)
		}
		return items, total, nil
	}

	// Non-company users: return basic job responses
	var items []JobResponse
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal job projection: %w", err)
	}
	return items, total, nil
}

// GetJobDetail fetches the detailed view of a job via repository and maps to JobResponse.
func (s *JobService) GetJobDetail(ctx context.Context, jobID uint) (*JobResponse, error) {
	// The repository returns a denormalized JobDetail projection.
	// Map that projection into the service-level JobResponse.
	mjob, err := s.jobRepo.GetJobDetail(ctx, jobID)
	if err != nil {
		return nil, err
	}

	resp := &JobResponse{
		ID:                  mjob.ID,
		CreatedAt:           mjob.CreatedAt,
		UpdatedAt:           mjob.UpdatedAt,
		Name:                mjob.Name,
		CompanyID:           mjob.CompanyID,
		PhotoID:             mjob.PhotoID,
		BannerID:            mjob.BannerID,
		CompanyName:         mjob.CompanyName,
		Position:            mjob.Position,
		Duration:            mjob.Duration,
		Description:         mjob.Description,
		Location:            mjob.Location,
		JobType:             mjob.JobType,
		Experience:          mjob.Experience,
		MinSalary:           mjob.MinSalary,
		MaxSalary:           mjob.MaxSalary,
		ApprovalStatus:      mjob.ApprovalStatus,
		IsOpen:              mjob.IsOpen,
		NotifyOnApplication: mjob.NotifyOnApplication,
	}
	return resp, nil
}

// FindCompanyByUserID returns the Company record for the given user id via repository.
func (s *JobService) FindCompanyByUserID(ctx context.Context, userID string) (*model.Company, error) {
	return s.jobRepo.FindCompanyByUserID(ctx, userID)
}

// ResolveRole asks the underlying repository for the role of a given user.
// This keeps role resolution behind the service abstraction so handlers don't touch DB.
func (s *JobService) ResolveRole(ctx context.Context, userID string) (helper.Role, error) {
	// If the repository implementation exposes GetRole, delegate to it.
	type roleResolver interface {
		GetRole(ctx context.Context, userID string) (helper.Role, error)
	}
	if rr, ok := s.jobRepo.(roleResolver); ok {
		return rr.GetRole(ctx, userID)
	}
	// Otherwise return Unknown as a safe default.
	return helper.Unknown, nil
}

// AcceptOrRejectJobApplication updates a job application's status ensuring the company owns the job.
func (s *JobService) AcceptOrRejectJobApplication(ctx context.Context, userId string, appID uint, accept bool) error {
	return s.jobRepo.AcceptOrRejectJobApplication(ctx, userId, appID, accept)
}
