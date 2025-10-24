package tests

import (
	"context"
	"fmt"
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/model"
	"ku-work/backend/services"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db *gorm.DB
var redisClient *redis.Client
var router *gin.Engine

// A 1x1 pixel black PNG for testing file uploads.
var pixel = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestMain(m *testing.M) {
	// Set cwd to parent, otherwise it can't find templates
	oldWorkingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(filepath.Dir(oldWorkingDir)); err != nil {
		panic(err)
	}
	ctx := context.Background()
	// Set XDG_RUNTIME_DIR if not set (required for testcontainers with Podman)
	if os.Getenv("XDG_RUNTIME_DIR") == "" {
		_ = os.Setenv("XDG_RUNTIME_DIR", fmt.Sprintf("/run/user/%d", os.Getuid()))
	}
	// Check for podman socket and set DOCKER_HOST if present to make it compatible with podman
	podmanSocketPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
	if _, err := os.Stat(podmanSocketPath); err == nil {
		_ = os.Setenv("DOCKER_HOST", "unix://"+podmanSocketPath)
	}
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	// Setup PostgreSQL testcontainer
	req := testcontainers.ContainerRequest{
		Name:         "kuwork-test-database",
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "tester",
			"POSTGRES_PASSWORD": "1234",
			"POSTGRES_DB":       "kuwork",
		},
		WaitingFor: wait.ForLog("ready to accept connections"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})

	if err != nil {
		panic(err)
	}

	// Setup Redis testcontainer
	redisReq := testcontainers.ContainerRequest{
		Name:         "kuwork-test-redis",
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		panic(err)
	}

	// Get PostgreSQL port
	port, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		_ = testcontainers.TerminateContainer(redisContainer)
		panic(err)
	}

	// idk why but sometimes it couldn't connect to db even though it said ready to accept connections in the logs
	time.Sleep(500000000)

	// Configure PostgreSQL
	_ = os.Setenv("DB_USERNAME", "tester")
	_ = os.Setenv("DB_PASSWORD", "1234")
	_ = os.Setenv("DB_HOST", "127.0.0.1")
	_ = os.Setenv("DB_PORT", port.Port())
	_ = os.Setenv("DB_NAME", "kuwork")

	db, err = database.LoadDB()
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		_ = testcontainers.TerminateContainer(redisContainer)
		panic(err)
	}

	// Get Redis port
	redisPort, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		_ = testcontainers.TerminateContainer(redisContainer)
		panic(err)
	}

	// Configure Redis
	_ = os.Setenv("REDIS_HOST", "127.0.0.1")
	_ = os.Setenv("REDIS_PORT", redisPort.Port())
	_ = os.Setenv("REDIS_PASSWORD", "")
	_ = os.Setenv("REDIS_DB", "0")

	redisClient, err = database.LoadRedis()
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		_ = testcontainers.TerminateContainer(redisContainer)
		panic(err)
	}

	// Configure other environment variables
	_ = os.Setenv("JWT_SECRET", "please-change-this-is-insecure!!")
	_ = os.Setenv("GOOGLE_CLIENT_SECRET", "GOCSPX-idklmao")
	_ = os.Setenv("GOOGLE_CLIENT_ID", "012345678901-1md5idklmao.apps.googleusercontent.com")
	_ = os.Setenv("APPROVAL_AI", "dummy")
	_ = os.Setenv("EMAIL_PROVIDER", "dummy")

	// Initialize Redis for rate limiting (optional for tests)
	redisClient, err = database.LoadRedis()
	if err != nil {
		// Redis is optional for tests - rate limiter will fail open
		redisClient = nil
	}

	// Initialize services for tests
	emailService, err := services.NewEmailService(db)
	if err != nil {
		panic(err)
	}

	aiService, err := services.NewAIService(db, emailService)
	if err != nil {
		panic(err)
	}

	router = gin.Default()
	if err := handlers.SetupRoutes(router, db, redisClient, emailService, aiService); err != nil {
		panic(err)
	}

	code := m.Run()

	// Clean up Redis connection
	if redisClient != nil {
		_ = redisClient.Close()
	}

	// Terminate containers
	_ = testcontainers.TerminateContainer(redisContainer)
	_ = testcontainers.TerminateContainer(postgresContainer)

	os.Exit(code)
}

type UserCreationInfo struct {
	Username  string
	IsAdmin   bool
	IsOAuth   bool
	IsCompany bool
	IsStudent bool
}

type UserCreationResult struct {
	User    model.User
	Admin   *model.Admin
	Company *model.Company
	Student *model.Student
	OAuth   *model.GoogleOAuthDetails
}

func CreateUser(config UserCreationInfo) (*UserCreationResult, error) {
	var result UserCreationResult
	if err := db.Transaction(func(tx *gorm.DB) error {
		result.User.Username = config.Username
		if result := tx.Create(&result.User); result.Error != nil {
			return result.Error
		}
		if config.IsAdmin {
			admin := model.Admin{
				UserID: result.User.ID,
			}
			if result := tx.Create(&admin); result.Error != nil {
				return result.Error
			}
			result.Admin = &admin
		}
		return nil
	}); err != nil {
		return nil, err
	}
	success := false
	if config.IsOAuth {
		oauth := model.GoogleOAuthDetails{
			UserID:     result.User.ID,
			ExternalID: "[External ID]",
			FirstName:  config.Username,
			LastName:   "LastName",
			Email:      fmt.Sprintf("%s@email.com", strings.ToLower(config.Username)),
		}
		if err := db.Create(&oauth).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&oauth)
			}
		})()
		result.OAuth = &oauth
	}
	if config.IsCompany {
		companyPhoto := model.File{
			UserID:   result.User.ID,
			FileType: model.FileTypePNG,
			Category: model.FileCategoryDocument,
		}
		if err := db.Create(&companyPhoto).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&companyPhoto)
			}
		})()
		companyBanner := model.File{
			UserID:   result.User.ID,
			FileType: model.FileTypePNG,
			Category: model.FileCategoryDocument,
		}
		if err := db.Create(&companyBanner).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&companyBanner)
			}
		})()
		company := model.Company{
			UserID:   result.User.ID,
			BannerID: companyBanner.ID,
			PhotoID:  companyPhoto.ID,
		}
		if err := db.Create(&company).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&company)
			}
		})()
		result.Company = &company
	}
	if config.IsStudent {
		photoFile := model.File{UserID: result.User.ID, FileType: model.FileTypeJPEG, Category: model.FileCategoryImage}
		if err := db.Create(&photoFile).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&photoFile)
			}
		})()
		statusFile := model.File{UserID: result.User.ID, FileType: model.FileTypePDF, Category: model.FileCategoryImage}
		if err := db.Create(&statusFile).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&statusFile)
			}
		})()
		student := model.Student{
			UserID:              result.User.ID,
			ApprovalStatus:      model.StudentApprovalAccepted,
			PhotoID:             photoFile.ID,
			StudentStatusFileID: statusFile.ID,
		}
		if err := db.Create(&student).Error; err != nil {
			return nil, err
		}
		defer (func() {
			if !success {
				_ = db.Delete(&student)
			}
		})()
		result.Student = &student
	}
	success = true
	return &result, nil
}
