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
	repo "ku-work/backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// JWTService encapsulates JWT and refresh-token related operations.
//
// Note: this service no longer performs raw GORM or Redis operations itself.
// Instead it depends on repository interfaces for refresh-token persistence and
// JWT revocation storage. The `DB` field remains for passing into UserRepo
// calls which expect a *gorm.DB (transaction support).
type JWTService struct {
	DB               *gorm.DB
	RefreshTokenRepo repo.RefreshTokenRepository
	RevocationRepo   repo.JWTRevocationRepository
	JWTSecret        []byte
	UserRepo         repo.UserRepository
}

// NewJWTService constructs a new JWTService wired with Redis/GORM dependencies.
// It reads JWT_SECRET from the environment (and requires it to be at least 32 bytes).
func NewJWTService(db *gorm.DB, refreshRepo repo.RefreshTokenRepository, revocationRepo repo.JWTRevocationRepository, userRepo repo.UserRepository) *JWTService {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes long for security")
	}

	if revocationRepo == nil {
		log.Fatal("revocation repository is required for JWT revocation")
	}
	if refreshRepo == nil {
		log.Fatal("refresh token repository is required")
	}

	return &JWTService{
		DB:               db,
		RefreshTokenRepo: refreshRepo,
		RevocationRepo:   revocationRepo,
		JWTSecret:        jwtSecret,
		UserRepo:         userRepo,
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

	// Count active tokens for this user via repository
	activeTokenCount, err := s.RefreshTokenRepo.CountActiveByUser(context.Background(), userID, time.Now())
	if err != nil {
		return "", "", err
	}

	// If at or over limit, revoke oldest tokens to make room using repository methods
	if activeTokenCount >= int64(maxSessions) {
		tokensToRevoke := int(activeTokenCount) - (maxSessions - 1)
		if tokensToRevoke > 0 {
			oldest, err := s.RefreshTokenRepo.FindOldestActiveByUserLimit(context.Background(), userID, tokensToRevoke)
			if err == nil && len(oldest) > 0 {
				now := time.Now()
				ids := make([]uint, 0, len(oldest))
				for _, t := range oldest {
					ids = append(ids, t.ID)
				}
				_ = s.RefreshTokenRepo.RevokeByIDs(context.Background(), ids, now)
			}
		}
	}

	// Create refresh token record via repository
	rt := model.RefreshToken{
		UserID:        userID,
		TokenSelector: selector,
		Token:         hashedValidator,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		RevokedAt:     nil,
	}

	if err := s.RefreshTokenRepo.Create(context.Background(), &rt); err != nil {
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

	stored, err := s.RefreshTokenRepo.FindBySelector(context.Background(), selector)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, fmt.Errorf("refresh token not found")
	}

	match, err := verifyToken(validator, stored.Token)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, fmt.Errorf("refresh token validation failed")
	}
	return stored, nil
}

// RevokeRefreshTokenBySelector revokes a refresh token record identified by selector.
func (s *JWTService) RevokeRefreshTokenBySelector(selector string) error {
	now := time.Now()
	return s.RefreshTokenRepo.RevokeBySelector(context.Background(), selector, now)
}

// RevokeAllRefreshTokensForUser revokes all active refresh tokens for a user.
func (s *JWTService) RevokeAllRefreshTokensForUser(userID string) error {
	now := time.Now()
	return s.RefreshTokenRepo.RevokeAllForUser(context.Background(), userID, now)
}

// IsJWTRevoked is a thin wrapper around the revocation repository.
func (s *JWTService) IsJWTRevoked(ctx context.Context, jti string) (bool, error) {
	return s.RevocationRepo.IsJWTRevoked(ctx, jti)
}

// RevokeJWT delegates to the revocation repository to blacklist a JTI until its expiry.
func (s *JWTService) RevokeJWT(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	return s.RevocationRepo.RevokeJWT(ctx, jti, userID, expiresAt)
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
		_ = s.RefreshTokenRepo.RevokeAllForUser(context.Background(), rt.UserID, now)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", "", -1, "/", "", helper.GetCookieSecure(), true)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Check expiry
	if rt.ExpiresAt.Before(time.Now()) {
		log.Printf("SECURITY: Expired refresh token from IP: %s, User: %s", clientIP, rt.UserID)
		now := time.Now()
		_ = s.RefreshTokenRepo.UpdateRevokedAt(context.Background(), rt.ID, now)
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

	// Resolve role using UserRepository (prefer explicit counts to avoid leaking DB access into helper).
	var role helper.Role = helper.Unknown
	if cnt, err := s.UserRepo.CountAdminByUserID(s.DB, rt.UserID); err == nil && cnt > 0 {
		role = helper.Admin
	} else if cnt, err := s.UserRepo.CountCompanyByUserID(s.DB, rt.UserID); err == nil && cnt > 0 {
		role = helper.Company
	} else {
		// Fallback: unknown (other role checks like Student/Viewer may be implemented in repo later)
		role = helper.Unknown
	}

	// Resolve username via UserRepository when role is admin/company; otherwise fallback to unknown.
	username := "unknown"
	if role == helper.Company || role == helper.Admin {
		if u, err := s.UserRepo.FindUserByID(s.DB, rt.UserID); err == nil && u != nil {
			username = u.Username
		}
	}

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

			tokenDB, err := s.RefreshTokenRepo.FindBySelector(context.Background(), selector)
			if err == nil && tokenDB != nil {
				// verify validator against stored hash
				match, verr := verifyToken(validator, tokenDB.Token)
				if verr == nil && match {
					now := time.Now()
					_ = s.RefreshTokenRepo.UpdateRevokedAt(context.Background(), tokenDB.ID, now)
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
