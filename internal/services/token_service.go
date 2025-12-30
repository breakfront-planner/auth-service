package services

import (
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/models"
)

type ITokenRepository interface {
	SaveToken(token *models.Token) error
	RevokeToken(token *models.Token) error
	CheckToken(token *models.Token) error
}

type IHashService interface {
	HashToken(token string) string
	HashPassword(password string) (string, error)
	ComparePasswords(passHash, input string) error
}

type TokenService struct {
	tokenRepo   ITokenRepository
	hashService IHashService
	jwtManager  *jwt.JWTManager
}

func NewTokenService(tokenRepo ITokenRepository, hashService IHashService, jwtManager *jwt.JWTManager) *TokenService {
	return &TokenService{
		tokenRepo:   tokenRepo,
		hashService: hashService,
		jwtManager:  jwtManager,
	}
}

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

func (s *TokenService) Refresh(refreshToken *models.Token, user *models.User) (newAccessToken, newRefreshToken *models.Token, err error) {

	refreshToken.HashedValue = s.hashService.HashToken(refreshToken.Value)

	err = s.tokenRepo.CheckToken(refreshToken)
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

func (s *TokenService) RevokeToken(token *models.Token) error {

	token.HashedValue = s.hashService.HashToken(token.Value)
	err := s.tokenRepo.CheckToken(token)
	if err != nil {
		return autherrors.ErrInvalidToken(err)
	}

	*token.RevokedAt = time.Now().UTC()
	err = s.tokenRepo.RevokeToken(token)
	if err != nil {
		return autherrors.ErrRevokeToken(err)
	}

	return nil
}
