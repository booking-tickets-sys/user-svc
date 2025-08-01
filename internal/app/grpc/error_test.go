package grpc

import (
	"context"
	"errors"
	"testing"

	"user-svc/internal/domain/errs"
	"user-svc/token"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorMapper_MapError(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "ErrInvalidEmail",
			err:          errs.ErrInvalidEmail,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "invalid email format",
		},
		{
			name:         "ErrInvalidUsername",
			err:          errs.ErrInvalidUsername,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "invalid username",
		},
		{
			name:         "ErrInvalidPassword",
			err:          errs.ErrInvalidPassword,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "invalid password",
		},
		{
			name:         "ErrUserNotFound",
			err:          errs.ErrUserNotFound,
			expectedCode: codes.NotFound,
			expectedMsg:  "user not found",
		},
		{
			name:         "ErrUserExists",
			err:          errs.ErrUserExists,
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "user already exists",
		},
		{
			name:         "ErrInvalidToken",
			err:          errs.ErrInvalidToken,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "invalid token",
		},
		{
			name:         "ErrTokenExpired",
			err:          errs.ErrTokenExpired,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token expired",
		},
		{
			name:         "ErrTokenRevoked",
			err:          errs.ErrTokenRevoked,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token revoked",
		},
		{
			name:         "ErrTokenNotFound",
			err:          errs.ErrTokenNotFound,
			expectedCode: codes.NotFound,
			expectedMsg:  "token not found",
		},
		{
			name:         "ErrInvalidCredentials",
			err:          errs.ErrInvalidCredentials,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "invalid credentials",
		},
		{
			name:         "ErrEmailIsRequired",
			err:          errs.ErrEmailIsRequired,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "email is required",
		},
		{
			name:         "TokenErrInvalidToken",
			err:          token.ErrInvalidToken,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "invalid token",
		},
		{
			name:         "TokenErrExpiredToken",
			err:          token.ErrExpiredToken,
			expectedCode: codes.Unauthenticated,
			expectedMsg:  "token expired",
		},
		{
			name:         "DatabaseConstraintError",
			err:          errors.New("UNIQUE constraint failed: users.email"),
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "resource already exists",
		},
		{
			name:         "ContextCanceled",
			err:          context.Canceled,
			expectedCode: codes.Canceled,
			expectedMsg:  "operation was canceled",
		},
		{
			name:         "ContextDeadlineExceeded",
			err:          context.DeadlineExceeded,
			expectedCode: codes.DeadlineExceeded,
			expectedMsg:  "operation deadline exceeded",
		},
		{
			name:         "UnknownError",
			err:          errors.New("unknown error"),
			expectedCode: codes.Internal,
			expectedMsg:  "internal server error: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapError(tt.err)

			st, ok := status.FromError(result)
			assert.True(t, ok, "Error should be a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}

func TestErrorMapper_MapErrorWithContext(t *testing.T) {
	mapper := NewErrorMapper()
	operation := "TestOperation"

	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "ErrInvalidEmail with context",
			err:          errs.ErrInvalidEmail,
			expectedCode: codes.InvalidArgument,
			expectedMsg:  "TestOperation: invalid email format",
		},
		{
			name:         "ErrUserNotFound with context",
			err:          errs.ErrUserNotFound,
			expectedCode: codes.NotFound,
			expectedMsg:  "TestOperation: user not found",
		},
		{
			name:         "ContextCanceled with context",
			err:          context.Canceled,
			expectedCode: codes.Canceled,
			expectedMsg:  "TestOperation: operation was canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapErrorWithContext(tt.err, operation)

			st, ok := status.FromError(result)
			assert.True(t, ok, "Error should be a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}

func TestErrorMapper_HandleContextError(t *testing.T) {
	mapper := NewErrorMapper()

	t.Run("Context not canceled", func(t *testing.T) {
		ctx := context.Background()
		err := mapper.HandleContextError(ctx, "TestOperation")
		assert.NoError(t, err)
	})

	t.Run("Context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := mapper.HandleContextError(ctx, "TestOperation")
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Canceled, st.Code())
		assert.Contains(t, st.Message(), "TestOperation")
	})
}

func TestErrorMapper_isDatabaseConstraintError(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "UNIQUE constraint failed",
			err:      errors.New("UNIQUE constraint failed: users.email"),
			expected: true,
		},
		{
			name:     "duplicate key value",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "unique constraint",
			err:      errors.New("unique constraint violation"),
			expected: true,
		},
		{
			name:     "duplicate entry",
			err:      errors.New("duplicate entry for key"),
			expected: true,
		},
		{
			name:     "already exists",
			err:      errors.New("resource already exists"),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.isDatabaseConstraintError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorMapper_isContextError(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Context canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "Context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "Regular error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.isContextError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
