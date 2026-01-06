package repositories

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/google/uuid"

	"github.com/breakfront-planner/auth-service/internal/models"
)

type TokenRepositoryTestSuite struct {
	RepositoryTestSuite
	TestUser *models.User
}

func (s *TokenRepositoryTestSuite) SetupSuite() {
	s.RepositoryTestSuite.SetupSuite()

	user, err := s.UserRepo.CreateUser(s.TestLogin, s.TestPassword)
	require.NoError(s.T(), err)
	s.TestUser = user
}

func (s *TokenRepositoryTestSuite) TestSaveAndCheckSuccess() {
	token := models.Token{
		HashedValue: s.TokenHashedValue,
		UserID:      s.TestUser.ID,
		ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
	}

	err := s.TokenRepo.SaveToken(&token)
	require.NoError(s.T(), err)

	err = s.TokenRepo.CheckToken(&token)
	require.NoError(s.T(), err)
}

func (s *TokenRepositoryTestSuite) TestSaveError() {
	token := models.Token{
		HashedValue: s.TokenHashedValue,
		ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
	}

	err := s.TokenRepo.SaveToken(&token)

	require.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to save token")
}

func (s *TokenRepositoryTestSuite) TestCheckTokenErrors() {
	validToken := models.Token{
		HashedValue: s.TokenHashedValue,
		UserID:      s.TestUser.ID,
		ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
	}

	err := s.TokenRepo.SaveToken(&validToken)
	require.NoError(s.T(), err)

	expiredToken := models.Token{
		HashedValue: s.TokenHashedValue[:30] + "expired",
		UserID:      s.TestUser.ID,
		ExpiresAt:   time.Now().UTC().Add(-s.RefreshDuration),
	}

	err = s.TokenRepo.SaveToken(&expiredToken)
	require.NoError(s.T(), err)

	revokedToken := models.Token{
		HashedValue: s.TokenHashedValue[:30] + "revoked",
		UserID:      s.TestUser.ID,
		ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
	}

	err = s.TokenRepo.SaveToken(&revokedToken)
	require.NoError(s.T(), err)

	err = s.TokenRepo.RevokeToken(&revokedToken)
	require.NoError(s.T(), err)

	testCases := []struct {
		name          string
		token         models.Token
		errorContains string
	}{
		{
			name: "non-existent token",
			token: models.Token{
				HashedValue: s.TokenHashedValue[:30] + "nonexistent",
				UserID:      s.TestUser.ID,
				ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
			},
			errorContains: "invalid token",
		},
		{
			name: "wrong user ID",
			token: models.Token{
				HashedValue: s.TokenHashedValue,
				UserID:      uuid.New(),
				ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
			},
			errorContains: "invalid token",
		},
		{
			name:          "expired token",
			token:         expiredToken,
			errorContains: "token expired",
		},
		{
			name:          "revoked token",
			token:         revokedToken,
			errorContains: "invalid token",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.TokenRepo.CheckToken(&tc.token)

			require.Error(s.T(), err)
			assert.ErrorContains(s.T(), err, tc.errorContains)
		})
	}
}

func (s *TokenRepositoryTestSuite) TestRevokeTokenSuccess() {

	token := models.Token{
		HashedValue: s.TokenHashedValue,
		UserID:      s.TestUser.ID,
		ExpiresAt:   time.Now().UTC().Add(s.RefreshDuration),
	}

	err := s.TokenRepo.SaveToken(&token)
	require.NoError(s.T(), err)

	err = s.TokenRepo.RevokeToken(&token)
	require.NoError(s.T(), err)

	err = s.TokenRepo.CheckToken(&token)
	require.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "invalid token")
}

func (s *TokenRepositoryTestSuite) TestRevokeNonExistentToken() {
	token := models.Token{
		HashedValue: "nonexistent_hash",
		UserID:      s.TestUser.ID,
	}

	err := s.TokenRepo.RevokeToken(&token)
	require.NoError(s.T(), err)
}

func (s *TokenRepositoryTestSuite) TearDownTest() {
	_, err := s.DB.Exec("DELETE FROM refresh_tokens")
	require.NoError(s.T(), err, "Failed to cleanup refresh_tokens")
}

func (s *TokenRepositoryTestSuite) TearDownSuite() {
	_, err := s.DB.Exec("DELETE FROM users")
	require.NoError(s.T(), err, "Failed to cleanup users")

	s.RepositoryTestSuite.TearDownSuite()
}

func TestTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TokenRepositoryTestSuite))
}
