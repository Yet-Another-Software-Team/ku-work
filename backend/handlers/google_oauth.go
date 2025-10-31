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

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

// OauthHandlers is a thin HTTP layer that delegates work to an internal service.
type OauthHandlers struct {
	service *oauthService
}

type oauthService struct {
	DB                *gorm.DB
	GoogleOauthConfig *oauth2.Config
	JWTHandlers       *JWTHandlers
	HTTPClient        *http.Client
}

func newOauthService(db *gorm.DB, jwtHandlers *JWTHandlers, cfg *oauth2.Config) *oauthService {
	return &oauthService{
		DB:                db,
		GoogleOauthConfig: cfg,
		JWTHandlers:       jwtHandlers,
		HTTPClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

func NewOAuthHandlers(db *gorm.DB, jwtHandlers *JWTHandlers) *OauthHandlers {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate random oauth state: %v", err)
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

	_ = oauthStateString // kept in case future use is needed

	return &OauthHandlers{
		service: newOauthService(db, jwtHandlers, googleOauthConfig),
	}
}

type oauthToken struct {
	Code string `json:"code"`
}

type userInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

// ExchangeCode exchanges an authorization code for an access token.
func (s *oauthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.GoogleOauthConfig.Exchange(ctx, code)
}

// FetchUserInfo retrieves user info from Google using an access token.
func (s *oauthService) FetchUserInfo(token *oauth2.Token) (userInfo, int, error) {
	var ui userInfo

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return ui, http.StatusInternalServerError, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return ui, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ui, http.StatusUnauthorized, fmt.Errorf("access token invalid or expired")
	}

	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		return ui, http.StatusInternalServerError, err
	}
	return ui, http.StatusOK, nil
}

// EnsureUser creates or updates local user records based on external user info.
func (s *oauthService) EnsureUser(ui userInfo) (model.User, model.GoogleOAuthDetails, int, error) {
	var oauthDetail model.GoogleOAuthDetails
	var status = http.StatusOK

	var count int64
	if err := s.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", ui.ID).Count(&count).Error; err != nil {
		return model.User{}, oauthDetail, http.StatusInternalServerError, err
	}

	if count == 0 {
		var newUser model.User
		if err := s.DB.FirstOrCreate(&newUser, model.User{
			Username: ui.Email,
			UserType: "oauth",
		}).Error; err != nil {
			return model.User{}, oauthDetail, http.StatusInternalServerError, err
		}

		oauthDetail = model.GoogleOAuthDetails{
			UserID:     newUser.ID,
			ExternalID: ui.ID,
			FirstName:  ui.GivenName,
			LastName:   ui.FamilyName,
			Email:      ui.Email,
		}
		if err := s.DB.Create(&oauthDetail).Error; err != nil {
			return model.User{}, oauthDetail, http.StatusInternalServerError, err
		}
		status = http.StatusCreated
	} else {
		// load existing and update fields
		if err := s.DB.Model(&model.GoogleOAuthDetails{}).Where("external_id = ?", ui.ID).First(&oauthDetail).Error; err != nil {
			return model.User{}, oauthDetail, http.StatusInternalServerError, err
		}
		oauthDetail.FirstName = ui.GivenName
		oauthDetail.LastName = ui.FamilyName
		oauthDetail.Email = ui.Email
		if err := s.DB.Save(&oauthDetail).Error; err != nil {
			return model.User{}, oauthDetail, http.StatusInternalServerError, err
		}
	}

	var user model.User
	if err := s.DB.Model(&user).Where("id = ?", oauthDetail.UserID).First(&user).Error; err != nil {
		return model.User{}, oauthDetail, http.StatusInternalServerError, err
	}

	return user, oauthDetail, status, nil
}

// GoogleOauthHandler handles the HTTP request; it keeps the handler small and delegates the work.
func (h *OauthHandlers) GoogleOauthHandler(ctx *gin.Context) {
	var req oauthToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	token, err := h.service.ExchangeCode(context.Background(), req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ui, code, err := h.service.FetchUserInfo(token)
	if err != nil {
		ctx.JSON(code, gin.H{"error": err.Error()})
		return
	}

	user, oauthDetail, status, err := h.service.EnsureUser(ui)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jwtToken, refreshToken, err := h.service.JWTHandlers.HandleToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate jwt token"})
		return
	}

	maxAge := int(time.Hour * 24 * 30 / time.Second)
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	username := oauthDetail.FirstName + " " + oauthDetail.LastName
	role := helper.Viewer
	isRegistered := false

	if status == http.StatusOK {
		var count int64
		if err := h.service.DB.Model(&model.Student{}).Where("user_id = ?", user.ID).Count(&count).Error; err == nil {
			isRegistered = count > 0
			if count > 0 {
				var student model.Student
				if err := h.service.DB.Model(&student).Where("user_id = ?", user.ID).First(&student).Error; err == nil {
					if student.ApprovalStatus == model.StudentApprovalAccepted {
						role = helper.Student
					}
				}
			}
		}
	}

	ctx.JSON(status, gin.H{
		"token":        jwtToken,
		"username":     username,
		"role":         role,
		"userId":       user.ID,
		"isRegistered": isRegistered,
	})
}
