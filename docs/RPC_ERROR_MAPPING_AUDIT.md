# RPC Error Mapping Audit Report

## Overview

This document provides a comprehensive audit of the gRPC error mapping implementation in the User Service application. The audit was conducted to ensure all domain errors are properly mapped to appropriate gRPC status codes.

## Current Error Mapping Status

### ✅ **Properly Mapped Domain Errors**

All domain errors from `internal/domain/errs/errors.go` are correctly mapped:

| Domain Error | gRPC Status Code | Message |
|--------------|------------------|---------|
| `ErrInvalidEmail` | `InvalidArgument` | "invalid email format" |
| `ErrInvalidUsername` | `InvalidArgument` | "invalid username" |
| `ErrInvalidPassword` | `InvalidArgument` | "invalid password" |
| `ErrUserNotFound` | `NotFound` | "user not found" |
| `ErrUserExists` | `AlreadyExists` | "user already exists" |
| `ErrInvalidToken` | `Unauthenticated` | "invalid token" |
| `ErrTokenExpired` | `Unauthenticated` | "token expired" |
| `ErrTokenRevoked` | `Unauthenticated` | "token revoked" |
| `ErrTokenNotFound` | `NotFound` | "token not found" |
| `ErrInvalidCredentials` | `Unauthenticated` | "invalid credentials" |
| `ErrEmailIsRequired` | `InvalidArgument` | "email is required" |

### ✅ **Token Package Errors**

Token package errors are now properly mapped:

| Token Error | gRPC Status Code | Message |
|-------------|------------------|---------|
| `token.ErrInvalidToken` | `Unauthenticated` | "invalid token" |
| `token.ErrExpiredToken` | `Unauthenticated` | "token expired" |

### ✅ **Database Constraint Errors**

Database constraint violations are now handled:

| Error Pattern | gRPC Status Code | Message |
|---------------|------------------|---------|
| "UNIQUE constraint failed" | `AlreadyExists` | "resource already exists" |
| "duplicate key value" | `AlreadyExists` | "resource already exists" |
| "unique constraint" | `AlreadyExists` | "resource already exists" |
| "duplicate entry" | `AlreadyExists` | "resource already exists" |
| "already exists" | `AlreadyExists` | "resource already exists" |

### ✅ **Context Errors**

Context-related errors are properly handled:

| Context Error | gRPC Status Code | Message |
|---------------|------------------|---------|
| `context.Canceled` | `Canceled` | "operation was canceled" |
| `context.DeadlineExceeded` | `DeadlineExceeded` | "operation deadline exceeded" |

### ✅ **Unknown Errors**

Any unmapped errors default to:

| Error Type | gRPC Status Code | Message |
|------------|------------------|---------|
| Unknown/Unmapped | `Internal` | "internal server error" |

## Implementation Details

### Error Mapper Structure

The `ErrorMapper` in `internal/app/grpc/errors.go` provides:

1. **`MapError(err error) error`** - Basic error mapping
2. **`MapErrorWithContext(err error, operation string) error`** - Error mapping with operation context
3. **`HandleContextError(ctx context.Context, operation string) error`** - Context error handling
4. **Helper methods** for detecting specific error types

### gRPC Server Integration

The gRPC server in `internal/app/grpc/server.go` now:

1. **Checks context errors first** in each RPC method
2. **Uses error mapper directly** for consistent error handling
3. **Provides operation context** in error messages for better debugging

### Middleware Integration

The middleware in `internal/app/grpc/middleware.go` provides:

1. **Automatic error mapping** for all RPC calls
2. **Panic recovery** with proper error mapping
3. **Request/response logging** with error context

## Error Flow

```
RPC Request → Context Check → Service Call → Error Mapping → gRPC Response
     ↓              ↓              ↓              ↓              ↓
   Validate    Check Cancel    Business     Map to gRPC    Return Status
   Context     Deadline       Logic        Status Code    with Message
```

## Testing Coverage

Comprehensive tests in `internal/app/grpc/error_test.go` cover:

- ✅ All domain error mappings
- ✅ Token package error mappings
- ✅ Database constraint error detection
- ✅ Context error handling
- ✅ Unknown error fallback
- ✅ Error mapping with context

## Recommendations

### ✅ **Completed Improvements**

1. **Added missing token package error mappings**
2. **Added database constraint error handling**
3. **Added context error handling**
4. **Improved gRPC server error handling**
5. **Added comprehensive test coverage**

### 🔄 **Future Enhancements**

1. **Add metrics collection** for error types and frequencies
2. **Implement structured logging** for better error tracking
3. **Add error correlation IDs** for distributed tracing
4. **Consider adding retry logic** for transient errors

## Conclusion

The RPC error mapping is now **comprehensive and robust**. All domain errors, token errors, database constraints, and context errors are properly mapped to appropriate gRPC status codes. The implementation includes:

- ✅ **Complete error coverage**
- ✅ **Consistent error handling**
- ✅ **Proper status code mapping**
- ✅ **Context-aware error messages**
- ✅ **Comprehensive test coverage**

The error handling system is production-ready and follows gRPC best practices. 