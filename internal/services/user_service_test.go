package services

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/breakfront-planner/auth-service/internal/services/mocks"
)

type UserServiceTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	mockUserRepo *mocks.MockIUserRepository
	hashService  *HashService
	userService  *UserService
	testLogin    string
	testPassword string
}

func (s *UserServiceTestSuite) SetupSuite() {
	err := godotenv.Load("../../.env.test")
	require.NoError(s.T(), err, "Failed to load .env.test")

	s.testLogin = os.Getenv("TEST_LOGIN")
	s.testPassword = os.Getenv("TEST_PASS")

	require.NotEmpty(s.T(), s.testLogin, "TEST_LOGIN must be set in .env.test")
	require.NotEmpty(s.T(), s.testPassword, "TEST_PASS must be set in .env.test")

	s.hashService = NewHashService()
}

func (s *UserServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockUserRepo = mocks.NewMockIUserRepository(s.ctrl)
	s.userService = NewUserService(s.mockUserRepo, s.hashService)
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *UserServiceTestSuite) TestCreateUserSuccess() {
	hashedPassword, err := s.hashService.HashPassword(s.testPassword)
	require.NoError(s.T(), err)

	expectedUser := &models.User{
		ID:           uuid.New(),
		Login:        s.testLogin,
		PasswordHash: hashedPassword,
	}

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, nil)

	s.mockUserRepo.EXPECT().
		CreateUser(s.testLogin, gomock.Any()).
		DoAndReturn(func(login string, passHash string) (*models.User, error) {
			return &models.User{
				ID:           expectedUser.ID,
				Login:        login,
				PasswordHash: passHash,
			}, nil
		})

	user, err := s.userService.CreateUser(s.testLogin, s.testPassword)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), s.testLogin, user.Login)
	assert.NotEmpty(s.T(), user.PasswordHash)
}

func (s *UserServiceTestSuite) TestCreateUserLoginTaken() {
	existingUser := &models.User{
		ID:    uuid.New(),
		Login: s.testLogin,
	}

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(existingUser, nil)

	user, err := s.userService.CreateUser(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.Equal(s.T(), autherrors.ErrLoginTaken, err)
}

func (s *UserServiceTestSuite) TestCreateUserRepositoryError() {
	repoError := errors.New("database error")

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, nil)

	s.mockUserRepo.EXPECT().
		CreateUser(s.testLogin, gomock.Any()).
		Return(nil, repoError)

	user, err := s.userService.CreateUser(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.ErrorContains(s.T(), err, "registration failed")
}

func (s *UserServiceTestSuite) TestCreateUserFindUserError() {
	findError := errors.New("find user error")

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, findError)

	user, err := s.userService.CreateUser(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.ErrorContains(s.T(), err, "registration failed")
}

func (s *UserServiceTestSuite) TestFindUserSuccess() {
	filter := &models.UserFilter{
		Login: &s.testLogin,
	}

	expectedUser := &models.User{
		ID:    uuid.New(),
		Login: s.testLogin,
	}

	s.mockUserRepo.EXPECT().
		FindUser(filter).
		Return(expectedUser, nil)

	user, err := s.userService.FindUser(filter)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedUser, user)
}

func (s *UserServiceTestSuite) TestFindUserNotFound() {
	nonexistent := "nonexistent"
	filter := &models.UserFilter{
		Login: &nonexistent,
	}

	s.mockUserRepo.EXPECT().
		FindUser(filter).
		Return(nil, nil)

	user, err := s.userService.FindUser(filter)

	assert.NoError(s.T(), err)
	assert.Nil(s.T(), user)
}

func (s *UserServiceTestSuite) TestFindUserError() {
	filter := &models.UserFilter{
		Login: &s.testLogin,
	}
	repoError := errors.New("database error")

	s.mockUserRepo.EXPECT().
		FindUser(filter).
		Return(nil, repoError)

	user, err := s.userService.FindUser(filter)

	assert.Error(s.T(), err)
	assert.Nil(s.T(), user)
	assert.ErrorContains(s.T(), err, "storage error")
}

func (s *UserServiceTestSuite) TestCheckPasswordSuccess() {
	hashedPassword, err := s.hashService.HashPassword(s.testPassword)
	require.NoError(s.T(), err)

	user := &models.User{
		ID:           uuid.New(),
		Login:        s.testLogin,
		PasswordHash: hashedPassword,
	}

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(user, nil)

	err = s.userService.CheckPassword(s.testLogin, s.testPassword)

	assert.NoError(s.T(), err)
}

func (s *UserServiceTestSuite) TestCheckPasswordWrongLogin() {
	findError := errors.New("user not found")

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(nil, findError)

	err := s.userService.CheckPassword(s.testLogin, s.testPassword)

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "failed to find user")
}

func (s *UserServiceTestSuite) TestCheckPasswordWrongPassword() {
	hashedPassword, err := s.hashService.HashPassword(s.testPassword)
	require.NoError(s.T(), err)

	user := &models.User{
		ID:           uuid.New(),
		Login:        s.testLogin,
		PasswordHash: hashedPassword,
	}

	s.mockUserRepo.EXPECT().
		FindUser(gomock.Any()).
		Return(user, nil)

	err = s.userService.CheckPassword(s.testLogin, "wrongpassword")

	assert.Error(s.T(), err)
	assert.ErrorContains(s.T(), err, "wrong password")
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
