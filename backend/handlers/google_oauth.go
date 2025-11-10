package handlers

import (
	"context"
	"net/http"

	"ku-work/backend/helper"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

// OauthHandlers wires HTTP requests to the OAuth and Auth services.
type OauthHandlers struct {
	OAuth *services.OAuthService
	Auth  *services.AuthService
}

// NewOAuthHandlers constructs a new OauthHandlers instance.
func NewOAuthHandlers(oauth *services.OAuthService, auth *services.AuthService) *OauthHandlers {
	if oauth == nil || auth == nil {
		panic("OAuth and Auth services must be provided")
	}
	return &OauthHandlers{
		OAuth: oauth,
		Auth:  auth,
	}
}

// oauthToken represents the incoming payload for Google OAuth code exchange.
// Kept unexported to match swagger definition path "handlers.oauthToken".
type oauthToken struct {
	Code string `json:"code"`
}

// GoogleOauthHandler
// @Summary Handle Google OAuth login
// @Description Handles the server-side flow for Google OAuth2. It receives an authorization code from the client, exchanges it for a token with Google, fetches the user's profile information, and then either creates a new user account or logs in an existing user. On success, it returns a JWT token and sets a refresh token cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code body handlers.oauthToken true "Google Authorization Code"
// @Success 200 {object} object{token=string,username=string,role=string,userId=string,isRegistered=bool} "Login successful"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/google/login [post]
func (h *OauthHandlers) GoogleOauthHandler(ctx *gin.Context) {
	var req oauthToken
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	// Exchange code and fetch user info from Google.
	ui, code, err := h.OAuth.ExchangeAndFetchUserInfo(context.Background(), req.Code)
	if err != nil {
		// code is an HTTP-like status from the OAuth service
		if code <= 0 {
			code = http.StatusInternalServerError
		}
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	// Delegate business logic (create/update users and issue tokens) to AuthService.
	jwtToken, refreshToken, username, role, userId, isRegistered, statusCode, err := h.Auth.HandleGoogleOAuth(struct {
		ID         string
		Email      string
		Name       string
		GivenName  string
		FamilyName string
	}{
		ID:         ui.ID,
		Email:      ui.Email,
		Name:       ui.Name,
		GivenName:  ui.GivenName,
		FamilyName: ui.FamilyName,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set refresh token cookie
	maxAge := helper.CookieMaxAge()
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(statusCode, gin.H{
		"token":        jwtToken,
		"username":     username,
		"role":         role,
		"userId":       userId,
		"isRegistered": isRegistered,
	})
}
