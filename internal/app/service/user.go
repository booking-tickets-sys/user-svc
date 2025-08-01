package service

import (
	"context"
	"time"

	models "user-svc/internal/domain"
	"user-svc/internal/domain/dto"
	"user-svc/internal/domain/errs"
	"user-svc/token"

	"user-svc/utils/tx"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TokenMaker interface {
	CreateToken(username string, duration int64) (string, error)
	CreateRefreshToken(username string, duration int64) (string, error)
	VerifyToken(token string) (interface{}, error)
	VerifyRefreshToken(token string) (interface{}, error)
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(*tx.TxWrapper) error) error
	WithTransactionResult(ctx context.Context, fn func(*tx.TxWrapper) (any, error)) (any, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *models.RefreshToken) error
	GetByID(ctx context.Context, id string) (*models.RefreshToken, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.RefreshToken, error)
	Update(ctx context.Context, refreshToken *models.RefreshToken) error
	RevokeByID(ctx context.Context, id string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	Delete(ctx context.Context, id string) error
}

type UserService struct {
	userRepo             UserRepository
	refreshTokenRepo     RefreshTokenRepository
	tokenMaker           TokenMaker
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
	txManager            TxManager
}

func NewUserService(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	tokenMaker TokenMaker,
	tokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	txManager TxManager,
) *UserService {
	return &UserService{
		userRepo:             userRepo,
		refreshTokenRepo:     refreshTokenRepo,
		tokenMaker:           tokenMaker,
		tokenDuration:        tokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		txManager:            txManager,
	}
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterReq) (*dto.RegisterResp, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user, err := models.NewUser(req.Email, string(hashedPassword), req.Username)
	if err != nil {
		return nil, err
	}

	// Validate user
	if err := user.IsValid(); err != nil {
		return nil, err
	}

	// Execute registration in transaction
	var result *dto.RegisterResp
	err = s.txManager.WithTransaction(ctx, func(tx *tx.TxWrapper) error {
		// Create user
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		// Generate access token
		accessToken, err := s.tokenMaker.CreateToken(user.Username.String(), int64(s.tokenDuration.Seconds()))
		if err != nil {
			return err
		}

		// Generate refresh token
		refreshToken, err := s.tokenMaker.CreateRefreshToken(user.Username.String(), int64(s.refreshTokenDuration.Seconds()))
		if err != nil {
			return err
		}

		// Hash the refresh token for storage
		_, tokenHash := token.GenerateTokenHash(refreshToken)

		// Create refresh token record
		refreshTokenRecord, err := models.NewRefreshToken(user.ID, tokenHash, int64(s.refreshTokenDuration.Seconds())+time.Now().Unix())
		if err != nil {
			return err
		}

		// Store refresh token in database
		if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
			return err
		}

		result = &dto.RegisterResp{
			User:         user,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginReq) (*dto.LoginResp, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String()), []byte(req.Password))
	if err != nil {
		return nil, errs.ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.tokenMaker.CreateToken(user.Username.String(), int64(s.tokenDuration.Seconds()))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenMaker.CreateRefreshToken(user.Username.String(), int64(s.refreshTokenDuration.Seconds()))
	if err != nil {
		return nil, err
	}

	// Hash the refresh token for storage
	_, tokenHash := token.GenerateTokenHash(refreshToken)

	// Create refresh token record
	refreshTokenRecord, err := models.NewRefreshToken(user.ID, tokenHash, int64(s.refreshTokenDuration.Seconds())+time.Now().Unix())
	if err != nil {
		return nil, err
	}

	// Store refresh token in database
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, err
	}

	return &dto.LoginResp{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) RefreshToken(ctx context.Context, req dto.RefreshTokenReq) (*dto.RefreshTokenResp, error) {
	if req.RefreshToken == "" {
		return nil, errs.ErrInvalidToken
	}

	// Verify the refresh token
	_, err := s.tokenMaker.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	// Hash the provided refresh token
	tokenHash := token.HashToken(req.RefreshToken)

	// Get the refresh token from database
	refreshTokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	// Validate the refresh token
	if err := refreshTokenRecord.IsValid(); err != nil {
		return nil, err
	}

	// Get the user to get the username
	user, err := s.userRepo.GetByID(ctx, refreshTokenRecord.UserID)
	if err != nil {
		return nil, errs.ErrUserNotFound
	}

	// Generate new access token
	newAccessToken, err := s.tokenMaker.CreateToken(user.Username.String(), int64(s.tokenDuration.Seconds()))
	if err != nil {
		return nil, err
	}

	// Revoke the old refresh token
	if err := s.refreshTokenRepo.RevokeByID(ctx, refreshTokenRecord.ID); err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResp{
		AccessToken: newAccessToken,
	}, nil
}

func (s *UserService) RevokeToken(ctx context.Context, req dto.RevokeTokenReq) error {
	if req.RefreshToken == "" {
		return errs.ErrInvalidToken
	}

	// Hash the refresh token
	tokenHash := token.HashToken(req.RefreshToken)

	// Get the refresh token from database
	refreshTokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return errs.ErrInvalidToken
	}

	// Revoke the token
	return s.refreshTokenRepo.RevokeByID(ctx, refreshTokenRecord.ID)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *UserService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeByUserID(ctx, userID)
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *UserService) CleanupExpiredTokens(ctx context.Context) error {
	return s.refreshTokenRepo.DeleteExpired(ctx)
}
