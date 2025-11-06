package gormrepo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormJobRepository struct {
	db *gorm.DB
}

func NewGormJobRepository(db *gorm.DB) *GormJobRepository {
	return &GormJobRepository{db: db}
}

// CreateJob persists a new job record (populates ID and timestamps).
func (r *GormJobRepository) CreateJob(ctx context.Context, job *model.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	return r.db.WithContext(ctx).Create(job).Error
}

// FindJobByID retrieves a job by its ID.
func (r *GormJobRepository) FindJobByID(ctx context.Context, id uint) (*model.Job, error) {
	var job model.Job
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateJob persists modifications to a job.
func (r *GormJobRepository) UpdateJob(ctx context.Context, job *model.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	return r.db.WithContext(ctx).Save(job).Error
}

// ApproveOrRejectJob updates the approval status and creates an audit entry within a transaction.
func (r *GormJobRepository) ApproveOrRejectJob(ctx context.Context, jobID uint, approve bool, actorID, reason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job model.Job
		if err := tx.Take(&job, jobID).Error; err != nil {
			return err
		}

		if approve {
			job.ApprovalStatus = model.JobApprovalAccepted
		} else {
			job.ApprovalStatus = model.JobApprovalRejected
		}

		if err := tx.Save(&job).Error; err != nil {
			return err
		}

		audit := model.Audit{
			ActorID:    actorID,
			Action:     string(job.ApprovalStatus),
			Reason:     reason,
			ObjectName: "Job",
			ObjectID:   strconv.FormatUint(uint64(job.ID), 10),
		}
		if err := tx.Create(&audit).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetRole returns the Role for the given user id by delegating to helper.GetRole.
func (r *GormJobRepository) GetRole(ctx context.Context, userID string) (helper.Role, error) {
	role := helper.GetRole(userID, r.db)
	return role, nil
}

// internal projection for job response (non-company)
type jobResp struct {
	ID                  uint      `gorm:"column:id" json:"id"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updatedAt"`
	Name                string    `gorm:"column:name" json:"name"`
	CompanyID           string    `gorm:"column:company_id" json:"companyId"`
	PhotoID             string    `gorm:"column:photo_id" json:"photoId"`
	BannerID            string    `gorm:"column:banner_id" json:"bannerId"`
	CompanyName         string    `gorm:"column:company_name" json:"companyName"`
	Position            string    `gorm:"column:position" json:"position"`
	Duration            string    `gorm:"column:duration" json:"duration"`
	Description         string    `gorm:"column:description" json:"description"`
	Location            string    `gorm:"column:location" json:"location"`
	JobType             string    `gorm:"column:job_type" json:"jobType"`
	Experience          string    `gorm:"column:experience" json:"experience"`
	MinSalary           uint      `gorm:"column:min_salary" json:"minSalary"`
	MaxSalary           uint      `gorm:"column:max_salary" json:"maxSalary"`
	ApprovalStatus      string    `gorm:"column:approval_status" json:"approvalStatus"`
	IsOpen              bool      `gorm:"column:is_open" json:"open"`
	NotifyOnApplication bool      `gorm:"column:notify_on_application" json:"notifyOnApplication"`
}

// internal projection for company view (with stats)
type jobWithStatsResp struct {
	jobResp
	Pending  int64 `gorm:"column:pending" json:"pending"`
	Accepted int64 `gorm:"column:accepted" json:"accepted"`
	Rejected int64 `gorm:"column:rejected" json:"rejected"`
}

// applyFilters centralizes the query filters used by FetchJobs.
func (r *GormJobRepository) applyFilters(query *gorm.DB, params *repo.FetchJobsParams) *gorm.DB {
	// Optional id filter
	if params.JobID != nil {
		query = query.Where(&model.Job{ID: *params.JobID})
	}

	// Keyword search (split by whitespace)
	if params.Keyword != "" {
		words := strings.Fields(params.Keyword)
		for _, w := range words {
			pat := fmt.Sprintf("%%%s%%", w)
			searchGroup := r.db.Where("name ILIKE ?", pat).
				Or("description ILIKE ?", pat).
				Or("position ILIKE ?", pat).
				Or("duration ILIKE ?", pat).
				Or("users.username ILIKE ?", pat)
			query = query.Where(searchGroup)
		}
	}

	// Salary filtering
	query = query.Where("min_salary >= ?", params.MinSalary)
	query = query.Where("max_salary <= ?", params.MaxSalary)

	// Company and role-specific restrictions
	if params.Role == helper.Company {
		query = query.Where("company_id = ?", params.UserID)
	} else {
		if params.CompanyID != "" {
			query = query.Where("company_id = ?", params.CompanyID)
		}
	}

	// Open filter
	if (params.Role == helper.Company || params.Role == helper.Admin) && params.Open != nil {
		query = query.Where("is_open = ?", *params.Open)
	} else if params.Role == helper.Viewer || params.Role == helper.Student || params.Role == helper.Unknown {
		// Only open jobs for non-company viewers by default
		query = query.Where("is_open = ?", true)
	}

	// Location filter
	if params.Location != "" {
		query = query.Where("location ILIKE ?", params.Location)
	}

	// Job type
	if len(params.JobType) != 0 {
		query = query.Where("job_type IN ?", params.JobType)
	}

	// Experience
	if len(params.Experience) != 0 {
		query = query.Where("experience IN ?", params.Experience)
	}

	// Approval status
	if params.Role == helper.Admin || params.Role == helper.Company {
		if params.ApprovalStatus != nil && *params.ApprovalStatus != "" {
			query = query.Where("approval_status = ?", *params.ApprovalStatus)
		}
	} else {
		// Only accepted jobs for non-admin/company users
		query = query.Where(&model.Job{ApprovalStatus: model.JobApprovalAccepted})
	}

	return query
}

func (r *GormJobRepository) FetchJobs(ctx context.Context, params *repo.FetchJobsParams) (interface{}, int64, error) {
	// Base query with required joins to enrich job rows with company info (used only for filtering/counting)
	base := r.db.WithContext(ctx).Model(&model.Job{}).
		Joins("INNER JOIN users ON users.id = jobs.company_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id")

	// Apply filters
	filtered := r.applyFilters(base, params)

	// Get total count before pagination
	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and pluck job ids
	paged := filtered.Offset(int(params.Offset)).Limit(int(params.Limit))
	var ids []uint
	if err := paged.Pluck("jobs.id", &ids).Error; err != nil {
		return nil, 0, err
	}

	// If no rows, return empty typed slices
	if len(ids) == 0 {
		if params.Role == helper.Company {
			return []jobWithStatsResp{}, total, nil
		}
		return []jobResp{}, total, nil
	}

	// Map JobDetail -> jobResp
	mapDetailToResp := func(d *repo.JobDetail) jobResp {
		return jobResp{
			ID:                  d.ID,
			CreatedAt:           d.CreatedAt,
			UpdatedAt:           d.UpdatedAt,
			Name:                d.Name,
			CompanyID:           d.CompanyID,
			PhotoID:             d.PhotoID,
			BannerID:            d.BannerID,
			CompanyName:         d.CompanyName,
			Position:            d.Position,
			Duration:            d.Duration,
			Description:         d.Description,
			Location:            d.Location,
			JobType:             d.JobType,
			Experience:          d.Experience,
			MinSalary:           d.MinSalary,
			MaxSalary:           d.MaxSalary,
			ApprovalStatus:      d.ApprovalStatus,
			IsOpen:              d.IsOpen,
			NotifyOnApplication: d.NotifyOnApplication,
		}
	}

	// Fetch each detailed row (denormalized projection)
	var itemsResp []jobResp
	for _, id := range ids {
		detail, err := r.GetJobDetail(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		itemsResp = append(itemsResp, mapDetailToResp(detail))
	}

	// If company role, fetch aggregated application statistics grouped by job_id and merge
	if params.Role == helper.Company {
		type statRow struct {
			JobID    uint  `gorm:"column:job_id"`
			Pending  int64 `gorm:"column:pending"`
			Accepted int64 `gorm:"column:accepted"`
			Rejected int64 `gorm:"column:rejected"`
		}

		var stats []statRow
		if err := r.db.WithContext(ctx).Model(&model.JobApplication{}).
			Select("job_id, COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending, COUNT(CASE WHEN status = 'accepted' THEN 1 END) AS accepted, COUNT(CASE WHEN status = 'rejected' THEN 1 END) AS rejected").
			Where("job_id IN ?", ids).
			Group("job_id").
			Find(&stats).Error; err != nil {
			return nil, 0, err
		}

		statsMap := make(map[uint]statRow, len(stats))
		for _, s := range stats {
			statsMap[s.JobID] = s
		}

		out := make([]jobWithStatsResp, 0, len(itemsResp))
		for _, it := range itemsResp {
			s := statsMap[it.ID]
			out = append(out, jobWithStatsResp{
				jobResp:  it,
				Pending:  s.Pending,
				Accepted: s.Accepted,
				Rejected: s.Rejected,
			})
		}
		return out, total, nil
	}

	// Non-company users: return the simple projection
	return itemsResp, total, nil
}

// GetJobDetail retrieves a job record by ID and returns a denormalized projection (job + company fields).
func (r *GormJobRepository) GetJobDetail(ctx context.Context, jobID uint) (*repo.JobDetail, error) {
	var detail repo.JobDetail
	if err := r.db.WithContext(ctx).
		Table("jobs").
		Joins("INNER JOIN users ON users.id = jobs.company_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id").
		Select("jobs.id as id, jobs.created_at as created_at, jobs.name as name, jobs.company_id as company_id, "+
			"companies.photo_id as photo_id, companies.banner_id as banner_id, users.username as company_name, "+
			"jobs.position as position, jobs.duration as duration, jobs.description as description, jobs.location as location, "+
			"jobs.job_type as job_type, jobs.experience as experience, jobs.min_salary as min_salary, jobs.max_salary as max_salary, "+
			"jobs.approval_status as approval_status, jobs.is_open as is_open, jobs.notify_on_application as notify_on_application").
		Where("jobs.id = ?", jobID).
		First(&detail).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

// FindCompanyByUserID returns the Company record associated with the given user id.
func (r *GormJobRepository) FindCompanyByUserID(ctx context.Context, userID string) (*model.Company, error) {
	var company model.Company
	if err := r.db.WithContext(ctx).Where(&model.Company{UserID: userID}).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// AcceptOrRejectJobApplication updates a job application's status ensuring the company owns the job.
func (r *GormJobRepository) AcceptOrRejectJobApplication(ctx context.Context, userId string, appID uint, accept bool) error {
	jobApplication := model.JobApplication{}
	if err := r.db.WithContext(ctx).Model(&jobApplication).
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Where("jobs.company_id = ?", userId).
		Where("job_applications.id = ?", appID).
		Take(&jobApplication).Error; err != nil {
		return err
	}

	if accept {
		jobApplication.Status = model.JobApplicationAccepted
	} else {
		jobApplication.Status = model.JobApplicationRejected
	}

	return r.db.WithContext(ctx).Save(&jobApplication).Error
}
