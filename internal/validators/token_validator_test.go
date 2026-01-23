package validators

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/breakfront-planner/auth-service/internal/validators/mocks"
)

type TokenValidatorTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockUserService *mocks.MockIUserService
	jwtManager      *jwt.Manager
	validator       *TokenValidator
	testUser        *models.User
	validToken      string
	expiredToken    string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func (s *TokenValidatorTestSuite) SetupSuite() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Failed to load .env.test")

	jwtSecret := os.Getenv("TEST_JWT_SECRET")
	require.NotEmpty(s.T(), jwtSecret, "TEST_JWT_SECRET must be set")

	s.accessDuration, err = time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	require.NoError(s.T(), err)

	s.refreshDuration, err = time.ParseDuration(os.Getenv("REFRESH_TOKEN_DURATION"))
	require.NoError(s.T(), err)

	s.jwtManager = jwt.NewManager(jwtSecret, s.accessDuration, s.refreshDuration)

	s.testUser = &models.User{
		ID:    uuid.New(),
		Login: "testuser",
	}
}

func (s *TokenValidatorTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserService = mocks.NewMockIUserService(s.ctrl)
	s.validator = NewTokenValidator(s.jwtManager, s.mockUserService)

	// Generate a valid refresh token
	token, err := s.jwtManager.GenerateToken(s.testUser, constants.TokenTypeRefresh)
	require.NoError(s.T(), err)
	s.validToken = token.Value

	// Generate an expired token (manually set expiration in the past)
	// For testing, we'll use a very short duration
	expiredManager := jwt.NewManager(os.Getenv("TEST_JWT_SECRET"), -1*time.Hour, -1*time.Hour)
	expiredTokenModel, err := expiredManager.GenerateToken(s.testUser, constants.TokenTypeRefresh)
	require.NoError(s.T(), err)
	s.expiredToken = expiredTokenModel.Value
}

func (s *TokenValidatorTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test ValidateRefreshToken - Success
func (s *TokenValidatorTestSuite) TestValidateRefreshTokenSuccess() {
	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(s.testUser, nil)

	parsedToken, err := s.validator.ValidateRefreshToken(s.validToken)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
	assert.Equal(s.T(), s.testUser.ID, parsedToken.UserID)
	assert.Equal(s.T(), string(constants.TokenTypeRefresh), parsedToken.Type)
}

// Test ValidateRefreshToken - Invalid Token
func (s *TokenValidatorTestSuite) TestValidateRefreshTokenInvalid() {
	invalidToken := "invalid.jwt.token"

	parsedToken, err := s.validator.ValidateRefreshToken(invalidToken)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), parsedToken)
}

// Test ValidateRefreshToken - Expired Token
func (s *TokenValidatorTestSuite) TestValidateRefreshTokenExpired() {
	parsedToken, err := s.validator.ValidateRefreshToken(s.expiredToken)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), parsedToken)
	// JWT library returns "token is expired" message
	assert.ErrorContains(s.T(), err, "token is expired")
}

// Test ValidateRefreshToken - Wrong Token Type (access instead of refresh)
func (s *TokenValidatorTestSuite) TestValidateRefreshTokenWrongType() {
	// Generate access token instead of refresh
	accessToken, err := s.jwtManager.GenerateToken(s.testUser, constants.TokenTypeAccess)
	require.NoError(s.T(), err)

	parsedToken, err := s.validator.ValidateRefreshToken(accessToken.Value)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), parsedToken)
	assert.ErrorContains(s.T(), err, "wrong token type")
}

// Test ValidateRefreshToken - User Not Found
func (s *TokenValidatorTestSuite) TestValidateRefreshTokenUserNotFound() {
	userNotFoundErr := errors.New("user not found")

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, userNotFoundErr)

	parsedToken, err := s.validator.ValidateRefreshToken(s.validToken)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), parsedToken)
	assert.ErrorContains(s.T(), err, "user not found")
}

// Test ValidateAccessToken - Success
func (s *TokenValidatorTestSuite) TestValidateAccessTokenSuccess() {
	accessToken, err := s.jwtManager.GenerateToken(s.testUser, constants.TokenTypeAccess)
	require.NoError(s.T(), err)

	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(s.testUser, nil)

	parsedToken, err := s.validator.ValidateAccessToken(accessToken.Value)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
	assert.Equal(s.T(), s.testUser.ID, parsedToken.UserID)
	assert.Equal(s.T(), string(constants.TokenTypeAccess), parsedToken.Type)
}

// Test Validate - Only Expiration Check (no type, no user)
func (s *TokenValidatorTestSuite) TestValidateOnlyExpirationCheck() {
	// No options = only parse + expiration check
	parsedToken, err := s.validator.Validate(s.validToken)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
	assert.Equal(s.T(), s.testUser.ID, parsedToken.UserID)
	// No user service call expected because WithUserExistenceCheck not used
}

// Test Validate - With Type Check Only
func (s *TokenValidatorTestSuite) TestValidateWithTypeCheckOnly() {
	parsedToken, err := s.validator.Validate(
		s.validToken,
		WithTokenType(constants.TokenTypeRefresh),
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
	assert.Equal(s.T(), string(constants.TokenTypeRefresh), parsedToken.Type)
}

// Test Validate - With User Check Only
func (s *TokenValidatorTestSuite) TestValidateWithUserCheckOnly() {
	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(s.testUser, nil)

	parsedToken, err := s.validator.Validate(
		s.validToken,
		WithUserExistenceCheck(),
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
}

// Test Validate - Multiple Options Combined
func (s *TokenValidatorTestSuite) TestValidateMultipleOptions() {
	s.mockUserService.EXPECT().
		FindUser(gomock.Any()).
		Return(s.testUser, nil)

	parsedToken, err := s.validator.Validate(
		s.validToken,
		WithTokenType(constants.TokenTypeRefresh),
		WithUserExistenceCheck(),
	)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), parsedToken)
	assert.Equal(s.T(), s.testUser.ID, parsedToken.UserID)
	assert.Equal(s.T(), string(constants.TokenTypeRefresh), parsedToken.Type)
}

func TestTokenValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(TokenValidatorTestSuite))
}
