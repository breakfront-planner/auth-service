package api

import (
	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/configs"
)

func validateCredentials(req CredentialsRequest, cfg *configs.CredentialsConfig) error {
	if len(req.Login) < cfg.LoginMinLen || len(req.Login) > cfg.LoginMaxLen {
		return autherrors.ErrLoginLength(cfg.LoginMinLen, cfg.LoginMaxLen)
	}
	if len(req.Password) < cfg.PasswordMinLen || len(req.Password) > cfg.PasswordMaxLen {
		return autherrors.ErrPasswordLength(cfg.PasswordMinLen, cfg.PasswordMaxLen)
	}
	return nil
}
