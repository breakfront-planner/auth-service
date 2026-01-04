package repositories

import (
	"database/sql"
	"fmt"

	"github.com/breakfront-planner/auth-service/internal/autherrors"
	"github.com/breakfront-planner/auth-service/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

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

func (r *UserRepository) FindUser(filter *models.UserFilter) (*models.User, error) {
	if filter.ID == nil && filter.Login == nil {
		return nil, autherrors.ErrFailToFindUser(fmt.Errorf("filter cannot be empty: at least one field must be specified"))
	}

	query := `SELECT id, login, password_hash, created_at, updated_at FROM users WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argIndex)
		args = append(args, *filter.ID)
		argIndex++
	}

	if filter.Login != nil {
		query += fmt.Sprintf(" AND login = $%d", argIndex)
		args = append(args, *filter.Login)
		argIndex++
	}

	var user models.User
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, autherrors.ErrFailToFindUser(err)
	}
	return &user, nil
}
