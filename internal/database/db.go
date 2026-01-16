package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
)

// Connect establishes a connection to the PostgreSQL database using environment variables.
// It validates required environment variables, configures connection pool settings,
// and verifies the connection with a ping.
func Connect() (*sql.DB, error) {

	requiredEnvVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

	envVars := make(map[string]string)
	var missingVars []string
	for _, varName := range requiredEnvVars {
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		} else {
			envVars[varName] = os.Getenv(varName)
		}
	}

	if len(missingVars) > 0 {
		return nil, autherrors.ErrMissingEnvVars(missingVars)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envVars["DB_HOST"],
		envVars["DB_PORT"],
		envVars["DB_USER"],
		envVars["DB_PASSWORD"],
		envVars["DB_NAME"],
		envVars["DB_SSLMODE"])

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
