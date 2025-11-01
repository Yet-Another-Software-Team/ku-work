package handlers

import (
	"ku-work/backend/model"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type JWTHandlers struct {
	Service *services.JWTService
}

func NewJWTHandlers(db *gorm.DB, redisClient *redis.Client) *JWTHandlers {
	// Construct the service-layer repositories/adapters and expose the service through a thin handler.
	userRepo := gormrepo.NewGormUserRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)

	// Wire the service with repository abstractions (no direct DB/Redis use inside the service).
	svc := services.NewJWTService(db, refreshRepo, revocationRepo, userRepo)
	return &JWTHandlers{
		Service: svc,
	}
}

// Token hashing/verification implemented in service layer (services.JWTService).
// Handlers delegate token operations to the service; helper implementations were moved.

// GenerateTokens delegates to the service implementation.
func (h *JWTHandlers) GenerateTokens(userID string) (string, string, error) {
	return h.Service.GenerateTokens(userID)
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
	// Delegate full refresh logic to the service implementation
	h.Service.RefreshTokenHandler(ctx)
}

// @Summary Logout user
// @Description Invalidates the user's session by revoking both the JWT token (blacklist) and refresh token. Complies with OWASP session termination requirements.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{message=string} "Logged out successfully"
// @Router /auth/logout [post]
func (h *JWTHandlers) LogoutHandler(ctx *gin.Context) {
	// Delegate logout / revocation to the service
	h.Service.LogoutHandler(ctx)
}

// HandleToken delegates to the service implementation.
func (h *JWTHandlers) HandleToken(user model.User) (string, string, error) {
	return h.Service.HandleToken(user)
}
