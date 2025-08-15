package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	userHandlers := NewUserHandlers(db)

	// Health check route
	router.GET("/", userHandlers.HealthCheck)

	// User routes
	router.GET("/users", userHandlers.GetUsers)
	router.POST("/create_user", userHandlers.CreateUser)
}
