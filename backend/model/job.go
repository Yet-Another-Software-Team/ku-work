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
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	Name        string
	CompanyID   uint
	Company     User `gorm:"foreignKey:CompanyRefer"`
	Position    string
	Duration    string
	Description string
	Location    string
	JobType     JobType        `gorm:"type:enum('fulltime', 'parttime', 'contract', 'casual', 'internship')"`
	Experience  ExperienceType `gorm:"type:enum('newgrad', 'junior', 'senior', 'manager', 'internship')"`
	MinSalary   uint
	MaxSalary   uint
	IsApproved  bool
}
