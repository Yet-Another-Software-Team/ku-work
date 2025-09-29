package tests

import (
	"context"
	"fmt"
	"ku-work/backend/database"
	"ku-work/backend/handlers"
	"ku-work/backend/model"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var db *gorm.DB
var router *gin.Engine

func TestMain(m *testing.M) {
	ctx := context.Background()
	// Check for podman socket and set DOCKER_HOST if present to make it compatible with podman
	podmanSocketPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
	if _, err := os.Stat(podmanSocketPath); err == nil {
		_ = os.Setenv("DOCKER_HOST", "unix://"+podmanSocketPath)
	}
	_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
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
	port, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		panic(err)
	}
	// idk why but sometimes it couldn't connect to db even though it said ready to accept connections in the logs
	time.Sleep(500000000)
	_ = os.Setenv("DB_USERNAME", "tester")
	_ = os.Setenv("DB_PASSWORD", "1234")
	_ = os.Setenv("DB_HOST", "127.0.0.1")
	_ = os.Setenv("DB_PORT", port.Port())
	_ = os.Setenv("DB_NAME", "kuwork")
	db, err = database.LoadDB()
	if err != nil {
		_ = testcontainers.TerminateContainer(postgresContainer)
		panic(err)
	}
	_ = os.Setenv("JWT_SECRET", "1234")
	_ = os.Setenv("GOOGLE_CLIENT_SECRET", "GOCSPX-idklmao")
	_ = os.Setenv("GOOGLE_CLIENT_ID", "012345678901-1md5idklmao.apps.googleusercontent.com")
	router = gin.Default()
	handlers.SetupRoutes(router, db)
	code := m.Run()
	_ = testcontainers.TerminateContainer(postgresContainer)
	os.Exit(code)
}

type UserCreationInfo struct {
	Username  string
	IsAdmin   bool
	IsCompany bool
	IsStudent bool
}

type UserCreationResult struct {
	User    model.User
	Admin   *model.Admin
	Company *model.Company
	Student *model.Student
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
			Approved:            true,
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
