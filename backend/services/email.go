package services

import (
	"context"
	"errors"
	"fmt"
	"ku-work/backend/model"
	"ku-work/backend/services/email"
	"log/slog"
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

// sanitizeEmailContent removes potential email injection sequences from content
// This prevents attackers from injecting email headers through the body
func sanitizeEmailContent(content string) string {
	// Remove any sequence that could break out of the email body section
	// Replace \r\n with just \n to normalize line endings
	sanitized := strings.ReplaceAll(content, "\r\n", "\n")
	// Remove any standalone \r characters
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	return sanitized
}

func (cur *EmailService) SendTo(target string, subject string, content string) error {
	// Escape header values
	escapedTarget := escapeHeaderValue(target)
	escapedSubject := escapeHeaderValue(subject)
	// Sanitize content to prevent email injection
	sanitizedContent := sanitizeEmailContent(content)

	// Create initial log entry
	mailLog := model.MailLog{
		To:         escapedTarget,
		Subject:    escapedSubject,
		Body:       sanitizedContent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     model.MailLogStatusTemporaryError, // If it fail for no reason, log as temporary error to retry later.
		RetryCount: 0,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cur.timeout)
	defer cancel()

	// Attempt to send email with timeout
	err := cur.provider.SendTo(ctx, escapedTarget, escapedSubject, sanitizedContent)

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
	} else {
		mailLog.Status = model.MailLogStatusDelivered
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

// RetryFailedEmails attempts to resend emails that failed with temporary errors
func (cur *EmailService) RetryFailedEmails() error {

	// Get retry configuration from environment variables
	maxAttempts := 3
	if maxAttemptsStr, hasMaxAttempts := os.LookupEnv("EMAIL_RETRY_MAX_ATTEMPTS"); hasMaxAttempts {
		if attempts, err := strconv.Atoi(maxAttemptsStr); err == nil && attempts > 0 {
			maxAttempts = attempts
		}
	}

	retryIntervalMinutes := 30
	if intervalStr, hasInterval := os.LookupEnv("EMAIL_RETRY_INTERVAL_MINUTES"); hasInterval {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			retryIntervalMinutes = interval
		}
	}

	maxAgeHours := 24
	if maxAgeStr, hasMaxAge := os.LookupEnv("EMAIL_RETRY_MAX_AGE_HOURS"); hasMaxAge {
		if age, err := strconv.Atoi(maxAgeStr); err == nil && age > 0 {
			maxAgeHours = age
		}
	}

	// Calculate cutoff times
	retryAfter := time.Now().Add(-time.Duration(retryIntervalMinutes) * time.Minute)
	maxAge := time.Now().Add(-time.Duration(maxAgeHours) * time.Hour)

	// Query for emails with temporary errors that are ready to retry
	var failedEmails []model.MailLog
	result := cur.db.Where(
		"status = ? AND updated_at < ? AND created_at > ?",
		model.MailLogStatusTemporaryError,
		retryAfter,
		maxAge,
	).Find(&failedEmails)

	if result.Error != nil {
		return fmt.Errorf("failed to query failed emails: %w", result.Error)
	}

	if len(failedEmails) == 0 {
		slog.Info("No failed emails to retry")
		return nil
	}

	slog.Info("Found failed emails to retry", "count", len(failedEmails))

	successCount := 0
	permanentFailCount := 0
	temporaryFailCount := 0

	for _, mailLog := range failedEmails {
		// Check if we've exceeded max attempts
		if mailLog.RetryCount >= maxAttempts {
			// Mark as permanent error after max attempts
			mailLog.Status = model.MailLogStatusPermanentError
			mailLog.ErrorDescription = fmt.Sprintf("Max retry attempts (%d) exceeded. Last error: %s",
				maxAttempts, mailLog.ErrorDescription)
			mailLog.UpdatedAt = time.Now()
			cur.db.Save(&mailLog)
			permanentFailCount++
			slog.Error("Email marked as permanent error", "id", mailLog.ID, "attempt", mailLog.RetryCount)
			continue
		}

		// Increment retry count
		mailLog.RetryCount++

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cur.timeout)

		// Attempt to resend
		err := cur.provider.SendTo(ctx, mailLog.To, mailLog.Subject, mailLog.Body)
		cancel()

		// Update mail log based on result
		mailLog.UpdatedAt = time.Now()
		if err != nil {
			errorMsg := err.Error()
			if isTemporaryError(errorMsg) {
				mailLog.Status = model.MailLogStatusTemporaryError
				mailLog.ErrorDescription = errorMsg
				temporaryFailCount++
				slog.Warn("Email retry failed with temporary error", "id", mailLog.ID, "message", errorMsg)
			} else {
				mailLog.Status = model.MailLogStatusPermanentError
				mailLog.ErrorDescription = errorMsg
				permanentFailCount++
				slog.Error("Email retry failed with permanent error", "id", mailLog.ID, "message", errorMsg)
			}
		} else {
			mailLog.Status = model.MailLogStatusDelivered
			mailLog.ErrorDescription = ""
			successCount++
			slog.Info("Email successfully resent", "id", mailLog.ID)
		}

		cur.db.Save(&mailLog)
	}

	slog.Info("Retry summary", "success_count", successCount, "temporary_fail_count", temporaryFailCount, "permanent_fail_count", permanentFailCount)
	return nil
}
