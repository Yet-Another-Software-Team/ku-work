package handlers

import (
	"ku-work/backend/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize handlers
	jwtHandlers := NewJWTHandlers(db)
	fileHandlers := NewFileHandlers(db)
	localAuthHandlers := NewLocalAuthHandlers(db, jwtHandlers)
	googleAuthHandlers := NewOAuthHandlers(db, jwtHandlers)
	jobHandlers := NewJobHandlers(db)
	applicationHandlers := NewApplicationHandlers(db)
	studentHandlers := NewStudentHandler(db, fileHandlers)
	companyHandlers := NewCompanyHandlers(db)
	userHandlers := NewUserHandlers(db)

	// Authentication Routes
	auth := router.Group("/auth")
	auth.POST("/admin/login", localAuthHandlers.AdminLoginHandler)
	auth.POST("/company/register", localAuthHandlers.CompanyRegisterHandler)
	auth.POST("/company/login", localAuthHandlers.CompanyLoginHandler)
	auth.POST("/google/login", googleAuthHandlers.GoogleOauthHandler)

	// Protected Authentication Routes
	authProtected := auth.Group("", middlewares.AuthMiddleware(jwtHandlers.JWTSecret))
	authProtected.POST("/student/register", studentHandlers.RegisterHandler)
	authProtected.POST("/refresh", jwtHandlers.RefreshTokenHandler)
	authProtected.POST("/logout", jwtHandlers.LogoutHandler)

	// File Routes (Currently public)
	router.GET("/files/:fileID", fileHandlers.ServeFileHandler)

	// User Routes
	protectedRouter := router.Group("", middlewares.AuthMiddleware(jwtHandlers.JWTSecret))
	protectedRouter.PATCH("/me", userHandlers.EditProfileHandler)
	protectedRouter.GET("/me", userHandlers.GetProfileHandler)

	// Company Routs
	company := protectedRouter.Group("/company")
	company.GET("/:id", companyHandlers.GetCompanyProfileHandler)

	companyAdmin := company.Group("", middlewares.AdminPermissionMiddleware(db))
	companyAdmin.GET("", companyHandlers.GetCompanyListHandler)

	// Job Routes
	job := protectedRouter.Group("/jobs")
	job.GET("", jobHandlers.FetchJobsHandler)
	job.POST("", jobHandlers.CreateJobHandler)
	job.GET("/:id/applications", applicationHandlers.GetJobApplicationsHandler)
	job.GET("/:id/applications/:studentId", applicationHandlers.GetJobApplicationHandler)
	job.PATCH("/:id/applications/:studentId", applicationHandlers.UpdateJobApplicationStatusHandler)
	job.GET("/:id", jobHandlers.GetJobDetailHandler)
	job.POST("/:id/apply", applicationHandlers.CreateJobApplicationHandler)
	job.PATCH("/:id", jobHandlers.EditJobHandler)

	jobAdmin := job.Group("", middlewares.AdminPermissionMiddleware(db))
	jobAdmin.POST("/:id/approval", jobHandlers.JobApprovalHandler)

	// Application Routes
	application := protectedRouter.Group("/applications")
	application.GET("", applicationHandlers.GetAllJobApplicationsHandler)

	// Student Routes
	student := protectedRouter.Group("/students")
	student.GET("", studentHandlers.GetProfileHandler)

	studentAdmin := student.Group("", middlewares.AdminPermissionMiddleware(db))
	studentAdmin.POST("/:id/approval", studentHandlers.ApproveHandler)

}
