package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthHandlers struct {
	DB                *gorm.DB
	OauthStateString  string
	GoogleOauthConfig *oauth2.Config
	JWTSecret         []byte
}

func NewAuthHandlers(db *gorm.DB) *AuthHandlers {
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
	
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	
	return &AuthHandlers {
		DB: db,
		OauthStateString: oauthStateString,
		GoogleOauthConfig: googleOauthConfig,
		JWTSecret: jwtSecret,
	}
}

// RegisterRequest is a custom struct to handle incoming registration data.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register handles user registration
func (h *AuthHandlers) RegisterHandler(ctx *gin.Context) {
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
	ctx.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginRequest is a custom struct to handle incoming login data.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login
func (h *AuthHandlers) LoginHandler(ctx *gin.Context) {
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
	
	jwtToken, refreshToken, err := h.generateTokens(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}
	
	// Corrected for production: Secure is true, Domain is empty to default to backend domain.
	ctx.SetCookie("refresh_token", refreshToken, int(time.Hour * 24 * 30 / time.Second), "/", "", true, true)

	// Return the JWT token in the JSON response body.
	ctx.JSON(http.StatusOK, gin.H{"token": jwtToken})
}


// Generate JWT and Refresh Tokens
func (h *AuthHandlers) generateTokens(userID uint) (string, string, error) {
	// JWT Token
	jwtClaims := &model.UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)), // JWT expires in 15 minutes.
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedJwtToken, err := jwtToken.SignedString(h.JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshTokenBytes := make([]byte, 32)
	rand.Read(refreshTokenBytes)
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)
	
	// Create or update the refresh token in the database.
	refreshTokenDB := model.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // Refresh token expires in 30 days.
	}
	
	h.DB.Save(&refreshTokenDB)

	return signedJwtToken, refreshTokenString, nil
}

// handle refresh token request
func (h *AuthHandlers) RefreshTokenHandler(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token cookie"})
		return
	}
	
	var refreshTokenDB model.RefreshToken
	if err := h.DB.Where("token = ?", refreshToken).First(&refreshTokenDB).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	
	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		h.DB.Delete(&refreshTokenDB)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}
	
	jwtToken, newRefreshToken, err := h.generateTokens(refreshTokenDB.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		return
	}
	
	h.DB.Delete(&refreshTokenDB)
	
	ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour * 24 * 30 / time.Second), "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

// LogoutHandler invalidates the user's session.
func (h *AuthHandlers) LogoutHandler(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	var refreshTokenDB model.RefreshToken
	if err := h.DB.Where("token = ?", refreshToken).First(&refreshTokenDB).Error; err == nil {
		h.DB.Delete(&refreshTokenDB)
	}

	ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
