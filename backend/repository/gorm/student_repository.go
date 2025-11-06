package gormrepo

import (
	"context"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

type GormStudentRepository struct {
	db *gorm.DB
}

func NewGormStudentRepository(db *gorm.DB) repo.StudentRepository {
	return &GormStudentRepository{db: db}
}

func (r *GormStudentRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormStudentRepository) CreateStudent(ctx context.Context, s *model.Student) error {
	if s == nil {
		return gorm.ErrInvalidData
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *GormStudentRepository) FindStudentByUserID(ctx context.Context, userID string) (*model.Student, error) {
	var s model.Student
	if err := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Where("user_id = ?", userID).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormStudentRepository) UpdateStudentFields(ctx context.Context, userID string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.Student{}).
		Unscoped().
		Where("user_id = ?", userID).
		Updates(fields).Error
}

func (r *GormStudentRepository) FindStudentProfileByUserID(ctx context.Context, userID string) (*repo.StudentProfile, error) {
	var out repo.StudentProfile
	q := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Select(`
			students.*,
			google_o_auth_details.first_name as first_name,
			google_o_auth_details.last_name as last_name,
			CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) as fullname,
			google_o_auth_details.email as email`,
		).
		Joins("INNER JOIN google_o_auth_details on google_o_auth_details.user_id = students.user_id").
		Where("students.user_id = ?", userID)

	if err := q.Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *GormStudentRepository) ListStudentProfiles(ctx context.Context, filter repo.StudentListFilter) ([]repo.StudentProfile, error) {
	q := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Select(`
			students.*,
			google_o_auth_details.first_name as first_name,
			google_o_auth_details.last_name as last_name,
			CONCAT(google_o_auth_details.first_name, ' ', google_o_auth_details.last_name) as fullname,
			google_o_auth_details.email as email`,
		).
		Joins("INNER JOIN google_o_auth_details on google_o_auth_details.user_id = students.user_id")

	if filter.ApprovalStatus != nil {
		q = q.Where("students.approval_status = ?", *filter.ApprovalStatus)
	}

	switch filter.SortBy {
	case "latest":
		q = q.Order("students.created_at DESC")
	case "oldest":
		q = q.Order("students.created_at ASC")
	case "name_az":
		q = q.Order("fullname ASC")
	case "name_za":
		q = q.Order("fullname DESC")
	}

	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}

	var items []repo.StudentProfile
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormStudentRepository) ApproveOrRejectStudent(ctx context.Context, userID string, approve bool, actorID, reason string) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	newStatus := model.StudentApprovalRejected
	if approve {
		newStatus = model.StudentApprovalAccepted
	}

	if err := tx.Model(&model.Student{}).
		Where("user_id = ?", userID).
		Update("approval_status", newStatus).Error; err != nil {
		tx.Rollback()
		return err
	}

	audit := model.Audit{
		ActorID:    actorID,
		Action:     string(newStatus),
		ObjectName: "Student",
		Reason:     reason,
		ObjectID:   userID,
	}
	if err := tx.Create(&audit).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
