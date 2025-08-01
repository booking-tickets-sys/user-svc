package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	models "user-svc/internal/domain"
	"user-svc/internal/domain/errs"
	"user-svc/utils/tx"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RefreshToken domain model
type RefreshToken struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	TokenHash string `db:"token_hash"`
	ExpiresAt int64  `db:"expires_at"`
	IsRevoked bool   `db:"is_revoked"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func (rt *RefreshToken) ToDomain() *models.RefreshToken {
	userID, err := uuid.Parse(rt.UserID)
	if err != nil {
		userID = uuid.Nil
	}

	return &models.RefreshToken{
		ID:        rt.ID,
		UserID:    userID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		IsRevoked: rt.IsRevoked,
		CreatedAt: rt.CreatedAt,
		UpdatedAt: rt.UpdatedAt,
	}
}

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, refreshToken *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, is_revoked, created_at, updated_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :is_revoked, :created_at, :updated_at)
	`

	repoRefreshToken := &RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID.String(),
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		IsRevoked: refreshToken.IsRevoked,
		CreatedAt: refreshToken.CreatedAt,
		UpdatedAt: refreshToken.UpdatedAt,
	}

	_, err := r.db.NamedExecContext(ctx, query, repoRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByID retrieves a refresh token by ID
func (r *RefreshTokenRepository) GetByID(ctx context.Context, id string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, is_revoked, created_at, updated_at
		FROM refresh_tokens 
		WHERE id = $1
	`

	var refreshToken RefreshToken
	err := r.db.GetContext(ctx, &refreshToken, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token by ID: %w", err)
	}

	return refreshToken.ToDomain(), nil
}

// GetByTokenHash retrieves a refresh token by token hash
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, is_revoked, created_at, updated_at
		FROM refresh_tokens 
		WHERE token_hash = $1
	`

	var refreshToken RefreshToken
	err := r.db.GetContext(ctx, &refreshToken, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token by hash: %w", err)
	}

	return refreshToken.ToDomain(), nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *RefreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, is_revoked, created_at, updated_at
		FROM refresh_tokens 
		WHERE user_id = $1
	`

	var refreshTokens []RefreshToken
	err := r.db.SelectContext(ctx, &refreshTokens, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens by user ID: %w", err)
	}

	domainRefreshTokens := make([]*models.RefreshToken, len(refreshTokens))
	for i, rt := range refreshTokens {
		domainRefreshTokens[i] = rt.ToDomain()
	}

	return domainRefreshTokens, nil
}

// Update updates a refresh token
func (r *RefreshTokenRepository) Update(ctx context.Context, refreshToken *models.RefreshToken) error {
	query := `
		UPDATE refresh_tokens 
		SET user_id = :user_id, token_hash = :token_hash, expires_at = :expires_at, 
		    is_revoked = :is_revoked, updated_at = :updated_at
		WHERE id = :id
	`

	repoRefreshToken := &RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID.String(),
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		IsRevoked: refreshToken.IsRevoked,
		CreatedAt: refreshToken.CreatedAt,
		UpdatedAt: refreshToken.UpdatedAt,
	}

	_, err := r.db.NamedExecContext(ctx, query, repoRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	return nil
}

// RevokeByID revokes a refresh token by ID
func (r *RefreshTokenRepository) RevokeByID(ctx context.Context, id string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = true, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token by ID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrTokenNotFound
	}

	return nil
}

// RevokeByUserID revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = true, updated_at = $2
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID.String(), time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens by user ID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrTokenNotFound
	}

	return nil
}

// DeleteExpired deletes expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens 
		WHERE expires_at < ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

// Delete deletes a refresh token by ID
func (r *RefreshTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrTokenNotFound
	}

	return nil
}

// CreateWithTx creates a new refresh token within a transaction
func (r *RefreshTokenRepository) CreateWithTx(ctx context.Context, tx *tx.TxWrapper, refreshToken *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, is_revoked, created_at, updated_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :is_revoked, :created_at, :updated_at)
	`

	repoRefreshToken := &RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID.String(),
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		IsRevoked: refreshToken.IsRevoked,
		CreatedAt: refreshToken.CreatedAt,
		UpdatedAt: refreshToken.UpdatedAt,
	}

	_, err := tx.NamedExecContext(ctx, query, repoRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}
