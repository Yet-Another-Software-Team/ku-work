package model

import (
	"time"

	"gorm.io/datatypes"
)

type Student struct {
	UserID              string         `gorm:"type:uuid;primarykey" json:"id"`
	User                User           `gorm:"foreignKey:UserID" json:"-"`
	Approved            bool           `json:"approved"`
	CreatedAt           time.Time      `json:"created"`
	Phone               string         `json:"phone"`
	PhotoID             string         `gorm:"type:uuid" json:"photoId"`
	Photo               File           `gorm:"foreignKey:PhotoID" json:"photo,omitempty"`
	BirthDate           datatypes.Date `json:"birthDate"`
	AboutMe             string         `json:"aboutMe"`
	GitHub              string         `json:"github"`
	LinkedIn            string         `json:"linkedIn"`
	StudentID           string         `json:"studentId"`
	Major               string         `json:"major"`
	StudentStatus       string         `json:"status"`
	StudentStatusFileID string         `gorm:"type:uuid" json:"statusFileId"`
	StudentStatusFile   File           `gorm:"foreignKey:StudentStatusFileID" json:"statusFile,omitempty"`
	JobApplications   []JobApplication `gorm:"foreignkey:UserID" json:"-"`
}
