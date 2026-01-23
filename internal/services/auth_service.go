package services

import (
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/breakfront-planner/auth-service/internal/validators"
)

// IUserService defines the interface for user-related operations.
type IUserService interface {
	CreateUser(login string, passHash string) (*models.User, error)
	FindUser(*models.UserFilter) (*models.User, error)
	CheckPassword(login string, password string) error
}

// ITokenService defines the interface for token management operations.
type ITokenService interface {
	CreateNewTokenPair(user *models.User) (accessToken, refreshToken *models.Token, err error)
	Refresh(refreshToken *models.Token, user *models.User) (newAccessToken, newRefreshToken *models.Token, err error)
	RevokeToken(token *models.Token) error
}

// ITokenValidator defines the interface for token validation.
type ITokenValidator interface {
	ValidateRefreshToken(tokenValue string) (*models.ParsedToken, error)
	ValidateAccessToken(tokenValue string) (*models.ParsedToken, error)
	Validate(tokenValue string, opts ...validators.ValidationOption) (*models.ParsedToken, error)
}

// AuthService provides authentication and authorization functionality.
// It coordinates between user, token, and validation services to handle registration, login, and logout flows.
type AuthService struct {
	tokenService   ITokenService
	userService    IUserService
	tokenValidator ITokenValidator
}

// NewAuthService creates a new authentication service instance.
func NewAuthService(tokenService ITokenService, userService IUserService, tokenValidator ITokenValidator) *AuthService {
	return &AuthService{
		tokenService:   tokenService,
		userService:    userService,
		tokenValidator: tokenValidator,
	}
}

// Register creates a new user account and returns access and refresh tokens.
// Returns an error if the user already exists or if token generation fails.
func (s *AuthService) Register(login string, password string) (accessToken, refreshToken *models.Token, err error) {

	user, err := s.userService.CreateUser(login, password)

	if err != nil {
		return nil, nil, err
	}

	return s.tokenService.CreateNewTokenPair(user)

}

// Login authenticates a user with their credentials and returns access and refresh tokens.
// Returns an error if credentials are invalid or token generation fails.
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

// Refresh generates a new token pair using a valid refresh token.
// The old refresh token is revoked after successful generation of new tokens.
func (s *AuthService) Refresh(oldRefreshTokenValue string) (newAccessToken, newRefreshToken *models.Token, err error) {
	// Validate refresh token using the validator
	parsedToken, err := s.tokenValidator.ValidateRefreshToken(oldRefreshTokenValue)
	if err != nil {
		return nil, nil, err
	}

	oldRefreshToken := models.Token{
		UserID: parsedToken.UserID,
		Value:  oldRefreshTokenValue,
	}

	user := models.User{
		ID: parsedToken.UserID,
	}

	return s.tokenService.Refresh(&oldRefreshToken, &user)
}

// Logout invalidates the user's refresh token, effectively ending their session.
func (s *AuthService) Logout(refreshTokenValue string) error {
	// Validate refresh token using the validator
	parsedToken, err := s.tokenValidator.ValidateRefreshToken(refreshTokenValue)
	if err != nil {
		return err
	}

	tokenToRevoke := models.Token{
		UserID: parsedToken.UserID,
		Value:  refreshTokenValue,
	}

	return s.tokenService.RevokeToken(&tokenToRevoke)
}
