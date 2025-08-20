package model

import (
	"time"
)

type ExperienceType string

const (
	NewGrad ExperienceType = "newgrad"
	Junior  ExperienceType = "junior"
	Senior  ExperienceType = "senior"
	Manager ExperienceType = "manager"
)

type JobType string

const (
	FullTime JobType = "fulltime"
	PartTime JobType = "parttime"
	Contract JobType = "contract"
	Casual   JobType = "casual"
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
	JobType     JobType        `gorm:"type:enum('fulltime', 'parttime', 'contract', 'casual')"`
	Experience  ExperienceType `gorm:"type:enum('newgrad', 'junior', 'senior', 'manager')"`
	MinSalary   uint
	MaxSalary   uint
	IsApproved  bool
}
