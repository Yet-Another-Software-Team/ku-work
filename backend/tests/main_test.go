package tests

import (
	"context"
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

func CreateAdminUser(name string) (*model.User, error) {
	var user model.User
	if err := db.Transaction(func(tx *gorm.DB) error {
		user = model.User{
			Username: name,
		}
		if result := tx.Create(&user); result.Error != nil {
			return result.Error
		}
		admin := model.Admin{
			UserID: user.ID,
		}
		if result := tx.Create(&admin); result.Error != nil {
			return result.Error
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &user, nil
}
