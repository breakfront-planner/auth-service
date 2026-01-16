package autherrors

import (
	"errors"
	"fmt"

	"github.com/breakfront-planner/auth-service/internal/models"
)

var (
	ErrLoginTaken     = errors.New("login already taken")
	ErrWrongTokenType = errors.New("wrong tokenType, should be 'access' or 'refresh'")
)

func ErrPassHash(err error) error {
	return fmt.Errorf("hashing password failed: %w", err)
}

func ErrRegisterFailed(err error) error {
	return fmt.Errorf("registration failed: %w", err)
}

func ErrFindUser(err error) error {
	return fmt.Errorf("storage error: %w", err)
}

func ErrUserNotFound(user *models.User) error {
	return fmt.Errorf("user with ID: %v doesn't found", user.ID)
}

func ErrCreateToken(err error) error {
	return fmt.Errorf("failed to create token: %w", err)
}

func ErrWrongLogin(err error) error {
	return fmt.Errorf("failed to find user: %w", err)
}

func ErrWrongPassword(err error) error {
	return fmt.Errorf("wrong password: %w", err)
}

func ErrRefreshToken(err error) error {
	return fmt.Errorf("failed to refresh tokens: %w", err)
}

func ErrRevokeToken(err error) error {
	return fmt.Errorf("failed to revoke token: %w", err)
}
