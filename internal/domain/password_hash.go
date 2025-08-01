package models

import "user-svc/internal/domain/errs"

// PasswordHash represents a hashed password
type PasswordHash string

// NewPasswordHash creates a new PasswordHash and validates it
func NewPasswordHash(hash string) (PasswordHash, error) {
	ph := PasswordHash(hash)
	if err := ph.Validate(); err != nil {
		return "", err
	}
	return ph, nil
}

// Validate checks if the password hash is valid (non-empty)
func (ph PasswordHash) Validate() error {
	if string(ph) == "" {
		return errs.ErrInvalidPassword
	}
	return nil
}

// String returns the password hash as a string
func (ph PasswordHash) String() string {
	return string(ph)
}
