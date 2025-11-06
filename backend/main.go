package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ku-work/backend/bootstrap"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	app, err := bootstrap.BuildApp()
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}

	// Start HTTP server
	app.Start()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received, starting graceful shutdown...")

	// Perform graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	app.Shutdown(ctx)
}
