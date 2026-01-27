package configs

import (
	_ "embed"
	"encoding/json"
)

//go:embed creds_config.json
var credentialsConfigData []byte

type CredentialsConfig struct {
	LoginMinLen    int `json:"login_min_len"`
	LoginMaxLen    int `json:"login_max_len"`
	PasswordMinLen int `json:"password_min_len"`
	PasswordMaxLen int `json:"password_max_len"`
}

func LoadCredentialsConfig() (*CredentialsConfig, error) {
	var cfg CredentialsConfig
	if err := json.Unmarshal(credentialsConfigData, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
