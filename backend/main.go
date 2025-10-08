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

	router := gin.Default()

	// Setup CORS middleware
	corsConfig := middlewares.SetupCORS()
	router.Use(cors.New(corsConfig))

	// Setup routes
	handlers.SetupRoutes(router, db)

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
	run_err := router.Run(listen_address)
	if run_err != nil {
		log.Fatal("Server is Failed to Run")
	}
}
