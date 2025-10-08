package handlers

import (
	"ku-work/backend/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	jwtHandler := NewJWTHandlers(db)
	fileHandlers := NewFileHandlers(db)
	localAuthHandlers := NewLocalAuthHandlers(db, jwtHandler)
	googleAuthHandlers := NewOAuthHandlers(db, jwtHandler)
	jobHandlers := NewJobHandlers(db)
	studentHandler := NewStudentHandler(db, fileHandlers)
	companyHandler := NewCompanyHandlers(db)

	//Authentication Routes
	router.POST("/admin/login", localAuthHandlers.AdminLoginHandler)
	router.POST("/company/register", localAuthHandlers.CompanyRegisterHandler)
	router.POST("/company/login", localAuthHandlers.CompanyLoginHandler)
	router.POST("/google/login", googleAuthHandlers.GoogleOauthHandler)
	router.POST("/refresh", jwtHandler.RefreshTokenHandler)
	router.POST("/logout", jwtHandler.LogoutHandler)

	router.GET("/files/:fileID", fileHandlers.ServeFile)

	// Authentication Protected Routes
	authed := router.Group("/", middlewares.AuthMiddleware(jwtHandler.JWTSecret))

	// Admin Routes
	admin := authed.Group("/", middlewares.AdminPermissionMiddleware(db))

	// Student routes
	authed.POST("/students/register", studentHandler.RegisterHandler)
	authed.PATCH("/students", studentHandler.EditProfileHandler)
	authed.GET("/students", studentHandler.GetProfileHandler)
	admin.POST("/students/approve", studentHandler.ApproveHandler)

	// Job routes
	authed.GET("/job", jobHandlers.FetchJobs)
	authed.POST("/job", jobHandlers.CreateJob)
	authed.PATCH("/job", jobHandlers.EditJob)
	admin.POST("/job/approve", jobHandlers.ApproveJob)
	authed.POST("/job/apply", jobHandlers.ApplyJob)
	authed.GET("/job/application", jobHandlers.FetchJobApplications)
	authed.POST("/job/application/accept", jobHandlers.AcceptJobApplication)

	// Company routes
	authed.PATCH("/company", companyHandler.EditProfileHandler)
	authed.GET("/company", companyHandler.GetProfileHandler)
}
