package models

import (
	"time"

	"github.com/google/uuid"
)

// Token represents a JWT token with its metadata.
type Token struct {
	Value       string
	HashedValue string
	UserID      uuid.UUID
	ExpiresAt   time.Time
	RevokedAt   *time.Time
}

// Token represents parsed JWT token claims.
type ParsedToken struct {
	UserID    uuid.UUID
	Type      string
	ExpiresAt time.Time
}
