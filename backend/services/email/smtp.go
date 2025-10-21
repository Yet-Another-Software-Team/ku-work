package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"os"
)

type SMTPEmailProvider struct {
	addr   string
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

func (cur *SMTPEmailProvider) SendTo(ctx context.Context, target string, subject string, content string) error {
	msg := fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s", subject, content)

	// Use a channel to handle the email sending with timeout
	errChan := make(chan error, 1)

	go func() {
		errChan <- smtp.SendMail(cur.addr, cur.auth, cur.sender, []string{target}, []byte(msg))
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending timeout: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}
