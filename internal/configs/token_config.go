package configs

import (
	"os"
	"time"
)

// Config holds the application configuration settings.
type TokenConfig struct {
	JWTSecret       string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

// Load reads configuration from environment variables.
// It parses token durations and falls back to default values if parsing fails.
func LoadTokenConfig() (*TokenConfig, error) {
	accessDur, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		accessDur = 10 * time.Minute
	}

	refreshDur, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		refreshDur = 48 * time.Hour
	}

	return &TokenConfig{
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessDuration:  accessDur,
		RefreshDuration: refreshDur,
	}, nil
}
