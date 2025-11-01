package model

import "time"

type MailLogStatus string

const (
	MailLogStatusDelivered      MailLogStatus = "delivered"
	MailLogStatusTemporaryError MailLogStatus = "temporary_error"
	MailLogStatusPermanentError MailLogStatus = "permanent_error"
)

type MailLog struct {
	ID               uint          `gorm:"primary_key" json:"id"`
	To               string        `gorm:"not null" json:"to"`
	Subject          string        `gorm:"not null" json:"subject"`
	Body             string        `gorm:"not null" json:"body"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	Status           MailLogStatus `gorm:"not null" json:"status"`
	ErrorCode        string        `json:"errorCode,omitempty"`
	ErrorDescription string        `json:"errorDesc,omitempty"`
	RetryCount       int           `gorm:"default:0" json:"retryCount"` // Number of retry attempts made
}
