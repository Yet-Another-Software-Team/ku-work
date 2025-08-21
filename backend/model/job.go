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
	Company     User `gorm:"foreignKey:CompanyID"`
	Position    string
	Duration    string
	Description string
	Location    string
	JobType     JobType
	Experience  ExperienceType
	MinSalary   uint
	MaxSalary   uint
	IsApproved  bool
}
