package helper

import (
	"ku-work/backend/model"
	"os"
	"strconv"

	"gorm.io/gorm"
)

// GetGracePeriodDays returns the grace period in days from environment variable
// Defaults to 30 days if not set or invalid
func GetGracePeriodDays() int {
	gracePeriodStr := os.Getenv("ACCOUNT_DELETION_GRACE_PERIOD_DAYS")
	if gracePeriodStr == "" {
		return 30 // Default to 30 days
	}

	days, err := strconv.Atoi(gracePeriodStr)
	if err != nil || days <= 0 {
		return 30 // Default to 30 days if invalid
	}

	return days
}

func IsDeactivated(db *gorm.DB, userId string) bool {
	user := model.User{
		ID: userId,
	}
	err := db.Unscoped().First(&user).Error
	if err != nil {
		return false
	}
	return user.DeletedAt.Valid
}
