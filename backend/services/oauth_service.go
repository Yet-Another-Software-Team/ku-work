package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// OAuthService encapsulates Google OAuth code exchange and userinfo retrieval.
type OAuthService struct {
	// Config is the Google OAuth2 configuration (client ID, secret, scopes, endpoint).
	Config *oauth2.Config
	// HTTPClient is used for outbound HTTP calls. If nil, a default client with a 10s timeout is used.
	HTTPClient *http.Client
	// UserInfoEndpoint is the Google userinfo endpoint. Defaults to "https://www.googleapis.com/oauth2/v2/userinfo".
	UserInfoEndpoint string
}

// UserInfo represents the subset of Google userinfo we use downstream.
type UserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

// NewOAuthService constructs a new OAuthService with sensible defaults.
// If client is nil, a default HTTP client with a 10 second timeout is used.
// If userInfoEndpoint is empty, it defaults to Google's v2 userinfo endpoint.
func NewOAuthService(cfg *oauth2.Config, client *http.Client, userInfoEndpoint string) *OAuthService {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if userInfoEndpoint == "" {
		userInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"
	}
	return &OAuthService{
		Config:           cfg,
		HTTPClient:       client,
		UserInfoEndpoint: userInfoEndpoint,
	}
}

// ExchangeCode exchanges an authorization code for an access token using the configured OAuth2 client.
func (s *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	if s == nil || s.Config == nil {
		return nil, fmt.Errorf("oauth service not configured")
	}
	return s.Config.Exchange(ctx, code)
}

// FetchUserInfo retrieves user information from Google using an access token.
// Returns the parsed UserInfo, an HTTP-like status code, and an error (if any).
func (s *OAuthService) FetchUserInfo(token *oauth2.Token) (UserInfo, int, error) {
	var ui UserInfo

	if s == nil || s.HTTPClient == nil {
		return ui, http.StatusInternalServerError, fmt.Errorf("oauth service http client not configured")
	}
	if s.UserInfoEndpoint == "" {
		return ui, http.StatusInternalServerError, fmt.Errorf("userinfo endpoint not configured")
	}
	if token == nil || token.AccessToken == "" {
		return ui, http.StatusBadRequest, fmt.Errorf("missing access token")
	}

	req, err := http.NewRequest("GET", s.UserInfoEndpoint, nil)
	if err != nil {
		return ui, http.StatusInternalServerError, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return ui, http.StatusInternalServerError, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Failed to close response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return ui, http.StatusUnauthorized, fmt.Errorf("access token invalid or expired")
	}

	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return ui, http.StatusInternalServerError, err
	}

	return ui, http.StatusOK, nil
}

// ExchangeAndFetchUserInfo is a convenience helper that exchanges an auth code and then
// fetches the corresponding Google user information.
func (s *OAuthService) ExchangeAndFetchUserInfo(ctx context.Context, code string) (UserInfo, int, error) {
	var ui UserInfo

	tok, err := s.ExchangeCode(ctx, code)
	if err != nil {
		return ui, http.StatusInternalServerError, err
	}
	return s.FetchUserInfo(tok)
}
