package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StudentApprovalStatus string

const (
	StudentApprovalAccepted StudentApprovalStatus = "accepted"
	StudentApprovalRejected StudentApprovalStatus = "rejected"
	StudentApprovalPending  StudentApprovalStatus = "pending"
)

type Student struct {
	UserID              string                `gorm:"type:uuid;primarykey" json:"id"`
	User                User                  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	ApprovalStatus      StudentApprovalStatus `json:"approvalStatus"`
	CreatedAt           time.Time             `json:"createdAt"`
	Phone               string                `json:"phone"`
	PhotoID             string                `gorm:"type:uuid" json:"photoId"`
	Photo               File                  `gorm:"foreignKey:PhotoID;constraint:OnDelete:CASCADE;" json:"photo,omitempty"`
	BirthDate           datatypes.Date        `json:"birthDate"`
	AboutMe             string                `json:"aboutMe"`
	GitHub              string                `json:"github"`
	LinkedIn            string                `json:"linkedIn"`
	StudentID           string                `json:"studentId"`
	Major               string                `json:"major"`
	StudentStatus       string                `json:"status"`
	StudentStatusFileID string                `gorm:"type:uuid" json:"statusFileId"`
	StudentStatusFile   File                  `gorm:"foreignKey:StudentStatusFileID;constraint:OnDelete:CASCADE;" json:"statusFile,omitempty"`
	JobApplications     []JobApplication      `gorm:"foreignkey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
}

func (student *Student) BeforeDelete(tx *gorm.DB) (err error) {
	newStudent := Student{
		UserID: student.UserID,
	}
	if err := tx.Preload("Photo").Preload("StudentStatusFile").Preload("JobApplications").First(&newStudent).Error; err != nil {
		return err
	}
	for _, application := range newStudent.JobApplications {
		if err := application.BeforeDelete(tx); err != nil {
			return err
		}
	}
	if err := newStudent.Photo.AfterDelete(tx); err != nil {
		return err
	}
	if err := newStudent.StudentStatusFile.AfterDelete(tx); err != nil {
		return err
	}
	return nil
}
