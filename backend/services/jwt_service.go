package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// JWTService encapsulates JWT and refresh-token related operations.
//
// This service is intended to hold the core token generation, hashing and
// refresh-token persistence logic so that HTTP handlers can delegate to it
// without carrying all cryptographic/storage details.
type JWTService struct {
	DB                *gorm.DB
	RedisClient       *redis.Client
	RevocationService *JWTRevocationService
	JWTSecret         []byte
}

// NewJWTService constructs a new JWTService wired with Redis/GORM dependencies.
// It reads JWT_SECRET from the environment (and requires it to be at least 32 bytes).
func NewJWTService(db *gorm.DB, redisClient *redis.Client) *JWTService {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes long for security")
	}

	if redisClient == nil {
		log.Fatal("Redis client is required for JWT revocation")
	}

	revocationService := NewJWTRevocationService(redisClient)

	return &JWTService{
		DB:                db,
		RedisClient:       redisClient,
		RevocationService: revocationService,
		JWTSecret:         jwtSecret,
	}
}

// hashToken hashes a refresh token validator using Argon2id.
// Format used for storage: base64(salt) + "$" + base64(hash)
func hashToken(token string) (string, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Argon2id parameters (OWASP-recommended-ish)
	// Memory: 64 MB, Iterations: 3, Parallelism: 1, Key length: 32 bytes
	hash := argon2.IDKey([]byte(token), salt, 3, 64*1024, 1, 32)

	encoded := base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash)
	return encoded, nil
}

// verifyToken verifies a validator against a stored Argon2id hash.
func verifyToken(token, storedHash string) (bool, error) {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid stored hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey([]byte(token), salt, 3, 64*1024, 1, 32)
	match := subtle.ConstantTimeCompare(hash, computed) == 1
	return match, nil
}

// GenerateTokens creates a signed JWT access token and a secure refresh token.
//
// The refresh token returned to the client is in the form "selector:validator".
// The validator is hashed before storage; the selector is stored in plaintext
// to allow O(1) lookup.
func (s *JWTService) GenerateTokens(userID string) (string, string, error) {
	// Generate unique JTI (JWT ID)
	jti := uuid.New().String()

	// Build JWT claims (short lived access token)
	jwtClaims := &model.UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedJwt, err := token.SignedString(s.JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Generate selector (public part) and validator (secret part)
	selectorBytes := make([]byte, 16)
	if _, err := rand.Read(selectorBytes); err != nil {
		return "", "", err
	}
	selector := base64.URLEncoding.EncodeToString(selectorBytes)

	validatorBytes := make([]byte, 32)
	if _, err := rand.Read(validatorBytes); err != nil {
		return "", "", err
	}
	validator := base64.URLEncoding.EncodeToString(validatorBytes)

	// Hash validator before persisting
	hashedValidator, err := hashToken(validator)
	if err != nil {
		return "", "", err
	}

	// Determine max sessions per user (default 10)
	maxSessions := 10
	if ms := os.Getenv("MAX_SESSIONS_PER_USER"); ms != "" {
		if parsed, perr := strconv.Atoi(ms); perr == nil && parsed > 0 {
			maxSessions = parsed
		}
	}

	// Count active tokens for this user
	var activeTokenCount int64
	if err := s.DB.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Count(&activeTokenCount).Error; err != nil {
		return "", "", err
	}

	// If at or over limit, revoke oldest tokens to make room
	if activeTokenCount >= int64(maxSessions) {
		tokensToRevoke := int(activeTokenCount) - (maxSessions - 1)
		if tokensToRevoke > 0 {
			var oldest []model.RefreshToken
			if err := s.DB.Where("user_id = ? AND revoked_at IS NULL", userID).
				Order("created_at ASC").
				Limit(tokensToRevoke).
				Find(&oldest).Error; err == nil && len(oldest) > 0 {
				now := time.Now()
				ids := make([]uint, 0, len(oldest))
				for _, t := range oldest {
					ids = append(ids, t.ID)
				}
				_ = s.DB.Model(&model.RefreshToken{}).
					Where("id IN ?", ids).
					Update("revoked_at", now).Error
			}
		}
	}

	// Create refresh token DB record
	rt := model.RefreshToken{
		UserID:        userID,
		TokenSelector: selector,
		Token:         hashedValidator,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		RevokedAt:     nil,
	}

	if err := s.DB.Create(&rt).Error; err != nil {
		return "", "", err
	}

	combined := selector + ":" + validator
	return signedJwt, combined, nil
}

// HandleToken is a convenience wrapper to generate tokens for a model.User.
func (s *JWTService) HandleToken(user model.User) (string, string, error) {
	return s.GenerateTokens(user.ID)
}

// VerifyRefreshToken takes a combined token (selector:validator) and verifies it against storage.
// Returns the stored model.RefreshToken entry when verification succeeds.
func (s *JWTService) VerifyRefreshToken(combined string) (*model.RefreshToken, error) {
	parts := strings.SplitN(combined, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid refresh token format")
	}
	selector := parts[0]
	validator := parts[1]

	var stored model.RefreshToken
	if err := s.DB.Where("token_selector = ?", selector).First(&stored).Error; err != nil {
		return nil, err
	}

	match, err := verifyToken(validator, stored.Token)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, fmt.Errorf("refresh token validation failed")
	}
	return &stored, nil
}

// RevokeRefreshTokenBySelector revokes a refresh token record identified by selector.
func (s *JWTService) RevokeRefreshTokenBySelector(selector string) error {
	var t model.RefreshToken
	if err := s.DB.Where("token_selector = ? AND revoked_at IS NULL", selector).First(&t).Error; err != nil {
		return err
	}
	now := time.Now()
	return s.DB.Model(&t).Update("revoked_at", now).Error
}

// RevokeAllRefreshTokensForUser revokes all active refresh tokens for a user.
func (s *JWTService) RevokeAllRefreshTokensForUser(userID string) error {
	now := time.Now()
	return s.DB.Model(&model.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked_at", now).Error
}

// IsJWTRevoked is a thin wrapper around the RevocationService.
func (s *JWTService) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	return s.RevocationService.IsJWTRevoked(ctx, jti)
}

// RevokeJWT delegates to the RevocationService to blacklist a JTI until its expiry.
func (s *JWTService) RevokeJWT(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	return s.RevocationService.RevokeJWT(ctx, jti, userID, expiresAt)
}

// RefreshTokenHandler is the HTTP handler that renews access tokens using a refresh token cookie.
// It is implemented on the service so handlers can simply forward requests to it.
func (s *JWTService) RefreshTokenHandler(ctx *gin.Context) {
	clientIP := ctx.ClientIP()

	refreshCookie, err := ctx.Cookie("refresh_token")
	if err != nil || refreshCookie == "" {
		log.Printf("SECURITY: Missing refresh token from IP: %s", clientIP)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Verify combined token (selector:validator)
	rt, err := s.VerifyRefreshToken(refreshCookie)
	if err != nil {
		log.Printf("SECURITY: Invalid refresh token from IP: %s - %v", clientIP, err)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check for reuse / revoked
	if rt.RevokedAt != nil {
		log.Printf("SECURITY ALERT: Revoked token reuse detected! User: %s, IP: %s - Revoking all tokens", rt.UserID, clientIP)
		now := time.Now()
		_ = s.DB.Model(&model.RefreshToken{}).
			Where("user_id = ?", rt.UserID).
			Update("revoked_at", now).Error
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check expiry
	if rt.ExpiresAt.Before(time.Now()) {
		log.Printf("SECURITY: Expired refresh token from IP: %s, User: %s", clientIP, rt.UserID)
		now := time.Now()
		_ = s.DB.Model(&rt).Update("revoked_at", now).Error
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Valid - generate new tokens
	jwtToken, newRefreshToken, err := s.GenerateTokens(rt.UserID)
	if err != nil {
		log.Printf("ERROR: Failed to generate new tokens for user: %s, IP: %s - %v", rt.UserID, clientIP, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new tokens"})
		return
	}

	role := helper.GetRole(rt.UserID, s.DB)
	username := helper.GetUsername(rt.UserID, role, s.DB)

	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour*24*30/time.Second), "/", "", helper.GetCookieSecure(), true)

	log.Printf("INFO: Token refreshed successfully for user: %s, IP: %s", rt.UserID, clientIP)

	ctx.JSON(http.StatusOK, gin.H{
		"token":    jwtToken,
		"username": username,
		"role":     role,
		"userId":   rt.UserID,
	})
}

// LogoutHandler invalidates a user's session: blacklist JWT and revoke refresh token cookie.
// Implemented here so HTTP handlers can forward to it.
func (s *JWTService) LogoutHandler(ctx *gin.Context) {
	clientIP := ctx.ClientIP()
	userID, _ := ctx.Get("userID")

	// Blacklist JWT if present in Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]

			// Parse token to extract JTI & expiry (if valid)
			token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				return s.JWTSecret, nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*model.UserClaims); ok {
					_ = s.RevokeJWT(context.Background(), claims.ID, claims.UserID, claims.ExpiresAt.Time)
					log.Printf("INFO: JWT blacklisted in Redis for user: %s, JTI: %s, IP: %s", claims.UserID, claims.ID, clientIP)
				}
			}
		}
	}

	// Revoke refresh token referenced in cookie (if present)
	refreshCookie, err := ctx.Cookie("refresh_token")
	if err == nil && refreshCookie != "" {
		parts := strings.SplitN(refreshCookie, ":", 2)
		if len(parts) == 2 {
			selector := parts[0]
			validator := parts[1]

			var tokenDB model.RefreshToken
			if err := s.DB.Where("token_selector = ? AND revoked_at IS NULL", selector).First(&tokenDB).Error; err == nil {
				// verify validator against stored hash
				match, verr := verifyToken(validator, tokenDB.Token)
				if verr == nil && match {
					now := time.Now()
					_ = s.DB.Model(&tokenDB).Update("revoked_at", now).Error
					log.Printf("INFO: Refresh token revoked for user: %s, IP: %s", tokenDB.UserID, clientIP)
				}
			}
		}
	}

	// Always clear cookie
	ctx.SetSameSite(helper.GetCookieSameSite())
	ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)

	if userID != nil {
		log.Printf("INFO: User logged out successfully: %s, IP: %s", userID, clientIP)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
