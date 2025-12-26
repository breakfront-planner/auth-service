package services

import (
	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/models"
)

type IUserRepository interface {
	CreateUser(login string, passHash string) (*models.User, error)
	FindUser(filter *models.UserFilter) (*models.User, error)
}

type UserService struct {
	userRepo    IUserRepository
	hashService IHashService
}

func NewUserService(userRepo IUserRepository, hashService IHashService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		hashService: hashService,
	}
}

func (s *UserService) CreateUser(login string, password string) (*models.User, error) {
	newUserFilter := models.UserFilter{
		Login: &login,
	}

	user, err := s.userRepo.FindUser(&newUserFilter)

	if err != nil {
		return nil, autherrors.ErrRegisterFailed(err)
	}

	if user == nil {
		passHash, err := s.hashService.HashPassword(password)
		if err != nil {
			return nil, err
		}

		user, err = s.userRepo.CreateUser(login, string(passHash))
		if err != nil {
			return nil, autherrors.ErrRegisterFailed(err)
		}
		return user, nil
	}

	return nil, autherrors.ErrLoginTaken

}

func (s *UserService) FindUser(filter *models.UserFilter) (*models.User, error) {

	user, err := s.userRepo.FindUser(filter)

	if err != nil {
		return nil, autherrors.ErrFindUser(err)
	}

	if user != nil {
		return user, nil
	}

	return nil, nil

}

func (s *UserService) CheckPassword(login string, password string) error {

	filter := models.UserFilter{
		Login: &login,
	}

	user, err := s.userRepo.FindUser(&filter)
	if err != nil {
		return autherrors.ErrWrongLogin(err)
	}

	err = s.hashService.ComparePasswords(user.PasswordHash, password)
	if err != nil {
		return err
	}

	return nil

}
