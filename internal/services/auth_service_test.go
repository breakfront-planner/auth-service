package services

import (
	"errors"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/breakfront-planner/auth-service/internal/services/mocks"
)

type AuthServiceTestSuite struct {
	suite.Suite
	ctrl             *gomock.Controller
	mockUserService  *mocks.MockIUserService
	mockTokenService *mocks.MockITokenService
	authService      *AuthService
	testLogin        string
	testPassword     string
	testTokenValue   string
}

func (s *AuthServiceTestSuite) SetupSuite() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Failed to load .env.test")

	s.testLogin = os.Getenv("TEST_LOGIN")
	s.testPassword = os.Getenv("TEST_PASS")
	s.testTokenValue = os.Getenv("TOKEN_TEST_VALUE")

	require.NotEmpty(s.T(), s.testLogin, "TEST_LOGIN must be set in .env.test")
	require.NotEmpty(s.T(), s.testPassword, "TEST_PASS must be set in .env.test")
	require.NotEmpty(s.T(), s.testTokenValue, "TOKEN_TEST_VALUE must be set in .env.test")

}

func (s *AuthServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserService = mocks.NewMockIUserService(s.ctrl)
	s.mockTokenService = mocks.NewMockITokenService(s.ctrl)
	s.authService = NewAuthService(s.mockTokenService, s.mockUserService)
}

func (s *AuthServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AuthServiceTestSuite) TestRegisterSuccess() {
	s.mockUserService.EXPECT().
		CreateUser(s.testLogin, s.testPassword).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		CreateNewTokenPair(gomock.Any()).
		Return(&models.Token{}, &models.Token{}, nil)

	accessToken, refreshToken, err := s.authService.Register(s.testLogin, s.testPassword)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessToken)
	assert.NotNil(s.T(), refreshToken)
}

func (s *AuthServiceTestSuite) TestRegisterCreateUserError() {
	createUserError := errors.New("login already taken")

	s.mockUserService.EXPECT().
		CreateUser(s.testLogin, s.testPassword).
		Return(nil, createUserError)

	accessToken, refreshToken, err := s.authService.Register(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), accessToken)
	assert.Nil(s.T(), refreshToken)
	assert.ErrorContains(s.T(), err, "login already taken")
}

func (s *AuthServiceTestSuite) TestRegisterCreateTokenPairError() {
	tokenError := errors.New("failed to create token")

	s.mockUserService.EXPECT().
		CreateUser(s.testLogin, s.testPassword).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		CreateNewTokenPair(gomock.Any()).
		Return(nil, nil, tokenError)

	accessToken, refreshToken, err := s.authService.Register(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), accessToken)
	assert.Nil(s.T(), refreshToken)
	assert.ErrorContains(s.T(), err, "failed to create token")
}

func (s *AuthServiceTestSuite) TestLoginSuccess() {
	s.mockUserService.EXPECT().
		CheckPassword(s.testLogin, s.testPassword).
		Return(nil)

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		CreateNewTokenPair(gomock.Any()).
		Return(&models.Token{}, &models.Token{}, nil)

	accessToken, refreshToken, err := s.authService.Login(s.testLogin, s.testPassword)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessToken)
	assert.NotNil(s.T(), refreshToken)
}

func (s *AuthServiceTestSuite) TestLoginWrongPassword() {
	passwordError := errors.New("wrong password")

	s.mockUserService.EXPECT().
		CheckPassword(s.testLogin, s.testPassword).
		Return(passwordError)

	accessToken, refreshToken, err := s.authService.Login(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), accessToken)
	assert.Nil(s.T(), refreshToken)
	assert.ErrorContains(s.T(), err, "wrong password")
}

func (s *AuthServiceTestSuite) TestLoginCreateTokenPairError() {
	tokenError := errors.New("failed to create token")

	s.mockUserService.EXPECT().
		CheckPassword(s.testLogin, s.testPassword).
		Return(nil)

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		CreateNewTokenPair(gomock.Any()).
		Return(nil, nil, tokenError)

	accessToken, refreshToken, err := s.authService.Login(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), accessToken)
	assert.Nil(s.T(), refreshToken)
	assert.ErrorContains(s.T(), err, "failed to create token")
}

func (s *AuthServiceTestSuite) TestRefreshSuccess() {

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		Refresh(gomock.Any(), gomock.Any()).
		Return(&models.Token{}, &models.Token{}, nil)

	newAccessToken, newRefreshToken, err := s.authService.Refresh(s.testTokenValue, s.testLogin)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), newAccessToken)
	assert.NotNil(s.T(), newRefreshToken)
}

func (s *AuthServiceTestSuite) TestRefreshFindUserError() {
	findUserError := errors.New("user not found")

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, findUserError)

	newAccessToken, newRefreshToken, err := s.authService.Refresh(s.testTokenValue, s.testLogin)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), newAccessToken)
	assert.Nil(s.T(), newRefreshToken)
	assert.ErrorContains(s.T(), err, "user not found")
}

func (s *AuthServiceTestSuite) TestRefreshTokenServiceError() {

	refreshError := errors.New("invalid token")

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		Refresh(gomock.Any(), gomock.Any()).
		Return(nil, nil, refreshError)

	newAccessToken, newRefreshToken, err := s.authService.Refresh(s.testTokenValue, s.testLogin)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), newAccessToken)
	assert.Nil(s.T(), newRefreshToken)
	assert.ErrorContains(s.T(), err, "invalid token")
}

func (s *AuthServiceTestSuite) TestLogoutSuccess() {

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		RevokeToken(gomock.Any()).
		Return(nil)

	err := s.authService.Logout(s.testTokenValue, s.testLogin)

	assert.NoError(s.T(), err)
}

func (s *AuthServiceTestSuite) TestLogoutFindUserError() {
	findUserError := errors.New("user not found")

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, findUserError)

	err := s.authService.Logout(s.testTokenValue, s.testLogin)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "user not found")
}

func (s *AuthServiceTestSuite) TestLogoutRevokeTokenError() {
	revokeError := errors.New("failed to revoke token")

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(&models.User{}, nil)

	s.mockTokenService.EXPECT().
		RevokeToken(gomock.Any()).
		Return(revokeError)

	err := s.authService.Logout(s.testTokenValue, s.testLogin)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to revoke token")
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
