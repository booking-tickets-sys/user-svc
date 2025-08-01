package models

import (
	"testing"
	"time"

	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		tokenHash string
		expiresAt int64
		wantErr   error
	}{
		{
			name:      "valid refresh token",
			userID:    uuid.New(),
			tokenHash: "valid-hash",
			expiresAt: time.Now().Add(time.Hour).Unix(),
			wantErr:   nil,
		},
		{
			name:      "nil user ID",
			userID:    uuid.Nil,
			tokenHash: "valid-hash",
			expiresAt: time.Now().Add(time.Hour).Unix(),
			wantErr:   errs.ErrInvalidToken,
		},
		{
			name:      "empty token hash",
			userID:    uuid.New(),
			tokenHash: "",
			expiresAt: time.Now().Add(time.Hour).Unix(),
			wantErr:   errs.ErrInvalidToken,
		},
		{
			name:      "expired token",
			userID:    uuid.New(),
			tokenHash: "valid-hash",
			expiresAt: time.Now().Add(-time.Hour).Unix(),
			wantErr:   errs.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken, err := NewRefreshToken(tt.userID, tt.tokenHash, tt.expiresAt)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, refreshToken)
				assert.NotEmpty(t, refreshToken.ID)
				assert.Equal(t, tt.userID, refreshToken.UserID)
				assert.Equal(t, tt.tokenHash, refreshToken.TokenHash)
				assert.Equal(t, tt.expiresAt, refreshToken.ExpiresAt)
				assert.False(t, refreshToken.IsRevoked)
				assert.Greater(t, refreshToken.CreatedAt, int64(0))
				assert.Greater(t, refreshToken.UpdatedAt, int64(0))
			}
		})
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	validUserID := uuid.New()
	validTokenHash := "valid-hash"
	validExpiresAt := time.Now().Add(time.Hour).Unix()

	tests := []struct {
		name         string
		refreshToken *RefreshToken
		wantErr      error
	}{
		{
			name: "valid refresh token",
			refreshToken: &RefreshToken{
				ID:        uuid.New().String(),
				UserID:    validUserID,
				TokenHash: validTokenHash,
				ExpiresAt: validExpiresAt,
				IsRevoked: false,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: nil,
		},
		{
			name: "empty ID",
			refreshToken: &RefreshToken{
				ID:        "",
				UserID:    validUserID,
				TokenHash: validTokenHash,
				ExpiresAt: validExpiresAt,
				IsRevoked: false,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: errs.ErrInvalidToken,
		},
		{
			name: "nil user ID",
			refreshToken: &RefreshToken{
				ID:        uuid.New().String(),
				UserID:    uuid.Nil,
				TokenHash: validTokenHash,
				ExpiresAt: validExpiresAt,
				IsRevoked: false,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: errs.ErrInvalidToken,
		},
		{
			name: "empty token hash",
			refreshToken: &RefreshToken{
				ID:        uuid.New().String(),
				UserID:    validUserID,
				TokenHash: "",
				ExpiresAt: validExpiresAt,
				IsRevoked: false,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: errs.ErrInvalidToken,
		},
		{
			name: "revoked token",
			refreshToken: &RefreshToken{
				ID:        uuid.New().String(),
				UserID:    validUserID,
				TokenHash: validTokenHash,
				ExpiresAt: validExpiresAt,
				IsRevoked: true,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: errs.ErrTokenRevoked,
		},
		{
			name: "expired token",
			refreshToken: &RefreshToken{
				ID:        uuid.New().String(),
				UserID:    validUserID,
				TokenHash: validTokenHash,
				ExpiresAt: time.Now().Add(-time.Hour).Unix(),
				IsRevoked: false,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			},
			wantErr: errs.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.refreshToken.IsValid()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
