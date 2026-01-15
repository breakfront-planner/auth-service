package main

import (
	"log"

	"github.com/breakfront-planner/auth-service/internal/database"

	// Register database drivers
	_ "github.com/lib/pq"

	/*"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/repositories"
	"github.com/breakfront-planner/auth-service/internal/services"
	"github.com/breakfront-planner/auth-service/internal/configs"

	"os"
	*/

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()
	log.Println("DB connected")

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Migrations ok")

	/*

		userRepo := repositories.NewUserRepository(db)
		tokenRepo := repositories.NewTokenRepository(db)

		cfg, _ := configs.Load()
		jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.AccessDuration, cfg.RefreshDuration)

		hashService := services.NewHashService()
		userService := services.NewUserService(userRepo, hashService)
		tokenService := services.NewTokenService(tokenRepo, hashService, jwtManager)
		authService := services.NewAuthService(tokenService, userService)

	*/

}
