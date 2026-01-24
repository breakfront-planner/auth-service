package configs

import (
	"encoding/json"
	"os"
)

type CredentialsConfig struct {
	LoginMinLen    int `json:"login_min_len"`
	LoginMaxLen    int `json:"login_max_len"`
	PasswordMinLen int `json:"password_min_len"`
	PasswordMaxLen int `json:"password_max_len"`
}

func LoadCredentialsConfig(path string) (*CredentialsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg CredentialsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
