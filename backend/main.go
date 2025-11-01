package main

import (
	"context"
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	"ku-work/backend/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	docs "ku-work/backend/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title KU-Work API
// @version 1.0
// @description This is a sample API for KU-Work
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the Token
func main() {
	_ = godotenv.Load()

	// Initialize infrastructure
	if err := initializeFilesDirectory(); err != nil {
		log.Printf("Failed to create files directory: %v", err)
		return
	}

	db, err := database.LoadDB()
	if err != nil {
		log.Printf("Database initialization failed: %v", err)
		return
	}

	redisClient := initializeRedis()
	emailService, aiService := initializeServices(db)

	// Setup background tasks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := setupScheduler(ctx, db, emailService)
	scheduler.Start()

	// Setup HTTP server
	router := setupRouter(db, redisClient, emailService, aiService)
	srv := startServer(router)

	// Graceful shutdown
	waitForShutdownSignal()
	performGracefulShutdown(srv, cancel, scheduler, redisClient)
}

// initializeFilesDirectory creates the files directory if it doesn't exist
func initializeFilesDirectory() error {
	return os.MkdirAll("./files", 0755)
}

// initializeRedis initializes Redis client with fail-open behavior
func initializeRedis() *redis.Client {
	redisClient, redis_err := database.LoadRedis()
	if redis_err != nil {
		log.Fatalf("FATAL: Redis initialization failed: %v.", redis_err)
		return nil
	}
	log.Println("Redis connected successfully")
	return redisClient
}

// initializeServices initializes email and AI services with proper error handling
func initializeServices(db *gorm.DB) (*services.EmailService, *services.AIService) {
	emailService, err := services.NewEmailService(db)
	if err != nil {
		log.Printf("Warning: Email service initialization failed: %v", err)
		return nil, nil
	}

	aiService, err := services.NewAIService(db, emailService)
	if err != nil {
		log.Printf("Warning: AI service initialization failed: %v", err)
		return emailService, nil
	}

	return emailService, aiService
}

// setupScheduler configures and returns the background task scheduler
func setupScheduler(ctx context.Context, db *gorm.DB, emailService *services.EmailService) *helper.Scheduler {
	scheduler := helper.NewScheduler(ctx)

	// Token cleanup task
	scheduler.AddTask("token-cleanup", time.Hour, func() error {
		return helper.CleanupExpiredTokens(db)
	})

	// Email retry task (if email service is available)
	if emailService != nil {
		interval := getEmailRetryInterval()
		scheduler.AddTask("email-retry", interval, func() error {
			return emailService.RetryFailedEmails()
		})
	}

	// Account anonymization task - runs daily to anonymize accounts past grace period
	accountDeletionInterval := getAccountDeletionInterval()
	gracePeriod := helper.GetGracePeriodDays()
	scheduler.AddTask("account-deletion", accountDeletionInterval, func() error {
		return services.AnonymizeExpiredAccounts(db, gracePeriod)
	})

	return scheduler
}

// getEmailRetryInterval reads the email retry interval from environment or returns default
func getEmailRetryInterval() time.Duration {
	defaultInterval := 5 * time.Minute

	intervalStr, hasInterval := os.LookupEnv("EMAIL_RETRY_INTERVAL_MINUTES")
	if !hasInterval {
		return defaultInterval
	}

	minutes, err := strconv.Atoi(intervalStr)
	if err != nil || minutes <= 0 {
		return defaultInterval
	}

	return time.Duration(minutes) * time.Minute
}

// getAccountDeletionInterval reads the account anonymization check interval from environment or returns default
func getAccountDeletionInterval() time.Duration {
	defaultInterval := 24 * time.Hour // Check once per day by default (PDPA compliant anonymization)

	intervalStr, hasInterval := os.LookupEnv("ACCOUNT_DELETION_CHECK_INTERVAL_HOURS")
	if !hasInterval {
		return defaultInterval
	}

	hours, err := strconv.Atoi(intervalStr)
	if err != nil || hours <= 0 {
		return defaultInterval
	}

	return time.Duration(hours) * time.Hour
}

// setupRouter configures the Gin router with middleware and routes
func setupRouter(db *gorm.DB, redisClient *redis.Client, emailService *services.EmailService, aiService *services.AIService) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	corsConfig := middlewares.SetupCORS()
	router.Use(cors.New(corsConfig))

	// Application routes
	if err := handlers.SetupRoutes(router, db, redisClient, emailService, aiService); err != nil {
		log.Fatal("Failed to setup routes:", err)
	}

	// Swagger documentation
	setupSwagger(router)

	return router
}

// setupSwagger configures Swagger documentation endpoint
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

// startServer creates and starts the HTTP server in a goroutine
func startServer(router *gin.Engine) *http.Server {
	listenAddress := getListenAddress()

	srv := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on %s", listenAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	return srv
}

// getListenAddress returns the server listen address from environment or default
func getListenAddress() string {
	listenAddress, hasAddress := os.LookupEnv("LISTEN_ADDRESS")
	if !hasAddress {
		return ":8000"
	}
	return listenAddress
}

// waitForShutdownSignal blocks until an interrupt or termination signal is received
func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received, starting graceful shutdown...")
}

// performGracefulShutdown gracefully shuts down all services
func performGracefulShutdown(srv *http.Server, cancel context.CancelFunc, scheduler *helper.Scheduler, redisClient *redis.Client) {
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Stop scheduler
	cancel()
	scheduler.Wait()

	// Close Redis connection
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}

	log.Println("Server stopped gracefully")
}
