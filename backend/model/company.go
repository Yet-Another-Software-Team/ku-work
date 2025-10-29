package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Company struct {
	UserID    string    `gorm:"type:uuid;primarykey" json:"id"`
	User      User      `gorm:"foreignKey:UserID" json:"User"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Website   string    `json:"website"`
	Phone     string    `json:"phone"`
	PhotoID   string    `gorm:"type:uuid" json:"photoId"`
	Photo     File      `gorm:"foreignKey:PhotoID" json:"-"`
	BannerID  string    `gorm:"type:uuid" json:"bannerId"`
	Banner    File      `gorm:"foreignKey:BannerID" json:"-"`
	AboutUs   string    `json:"about"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	Jobs      []Job     `gorm:"foreignkey:CompanyID;constraint:OnDelete:CASCADE;" json:"-"`
}

func (company *Company) BeforeDelete(tx *gorm.DB) (err error) {
	newCompany := Company{
		UserID: company.UserID,
	}
	if err := tx.Preload("Photo").Preload("Banner").First(&newCompany).Error; err != nil {
		return err
	}
	// Use the registered storage deletion hook to remove underlying stored objects.
	// CallStorageDeleteHook is a no-op when no hook/provider is registered.
	if newCompany.Photo.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newCompany.Photo.ID); err != nil {
			return err
		}
	}
	if newCompany.Banner.ID != "" {
		if err := CallStorageDeleteHook(context.Background(), newCompany.Banner.ID); err != nil {
			return err
		}
	}
	return nil
}
