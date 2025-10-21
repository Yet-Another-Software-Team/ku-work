package model

import "time"

type MailLogStatus string

const (
	MailLogStatusDelivered      MailLogStatus = "delivered"
	MailLogStatusTemporaryError MailLogStatus = "temporary_error"
	MailLogStatusPermanentError MailLogStatus = "permanent_error"
)

type MailLog struct {
	ID        uint   `gorm:"primary_key"`
	To        string `gorm:"not null"`
	Subject   string `gorm:"not null"`
	Body      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    MailLogStatus `gorm:"not null"`
	ErrorCode string
	ErrorDescription string
}
