package main

import (
	"errors"
	"fmt"
	"ku-work/backend/model"
	"net/http"
	"net/url"
	"os"

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
			Username string `json:"usr" binding:"required"`
			Password string `json:"passwd" binding:"required"`
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
		user := model.User {
			Username: input.Username,
			Password: hashedPassword,
			IsAdmin: false,
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
