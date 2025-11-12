package helper

import (
	"os"
	"strconv"
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
