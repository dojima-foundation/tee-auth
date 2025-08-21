package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server represents the gRPC server
type Server struct {
	pb.UnimplementedGAuthServiceServer
	config  *config.Config
	logger  *logger.Logger
	service *service.GAuthService
	server  *grpc.Server
}

// NewServer creates a new gRPC server instance
func NewServer(cfg *config.Config, logger *logger.Logger, svc *service.GAuthService) *Server {
	return &Server{
		config:  cfg,
		logger:  logger,
		service: svc,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	addr := s.config.GetGRPCAddr()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Configure gRPC server options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(s.config.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(s.config.GRPC.MaxSendMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    s.config.GRPC.KeepAliveTime,
			Timeout: s.config.GRPC.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             s.config.GRPC.KeepAliveTime,
			PermitWithoutStream: s.config.GRPC.PermitWithoutStream,
		}),
		grpc.UnaryInterceptor(s.unaryInterceptor),
	}

	// Create gRPC server
	s.server = grpc.NewServer(opts...)

	// Register service
	pb.RegisterGAuthServiceServer(s.server, s)

	// Enable reflection for development
	reflection.Register(s.server)

	s.logger.Info("Starting gRPC server", "address", addr)

	// Start serving
	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.server != nil {
		s.logger.Info("Stopping gRPC server")
		s.server.GracefulStop()
	}
}

// Unary interceptor for logging and error handling
func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the handler
	resp, err := handler(ctx, req)

	// Log the request
	duration := time.Since(start)
	s.logger.Info("gRPC request",
		"method", info.FullMethod,
		"duration", duration,
		"error", err,
	)

	return resp, err
}

// Organization management methods

func (s *Server) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	org, err := s.service.CreateOrganization(ctx, req.Name, req.InitialUserEmail, req.InitialUserPublicKey)
	if err != nil {
		s.logger.Error("Failed to create organization", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create organization: %v", err)
	}

	return &pb.CreateOrganizationResponse{
		Organization: convertOrganizationToProto(org),
		Status:       "created",
	}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
	org, err := s.service.GetOrganization(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get organization", "error", err, "id", req.Id)
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}

	return &pb.GetOrganizationResponse{
		Organization: convertOrganizationToProto(org),
	}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.UpdateOrganizationResponse, error) {
	org, err := s.service.UpdateOrganization(ctx, req.Id, req.Name)
	if err != nil {
		s.logger.Error("Failed to update organization", "error", err, "id", req.Id)
		return nil, status.Errorf(codes.Internal, "failed to update organization: %v", err)
	}

	return &pb.UpdateOrganizationResponse{
		Organization: convertOrganizationToProto(org),
	}, nil
}

func (s *Server) ListOrganizations(ctx context.Context, req *pb.ListOrganizationsRequest) (*pb.ListOrganizationsResponse, error) {
	orgs, nextToken, err := s.service.ListOrganizations(ctx, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("Failed to list organizations", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list organizations: %v", err)
	}

	protoOrgs := make([]*pb.Organization, len(orgs))
	for i, org := range orgs {
		protoOrgs[i] = convertOrganizationToProto(&org)
	}

	return &pb.ListOrganizationsResponse{
		Organizations: protoOrgs,
		NextPageToken: nextToken,
	}, nil
}

// User management methods

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := s.service.CreateUser(ctx, req.OrganizationId, req.Username, req.Email, req.PublicKey, req.Tags)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.CreateUserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.service.GetUser(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err, "id", req.Id)
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err := s.service.UpdateUser(ctx, req.Id, req.Username, req.Email, req.Tags, req.IsActive)
	if err != nil {
		s.logger.Error("Failed to update user", "error", err, "id", req.Id)
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UpdateUserResponse{
		User: convertUserToProto(user),
	}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, nextToken, err := s.service.ListUsers(ctx, req.OrganizationId, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("Failed to list users", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = convertUserToProto(&user)
	}

	return &pb.ListUsersResponse{
		Users:         protoUsers,
		NextPageToken: nextToken,
	}, nil
}

// Activity management methods

func (s *Server) CreateActivity(ctx context.Context, req *pb.CreateActivityRequest) (*pb.CreateActivityResponse, error) {
	activity, err := s.service.CreateActivity(ctx, req.OrganizationId, req.Type, req.Parameters, req.CreatedBy)
	if err != nil {
		s.logger.Error("Failed to create activity", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create activity: %v", err)
	}

	return &pb.CreateActivityResponse{
		Activity: convertActivityToProto(activity),
	}, nil
}

func (s *Server) GetActivity(ctx context.Context, req *pb.GetActivityRequest) (*pb.GetActivityResponse, error) {
	activity, err := s.service.GetActivity(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get activity", "error", err, "id", req.Id)
		return nil, status.Errorf(codes.NotFound, "activity not found: %v", err)
	}

	return &pb.GetActivityResponse{
		Activity: convertActivityToProto(activity),
	}, nil
}

func (s *Server) ListActivities(ctx context.Context, req *pb.ListActivitiesRequest) (*pb.ListActivitiesResponse, error) {
	activities, nextToken, err := s.service.ListActivities(ctx, req.OrganizationId, req.Type, req.Status, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("Failed to list activities", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list activities: %v", err)
	}

	protoActivities := make([]*pb.Activity, len(activities))
	for i, activity := range activities {
		protoActivities[i] = convertActivityToProto(&activity)
	}

	return &pb.ListActivitiesResponse{
		Activities:    protoActivities,
		NextPageToken: nextToken,
	}, nil
}

// Communication with renclave-v2

func (s *Server) RequestSeedGeneration(ctx context.Context, req *pb.SeedGenerationRequest) (*pb.SeedGenerationResponse, error) {
	response, err := s.service.RequestSeedGeneration(ctx, req.OrganizationId, req.UserId, int(req.Strength), req.Passphrase)
	if err != nil {
		s.logger.Error("Failed to request seed generation", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to generate seed: %v", err)
	}

	return &pb.SeedGenerationResponse{
		SeedPhrase: response.SeedPhrase,
		Entropy:    response.Entropy,
		Strength:   int32(response.Strength),
		WordCount:  int32(response.WordCount),
		RequestId:  response.RequestID,
	}, nil
}

func (s *Server) ValidateSeed(ctx context.Context, req *pb.SeedValidationRequest) (*pb.SeedValidationResponse, error) {
	response, err := s.service.ValidateSeed(ctx, req.SeedPhrase)
	if err != nil {
		s.logger.Error("Failed to validate seed", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to validate seed: %v", err)
	}

	return &pb.SeedValidationResponse{
		IsValid:   response.IsValid,
		Strength:  int32(response.Strength),
		WordCount: int32(response.WordCount),
		Errors:    response.Errors,
	}, nil
}

func (s *Server) GetEnclaveInfo(ctx context.Context, req *emptypb.Empty) (*pb.EnclaveInfoResponse, error) {
	info, err := s.service.GetEnclaveInfo(ctx)
	if err != nil {
		s.logger.Error("Failed to get enclave info", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get enclave info: %v", err)
	}

	return &pb.EnclaveInfoResponse{
		Version:      info.Version,
		EnclaveId:    info.EnclaveID,
		Capabilities: info.Capabilities,
		Healthy:      info.Healthy,
	}, nil
}

// Authentication and authorization

func (s *Server) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	response, err := s.service.Authenticate(ctx, req.OrganizationId, req.UserId, req.AuthMethodId, req.Signature, req.Timestamp)
	if err != nil {
		s.logger.Error("Authentication failed", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
	}

	return &pb.AuthenticateResponse{
		Authenticated: response.Authenticated,
		SessionToken:  response.SessionToken,
		ExpiresAt:     timestamppb.New(response.ExpiresAt),
		User:          convertUserToProto(response.User),
	}, nil
}

func (s *Server) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	response, err := s.service.Authorize(ctx, req.SessionToken, req.ActivityType, req.Parameters)
	if err != nil {
		s.logger.Error("Authorization failed", "error", err)
		return nil, status.Errorf(codes.PermissionDenied, "authorization failed: %v", err)
	}

	return &pb.AuthorizeResponse{
		Authorized:        response.Authorized,
		Reason:            response.Reason,
		RequiredApprovals: response.RequiredApprovals,
	}, nil
}

// Health and status

func (s *Server) Health(ctx context.Context, req *emptypb.Empty) (*pb.HealthResponse, error) {
	health, err := s.service.Health(ctx)
	if err != nil {
		s.logger.Error("Health check failed", "error", err)
		return nil, status.Errorf(codes.Internal, "health check failed: %v", err)
	}

	services := make([]*pb.ServiceStatus, len(health.Services))
	for i, svc := range health.Services {
		services[i] = &pb.ServiceStatus{
			Name:   svc.Name,
			Status: svc.Status,
			Error:  svc.Error,
		}
	}

	return &pb.HealthResponse{
		Status:    health.Status,
		Services:  services,
		Timestamp: timestamppb.New(health.Timestamp),
	}, nil
}

func (s *Server) Status(ctx context.Context, req *emptypb.Empty) (*pb.StatusResponse, error) {
	statusResp, err := s.service.Status(ctx)
	if err != nil {
		s.logger.Error("Status check failed", "error", err)
		return nil, status.Errorf(codes.Internal, "status check failed: %v", err)
	}

	return &pb.StatusResponse{
		Version:   statusResp.Version,
		BuildTime: statusResp.BuildTime,
		GitCommit: statusResp.GitCommit,
		Uptime:    timestamppb.New(statusResp.Uptime),
		Metrics:   statusResp.Metrics,
	}, nil
}
