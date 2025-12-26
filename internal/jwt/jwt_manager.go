package jwt

import (
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secret          string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewJWTManager(secret string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secret:          secret,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (m *JWTManager) GenerateToken(user *models.User, tokenType constants.TokenType) (*models.Token, error) {
	var duration time.Duration

	switch tokenType {
	case constants.TokenTypeAccess:
		duration = m.accessDuration
	case constants.TokenTypeRefresh:
		duration = m.refreshDuration
	default:
		return nil, autherrors.ErrWrongTokenType
	}

	expiresAt := time.Now().UTC().Add(duration)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     expiresAt.Unix(),
		"type":    tokenType,
		"jti":     uuid.New().String(),
	}
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	value, err := unsignedToken.SignedString([]byte(m.secret))
	if err != nil {
		return nil, err
	}
	token := models.Token{
		Value:     value,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	return &token, nil
}
