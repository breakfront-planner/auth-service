package services

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
)

// HashService provides cryptographic hashing functionality for tokens and passwords.
type HashService struct{}

// NewHashService creates a new hash service instance.
func NewHashService() *HashService {
	return &HashService{}
}

// HashToken creates a SHA-256 hash of the provided token string.
func (s *HashService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashPassword generates a bcrypt hash of the provided password.
func (s *HashService) HashPassword(password string) (string, error) {

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", autherrors.ErrPassHash(err)
	}

	return string(passHash), nil

}

// ComparePasswords verifies that the input password matches the stored hash.
func (s *HashService) ComparePasswords(passHash, input string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(input))
	if err != nil {
		return autherrors.ErrWrongPassword(err)
	}

	return nil

}
