package repositories

import (
	//"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	//"github.com/breakfront-planner/auth-service/internal/autherrors"

	"os"

	"github.com/joho/godotenv"
)

// UserRepositoryTestSuite extends RepositoryTestSuite
type UserRepositoryTestSuite struct {
	RepositoryTestSuite
}

// TestCreateSuccess tests successful user creation
func (s *UserRepositoryTestSuite) TestCreateSuccess() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Error loading .env file")

	// ACT
	user, err := s.UserRepo.CreateUser(os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASS"))

	// ASSERT
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), user.ID, "User ID should be generated")
	assert.Equal(s.T(), "testuser", user.Login)
}

//func (s *UserRepositoryTestSuite) TestCreateLoginTaken()

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
