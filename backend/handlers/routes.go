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
	
	// Middlewares
	turnstileMiddleware := middlewares.TurnstileMiddleware()
	authedMiddleware := middlewares.AuthMiddleware(jwtHandlers.JWTSecret, redisClient)
	activeMiddleware := middlewares.AccountActiveMiddleware(db)
	adminMiddleware := middlewares.AdminPermissionMiddleware(db)

	// Rate Limiter
	trustedRateLimiter := middlewares.RateLimiterWithLimits(redisClient, 100, 100*60)
	loginRateLimiter := middlewares.RateLimiterWithLimits(redisClient, 5, 20)
	authedRateLimiter := middlewares.RateLimiterWithLimits(redisClient, 60, 60*60)

	if fileService == nil {
		return fmt.Errorf("fileService must be provided")
	}
	// Register the FileService with package-level handlers so handler functions
	// such as SaveFile and ServeFileHandler can use the configured service.
	SetFileService(fileService)
	// File Routes
	router.GET("/files/:fileID", fileHandlers.ServeFileHandler)

	// Authentication Routes
	auth := router.Group("/auth", loginRateLimiter)
	auth.POST("/admin/login", turnstileMiddleware, localAuthHandlers.AdminLoginHandler)
	auth.POST("/company/register", turnstileMiddleware, localAuthHandlers.CompanyRegisterHandler)
	auth.POST("/company/login", turnstileMiddleware, localAuthHandlers.CompanyLoginHandler)
	auth.POST("/google/login", googleAuthHandlers.GoogleOauthHandler)

	
	// Logout and Student Register treated as normal API
	authLooseProtected := router.Group("/auth", authedRateLimiter, authedMiddleware)
	authLooseProtected.POST("/refresh", jwtHandlers.RefreshTokenHandler)
	authLooseProtected.POST("/logout", jwtHandlers.LogoutHandler)
	// Only active account can register
	authProtectedActive := authLooseProtected.Group("", activeMiddleware)
	authProtectedActive.POST("/student/register", turnstileMiddleware, studentHandlers.RegisterHandler)

	// User Routes
	protectedRouter := router.Group("", authedMiddleware, authedRateLimiter)
	// Keep reactivation available even for deactivated accounts
	protectedRouter.POST("/me/reactivate", turnstileMiddleware, userHandlers.ReactivateAccount)

	// Routes that require the account to be active
	protectedActive := protectedRouter.Group("", activeMiddleware)
	trustedProtectedActive := router.Group("", authedMiddleware, adminMiddleware, activeMiddleware, trustedRateLimiter)
	protectedActive.PATCH("/me", turnstileMiddleware, userHandlers.EditProfileHandler)
	protectedActive.GET("/me", userHandlers.GetProfileHandler)
	protectedActive.POST("/me/deactivate", turnstileMiddleware, userHandlers.DeactivateAccount)

	// Company Routs
	company := protectedActive.Group("/company")
	company.GET("/:id", companyHandlers.GetCompanyProfileHandler)

	companyAdmin := trustedProtectedActive.Group("/company")
	companyAdmin.GET("", companyHandlers.GetCompanyListHandler)

	// Job Routes
	job := protectedActive.Group("/jobs")
	job.GET("", jobHandlers.FetchJobsHandler)
	job.POST("", turnstileMiddleware, jobHandlers.CreateJobHandler)
	job.GET("/:id/applications", applicationHandlers.GetJobApplicationsHandler)
	job.DELETE("/:id/applications", applicationHandlers.ClearJobApplicationsHandler)
	job.GET("/:id/applications/:email", applicationHandlers.GetJobApplicationHandler)
	job.PATCH("/:id/applications/:studentUserId/status", applicationHandlers.UpdateJobApplicationStatusHandler)
	job.GET("/:id", jobHandlers.GetJobDetailHandler)
	job.POST("/:id/apply", turnstileMiddleware, applicationHandlers.CreateJobApplicationHandler)
	job.PATCH("/:id", middlewares.TurnstileExceptionMiddleware(), jobHandlers.EditJobHandler, turnstileMiddleware, jobHandlers.EditJobHandler)

	jobAdmin := trustedProtectedActive.Group("/jobs")
	jobAdmin.POST("/:id/approval", jobHandlers.JobApprovalHandler)

	// Application Routes
	application := protectedActive.Group("/applications")
	application.GET("", applicationHandlers.GetAllJobApplicationsHandler)

	// Student Routes
	student := protectedActive.Group("/students")
	student.GET("", studentHandlers.GetProfileHandler)

	studentAdmin := trustedProtectedActive.Group("/students")
	studentAdmin.POST("/:id/approval", studentHandlers.ApproveHandler)

	// Admin Routes
	admin := trustedProtectedActive.Group("/admin")
	admin.GET("/audits", adminHandlers.FetchAuditLog)
	admin.GET("/emaillog", adminHandlers.FetchEmailLog)
	return nil
}
