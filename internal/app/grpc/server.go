package grpc

import (
	"context"
	"time"

	"user-svc/internal/app/service"
	"user-svc/internal/domain/dto"
	"user-svc/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
	errorMapper *ErrorMapper
}

func NewUserServiceServer(userService *service.UserService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
		errorMapper: NewErrorMapper(),
	}
}

// Register handles user registration
func (s *UserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "Register"); err != nil {
		return nil, err
	}

	// Call service
	result, err := s.userService.Register(ctx, dto.RegisterReq{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "Register")
	}

	// Map response
	user := &pb.User{
		Id:        result.User.ID.String(),
		Email:     result.User.Email.String(),
		Username:  result.User.Username.String(),
		CreatedAt: timestamppb.New(time.UnixMilli(result.User.CreatedAt)),
		UpdatedAt: timestamppb.New(time.UnixMilli(result.User.UpdatedAt)),
	}

	return &pb.RegisterResponse{
		User:         user,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

// Login handles user login
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "Login"); err != nil {
		return nil, err
	}

	// Call service
	result, err := s.userService.Login(ctx, dto.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "Login")
	}

	// Map response
	user := &pb.User{
		Id:        result.User.ID.String(),
		Email:     result.User.Email.String(),
		Username:  result.User.Username.String(),
		CreatedAt: timestamppb.New(time.UnixMilli(result.User.CreatedAt)),
		UpdatedAt: timestamppb.New(time.UnixMilli(result.User.UpdatedAt)),
	}

	return &pb.LoginResponse{
		User:         user,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *UserServiceServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "RefreshToken"); err != nil {
		return nil, err
	}

	// Call service
	result, err := s.userService.RefreshToken(ctx, dto.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "RefreshToken")
	}

	return &pb.RefreshTokenResponse{
		AccessToken: result.AccessToken,
	}, nil
}

func (s *UserServiceServer) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "RevokeToken"); err != nil {
		return nil, err
	}

	// Call service
	err := s.userService.RevokeToken(ctx, dto.RevokeTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "RevokeToken")
	}

	return &pb.RevokeTokenResponse{
		Success: true,
		Message: "Token revoked successfully",
	}, nil
}

func (s *UserServiceServer) RevokeAllUserTokens(ctx context.Context, req *pb.RevokeAllUserTokensRequest) (*pb.RevokeAllUserTokensResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "RevokeAllUserTokens"); err != nil {
		return nil, err
	}

	// Parse user ID from string to uuid.UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	// Call service
	err = s.userService.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "RevokeAllUserTokens")
	}

	return &pb.RevokeAllUserTokensResponse{
		Success: true,
		Message: "All user tokens revoked successfully",
	}, nil
}

func (s *UserServiceServer) CleanupExpiredTokens(ctx context.Context, req *pb.CleanupExpiredTokensRequest) (*pb.CleanupExpiredTokensResponse, error) {
	// Check context error first
	if err := s.errorMapper.HandleContextError(ctx, "CleanupExpiredTokens"); err != nil {
		return nil, err
	}

	// Call service
	err := s.userService.CleanupExpiredTokens(ctx)
	if err != nil {
		return nil, s.errorMapper.MapErrorWithContext(err, "CleanupExpiredTokens")
	}

	return &pb.CleanupExpiredTokensResponse{
		Success:       true,
		Message:       "Expired tokens cleaned up successfully",
		TokensRemoved: 0, // Note: The service doesn't return the count, so we set it to 0
	}, nil
}
