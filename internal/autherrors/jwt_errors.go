package autherrors

import (
	"errors"
	"fmt"
	//"github.com/breakfront-planner/auth-service/internal/models"
)

var (
	ErrWrongTokenType  = errors.New("wrong tokenType, should be 'access' or 'refresh'")
	ErrTokenSignMethod = errors.New("unexpected signing method")
	ErrInvalidJWT      = errors.New("invalid JWT")
	ErrInvalidUserID   = errors.New("invalid user_id format")
)

func ErrNoClaimInToken(claim string) error {
	return fmt.Errorf("not found in token: %v", claim)
}
