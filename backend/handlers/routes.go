package handlers

import (
	"ku-work/backend/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	localAuthHandlers := NewLocalAuthHandlers(db)
	googleAuthHandlers := NewOAuthHandlers(db)
	jwtHandler := NewJWTHandlers(db)

	router.POST("/register", localAuthHandlers.RegisterHandler)
	router.POST("/google/login", googleAuthHandlers.GoogleOauthHandler)
	router.POST("/login", localAuthHandlers.LoginHandler)
	router.POST("/refresh", jwtHandler.RefreshTokenHandler)
	router.POST("/logout", jwtHandler.LogoutHandler)

	// Authentication Protected Routes
	protected := router.Use(middlewares.AuthMiddleware(jwtHandler.JWTSecret))
	protected.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Protected route"})
	})
}
