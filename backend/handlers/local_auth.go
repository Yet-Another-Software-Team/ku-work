package handlers

import (
	"net/http"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LocalAuthHandlers struct {
	DB          *gorm.DB
	JWTHandlers *JWTHandlers
}

func NewLocalAuthHandlers(db *gorm.DB) *LocalAuthHandlers {
	return &LocalAuthHandlers{
		DB:          db,
		JWTHandlers: NewJWTHandlers(db),
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
	newUser := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := transaction.Create(&newUser).Error; err != nil {
		transaction.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	transaction.Commit()

	ctx = h.JWTHandlers.HandleToken(ctx, newUser)
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

	ctx = h.JWTHandlers.HandleToken(ctx, user)
}
