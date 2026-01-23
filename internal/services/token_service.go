package services

import (
	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/models"
)

// ITokenRepository defines the interface for token data persistence operations.
type ITokenRepository interface {
	SaveToken(token *models.Token) error
	RevokeToken(token *models.Token) error
	FindToken(token *models.Token) error
}

// IHashService defines the interface for hashing operations.
type IHashService interface {
	HashToken(token string) string
	HashPassword(password string) (string, error)
	ComparePasswords(passHash, input string) error
}

// TokenService manages JWT token lifecycle including creation, validation, and revocation.
type TokenService struct {
	tokenRepo   ITokenRepository
	hashService IHashService
	jwtManager  *jwt.Manager
}

// NewTokenService creates a new token service instance.
func NewTokenService(tokenRepo ITokenRepository, hashService IHashService, jwtManager *jwt.Manager) *TokenService {
	return &TokenService{
		tokenRepo:   tokenRepo,
		hashService: hashService,
		jwtManager:  jwtManager,
	}
}

// CreateNewTokenPair generates a new access and refresh token pair for the user.
// The refresh token is hashed and persisted in the repository.
func (s *TokenService) CreateNewTokenPair(user *models.User) (accessToken, refreshToken *models.Token, err error) {

	accessToken, err = s.jwtManager.GenerateToken(user, constants.TokenTypeAccess)
	if err != nil {
		return nil, nil, autherrors.ErrCreateToken(err)
	}

	refreshToken, err = s.jwtManager.GenerateToken(user, constants.TokenTypeRefresh)
	if err != nil {
		return nil, nil, autherrors.ErrCreateToken(err)
	}

	refreshToken.HashedValue = s.hashService.HashToken(refreshToken.Value)

	err = s.tokenRepo.SaveToken(refreshToken)
	if err != nil {
		return nil, nil, autherrors.ErrSaveToken(err)
	}

	return accessToken, refreshToken, nil

}

// Refresh validates the provided refresh token and generates a new token pair.
// The old refresh token is revoked after successful validation.
func (s *TokenService) Refresh(refreshToken *models.Token, user *models.User) (newAccessToken, newRefreshToken *models.Token, err error) {

	refreshToken.HashedValue = s.hashService.HashToken(refreshToken.Value)

	err = s.tokenRepo.FindToken(refreshToken)
	if err != nil {
		return nil, nil, autherrors.ErrRefreshToken(err)
	}

	newAccessToken, newRefreshToken, err = s.CreateNewTokenPair(user)
	if err != nil {
		return nil, nil, autherrors.ErrCreateToken(err)
	}

	err = s.tokenRepo.RevokeToken(refreshToken)
	if err != nil {
		return newAccessToken, newRefreshToken, autherrors.ErrRevokeToken(err)
	}

	return newAccessToken, newRefreshToken, nil

}

// RevokeToken invalidates the specified token by marking it as revoked in the repository.
func (s *TokenService) RevokeToken(token *models.Token) error {

	token.HashedValue = s.hashService.HashToken(token.Value)
	err := s.tokenRepo.FindToken(token)
	if err != nil {
		return err
	}

	err = s.tokenRepo.RevokeToken(token)
	if err != nil {
		return autherrors.ErrRevokeToken(err)
	}

	return nil
}
