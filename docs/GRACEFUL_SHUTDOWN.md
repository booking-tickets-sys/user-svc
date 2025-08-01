# Graceful Shutdown Implementation

This document describes the graceful shutdown implementation in the user-svc gRPC server.

## Overview

The server implements graceful shutdown to ensure that:
- In-flight requests are completed before shutdown
- Database connections are properly closed
- Resources are cleaned up in the correct order
- The server responds to OS signals (SIGINT, SIGTERM)

## How It Works

### Signal Handling
The server listens for OS signals:
- `SIGINT` (Ctrl+C)
- `SIGTERM` (termination signal)

### Shutdown Process
1. **Signal Reception**: When a shutdown signal is received, the server logs the event
2. **gRPC Server Graceful Stop**: Calls `grpcServer.GracefulStop()` which:
   - Stops accepting new connections
   - Allows existing requests to complete
   - Waits for all active streams to finish
3. **Database Connection Cleanup**: Properly closes the database connection
4. **Resource Cleanup**: Any additional cleanup tasks

### Configuration

The graceful shutdown timeout can be configured in `config.yaml`:

```yaml
server:
  grpc:
    port: 9090
    host: "localhost"
    graceful_shutdown_timeout: 30s  # Default: 30s if not specified
```

### Timeout Behavior

- The server will wait up to the configured timeout for graceful shutdown
- If the timeout is reached, the server will force shutdown
- Default timeout is 30 seconds if not configured

## Usage

### Starting the Server
```bash
go run main.go
# or
./user-svc
```

### Stopping the Server
```bash
# Send SIGINT (Ctrl+C)
# or
kill -TERM <pid>
```

### Expected Output
```
gRPC server listening on localhost:9090
Server configuration:
  - Database: localhost:5432
  - Token Duration: 15m0s
  - Reflection: enabled
^CReceived signal interrupt, initiating graceful shutdown...
Stopping gRPC server...
gRPC server stopped
Closing database connection...
Database connection closed
Graceful shutdown completed
```

## Benefits

1. **Data Integrity**: Ensures no data loss during shutdown
2. **Client Experience**: Clients can complete their requests
3. **Resource Management**: Proper cleanup prevents resource leaks
4. **Container Compatibility**: Works well with container orchestration systems
5. **Monitoring**: Clear logging of shutdown process

## Implementation Details

The implementation uses:
- `os/signal` package for signal handling
- `context.WithTimeout` for timeout management
- `grpc.GracefulStop()` for graceful gRPC server shutdown
- Proper error handling and logging throughout the process 