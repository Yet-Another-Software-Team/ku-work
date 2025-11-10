package gormrepo

import (
	"context"
	"fmt"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormApplicationRepository struct {
	db *gorm.DB
}

func NewGormApplicationRepository(db *gorm.DB) *GormApplicationRepository {
	return &GormApplicationRepository{db: db}
}

func (r *GormApplicationRepository) CreateApplication(ctx context.Context, app *model.JobApplication) error {
	if app == nil {
		return fmt.Errorf("application is nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Files.*").Create(app).Error; err != nil {
			return err
		}

		if len(app.Files) > 0 {
			if err := tx.Model(app).Association("Files").Append(app.Files); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormApplicationRepository) GetApplicationsForJob(ctx context.Context, jobID uint, params *repo.FetchJobApplicationsParams) ([]repo.ShortApplicationDetail, error) {
	if params == nil {
		params = &repo.FetchJobApplicationsParams{}
	}

	query := r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Joins("INNER JOIN users ON users.id = job_applications.user_id").
		Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.user_id = job_applications.user_id").
		Joins("INNER JOIN students ON students.user_id = job_applications.user_id").
		Select(`
			job_applications.*,
			CASE WHEN users.deleted_at IS NOT NULL THEN 'Deactivated User'
				 ELSE CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name)
			END AS username,
			CASE WHEN students.major IS NULL THEN 'Unknown' ELSE students.major END AS major,
			CASE WHEN students.student_id IS NULL THEN 'Unknown' ELSE students.student_id END AS student_id,
			job_applications.status AS status`).
		Where("job_applications.job_id = ?", jobID)

	if params.Status != nil && *params.Status != "" {
		query = query.Where("job_applications.status = ?", *params.Status)
	}

	switch params.SortBy {
	case "latest":
		query = query.Order("job_applications.created_at DESC")
	case "oldest":
		query = query.Order("job_applications.created_at ASC")
	case "name_az":
		query = query.Order("username ASC")
	case "name_za":
		query = query.Order("username DESC")
	default:
		query = query.Order("job_applications.created_at DESC")
	}

	if params.Offset > 0 {
		query = query.Offset(int(params.Offset))
	}
	if params.Limit > 0 {
		query = query.Limit(int(params.Limit))
	}

	var rows []repo.ShortApplicationDetail
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	for i := range rows {
		files, err := r.loadApplicationFiles(ctx, rows[i].JobID, rows[i].UserID)
		if err != nil {
			return nil, err
		}
		rows[i].Files = files
	}

	return rows, nil
}

func (r *GormApplicationRepository) ClearJobApplications(ctx context.Context, jobID uint, includePending, includeRejected, includeAccepted bool) (int64, error) {
	q := r.db.WithContext(ctx).Where("job_id = ?", jobID)

	if !includeAccepted {
		q = q.Not("status = ?", model.JobApplicationAccepted)
	}
	if !includeRejected {
		q = q.Not("status = ?", model.JobApplicationRejected)
	}
	if !includePending {
		q = q.Not("status = ?", model.JobApplicationPending)
	}

	res := q.Delete(&model.JobApplication{})
	return res.RowsAffected, res.Error
}

func (r *GormApplicationRepository) GetApplicationByJobAndEmail(ctx context.Context, jobID uint, email string) (*repo.FullApplicantDetail, error) {
	var out repo.FullApplicantDetail

	q := r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Joins("INNER JOIN users ON users.id = job_applications.user_id").
		Joins("INNER JOIN students ON students.user_id = job_applications.user_id").
		Joins("INNER JOIN google_o_auth_details ON google_o_auth_details.user_id = job_applications.user_id").
		Select(`
			job_applications.*,
		 	CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) AS username,
			users.username AS email,
			students.phone AS phone,
			students.photo_id AS photo_id,
			students.birth_date AS birth_date,
			students.about_me AS about_me,
			students.git_hub AS github,
			students.linked_in AS linked_in,
			students.student_id AS student_id,
			students.major AS major`).
		Where("job_applications.job_id = ? AND google_o_auth_details.email = ?", jobID, email).
		Where("users.deleted_at IS NULL").
		Where("users.username NOT LIKE 'ANON-%'")

	if err := q.First(&out).Error; err != nil {
		return nil, err
	}

	files, err := r.loadApplicationFiles(ctx, out.JobID, out.UserID)
	if err != nil {
		return nil, err
	}
	out.Files = files

	return &out, nil
}

func (r *GormApplicationRepository) GetAllApplicationsForUser(ctx context.Context, userID string, params *repo.FetchAllApplicationsParams) ([]repo.ApplicationWithJobDetails, int64, error) {
	if params == nil {
		params = &repo.FetchAllApplicationsParams{}
	}

	db := r.db.WithContext(ctx)

	isCompany := false
	{
		var count int64
		if err := db.Model(&model.Company{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			return nil, 0, err
		}
		isCompany = count > 0
	}

	base := db.Model(&model.JobApplication{}).
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id").
		Joins("INNER JOIN users ON users.id = companies.user_id").
		Joins("INNER JOIN users AS student_users ON student_users.id = job_applications.user_id").
		Select(`
			job_applications.*,
			jobs.position AS job_position,
			jobs.name AS job_name,
			jobs.job_type AS job_type,
			jobs.experience AS experience,
			jobs.min_salary AS min_salary,
			jobs.max_salary AS max_salary,
			jobs.is_open AS is_open,
			users.username AS company_name,
			companies.photo_id AS company_logo_id`)

	if isCompany {
		base = base.Where("jobs.company_id = ?", userID)
	} else {
		var studentCount int64
		if err := db.Model(&model.Student{}).Where("user_id = ?", userID).Count(&studentCount).Error; err != nil {
			return nil, 0, err
		}
		if studentCount == 0 {
			return []repo.ApplicationWithJobDetails{}, 0, nil
		}
		base = base.Where("job_applications.user_id = ?", userID)
	}

	if params.Status != nil && *params.Status != "" {
		base = base.Where("job_applications.status = ?", *params.Status)
	}

	switch params.SortBy {
	case "name":
		base = base.Order("users.username ASC")
	case "date-asc":
		base = base.Order("job_applications.created_at ASC")
	default:
		base = base.Order("job_applications.created_at DESC")
	}

	countQ := db.Model(&model.JobApplication{}).
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Joins("INNER JOIN companies ON companies.user_id = jobs.company_id").
		Joins("INNER JOIN users ON users.id = companies.user_id")
	if isCompany {
		countQ = countQ.Where("jobs.company_id = ?", userID)
	} else {
		countQ = countQ.Where("job_applications.user_id = ?", userID)
	}
	if params.Status != nil && *params.Status != "" {
		countQ = countQ.Where("job_applications.status = ?", *params.Status)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.Offset > 0 {
		base = base.Offset(int(params.Offset))
	}
	if params.Limit > 0 {
		base = base.Limit(int(params.Limit))
	}

	var rows []repo.ApplicationWithJobDetails
	if err := base.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	for i := range rows {
		files, err := r.loadApplicationFiles(ctx, rows[i].JobID, rows[i].UserID)
		if err != nil {
			return nil, 0, err
		}
		rows[i].Files = files
	}

	return rows, total, nil
}

func (r *GormApplicationRepository) GetApplication(ctx context.Context, jobID uint, studentUserID string) (*repo.ApplicationWithJobDetails, error) {
	var row repo.ApplicationWithJobDetails
	err := r.db.WithContext(ctx).
		Table("job_applications").
		Select("job_applications.*, jobs.title AS job_title").
		Joins("INNER JOIN jobs ON jobs.id = job_applications.job_id").
		Where("job_applications.job_id = ? AND job_applications.user_id = ?", jobID, studentUserID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}

	files, err := r.loadApplicationFiles(ctx, jobID, studentUserID)
	if err != nil {
		return nil, err
	}
	row.Files = files

	return &row, nil
}

func (r *GormApplicationRepository) UpdateApplicationStatus(ctx context.Context, jobID uint, studentUserID string, status model.JobApplicationStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.JobApplication{}).
		Where("job_id = ? AND user_id = ?", jobID, studentUserID).
		Clauses(clause.Returning{}).
		Update("status", status)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormApplicationRepository) loadApplicationFiles(ctx context.Context, jobID uint, userID string) ([]model.File, error) {
	var files []model.File
	err := r.db.WithContext(ctx).
		Table("job_application_has_file").
		Select("files.*").
		Joins("INNER JOIN files ON files.id = job_application_has_file.file_id").
		Where("job_application_has_file.job_application_job_id = ? AND job_application_has_file.job_application_user_id = ?", jobID, userID).
		Find(&files).Error
	return files, err
}
