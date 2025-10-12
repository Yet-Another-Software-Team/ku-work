package model

import "time"

type Audit struct {
	ID         uint `gorm:"primarykey"`
	ActorID    string
	CreatedAt  time.Time
	Action     string
	Reason     string
	ObjectName string
	ObjectID   string
}
