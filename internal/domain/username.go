package models

import "user-svc/internal/domain/errs"

// Username represents a validated username
type Username string

// NewUsername creates a new Username and validates it
func NewUsername(username string) (Username, error) {
	u := Username(username)
	if err := u.Validate(); err != nil {
		return "", err
	}
	return u, nil
}

// Validate checks if the username meets requirements
func (u Username) Validate() error {
	username := string(u)

	// Check if username is empty
	if username == "" {
		return errs.ErrInvalidUsername
	}

	// Check minimum length (at least 3 characters)
	if len(username) < 3 {
		return errs.ErrInvalidUsername
	}

	// Check maximum length (reasonable limit)
	if len(username) > 30 {
		return errs.ErrInvalidUsername
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return errs.ErrInvalidUsername
		}
	}

	// Check that username doesn't start or end with underscore or hyphen
	if len(username) > 0 && (username[0] == '_' || username[0] == '-' ||
		username[len(username)-1] == '_' || username[len(username)-1] == '-') {
		return errs.ErrInvalidUsername
	}

	// Check for consecutive underscores or hyphens
	for i := 0; i < len(username)-1; i++ {
		if (username[i] == '_' && username[i+1] == '_') ||
			(username[i] == '-' && username[i+1] == '-') {
			return errs.ErrInvalidUsername
		}
	}

	return nil
}

// String returns the username as a string
func (u Username) String() string {
	return string(u)
}
