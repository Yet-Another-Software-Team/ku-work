package model

import "time"

type Audit struct {
	ID         uint   `gorm:"primarykey"`
	ActorID    string `gorm:"type:uuid;foreignkey:UserID"`
	CreatedAt  time.Time
	Action     string
	ObjectName string
	ObjectID   string
}
