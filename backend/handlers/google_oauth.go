package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type OauthHandlers struct {
	DB                *gorm.DB
	OauthStateString  string
	GoogleOauthConfig *oauth2.Config
	JWTHandlers       *JWTHandlers
}

func NewOAuthHandlers(db *gorm.DB, jwtHandlers *JWTHandlers) *OauthHandlers {
	b := make([]byte, 16)
	_, rand_err := rand.Read(b)
	if rand_err != nil {
		slog.Error("Failed to generate random oauth state", "message", rand_err)
		os.Exit(1)
	}
	oauthStateString := base64.URLEncoding.EncodeToString(b)

	googleOauthConfig := &oauth2.Config{
		RedirectURL:  "postmessage",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" {
		slog.Error("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables are not set")
		os.Exit(1)
	}

	return &OauthHandlers{
		DB:                db,
		OauthStateString:  oauthStateString,
		GoogleOauthConfig: googleOauthConfig,
		JWTHandlers:       jwtHandlers,
	}
}

type oauthToken struct {
	Code string `json:"code"`
}

// @Summary Handle Google OAuth login
// @Description Handles the server-side flow for Google OAuth2. It receives an authorization code from the client, exchanges it for a token with Google, fetches the user's profile information, and then either creates a new user account or logs in an existing user. On success, it returns a JWT token and sets a refresh token cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param code body oauthToken true "Google Authorization Code"
// @Success 200 {object} object{token=string, username=string, role=string, userId=string, isRegistered=bool} "Login successful"
// @Success 201 {object} object{token=string, username=string, role=string, userId=string, isRegistered=bool} "User registration successful"
// @Failure 400 {object} object{error=string} "Bad Request: Authorization code is required"
// @Failure 401 {object} object{error=string} "Unauthorized: Invalid access token"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/google/login [post]
func (h *OauthHandlers) GoogleOauthHandler(ctx *gin.Context) {
	// Validate request
	var req oauthToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	// Exchange authorization code for access token
	exchange_ctx := context.Background()
	token, err := h.GoogleOauthConfig.Exchange(exchange_ctx, req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Create API request to exchange access token for user info
	api_req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	api_req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	api_res, err := client.Do(api_req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Close API body
	defer func() {
		if err := api_res.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// Check if API request was successful
	if api_res.StatusCode != http.StatusOK {
		slog.Warn("User attempt to login using OAuth", "ip", ctx.ClientIP())
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Access Token is invalid, expired or insufficient"})
		return
	}

	// Parse user info from API response
	type UserInfo struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}

	var userInfo UserInfo
	if err := json.NewDecoder(api_res.Body).Decode(&userInfo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if this user already exists on the DB via their external ID.
	var userCount int64
	h.DB.Model(&model.GoogleOAuthDetails{}).Unscoped().Where("external_id = ?", userInfo.ID).Count(&userCount)

	var oauthDetail model.GoogleOAuthDetails

	// User is not exist, registerings
	status := http.StatusOK

	if userCount == 0 {
		var newUser model.User
		h.DB.Unscoped().FirstOrCreate(&newUser, model.User{
			Username: userInfo.Email,
			UserType: "oauth",
		})

		oauthDetail = model.GoogleOAuthDetails{
			UserID:     newUser.ID,
			ExternalID: userInfo.ID,
			FirstName:  userInfo.GivenName,
			LastName:   userInfo.FamilyName,
			Email:      userInfo.Email,
		}
		h.DB.Create(&oauthDetail)
		status = http.StatusCreated
	}

	// Update user details if necessary, to keep them up-to-date
	if oauthDetail.UserID != "" {
		h.DB.Model(&oauthDetail).Updates(model.GoogleOAuthDetails{
			FirstName: userInfo.GivenName,
			LastName:  userInfo.FamilyName,
			Email:     userInfo.Email,
		})
	}

	// Get updated oauthDetail
	h.DB.Model(&model.GoogleOAuthDetails{}).Unscoped().Where("external_id = ?", userInfo.ID).First(&oauthDetail)

	// getUser model and return JWT Token to frontend.
	var user model.User
	h.DB.Model(&user).Unscoped().Where("id = ?", oauthDetail.UserID).First(&user)

	//Return JWT Token to context
	jwtToken, refreshToken, err := h.JWTHandlers.HandleToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	maxAge := int(time.Hour * 24 * 30 / time.Second)
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie(helper.GetRefreshCookieName(), refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	username := oauthDetail.FirstName + " " + oauthDetail.LastName
	role := helper.Viewer
	isRegistered := false
	if status == http.StatusOK {
		// Check if user is a valid and approved student
		var count int64
		h.DB.Model(&model.Student{}).Where("user_id = ?", user.ID).Count(&count)
		isRegistered = count > 0
		if count > 0 {
			var student model.Student
			h.DB.Model(&student).Where("user_id = ?", user.ID).First(&student)
			if student.ApprovalStatus == model.StudentApprovalAccepted {
				role = helper.Student
			}
		}
	}

	slog.Info("User logged in using OAuth", "user_id", user.ID, "ip", ctx.ClientIP())

	ctx.JSON(status, gin.H{
		"token":         jwtToken,
		"username":      username,
		"role":          role,
		"userId":        user.ID,
		"isRegistered":  isRegistered, // To tell frontend whether user is registered or not
		"isDeactivated": user.DeletedAt.Valid,
	})
}
