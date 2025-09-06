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
	Photo     string    `json:"photo"`
	Banner    string    `json:"banner"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
}
