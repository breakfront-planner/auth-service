package validators

import (
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/constants"
	"github.com/breakfront-planner/auth-service/internal/jwt"
	"github.com/breakfront-planner/auth-service/internal/models"
)

// IUserService defines operations for user management.
type IUserService interface {
	FindUser(filter *models.UserFilter) (*models.User, error)
}

// ValidationConfig holds the settings for token validation checks.
type ValidationConfig struct {
	RequiredType    *constants.TokenType
	CheckUserExists bool
}

// ValidationOption is a function that modifies ValidationConfig.
type ValidationOption func(*ValidationConfig)

// WithTokenType adds token type validation.
func WithTokenType(tokenType constants.TokenType) ValidationOption {
	return func(config *ValidationConfig) {
		config.RequiredType = &tokenType
	}
}

// WithUserExistenceCheck enables user existence validation.
func WithUserExistenceCheck() ValidationOption {
	return func(config *ValidationConfig) {
		config.CheckUserExists = true
	}
}

// TokenValidator validates JWT tokens with flexible configuration.
type TokenValidator struct {
	jwtManager  *jwt.Manager
	userService IUserService
}

// NewTokenValidator creates a new token validator instance.
func NewTokenValidator(jwtManager *jwt.Manager, userService IUserService) *TokenValidator {
	return &TokenValidator{
		jwtManager:  jwtManager,
		userService: userService,
	}
}

// Validate performs token validation with the given options.
func (v *TokenValidator) Validate(tokenValue string, opts ...ValidationOption) (*models.ParsedToken, error) {
	// Apply configuration options
	config := &ValidationConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Parse token and verify signature
	parsedToken, err := v.jwtManager.ParseToken(tokenValue)
	if err != nil {
		return nil, autherrors.ErrParseToken(err)
	}

	// Always check expiration
	if parsedToken.ExpiresAt.Before(time.Now().UTC()) {
		return nil, autherrors.ErrTokenExpired
	}

	// Validate token type if specified
	if config.RequiredType != nil {
		if parsedToken.Type != string(*config.RequiredType) {
			return nil, autherrors.ErrTokenType
		}
	}

	// Validate user existence if specified
	if config.CheckUserExists {
		filter := &models.UserFilter{
			ID: &parsedToken.UserID,
		}
		if _, err := v.userService.FindUser(filter); err != nil {
			return nil, err
		}
	}

	return parsedToken, nil
}

// ValidateRefreshToken validates a refresh token with all checks enabled.
func (v *TokenValidator) ValidateRefreshToken(tokenValue string) (*models.ParsedToken, error) {
	refreshType := constants.TokenTypeRefresh
	return v.Validate(
		tokenValue,
		WithTokenType(refreshType),
		WithUserExistenceCheck(),
	)
}

// ValidateAccessToken validates an access token with all checks enabled.
func (v *TokenValidator) ValidateAccessToken(tokenValue string) (*models.ParsedToken, error) {
	accessType := constants.TokenTypeAccess
	return v.Validate(
		tokenValue,
		WithTokenType(accessType),
		WithUserExistenceCheck(),
	)
}
