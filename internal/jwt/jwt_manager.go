package jwt

import (
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Manager manages JWT token generation and validation.
// It handles both access and refresh tokens with configurable expiration durations.
type Manager struct {
	secret          string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewManager creates a new JWT manager instance.
// The secret is used for signing tokens, while accessDuration and refreshDuration
// define the expiration time for access and refresh tokens respectively.
func NewManager(secret string, accessDuration, refreshDuration time.Duration) *Manager {
	return &Manager{
		secret:          secret,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateToken creates a new JWT token for the specified user.
// The tokenType parameter determines whether to generate an access or refresh token,
// which affects the token's expiration duration and claims.
func (m *Manager) GenerateToken(user *models.User, tokenType constants.TokenType) (*models.Token, error) {
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
