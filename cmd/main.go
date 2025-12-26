package main

import (
	"log"

	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/database"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/repositories"
	"github.com/breakfront-planner/auth-service/internal/services"

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
	defer db.Close()
	log.Println("DB connected")

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Migrations ok")

	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	cfg, _ := configs.Load()
	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.AccessDuration, cfg.RefreshDuration)

	hashService := services.NewHashService()
	userService := services.NewUserService(userRepo, hashService)
	tokenService := services.NewTokenService(tokenRepo, hashService, jwtManager)
	authService := services.NewAuthService(tokenService, userService)
	login := os.Getenv("TEST_LOGIN")
	pass := os.Getenv("TEST_PASS")

	/*


		accessToken, refreshToken, err := authService.Register(login, pass)

		if err != nil {
			log.Fatal("registration failed: ", err)
		}

		log.Println("registration success")
		log.Printf("refresh token expires at: %v", refreshToken.ExpiresAt)
		log.Printf("access token expires at: %v", accessToken.ExpiresAt)
	*/

	newAccessToken, newRefreshToken, err := authService.Login(login, pass)

	if err != nil {
		log.Fatal("login failed: ", err)
	}

	log.Println("login success")
	log.Printf("new refresh token expires at: %v", newRefreshToken.ExpiresAt)
	log.Printf("new access token expires at: %v", newAccessToken.ExpiresAt)

	oldRefreshTokenValue := newRefreshToken.Value

	newAccessToken, newRefreshToken, err = authService.Refresh(oldRefreshTokenValue, login)

	if err != nil {
		log.Fatal("refresh failed: ", err)
	}

	log.Println("refresh success")
	log.Printf("new refresh token expires at: %v", newRefreshToken.ExpiresAt)
	log.Printf("new access token expires at: %v", newAccessToken.ExpiresAt)

}
