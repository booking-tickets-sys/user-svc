# gRPC Testing Guide

This document provides comprehensive information about testing the gRPC user service, including integration tests, error handling, and debugging.

## Overview

The gRPC service includes comprehensive integration tests that cover:
- User registration and login flows
- Refresh token functionality
- Token revocation
- Error handling scenarios
- Concurrent request handling

## Test Structure

### Integration Tests (`internal/app/grpc/integration_test.go`)

The integration tests use an in-memory SQLite database and a buffered gRPC server to test the complete flow without external dependencies.

#### Test Categories

1. **User Registration Tests**
   - Successful registration
   - Duplicate email/username handling
   - Invalid input validation
   - Empty field validation

2. **User Login Tests**
   - Successful login
   - Invalid credentials handling
   - Non-existent user handling

3. **Refresh Token Tests**
   - Successful token refresh
   - Single-use token validation (security feature)
   - Invalid token handling
   - Expired token handling

4. **Token Revocation Tests**
   - Individual token revocation
   - Bulk user token revocation
   - Revoked token access prevention

5. **Error Handling Tests**
   - Context cancellation
   - Invalid request data
   - Database constraint violations

6. **Concurrent Request Tests**
   - Multiple simultaneous registrations
   - Race condition testing

## Running Tests

### Quick Start

```bash
# Run all gRPC tests
./test_grpc.sh

# Run with coverage report
./test_grpc.sh --coverage
```

### Manual Test Execution

```bash
# Run specific test
go test -v ./internal/app/grpc/... -run TestUserRegistration

# Run all tests with verbose output
go test -v ./internal/app/grpc/...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./internal/app/grpc/...
go tool cover -html=coverage.out -o coverage.html
```

## Test Setup

### In-Memory Database

Tests use SQLite in-memory database for fast execution:

```go
db, err := sqlx.Connect("sqlite3", ":memory:")
```

### Buffered gRPC Server

Tests use `bufconn` for in-memory gRPC communication:

```go
lis = bufconn.Listen(bufSize)
grpcServer.Serve(lis)
```

### Test Dependencies

- **Database**: In-memory SQLite
- **Token Maker**: JWT with test secret key
- **Repositories**: Full repository implementations
- **Service**: Complete user service with all dependencies

## Error Handling and Logging

### Enhanced Error Messages

The error mapper provides detailed error information:

```go
// Before
return status.Errorf(codes.Internal, "internal server error")

// After  
return status.Errorf(codes.Internal, "%s: internal server error - %v", operation, err)
```

### Improved Logging

The middleware now provides detailed request/response logging:

```
REQUEST: /user.UserService/Register started at 14:17:50.123
SUCCESS in /user.UserService/Register (duration: 89.354833ms)
ERROR in /user.UserService/RefreshToken (duration: 10.4485ms): rpc error: code = Unauthenticated desc = RefreshToken: token revoked
```

### Error Code Mapping

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

## Common Test Scenarios

### 1. Refresh Token Security

**Expected Behavior**: Refresh tokens are single-use and get revoked after use.

```go
// First use - should succeed
resp1, err := client.RefreshToken(ctx, req)
assert.NoError(t, err)

// Second use - should fail with "token revoked"
resp2, err := client.RefreshToken(ctx, req)
assert.Error(t, err)
st, _ := status.FromError(err)
assert.Equal(t, codes.Unauthenticated, st.Code())
```

### 2. Concurrent Registration

**Expected Behavior**: Multiple simultaneous registrations should succeed.

```go
const numGoroutines = 10
results := make(chan error, numGoroutines)

for i := 0; i < numGoroutines; i++ {
    go func(id int) {
        req := &pb.RegisterRequest{
            Email:    fmt.Sprintf("concurrent%d@example.com", id),
            Username: fmt.Sprintf("concurrentuser%d", id),
            Password: "password123",
        }
        _, err := client.Register(ctx, req)
        results <- err
    }(i)
}

// All should succeed
for i := 0; i < numGoroutines; i++ {
    err := <-results
    assert.NoError(t, err)
}
```

### 3. Error Validation

**Expected Behavior**: Invalid inputs should return appropriate error codes.

```go
req := &pb.RegisterRequest{
    Email:    "invalid-email",
    Username: "validuser",
    Password: "password123",
}

resp, err := client.Register(ctx, req)
assert.Error(t, err)
st, _ := status.FromError(err)
assert.Equal(t, codes.InvalidArgument, st.Code())
```

## Debugging Tips

### 1. Enable Verbose Logging

```bash
go test -v ./internal/app/grpc/...
```

### 2. Check Error Details

```go
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        fmt.Printf("Error Code: %s\n", st.Code())
        fmt.Printf("Error Message: %s\n", st.Message())
        fmt.Printf("Error Details: %v\n", st.Details())
    }
}
```

### 3. Database State Inspection

```go
// In tests, you can inspect the database directly
var count int
err := db.Get(&count, "SELECT COUNT(*) FROM users")
assert.NoError(t, err)
fmt.Printf("User count: %d\n", count)
```

### 4. Token Validation

```go
// Verify token structure
tokenMaker := token.NewJWTTokenMaker("test-secret")
payload, err := tokenMaker.VerifyToken(accessToken)
assert.NoError(t, err)
fmt.Printf("Token payload: %+v\n", payload)
```

## Performance Considerations

### Test Execution Time

- **Individual tests**: < 100ms
- **Full test suite**: < 5 seconds
- **With coverage**: < 10 seconds

### Memory Usage

- **In-memory database**: ~1MB
- **Buffered connections**: ~1MB
- **Total test memory**: < 10MB

## Continuous Integration

### GitHub Actions Example

```yaml
name: gRPC Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go mod download
      - run: ./test_grpc.sh --coverage
      - run: go vet ./...
      - run: golangci-lint run
```

## Troubleshooting

### Common Issues

1. **"token revoked" errors**: This is expected behavior for refresh tokens
2. **Database connection errors**: Ensure SQLite is available
3. **Timeout errors**: Increase test timeout with `-timeout 60s`

### Debug Commands

```bash
# Run specific failing test
go test -v ./internal/app/grpc/... -run TestSpecificTest

# Run with race detection
go test -race ./internal/app/grpc/...

# Run with memory profiling
go test -memprofile=mem.prof ./internal/app/grpc/...
```

## Best Practices

1. **Test Isolation**: Each test uses a fresh database
2. **Cleanup**: Proper resource cleanup in test teardown
3. **Assertions**: Use `require` for setup, `assert` for validation
4. **Error Handling**: Test both success and failure scenarios
5. **Concurrency**: Test race conditions and concurrent access

## Future Enhancements

- [ ] Add performance benchmarks
- [ ] Implement stress testing
- [ ] Add integration with external services
- [ ] Create test data factories
- [ ] Add property-based testing 