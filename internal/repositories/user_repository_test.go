package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/breakfront-planner/auth-service/internal/models"
)

type UserRepositoryTestSuite struct {
	RepositoryTestSuite
}

func (s *UserRepositoryTestSuite) TestCreateSuccess() {

	user, err := s.UserRepo.CreateUser(s.TestLogin, s.TestPassword)

	require.NoError(s.T(), err)
	assert.NotZero(s.T(), user.ID, "User ID should be generated")
	assert.Equal(s.T(), s.TestLogin, user.Login)
}

func (s *UserRepositoryTestSuite) TestCreateError() {

	_, err := s.UserRepo.CreateUser(s.TestLogin, s.TestPassword)
	require.NoError(s.T(), err)

	user, err := s.UserRepo.CreateUser(s.TestLogin, s.TestPassword)

	require.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to create user", "Should return ErrFailToCreateUser if error")
	assert.Nil(s.T(), user, "User should be nil when error")

}

func (s *UserRepositoryTestSuite) TestFindSuccess() {
	createdUser, err := s.UserRepo.CreateUser(s.TestLogin, s.TestPassword)
	require.NoError(s.T(), err)

	nonExistentID := uuid.New()
	nonExistentLogin := "nonexistent_user"

	testCases := []struct {
		name        string
		filter      models.UserFilter
		expectFound bool
	}{
		{
			name:        "find by login",
			filter:      models.UserFilter{Login: &s.TestLogin},
			expectFound: true,
		},
		{
			name:        "find by ID",
			filter:      models.UserFilter{ID: &createdUser.ID},
			expectFound: true,
		},
		{
			name:        "find by both filters",
			filter:      models.UserFilter{ID: &createdUser.ID, Login: &s.TestLogin},
			expectFound: true,
		},
		{
			name:        "find by non-existent login",
			filter:      models.UserFilter{Login: &nonExistentLogin},
			expectFound: false,
		},
		{
			name:        "find by non-existent ID",
			filter:      models.UserFilter{ID: &nonExistentID},
			expectFound: false,
		},
		{
			name:        "find by existing ID and non-existent login",
			filter:      models.UserFilter{ID: &createdUser.ID, Login: &nonExistentLogin},
			expectFound: false,
		},
		{
			name:        "find by non-existent ID and existing login",
			filter:      models.UserFilter{ID: &nonExistentID, Login: &s.TestLogin},
			expectFound: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			user, err := s.UserRepo.FindUser(&tc.filter)

			require.NoError(s.T(), err)

			if tc.expectFound {
				assert.NotNil(s.T(), user)
				assert.Equal(s.T(), createdUser.ID, user.ID)
				assert.Equal(s.T(), s.TestLogin, user.Login)
				assert.NotZero(s.T(), user.PasswordHash)
				assert.NotZero(s.T(), user.CreatedAt)
				assert.NotZero(s.T(), user.UpdatedAt)
			} else {
				assert.Nil(s.T(), user)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestFindWithEmptyFilter() {

	emptyFilter := models.UserFilter{}

	user, err := s.UserRepo.FindUser(&emptyFilter)

	require.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to find user")
	assert.ErrorContains(s.T(), err, "filter cannot be empty")
	assert.Nil(s.T(), user, "User should be nil when filter is empty")
}

func (s *UserRepositoryTestSuite) TearDownTest() {

	_, err := s.DB.Exec("DELETE FROM users")
	require.NoError(s.T(), err, "Failed to cleanup users")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
