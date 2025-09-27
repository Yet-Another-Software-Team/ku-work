package model

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	UserID    string    `gorm:"type:uuid;primarykey" json:"id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   string    `gorm:"type:uuid" json:"photoId"`
	Photo     File      `gorm:"foreignKey:PhotoID" json:"photo,omitempty"`
	BannerID  string    `gorm:"type:uuid" json:"bannerId"`
	Banner    File      `gorm:"foreignKey:BannerID" json:"banner,omitempty"`
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
	if err := newCompany.Photo.AfterDelete(tx); err != nil {
		return err
	}
	if err := newCompany.Banner.AfterDelete(tx); err != nil {
		return err
	}
	return nil
}
