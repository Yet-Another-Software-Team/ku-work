package handlers

import (
	"ku-work/backend/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	authHandlers := NewAuthHandlers(db)
	
	
	router.POST("/register", authHandlers.RegisterHandler)
	router.POST("/login", authHandlers.LoginHandler)	
	router.POST("/refresh", authHandlers.RefreshTokenHandler)
	router.POST("/logout", authHandlers.LogoutHandler)

	
	protected := router.Use(middlewares.AuthMiddleware(authHandlers.JWTSecret))
	protected.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Protected route"})
	})
}
