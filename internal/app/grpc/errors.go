package grpc

import (
	"context"
	"fmt"
	"strings"

	"user-svc/internal/domain/errs"
	"user-svc/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorMapper maps domain errors to gRPC status codes
type ErrorMapper struct{}

// NewErrorMapper creates a new error mapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapError maps domain errors to gRPC status codes
func (e *ErrorMapper) MapError(err error) error {
	switch {
	case err == errs.ErrInvalidEmail:
		return status.Errorf(codes.InvalidArgument, "invalid email format")
	case err == errs.ErrInvalidUsername:
		return status.Errorf(codes.InvalidArgument, "invalid username")
	case err == errs.ErrInvalidPassword:
		return status.Errorf(codes.InvalidArgument, "invalid password")
	case err == errs.ErrUserNotFound:
		return status.Errorf(codes.NotFound, "user not found")
	case err == errs.ErrUserExists:
		return status.Errorf(codes.AlreadyExists, "user already exists")
	case err == errs.ErrInvalidToken:
		return status.Errorf(codes.Unauthenticated, "invalid token")
	case err == errs.ErrTokenExpired:
		return status.Errorf(codes.Unauthenticated, "token expired")
	case err == errs.ErrTokenRevoked:
		return status.Errorf(codes.Unauthenticated, "token revoked")
	case err == errs.ErrTokenNotFound:
		return status.Errorf(codes.NotFound, "token not found")
	case err == errs.ErrInvalidCredentials:
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	case err == errs.ErrEmailIsRequired:
		return status.Errorf(codes.InvalidArgument, "email is required")
	// Token package errors
	case err == token.ErrInvalidToken:
		return status.Errorf(codes.Unauthenticated, "invalid token")
	case err == token.ErrExpiredToken:
		return status.Errorf(codes.Unauthenticated, "token expired")
	// Database constraint errors
	case e.isDatabaseConstraintError(err):
		return status.Errorf(codes.AlreadyExists, "resource already exists")
	// Context errors
	case e.isContextError(err):
		return e.mapContextError(err)
	default:
		// Log unknown errors for debugging
		e.LogError(err, "UNKNOWN_ERROR")
		return status.Errorf(codes.Internal, "internal server error: %v", err)
	}
}

// isDatabaseConstraintError checks if the error is a database constraint violation
func (e *ErrorMapper) isDatabaseConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "duplicate key value") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "already exists")
}

// isContextError checks if the error is a context-related error
func (e *ErrorMapper) isContextError(err error) bool {
	if err == nil {
		return false
	}

	return err == context.Canceled || err == context.DeadlineExceeded
}

// mapContextError maps context errors to appropriate gRPC status codes
func (e *ErrorMapper) mapContextError(err error) error {
	switch err {
	case context.Canceled:
		return status.Errorf(codes.Canceled, "operation was canceled")
	case context.DeadlineExceeded:
		return status.Errorf(codes.DeadlineExceeded, "operation deadline exceeded")
	default:
		return status.Errorf(codes.Internal, "context error")
	}
}

// LogError logs the error with operation context and additional debugging info
func (e *ErrorMapper) LogError(err error, operation string) {
	if err != nil {
		// Extract gRPC status details if available
		if st, ok := status.FromError(err); ok {
			fmt.Printf("ERROR in %s: code=%s, message=%s, details=%v\n",
				operation, st.Code(), st.Message(), st.Details())
		} else {
			fmt.Printf("ERROR in %s: %v\n", operation, err)
		}
	} else {
		fmt.Printf("SUCCESS in %s\n", operation)
	}
}

// HandleContextError checks if the context was cancelled and returns appropriate error
func (e *ErrorMapper) HandleContextError(ctx context.Context, operation string) error {
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.Canceled:
			return status.Errorf(codes.Canceled, "%s: operation was canceled", operation)
		case context.DeadlineExceeded:
			return status.Errorf(codes.DeadlineExceeded, "%s: operation deadline exceeded", operation)
		default:
			return status.Errorf(codes.Internal, "%s: context error", operation)
		}
	default:
		return nil
	}
}

// MapErrorWithContext maps domain errors to gRPC status codes with operation context
func (e *ErrorMapper) MapErrorWithContext(err error, operation string) error {
	switch {
	case err == errs.ErrInvalidEmail:
		return status.Errorf(codes.InvalidArgument, "%s: invalid email format", operation)
	case err == errs.ErrInvalidUsername:
		return status.Errorf(codes.InvalidArgument, "%s: invalid username", operation)
	case err == errs.ErrInvalidPassword:
		return status.Errorf(codes.InvalidArgument, "%s: invalid password", operation)
	case err == errs.ErrUserNotFound:
		return status.Errorf(codes.NotFound, "%s: user not found", operation)
	case err == errs.ErrUserExists:
		return status.Errorf(codes.AlreadyExists, "%s: user already exists", operation)
	case err == errs.ErrInvalidToken:
		return status.Errorf(codes.Unauthenticated, "%s: invalid token", operation)
	case err == errs.ErrTokenExpired:
		return status.Errorf(codes.Unauthenticated, "%s: token expired", operation)
	case err == errs.ErrTokenRevoked:
		return status.Errorf(codes.Unauthenticated, "%s: token revoked", operation)
	case err == errs.ErrTokenNotFound:
		return status.Errorf(codes.NotFound, "%s: token not found", operation)
	case err == errs.ErrInvalidCredentials:
		return status.Errorf(codes.Unauthenticated, "%s: invalid credentials", operation)
	case err == errs.ErrEmailIsRequired:
		return status.Errorf(codes.InvalidArgument, "%s: email is required", operation)
	// Token package errors
	case err == token.ErrInvalidToken:
		return status.Errorf(codes.Unauthenticated, "%s: invalid token", operation)
	case err == token.ErrExpiredToken:
		return status.Errorf(codes.Unauthenticated, "%s: token expired", operation)
	// Database constraint errors
	case e.isDatabaseConstraintError(err):
		return status.Errorf(codes.AlreadyExists, "%s: resource already exists", operation)
	// Context errors
	case e.isContextError(err):
		return e.mapContextErrorWithContext(err, operation)
	default:
		// Log unknown errors for debugging with more context
		e.LogError(err, fmt.Sprintf("UNKNOWN_ERROR_%s", operation))
		return status.Errorf(codes.Internal, "%s: internal server error - %v", operation, err)
	}
}

// mapContextErrorWithContext maps context errors with operation context
func (e *ErrorMapper) mapContextErrorWithContext(err error, operation string) error {
	switch err {
	case context.Canceled:
		return status.Errorf(codes.Canceled, "%s: operation was canceled", operation)
	case context.DeadlineExceeded:
		return status.Errorf(codes.DeadlineExceeded, "%s: operation deadline exceeded", operation)
	default:
		return status.Errorf(codes.Internal, "%s: context error", operation)
	}
}
