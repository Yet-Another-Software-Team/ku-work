package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type JWTHandlers struct {
	DB        *gorm.DB
	JWTSecret []byte
}

func NewJWTHandlers(db *gorm.DB) *JWTHandlers {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	return &JWTHandlers{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}

// Generate JWT and Refresh Tokens
func (h *JWTHandlers) GenerateTokens(userID string) (string, string, error) {
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
	_, rand_err := rand.Read(refreshTokenBytes)
	if rand_err != nil {
		return "", "", err
	}
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	// Create or update the refresh token in the database.
	refreshTokenDB := model.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // Refresh token expires in 30 days.
	}

	h.DB.Create(&refreshTokenDB)

	return signedJwtToken, refreshTokenString, nil
}

// @Summary Refresh JWT token
// @Description Renews an access token using a valid refresh token provided in a cookie. It returns a new JWT and user details, and sets a new refresh token cookie.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{token=string, username=string, role=string, userId=string} "Successfully refreshed token"
// @Failure 401 {object} object{error=string} "Unauthorized: Missing, invalid, or expired refresh token"
// @Failure 500 {object} object{error=string} "Internal Server Error: Failed to generate new tokens"
// @Router /auth/refresh [post]
func (h *JWTHandlers) RefreshTokenHandler(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token cookie"})
		return
	}

	var refreshTokenDB model.RefreshToken
	if err := h.DB.Where("token = ?", refreshToken).First(&refreshTokenDB).Error; err != nil {
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		h.DB.Unscoped().Delete(&refreshTokenDB)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Clear user's expired token
	h.DB.Unscoped().Where("user_id = ? AND expires_at < ?", refreshTokenDB.UserID, time.Now()).Delete(&model.RefreshToken{})

	h.DB.Unscoped().Delete(&refreshTokenDB)
	jwtToken, newRefreshToken, err := h.GenerateTokens(refreshTokenDB.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		return
	}

	role := helper.GetRole(refreshTokenDB.UserID, h.DB)
	username := helper.GetUsername(refreshTokenDB.UserID, role, h.DB)

	ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour*24*30/time.Second), "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": username,
		"role":     role,
		"userId":   refreshTokenDB.UserID,
	})
}

// @Summary Logout user
// @Description Invalidates the user's session by deleting their refresh token from the database and clearing the refresh token cookie.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string} "Logged out successfully"
// @Router /auth/logout [post]
func (h *JWTHandlers) LogoutHandler(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	var refreshTokenDB model.RefreshToken
	if err := h.DB.Where("token = ?", refreshToken).First(&refreshTokenDB).Error; err == nil {
		h.DB.Unscoped().Delete(&refreshTokenDB)
	}

	ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// HandleToken is a helper function to generate and return JWT and refresh tokens for a user.
func (h *JWTHandlers) HandleToken(user model.User) (string, string, error) {
	jwtToken, refreshToken, err := h.GenerateTokens(user.ID)
	if err != nil {
		return "", "", err
	}

	return jwtToken, refreshToken, nil
}
