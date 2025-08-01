# gRPC Refresh Token Service

This document describes the gRPC service implementation for refresh token functionality in the user service.

## Overview

The gRPC service provides a complete API for managing refresh tokens, including token refresh, revocation, and cleanup operations. All operations are exposed through the `UserService` gRPC service.

## Service Definition

### Proto Files

#### `pb/user-svc.pb.go`

Defines all the message types for refresh token operations:

```protobuf
// Refresh token request message - used for refreshing access tokens
message RefreshTokenRequest {
  string refresh_token = 1;
}

// Refresh token response message - returned after successful token refresh
message RefreshTokenResponse {
  string access_token = 1;
  string refresh_token = 2;
}

// Revoke token request message - used for revoking refresh tokens
message RevokeTokenRequest {
  string refresh_token = 1;
}

// Revoke token response message - returned after successful token revocation
message RevokeTokenResponse {
  bool success = 1;
  string message = 2;
}

// Revoke all user tokens request message - used for revoking all tokens for a user
message RevokeAllUserTokensRequest {
  string user_id = 1;
}

// Revoke all user tokens response message - returned after successful bulk token revocation
message RevokeAllUserTokensResponse {
  bool success = 1;
  string message = 2;
}

// Cleanup expired tokens request message - used for cleaning up expired tokens
message CleanupExpiredTokensRequest {
  // Empty request - no parameters needed
}

// Cleanup expired tokens response message - returned after successful cleanup
message CleanupExpiredTokensResponse {
  bool success = 1;
  string message = 2;
  int32 tokens_removed = 3;
}
```

#### `pb/user-svc_grpc.pb.go`

Defines the gRPC service interface:

```protobuf
service UserService {
  // Register creates a new user account
  // Returns user information, access token, and refresh token on success
  rpc Register(RegisterRequest) returns (RegisterResponse);
  
  // Login authenticates an existing user
  // Returns user information, access token, and refresh token on success
  rpc Login(LoginRequest) returns (LoginResponse);
  
  // RefreshToken exchanges a refresh token for a new access token and refresh token pair
  // Returns new access token and refresh token on success
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  
  // RevokeToken revokes a specific refresh token
  // Returns success status and message
  rpc RevokeToken(RevokeTokenRequest) returns (RevokeTokenResponse);
  
  // RevokeAllUserTokens revokes all refresh tokens for a specific user
  // Returns success status and message
  rpc RevokeAllUserTokens(RevokeAllUserTokensRequest) returns (RevokeAllUserTokensResponse);
  
  // CleanupExpiredTokens removes expired refresh tokens from the database
  // Returns success status, message, and number of tokens removed
  rpc CleanupExpiredTokens(CleanupExpiredTokensRequest) returns (CleanupExpiredTokensResponse);
}
```

## Server Implementation

### UserServer (`internal/app/grpc/server.go`)

The gRPC server implementation provides the following methods:

#### RefreshToken

```go
func (s *UserServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error)
```

**Purpose**: Exchanges a refresh token for a new access token and refresh token pair.

**Request**:
- `refresh_token`: The refresh token to exchange

**Response**:
- `access_token`: New access token
- `refresh_token`: New refresh token

**Error Handling**:
- `InvalidArgument`: Invalid or malformed refresh token
- `Unauthenticated`: Expired or revoked refresh token
- `NotFound`: Refresh token not found in database

#### RevokeToken

```go
func (s *UserServer) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error)
```

**Purpose**: Revokes a specific refresh token.

**Request**:
- `refresh_token`: The refresh token to revoke

**Response**:
- `success`: Boolean indicating success
- `message`: Success or error message

**Error Handling**:
- `InvalidArgument`: Invalid refresh token
- `NotFound`: Refresh token not found

#### RevokeAllUserTokens

```go
func (s *UserServer) RevokeAllUserTokens(ctx context.Context, req *pb.RevokeAllUserTokensRequest) (*pb.RevokeAllUserTokensResponse, error)
```

**Purpose**: Revokes all refresh tokens for a specific user.

**Request**:
- `user_id`: The user ID whose tokens should be revoked

**Response**:
- `success`: Boolean indicating success
- `message`: Success or error message

**Error Handling**:
- `NotFound`: User not found
- `InvalidArgument`: Invalid user ID

#### CleanupExpiredTokens

```go
func (s *UserServer) CleanupExpiredTokens(ctx context.Context, req *pb.CleanupExpiredTokensRequest) (*pb.CleanupExpiredTokensResponse, error)
```

**Purpose**: Removes expired refresh tokens from the database.

**Request**: Empty request (no parameters needed)

**Response**:
- `success`: Boolean indicating success
- `message`: Success or error message
- `tokens_removed`: Number of tokens removed (currently always 0)

**Error Handling**:
- `Internal`: Database error during cleanup

## Error Mapping

The gRPC server uses an `ErrorMapper` to convert domain errors to appropriate gRPC status codes:

| Domain Error | gRPC Status Code | Description |
|--------------|------------------|-------------|
| `ErrInvalidToken` | `InvalidArgument` | Invalid or malformed token |
| `ErrTokenExpired` | `Unauthenticated` | Token has expired |
| `ErrTokenRevoked` | `Unauthenticated` | Token has been revoked |
| `ErrTokenNotFound` | `NotFound` | Token not found in database |
| `ErrUserNotFound` | `NotFound` | User not found |

## Usage Examples

### Client Implementation

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "user-svc/pb"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewUserServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Refresh token
    refreshResp, err := client.RefreshToken(ctx, &pb.RefreshTokenRequest{
        RefreshToken: "your-refresh-token",
    })
    if err != nil {
        log.Fatalf("Failed to refresh token: %v", err)
    }

    log.Printf("New access token: %s", refreshResp.AccessToken)
    log.Printf("New refresh token: %s", refreshResp.RefreshToken)

    // Revoke token
    revokeResp, err := client.RevokeToken(ctx, &pb.RevokeTokenRequest{
        RefreshToken: "token-to-revoke",
    })
    if err != nil {
        log.Fatalf("Failed to revoke token: %v", err)
    }

    log.Printf("Revoke result: %s", revokeResp.Message)

    // Revoke all user tokens
    revokeAllResp, err := client.RevokeAllUserTokens(ctx, &pb.RevokeAllUserTokensRequest{
        UserId: "user-id",
    })
    if err != nil {
        log.Fatalf("Failed to revoke all tokens: %v", err)
    }

    log.Printf("Revoke all result: %s", revokeAllResp.Message)

    // Cleanup expired tokens
    cleanupResp, err := client.CleanupExpiredTokens(ctx, &pb.CleanupExpiredTokensRequest{})
    if err != nil {
        log.Fatalf("Failed to cleanup tokens: %v", err)
    }

    log.Printf("Cleanup result: %s", cleanupResp.Message)
}
```

### Server Setup

```go
package main

import (
    "log"
    "net"

    "google.golang.org/grpc"
    "user-svc/internal/app/grpc"
    "user-svc/internal/app/service"
    "user-svc/pb"
)

func main() {
    // Initialize your service dependencies
    userService := service.NewUserService(/* dependencies */)
    
    // Create gRPC server
    grpcServer := grpc.NewServer()
    userServer := grpc.NewUserServer(userService)
    
    // Register service
    pb.RegisterUserServiceServer(grpcServer, userServer)
    
    // Start server
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    log.Printf("Server listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

## Testing

The gRPC service includes comprehensive tests:

### Integration Tests

```go
func TestRefreshTokenProtoMessages(t *testing.T) {
    // Tests that protobuf messages are correctly defined
}

func TestServiceRequestTypes(t *testing.T) {
    // Tests that service request types match proto messages
}
```

### Running Tests

```bash
# Run gRPC tests
go test ./internal/app/grpc -v

# Run all tests
go test ./... -v
```

## Security Considerations

1. **Token Validation**: All refresh tokens are validated for expiration and revocation status
2. **Error Handling**: Sensitive information is not exposed in error messages
3. **Context Handling**: Proper context cancellation is handled for all operations
4. **Input Validation**: All input parameters are validated before processing

## Performance Considerations

1. **Database Indexes**: Refresh token lookups are optimized with database indexes
2. **Connection Pooling**: gRPC connections are pooled for better performance
3. **Error Mapping**: Efficient error mapping without unnecessary allocations
4. **Context Propagation**: Proper context propagation for tracing and monitoring

## Monitoring and Observability

The gRPC server includes:

1. **Error Logging**: All errors are logged with appropriate context
2. **Context Handling**: Proper context cancellation and timeout handling
3. **Status Codes**: Standard gRPC status codes for monitoring
4. **Request Tracing**: Context propagation for distributed tracing

## Best Practices

1. **Use HTTPS**: Always use TLS in production
2. **Implement Retry Logic**: Handle transient failures in clients
3. **Monitor Token Usage**: Track refresh patterns for security analysis
4. **Regular Cleanup**: Schedule regular cleanup of expired tokens
5. **Rate Limiting**: Implement rate limiting for token operations
6. **Audit Logging**: Log all token operations for security auditing 