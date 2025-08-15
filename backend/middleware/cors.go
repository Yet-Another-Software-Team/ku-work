package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
)

func SetupCORS() cors.Config {
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

	return corsConfig
}
