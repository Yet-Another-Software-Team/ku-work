package helper

import (
	"ku-work/backend/model"
	"log"
	"time"

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

// DEPRECATED: GetRole - Database operations moved to IdentityService.ResolveRole
// This function is kept for backward compatibility but should not be used in new code.
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

// DEPRECATED: GetUsername - Database operations moved to IdentityService.GetUsername
// This function is kept for backward compatibility but should not be used in new code.
func GetUsername(userID string, role Role, db *gorm.DB) string {
	if userID == "" {
		return "unknown"
	}
	if role == Company || role == Admin {
		var user model.User
		if db.Where("id = ?", userID).First(&user).Error == nil {
			return user.Username
		}
	}
	if role == Viewer || role == Student {
		var profile model.GoogleOAuthDetails
		if db.Where("user_id = ?", userID).First(&profile).Error == nil {
			return profile.FirstName + " " + profile.LastName
		}
	}
	return "unknown"
}

// CleanupExpiredTokens removes expired refresh tokens from the database.
// Keeps revoked tokens for 7 days for token reuse detection.
// This function is designed to be called by the scheduler.
func CleanupExpiredTokens(db *gorm.DB) error {
	now := time.Now()
	gracePeriod := now.Add(-7 * 24 * time.Hour) // 7 days ago

	// Delete tokens that are:
	// 1. Expired AND not revoked (normal expiration), OR
	// 2. Revoked more than 7 days ago (grace period for reuse detection)
	result := db.Unscoped().
		Where("(expires_at < ? AND revoked_at IS NULL) OR (revoked_at IS NOT NULL AND revoked_at < ?)", now, gracePeriod).
		Delete(&model.RefreshToken{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired refresh tokens", result.RowsAffected)
	}
	return nil
}
