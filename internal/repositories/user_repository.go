package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/models"
)

// UserRepository handles user data persistence operations.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user record into the database and returns the created user.
func (r *UserRepository) CreateUser(login string, passHash string) (*models.User, error) {
	var user models.User
	query := `
        INSERT INTO users (login, password_hash)
        VALUES ($1, $2)
        RETURNING id, login, password_hash, created_at, updated_at
    `

	err := r.db.QueryRow(query, login, passHash).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, autherrors.ErrFailToCreateUser(err)
	}

	return &user, nil
}

// FindUser searches for a user in the database using the provided filter criteria.
// Returns nil if no matching user is found.
func (r *UserRepository) FindUser(filter *models.UserFilter) (*models.User, error) {
	fields, err := ParseFilter(filter)

	if err != nil {
		return nil, autherrors.ErrFailToFindUser(err)
	}

	var conditions []string
	var args []interface{}
	for i, value := range fields {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", value.DBName, i+1))
		args = append(args, value.Value)
	}
	query := `SELECT id, login, password_hash, created_at, updated_at FROM users WHERE ` + strings.Join(conditions, " AND ")

	var user models.User
	err = r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, autherrors.ErrFailToFindUser(err)
	}
	return &user, nil
}
