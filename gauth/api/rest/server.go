package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server represents the REST API server that communicates with gRPC internally
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	grpcClient pb.GAuthServiceClient
	grpcConn   *grpc.ClientConn
	router     *gin.Engine
	httpServer *http.Server
}

// NewServer creates a new REST API server instance
func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

// Start starts the REST API server
func (s *Server) Start() error {
	// Connect to gRPC server
	if err := s.connectToGRPC(); err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Setup router
	s.setupRouter()

	// Create HTTP server
	addr := s.config.GetServerAddr()
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	s.logger.Info("Starting REST API server", "address", addr)

	// Start serving
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve REST API server: %w", err)
	}

	return nil
}

// Stop gracefully stops the REST API server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.logger.Info("Stopping REST API server")

	// Close gRPC connection
	if s.grpcConn != nil {
		s.grpcConn.Close()
	}

	// Shutdown HTTP server
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

// connectToGRPC establishes connection to the gRPC server
func (s *Server) connectToGRPC() error {
	grpcAddr := s.config.GetGRPCAddr()

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GRPC.ConnectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at %s: %w", grpcAddr, err)
	}

	s.grpcConn = conn
	s.grpcClient = pb.NewGAuthServiceClient(conn)

	s.logger.Info("Connected to gRPC server", "address", grpcAddr)
	return nil
}

// setupRouter configures the Gin router with all routes
func (s *Server) setupRouter() {
	// Set Gin mode based on log level
	if s.config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.New()

	// Add middleware
	s.router.Use(s.loggingMiddleware())
	s.router.Use(s.recoveryMiddleware())

	if s.config.Security.CORSEnabled {
		s.router.Use(s.corsMiddleware())
	}

	if s.config.Security.RateLimitEnabled {
		s.router.Use(s.rateLimitMiddleware())
	}

	s.SetupAPIRoutes(s.router)
}

// ConnectToGRPCForTesting manually connects to gRPC server for testing
func (s *Server) ConnectToGRPCForTesting(grpcAddr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at %s: %w", grpcAddr, err)
	}

	s.grpcConn = conn
	s.grpcClient = pb.NewGAuthServiceClient(conn)

	s.logger.Info("Connected to gRPC server for testing", "address", grpcAddr)
	return nil
}

// SetupAPIRoutes sets up API routes on the given router (for testing)
func (s *Server) SetupAPIRoutes(router *gin.Engine) {
	// API versioning
	v1 := router.Group("/api/v1")
	{
		// Health endpoints
		v1.GET("/health", s.handleHealth)
		v1.GET("/status", s.handleStatus)

		// Organization endpoints
		orgs := v1.Group("/organizations")
		{
			orgs.POST("", s.handleCreateOrganization)
			orgs.GET("/:id", s.handleGetOrganization)
			orgs.PUT("/:id", s.handleUpdateOrganization)
			orgs.GET("", s.handleListOrganizations)
		}

		// User endpoints
		users := v1.Group("/users")
		{
			users.POST("", s.handleCreateUser)
			users.GET("/:id", s.handleGetUser)
			users.PUT("/:id", s.handleUpdateUser)
			users.GET("", s.handleListUsers)
		}

		// Wallet endpoints
		wallets := v1.Group("/wallets")
		{
			wallets.POST("", s.handleCreateWallet)
			wallets.GET("/:id", s.handleGetWallet)
			wallets.GET("", s.handleListWallets)
			wallets.DELETE("/:id", s.handleDeleteWallet)
		}

		// Private key endpoints
		privateKeys := v1.Group("/private-keys")
		{
			privateKeys.POST("", s.handleCreatePrivateKey)
			privateKeys.GET("", s.handleListPrivateKeys)
			privateKeys.GET("/:id", s.handleGetPrivateKey)
			privateKeys.DELETE("/:id", s.handleDeletePrivateKey)
		}

		// Activity endpoints
		activities := v1.Group("/activities")
		{
			activities.POST("", s.handleCreateActivity)
			activities.GET("/:id", s.handleGetActivity)
			activities.GET("", s.handleListActivities)
		}

		// Authentication endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/authenticate", s.handleAuthenticate)
			auth.POST("/authorize", s.handleAuthorize)
		}

		// Renclave endpoints
		renclave := v1.Group("/renclave")
		{
			renclave.GET("/info", s.handleGetEnclaveInfo)
			renclave.POST("/seed/generate", s.handleRequestSeedGeneration)
			renclave.POST("/seed/validate", s.handleValidateSeed)
		}
	}
}
