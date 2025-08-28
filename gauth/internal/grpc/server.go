package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"

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
	config    *config.Config
	logger    *logger.Logger
	service   *service.GAuthService
	server    *grpc.Server
	telemetry *telemetry.Telemetry
}

// NewServer creates a new gRPC server instance
func NewServer(cfg *config.Config, logger *logger.Logger, svc *service.GAuthService, tel *telemetry.Telemetry) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		service:   svc,
		telemetry: tel,
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
	}

	// Add OpenTelemetry middleware if telemetry is available and tracing is enabled
	if s.telemetry != nil && s.config.Telemetry.TracingEnabled {
		// Add OpenTelemetry interceptors
		opts = append(opts,
			grpc.UnaryInterceptor(telemetry.GRPCUnaryServerInterceptor(s.telemetry, s.logger)),
			grpc.StreamInterceptor(telemetry.GRPCStreamServerInterceptor(s.telemetry, s.logger)),
		)
	} else {
		// Fall back to basic logging interceptor
		opts = append(opts, grpc.UnaryInterceptor(s.unaryInterceptor))
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

// Wallet management methods

func (s *Server) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	s.logger.Info("CreateWallet gRPC request received", "organization_id", req.OrganizationId, "name", req.Name)

	// Convert protobuf accounts to models
	accounts := make([]models.WalletAccount, len(req.Accounts))
	for i, acc := range req.Accounts {
		accounts[i] = models.WalletAccount{
			Curve:         acc.Curve,
			Path:          acc.Path,
			AddressFormat: acc.AddressFormat,
			// PathFormat is handled internally - we store the path but not the format
		}
	}

	wallet, addresses, err := s.service.CreateWallet(ctx, req.OrganizationId, req.Name, accounts, req.MnemonicLength, req.Tags)
	if err != nil {
		s.logger.Error("Failed to create wallet", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	// Convert to protobuf
	pbWallet := convertWalletToProto(wallet)

	return &pb.CreateWalletResponse{
		Wallet:    pbWallet,
		Addresses: addresses,
	}, nil
}

func (s *Server) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	s.logger.Debug("GetWallet gRPC request received", "wallet_id", req.Id)

	wallet, err := s.service.GetWallet(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get wallet", "error", err, "wallet_id", req.Id)
		return nil, status.Errorf(codes.NotFound, "wallet not found: %v", err)
	}

	return &pb.GetWalletResponse{
		Wallet: convertWalletToProto(wallet),
	}, nil
}

func (s *Server) ListWallets(ctx context.Context, req *pb.ListWalletsRequest) (*pb.ListWalletsResponse, error) {
	s.logger.Debug("ListWallets gRPC request received", "organization_id", req.OrganizationId)

	wallets, nextToken, err := s.service.ListWallets(ctx, req.OrganizationId, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("Failed to list wallets", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list wallets: %v", err)
	}

	// Convert to protobuf
	pbWallets := make([]*pb.Wallet, len(wallets))
	for i, wallet := range wallets {
		pbWallets[i] = convertWalletToProto(&wallet)
	}

	return &pb.ListWalletsResponse{
		Wallets:       pbWallets,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) DeleteWallet(ctx context.Context, req *pb.DeleteWalletRequest) (*pb.DeleteWalletResponse, error) {
	s.logger.Info("DeleteWallet gRPC request received", "wallet_id", req.Id)

	deleteWithoutExport := false
	if req.DeleteWithoutExport != nil {
		deleteWithoutExport = *req.DeleteWithoutExport
	}

	err := s.service.DeleteWallet(ctx, req.Id, deleteWithoutExport)
	if err != nil {
		s.logger.Error("Failed to delete wallet", "error", err, "wallet_id", req.Id)
		return nil, status.Errorf(codes.Internal, "failed to delete wallet: %v", err)
	}

	return &pb.DeleteWalletResponse{
		Success: true,
		Message: "Wallet deleted successfully",
	}, nil
}

// Private key management methods

func (s *Server) CreatePrivateKey(ctx context.Context, req *pb.CreatePrivateKeyRequest) (*pb.CreatePrivateKeyResponse, error) {
	s.logger.Info("CreatePrivateKey gRPC request received", "organization_id", req.OrganizationId, "wallet_id", req.WalletId, "name", req.Name)

	privateKey, err := s.service.CreatePrivateKey(ctx, req.OrganizationId, req.WalletId, req.Name, req.Curve, req.PrivateKeyMaterial, req.Tags)
	if err != nil {
		s.logger.Error("Failed to create private key", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create private key: %v", err)
	}

	return &pb.CreatePrivateKeyResponse{
		PrivateKey: convertPrivateKeyToProto(privateKey),
	}, nil
}

func (s *Server) GetPrivateKey(ctx context.Context, req *pb.GetPrivateKeyRequest) (*pb.GetPrivateKeyResponse, error) {
	s.logger.Debug("GetPrivateKey gRPC request received", "private_key_id", req.Id)

	privateKey, err := s.service.GetPrivateKey(ctx, req.Id)
	if err != nil {
		s.logger.Error("Failed to get private key", "error", err, "private_key_id", req.Id)
		return nil, status.Errorf(codes.NotFound, "private key not found: %v", err)
	}

	return &pb.GetPrivateKeyResponse{
		PrivateKey: convertPrivateKeyToProto(privateKey),
	}, nil
}

func (s *Server) ListPrivateKeys(ctx context.Context, req *pb.ListPrivateKeysRequest) (*pb.ListPrivateKeysResponse, error) {
	s.logger.Debug("ListPrivateKeys gRPC request received", "organization_id", req.OrganizationId)

	privateKeys, nextToken, err := s.service.ListPrivateKeys(ctx, req.OrganizationId, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("Failed to list private keys", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list private keys: %v", err)
	}

	// Convert to protobuf
	pbPrivateKeys := make([]*pb.PrivateKey, len(privateKeys))
	for i, pk := range privateKeys {
		pbPrivateKeys[i] = convertPrivateKeyToProto(&pk)
	}

	return &pb.ListPrivateKeysResponse{
		PrivateKeys:   pbPrivateKeys,
		NextPageToken: nextToken,
	}, nil
}

func (s *Server) DeletePrivateKey(ctx context.Context, req *pb.DeletePrivateKeyRequest) (*pb.DeletePrivateKeyResponse, error) {
	s.logger.Info("DeletePrivateKey gRPC request received", "private_key_id", req.Id)

	deleteWithoutExport := false
	if req.DeleteWithoutExport != nil {
		deleteWithoutExport = *req.DeleteWithoutExport
	}

	err := s.service.DeletePrivateKey(ctx, req.Id, deleteWithoutExport)
	if err != nil {
		s.logger.Error("Failed to delete private key", "error", err, "private_key_id", req.Id)
		return nil, status.Errorf(codes.Internal, "failed to delete private key: %v", err)
	}

	return &pb.DeletePrivateKeyResponse{
		Success: true,
		Message: "Private key deleted successfully",
	}, nil
}
