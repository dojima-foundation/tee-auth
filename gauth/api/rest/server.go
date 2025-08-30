package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server represents the REST API server that communicates with gRPC internally
type Server struct {
	config             *config.Config
	logger             *logger.Logger
	grpcClient         pb.GAuthServiceClient
	grpcConn           *grpc.ClientConn
	router             *gin.Engine
	httpServer         *http.Server
	telemetry          *telemetry.Telemetry
	googleOAuthService *service.GoogleOAuthService
	sessionManager     *SessionManager
	redis              db.RedisInterface
}

// NewServer creates a new REST API server instance
func NewServer(cfg *config.Config, logger *logger.Logger, tel *telemetry.Telemetry) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		telemetry: tel,
	}
}

// SetRedis sets the Redis interface for the server
func (s *Server) SetRedis(redis db.RedisInterface) {
	s.redis = redis
	s.sessionManager = NewSessionManager(s)
}

// GetSessionManager returns the session manager instance
func (s *Server) GetSessionManager() *SessionManager {
	return s.sessionManager
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

	// Configure gRPC client options
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Add OpenTelemetry middleware if telemetry is available and tracing is enabled
	if s.telemetry != nil && s.config.Telemetry.TracingEnabled {
		dialOpts = append(dialOpts,
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
	}

	conn, err := grpc.DialContext(ctx, grpcAddr, dialOpts...)
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
	s.router.Use(s.recoveryMiddleware())

	// Add OpenTelemetry middleware if telemetry is available and tracing is enabled
	if s.telemetry != nil && s.config.Telemetry.TracingEnabled {
		s.router.Use(telemetry.HTTPMiddleware(s.telemetry, s.logger))
	} else {
		// Fall back to basic logging middleware
		s.router.Use(s.loggingMiddleware())
	}

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

	// Configure gRPC client options
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Add OpenTelemetry middleware if telemetry is available and tracing is enabled
	if s.telemetry != nil && s.config.Telemetry.TracingEnabled {
		dialOpts = append(dialOpts,
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
	}

	conn, err := grpc.DialContext(ctx, grpcAddr, dialOpts...)
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
			users.POST("", s.sessionManager.SessionMiddleware(), s.handleCreateUser)
			users.GET("/:id", s.sessionManager.SessionMiddleware(), s.handleGetUser)
			users.PUT("/:id", s.sessionManager.SessionMiddleware(), s.handleUpdateUser)
			users.GET("", s.sessionManager.SessionMiddleware(), s.handleListUsers)
		}

		// Wallet endpoints
		wallets := v1.Group("/wallets")
		{
			wallets.POST("", s.sessionManager.SessionMiddleware(), s.handleCreateWallet)
			wallets.GET("/:id", s.sessionManager.SessionMiddleware(), s.handleGetWallet)
			wallets.GET("", s.sessionManager.SessionMiddleware(), s.handleListWallets)
			wallets.DELETE("/:id", s.sessionManager.SessionMiddleware(), s.handleDeleteWallet)
		}

		// Private key endpoints
		privateKeys := v1.Group("/private-keys")
		{
			privateKeys.POST("", s.sessionManager.SessionMiddleware(), s.handleCreatePrivateKey)
			privateKeys.GET("", s.sessionManager.SessionMiddleware(), s.handleListPrivateKeys)
			privateKeys.GET("/:id", s.sessionManager.SessionMiddleware(), s.handleGetPrivateKey)
			privateKeys.DELETE("/:id", s.sessionManager.SessionMiddleware(), s.handleDeletePrivateKey)
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

		// Session endpoints
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/info", s.sessionManager.SessionMiddleware(), s.handleSessionInfo)
			sessions.POST("/refresh", s.handleSessionRefresh)
			sessions.POST("/logout", s.handleSessionLogout)
			sessions.GET("/validate", s.handleSessionValidate)
			sessions.GET("/list", s.sessionManager.SessionMiddleware(), s.handleSessionList)
			sessions.DELETE("/:id", s.sessionManager.SessionMiddleware(), s.handleSessionDestroy)
		}

		// Google OAuth endpoints
		googleAuth := v1.Group("/auth/google")
		{
			googleAuth.POST("/login", s.handleGoogleOAuthLogin)
			googleAuth.GET("/callback", s.handleGoogleOAuthCallback)
			googleAuth.POST("/refresh/:id", s.handleGoogleOAuthRefresh)
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
