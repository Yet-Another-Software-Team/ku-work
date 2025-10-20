package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
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
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes long for security")
	}
	return &JWTHandlers{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}

// hashToken hashes a refresh token using Argon2id before storage
// Uses OWASP recommended parameters for password hashing
func hashToken(token string) (string, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Argon2id parameters (OWASP recommended)
	// Memory: 64 MB, Iterations: 3, Parallelism: 2, Key length: 32 bytes
	hash := argon2.IDKey([]byte(token), salt, 3, 64*1024, 2, 32)

	// Encode salt and hash together for storage
	// Format: base64(salt) + "$" + base64(hash)
	encoded := base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash)
	return encoded, nil
}

// verifyToken verifies a token against a stored Argon2id hash
func verifyToken(token, storedHash string) (bool, error) {
	// Parse the stored hash to extract salt
	parts := make([]byte, len(storedHash))
	copy(parts, storedHash)

	var salt, hash []byte
	var err error

	// Find the separator
	sepIdx := -1
	for i, b := range parts {
		if b == '$' {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err = base64.RawStdEncoding.DecodeString(string(parts[:sepIdx]))
	if err != nil {
		return false, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(string(parts[sepIdx+1:]))
	if err != nil {
		return false, err
	}

	// Hash the provided token with the same salt
	computedHash := argon2.IDKey([]byte(token), salt, 3, 64*1024, 2, 32)

	// Constant-time comparison
	if len(hash) != len(computedHash) {
		return false, nil
	}

	match := true
	for i := range hash {
		if hash[i] != computedHash[i] {
			match = false
		}
	}

	return match, nil
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

	// Hash the refresh token before storing
	hashedToken, err := hashToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// Mark old refresh tokens as revoked for this user (for reuse detection)
	now := time.Now()
	h.DB.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)

	// Create the new refresh token in the database with hashed value
	refreshTokenDB := model.RefreshToken{
		UserID:    userID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30), // Refresh token expires in 30 days.
		RevokedAt: nil,                                 // Active token
	}

	if err := h.DB.Create(&refreshTokenDB).Error; err != nil {
		return "", "", err
	}

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
	clientIP := ctx.ClientIP()

	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		log.Printf("SECURITY: Missing refresh token from IP: %s", clientIP)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Find all tokens and verify against each (since we can't query by hash directly)
	var allTokens []model.RefreshToken
	if err := h.DB.Find(&allTokens).Error; err != nil {
		log.Printf("SECURITY: Database error during token lookup from IP: %s", clientIP)
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Find matching token
	var refreshTokenDB *model.RefreshToken
	for i := range allTokens {
		match, err := verifyToken(refreshToken, allTokens[i].Token)
		if err != nil {
			continue
		}
		if match {
			refreshTokenDB = &allTokens[i]
			break
		}
	}

	if refreshTokenDB == nil {
		log.Printf("SECURITY: Invalid refresh token from IP: %s", clientIP)
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check if token was revoked (reuse detection)
	if refreshTokenDB.RevokedAt != nil {
		log.Printf("SECURITY ALERT: Revoked token reuse detected! User: %s, IP: %s - Revoking all tokens",
			refreshTokenDB.UserID, clientIP)

		// Token reuse detected - revoke ALL tokens for this user
		now := time.Now()
		h.DB.Model(&model.RefreshToken{}).
			Where("user_id = ?", refreshTokenDB.UserID).
			Update("revoked_at", now)

		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check if token is expired
	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		log.Printf("SECURITY: Expired refresh token from IP: %s, User: %s", clientIP, refreshTokenDB.UserID)
		now := time.Now()
		h.DB.Model(refreshTokenDB).Update("revoked_at", now)
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Token is valid - generate new tokens
	jwtToken, newRefreshToken, err := h.GenerateTokens(refreshTokenDB.UserID)
	if err != nil {
		log.Printf("ERROR: Failed to generate new tokens for user: %s, IP: %s - %v",
			refreshTokenDB.UserID, clientIP, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		return
	}

	role := helper.GetRole(refreshTokenDB.UserID, h.DB)
	username := helper.GetUsername(refreshTokenDB.UserID, role, h.DB)

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour*24*30/time.Second), "/", "", true, true)

	log.Printf("INFO: Token refreshed successfully for user: %s, IP: %s", refreshTokenDB.UserID, clientIP)

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
	clientIP := ctx.ClientIP()

	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("refresh_token", "", -1, "/", "", true, true)
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	// Find and revoke the token
	var allTokens []model.RefreshToken
	if err := h.DB.Find(&allTokens).Error; err == nil {
		for i := range allTokens {
			match, err := verifyToken(refreshToken, allTokens[i].Token)
			if err == nil && match {
				now := time.Now()
				h.DB.Model(&allTokens[i]).Update("revoked_at", now)
				log.Printf("INFO: User logged out: %s, IP: %s", allTokens[i].UserID, clientIP)
				break
			}
		}
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
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
