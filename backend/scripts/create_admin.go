package main

import (
	"fmt"
	"ku-work/backend/database"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log"
	"os"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/term"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./create_admin <username>")
		return
	}

	username := os.Args[1]

	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println()

	_ = godotenv.Load()

	db, err := database.LoadDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check if user already exists
	var count int64
	err = db.Model(&model.User{}).Where("username = ? AND user_type = ?", username, "admin").Count(&count).Error
	if count > 0 {
		log.Fatalf("User with username '%s' already exists.", username)
	} else if err != nil {
		log.Fatalf("Failed to check for existing user: %v", err)
	}

	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		user := model.User{
			Username:     username,
			UserType:     "admin",
			PasswordHash: hashedPassword,
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		admin := model.Admin{
			UserID: user.ID,
		}

		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Admin user '%s' created successfully.\n", username)
}
