package dto

import (
	models "user-svc/internal/domain"
	"user-svc/internal/domain/errs"
)

// RegisterReq represents a user registration request
type RegisterReq struct {
	Email    string
	Username string
	Password string
}

// Validate validates the registration request
func (req RegisterReq) Validate() error {
	// Check if email is provided
	if req.Email == "" {
		return errs.ErrEmailIsRequired
	}

	// Validate email using the Email type
	if _, err := models.NewEmail(req.Email); err != nil {
		return err
	}

	// Validate username using the Username type
	if _, err := models.NewUsername(req.Username); err != nil {
		return err
	}

	// Validate password using the Password type
	if _, err := models.NewPassword(req.Password); err != nil {
		return err
	}

	return nil
}

// RegisterResp represents a user registration response
type RegisterResp struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

// LoginReq represents a user login request
type LoginReq struct {
	Email    string
	Password string
}

// Validate validates the login request
func (req LoginReq) Validate() error {
	// Check if email is provided
	if req.Email == "" {
		return errs.ErrEmailIsRequired
	}

	// Validate email using the Email type
	if _, err := models.NewEmail(req.Email); err != nil {
		return err
	}

	// Validate password
	if req.Password == "" {
		return errs.ErrInvalidPassword
	}

	return nil
}

// LoginResp represents a user login response
type LoginResp struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

// RefreshTokenReq represents a refresh token request
type RefreshTokenReq struct {
	RefreshToken string
}

// RefreshTokenResp represents a refresh token response
type RefreshTokenResp struct {
	AccessToken string
}

// RevokeTokenReq represents a token revocation request
type RevokeTokenReq struct {
	RefreshToken string
}
