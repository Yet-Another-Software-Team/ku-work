package handlers

import (
	"log"
	"mime/multipart"
	"net/http"

	"ku-work/backend/helper"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type LocalAuthHandlers struct {
	DB          *gorm.DB
	JWTHandlers *JWTHandlers
	Service     *services.AuthService
}

func NewLocalAuthHandlers(db *gorm.DB, jwtHandlers *JWTHandlers) *LocalAuthHandlers {
	return &LocalAuthHandlers{
		DB:          db,
		JWTHandlers: jwtHandlers,
		// Inject the handler-level SaveFile function into the service to avoid import cycles.
		// handlers.SaveFile matches the services.SaveFileFunc signature.
		Service: services.NewAuthService(db, jwtHandlers, SaveFile),
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

// LoginRequest defines incoming login payload.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CompanyRegisterHandler handles the HTTP request and delegates to the shared AuthService.
func (h *LocalAuthHandlers) CompanyRegisterHandler(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.MustBindWith(&req, binding.FormMultipart); err != nil {
		log.Printf("Error binding request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Map handler-level DTO to service-level DTO to avoid importing handler types into services.
	svcInput := services.RegisterCompanyInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Website:  req.Website,
		Phone:    req.Phone,
		Address:  req.Address,
		City:     req.City,
		Country:  req.Country,
		Photo:    req.Photo,
		Banner:   req.Banner,
		AboutUs:  req.AboutUs,
	}

	createdUser, _, jwtToken, refreshToken, err := h.Service.RegisterCompany(ctx, svcInput)
	if err != nil {
		// map known service errors to HTTP responses
		if err == services.ErrUsernameExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		if err == services.ErrInvalidWebsite {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website URL"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	maxAge := services.CookieMaxAge()
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": createdUser.Username,
		"role":     helper.Company,
		"userId":   createdUser.ID,
	})
}

// CompanyLoginHandler delegates company login to the shared AuthService.
func (h *LocalAuthHandlers) CompanyLoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, jwtToken, refreshToken, err := h.Service.CompanyLogin(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	maxAge := services.CookieMaxAge()
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": user.Username,
		"role":     helper.Company,
		"userId":   user.ID,
	})
}

// AdminLoginHandler delegates admin login to the shared AuthService.
func (h *LocalAuthHandlers) AdminLoginHandler(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, jwtToken, refreshToken, err := h.Service.AdminLogin(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	maxAge := services.CookieMaxAge()
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": user.Username,
		"role":     helper.Admin,
		"userId":   user.ID,
	})
}
