package handlers

import (
	"fmt"
	"html/template"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	gormrepo "ku-work/backend/repository/gorm"
	redisrepo "ku-work/backend/repository/redis"
	"ku-work/backend/services"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, redisClient *redis.Client, emailService *services.EmailService, aiService *services.AIService, jobService *services.JobService, fileService *services.FileService) error {
	// Wire services (DI) and initialize handlers

	// JWT service and handlers
	userRepo := gormrepo.NewGormUserRepository(db)
	refreshRepo := gormrepo.NewGormRefreshTokenRepository(db)
	revocationRepo := redisrepo.NewRedisRevocationRepository(redisClient)
	jwtService := services.NewJWTService(refreshRepo, revocationRepo, userRepo)
	jwtHandlers := NewJWTHandlers(jwtService)

	// File handlers
	fileHandlers := NewFileHandlers(db)

	// Local auth service and handlers
	authService := services.NewAuthService(db, jwtHandlers, userRepo)
	localAuthHandlers := NewLocalAuthHandlers(authService)

	// Google OAuth handlers
	googleOauthConfig := &oauth2.Config{
		RedirectURL:  "postmessage",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables are not set")
	}
	oauthSvc := newOauthService(db, jwtHandlers, googleOauthConfig)
	googleAuthHandlers := NewOAuthHandlers(oauthSvc)

	// Job service (with optional email notifications)
	if jobService == nil {
		jobRepo := gormrepo.NewGormJobRepository(db)
		if emailService != nil {
			tpl, err := template.New("job_approval_status_update.tmpl").ParseFiles("email_templates/job_approval_status_update.tmpl")
			if err != nil {
				return err
			}
			jobService = services.NewJobServiceWithEmail(jobRepo, emailService, tpl)
		} else {
			jobService = services.NewJobService(jobRepo)
		}
	}

	// AI service wiring
	if aiService != nil {
		aiService.JobService = jobService
	} else if emailService != nil {
		var err error
		aiService, err = services.NewAIService(db, emailService, jobService)
		if err != nil {
			return err
		}
	}

	jobHandlers, err := NewJobHandlers(fileHandlers, aiService, jobService)
	if err != nil {
		return err
	}

	applicationHandlers, err := NewApplicationHandlers(db, fileHandlers, emailService)
	if err != nil {
		return err
	}
	studentHandlers, err := NewStudentHandler(db, fileHandlers, aiService, emailService)
	if err != nil {
		return err
	}
	companyRepo := gormrepo.NewGormCompanyRepository(db)
	companyService := services.NewCompanyServiceWithRepo(companyRepo)
	companyHandlers := NewCompanyHandlers(companyService)
	userHandlers := NewUserHandlers(db, helper.GetGracePeriodDays())
	// Admin handlers with injected service
	auditRepo := gormrepo.NewGormAuditRepository(db)
	adminSvc := services.NewAdminService(auditRepo)
	adminHandlers := NewAdminHandlersWithService(adminSvc)

	if fileService == nil {
		return fmt.Errorf("fileService must be provided")
	}
	// Register the FileService with package-level handlers so they can use the configured service
	// (e.g., ServeFileHandler and file upload operations).
	SetFileService(fileService)

	// File Routes
	router.GET("/files/:fileID", fileHandlers.ServeFileHandler)

	// Authentication Routes
	auth := router.Group("/auth")
	auth.POST("/admin/login", middlewares.RateLimiterWithLimits(redisClient, 5, 20), localAuthHandlers.AdminLoginHandler)
	auth.POST("/company/register", localAuthHandlers.CompanyRegisterHandler)
	auth.POST("/company/login", middlewares.RateLimiterWithLimits(redisClient, 5, 20), localAuthHandlers.CompanyLoginHandler)
	auth.POST("/google/login", middlewares.RateLimiterWithLimits(redisClient, 5, 20), googleAuthHandlers.GoogleOauthHandler)

	// Protected Authentication Routes
	authProtected := auth.Group("", middlewares.AuthMiddlewareWithRedis(jwtHandlers.Service.JWTSecret, redisClient))
	// Student registration requires an authenticated and active account.
	authProtected.POST("/refresh", middlewares.RateLimiterWithLimits(redisClient, 5, 20), jwtHandlers.RefreshTokenHandler)
	authProtected.POST("/logout", jwtHandlers.LogoutHandler)
	// Only active account can register
	authProtectedActive := authProtected.Group("", middlewares.AccountActiveMiddleware(db))
	authProtectedActive.POST("/student/register", studentHandlers.RegisterHandler)

	// User Routes
	protectedRouter := router.Group("", middlewares.AuthMiddlewareWithRedis(jwtHandlers.Service.JWTSecret, redisClient))
	// Keep reactivation available even for deactivated accounts
	protectedRouter.POST("/me/reactivate", userHandlers.ReactivateAccount)

	// Routes that require the account to be active
	protectedActive := protectedRouter.Group("", middlewares.AccountActiveMiddleware(db))
	protectedActive.PATCH("/me", userHandlers.EditProfileHandler)
	protectedActive.GET("/me", userHandlers.GetProfileHandler)
	protectedActive.POST("/me/deactivate", userHandlers.DeactivateAccount)

	// Company Routs
	company := protectedActive.Group("/company")
	company.GET("/:id", companyHandlers.GetCompanyProfileHandler)

	companyAdmin := company.Group("", middlewares.AdminPermissionMiddleware(db))
	companyAdmin.GET("", companyHandlers.GetCompanyListHandler)

	// Job Routes
	job := protectedActive.Group("/jobs")
	job.GET("", jobHandlers.FetchJobsHandler)
	job.POST("", jobHandlers.CreateJobHandler)
	job.GET("/:id/applications", applicationHandlers.GetJobApplicationsHandler)
	job.DELETE("/:id/applications", applicationHandlers.ClearJobApplicationsHandler)
	job.GET("/:id/applications/:email", applicationHandlers.GetJobApplicationHandler)
	job.PATCH("/:id/applications/:studentUserId/status", applicationHandlers.UpdateJobApplicationStatusHandler)
	job.GET("/:id", jobHandlers.GetJobDetailHandler)
	job.POST("/:id/apply", applicationHandlers.CreateJobApplicationHandler)
	job.PATCH("/:id", jobHandlers.EditJobHandler)

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
