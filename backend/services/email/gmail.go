package email

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailEmailProvider struct {
	gmailService *gmail.Service
}

func NewGmailEmailProvider() (*GmailEmailProvider, error) {
	clientId, hasClientId := os.LookupEnv("GMAIL_OAUTH_CLIENT_ID")
	if !hasClientId {
		return nil, errors.New("GMAIL_OAUTH_CLIENT_ID not specified")
	}
	clientSecret, hasClientSecret := os.LookupEnv("GMAIL_OAUTH_CLIENT_SECRET")
	if !hasClientSecret {
		return nil, errors.New("GMAIL_OAUTH_CLIENT_SECRET not specified")
	}
	accessToken, hasAccessToken := os.LookupEnv("GMAIL_OAUTH_ACCESS_TOKEN")
	if !hasAccessToken {
		return nil, errors.New("GMAIL_OAUTH_ACCESS_TOKEN not specified")
	}
	refreshToken, hasRefreshToken := os.LookupEnv("GMAIL_OAUTH_REFRESH_TOKEN")
	if !hasRefreshToken {
		return nil, errors.New("GMAIL_OAUTH_REFRESH_TOKEN not specified")
	}
	config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}
	token := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}
	tokenSource := config.TokenSource(context.Background(), &token)
	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}
	return &GmailEmailProvider{
		gmailService: srv,
	}, nil
}

func (cur *GmailEmailProvider) SendTo(ctx context.Context, target string, subject string, content string) error {
	message := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s", target, subject, content))),
	}

	// Use a channel to handle the email sending with timeout
	errChan := make(chan error, 1)

	go func() {
		_, err := cur.gmailService.Users.Messages.Send("me", &message).Context(ctx).Do()
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending timeout: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}
