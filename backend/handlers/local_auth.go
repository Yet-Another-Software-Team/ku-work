package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type LocalAuthHandlers struct {
	DB                *gorm.DB
	OauthStateString  string
	GoogleOauthConfig *oauth2.Config
	JWTHandler        *JWTHandlers
}

func NewLocalAuthHandlers(db *gorm.DB) *LocalAuthHandlers {
	// TODO: MOVE THIS TO OAUTH HANDLERS
	b := make([]byte, 16)
	rand.Read(b)
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
	
	return &LocalAuthHandlers {
		DB: db,
		OauthStateString: oauthStateString,
		GoogleOauthConfig: googleOauthConfig,
		JWTHandler: NewJWTHandlers(db),
	}
}

// struct to handle incoming registration data.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register handles user registration
func (h *LocalAuthHandlers) RegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Check if a user with the same username already exists.
	var existingUser model.User
	if err := h.DB.Model(&model.User{}).Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}
	
	// Hash the password before saving it.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// start db transaction
	transaction := h.DB.Begin()
	
	// Create the new user.
	newUser := model.User {
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}
	
	if err := transaction.Create(&newUser).Error; err != nil {
		transaction.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	transaction.Commit()
	
	ctx = h.JWTHandler.HandleToken(ctx, newUser)
}

// struct to handle incoming login data.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login
func (h *LocalAuthHandlers) LoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	var user model.User
	if err := h.DB.Model(&model.User{}).Where("username = ?", req.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	// Compare the provided password with the stored hashed password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	ctx = h.JWTHandler.HandleToken(ctx, user)
}
