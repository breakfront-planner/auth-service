package models

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Value       string
	HashedValue string
	UserID      uuid.UUID
	ExpiresAt   time.Time
	RevokedAt   *time.Time
}
