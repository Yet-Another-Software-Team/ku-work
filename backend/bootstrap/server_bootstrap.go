package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ku-work/backend/database"
	"ku-work/backend/helper"
	"ku-work/backend/services"
)

// BuildApp wires infrastructure, repositories, services, handlers, router, scheduler, and server.
func BuildApp() (*App, error) {
	// Infrastructure
	db, err := database.LoadDB()
	if err != nil {
		return nil, fmt.Errorf("database init failed: %w", err)
	}

	redisClient, err := database.LoadRedis()
	if err != nil {
		return nil, fmt.Errorf("redis init failed: %w", err)
	}

	// Repositories
	repos, err := NewRepositories(db, redisClient)
	if err != nil {
		return nil, fmt.Errorf("repository init failed: %w", err)
	}

	// Composition context for background services
	ctx, cancel := context.WithCancel(context.Background())

	// Services
	svcs, err := BuildServices(ctx, db, repos)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("services init failed: %w", err)
	}

	// Handlers
	handlers, err := BuildHandlers(db, svcs)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("handlers init failed: %w", err)
	}

	// Router
	router := NewRouter(RouterDeps{
		DB:       db,
		Redis:    redisClient,
		Services: svcs,
		Handlers: handlers,
	})

	// Scheduler and background tasks
	scheduler := services.NewScheduler(ctx)

	// 1) Token cleanup (refresh tokens)
	scheduler.AddTask("token-cleanup", time.Hour, func() error {
		return helper.CleanupExpiredTokens(db)
	})

	// 2) Email retry (optional)
	if svcs.Email != nil {
		scheduler.AddTask("email-retry", getEmailRetryInterval(), func() error {
			return svcs.Email.RetryFailedEmails()
		})
	}

	// 3) Account anonymization for soft-deleted users after grace period (daily)
	if svcs.Identity != nil {
		scheduler.AddTask("account-deletion", getAccountDeletionInterval(), func() error {
			return svcs.Identity.AnonymizeExpiredAccounts(context.Background(), helper.GetGracePeriodDays())
		})
	}

	scheduler.Start()

	// HTTP server
	server := &http.Server{
		Addr:    getListenAddress(),
		Handler: router,
	}

	return &App{
		DB:         db,
		Redis:      redisClient,
		Router:     router,
		Server:     server,
		Scheduler:  scheduler,
		cancelFunc: cancel,
	}, nil
}

// getListenAddress returns the server listen address from environment or default.
func getListenAddress() string {
	if addr := strings.TrimSpace(os.Getenv("LISTEN_ADDRESS")); addr != "" {
		return addr
	}
	return ":8000"
}

// getEmailRetryInterval reads the email retry interval from environment or returns default (5m).
func getEmailRetryInterval() time.Duration {
	const def = 5 * time.Minute
	if s := strings.TrimSpace(os.Getenv("EMAIL_RETRY_INTERVAL_MINUTES")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Minute
		}
	}
	return def
}

// getAccountDeletionInterval reads the account anonymization check interval from environment or returns default (24h).
func getAccountDeletionInterval() time.Duration {
	const def = 24 * time.Hour
	if s := strings.TrimSpace(os.Getenv("ACCOUNT_DELETION_CHECK_INTERVAL_HOURS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Hour
		}
	}
	return def
}
