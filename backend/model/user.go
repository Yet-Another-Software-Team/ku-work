// Database Model for Users related entities
package model

import "gorm.io/gorm"

// Represents a user in the system.
type User struct {
	gorm.Model
	Username     string `gorm:"unique"`
	PasswordHash string
}

// Represents a user's Google OAuth details.
type GoogleOAuthDetails struct {
	gorm.Model
	UserID     uint `gorm:"unique"`
	ExternalID string
	FirstName  string
	LastName   string
	Email      string
}

// Represents a user's who is an Admin without any additional fields.
type Admin struct {
	gorm.Model
	UserID uint `gorm:"unique"`
}
