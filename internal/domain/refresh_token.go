package models

import (
	"time"

	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token domain model
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt int64     `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// NewRefreshToken creates a new RefreshToken
func NewRefreshToken(userID uuid.UUID, tokenHash string, expiresAt int64) (*RefreshToken, error) {
	if userID == uuid.Nil {
		return nil, errs.ErrInvalidToken
	}

	if tokenHash == "" {
		return nil, errs.ErrInvalidToken
	}

	if expiresAt <= time.Now().Unix() {
		return nil, errs.ErrTokenExpired
	}

	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

// IsValid checks if the refresh token is valid
func (rt *RefreshToken) IsValid() error {
	if rt.ID == "" {
		return errs.ErrInvalidToken
	}

	if rt.UserID == uuid.Nil {
		return errs.ErrInvalidToken
	}

	if rt.TokenHash == "" {
		return errs.ErrInvalidToken
	}

	if rt.IsRevoked {
		return errs.ErrTokenRevoked
	}

	if rt.ExpiresAt <= time.Now().Unix() {
		return errs.ErrTokenExpired
	}

	return nil
}
