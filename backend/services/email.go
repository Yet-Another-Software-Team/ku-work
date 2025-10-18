package services

import (
	"errors"
	"ku-work/backend/services/email"
	"os"

	"gorm.io/gorm"
)

type EmailService struct {
	provider email.EmailProvider
}

func NewEmailService(DB *gorm.DB) (*EmailService, error) {
	emailProvider, hasEmailProvider := os.LookupEnv("EMAIL_PROVIDER")
	if !hasEmailProvider {
		return nil, errors.New("EMAIL_PROVIDER not specified")
	}
	switch emailProvider {
	case "SMTP":
		smtpProvider, err := email.NewSMTPEmailProvider()
		if err != nil {
			return nil, err
		}
		return &EmailService{
			provider: smtpProvider,
		}, nil
	case "gmail":
		gmailProvider, err := email.NewGmailEmailProvider()
		if err != nil {
			return nil, err
		}
		return &EmailService{
			provider: gmailProvider,
		}, nil
	case "dummy":
		return &EmailService{
			provider: email.NewDummyEmailProvider(),
		}, nil
	}
	return nil, errors.New("invalid EMAIL_PROVIDER specified")
}

func (cur *EmailService) SendTo(target string, subject string, content string) error {
	return cur.provider.SendTo(target, subject, content)
}
