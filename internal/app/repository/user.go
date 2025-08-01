package repository

import (
	"context"
	"database/sql"
	"fmt"

	models "user-svc/internal/domain"
	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// User domain model
type User struct {
	ID           string `db:"id"`
	Email        string `db:"email"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	CreatedAt    int64  `db:"created_at"`
	UpdatedAt    int64  `db:"updated_at"`
}

func (u *User) ToDomain() *models.User {
	email, err := models.NewEmail(u.Email)
	if err != nil {
		// This should not happen in normal operation since we store validated emails
		// But we need to handle it for backward compatibility
		email = models.Email(u.Email)
	}

	username, err := models.NewUsername(u.Username)
	if err != nil {
		// This should not happen in normal operation since we store validated usernames
		// But we need to handle it for backward compatibility
		username = models.Username(u.Username)
	}

	id, err := uuid.Parse(u.ID)
	if err != nil {
		id = uuid.Nil
	}

	return &models.User{
		ID:           id,
		Email:        email,
		Username:     username,
		PasswordHash: models.PasswordHash(u.PasswordHash),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, created_at, updated_at)
		VALUES (:id, :email, :username, :password_hash, :created_at, :updated_at)
	`

	// Convert domain user to repository user
	repoUser := &User{
		ID:           user.ID.String(),
		Email:        user.Email.String(),
		Username:     user.Username.String(),
		PasswordHash: user.PasswordHash.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	_, err := r.db.NamedExecContext(ctx, query, repoUser)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrUserNotFound
	}

	return nil
}
