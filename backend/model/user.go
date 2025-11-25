// Database Model for Users related entities
package model

import (
	"time"

	"gorm.io/gorm"
)

// Represents a user in the system.
type User struct {
	ID           string `gorm:"type:uuid;primarykey;default:gen_random_uuid()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Username     string         `gorm:"index:idx_username_user_type,unique"`
	UserType     string         `gorm:"index:idx_username_user_type,unique"`
	PasswordHash string         `json:"-"`
	AcceptPDPA   bool			`gorm:"default:false"`
	AcceptGDPR   bool			`gorm:"default:false"`
}

func (user *User) BeforeDelete(tx *gorm.DB) (err error) {
	company := Company{
		UserID: user.ID,
	}
	result := tx.Limit(1).Find(&company)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 0 {
		if err := company.BeforeDelete(tx); err != nil {
			return err
		}
	}
	student := Student{
		UserID: user.ID,
	}
	result = tx.Limit(1).Find(&student)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected != 0 {
		if err := student.BeforeDelete(tx); err != nil {
			return err
		}
	}
	return nil
}

// Represents a user's Google OAuth details.
type GoogleOAuthDetails struct {
	UserID     string `gorm:"type:uuid;foreignkey:UserID;primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	ExternalID string
	FirstName  string
	LastName   string
	Email      string `gorm:"unique;index"`
}

// Represents a user's who is an Admin without any additional fields.
type Admin struct {
	UserID string `gorm:"type:uuid;foreignkey:UserID;primarykey"`
}
