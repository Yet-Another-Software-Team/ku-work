// Database Model for Authentication related entities
package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Represent JWT refresh token and its details.
type RefreshToken struct {
	gorm.Model
	UserID    string `gorm:"type:uuid;foreignkey:UserID"`
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
}

// JWT Payload (NOT DATABASE INSTANCE)
type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
