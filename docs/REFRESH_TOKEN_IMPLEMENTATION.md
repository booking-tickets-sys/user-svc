# Refresh Token Implementation

This document describes the implementation of session refresh token functionality in the user service.

## Overview

The refresh token system provides a secure way to maintain user sessions by allowing clients to obtain new access tokens without requiring re-authentication. This implementation includes:

- Database storage of refresh tokens with proper hashing
- Token validation and expiration handling
- Token revocation capabilities
- Automatic cleanup of expired tokens

## Architecture

### Domain Layer

#### RefreshToken Model (`internal/domain/refresh_token.go`)

The `RefreshToken` domain model includes:

- **ID**: Unique identifier for the refresh token
- **UserID**: Reference to the user who owns the token
- **TokenHash**: SHA-256 hash of the actual token (for secure storage)
- **ExpiresAt**: Unix timestamp when the token expires
- **IsRevoked**: Boolean flag indicating if the token has been revoked
- **CreatedAt/UpdatedAt**: Timestamps for audit purposes

Key methods:
- `NewRefreshToken()`: Creates a new refresh token with validation
- `IsValid()`: Validates the token (not expired, not revoked, etc.)
- `IsExpired()`: Checks if the token has expired
- `Revoke()`: Marks the token as revoked

### Database Layer

#### Migration (`db/migrations/000002_create_refresh_tokens_table.up.sql`)

Creates the `refresh_tokens` table with:

- UUID primary key
- Foreign key relationship to users table
- Indexes for performance optimization
- Automatic timestamp updates via triggers

#### Repository (`internal/app/repository/refresh_token.go`)

The `RefreshTokenRepository` provides:

- **CRUD Operations**: Create, read, update, delete refresh tokens
- **Token Lookup**: Find tokens by hash, user ID, or token ID
- **Bulk Operations**: Revoke all tokens for a user, delete expired tokens
- **Security**: Stores only hashed tokens, never plain text

### Service Layer

#### User Service (`internal/app/service/user.go`)

Enhanced with refresh token functionality:

- **Registration**: Creates both access and refresh tokens
- **Login**: Issues new access and refresh tokens
- **Token Refresh**: Exchanges refresh token for new access/refresh token pair
- **Token Revocation**: Allows users to revoke specific or all tokens
- **Cleanup**: Removes expired tokens from database

#### Token Utilities (`internal/app/service/token_utils.go`)

Utility functions for token security:

- `HashToken()`: Creates SHA-256 hash of tokens
- `ValidateTokenHash()`: Validates token against its hash
- `GenerateTokenHash()`: Generates both token and hash

### Token Layer

#### Token Maker Interface (`token/maker.go`)

Extended to support refresh tokens:

- `CreateRefreshToken()`: Creates refresh tokens
- `VerifyRefreshToken()`: Verifies refresh tokens

Both JWT and PASETO implementations support these methods.

## Security Features

### Token Hashing
- Refresh tokens are hashed using SHA-256 before storage
- Only hashes are stored in the database, never plain text tokens
- Token validation compares hashes, not plain text

### Token Rotation
- Each refresh operation generates a new refresh token
- Old refresh tokens are automatically revoked
- Prevents token reuse and reduces attack surface

### Expiration Management
- Refresh tokens have configurable expiration times
- Expired tokens are automatically invalidated
- Cleanup process removes expired tokens from database

### Revocation Support
- Individual tokens can be revoked
- All tokens for a user can be revoked at once
- Revoked tokens are permanently invalid

## API Endpoints

The service provides the following refresh token operations:

### Refresh Token
```go
type RefreshTokenReq struct {
    RefreshToken string
}

type RefreshTokenResp struct {
    AccessToken  string
    RefreshToken string
}

func (s *UserService) RefreshToken(ctx context.Context, req RefreshTokenReq) (*RefreshTokenResp, error)
```

### Revoke Token
```go
type RevokeTokenReq struct {
    RefreshToken string
}

func (s *UserService) RevokeToken(ctx context.Context, req RevokeTokenReq) error
```

### Revoke All User Tokens
```go
func (s *UserService) RevokeAllUserTokens(ctx context.Context, userID string) error
```

### Cleanup Expired Tokens
```go
func (s *UserService) CleanupExpiredTokens(ctx context.Context) error
```

## Configuration

The service accepts configuration for token durations:

- **Access Token Duration**: Default 15 minutes
- **Refresh Token Duration**: Default 7 days

These can be configured when creating the service:

```go
service := NewUserService(
    userRepo,
    refreshTokenRepo,
    tokenMaker,
    15*time.Minute,    // Access token duration
    7*24*time.Hour,    // Refresh token duration
    txManager,
)
```

## Testing

Comprehensive test coverage includes:

### Domain Tests
- Token creation and validation
- Expiration handling
- Revocation functionality

### Service Tests
- Token refresh flow
- Error handling
- Token revocation
- Cleanup operations

### Utility Tests
- Token hashing
- Hash validation
- Token generation

## Usage Example

```go
// Register a new user (returns access + refresh tokens)
resp, err := service.Register(ctx, RegisterReq{
    Email:    "user@example.com",
    Username: "username",
    Password: "password123",
})

// Later, refresh the access token
refreshResp, err := service.RefreshToken(ctx, RefreshTokenReq{
    RefreshToken: resp.RefreshToken,
})

// Use new access token
accessToken := refreshResp.AccessToken

// Revoke token when user logs out
err = service.RevokeToken(ctx, RevokeTokenReq{
    RefreshToken: refreshResp.RefreshToken,
})
```

## Best Practices

1. **Store refresh tokens securely**: Use HTTP-only cookies or secure storage
2. **Implement token rotation**: Always issue new refresh tokens on refresh
3. **Set appropriate expiration times**: Balance security with user experience
4. **Monitor token usage**: Track refresh patterns for security analysis
5. **Implement cleanup**: Regularly remove expired tokens from database
6. **Handle revocation**: Provide logout functionality that revokes tokens

## Error Handling

The implementation includes comprehensive error handling:

- `ErrInvalidToken`: Invalid or malformed tokens
- `ErrTokenExpired`: Expired tokens
- `ErrTokenRevoked`: Revoked tokens
- `ErrTokenNotFound`: Tokens not found in database
- `ErrUserNotFound`: Associated user not found

## Performance Considerations

- Database indexes optimize token lookups
- Token hashing uses efficient SHA-256
- Cleanup process can be run asynchronously
- Connection pooling for database operations 