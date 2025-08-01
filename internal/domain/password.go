package models

import "user-svc/internal/domain/errs"

// Password represents a validated password
type Password string

// NewPassword creates a new Password and validates it
func NewPassword(password string) (Password, error) {
	p := Password(password)
	if err := p.Validate(); err != nil {
		return "", err
	}
	return p, nil
}

// Validate checks if the password meets security requirements
func (p Password) Validate() error {
	password := string(p)

	// Check if password is empty
	if password == "" {
		return errs.ErrInvalidPassword
	}

	// Check minimum length (at least 8 characters)
	if len(password) < 8 {
		return errs.ErrInvalidPassword
	}

	// Check maximum length (reasonable limit)
	if len(password) > 32 {
		return errs.ErrInvalidPassword
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			hasSpecial = true
		}
	}

	// Password must have at least 3 of the 4 character types
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	if score < 3 {
		return errs.ErrInvalidPassword
	}

	return nil
}

// String returns the password as a string
func (p Password) String() string {
	return string(p)
}
