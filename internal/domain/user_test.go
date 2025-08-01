package models

import (
	"strings"
	"testing"

	"user-svc/internal/domain/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEmail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with subdomain",
			email:   "test@sub.example.com",
			wantErr: nil,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "email too short",
			email:   "a@b",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "email too long",
			email:   "a" + string(make([]byte, 254)) + "@example.com",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "no @ symbol",
			email:   "testexample.com",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "@ at beginning",
			email:   "@example.com",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "@ at end",
			email:   "test@",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "multiple @ symbols",
			email:   "test@@example.com",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "domain too short",
			email:   "test@a",
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "domain too long",
			email:   "test@" + string(make([]byte, 254)),
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name:    "no dot in domain",
			email:   "test@example",
			wantErr: errs.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := Email(tt.email)
			err := email.Validate()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		want    Email
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			want:    Email("test@example.com"),
			wantErr: nil,
		},
		{
			name:    "invalid email",
			email:   "invalid-email",
			want:    "",
			wantErr: errs.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.email)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestEmail_String(t *testing.T) {
	email := Email("test@example.com")
	assert.Equal(t, "test@example.com", email.String())
}

func TestNewUser_WithEmailValidation(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		passwordHash string
		username     string
		wantErr      error
	}{
		{
			name:         "valid user",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			username:     "testuser",
			wantErr:      nil,
		},
		{
			name:         "invalid email",
			email:        "invalid-email",
			passwordHash: "hashedpassword",
			username:     "testuser",
			wantErr:      errs.ErrInvalidEmail,
		},
		{
			name:         "empty email",
			email:        "",
			passwordHash: "hashedpassword",
			username:     "testuser",
			wantErr:      errs.ErrEmailIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.passwordHash, tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, Email(tt.email), user.Email)
				assert.Equal(t, Username(tt.username), user.Username)
				assert.Equal(t, tt.passwordHash, user.PasswordHash.String())
			}
		})
	}
}

func TestUser_IsValid(t *testing.T) {
	testID := uuid.New()
	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name: "valid user",
			user: &User{
				ID:           testID,
				Email:        Email("test@example.com"),
				Username:     Username("testuser"),
				PasswordHash: PasswordHash("hashedpassword"),
				CreatedAt:    1234567890,
				UpdatedAt:    1234567890,
			},
			wantErr: nil,
		},
		{
			name: "empty email",
			user: &User{
				ID:           testID,
				Email:        "",
				Username:     Username("testuser"),
				PasswordHash: PasswordHash("hashedpassword"),
				CreatedAt:    1234567890,
				UpdatedAt:    1234567890,
			},
			wantErr: errs.ErrEmailIsRequired,
		},
		{
			name: "invalid email",
			user: &User{
				ID:           testID,
				Email:        Email("invalid-email"),
				Username:     Username("testuser"),
				PasswordHash: PasswordHash("hashedpassword"),
				CreatedAt:    1234567890,
				UpdatedAt:    1234567890,
			},
			wantErr: errs.ErrInvalidEmail,
		},
		{
			name: "empty username",
			user: &User{
				ID:           testID,
				Email:        Email("test@example.com"),
				Username:     Username(""),
				PasswordHash: PasswordHash("hashedpassword"),
				CreatedAt:    1234567890,
				UpdatedAt:    1234567890,
			},
			wantErr: errs.ErrInvalidUsername,
		},
		{
			name: "empty password hash",
			user: &User{
				ID:           testID,
				Email:        Email("test@example.com"),
				Username:     Username("testuser"),
				PasswordHash: PasswordHash(""),
				CreatedAt:    1234567890,
				UpdatedAt:    1234567890,
			},
			wantErr: errs.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.IsValid()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPassword_Validate(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password with all character types",
			password: "TestPass123!",
			wantErr:  nil,
		},
		{
			name:     "valid password with uppercase, lowercase, and digits",
			password: "TestPass123",
			wantErr:  nil,
		},
		{
			name:     "valid password with uppercase, lowercase, and special chars",
			password: "TestPass!@#",
			wantErr:  nil,
		},
		{
			name:     "valid password with lowercase, digits, and special chars",
			password: "testpass123!",
			wantErr:  nil,
		},
		{
			name:     "valid password with uppercase, digits, and special chars",
			password: "TESTPASS123!",
			wantErr:  nil,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password too short",
			password: "Test1!",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password too long",
			password: "TestPass123!" + string(make([]byte, 120)),
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only lowercase and digits",
			password: "testpass123",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only uppercase and digits",
			password: "TESTPASS123",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only lowercase and special chars",
			password: "testpass!@#",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only uppercase and special chars",
			password: "TESTPASS!@#",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only digits and special chars",
			password: "12345678!@#",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only lowercase letters",
			password: "testpassword",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only uppercase letters",
			password: "TESTPASSWORD",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only digits",
			password: "12345678",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "password with only special characters",
			password: "!@#$%^&*",
			wantErr:  errs.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password := Password(tt.password)
			err := password.Validate()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     Password
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "TestPass123!",
			want:     Password("TestPass123!"),
			wantErr:  nil,
		},
		{
			name:     "invalid password",
			password: "weak",
			want:     "",
			wantErr:  errs.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPassword(tt.password)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPassword_String(t *testing.T) {
	password := Password("TestPass123!")
	assert.Equal(t, "TestPass123!", password.String())
}

func TestNewUserWithPassword(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		wantErr  error
	}{
		{
			name:     "valid user with password",
			email:    "test@example.com",
			password: "TestPass123!",
			username: "testuser",
			wantErr:  nil,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "weak",
			username: "testuser",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			password: "TestPass123!",
			username: "testuser",
			wantErr:  errs.ErrInvalidEmail,
		},
		{
			name:     "empty email",
			email:    "",
			password: "TestPass123!",
			username: "testuser",
			wantErr:  errs.ErrEmailIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserWithPassword(tt.email, tt.password, tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, Email(tt.email), user.Email)
				assert.Equal(t, Username(tt.username), user.Username)
				assert.Equal(t, "", user.PasswordHash.String()) // Should be empty as it needs to be hashed
			}
		})
	}
}

func TestUsername_Validate(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "valid username with letters and numbers",
			username: "testuser123",
			wantErr:  nil,
		},
		{
			name:     "valid username with underscore",
			username: "test_user",
			wantErr:  nil,
		},
		{
			name:     "valid username with hyphen",
			username: "test-user",
			wantErr:  nil,
		},
		{
			name:     "valid username with mixed characters",
			username: "TestUser_123",
			wantErr:  nil,
		},
		{
			name:     "valid username minimum length",
			username: "abc",
			wantErr:  nil,
		},
		{
			name:     "valid username maximum length",
			username: "a" + strings.Repeat("b", 29),
			wantErr:  nil,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username too short",
			username: "ab",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username too long",
			username: "a" + strings.Repeat("b", 30),
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username with invalid characters",
			username: "test@user",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username with spaces",
			username: "test user",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username starting with underscore",
			username: "_testuser",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username ending with underscore",
			username: "testuser_",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username starting with hyphen",
			username: "-testuser",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username ending with hyphen",
			username: "testuser-",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username with consecutive underscores",
			username: "test__user",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username with consecutive hyphens",
			username: "test--user",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "username with special characters",
			username: "test!user",
			wantErr:  errs.ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := Username(tt.username)
			err := username.Validate()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     Username
		wantErr  error
	}{
		{
			name:     "valid username",
			username: "testuser123",
			want:     Username("testuser123"),
			wantErr:  nil,
		},
		{
			name:     "invalid username",
			username: "ab",
			want:     "",
			wantErr:  errs.ErrInvalidUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUsername(tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUsername_String(t *testing.T) {
	username := Username("testuser123")
	assert.Equal(t, "testuser123", username.String())
}

func TestNewUser_WithUsernameValidation(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		passwordHash string
		username     string
		wantErr      error
	}{
		{
			name:         "valid user with username",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			username:     "testuser123",
			wantErr:      nil,
		},
		{
			name:         "invalid username",
			email:        "test@example.com",
			passwordHash: "hashedpassword",
			username:     "ab",
			wantErr:      errs.ErrInvalidUsername,
		},
		{
			name:         "invalid email",
			email:        "invalid-email",
			passwordHash: "hashedpassword",
			username:     "testuser123",
			wantErr:      errs.ErrInvalidEmail,
		},
		{
			name:         "empty email",
			email:        "",
			passwordHash: "hashedpassword",
			username:     "testuser123",
			wantErr:      errs.ErrEmailIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.passwordHash, tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, Email(tt.email), user.Email)
				assert.Equal(t, Username(tt.username), user.Username)
				assert.Equal(t, tt.passwordHash, user.PasswordHash.String())
			}
		})
	}
}

func TestNewUserWithPassword_WithUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		wantErr  error
	}{
		{
			name:     "valid user with password and username",
			email:    "test@example.com",
			password: "TestPass123!",
			username: "testuser123",
			wantErr:  nil,
		},
		{
			name:     "invalid username",
			email:    "test@example.com",
			password: "TestPass123!",
			username: "ab",
			wantErr:  errs.ErrInvalidUsername,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "weak",
			username: "testuser123",
			wantErr:  errs.ErrInvalidPassword,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			password: "TestPass123!",
			username: "testuser123",
			wantErr:  errs.ErrInvalidEmail,
		},
		{
			name:     "empty email",
			email:    "",
			password: "TestPass123!",
			username: "testuser123",
			wantErr:  errs.ErrEmailIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserWithPassword(tt.email, tt.password, tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, Email(tt.email), user.Email)
				assert.Equal(t, Username(tt.username), user.Username)
				assert.Equal(t, "", user.PasswordHash.String()) // Should be empty as it needs to be hashed
			}
		})
	}
}
