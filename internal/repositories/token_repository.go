package repositories

import (
	"database/sql"
	"log"
	"time"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/models"
)

// TokenRepository handles refresh token data persistence operations.
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new token repository instance.
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// SaveToken persists a refresh token to the database.
func (r *TokenRepository) SaveToken(token *models.Token) error {

	_, err := r.db.Exec(`INSERT INTO refresh_tokens (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		token.HashedValue, token.UserID, token.ExpiresAt)
	if err != nil {
		return autherrors.ErrSaveToken(err)
	}

	return nil

}

// RevokeToken marks a refresh token as revoked by setting its revoked_at timestamp.
func (r *TokenRepository) RevokeToken(token *models.Token) error {

	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = $1`, token.HashedValue)
	if err != nil {
		return autherrors.ErrDeleteToken(err)
	}

	return nil

}

// CheckToken validates a refresh token by verifying it exists, is not revoked, and has not expired.
func (r *TokenRepository) CheckToken(token *models.Token) error {

	var dbToken models.Token

	query := `SELECT user_id, expires_at, revoked_at 
	FROM refresh_tokens 
	WHERE token_hash = $1`

	err := r.db.QueryRow(query, token.HashedValue).Scan(
		&dbToken.UserID, &dbToken.ExpiresAt, &dbToken.RevokedAt)

	if err == sql.ErrNoRows || dbToken.UserID != token.UserID || dbToken.RevokedAt != nil {
		log.Printf("invalid token: %v ", err)
		return autherrors.ErrInvalidToken(err)
	}

	if err != nil {
		return autherrors.ErrCheckToken(err)
	}

	if dbToken.ExpiresAt.Before(time.Now().UTC()) {

		return autherrors.ErrExpiredToken(err)

	}

	return nil

}
