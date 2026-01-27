package autherrors

import (
	"errors"
	"fmt"
)

var (
	ErrNoPtrsFilterFields = errors.New("all filter fields must be pointers")
	ErrEmptyFilter        = errors.New("filter cannot be empty")
)

func ErrMissingEnvVars(varNames []string) error {
	return fmt.Errorf("missing required environment variables: %v", varNames)
}

func ErrFailToCreateUser(err error) error {
	return fmt.Errorf("failed to create user: %w", err)
}

func ErrFailToFindUser(err error) error {
	return fmt.Errorf("failed to find user: %w", err)
}

func ErrSaveToken(err error) error {
	return fmt.Errorf("failed to save token: %w", err)
}

func ErrDeleteToken(err error) error {
	return fmt.Errorf("failed to delete token: %w", err)
}

func ErrDBTransactionFailed(err error) error {
	return fmt.Errorf("transaction failed: %w", err)
}

func ErrInvalidToken(err error) error {
	return fmt.Errorf("invalid token: %w", err)
}

func ErrCheckToken(err error) error {
	return fmt.Errorf("failed to check token: %w", err)
}
