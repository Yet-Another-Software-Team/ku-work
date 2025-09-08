package model

import (
	"time"
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
	Company         User             `gorm:"foreignKey:CompanyID" json:"-"`
	Position        string           `json:"position"`
	Duration        string           `json:"duration"`
	Description     string           `json:"description"`
	Location        string           `json:"location"`
	JobType         JobType          `json:"jobType"`
	Experience      ExperienceType   `json:"experienceType"`
	MinSalary       uint             `json:"minSalary"`
	MaxSalary       uint             `json:"maxSalary"`
	IsApproved      bool             `json:"approved"`
	JobApplications []JobApplication `gorm:"foreignkey:JobID" json:"-"`
}

type JobApplication struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	JobID     uint      `json:"jobId"`
	UserID    string    `gorm:"type:uuid" json:"userId"`
	AltPhone  string    `json:"phone"`
	AltEmail  string    `json:"email"`
	Files     []File    `json:"files"`
}
