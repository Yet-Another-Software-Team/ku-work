package handlers

import (
	"mime/multipart"
	"net/http"
	"time"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	Username string                `form:"username" binding:"required,max=256"`
	Password string                `form:"password" binding:"required,min=8"`
	Email    string                `form:"email" binding:"required,max=100"`
	Phone    string                `form:"phone" binding:"required,max=20"`
	Address  string                `form:"address" binding:"required,max=200"`
	City     string                `form:"city" binding:"required,max=100"`
	Country  string                `form:"country" binding:"required,max=100"`
	Photo    *multipart.FileHeader `form:"photo" binding:"required"`
	Banner   *multipart.FileHeader `form:"banner" binding:"required"`
}

// Register handles user registration
func (h *LocalAuthHandlers) CompanyRegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.MustBindWith(&req, binding.FormMultipart); err != nil {
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

	tx := h.DB.Begin()

	defer tx.Rollback()

	// Create the new user.
	newUser := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := tx.Create(&newUser).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create file handler
	fileHandler := NewFileHandlers(tx)

	// Create Company
	photoID, err := fileHandler.SaveFile(ctx, newUser.ID, req.Photo, model.FileCategoryImage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bannerID, err := fileHandler.SaveFile(ctx, newUser.ID, req.Banner, model.FileCategoryImage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newCompany := model.Company{
		UserID:   newUser.ID,
		Email:    req.Email,
		Phone:    req.Phone,
		PhotoID:  photoID,
		BannerID: bannerID,
		Address:  req.Address,
		City:     req.City,
		Country:  req.Country,
	}

	if err := tx.Create(&newCompany).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Company Data"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
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
		"token":     jwtToken,
		"username":  newUser.Username,
		"isStudent": false,
		"isCompany": true, // This registration flow only used for Company.
	})
}

// struct to handle incoming login data.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login
func (h *LocalAuthHandlers) CompanyLoginHandler(ctx *gin.Context) {
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

	var companyCount int64
	h.DB.Model(&model.Company{}).Where("user_id = ?", user.ID).Count(&companyCount)
	if companyCount == 0 {
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
		"token":     jwtToken,
		"username":  user.Username,
		"isStudent": false,
		"isCompany": true,
	})
}

// Login handles user login
func (h *LocalAuthHandlers) AdminLoginHandler(ctx *gin.Context) {
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

	var adminCount int64
	h.DB.Model(&model.Admin{}).Where("user_id = ?", user.ID).Count(&adminCount)
	if adminCount == 0 {
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
		"token":     jwtToken,
		"username":  user.Username,
		"isStudent": false,
		"isCompany": false,
	})
}
