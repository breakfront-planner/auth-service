package configs

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const configPath = ".env"

type Configuration struct {
	Credentials CredentialsConfig
	Token       TokenConfig
	GRPCServer  GRPCServerConfig
}

type CredentialsConfig struct {
	LoginMinLen    int `env:"LOGIN_MIN_LEN" envDefault:"3"`
	LoginMaxLen    int `env:"LOGIN_MAX_LEN" envDefault:"32"`
	PasswordMinLen int `env:"PASSWORD_MIN_LEN" envDefault:"8"`
	PasswordMaxLen int `env:"PASSWORD_MAX_LEN" envDefault:"64"`
}

type TokenConfig struct {
	JWTSecret       string        `env:"JWT_SECRET"`
	AccessDuration  time.Duration `env:"ACCESS_TOKEN_DURATION" envDefault:"10m"`
	RefreshDuration time.Duration `env:"REFRESH_TOKEN_DURATION" envDefault:"48h"`
}

type GRPCServerConfig struct {
	GRPCServerAddress string `env:"GRPC_SERVER_ADDRESS" envDefault:":50051"`
}

func New() (*Configuration, error) {
	return LoadAndParseConfig(configPath)
}

func LoadAndParseConfig(path string) (*Configuration, error) {
	if err := loadEnvFile(path); err != nil {
		return nil, err
	}

	cfg := &Configuration{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	return cfg, nil
}

func loadEnvFile(path string) error {
	if err := godotenv.Load(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("load env file %s: %w", path, err)
	}
	return nil
}
