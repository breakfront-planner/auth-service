package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/models"
)

type IAuthService interface {
	Login(login, password string) (accessToken, refreshToken *models.Token, err error)
	Register(login, password string) (accessToken, refreshToken *models.Token, err error)
	Refresh(oldRefreshTokenValue string) (newAccessToken, newRefreshToken *models.Token, err error)
	Logout(refreshTokenValue string) error
}

type AuthHandler struct {
	authService    IAuthService
	credentialsCfg *configs.CredentialsConfig
}

func NewAuthHandler(authService IAuthService, credentialsCfg *configs.CredentialsConfig) *AuthHandler {
	return &AuthHandler{authService: authService, credentialsCfg: credentialsCfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "login and password are required")
		return
	}

	accessToken, refreshToken, err := h.authService.Login(req.Login, req.Password)
	if err != nil {
		status, msg := mapError(err)
		if status == http.StatusInternalServerError {
			slog.Error("login failed", "error", err)
		}
		writeError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, TokenPairResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validateCredentials(req, h.credentialsCfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authService.Register(req.Login, req.Password)
	if err != nil {
		status, msg := mapError(err)
		if status == http.StatusInternalServerError {
			slog.Error("registration failed", "error", err)
		}
		writeError(w, status, msg)
		return
	}

	writeJSON(w, http.StatusOK, TokenPairResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	})
}

func mapError(err error) (int, string) {
	if errors.Is(err, autherrors.ErrInvalidCredentials) {
		return http.StatusUnauthorized, "invalid credentials"
	}
	if errors.Is(err, autherrors.ErrLoginTaken) {
		return http.StatusConflict, "login already taken"
	}
	if errors.Is(err, autherrors.ErrTokenExpired) {
		return http.StatusUnauthorized, "token expired"
	}
	return http.StatusInternalServerError, "internal server error"
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
