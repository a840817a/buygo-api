package main

import (
	"log"
	"os"

	"github.com/buygo/buygo-api/internal/adapter/db"
	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
)

func main() {
	dbConfig := db.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "buygo",
		Password: "password",
		DBName:   "buygo",
		SSLMode:  "disable",
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		dbConfig.Host = host
	}

	database, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Upsert Admin User
	// Mock Token 'mock-token-admin' -> UID: 'admin', Email: 'admin@example.com'

	uid := "admin"
	email := "admin@example.com"
	role := 3 // USER_ROLE_SYS_ADMIN

	var user model.User
	result := database.Where("email = ?", email).First(&user)

	if result.Error != nil {
		log.Printf("Creating admin user %s...", email)
		user = model.User{
			ID:    uid,
			Email: email,
			Name:  "Admin User",
			Role:  role,
		}
		if err := database.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
	} else {
		log.Printf("Updating admin user %s role to %d...", email, role)
		if err := database.Model(&user).Update("role", role).Error; err != nil {
			log.Fatalf("Failed to update admin: %v", err)
		}
	}
	log.Println("Admin seeded successfully.")
}
