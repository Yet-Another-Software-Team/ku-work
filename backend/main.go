package main

import (
	"context"
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/helper"
	"ku-work/backend/middlewares"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

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

	// Create files directory if it doesn't exist
	if err := os.MkdirAll("./files", 0755); err != nil {
		log.Printf("Failed to create files directory: %v", err)
		return
	}

	db, db_err := database.LoadDB()
	if db_err != nil {
		log.Printf("%v", db_err)
		return
	}

	// Initialize Redis for rate limiting
	redisClient, redis_err := database.LoadRedis()
	if redis_err != nil {
		log.Printf("Warning: Redis initialization failed: %v. Rate limiting will fail open.", redis_err)
		redisClient = nil
	} else {
		log.Println("Redis connected successfully")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup scheduler for background tasks
	scheduler := helper.NewScheduler(ctx)
	scheduler.AddTask("token-cleanup", time.Hour, func() error {
		return helper.CleanupExpiredTokens(db)
	})
	scheduler.AddTask("jwt-blacklist-cleanup", time.Hour, func() error {
		return helper.CleanupExpiredRevokedJWTs(db)
	})
	scheduler.Start()

	router := gin.Default()

	// Setup CORS middleware
	corsConfig := middlewares.SetupCORS()
	router.Use(cors.New(corsConfig))

	// Setup routes
	if err := handlers.SetupRoutes(router, db, redisClient); err != nil {
		log.Fatal(err)
	}

	// Setup Swagger
	swagger_host, has_swagger_host := os.LookupEnv("SWAGGER_HOST")
	if has_swagger_host {
		docs.SwaggerInfo.Host = swagger_host
	} else {
		docs.SwaggerInfo.Host = "localhost:8000"
	}

	docs.SwaggerInfo.BasePath = ""
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	listen_address, has_listen_address := os.LookupEnv("LISTEN_ADDRESS")
	if !has_listen_address {
		listen_address = ":8000"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    listen_address,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", listen_address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, starting graceful shutdown...")

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
