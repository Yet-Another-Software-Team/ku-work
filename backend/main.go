package main

import (
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/middlewares"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	env_err := godotenv.Load()
	if env_err != nil {
		return
	}

	db, db_err := database.LoadDB()
	if db_err != nil {
		return
	}

	router := gin.Default()

	// Setup CORS middleware
	corsConfig := middlewares.SetupCORS()
	router.Use(cors.New(corsConfig))

	// Setup routes
	handlers.SetupRoutes(router, db)

	listen_address, has_listen_address := os.LookupEnv("LISTEN_ADDRESS")
	if !has_listen_address {
		listen_address = ":8000"
	}
	run_err := router.Run(listen_address)
	if run_err != nil {
		log.Fatal("Server is Failed to Run")
	}
}
