package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	docs "ku-work/backend/docs"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	gormrepo "ku-work/backend/repository/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

	// Google OAuth - only register if env is configured
	if googleOAuthConfigured() {
		auth.POST("/google/login", middlewares.RateLimiterWithLimits(d.Services.RateLimiter, 5, 20), googleOAuthLoginHandler(d.Services))
	} else {
		// Intentionally silent to avoid leaking configuration details
	}

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

func googleOAuthConfigured() bool {
	return strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")) != "" &&
		strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")) != ""
}

func googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "postmessage",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

// googleOAuthLoginHandler performs Google OAuth code exchange, fetches userinfo,
// delegates to AuthService.HandleGoogleOAuth, and sets refresh cookie + response.
func googleOAuthLoginHandler(svcs *ServicesBundle) gin.HandlerFunc {
	type oauthToken struct {
		Code string `json:"code"`
	}
	type userInfo struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}

	return func(ctx *gin.Context) {
		var req oauthToken
		if err := ctx.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Code) == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
			return
		}

		cfg := googleOAuthConfig()
		tok, err := cfg.Exchange(context.Background(), req.Code)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
			return
		}

		reqHttp, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to init userinfo request"})
			return
		}
		reqHttp.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok.AccessToken))

		client := &http.Client{Timeout: 10 * time.Second}
		res, err := client.Do(reqHttp)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch userinfo"})
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "access token invalid or expired"})
			return
		}

		var ui userInfo
		if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode userinfo"})
			return
		}

		jwtToken, refreshToken, username, role, userId, isRegistered, statusCode, err := svcs.Auth.HandleGoogleOAuth(struct {
			ID         string
			Email      string
			Name       string
			GivenName  string
			FamilyName string
		}{
			ID:         ui.ID,
			Email:      ui.Email,
			Name:       ui.Name,
			GivenName:  ui.GivenName,
			FamilyName: ui.FamilyName,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		maxAge := int(time.Hour * 24 * 30 / time.Second)
		ctx.SetSameSite(helper.GetCookieSameSite())
		ctx.SetCookie("refresh_token", refreshToken, maxAge, "/", "", helper.GetCookieSecure(), true)

		ctx.JSON(statusCode, gin.H{
			"token":        jwtToken,
			"username":     username,
			"role":         role,
			"userId":       userId,
			"isRegistered": isRegistered,
		})
	}
}

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
