package email

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
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

func (cur *GmailEmailProvider) SendTo(target string, subject string, content string) error {
	message := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n\n%s", strings.ReplaceAll(strings.ReplaceAll(target, "\n", ""), "\r", ""), subject, content))),
	}
	_, err := cur.gmailService.Users.Messages.Send("me", &message).Do()
	return err
}
