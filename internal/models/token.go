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
