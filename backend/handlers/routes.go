package handlers

import (
	"ku-work/backend/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	jwtHandler := NewJWTHandlers(db)
	localAuthHandlers := NewLocalAuthHandlers(db, jwtHandler)
	googleAuthHandlers := NewOAuthHandlers(db, jwtHandler)
	jobHandlers := NewJobHandlers(db)
	studentHandler := NewStudentHandler(db)

	router.POST("/register", localAuthHandlers.RegisterHandler)
	router.POST("/google/login", googleAuthHandlers.GoogleOauthHandler)
	router.POST("/login", localAuthHandlers.LoginHandler)
	router.POST("/refresh", jwtHandler.RefreshTokenHandler)
	router.POST("/logout", jwtHandler.LogoutHandler)

	// Authentication Protected Routes
	authed := router.Group("/", middlewares.AuthMiddleware(jwtHandler.JWTSecret))
	authed.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Protected route"})
	})

	// Admin Routes
	admin := authed.Group("/", middlewares.AdminPermissionMiddleware(db))
	admin.GET("/admin", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Admin route"})
	})

	// Student routes
	authed.POST("/students/register", studentHandler.RegisterHandler)
	authed.PATCH("/students", studentHandler.EditProfileHandler)
	authed.GET("/students", studentHandler.GetProfileHandler)
	admin.POST("/students/approve", studentHandler.ApproveHandler)

	// Job routes
	router.GET("/job", jobHandlers.FetchJobs)
	admin.POST("/job", jobHandlers.CreateJob)
	authed.PATCH("/job", jobHandlers.EditJob)
	admin.POST("/job/approve", jobHandlers.ApproveJob)
}
