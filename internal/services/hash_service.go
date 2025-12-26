package services

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
)

type HashService struct{}

func NewHashService() *HashService {
	return &HashService{}
}

func (s *HashService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *HashService) HashPassword(password string) (string, error) {

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", autherrors.ErrPassHash(err)
	}

	return string(passHash), nil

}

func (s *HashService) ComparePasswords(passHash, input string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(input))
	if err != nil {
		return autherrors.ErrWrongPassword(err)
	}

	return nil

}
