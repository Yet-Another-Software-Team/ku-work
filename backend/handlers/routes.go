package handlers

import (
	"fmt"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, redisClient *redis.Client, emailService *services.EmailService, aiService *services.AIService, fileService *services.FileService) error {
	// Initialize handlers
	jwtHandlers := NewJWTHandlers(db, redisClient)
	fileHandlers := NewFileHandlers(db)
	localAuthHandlers := NewLocalAuthHandlers(db, jwtHandlers)
	googleAuthHandlers := NewOAuthHandlers(db, jwtHandlers)

	jobHandlers, err := NewJobHandlers(db, aiService, emailService)
	if err != nil {
		return err
	}
	applicationHandlers, err := NewApplicationHandlers(db, emailService)
	if err != nil {
		return err
	}
	studentHandlers, err := NewStudentHandler(db, fileHandlers, aiService, emailService)
	if err != nil {
		return err
	}
	companyHandlers := NewCompanyHandlers(db)
	userHandlers := NewUserHandlers(db, helper.GetGracePeriodDays())
	adminHandlers := NewAdminHandlers(db)

	if fileService == nil {
		return fmt.Errorf("fileService must be provided")
	}
	// Register the FileService with package-level handlers so handler functions
	// such as SaveFile and ServeFileHandler can use the configured service.
	SetFileService(fileService)

	// File Routes
	router.GET("/files/:fileID", fileHandlers.ServeFileHandler)

	// Authentication Routes
	auth := router.Group("/auth")
	auth.POST("/admin/login", middlewares.TurnstileMiddleware(), middlewares.RateLimiterWithLimits(redisClient, 5, 20), localAuthHandlers.AdminLoginHandler)
	auth.POST("/company/register", middlewares.TurnstileMiddleware(), localAuthHandlers.CompanyRegisterHandler)
	auth.POST("/company/login", middlewares.TurnstileMiddleware(), middlewares.RateLimiterWithLimits(redisClient, 5, 20), localAuthHandlers.CompanyLoginHandler)
	auth.POST("/google/login", middlewares.RateLimiterWithLimits(redisClient, 5, 20), googleAuthHandlers.GoogleOauthHandler)

	// Protected Authentication Routes
	authProtected := auth.Group("", middlewares.AuthMiddleware(jwtHandlers.JWTSecret, redisClient))
	authProtected.POST("/refresh", middlewares.RateLimiterWithLimits(redisClient, 5, 20), jwtHandlers.RefreshTokenHandler)
	authProtected.POST("/logout", jwtHandlers.LogoutHandler)
	// Only active account can register
	authProtectedActive := authProtected.Group("", middlewares.AccountActiveMiddleware(db))
	authProtectedActive.POST("/student/register", middlewares.TurnstileMiddleware(), studentHandlers.RegisterHandler)

	// User Routes
	protectedRouter := router.Group("", middlewares.AuthMiddleware(jwtHandlers.JWTSecret, redisClient))
	// Keep reactivation available even for deactivated accounts
	protectedRouter.POST("/me/reactivate", middlewares.TurnstileMiddleware(), userHandlers.ReactivateAccount)

	// Routes that require the account to be active
	protectedActive := protectedRouter.Group("", middlewares.AccountActiveMiddleware(db))
	protectedActive.PATCH("/me", middlewares.TurnstileMiddleware(), userHandlers.EditProfileHandler)
	protectedActive.GET("/me", userHandlers.GetProfileHandler)
	protectedActive.POST("/me/deactivate", middlewares.TurnstileMiddleware(), userHandlers.DeactivateAccount)

	// Company Routs
	company := protectedActive.Group("/company")
	company.GET("/:id", companyHandlers.GetCompanyProfileHandler)

	companyAdmin := company.Group("", middlewares.AdminPermissionMiddleware(db))
	companyAdmin.GET("", companyHandlers.GetCompanyListHandler)

	// Job Routes
	job := protectedActive.Group("/jobs")
	job.GET("", jobHandlers.FetchJobsHandler)
	job.POST("", middlewares.TurnstileMiddleware(), jobHandlers.CreateJobHandler)
	job.GET("/:id/applications", applicationHandlers.GetJobApplicationsHandler)
	job.DELETE("/:id/applications", applicationHandlers.ClearJobApplicationsHandler)
	job.GET("/:id/applications/:email", applicationHandlers.GetJobApplicationHandler)
	job.PATCH("/:id/applications/:studentUserId/status", applicationHandlers.UpdateJobApplicationStatusHandler)
	job.GET("/:id", jobHandlers.GetJobDetailHandler)
	job.POST("/:id/apply", middlewares.TurnstileMiddleware(), applicationHandlers.CreateJobApplicationHandler)
	job.PATCH("/:id", middlewares.TurnstileExceptionMiddleware(), jobHandlers.EditJobHandler, middlewares.TurnstileMiddleware(), jobHandlers.EditJobHandler)

	jobAdmin := job.Group("", middlewares.AdminPermissionMiddleware(db))
	jobAdmin.POST("/:id/approval", jobHandlers.JobApprovalHandler)

	// Application Routes
	application := protectedActive.Group("/applications")
	application.GET("", applicationHandlers.GetAllJobApplicationsHandler)

	// Student Routes
	student := protectedActive.Group("/students")
	student.GET("", studentHandlers.GetProfileHandler)

	studentAdmin := student.Group("", middlewares.AdminPermissionMiddleware(db))
	studentAdmin.POST("/:id/approval", studentHandlers.ApproveHandler)

	// Admin Routes
	admin := protectedActive.Group("/admin", middlewares.AdminPermissionMiddleware(db))
	admin.GET("/audits", adminHandlers.FetchAuditLog)
	admin.GET("/emaillog", adminHandlers.FetchEmailLog)
	return nil
}
