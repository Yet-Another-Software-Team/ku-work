package model

import (
	"context"
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
	UpdatedAt           time.Time             `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt        `gorm:"index" json:"-"`
	Phone               string                `json:"phone"`
	PhotoID             string                `gorm:"type:uuid" json:"photoId"`
	Photo               File                  `gorm:"foreignKey:PhotoID;constraint:OnDelete:CASCADE;" json:"photo"`
	BirthDate           datatypes.Date        `json:"birthDate"`
	AboutMe             string                `json:"aboutMe"`
	GitHub              string                `json:"github"`
	LinkedIn            string                `json:"linkedIn"`
	StudentID           string                `json:"studentId"`
	Major               string                `json:"major"`
	StudentStatus       string                `json:"status"`
	StudentStatusFileID string                `gorm:"type:uuid" json:"statusFileId"`
	StudentStatusFile   File                  `gorm:"foreignKey:StudentStatusFileID;constraint:OnDelete:CASCADE;" json:"statusFile"`
	JobApplications     []JobApplication      `gorm:"foreignkey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
}

func (student *Student) BeforeDelete(tx *gorm.DB) (err error) {
	newStudent := Student{
		UserID: student.UserID,
	}
	if err := tx.Preload("Photo").Preload("StudentStatusFile").Preload("JobApplications").First(&newStudent).Error; err != nil {
		return err
	}
	// Ensure any JobApplication-associated files are cleaned up via their hooks/logic.
	for _, application := range newStudent.JobApplications {
		if err := application.BeforeDelete(tx); err != nil {
			return err
		}
	}
	// Delete associated stored objects (photo and student status file) via the registered hook.
	// CallStorageDeleteHook is a no-op when no hook/provider is registered.
	if newStudent.Photo.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newStudent.Photo.ID); err != nil {
			return err
		}
	}
	if newStudent.StudentStatusFile.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newStudent.StudentStatusFile.ID); err != nil {
			return err
		}
	}
	return nil
}
