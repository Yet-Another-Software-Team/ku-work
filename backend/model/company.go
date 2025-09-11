package model

import (
	"time"
)

type Company struct {
	UserID    string    `gorm:"type:uuid;primarykey" json:"id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt time.Time `json:"created"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	PhotoID   string    `gorm:"type:uuid" json:"photoId"`
	Photo     File      `gorm:"foreignKey:PhotoID" json:"photo,omitempty"`
	BannerID  string    `gorm:"type:uuid" json:"bannerId"`
	Banner    File      `gorm:"foreignKey:BannerID" json:"banner,omitempty"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
}
