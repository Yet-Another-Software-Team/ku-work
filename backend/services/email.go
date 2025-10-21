package services

import (
	"context"
	"errors"
	"ku-work/backend/model"
	"ku-work/backend/services/email"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type EmailService struct {
	provider email.EmailProvider
	db       *gorm.DB
	timeout  time.Duration
}

func NewEmailService(DB *gorm.DB) (*EmailService, error) {
	emailProvider, hasEmailProvider := os.LookupEnv("EMAIL_PROVIDER")
	if !hasEmailProvider {
		return nil, errors.New("EMAIL_PROVIDER not specified")
	}
	var provider email.EmailProvider

	switch emailProvider {
	case "SMTP":
		smtpProvider, err := email.NewSMTPEmailProvider()
		if err != nil {
			return nil, err
		}
		provider = smtpProvider
	case "gmail":
		gmailProvider, err := email.NewGmailEmailProvider()
		if err != nil {
			return nil, err
		}
		provider = gmailProvider
	case "dummy":
		provider = email.NewDummyEmailProvider()
	default:
		return nil, errors.New("invalid EMAIL_PROVIDER specified")
	}

	// Get timeout from environment variable, default to 30 seconds
	timeout := 30 * time.Second
	if timeoutStr, hasTimeout := os.LookupEnv("EMAIL_TIMEOUT_SECONDS"); hasTimeout {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			timeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	return &EmailService{
		provider: provider,
		db:       DB,
		timeout:  timeout,
	}, nil
}

func escapeHeaderValue(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\n", ""), "\r", "")
}

func (cur *EmailService) SendTo(target string, subject string, content string) error {
	// Escape header values
	escapedTarget := escapeHeaderValue(target)
	escapedSubject := escapeHeaderValue(subject)

	// Create initial log entry
	mailLog := model.MailLog{
		To:        escapedTarget,
		Subject:   escapedSubject,
		Body:      content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    model.MailLogStatusDelivered,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cur.timeout)
	defer cancel()

	// Attempt to send email with timeout
	err := cur.provider.SendTo(ctx, escapedTarget, escapedSubject, content)

	// Update log status based on result
	if err != nil {
		// Determine if error is temporary or permanent
		errorMsg := err.Error()
		if isTemporaryError(errorMsg) {
			mailLog.Status = model.MailLogStatusTemporaryError
		} else {
			mailLog.Status = model.MailLogStatusPermanentError
		}
		mailLog.ErrorDescription = errorMsg
	}

	// Save log to database
	if cur.db != nil {
		cur.db.Create(&mailLog)
	}

	return err
}

// isTemporaryError determines if an email error is temporary (can be retried)
func isTemporaryError(errorMsg string) bool {
	// Common temporary error patterns
	temporaryPatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"try again",
		"rate limit",
		"too many",
		"421", // SMTP temporary failure code
		"450", // SMTP mailbox unavailable
		"451", // SMTP local error
		"452", // SMTP insufficient storage
		"deadline exceeded",
		"context deadline exceeded",
	}

	errorLower := strings.ToLower(errorMsg)
	for _, pattern := range temporaryPatterns {
		if strings.Contains(errorLower, pattern) {
			return true
		}
	}

	return false
}
