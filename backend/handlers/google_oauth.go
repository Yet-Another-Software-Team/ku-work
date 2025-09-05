package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
		log.Fatalf("Failed to generate random oauth state %v", rand_err)
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
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables are not set")
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

func (h *OauthHandlers) GoogleOauthHandler(ctx *gin.Context) {
	var req oauthToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	exchange_ctx := context.Background()
	token, err := h.GoogleOauthConfig.Exchange(exchange_ctx, req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
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

	if api_res.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Access Token is invalid, expired or insufficient"})
		return
	}

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
	h.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", userInfo.ID).Count(&userCount)

	var oauthDetail model.GoogleOAuthDetails

	// User is not exist, registerings
	status := http.StatusOK

	if userCount == 0 {
		var newUser model.User
		h.DB.FirstOrCreate(&newUser, model.User{
			Username: userInfo.Email,
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

	// Update user details if necessary
	if oauthDetail.UserID != "" {
		h.DB.Model(&oauthDetail).Updates(model.GoogleOAuthDetails{
			FirstName: userInfo.GivenName,
			LastName:  userInfo.FamilyName,
			Email:     userInfo.Email,
		})
	}

	// Get updated oauthDetail
	h.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", userInfo.ID).First(&oauthDetail)

	// getUser model and return JWT Token to frontend.
	var user model.User
	h.DB.Model(&user).Where("id = ?", oauthDetail.UserID).First(&user)

	//Return JWT Token to context
	jwtToken, refreshToken, err := h.JWTHandlers.HandleToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	maxAge := int(time.Hour * 24 * 30 / time.Second)
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", true, true)

	username := oauthDetail.FirstName + " " + oauthDetail.LastName
	isStudent := false

	if status == http.StatusOK {
		// Check if user is a valid and approved student
		var count int64
		h.DB.Model(&model.Student{}).Where("user_id = ? AND approved = ?", user.ID, true).Count(&count)
		isStudent = count > 0
	}

	ctx.JSON(status, gin.H{
		"token":     jwtToken,
		"username":  username,
		"isCompany": false, // Company doesn't have OAuth Login option
		"isStudent": isStudent,
	})

}
