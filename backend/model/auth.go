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
	UserID        string `gorm:"type:uuid;foreignkey:UserID"`
	TokenSelector string `gorm:"type:varchar(32);uniqueIndex"` // Non-sensitive selector for fast lookup
	Token         string `gorm:"unique"`                       // Hashed token value
	ExpiresAt     time.Time
	RevokedAt     *time.Time `gorm:"index"` // NULL = active, set = revoked (for reuse detection)
}

// RevokedJWT represents a blacklisted JWT token (for logout/session termination)
// OWASP requirement: Maintain a list of terminated tokens to prevent reuse after logout
type RevokedJWT struct {
	ID        uint      `gorm:"primarykey"`
	JTI       string    `gorm:"type:varchar(36);uniqueIndex;not null"` // JWT ID (unique identifier)
	UserID    string    `gorm:"type:uuid;index;not null"`              // User who owned the token
	ExpiresAt time.Time `gorm:"index;not null"`                        // Original token expiration (for cleanup)
	RevokedAt time.Time `gorm:"not null"`                              // When the token was revoked
}

// JWT Payload (NOT DATABASE INSTANCE)
type UserClaims struct {
	UserID               string `json:"user_id"`
	jwt.RegisteredClaims        // Includes JTI (JWT ID) via RegisteredClaims.ID field
}
