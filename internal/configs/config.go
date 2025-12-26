package configs

import (
	"os"
	"time"
)

type Config struct {
	JWTSecret       string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

func Load() (*Config, error) {
	accessDur, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		accessDur = 2 * time.Hour // дефолт
	}

	refreshDur, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	if err != nil {
		refreshDur = 48 * time.Hour // дефолт
	}

	return &Config{
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AccessDuration:  accessDur,
		RefreshDuration: refreshDur,
	}, nil
}
