package models

import "user-svc/internal/domain/errs"

// Email represents a validated email address
type Email string

// NewEmail creates a new Email and validates it
func NewEmail(email string) (Email, error) {
	e := Email(email)
	if err := e.Validate(); err != nil {
		return "", err
	}
	return e, nil
}

// Validate checks if the email format is valid
func (e Email) Validate() error {
	email := string(e)

	// Check length
	if len(email) < 5 || len(email) > 254 {
		return errs.ErrInvalidEmail
	}

	// Check for @ symbol
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return errs.ErrInvalidEmail // Multiple @ symbols
			}
			atIndex = i
		}
	}

	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return errs.ErrInvalidEmail
	}

	// Check for domain part
	domain := email[atIndex+1:]
	if len(domain) < 2 || len(domain) > 253 {
		return errs.ErrInvalidEmail
	}

	// Check for dot in domain
	hasDot := false
	for _, char := range domain {
		if char == '.' {
			hasDot = true
			break
		}
	}

	if !hasDot {
		return errs.ErrInvalidEmail
	}

	return nil
}

// String returns the email as a string
func (e Email) String() string {
	return string(e)
}
