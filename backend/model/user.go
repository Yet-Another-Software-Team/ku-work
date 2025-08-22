// Database Model for Users related entities
package model

import (
	"time"

	"gorm.io/gorm"
)

// Represents a user in the system.
type User struct {
	ID        string `gorm:"type:uuid;primarykey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Username  string         `gorm:"unique"`
	PasswordHash string
}

// Represents a user's Google OAuth details.
type GoogleOAuthDetails struct {
	gorm.Model
	UserID     string `gorm:"type:uuid;foreignkey:UserID"`
	ExternalID string
	FirstName  string
	LastName   string
	Email      string
}

// Represents a user's who is an Admin without any additional fields.
type Admin struct {
	gorm.Model
	UserID string `gorm:"type:uuid;foreignkey:UserID"`
}
