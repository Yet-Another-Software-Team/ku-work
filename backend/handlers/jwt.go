package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	// Hash the provided token with the same salt
	computedHash := argon2.IDKey([]byte(token), salt, 3, 64*1024, 2, 32)

	// Constant-time comparison using crypto/subtle
	match := subtle.ConstantTimeCompare(hash, computedHash) == 1

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

	// Generate token selector (non-sensitive, for fast lookup)
	selectorBytes := make([]byte, 16)
	_, err = rand.Read(selectorBytes)
	if err != nil {
		return "", "", err
	}
	selector := base64.URLEncoding.EncodeToString(selectorBytes)

	// Generate refresh token validator (sensitive part)
	validatorBytes := make([]byte, 32)
	_, err = rand.Read(validatorBytes)
	if err != nil {
		return "", "", err
	}
	validator := base64.URLEncoding.EncodeToString(validatorBytes)

	// Hash the validator before storing
	hashedValidator, err := hashToken(validator)
	if err != nil {
		return "", "", err
	}

	// Get session limit from environment variable (default: 10)
	maxSessions := 10
	if maxSessionsStr := os.Getenv("MAX_SESSIONS_PER_USER"); maxSessionsStr != "" {
		if parsed, err := strconv.Atoi(maxSessionsStr); err == nil && parsed > 0 {
			maxSessions = parsed
		}
	}

	// Limit concurrent sessions per user
	// Count active tokens for this user
	var activeTokenCount int64
	h.DB.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Count(&activeTokenCount)

	// If at or over limit, revoke the oldest tokens
	if activeTokenCount >= int64(maxSessions) {
		// Find the oldest active tokens to revoke
		var oldestTokens []model.RefreshToken
		tokensToRevoke := int(activeTokenCount) - (maxSessions - 1) // Keep (max-1), revoke excess to make room for the new one
		h.DB.Where("user_id = ? AND revoked_at IS NULL", userID).
			Order("created_at ASC").
			Limit(tokensToRevoke).
			Find(&oldestTokens)

		// Revoke the oldest tokens
		if len(oldestTokens) > 0 {
			now := time.Now()
			tokenIDs := make([]uint, len(oldestTokens))
			for i, token := range oldestTokens {
				tokenIDs[i] = token.ID
			}
			h.DB.Model(&model.RefreshToken{}).
				Where("id IN ?", tokenIDs).
				Update("revoked_at", now)
			log.Printf("INFO: Revoked %d oldest tokens for user %s (session limit: %d)", len(oldestTokens), userID, maxSessions)
		}
	}

	// Create the new refresh token in the database
	refreshTokenDB := model.RefreshToken{
		UserID:        userID,
		TokenSelector: selector,
		Token:         hashedValidator,
		ExpiresAt:     time.Now().Add(time.Hour * 24 * 30), // Refresh token expires in 30 days.
		RevokedAt:     nil,                                 // Active token
	}

	if err := h.DB.Create(&refreshTokenDB).Error; err != nil {
		return "", "", err
	}

	// Return combined token: selector:validator
	combinedToken := selector + ":" + validator

	return signedJwtToken, combinedToken, nil
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

	// Split token into selector and validator
	parts := strings.SplitN(refreshToken, ":", 2)
	if len(parts) != 2 {
		log.Printf("SECURITY: Invalid refresh token format from IP: %s", clientIP)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	selector := parts[0]
	validator := parts[1]

	var refreshTokenDB model.RefreshToken
	if err := h.DB.Where("token_selector = ?", selector).First(&refreshTokenDB).Error; err != nil {
		log.Printf("SECURITY: Invalid refresh token from IP: %s", clientIP)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Verify the validator against stored hash
	match, err := verifyToken(validator, refreshTokenDB.Token)
	if err != nil || !match {
		log.Printf("SECURITY: Invalid refresh token from IP: %s", clientIP)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
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

		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check if token is expired
	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		log.Printf("SECURITY: Expired refresh token from IP: %s, User: %s", clientIP, refreshTokenDB.UserID)
		now := time.Now()
		h.DB.Model(&refreshTokenDB).Update("revoked_at", now)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
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

	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour*24*30/time.Second), "/", "", helper.GetCookieSecure(), true)

	log.Printf("INFO: Token refreshed successfully for user: %s, IP: %s", refreshTokenDB.UserID, clientIP)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": username,
		"role":     role,
		"userId":   refreshTokenDB.UserID,
	})
}

// @Summary Logout user
// @Description Invalidates the user's session by revoking the refresh token and clearing the cookie. This is a best-effort operation that always succeeds.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string} "Logged out successfully"
// @Router /auth/logout [post]
func (h *JWTHandlers) LogoutHandler(ctx *gin.Context) {
	clientIP := ctx.ClientIP()

	refreshToken, err := ctx.Cookie("refresh_token")
	if err == nil && refreshToken != "" {
		// Split token into selector and validator
		parts := strings.SplitN(refreshToken, ":", 2)
		if len(parts) == 2 {
			selector := parts[0]
			validator := parts[1]

			// Fast O(1) lookup using selector
			var tokenDB model.RefreshToken
			if err := h.DB.Where("token_selector = ? AND revoked_at IS NULL", selector).First(&tokenDB).Error; err == nil {
				// Verify the validator
				match, err := verifyToken(validator, tokenDB.Token)
				if err == nil && match {
					now := time.Now()
					h.DB.Model(&tokenDB).Update("revoked_at", now)
					log.Printf("INFO: User logged out and token revoked: %s, IP: %s", tokenDB.UserID, clientIP)
				}
			}
		}
	}

	// Always clear the cookie, even if token revocation failed
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
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
