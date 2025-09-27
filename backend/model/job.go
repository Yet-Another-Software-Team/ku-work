package model

import (
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

type Job struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time        `json:"createdAt"`
	Name            string           `json:"name"`
	CompanyID       string           `gorm:"type:uuid" json:"companyId"`
	Company         Company          `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE;" json:"company"`
	Position        string           `json:"position"`
	Duration        string           `json:"duration"`
	Description     string           `json:"description"`
	Location        string           `json:"location"`
	JobType         JobType          `json:"jobType"`
	Experience      ExperienceType   `json:"experienceType"`
	MinSalary       uint             `json:"minSalary"`
	MaxSalary       uint             `json:"maxSalary"`
	IsApproved      bool             `json:"approved"`
	IsOpen          bool             `json:"open"`
	JobApplications []JobApplication `gorm:"foreignkey:JobID;constraint:OnDelete:CASCADE;" json:"-"`
}

type JobApplicationStatus string

const (
	JobApplicationAccepted JobApplicationStatus = "accepted"
	JobApplicationRejected JobApplicationStatus = "rejected"
	JobApplicationPending  JobApplicationStatus = "pending"
)

type JobApplication struct {
	ID        uint                 `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time            `json:"createdAt"`
	JobID     uint                 `json:"jobId"`
	UserID    string               `gorm:"type:uuid" json:"userId"`
	AltPhone  string               `json:"phone"`
	AltEmail  string               `json:"email"`
	Status    JobApplicationStatus `json:"status"`
	Files     []File               `gorm:"many2many:job_application_has_file;constraint:OnDelete:CASCADE;" json:"files"`
}

func (jobApplication *JobApplication) BeforeDelete(tx *gorm.DB) (err error) {
	newJobApplication := JobApplication{
		ID: jobApplication.ID,
	}
	if err := tx.Preload("Files").First(&newJobApplication).Error; err != nil {
		return err
	}
	for _, file := range newJobApplication.Files {
		if err := file.AfterDelete(tx); err != nil {
			return err
		}
	}
	return nil
}
