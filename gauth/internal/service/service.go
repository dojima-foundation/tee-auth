package service

import (
	"context"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/google/uuid"
)

// GAuthService provides business logic for the gauth service
type GAuthService struct {
	config    *config.Config
	logger    *logger.Logger
	db        db.DatabaseInterface
	redis     db.RedisInterface
	renclave  EnclaveClientInterface
	startTime time.Time
}

// NewGAuthService creates a new GAuthService instance
func NewGAuthService(
	cfg *config.Config,
	logger *logger.Logger,
	database db.DatabaseInterface,
	redis db.RedisInterface,
) *GAuthService {
	return NewGAuthServiceWithEnclave(cfg, logger, database, redis, nil)
}

// NewGAuthServiceWithEnclave creates a new GAuthService instance with a custom enclave client
func NewGAuthServiceWithEnclave(
	cfg *config.Config,
	logger *logger.Logger,
	database db.DatabaseInterface,
	redis db.RedisInterface,
	enclaveClient EnclaveClientInterface,
) *GAuthService {
	if enclaveClient == nil {
		enclaveClient = NewRenclaveClient(cfg.GetRenclaveAddr(), cfg.Renclave.Timeout)
	}

	return &GAuthService{
		config:    cfg,
		logger:    logger,
		db:        database,
		redis:     redis,
		renclave:  enclaveClient,
		startTime: time.Now(),
	}
}

// Organization management methods

func (s *GAuthService) CreateOrganization(ctx context.Context, name, initialUserEmail, initialUserPublicKey string) (*models.Organization, error) {
	// Delegate to OrganizationService
	orgService := &OrganizationService{GAuthService: s}
	return orgService.CreateOrganization(ctx, name, initialUserEmail, initialUserPublicKey)
}

func (s *GAuthService) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	// Delegate to OrganizationService
	orgService := &OrganizationService{GAuthService: s}
	return orgService.GetOrganization(ctx, id)
}

func (s *GAuthService) UpdateOrganization(ctx context.Context, id string, name *string) (*models.Organization, error) {
	// Delegate to OrganizationService
	orgService := &OrganizationService{GAuthService: s}
	return orgService.UpdateOrganization(ctx, id, name)
}

func (s *GAuthService) ListOrganizations(ctx context.Context, pageSize int, pageToken string) ([]models.Organization, string, error) {
	// Delegate to OrganizationService
	orgService := &OrganizationService{GAuthService: s}
	return orgService.ListOrganizations(ctx, pageSize, pageToken)
}

// User management methods

func (s *GAuthService) CreateUser(ctx context.Context, organizationID, username, email, publicKey string, tags []string) (*models.User, error) {
	// Delegate to UserService
	userService := &UserService{GAuthService: s}
	return userService.CreateUser(ctx, organizationID, username, email, publicKey, tags)
}

func (s *GAuthService) GetUser(ctx context.Context, id string) (*models.User, error) {
	// Delegate to UserService
	userService := &UserService{GAuthService: s}
	return userService.GetUser(ctx, id)
}

func (s *GAuthService) UpdateUser(ctx context.Context, id string, username, email *string, tags []string, isActive *bool) (*models.User, error) {
	// Delegate to UserService
	userService := &UserService{GAuthService: s}
	return userService.UpdateUser(ctx, id, username, email, tags, isActive)
}

func (s *GAuthService) ListUsers(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.User, string, error) {
	// Delegate to UserService
	userService := &UserService{GAuthService: s}
	return userService.ListUsers(ctx, organizationID, pageSize, pageToken)
}

// Activity management methods

func (s *GAuthService) CreateActivity(ctx context.Context, organizationID, activityType, parameters, createdBy string) (*models.Activity, error) {
	// Delegate to ActivityService
	activityService := &ActivityService{GAuthService: s}
	return activityService.CreateActivity(ctx, organizationID, activityType, parameters, createdBy)
}

func (s *GAuthService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	// Delegate to ActivityService
	activityService := &ActivityService{GAuthService: s}
	return activityService.GetActivity(ctx, id)
}

func (s *GAuthService) ListActivities(ctx context.Context, organizationID string, activityType, status *string, pageSize int, pageToken string) ([]models.Activity, string, error) {
	// Delegate to ActivityService
	activityService := &ActivityService{GAuthService: s}
	return activityService.ListActivities(ctx, organizationID, activityType, status, pageSize, pageToken)
}

// Response types for service methods
type SeedGenerationResponse struct {
	SeedPhrase string `json:"seed_phrase"`
	Entropy    string `json:"entropy"`
	Strength   int    `json:"strength"`
	WordCount  int    `json:"word_count"`
	RequestID  string `json:"request_id"`
}

type SeedValidationResponse struct {
	IsValid   bool     `json:"is_valid"`
	Strength  int      `json:"strength"`
	WordCount int      `json:"word_count"`
	Errors    []string `json:"errors"`
}

type EnclaveInfoResponse struct {
	Version      string   `json:"version"`
	EnclaveID    string   `json:"enclave_id"`
	Capabilities []string `json:"capabilities"`
	Healthy      bool     `json:"healthy"`
}

type AuthenticationResponse struct {
	Authenticated bool         `json:"authenticated"`
	SessionToken  string       `json:"session_token"`
	ExpiresAt     time.Time    `json:"expires_at"`
	User          *models.User `json:"user"`
}

type AuthorizationResponse struct {
	Authorized        bool     `json:"authorized"`
	Reason            string   `json:"reason"`
	RequiredApprovals []string `json:"required_approvals"`
}

type ServiceHealthResponse struct {
	Status    string          `json:"status"`
	Services  []ServiceStatus `json:"services"`
	Timestamp time.Time       `json:"timestamp"`
}

type ServiceStatus struct {
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Error  *string `json:"error,omitempty"`
}

type StatusResponse struct {
	Version   string            `json:"version"`
	BuildTime string            `json:"build_time"`
	GitCommit string            `json:"git_commit"`
	Uptime    time.Time         `json:"uptime"`
	Metrics   map[string]string `json:"metrics"`
}

// Enclave communication methods

func (s *GAuthService) RequestSeedGeneration(ctx context.Context, organizationID, userID string, strength int, passphrase *string) (*SeedGenerationResponse, error) {
	// Delegate to EnclaveService
	enclaveService := &EnclaveService{GAuthService: s}
	return enclaveService.RequestSeedGeneration(ctx, organizationID, userID, strength, passphrase)
}

func (s *GAuthService) ValidateSeed(ctx context.Context, seedPhrase string) (*SeedValidationResponse, error) {
	// Delegate to EnclaveService
	enclaveService := &EnclaveService{GAuthService: s}
	return enclaveService.ValidateSeed(ctx, seedPhrase)
}

func (s *GAuthService) GetEnclaveInfo(ctx context.Context) (*EnclaveInfoResponse, error) {
	// Delegate to EnclaveService
	enclaveService := &EnclaveService{GAuthService: s}
	return enclaveService.GetEnclaveInfo(ctx)
}

// Authentication and authorization methods

func (s *GAuthService) Authenticate(ctx context.Context, organizationID, userID, authMethodID, signature, timestamp string) (*AuthenticationResponse, error) {
	// Delegate to AuthService
	authService := &AuthService{GAuthService: s}
	return authService.Authenticate(ctx, organizationID, userID, authMethodID, signature, timestamp)
}

func (s *GAuthService) Authorize(ctx context.Context, sessionToken, activityType, parameters string) (*AuthorizationResponse, error) {
	// Delegate to AuthService
	authService := &AuthService{GAuthService: s}
	return authService.Authorize(ctx, sessionToken, activityType, parameters)
}

// Health and status methods

func (s *GAuthService) Health(ctx context.Context) (*ServiceHealthResponse, error) {
	// Delegate to HealthService
	healthService := &HealthService{GAuthService: s}
	return healthService.Health(ctx)
}

func (s *GAuthService) Status(ctx context.Context) (*StatusResponse, error) {
	// Delegate to HealthService
	healthService := &HealthService{GAuthService: s}
	return healthService.Status(ctx)
}

// Helper types for database operations that might be missing
type QuorumMember struct {
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id" gorm:"type:uuid;not null"`
	UserID         uuid.UUID `json:"user_id" db:"user_id" gorm:"type:uuid;not null"`
}

func (QuorumMember) TableName() string {
	return "quorum_members"
}

// Wallet management methods

func (s *GAuthService) CreateWallet(ctx context.Context, organizationID, name string, accounts []models.WalletAccount, mnemonicLength *int32, tags []string) (*models.Wallet, []string, error) {
	// Delegate to WalletService
	walletService := &WalletService{GAuthService: s}
	return walletService.CreateWallet(ctx, organizationID, name, accounts, mnemonicLength, tags)
}

func (s *GAuthService) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	// Delegate to WalletService
	walletService := &WalletService{GAuthService: s}
	return walletService.GetWallet(ctx, walletID)
}

func (s *GAuthService) ListWallets(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.Wallet, string, error) {
	// Delegate to WalletService
	walletService := &WalletService{GAuthService: s}
	return walletService.ListWallets(ctx, organizationID, pageSize, pageToken)
}

func (s *GAuthService) DeleteWallet(ctx context.Context, walletID string, deleteWithoutExport bool) error {
	// Delegate to WalletService
	walletService := &WalletService{GAuthService: s}
	return walletService.DeleteWallet(ctx, walletID, deleteWithoutExport)
}

// Private key management methods

func (s *GAuthService) CreatePrivateKey(ctx context.Context, organizationID, walletID, name, curve string, privateKeyMaterial *string, tags []string) (*models.PrivateKey, error) {
	// Delegate to PrivateKeyService
	privateKeyService := &PrivateKeyService{GAuthService: s}
	return privateKeyService.CreatePrivateKey(ctx, organizationID, walletID, name, curve, privateKeyMaterial, tags)
}

func (s *GAuthService) GetPrivateKey(ctx context.Context, privateKeyID string) (*models.PrivateKey, error) {
	// Delegate to PrivateKeyService
	privateKeyService := &PrivateKeyService{GAuthService: s}
	return privateKeyService.GetPrivateKey(ctx, privateKeyID)
}

func (s *GAuthService) ListPrivateKeys(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.PrivateKey, string, error) {
	// Delegate to PrivateKeyService
	privateKeyService := &PrivateKeyService{GAuthService: s}
	return privateKeyService.ListPrivateKeys(ctx, organizationID, pageSize, pageToken)
}

func (s *GAuthService) DeletePrivateKey(ctx context.Context, privateKeyID string, deleteWithoutExport bool) error {
	// Delegate to PrivateKeyService
	privateKeyService := &PrivateKeyService{GAuthService: s}
	return privateKeyService.DeletePrivateKey(ctx, privateKeyID, deleteWithoutExport)
}
