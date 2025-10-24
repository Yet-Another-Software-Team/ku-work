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

// JWT Payload (NOT DATABASE INSTANCE)
type UserClaims struct {
	UserID               string `json:"user_id"`
	jwt.RegisteredClaims        // Includes JTI (JWT ID) via RegisteredClaims.ID field
}
