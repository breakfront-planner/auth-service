package services

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/breakfront-planner/auth-service/internal/configs"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/breakfront-planner/auth-service/internal/services/mocks"
)

type TokenServiceTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockTokenRepo   *mocks.MockITokenRepository
	mockHashService *mocks.MockIHashService
	jwtManager      *jwt.Manager
	tokenService    *TokenService
	testUser        *models.User
	accessDuration  time.Duration
	refreshDuration time.Duration
	testHashedValue string
	testTokenValue  string
}

func (s *TokenServiceTestSuite) SetupSuite() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Failed to load .env.test")

	tokenCfg := &configs.TokenConfig{}
	require.NoError(s.T(), env.ParseWithOptions(tokenCfg, env.Options{Prefix: "TEST_"}), "Failed to parse test token configuration")

	s.accessDuration = tokenCfg.AccessDuration
	s.refreshDuration = tokenCfg.RefreshDuration
	s.jwtManager = jwt.NewManager(tokenCfg.JWTSecret, s.accessDuration, s.refreshDuration)

	s.testHashedValue = os.Getenv("TOKEN_HASHED_VALUE")
	s.testTokenValue = os.Getenv("TOKEN_TEST_VALUE")
	require.NotEmpty(s.T(), s.testHashedValue, "TOKEN_HASHED_VALUE must be set in .env.test")
	require.NotEmpty(s.T(), s.testTokenValue, "TOKEN_TEST_VALUE must be set in .env.test")

	testLogin := os.Getenv("TEST_LOGIN")
	testPass := os.Getenv("TEST_PASS")
	require.NotEmpty(s.T(), testLogin, "TEST_LOGIN must be set in .env.test")
	require.NotEmpty(s.T(), testPass, "TEST_PASS must be set in .env.test")

	s.testUser = &models.User{
		ID:           uuid.New(),
		Login:        testLogin,
		PasswordHash: testPass,
	}
}

func (s *TokenServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockTokenRepo = mocks.NewMockITokenRepository(s.ctrl)
	s.mockHashService = mocks.NewMockIHashService(s.ctrl)
	s.tokenService = NewTokenService(s.mockTokenRepo, s.mockHashService, s.jwtManager)
}

func (s *TokenServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *TokenServiceTestSuite) TestCreateNewTokenPairSuccess() {
	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		SaveToken(gomock.Any()).
		Return(nil)

	accessToken, refreshToken, err := s.tokenService.CreateNewTokenPair(s.testUser)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessToken)
	assert.NotNil(s.T(), refreshToken)
	assert.NotEmpty(s.T(), accessToken.Value)
	assert.NotEmpty(s.T(), refreshToken.Value)
	assert.Equal(s.T(), s.testUser.ID, accessToken.UserID)
	assert.Equal(s.T(), s.testUser.ID, refreshToken.UserID)
	assert.Equal(s.T(), s.testHashedValue, refreshToken.HashedValue)
}

func (s *TokenServiceTestSuite) TestCreateNewTokenPairSaveTokenError() {
	saveError := errors.New("database error")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		SaveToken(gomock.Any()).
		Return(saveError)

	accessToken, refreshToken, err := s.tokenService.CreateNewTokenPair(s.testUser)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), accessToken)
	assert.Nil(s.T(), refreshToken)
	assert.ErrorContains(s.T(), err, "failed to save token")
}

func (s *TokenServiceTestSuite) TestRefreshSuccess() {
	oldRefreshToken := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		DoAndReturn(func(token *models.Token) error {
			assert.Equal(s.T(), s.testHashedValue, token.HashedValue)
			return nil
		})

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		SaveToken(gomock.Any()).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		RevokeToken(gomock.Any()).
		DoAndReturn(func(token *models.Token) error {
			assert.Equal(s.T(), s.testHashedValue, token.HashedValue)
			return nil
		})

	newAccessToken, newRefreshToken, err := s.tokenService.Refresh(oldRefreshToken, s.testUser)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), newAccessToken)
	assert.NotNil(s.T(), newRefreshToken)
	assert.NotEmpty(s.T(), newAccessToken.Value)
	assert.NotEmpty(s.T(), newRefreshToken.Value)
	assert.NotEqual(s.T(), oldRefreshToken.Value, newRefreshToken.Value)
}

func (s *TokenServiceTestSuite) TestRefreshInvalidToken() {
	oldRefreshToken := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	checkError := errors.New("token not found")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		Return(checkError)

	newAccessToken, newRefreshToken, err := s.tokenService.Refresh(oldRefreshToken, s.testUser)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), newAccessToken)
	assert.Nil(s.T(), newRefreshToken)
	assert.ErrorContains(s.T(), err, "failed to refresh token")
}

func (s *TokenServiceTestSuite) TestRefreshRevokeError() {
	oldRefreshToken := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	revokeError := errors.New("revoke failed")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		Return(nil)

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		SaveToken(gomock.Any()).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		RevokeToken(gomock.Any()).
		Return(revokeError)

	newAccessToken, newRefreshToken, err := s.tokenService.Refresh(oldRefreshToken, s.testUser)

	assert.Error(s.T(), err)
	assert.NotNil(s.T(), newAccessToken)
	assert.NotNil(s.T(), newRefreshToken)
	assert.ErrorContains(s.T(), err, "failed to revoke token")
}

func (s *TokenServiceTestSuite) TestRefreshCreateTokenPairError() {
	oldRefreshToken := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	saveError := errors.New("save failed")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		Return(nil)

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue).
		Times(1)

	s.mockTokenRepo.EXPECT().
		SaveToken(gomock.Any()).
		Return(saveError)

	newAccessToken, newRefreshToken, err := s.tokenService.Refresh(oldRefreshToken, s.testUser)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), newAccessToken)
	assert.Nil(s.T(), newRefreshToken)
	assert.ErrorContains(s.T(), err, "failed to create token")
}

func (s *TokenServiceTestSuite) TestRevokeTokenSuccess() {
	token := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		DoAndReturn(func(token *models.Token) error {
			assert.Equal(s.T(), s.testHashedValue, token.HashedValue)
			return nil
		})

	s.mockTokenRepo.EXPECT().
		RevokeToken(gomock.Any()).
		DoAndReturn(func(token *models.Token) error {
			assert.Equal(s.T(), s.testHashedValue, token.HashedValue)
			return nil
		})

	err := s.tokenService.RevokeToken(token)

	assert.NoError(s.T(), err)
}

func (s *TokenServiceTestSuite) TestRevokeTokenCheckError() {
	token := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	checkError := errors.New("token not found")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		Return(checkError)

	err := s.tokenService.RevokeToken(token)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "token not found")
}

func (s *TokenServiceTestSuite) TestRevokeTokenRevokeError() {
	token := &models.Token{
		Value:     s.testTokenValue,
		UserID:    s.testUser.ID,
		ExpiresAt: time.Now().UTC().Add(s.refreshDuration),
	}

	revokeError := errors.New("revoke failed")

	s.mockHashService.EXPECT().
		HashToken(gomock.Any()).
		Return(s.testHashedValue)

	s.mockTokenRepo.EXPECT().
		FindToken(gomock.Any()).
		Return(nil)

	s.mockTokenRepo.EXPECT().
		RevokeToken(gomock.Any()).
		Return(revokeError)

	err := s.tokenService.RevokeToken(token)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to revoke token")
}

func TestTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceTestSuite))
}
