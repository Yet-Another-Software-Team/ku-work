package bootstrap

import (
	"context"
	"log"
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App aggregates all runtime components of the server.
// Construction of App is done by the bootstrap wiring code.
// This file only defines the type and lifecycle helpers.
type App struct {
	// Core infrastructure
	DB    *gorm.DB
	Redis *redis.Client

	// HTTP stack
	Router *gin.Engine
	Server *http.Server

	// Background scheduler
	Scheduler  *services.Scheduler
	cancelFunc context.CancelFunc
}

// Start launches the HTTP server in a goroutine.
// It logs startup and captures any fatal errors.
func (a *App) Start() {
	if a == nil || a.Server == nil {
		return
	}
	go func() {
		addr := a.Server.Addr
		if addr == "" {
			addr = ":http"
		}
		log.Printf("Starting server on %s", addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
}

// Shutdown gracefully shuts down the HTTP server, stops background schedulers,
// cancels the bootstrap context, and closes external connections.
func (a *App) Shutdown(ctx context.Context) {
	if a == nil {
		return
	}

	// Shutdown HTTP server
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}

	// Cancel any bootstrap-level context
	if a.cancelFunc != nil {
		a.cancelFunc()
	}

	// Wait for background tasks to finish
	if a.Scheduler != nil {
		a.Scheduler.Wait()
	}

	// Close Redis connection
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}

	log.Println("Server stopped gracefully")
}
