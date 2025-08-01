package service

import (
	"context"
	"errors"
	"testing"
	"time"

	models "user-svc/internal/domain"
	"user-svc/internal/domain/dto"
	"user-svc/internal/domain/errs"
	"user-svc/mocks"
	"user-svc/utils/tx"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful registration", func(t *testing.T) {
		req := dto.RegisterReq{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "Password123!",
		}

		expectedUser := &models.User{
			ID:        uuid.New(),
			Email:     models.Email("test@example.com"),
			Username:  models.Username("testuser"),
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		}

		accessToken := "access_token_123"
		refreshToken := "refresh_token_123"

		mockTxManager.EXPECT().
			WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(*tx.TxWrapper) error) error {
				return fn(&tx.TxWrapper{})
			})

		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				user.ID = expectedUser.ID
				user.CreatedAt = expectedUser.CreatedAt
				user.UpdatedAt = expectedUser.UpdatedAt
				return nil
			})

		mockTokenMaker.EXPECT().
			CreateToken("testuser", int64(3600)).
			Return(accessToken, nil)

		mockTokenMaker.EXPECT().
			CreateRefreshToken("testuser", int64(604800)).
			Return(refreshToken, nil)

		mockRefreshTokenRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		result, err := service.Register(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedUser.Email, result.User.Email)
		assert.Equal(t, expectedUser.Username, result.User.Username)
		assert.Equal(t, accessToken, result.AccessToken)
		assert.Equal(t, refreshToken, result.RefreshToken)
	})

	t.Run("invalid request validation", func(t *testing.T) {
		req := dto.RegisterReq{
			Email:    "invalid-email",
			Username: "ab", // too short
			Password: "weak",
		}

		result, err := service.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("user creation error", func(t *testing.T) {
		req := dto.RegisterReq{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "Password123!",
		}

		mockTxManager.EXPECT().
			WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(*tx.TxWrapper) error) error {
				return fn(&tx.TxWrapper{})
			})

		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		result, err := service.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("token creation error", func(t *testing.T) {
		req := dto.RegisterReq{
			Email:    "test@example.com",
			Username: "testuser",
			Password: "Password123!",
		}

		mockTxManager.EXPECT().
			WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(*tx.TxWrapper) error) error {
				return fn(&tx.TxWrapper{})
			})

		mockUserRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		mockTokenMaker.EXPECT().
			CreateToken("testuser", int64(3600)).
			Return("", errors.New("token creation error"))

		result, err := service.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful login", func(t *testing.T) {
		req := dto.LoginReq{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		// Generate a proper hashed password for the test password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:           uuid.New(),
			Email:        models.Email("test@example.com"),
			Username:     models.Username("testuser"),
			PasswordHash: models.PasswordHash(string(hashedPassword)),
			CreatedAt:    time.Now().UnixMilli(),
			UpdatedAt:    time.Now().UnixMilli(),
		}

		accessToken := "access_token_123"
		refreshToken := "refresh_token_123"

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), "test@example.com").
			Return(user, nil)

		mockTokenMaker.EXPECT().
			CreateToken("testuser", int64(3600)).
			Return(accessToken, nil)

		mockTokenMaker.EXPECT().
			CreateRefreshToken("testuser", int64(604800)).
			Return(refreshToken, nil)

		mockRefreshTokenRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		result, err := service.Login(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user, result.User)
		assert.Equal(t, accessToken, result.AccessToken)
		assert.Equal(t, refreshToken, result.RefreshToken)
	})

	t.Run("invalid request validation", func(t *testing.T) {
		req := dto.LoginReq{
			Email:    "invalid-email",
			Password: "",
		}

		result, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("user not found", func(t *testing.T) {
		req := dto.LoginReq{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), "test@example.com").
			Return(nil, errors.New("user not found"))

		result, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidCredentials, err)
		assert.Nil(t, result)
	})

	t.Run("invalid password", func(t *testing.T) {
		req := dto.LoginReq{
			Email:    "test@example.com",
			Password: "WrongPassword123!",
		}

		hashedPassword := "$2a$10$abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz"
		user := &models.User{
			ID:           uuid.New(),
			Email:        models.Email("test@example.com"),
			Username:     models.Username("testuser"),
			PasswordHash: models.PasswordHash(hashedPassword),
			CreatedAt:    time.Now().UnixMilli(),
			UpdatedAt:    time.Now().UnixMilli(),
		}

		mockUserRepo.EXPECT().
			GetByEmail(gomock.Any(), "test@example.com").
			Return(user, nil)

		result, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidCredentials, err)
		assert.Nil(t, result)
	})
}

func TestUserService_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful token refresh", func(t *testing.T) {
		req := dto.RefreshTokenReq{
			RefreshToken: "valid_refresh_token",
		}

		userID := uuid.New()
		user := &models.User{
			ID:       userID,
			Email:    models.Email("test@example.com"),
			Username: models.Username("testuser"),
		}

		refreshTokenRecord := &models.RefreshToken{
			ID:        "token_id_123",
			UserID:    userID,
			TokenHash: "hashed_token",
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IsRevoked: false,
		}

		newAccessToken := "new_access_token_123"

		mockTokenMaker.EXPECT().
			VerifyRefreshToken("valid_refresh_token").
			Return("payload", nil)

		mockRefreshTokenRepo.EXPECT().
			GetByTokenHash(gomock.Any(), gomock.Any()).
			Return(refreshTokenRecord, nil)

		mockUserRepo.EXPECT().
			GetByID(gomock.Any(), userID).
			Return(user, nil)

		mockTokenMaker.EXPECT().
			CreateToken("testuser", int64(3600)).
			Return(newAccessToken, nil)

		mockRefreshTokenRepo.EXPECT().
			RevokeByID(gomock.Any(), "token_id_123").
			Return(nil)

		result, err := service.RefreshToken(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newAccessToken, result.AccessToken)
	})

	t.Run("empty refresh token", func(t *testing.T) {
		req := dto.RefreshTokenReq{
			RefreshToken: "",
		}

		result, err := service.RefreshToken(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidToken, err)
		assert.Nil(t, result)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		req := dto.RefreshTokenReq{
			RefreshToken: "invalid_token",
		}

		mockTokenMaker.EXPECT().
			VerifyRefreshToken("invalid_token").
			Return(nil, errors.New("invalid token"))

		result, err := service.RefreshToken(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidToken, err)
		assert.Nil(t, result)
	})

	t.Run("token not found in database", func(t *testing.T) {
		req := dto.RefreshTokenReq{
			RefreshToken: "valid_refresh_token",
		}

		mockTokenMaker.EXPECT().
			VerifyRefreshToken("valid_refresh_token").
			Return("payload", nil)

		mockRefreshTokenRepo.EXPECT().
			GetByTokenHash(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("token not found"))

		result, err := service.RefreshToken(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidToken, err)
		assert.Nil(t, result)
	})
}

func TestUserService_RevokeToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful token revocation", func(t *testing.T) {
		req := dto.RevokeTokenReq{
			RefreshToken: "valid_refresh_token",
		}

		refreshTokenRecord := &models.RefreshToken{
			ID:        "token_id_123",
			UserID:    uuid.New(),
			TokenHash: "hashed_token",
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IsRevoked: false,
		}

		mockRefreshTokenRepo.EXPECT().
			GetByTokenHash(gomock.Any(), gomock.Any()).
			Return(refreshTokenRecord, nil)

		mockRefreshTokenRepo.EXPECT().
			RevokeByID(gomock.Any(), "token_id_123").
			Return(nil)

		err := service.RevokeToken(context.Background(), req)

		assert.NoError(t, err)
	})

	t.Run("empty refresh token", func(t *testing.T) {
		req := dto.RevokeTokenReq{
			RefreshToken: "",
		}

		err := service.RevokeToken(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidToken, err)
	})

	t.Run("token not found", func(t *testing.T) {
		req := dto.RevokeTokenReq{
			RefreshToken: "valid_refresh_token",
		}

		mockRefreshTokenRepo.EXPECT().
			GetByTokenHash(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("token not found"))

		err := service.RevokeToken(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrInvalidToken, err)
	})
}

func TestUserService_RevokeAllUserTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful revocation of all user tokens", func(t *testing.T) {
		userID := uuid.New()

		mockRefreshTokenRepo.EXPECT().
			RevokeByUserID(gomock.Any(), userID).
			Return(nil)

		err := service.RevokeAllUserTokens(context.Background(), userID)

		assert.NoError(t, err)
	})

	t.Run("error revoking user tokens", func(t *testing.T) {
		userID := uuid.New()

		mockRefreshTokenRepo.EXPECT().
			RevokeByUserID(gomock.Any(), userID).
			Return(errors.New("database error"))

		err := service.RevokeAllUserTokens(context.Background(), userID)

		assert.Error(t, err)
	})
}

func TestUserService_CleanupExpiredTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRefreshTokenRepo := mocks.NewMockRefreshTokenRepository(ctrl)
	mockTokenMaker := mocks.NewMockTokenMaker(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	service := NewUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockTokenMaker,
		1*time.Hour,
		7*24*time.Hour,
		mockTxManager,
	)

	t.Run("successful cleanup of expired tokens", func(t *testing.T) {
		mockRefreshTokenRepo.EXPECT().
			DeleteExpired(gomock.Any()).
			Return(nil)

		err := service.CleanupExpiredTokens(context.Background())

		assert.NoError(t, err)
	})

	t.Run("error cleaning up expired tokens", func(t *testing.T) {
		mockRefreshTokenRepo.EXPECT().
			DeleteExpired(gomock.Any()).
			Return(errors.New("database error"))

		err := service.CleanupExpiredTokens(context.Background())

		assert.Error(t, err)
	})
}
