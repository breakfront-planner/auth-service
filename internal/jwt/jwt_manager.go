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

// ParseToken extract user info from token
func (m *Manager) ParseToken(tokenString string) (parsedToken *models.ParsedToken, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, autherrors.ErrTokenSignMethod
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, autherrors.ErrInvalidJWT
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, autherrors.ErrNoClaimInToken("user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, autherrors.ErrInvalidUserID
	}
	tokenType, ok := claims["type"].(string)
	if !ok {
		return nil, autherrors.ErrNoClaimInToken("type")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, autherrors.ErrNoClaimInToken("exp")
	}
	exp := time.Unix(int64(expFloat), 0)

	parsedToken = &models.ParsedToken{
		UserID:    userID,
		Type:      tokenType,
		ExpiresAt: exp,
	}

	return parsedToken, nil
}
