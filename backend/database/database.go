package database

import (
	"errors"
	"fmt"
	"ku-work/backend/model"
	"net/url"
	"os"

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

	allModels := []any{
		&model.User{},
		&model.Admin{},
		&model.GoogleOAuthDetails{},
		&model.RefreshToken{},
		&model.Job{},
		&model.Student{},
	}

	db_err := db.AutoMigrate(allModels...)
	if db_err != nil {
		return nil, err
	}
	return db, nil
}
