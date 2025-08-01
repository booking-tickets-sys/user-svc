# gRPC Error Handling Guide

This document explains how to properly handle gRPC errors in the User Service application.

## Overview

The application implements a comprehensive error handling system that maps domain errors to appropriate gRPC status codes and provides detailed error information to clients. **Validation logic is centralized in the service layer** for better separation of concerns.

## Architecture

### 1. Service Layer Validation (`internal/app/service/user.go`)

**All validation logic is now centralized in the service layer:**

```go
// validateRegisterRequest validates the registration request
func (s *UserService) validateRegisterRequest(req RegisterReq) error {
    // Check if email is provided
    if req.Email == "" {
        return errs.ErrEmailIsRequired
    }

    // Validate email
    if !s.isValidEmail(req.Email) {
        return errs.ErrInvalidEmail
    }

    // Validate username
    if req.Username == "" {
        return errs.ErrInvalidUsername
    }
    if len(req.Username) < 3 {
        return errs.ErrInvalidUsername
    }

    // Validate password
    if req.Password == "" {
        return errs.ErrInvalidPassword
    }
    if len(req.Password) < 6 {
        return errs.ErrInvalidPassword
    }

    return nil
}
```

### 2. Error Mapper (`internal/app/grpc/errors.go`)

The `ErrorMapper` is responsible for converting domain errors to gRPC status codes:

```go
type ErrorMapper struct{}

func (e *ErrorMapper) MapError(err error) error
func (e *ErrorMapper) MapErrorWithContext(err error, operation string) error
```

### 3. Domain Errors (`internal/domain/errs/errors.go`)

Domain-specific errors are defined in the `errs` package:

```go
var (
    ErrInvalidEmail       = errors.New("invalid email")
    ErrInvalidUsername    = errors.New("invalid username")
    ErrInvalidPassword    = errors.New("invalid password")
    ErrUserNotFound       = errors.New("user not found")
    ErrUserExists         = errors.New("user already exists")
    ErrInvalidToken       = errors.New("invalid token")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenRevoked       = errors.New("token revoked")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrEmailIsRequired    = errors.New("email is required")
)
```

### 4. gRPC Middleware (`internal/app/grpc/middleware.go`)

Middleware provides:
- Error logging
- Panic recovery
- Request/response timing
- Rate limiting
- Timeout handling

## Error Status Code Mapping

| Domain Error | gRPC Status Code | Description |
|--------------|------------------|-------------|
| `ErrInvalidEmail` | `InvalidArgument` | Invalid email format |
| `ErrInvalidUsername` | `InvalidArgument` | Invalid username |
| `ErrInvalidPassword` | `InvalidArgument` | Invalid password |
| `ErrUserNotFound` | `NotFound` | User not found |
| `ErrUserExists` | `AlreadyExists` | User already exists |
| `ErrInvalidToken` | `Unauthenticated` | Invalid token |
| `ErrTokenExpired` | `Unauthenticated` | Token expired |
| `ErrTokenRevoked` | `Unauthenticated` | Token revoked |
| `ErrInvalidCredentials` | `Unauthenticated` | Invalid credentials |
| `ErrEmailIsRequired` | `InvalidArgument` | Email is required |

## Usage Examples

### Server-Side Error Handling (Simplified)

**The gRPC server is now much cleaner since validation is handled in the service layer:**

```go
// In gRPC service implementation
func (s *UserServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
    // Check for context cancellation
    if err := s.errorMapper.HandleContextError(ctx, "Register"); err != nil {
        return nil, err
    }

    // Call service (validation is handled in service layer)
    result, err := s.userService.Register(ctx, service.RegisterReq{
        Email:    req.Email,
        Username: req.Username,
        Password: req.Password,
    })
    if err != nil {
        // Map domain errors to gRPC status codes
        mappedErr := s.errorMapper.MapError(err)
        s.errorMapper.LogError(mappedErr, "Register")
        return nil, mappedErr
    }

    return result, nil
}
```

### Service Layer Validation

**All business logic validation is centralized in the service layer:**

```go
func (s *UserService) Register(ctx context.Context, req RegisterReq) (*RegisterResp, error) {
    // Validate request (all validation logic here)
    if err := s.validateRegisterRequest(req); err != nil {
        return nil, err
    }

    // Business logic continues...
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // Create user and continue with business logic...
}
```

### Client-Side Error Handling

```go
func handleError(operation string, err error) {
    if err == nil {
        return
    }

    // Get gRPC status from error
    st, ok := status.FromError(err)
    if !ok {
        fmt.Printf("❌ %s: Unknown error: %v\n", operation, err)
        return
    }

    // Handle different status codes
    switch st.Code() {
    case codes.InvalidArgument:
        fmt.Printf("❌ %s: Invalid argument - %s\n", operation, st.Message())
    case codes.NotFound:
        fmt.Printf("❌ %s: Resource not found - %s\n", operation, st.Message())
    case codes.AlreadyExists:
        fmt.Printf("❌ %s: Resource already exists - %s\n", operation, st.Message())
    case codes.Unauthenticated:
        fmt.Printf("❌ %s: Authentication failed - %s\n", operation, st.Message())
    case codes.Internal:
        fmt.Printf("❌ %s: Internal server error - %s\n", operation, st.Message())
    default:
        fmt.Printf("❌ %s: Unexpected error (%s) - %s\n", operation, st.Code(), st.Message())
    }
}
```

## Validation Logic

### Email Validation

```go
func (s *UserService) isValidEmail(email string) bool {
    // Simple email validation - in production, use a proper email validation library
    if len(email) < 5 || len(email) > 254 {
        return false
    }

    // Check for @ symbol
    atIndex := -1
    for i, char := range email {
        if char == '@' {
            if atIndex != -1 {
                return false // Multiple @ symbols
            }
            atIndex = i
        }
    }

    if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
        return false
    }

    // Check for domain part
    domain := email[atIndex+1:]
    if len(domain) < 2 || len(domain) > 253 {
        return false
    }

    // Check for dot in domain
    hasDot := false
    for _, char := range domain {
        if char == '.' {
            hasDot = true
            break
        }
    }

    return hasDot
}
```

## Middleware Configuration

The server is configured with middleware in `main.go`:

```go
// Create gRPC server with middleware
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(grpcserver.UnaryServerInterceptor(errorMapper)),
    grpc.StreamInterceptor(grpcserver.StreamServerInterceptor(errorMapper)),
)
```

## Error Logging

The error mapper provides structured logging:

```go
// Log error with appropriate level based on gRPC status code
func (em *ErrorMapper) LogError(err error, operation string) {
    st, ok := status.FromError(err)
    if !ok {
        fmt.Printf("ERROR [%s]: %v\n", operation, err)
        return
    }

    switch st.Code() {
    case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists:
        fmt.Printf("INFO [%s]: %s\n", operation, st.Message())
    case codes.Unauthenticated, codes.PermissionDenied:
        fmt.Printf("WARN [%s]: %s\n", operation, st.Message())
    case codes.Internal, codes.Unavailable, codes.DataLoss:
        fmt.Printf("ERROR [%s]: %s\n", operation, st.Message())
    default:
        fmt.Printf("ERROR [%s]: %s\n", operation, st.Message())
    }
}
```

## Testing Error Handling

### Using grpcurl

```bash
# Test validation error
grpcurl -plaintext -d '{"email": "invalid-email", "username": "testuser", "password": "123"}' \
  localhost:9090 user.UserService/Register

# Test duplicate user error
grpcurl -plaintext -d '{"email": "existing@example.com", "username": "testuser", "password": "password123"}' \
  localhost:9090 user.UserService/Register

# Test authentication error
grpcurl -plaintext -d '{"email": "test@example.com", "password": "wrongpassword"}' \
  localhost:9090 user.UserService/Login
```

### Using the Client

The client demonstrates comprehensive error handling:

```bash
# Run the client demo
./client
```

## Best Practices

### 1. Centralize Validation in Service Layer

**✅ Good**: All validation logic is in the service layer
```go
// Service layer handles all validation
if err := s.validateRegisterRequest(req); err != nil {
    return nil, err
}
```

**❌ Bad**: Validation scattered across layers
```go
// Don't validate in gRPC layer
if req.Email == "" {
    return nil, status.Errorf(codes.InvalidArgument, "email required")
}
```

### 2. Always Map Domain Errors

Never return raw domain errors from gRPC services. Always use the error mapper:

```go
// ❌ Bad
return nil, errs.ErrUserNotFound

// ✅ Good
return nil, s.errorMapper.MapError(errs.ErrUserNotFound)
```

### 3. Provide Context

Use `MapErrorWithContext` when you need to add operation context:

```go
return nil, s.errorMapper.MapErrorWithContext(err, "Register")
```

### 4. Log Errors Appropriately

Use the error mapper's logging function to ensure consistent log levels:

```go
s.errorMapper.LogError(err, "Register")
```

### 5. Handle Context Cancellation

Always check for context cancellation:

```go
if err := s.errorMapper.HandleContextError(ctx, "Register"); err != nil {
    return nil, err
}
```

### 6. Keep gRPC Layer Thin

The gRPC layer should only handle:
- Protocol conversion (proto ↔ domain)
- Error mapping
- Context handling
- Logging

Business logic and validation belong in the service layer.

## Error Recovery

### Panic Recovery

The middleware automatically recovers from panics:

```go
defer func() {
    if r := recover(); r != nil {
        errorMapper.LogError(status.Errorf(codes.Internal, "panic recovered: %v", r), "PANIC: "+info.FullMethod)
    }
}()
```

### Timeout Handling

Use timeout interceptors for long-running operations:

```go
grpc.UnaryInterceptor(grpcserver.TimeoutInterceptor(30 * time.Second))
```

### Rate Limiting

Implement rate limiting for API protection:

```go
grpc.UnaryInterceptor(grpcserver.RateLimitInterceptor(100, time.Minute))
```

## Monitoring and Metrics

The middleware provides basic metrics collection:

```go
// Log metrics (in production, send to metrics system)
if err != nil {
    // metrics.IncrementCounter("grpc_errors_total", info.FullMethod)
} else {
    // metrics.RecordHistogram("grpc_duration_seconds", duration.Seconds(), info.FullMethod)
}
```

## Troubleshooting

### Common Issues

1. **Double-wrapped errors**: Ensure middleware doesn't wrap already-mapped errors
2. **Missing error mapping**: Add new domain errors to the error mapper
3. **Incorrect status codes**: Verify the mapping in `MapError` function
4. **Validation in wrong layer**: Ensure all validation is in the service layer

### Debug Mode

Enable debug logging for internal errors:

```go
if st.Code() == codes.Internal {
    fmt.Printf("   Debug: Full error details: %+v\n", err)
}
```

## Architecture Benefits

### ✅ Separation of Concerns

- **gRPC Layer**: Protocol handling, error mapping, logging
- **Service Layer**: Business logic, validation, domain operations
- **Repository Layer**: Data access and persistence

### ✅ Reusability

Service layer validation can be reused by:
- gRPC services
- HTTP handlers
- CLI commands
- Background jobs

### ✅ Testability

Service layer validation is easier to test in isolation:

```go
func TestValidateRegisterRequest(t *testing.T) {
    service := NewUserService(...)
    
    req := RegisterReq{Email: "invalid-email"}
    err := service.validateRegisterRequest(req)
    
    assert.Equal(t, errs.ErrInvalidEmail, err)
}
```

### ✅ Maintainability

- Single source of truth for validation rules
- Consistent error messages across all interfaces
- Easy to update validation logic

## Conclusion

This error handling system provides:
- ✅ **Centralized validation** in service layer
- ✅ **Consistent error responses** across all interfaces
- ✅ **Proper gRPC status codes**
- ✅ **Structured logging**
- ✅ **Panic recovery**
- ✅ **Request/response timing**
- ✅ **Comprehensive client-side handling**
- ✅ **Clean separation of concerns**

Follow these patterns to ensure robust error handling throughout your gRPC application with proper architectural separation. 