package main

import (
	"log"
	"net/http"

	"github.com/breakfront-planner/auth-service/internal/api"
	"github.com/breakfront-planner/auth-service/internal/database"

	// Register database drivers
	_ "github.com/lib/pq"

	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/repositories"
	"github.com/breakfront-planner/auth-service/internal/services"
	"github.com/breakfront-planner/auth-service/internal/validators"

	"os"

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

	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	tokenCfg, _ := configs.LoadTokenConfig()
	jwtManager := jwt.NewManager(tokenCfg.JWTSecret, tokenCfg.AccessDuration, tokenCfg.RefreshDuration)

	hashService := services.NewHashService()
	userService := services.NewUserService(userRepo, hashService)
	tokenService := services.NewTokenService(tokenRepo, hashService, jwtManager)
	validator := validators.NewTokenValidator(jwtManager, userService)
	authService := services.NewAuthService(tokenService, userService, validator)

	credentialsCfg, _ := configs.LoadCredentialsConfig()
	authHandler := api.NewAuthHandler(authService, credentialsCfg)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter(authHandler)
	server := &http.Server{Addr: ":" + port, Handler: router}

	log.Printf("Server starting on :%s", port)
	log.Fatal(server.ListenAndServe())

}
