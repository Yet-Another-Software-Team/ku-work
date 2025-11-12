package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Company represents a company profile owned by a User.
// The struct includes file references (photo/banner) and uses GORM's
// soft-delete via DeletedAt so records can be retained for audit.
type Company struct {
	// UserID is the primary key and also references users.id
	UserID    string         `gorm:"type:uuid;primarykey" json:"id"`
	User      User           `gorm:"foreignKey:UserID" json:"User"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
<<<<<<< HEAD

	Email    string `json:"email"`
	Website  string `json:"website"`
	Phone    string `json:"phone"`
	PhotoID  string `gorm:"type:uuid" json:"photoId"`
	Photo    File   `gorm:"foreignKey:PhotoID" json:"-"`
	BannerID string `gorm:"type:uuid" json:"bannerId"`
	Banner   File   `gorm:"foreignKey:BannerID" json:"-"`
	AboutUs  string `json:"about"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`

	// Jobs relation (not returned by JSON)
	Jobs []Job `gorm:"foreignkey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"`
}

// BeforeDelete is a GORM hook executed before a company record is deleted.
// It attempts to remove associated files (photo and banner) from the configured
// storage provider via the model-level storage hook.
=======
	Email     string         `json:"email"`
	Website   string         `json:"website"`
	Phone     string         `json:"phone"`
	PhotoID   string         `gorm:"type:uuid" json:"photoId"`
	Photo     File           `gorm:"foreignKey:PhotoID" json:"-"`
	BannerID  string         `gorm:"type:uuid" json:"bannerId"`
	Banner    File           `gorm:"foreignKey:BannerID" json:"-"`
	AboutUs   string         `json:"about"`
	Address   string         `json:"address"`
	City      string         `json:"city"`
	Country   string         `json:"country"`
	Jobs      []Job          `gorm:"foreignkey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"`
}

// BeforeDelete is a GORM hook that deletes associated files from storage.
>>>>>>> main
func (company *Company) BeforeDelete(tx *gorm.DB) (err error) {
	// Load the company along with file relations to obtain file IDs.
	var existing Company
	if err := tx.Preload("Photo").Preload("Banner").First(&existing, "user_id = ?", company.UserID).Error; err != nil {
		return err
	}

<<<<<<< HEAD
	// If a photo is present, call the storage deletion hook (best-effort; surface errors).
	if existing.Photo.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), existing.Photo.ID); err != nil {
			return err
		}
	}

	// If a banner is present, call the storage deletion hook.
	if existing.Banner.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), existing.Banner.ID); err != nil {
=======
	if newCompany.Photo.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newCompany.Photo.ID); err != nil {
			return err
		}
	}
	if newCompany.Banner.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newCompany.Banner.ID); err != nil {
>>>>>>> main
			return err
		}
	}

	return nil
}
