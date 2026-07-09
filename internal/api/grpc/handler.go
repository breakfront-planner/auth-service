package grpc

import (
	"context"
	"errors"

	authv1 "github.com/breakfront-planner/proto/gen/go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	authv1.UnimplementedAuthServiceServer
	authService    IAuthService
	credentialsCfg *configs.CredentialsConfig
}

func NewAuthHandler(authService IAuthService, credentialsCfg *configs.CredentialsConfig) *AuthHandler {
	return &AuthHandler{authService: authService, credentialsCfg: credentialsCfg}
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.CredentialsRequest) (*authv1.TokenPairResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "login and password are required")
	}

	accessToken, refreshToken, err := h.authService.Login(req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.TokenPairResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.CredentialsRequest) (*authv1.TokenPairResponse, error) {
	if err := h.credentialsCfg.ValidateCredentials(req.GetLogin(), req.GetPassword()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accessToken, refreshToken, err := h.authService.Register(req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.TokenPairResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *authv1.TokenRequest) (*authv1.TokenPairResponse, error) {
	accessToken, refreshToken, err := h.authService.Refresh(req.GetRefreshToken())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.TokenPairResponse{
		AccessToken:  accessToken.Value,
		RefreshToken: refreshToken.Value,
	}, nil
}

// Logout always succeeds regardless of token validity, to avoid leaking
// whether a given refresh token exists (mirrors the HTTP handler's behavior).
func (h *AuthHandler) Logout(ctx context.Context, req *authv1.TokenRequest) (*authv1.LogoutResponse, error) {
	h.authService.Logout(req.GetRefreshToken())
	return &authv1.LogoutResponse{}, nil
}

func mapError(err error) error {
	if errors.Is(err, autherrors.ErrInvalidCredentials) {
		return status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if errors.Is(err, autherrors.ErrLoginTaken) {
		return status.Error(codes.AlreadyExists, "login already taken")
	}
	if errors.Is(err, autherrors.ErrTokenExpired) ||
		errors.Is(err, autherrors.ErrTokenType) ||
		errors.Is(err, autherrors.ErrInvalidJWT) ||
		errors.Is(err, autherrors.ErrTokenSignMethod) ||
		errors.Is(err, autherrors.ErrInvalidUserID) {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}
	return status.Error(codes.Internal, "internal server error")
}
