package helper

import (
	"ku-work/backend/model"

	"gorm.io/gorm"
)

type Role string

const (
	Admin   Role = "admin"
	Company Role = "company"
	Student Role = "student"
	Viewer  Role = "viewer"
	Unknown Role = "unknown"
)

// Helper function to get the role of a user based on their ID and database connection.
func GetRole(userID string, db *gorm.DB) Role {
	if userID == "" {
		return Unknown
	}
	if result := db.Find(&model.Admin{UserID: userID}); result.Error == nil && result.RowsAffected > 0 {
		return Admin
	}
	if result := db.Find(&model.Company{UserID: userID}); result.Error == nil && result.RowsAffected > 0 {
		return Company
	}
	if result := db.Find(&model.Student{UserID: userID, ApprovalStatus: model.StudentApprovalAccepted}); result.Error == nil && result.RowsAffected > 0 {
		return Student
	}
	if result := db.Find(&model.GoogleOAuthDetails{UserID: userID}); result.Error == nil && result.RowsAffected > 0 {
		return Viewer
	}
	return Unknown
}

// Helper function to get username of user based on their ID, role and database connection.
func GetUsername(userID string, role Role, db *gorm.DB) string {
	if userID == "" {
		return "unknown"
	}
	if role == Company {
		var user model.User
		if db.Find(&user, userID).Error == nil {
			return user.Username
		}
	}
	if role == Viewer || role == Student {
		var profile model.GoogleOAuthDetails
		if db.Find(&profile, userID).Error == nil {
			return profile.FirstName + " " + profile.LastName
		}
	}
	return "unknown"
}
