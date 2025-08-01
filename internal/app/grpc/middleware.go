package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor provides error handling and logging for unary RPCs
func UnaryServerInterceptor(errorMapper *ErrorMapper) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Log request with method name
		fmt.Printf("REQUEST: %s started at %s\n", info.FullMethod, start.Format("15:04:05.000"))

		// Handle panic recovery
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("PANIC in %s: %v\n", info.FullMethod, r)
				errorMapper.LogError(status.Errorf(codes.Internal, "panic recovered: %v", r), "PANIC: "+info.FullMethod)
			}
		}()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Handle response
		if err != nil {
			// Log the error with detailed information
			fmt.Printf("ERROR in %s (duration: %s): %v\n", info.FullMethod, duration, err)
			errorMapper.LogError(err, "ERROR: "+info.FullMethod)
			// Map the error to gRPC status code only if it's not already a gRPC status
			if _, ok := status.FromError(err); ok {
				// Error is already a gRPC status, return it as is
				return nil, err
			}
			// Map the error to gRPC status code
			return nil, errorMapper.MapError(err)
		}

		// Log successful response with duration
		fmt.Printf("SUCCESS in %s (duration: %s)\n", info.FullMethod, duration)
		errorMapper.LogError(nil, "SUCCESS: "+info.FullMethod)

		return resp, nil
	}
}

// StreamServerInterceptor provides error handling and logging for streaming RPCs
func StreamServerInterceptor(errorMapper *ErrorMapper) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Log request with method name
		fmt.Printf("STREAM_REQUEST: %s started at %s\n", info.FullMethod, start.Format("15:04:05.000"))

		// Handle panic recovery
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("STREAM_PANIC in %s: %v\n", info.FullMethod, r)
				errorMapper.LogError(status.Errorf(codes.Internal, "panic recovered: %v", r), "STREAM_PANIC: "+info.FullMethod)
			}
		}()

		// Call the handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start)

		// Handle response
		if err != nil {
			// Log the error with detailed information
			fmt.Printf("STREAM_ERROR in %s (duration: %s): %v\n", info.FullMethod, duration, err)
			errorMapper.LogError(err, "STREAM_ERROR: "+info.FullMethod)
			// Map the error to gRPC status code
			return errorMapper.MapError(err)
		}

		// Log successful response with duration
		fmt.Printf("STREAM_SUCCESS in %s (duration: %s)\n", info.FullMethod, duration)
		errorMapper.LogError(nil, "STREAM_SUCCESS: "+info.FullMethod)

		return nil
	}
}

// RateLimitInterceptor provides basic rate limiting
func RateLimitInterceptor(maxRequests int, window time.Duration) grpc.UnaryServerInterceptor {
	// Simple in-memory rate limiter - in production, use Redis or similar
	requestCounts := make(map[string]int)
	lastReset := time.Now()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		now := time.Now()

		// Reset counters if window has passed
		if now.Sub(lastReset) > window {
			requestCounts = make(map[string]int)
			lastReset = now
		}

		// Check rate limit
		if requestCounts[info.FullMethod] >= maxRequests {
			fmt.Printf("RATE_LIMIT exceeded for %s\n", info.FullMethod)
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded for %s", info.FullMethod)
		}

		// Increment counter
		requestCounts[info.FullMethod]++

		// Call the handler
		return handler(ctx, req)
	}
}

// TimeoutInterceptor adds timeout to RPC calls
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}

// ValidationInterceptor provides request validation
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add validation logic here if needed
		// For now, just pass through
		return handler(ctx, req)
	}
}

// MetricsInterceptor provides basic metrics collection
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate metrics
		duration := time.Since(start)

		// Log metrics (in production, send to metrics system)
		if err != nil {
			// Log error metrics
			fmt.Printf("METRICS: %s failed after %s\n", info.FullMethod, duration)
			// metrics.IncrementCounter("grpc_errors_total", info.FullMethod)
		} else {
			// Log success metrics
			fmt.Printf("METRICS: %s succeeded in %s\n", info.FullMethod, duration)
			// metrics.RecordHistogram("grpc_duration_seconds", duration.Seconds(), info.FullMethod)
		}

		return resp, err
	}
}
