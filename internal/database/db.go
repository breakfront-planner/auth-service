package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/breakfront-planner/auth-service/internal/configs"

	_ "github.com/lib/pq"
)

// Connect establishes a connection to the PostgreSQL database using the given configuration.
// It configures connection pool settings and verifies the connection with a ping.
func Connect(cfg configs.PostgresConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode)

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
