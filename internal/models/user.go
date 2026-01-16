package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system.
type User struct {
	ID           uuid.UUID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserFilter provides criteria for searching users.
type UserFilter struct {
	ID    *uuid.UUID `db:"id"`
	Login *string    `db:"login"`
}
