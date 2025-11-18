package handlers

import (
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	Website  string                `form:"website" binding:"max=200"`
	Phone    string                `form:"phone" binding:"required,max=20"`
	Address  string                `form:"address" binding:"required,max=200"`
	City     string                `form:"city" binding:"required,max=100"`
	Country  string                `form:"country" binding:"required,max=100"`
	Photo    *multipart.FileHeader `form:"photo" binding:"required"`
	Banner   *multipart.FileHeader `form:"banner" binding:"required"`
	AboutUs  string                `form:"about" binding:"max=16384"`
}

// @Summary Register a new company
// @Description Handles the registration of a new company account. It takes company details and credentials, creates a new user and company profile, and returns JWT tokens upon successful registration.
// @Tags Authentication
// @Accept multipart/form-data
// @Produce json
// @Param username formData string true "Company's username"
// @Param password formData string true "Password (min 8 characters)"
// @Param email formData string true "Company's contact email"
// @Param website formData string false "Company's website URL"
// @Param phone formData string true "Company's contact phone number"
// @Param address formData string true "Company's physical address"
// @Param city formData string true "City"
// @Param country formData string true "Country"
// @Param photo formData file true "Company's profile photo"
// @Param banner formData file true "Company's banner image"
// @Param about formData string false "About the company"
// @Success 200 {object} object{token=string, username=string, role=string, userId=string} "Registration successful"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 409 {object} object{error=string} "Conflict: Username already exists"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/company/register [post]
func (h *LocalAuthHandlers) CompanyRegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.MustBindWith(&req, binding.FormMultipart); err != nil {
		log.Printf("Error binding request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if a user with the same username already exists.
	var count int64 = 0
	h.DB.Model(&model.User{}).Where("username = ? AND user_type = ?", req.Username, "company").Count(&count)

	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash the password before saving it.
	hashedPassword, err := helper.HashPassword(req.Password)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	tx := h.DB.Begin()

	defer tx.Rollback()

	// Create the new user.
	newUser := model.User{
		Username:     req.Username,
		UserType:     "company",
		PasswordHash: hashedPassword,
	}

	if err := tx.Create(&newUser).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create Company
	photo, err := SaveFile(ctx, tx, newUser.ID, req.Photo, model.FileCategoryImage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	banner, err := SaveFile(ctx, tx, newUser.ID, req.Banner, model.FileCategoryImage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Website != "" {
		// Parse website URL and check for invalid URL (Only Basic Check)
		parsedURL, err := url.Parse(req.Website)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website URL"})
			return
		}
	}

	newCompany := model.Company{
		UserID:   newUser.ID,
		Email:    req.Email,
		Phone:    req.Phone,
		PhotoID:  photo.ID,
		BannerID: banner.ID,
		Address:  req.Address,
		City:     req.City,
		AboutUs:  req.AboutUs,
		Country:  req.Country,
		Website:  req.Website,
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
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": newUser.Username,
		"role":     helper.Company,
		"userId":   newUser.ID,
	})
}

// struct to handle incoming login data.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Summary Company login
// @Description Authenticates a company user with their username and password. On successful authentication, it returns a JWT token for session management and sets a refresh token in a cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login Credentials"
// @Success 200 {object} object{token=string, username=string, role=string, userId=string} "Login successful"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized: Invalid credentials"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/company/login [post]
func (h *LocalAuthHandlers) CompanyLoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user model.User
	if err := h.DB.Model(&model.User{}).Unscoped().Where("username = ? AND user_type = ?", req.Username, "company").First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the provided password with the stored hashed password.
	match, err := helper.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
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
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":         jwtToken,
		"username":      user.Username,
		"role":          helper.Company,
		"userId":        user.ID,
		"isDeactivated": user.DeletedAt.Valid,
	})
}

// @Summary Admin login
// @Description Authenticates an admin user with their username and password. On successful authentication, it returns a JWT token for session management and sets a refresh token in a cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login Credentials"
// @Success 200 {object} object{token=string, username=string, role=string, userId=string} "Login successful"
// @Failure 400 {object} object{error=string} "Bad Request"
// @Failure 401 {object} object{error=string} "Unauthorized: Invalid credentials"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /auth/admin/login [post]
func (h *LocalAuthHandlers) AdminLoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user model.User
	if err := h.DB.Model(&model.User{}).Where("username = ? AND user_type = ?", req.Username, "admin").First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the provided password with the stored hashed password.
	match, err := helper.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
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
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": user.Username,
		"role":     helper.Admin,
		"userId":   user.ID,
	})
}
