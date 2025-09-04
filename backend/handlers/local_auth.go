package handlers

import (
	"net/http"
	"time"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LocalAuthHandlers struct {
	DB          *gorm.DB
	JWTHandlers *JWTHandlers
}

func NewLocalAuthHandlers(db *gorm.DB, jwtHandlers *JWTHandlers) *LocalAuthHandlers {
	return &LocalAuthHandlers{
		DB:          db,
		JWTHandlers: jwtHandlers,
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
	var count int64 = 0
	h.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)

	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash the password before saving it.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the new user.
	newUser := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := h.DB.Create(&newUser).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	jwtToken, refreshToken, err := h.JWTHandlers.HandleToken(newUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	maxAge := int(time.Hour * 24 * 30 / time.Second)
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"username": newUser.Username,
		"isStudent": false,
		"isCompany": false,
	})
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

	jwtToken, refreshToken, err := h.JWTHandlers.HandleToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	maxAge := int(time.Hour * 24 * 30 / time.Second)
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"username": user.Username,
		"isStudent": false,
		"isCompany": false,
	})
}
