package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HashServiceTestSuite struct {
	suite.Suite
	hashService *HashService
}

func (s *HashServiceTestSuite) SetupSuite() {
	s.hashService = NewHashService()
}

func (s *HashServiceTestSuite) TestHashTokenSuccess() {
	token := "sample_token_12345"

	hashedToken := s.hashService.HashToken(token)

	assert.NotEmpty(s.T(), hashedToken)
	assert.Len(s.T(), hashedToken, 64) // SHA256 produces 64 hex characters
}

func (s *HashServiceTestSuite) TestHashTokenConsistent() {
	token := "consistent_token"

	hash1 := s.hashService.HashToken(token)
	hash2 := s.hashService.HashToken(token)

	assert.Equal(s.T(), hash1, hash2, "Same token should produce same hash")
}

func (s *HashServiceTestSuite) TestHashTokenDifferentInputs() {
	token1 := "token_one"
	token2 := "token_two"

	hash1 := s.hashService.HashToken(token1)
	hash2 := s.hashService.HashToken(token2)

	assert.NotEqual(s.T(), hash1, hash2, "Different tokens should produce different hashes")
}

func (s *HashServiceTestSuite) TestHashPasswordSuccess() {
	password := "securePassword123"

	hashedPassword, err := s.hashService.HashPassword(password)

	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), hashedPassword)
	assert.True(s.T(), strings.HasPrefix(hashedPassword, "$2a$"), "Bcrypt hash should start with $2a$")
}

func (s *HashServiceTestSuite) TestHashPasswordUnique() {
	password := "samePassword"

	hash1, err1 := s.hashService.HashPassword(password)
	require.NoError(s.T(), err1)

	hash2, err2 := s.hashService.HashPassword(password)
	require.NoError(s.T(), err2)

	assert.NotEqual(s.T(), hash1, hash2, "Bcrypt should produce different hashes for same password due to random salt")
}

func (s *HashServiceTestSuite) TestComparePasswordsSuccess() {
	password := "correctPassword"

	hashedPassword, err := s.hashService.HashPassword(password)
	require.NoError(s.T(), err)

	err = s.hashService.ComparePasswords(hashedPassword, password)

	assert.NoError(s.T(), err)
}

func (s *HashServiceTestSuite) TestComparePasswordsWrongPassword() {
	originalPassword := "correctPassword"
	wrongPassword := "wrongPassword"

	hashedPassword, err := s.hashService.HashPassword(originalPassword)
	require.NoError(s.T(), err)

	err = s.hashService.ComparePasswords(hashedPassword, wrongPassword)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "wrong password")
}

func (s *HashServiceTestSuite) TestComparePasswordsInvalidHash() {
	invalidHash := "not_a_valid_bcrypt_hash"
	password := "password123"

	err := s.hashService.ComparePasswords(invalidHash, password)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "wrong password")
}

func TestHashServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HashServiceTestSuite))
}
