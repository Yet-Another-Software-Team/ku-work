package handlers

import (
	"errors"
	"ku-work/backend/handlers/email"
	"os"

	"gorm.io/gorm"
)

type EmailHandler struct {
	provider email.EmailProvider
}

func NewEmailHandler(DB *gorm.DB) (*EmailHandler, error) {
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
		return &EmailHandler{
			provider: smtpProvider,
		}, nil
	case "dummy":
		return &EmailHandler{
			provider: email.NewDummyEmailProvider(),
		}, nil
	}
	return nil, errors.New("invalid EMAIL_PROVIDER specified")
}

func (cur *EmailHandler) SendTo(target string, subject string, content string) error {
	return cur.provider.SendTo(target, subject, content)
}
