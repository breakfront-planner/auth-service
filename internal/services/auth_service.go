package services

import (
	"github.com/breakfront-planner/auth-service/internal/models"
)

type IUserService interface {
	CreateUser(login string, passHash string) (*models.User, error)
	FindUser(*models.UserFilter) (*models.User, error)
	CheckPassword(login string, password string) error
}

type ITokenService interface {
	CreateNewTokenPair(user *models.User) (accessToken, refreshToken *models.Token, err error)
	Refresh(refreshToken *models.Token, user *models.User) (newAccessToken, newRefreshToken *models.Token, err error)
}

type AuthService struct {
	tokenService ITokenService
	userService  IUserService
}

func NewAuthService(tokenService ITokenService, userService IUserService) *AuthService {
	return &AuthService{
		tokenService: tokenService,
		userService:  userService,
	}
}

func (s *AuthService) Register(login string, password string) (accessToken, refreshToken *models.Token, err error) {

	user, err := s.userService.CreateUser(login, password)

	if err != nil {
		return nil, nil, err
	}

	return s.tokenService.CreateNewTokenPair(user)

}

func (s *AuthService) Login(login string, password string) (accessToken, refreshToken *models.Token, err error) {

	err = s.userService.CheckPassword(login, password)
	if err != nil {
		return nil, nil, err
	}
	filter := models.UserFilter{
		Login: &login,
	}

	user, err := s.userService.FindUser(&filter)
	if err != nil {
		return nil, nil, err
	}

	return s.tokenService.CreateNewTokenPair(user)

}

func (s *AuthService) Refresh(oldRefreshTokenValue string, login string) (newAccessToken, newRefreshToken *models.Token, err error) {

	newUserFilter := models.UserFilter{
		Login: &login,
	}

	user, err := s.userService.FindUser(&newUserFilter)

	if err != nil {
		return nil, nil, err
	}

	oldRefreshToken := models.Token{
		UserID: user.ID,
		Value:  oldRefreshTokenValue,
	}

	return s.tokenService.Refresh(&oldRefreshToken, user)

}
