package bootstrap

import (
	"os"

	docs "ku-work/backend/docs"
	"ku-work/backend/middlewares"
	gormrepo "ku-work/backend/repository/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// RouterDeps contains everything needed to construct the HTTP router and register routes.
type RouterDeps struct {
	DB       *gorm.DB
	Redis    *redis.Client
	Services *ServicesBundle
	Handlers *HandlersBundle
}

// NewRouter creates the Gin engine, registers middlewares/routes, and configures Swagger.
func NewRouter(d RouterDeps) *gin.Engine {
	if d.DB == nil || d.Services == nil || d.Handlers == nil {
		panic("router deps must include DB, Services, and Handlers")
	}

	router := gin.Default()
	router.Use(cors.New(middlewares.SetupCORS()))

	registerRoutes(router, d)
	setupSwagger(router)

	return router
}

func registerRoutes(router *gin.Engine, d RouterDeps) {
	// Public file route
	router.GET("/files/:fileID", d.Handlers.File.ServeFileHandler)

	// Authentication
	auth := router.Group("/auth")
	auth.POST("/admin/login", middlewares.RateLimiterWithLimits(d.Services.RateLimiter, 5, 20), d.Handlers.LocalAuth.AdminLoginHandler)
	auth.POST("/company/register", d.Handlers.LocalAuth.CompanyRegisterHandler)
	auth.POST("/company/login", middlewares.RateLimiterWithLimits(d.Services.RateLimiter, 5, 20), d.Handlers.LocalAuth.CompanyLoginHandler)
	auth.POST("/google/login", middlewares.RateLimiterWithLimits(d.Services.RateLimiter, 5, 20), d.Handlers.OAuth.GoogleOauthHandler)

	// Protected Authentication Routes
	authProtected := auth.Group("", middlewares.AuthMiddleware(d.Services.JWT.JWTSecret, d.Services.JWT))
	authProtected.POST("/refresh", middlewares.RateLimiterWithLimits(d.Services.RateLimiter, 5, 20), d.Handlers.JWT.RefreshTokenHandler)
	authProtected.POST("/logout", d.Handlers.JWT.LogoutHandler)

	// Student registration requires active account
	authProtectedActive := authProtected.Group("", middlewares.AccountActiveMiddleware(d.DB))
	authProtectedActive.POST("/student/register", d.Handlers.Student.RegisterHandler)

	// User routes
	protectedRouter := router.Group("", middlewares.AuthMiddleware(d.Services.JWT.JWTSecret, d.Services.JWT))
	protectedRouter.POST("/me/reactivate", d.Handlers.User.ReactivateAccount)

	protectedActive := protectedRouter.Group("", middlewares.AccountActiveMiddleware(d.DB))
	protectedActive.PATCH("/me", d.Handlers.User.EditProfileHandler)
	protectedActive.GET("/me", d.Handlers.User.GetProfileHandler)
	protectedActive.POST("/me/deactivate", d.Handlers.User.DeactivateAccount)

	// Company routes
	company := protectedActive.Group("/company")
	company.GET("/:id", d.Handlers.Company.GetCompanyProfileHandler)
	companyAdmin := company.Group("", middlewares.AdminPermissionMiddleware(gormrepo.NewGormIdentityRepository(d.DB)))
	companyAdmin.GET("", d.Handlers.Company.GetCompanyListHandler)

	// Job routes
	job := protectedActive.Group("/jobs")
	job.GET("", d.Handlers.Job.FetchJobsHandler)
	job.POST("", d.Handlers.Job.CreateJobHandler)
	job.GET("/:id", d.Handlers.Job.GetJobDetailHandler)
	job.PATCH("/:id", d.Handlers.Job.EditJobHandler)
	job.POST("/:id/apply", d.Handlers.Application.CreateJobApplicationHandler)
	job.GET("/:id/applications", d.Handlers.Application.GetJobApplicationsHandler)
	job.DELETE("/:id/applications", d.Handlers.Application.ClearJobApplicationsHandler)
	job.GET("/:id/applications/:email", d.Handlers.Application.GetJobApplicationHandler)
	job.PATCH("/:id/applications/:studentUserId/status", d.Handlers.Application.UpdateJobApplicationStatusHandler)

	jobAdmin := job.Group("", middlewares.AdminPermissionMiddleware(gormrepo.NewGormIdentityRepository(d.DB)))
	jobAdmin.POST("/:id/approval", d.Handlers.Job.JobApprovalHandler)

	// Application routes
	application := protectedActive.Group("/applications")
	application.GET("", d.Handlers.Application.GetAllJobApplicationsHandler)

	// Student routes
	student := protectedActive.Group("/students")
	student.GET("", d.Handlers.Student.GetProfileHandler)

	studentAdmin := student.Group("", middlewares.AdminPermissionMiddleware(gormrepo.NewGormIdentityRepository(d.DB)))
	studentAdmin.POST("/:id/approval", d.Handlers.Student.ApproveHandler)

	// Admin routes
	admin := protectedActive.Group("/admin", middlewares.AdminPermissionMiddleware(gormrepo.NewGormIdentityRepository(d.DB)))
	admin.GET("/audits", d.Handlers.Admin.FetchAuditLog)
	admin.GET("/emaillog", d.Handlers.Admin.FetchEmailLog)
}

// googleOAuthLoginHandler performs Google OAuth code exchange, fetches userinfo,
// delegates to AuthService.HandleGoogleOAuth, and sets refresh cookie + response.

func setupSwagger(router *gin.Engine) {
	swaggerHost, hasHost := os.LookupEnv("SWAGGER_HOST")
	if hasHost {
		docs.SwaggerInfo.Host = swaggerHost
	} else {
		docs.SwaggerInfo.Host = "localhost:8000"
	}
	docs.SwaggerInfo.BasePath = ""
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
