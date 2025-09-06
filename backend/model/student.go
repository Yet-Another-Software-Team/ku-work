package model

import (
	"gorm.io/datatypes"
	"time"
)

type Student struct {
	UserID            string           `gorm:"type:uuid;primarykey" json:"id"`
	User              User             `gorm:"foreignKey:UserID" json:"-"`
	Approved          bool             `json:"approved"`
	CreatedAt         time.Time        `json:"created"`
	Phone             string           `json:"phone"`
	Photo             string           `json:"photo"`
	BirthDate         datatypes.Date   `json:"birthDate"`
	AboutMe           string           `json:"aboutMe"`
	GitHub            string           `json:"github"`
	LinkedIn          string           `json:"linkedIn"`
	StudentID         string           `json:"studentId"`
	Major             string           `json:"major"`
	StudentStatus     string           `json:"status"`
	StudentStatusFile string           `json:"statusFile"`
	JobApplications   []JobApplication `gorm:"foreignkey:UserID" json:"-"`
}
