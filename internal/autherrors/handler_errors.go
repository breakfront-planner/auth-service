package autherrors

import "fmt"

func ErrLoginLength(min, max int) error {
	return fmt.Errorf("login must be between %d and %d characters", min, max)
}

func ErrPasswordLength(min, max int) error {
	return fmt.Errorf("password must be between %d and %d characters", min, max)
}
