package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type ExperienceType string

const (
	ExperienceNewGrad    ExperienceType = "newgrad"
	ExperienceJunior     ExperienceType = "junior"
	ExperienceSenior     ExperienceType = "senior"
	ExperienceManager    ExperienceType = "manager"
	ExperienceInternship ExperienceType = "internship"
)

type JobType string

const (
	JobTypeFullTime   JobType = "fulltime"
	JobTypePartTime   JobType = "parttime"
	JobTypeContract   JobType = "contract"
	JobTypeCasual     JobType = "casual"
	JobTypeInternship JobType = "internship"
)

type JobApprovalStatus string

const (
	JobApprovalAccepted JobApprovalStatus = "accepted"
	JobApprovalRejected JobApprovalStatus = "rejected"
	JobApprovalPending  JobApprovalStatus = "pending"
)

type Job struct {
	ID                  uint              `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time         `json:"createdAt"`
	Name                string            `json:"name"`
	CompanyID           string            `gorm:"type:uuid" json:"companyId"`
	Company             Company           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company"`
	Position            string            `json:"position"`
	Duration            string            `json:"duration"`
	Description         string            `json:"description"`
	Location            string            `json:"location"`
	JobType             JobType           `json:"jobType"`
	Experience          ExperienceType    `json:"experienceType"`
	MinSalary           uint              `json:"minSalary"`
	MaxSalary           uint              `json:"maxSalary"`
	ApprovalStatus      JobApprovalStatus `json:"approvalStatus"`
	IsOpen              bool              `json:"open"`
	NotifyOnApplication bool              `json:"notifyOnApplication default:true"`
	JobApplications     []JobApplication  `gorm:"foreignkey:JobID;constraint:OnDelete:CASCADE;" json:"-"`
}

type JobApplicationStatus string

const (
	JobApplicationAccepted JobApplicationStatus = "accepted"
	JobApplicationRejected JobApplicationStatus = "rejected"
	JobApplicationPending  JobApplicationStatus = "pending"
)

type JobApplication struct {
	CreatedAt    time.Time            `json:"createdAt"`
	JobID        uint                 `gorm:"primaryKey" json:"jobId"`
	UserID       string               `gorm:"primaryKey;type:uuid" json:"userId"`
	ContactPhone string               `json:"phone"`
	ContactEmail string               `json:"email"`
	Status       JobApplicationStatus `json:"status"`
	Files        []File               `gorm:"many2many:job_application_has_file;constraint:OnDelete:CASCADE;" json:"files"`
}

func (jobApplication *JobApplication) BeforeDelete(tx *gorm.DB) (err error) {
	newJobApplication := JobApplication{
		JobID:  jobApplication.JobID,
		UserID: jobApplication.UserID,
	}
	if err := tx.Preload("Files").First(&newJobApplication).Error; err != nil {
		return err
	}
	// Use the registered storage deletion hook to remove stored files.
	// CallStorageDeleteHook is idempotent and will return nil if the object is already absent
	// or if no deletion hook/provider is registered.
	for _, file := range newJobApplication.Files {
		if file.ID == "" {
			continue
		}
		if err := CallStorageDeleteHook(context.Background(), file.ID); err != nil {
			return err
		}
	}
	return nil
}
