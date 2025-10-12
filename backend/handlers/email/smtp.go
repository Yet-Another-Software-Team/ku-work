package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

type SMTPEmailProvider struct {
	addr   string
	port   string
	sender string
	auth   smtp.Auth
}

func NewSMTPEmailProvider() (*SMTPEmailProvider, error) {
	host, hasHost := os.LookupEnv("SMTP_SERVER_HOST")
	if !hasHost {
		return nil, errors.New("SMTP_SERVER_HOST not specified")
	}
	port, hasPort := os.LookupEnv("SMTP_SERVER_PORT")
	if !hasPort {
		return nil, errors.New("SMTP_SERVER_PORT not specified")
	}
	sender, hasSender := os.LookupEnv("SMTP_SENDER")
	if !hasSender {
		return nil, errors.New("SMTP_SENDER not specified")
	}
	password, hasPassword := os.LookupEnv("SMTP_PASSWORD")
	if !hasPassword {
		return nil, errors.New("SMTP_PASSWORD not specified")
	}
	return &SMTPEmailProvider{
		sender: sender,
		addr:   fmt.Sprintf("%s:%s", host, port),
		auth:   smtp.PlainAuth("", sender, password, host),
	}, nil
}

func (cur *SMTPEmailProvider) SendTo(target string, subject string, content string) error {
	msg := fmt.Sprintf("Subject: %s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n\n%s", subject, content)
	return smtp.SendMail(cur.addr, cur.auth, cur.sender, []string{target}, []byte(msg))
}
