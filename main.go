package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-svc/config"
	database "user-svc/db"
	grpcserver "user-svc/internal/app/grpc"
	"user-svc/internal/app/repository"
	"user-svc/internal/app/service"
	"user-svc/pb"
	"user-svc/token"
	"user-svc/utils/tx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbConn, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(dbConn.DB)

	// Initialize token maker
	tokenMaker := token.NewJWTTokenMaker(cfg.Security.JWT.SecretKey)

	txManager := tx.NewTransactionManager(dbConn.DB)
	// Initialize services
	userService := service.NewUserService(userRepo, refreshTokenRepo, tokenMaker, cfg.Security.JWT.TokenDuration, 7*24*time.Hour, txManager)

	// Initialize gRPC server with error handling
	userServer := grpcserver.NewUserServiceServer(userService)

	// Create error mapper for middleware
	errorMapper := grpcserver.NewErrorMapper()

	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.UnaryServerInterceptor(errorMapper)),
		grpc.StreamInterceptor(grpcserver.StreamServerInterceptor(errorMapper)),
	)

	// Register services
	pb.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s", grpcAddr)
	log.Printf("Server configuration:")
	log.Printf("  - Database: %s:%d", cfg.Database.Host, cfg.Database.Port)
	log.Printf("  - Token Duration: %v", cfg.Security.JWT.TokenDuration)
	log.Printf("  - Reflection: enabled")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Create a context with timeout for graceful shutdown
	shutdownTimeout := cfg.Server.GRPC.GracefulShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second // default timeout
	}
	_, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Gracefully stop the gRPC server
	log.Println("Stopping gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped")

	// Close database connection
	log.Println("Closing database connection...")
	if err := dbConn.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	} else {
		log.Println("Database connection closed")
	}

	log.Println("Graceful shutdown completed")
}
