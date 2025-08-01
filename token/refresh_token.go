package token

import (
	"time"

	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the authentication system
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt int64     `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// NewRefreshToken creates a new refresh token with generated ID and timestamps
func NewRefreshToken(userID uuid.UUID, tokenHash string, expiresAt int64) (*RefreshToken, error) {
	if userID == uuid.Nil {
		return nil, errs.ErrUserNotFound
	}
	if tokenHash == "" {
		return nil, errs.ErrInvalidToken
	}
	if expiresAt <= time.Now().Unix() {
		return nil, errs.ErrInvalidToken
	}

	now := time.Now().UnixMilli()

	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsValid checks if the refresh token is valid
func (rt *RefreshToken) IsValid() error {
	if rt.UserID == uuid.Nil {
		return errs.ErrUserNotFound
	}
	if rt.TokenHash == "" {
		return errs.ErrInvalidToken
	}
	if rt.IsRevoked {
		return errs.ErrTokenRevoked
	}
	if time.Now().Unix() > rt.ExpiresAt {
		return errs.ErrTokenExpired
	}

	return nil
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().Unix() > rt.ExpiresAt
}

// Revoke marks the refresh token as revoked
func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
	rt.UpdatedAt = time.Now().UnixMilli()
}
