package main

import (
	"errors"
	"fmt"
	"ku-work/backend/model"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func LoadDB() (*gorm.DB, error) {
	db_username, db_has_username := os.LookupEnv("DB_USERNAME")
	if !db_has_username {
		return nil, errors.New("no database username specified")
	}
	db_password, db_has_password := os.LookupEnv("DB_PASSWORD")
	if !db_has_password {
		return nil, errors.New("no database password specified")
	}
	db_host, db_has_host := os.LookupEnv("DB_HOST")
	if !db_has_host {
		return nil, errors.New("no database host specified")
	}
	db_port, db_has_port := os.LookupEnv("DB_PORT")
	if !db_has_port {
		return nil, errors.New("no database port specified")
	}
	db_name, db_has_name := os.LookupEnv("DB_NAME")
	if !db_has_name {
		return nil, errors.New("no database name specified")
	}
	connection_info := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		url.PathEscape(db_username),
		url.PathEscape(db_password),
		db_host,
		db_port,
		url.PathEscape(db_name),
	)
	db, err := gorm.Open(postgres.Open(connection_info), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&model.User{})
	return db, nil
}

func main() {
	godotenv.Load()
	db, err := LoadDB()
	if err != nil {
		return
	}
	router := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()

	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		corsConfig.AllowOrigins = strings.Split(allowedOrigins, ",")
		for i, origin := range corsConfig.AllowOrigins {
			corsConfig.AllowOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// Default to allow localhost for development
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	// Get allowed methods from environment variable
	allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
	if allowedMethods != "" {
		corsConfig.AllowMethods = strings.Split(allowedMethods, ",")
		for i, method := range corsConfig.AllowMethods {
			corsConfig.AllowMethods[i] = strings.TrimSpace(method)
		}
	} else {
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	// Get allowed headers from environment variable
	allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	if allowedHeaders != "" {
		corsConfig.AllowHeaders = strings.Split(allowedHeaders, ",")
		for i, header := range corsConfig.AllowHeaders {
			corsConfig.AllowHeaders[i] = strings.TrimSpace(header)
		}
	} else {
		corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	}

	// Allow credentials if specified in environment
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS")
	if allowCredentials == "true" {
		corsConfig.AllowCredentials = true
	}

	// Set max age if specified in environment
	maxAge := os.Getenv("CORS_MAX_AGE")
	if maxAge != "" {
		corsConfig.MaxAge = 12 * 60 * 60 // 12 hours default
	}

	router.Use(cors.New(corsConfig))
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Simple response",
		})
	})
	router.GET("/users", func(ctx *gin.Context) {
		var users []model.User
		result := db.Model(&model.User{}).Find(&users)
		if result.Error != nil {
			ctx.String(http.StatusInternalServerError, result.Error.Error())
			return
		}
		ctx.JSON(200, gin.H{
			"users": users,
		})
	})
	router.POST("/create_user", func(ctx *gin.Context) {
		type CreateUserInput struct {
			Username string `json:"user" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		input := CreateUserInput{}
		err := ctx.Bind(&input)
		if err != nil {
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		user := model.User{
			Username: input.Username,
			Password: hashedPassword,
			IsAdmin:  false,
		}
		result := db.Create(&user)
		if result.Error != nil {
			ctx.String(http.StatusInternalServerError, result.Error.Error())
			return
		}
		ctx.String(http.StatusOK, "OK ;)")
	})
	listen_address, has_listen_address := os.LookupEnv("LISTEN_ADDRESS")
	if !has_listen_address {
		listen_address = ":8000"
	}
	router.Run(listen_address)
}
