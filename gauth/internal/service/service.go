package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GAuthService provides business logic for the gauth service
type GAuthService struct {
	config   *config.Config
	logger   *logger.Logger
	db       db.DatabaseInterface
	redis    db.RedisInterface
	renclave *RenclaveClient
	startTime time.Time
}

// NewGAuthService creates a new GAuthService instance
func NewGAuthService(
	cfg *config.Config,
	logger *logger.Logger,
	database db.DatabaseInterface,
	redis db.RedisInterface,
) *GAuthService {
	renclave := NewRenclaveClient(cfg.GetRenclaveAddr(), cfg.Renclave.Timeout)

	return &GAuthService{
		config:    cfg,
		logger:    logger,
		db:        database,
		redis:     redis,
		renclave:  renclave,
		startTime: time.Now(),
	}
}

// Organization management methods

func (s *GAuthService) CreateOrganization(ctx context.Context, name, initialUserEmail, initialUserPublicKey string) (*models.Organization, error) {
	s.logger.Info("Creating organization", "name", name, "initial_user_email", initialUserEmail)

	org := &models.Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    name,
		RootQuorum: models.Quorum{
			Threshold: s.config.Auth.DefaultQuorumThreshold,
		},
	}

	// Create initial user
	initialUser := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Username:       "admin",
		Email:          initialUserEmail,
		IsActive:       true,
	}

	if initialUserPublicKey != "" {
		initialUser.PublicKey = initialUserPublicKey
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Create organization
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		// Create initial user
		if err := tx.Create(&initialUser).Error; err != nil {
			return fmt.Errorf("failed to create initial user: %w", err)
		}

		// Add user to root quorum
		if err := tx.Create(&QuorumMember{
			OrganizationID: org.ID,
			UserID:         initialUser.ID,
		}).Error; err != nil {
			return fmt.Errorf("failed to add user to quorum: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create organization", "error", err)
		return nil, err
	}

	// Load users for the organization before returning
	if err := s.db.GetDB().WithContext(ctx).Preload("Users").First(org, org.ID).Error; err != nil {
		s.logger.Error("Failed to load organization with users", "error", err)
		return nil, fmt.Errorf("failed to load organization with users: %w", err)
	}

	s.logger.Info("Organization created successfully", "organization_id", org.ID.String())
	return org, nil
}

func (s *GAuthService) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	orgID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var org models.Organization
	if err := s.db.GetDB().WithContext(ctx).
		Preload("Users").
		Preload("Invitations").
		Preload("Policies").
		Preload("Tags").
		Preload("PrivateKeys").
		Preload("Wallets.Accounts").
		First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

func (s *GAuthService) UpdateOrganization(ctx context.Context, id string, name *string) (*models.Organization, error) {
	orgID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	var org models.Organization
	if err := s.db.GetDB().WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if name != nil {
		org.Name = *name
	}

	if err := s.db.GetDB().WithContext(ctx).Save(&org).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &org, nil
}

func (s *GAuthService) ListOrganizations(ctx context.Context, pageSize int, pageToken string) ([]models.Organization, string, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var orgs []models.Organization
	query := s.db.GetDB().WithContext(ctx).Limit(pageSize + 1) // +1 to check if there are more records

	if pageToken != "" {
		// Simple pagination using ID - in production, you might want to use cursor-based pagination
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	if err := query.Find(&orgs).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list organizations: %w", err)
	}

	var nextToken string
	if len(orgs) > pageSize {
		nextToken = orgs[pageSize-1].ID.String()
		orgs = orgs[:pageSize]
	}

	return orgs, nextToken, nil
}

// User management methods

func (s *GAuthService) CreateUser(ctx context.Context, organizationID, username, email, publicKey string, tags []string) (*models.User, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	user := &models.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Username:       username,
		Email:          email,
		Tags:           tags,
		IsActive:       true,
	}

	if publicKey != "" {
		user.PublicKey = publicKey
	}

	if err := s.db.GetDB().WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created", "user_id", user.ID.String(), "organization_id", organizationID)
	return user, nil
}

func (s *GAuthService) GetUser(ctx context.Context, id string) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	if err := s.db.GetDB().WithContext(ctx).
		Preload("AuthMethods").
		First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *GAuthService) UpdateUser(ctx context.Context, id string, username, email *string, tags []string, isActive *bool) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user models.User
	if err := s.db.GetDB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if username != nil {
		user.Username = *username
	}
	if email != nil {
		user.Email = *email
	}
	if tags != nil {
		user.Tags = tags
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	if err := s.db.GetDB().WithContext(ctx).Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (s *GAuthService) ListUsers(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.User, string, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid organization ID: %w", err)
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var users []models.User
	query := s.db.GetDB().WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("id ASC").
		Limit(pageSize + 1)

	if pageToken != "" {
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list users: %w", err)
	}

	var nextToken string
	if len(users) > pageSize {
		nextToken = users[pageSize-1].ID.String()
		users = users[:pageSize]
	}

	return users, nextToken, nil
}

// Activity management methods

func (s *GAuthService) CreateActivity(ctx context.Context, organizationID, activityType, parameters, createdBy string) (*models.Activity, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	createdByID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created_by user ID: %w", err)
	}

	// Validate and convert JSON parameters
	var parametersRaw json.RawMessage
	if parameters != "" {
		if err := json.Unmarshal([]byte(parameters), &parametersRaw); err != nil {
			return nil, fmt.Errorf("invalid JSON parameters: %w", err)
		}
	}

	activity := &models.Activity{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Type:           activityType,
		Status:         "PENDING",
		Parameters:     parametersRaw,
		CreatedBy:      createdByID,
	}

	if err := s.db.GetDB().WithContext(ctx).Create(activity).Error; err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	s.logger.LogActivity(activityType, activity.ID.String(), createdBy, organizationID)
	return activity, nil
}

func (s *GAuthService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	activityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid activity ID: %w", err)
	}

	var activity models.Activity
	if err := s.db.GetDB().WithContext(ctx).First(&activity, "id = ?", activityID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("activity not found")
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

func (s *GAuthService) ListActivities(ctx context.Context, organizationID string, activityType, status *string, pageSize int, pageToken string) ([]models.Activity, string, error) {
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid organization ID: %w", err)
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := s.db.GetDB().WithContext(ctx).
		Where("organization_id = ?", orgID).
		Limit(pageSize + 1)

	if activityType != nil {
		query = query.Where("type = ?", *activityType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if pageToken != "" {
		if tokenID, err := uuid.Parse(pageToken); err == nil {
			query = query.Where("id > ?", tokenID)
		}
	}

	var activities []models.Activity
	if err := query.Find(&activities).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list activities: %w", err)
	}

	var nextToken string
	if len(activities) > pageSize {
		nextToken = activities[pageSize-1].ID.String()
		activities = activities[:pageSize]
	}

	return activities, nextToken, nil
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

// Communication with renclave-v2

func (s *GAuthService) RequestSeedGeneration(ctx context.Context, organizationID, userID string, strength int, passphrase *string) (*SeedGenerationResponse, error) {
	s.logger.Info("Requesting seed generation",
		"organization_id", organizationID,
		"user_id", userID,
		"strength", strength,
	)

	response, err := s.renclave.GenerateSeed(ctx, strength, passphrase)
	if err != nil {
		s.logger.Error("Failed to generate seed", "error", err)
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// Log the activity
	s.logger.LogActivity("SEED_GENERATION", uuid.New().String(), userID, organizationID,
		"strength", strength,
		"has_passphrase", passphrase != nil,
	)

	return &SeedGenerationResponse{
		SeedPhrase: response.SeedPhrase,
		Entropy:    response.Entropy,
		Strength:   response.Strength,
		WordCount:  response.WordCount,
		RequestID:  uuid.New().String(),
	}, nil
}

func (s *GAuthService) ValidateSeed(ctx context.Context, seedPhrase string) (*SeedValidationResponse, error) {
	response, err := s.renclave.ValidateSeed(ctx, seedPhrase)
	if err != nil {
		s.logger.Error("Failed to validate seed", "error", err)
		return nil, fmt.Errorf("failed to validate seed: %w", err)
	}

	return &SeedValidationResponse{
		IsValid:   response.IsValid,
		Strength:  response.Strength,
		WordCount: response.WordCount,
		Errors:    response.Errors,
	}, nil
}

func (s *GAuthService) GetEnclaveInfo(ctx context.Context) (*EnclaveInfoResponse, error) {
	info, err := s.renclave.GetInfo(ctx)
	if err != nil {
		s.logger.Error("Failed to get enclave info", "error", err)
		return nil, fmt.Errorf("failed to get enclave info: %w", err)
	}

	return &EnclaveInfoResponse{
		Version:      info.Version,
		EnclaveID:    info.EnclaveID,
		Capabilities: info.Capabilities,
		Healthy:      info.Healthy,
	}, nil
}

// Authentication and authorization (placeholder implementations)

func (s *GAuthService) Authenticate(ctx context.Context, organizationID, userID, authMethodID, signature, timestamp string) (*AuthenticationResponse, error) {
	// This is a simplified implementation - in production you would:
	// 1. Verify the signature against the user's public key
	// 2. Check timestamp to prevent replay attacks
	// 3. Validate the auth method
	// 4. Create a session token

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(s.config.Auth.SessionTimeout)

	// Store session in Redis
	if err := s.redis.SetSession(ctx, sessionToken, userID, s.config.Auth.SessionTimeout); err != nil {
		s.logger.Error("Failed to store session", "error", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.LogAuthenticationAttempt(userID, organizationID, true)

	return &AuthenticationResponse{
		Authenticated: true,
		SessionToken:  sessionToken,
		ExpiresAt:     expiresAt,
		User:          user,
	}, nil
}

func (s *GAuthService) Authorize(ctx context.Context, sessionToken, activityType, parameters string) (*AuthorizationResponse, error) {
	// This is a simplified implementation - in production you would:
	// 1. Validate the session token
	// 2. Get the user from the session
	// 3. Check policies and permissions
	// 4. Evaluate quorum requirements

	userID, err := s.redis.GetSession(ctx, sessionToken)
	if err != nil {
		return &AuthorizationResponse{
			Authorized: false,
			Reason:     "Invalid session",
		}, nil
	}

	// Simple authorization - in production, implement policy engine
	_ = userID // Mark as used
	return &AuthorizationResponse{
		Authorized:        true,
		Reason:            "Policy evaluation passed",
		RequiredApprovals: []string{}, // No additional approvals needed
	}, nil
}

// Health and status

func (s *GAuthService) Health(ctx context.Context) (*ServiceHealthResponse, error) {
	services := []ServiceStatus{}

	// Check database health
	if err := s.db.Health(ctx); err != nil {
		errStr := err.Error()
		services = append(services, ServiceStatus{
			Name:   "database",
			Status: "unhealthy",
			Error:  &errStr,
		})
	} else {
		services = append(services, ServiceStatus{
			Name:   "database",
			Status: "healthy",
		})
	}

	// Check Redis health
	if err := s.redis.Health(ctx); err != nil {
		errStr := err.Error()
		services = append(services, ServiceStatus{
			Name:   "redis",
			Status: "unhealthy",
			Error:  &errStr,
		})
	} else {
		services = append(services, ServiceStatus{
			Name:   "redis",
			Status: "healthy",
		})
	}

	// Check renclave health
	if err := s.renclave.Health(ctx); err != nil {
		errStr := err.Error()
		services = append(services, ServiceStatus{
			Name:   "renclave",
			Status: "unhealthy",
			Error:  &errStr,
		})
	} else {
		services = append(services, ServiceStatus{
			Name:   "renclave",
			Status: "healthy",
		})
	}

	// Determine overall status
	status := "healthy"
	for _, svc := range services {
		if svc.Status != "healthy" {
			status = "degraded"
			break
		}
	}

	return &ServiceHealthResponse{
		Status:    status,
		Services:  services,
		Timestamp: time.Now().UTC(),
	}, nil
}

func (s *GAuthService) Status(ctx context.Context) (*StatusResponse, error) {
	return &StatusResponse{
		Version:   "1.0.0",
		BuildTime: "2024-01-01T00:00:00Z", // This would be set at build time
		GitCommit: "unknown",              // This would be set at build time
		Uptime:    time.Now().UTC(),       // This would be the actual service start time
		Metrics: map[string]string{
			"go_version": "go1.21",
		},
	}, nil
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
	s.logger.Info("Creating wallet", "organization_id", organizationID, "name", name)

	// Default mnemonic length to 12 if not specified
	length := int32(12)
	if mnemonicLength != nil {
		length = *mnemonicLength
	}

	// Validate mnemonic length
	validLengths := []int32{12, 15, 18, 21, 24}
	valid := false
	for _, vl := range validLengths {
		if length == vl {
			valid = true
			break
		}
	}
	if !valid {
		return nil, nil, fmt.Errorf("invalid mnemonic length: %d. Must be one of: 12, 15, 18, 21, 24", length)
	}

	wallet := &models.Wallet{
		ID:             uuid.New(),
		OrganizationID: uuid.MustParse(organizationID),
		Name:           name,
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate addresses for each account
	addresses := make([]string, len(accounts))
	for i := range accounts {
		// Set wallet ID for the account
		accounts[i].ID = uuid.New()
		accounts[i].WalletID = wallet.ID
		accounts[i].IsActive = true
		accounts[i].CreatedAt = time.Now()
		accounts[i].UpdatedAt = time.Now()

		// For now, generate mock addresses - in production, integrate with renclave
		addresses[i] = fmt.Sprintf("0x%s%d", uuid.New().String()[:32], i)
		accounts[i].Address = addresses[i]
		accounts[i].PublicKey = fmt.Sprintf("pub_%s", uuid.New().String()[:16])
	}

	wallet.Accounts = accounts

	// Save to database
	if err := s.db.GetDB().WithContext(ctx).Create(wallet).Error; err != nil {
		s.logger.Error("Failed to create wallet", "error", err)
		return nil, nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	s.logger.Info("Wallet created successfully", "wallet_id", wallet.ID.String())
	return wallet, addresses, nil
}

func (s *GAuthService) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	s.logger.Debug("Getting wallet", "wallet_id", walletID)

	var wallet models.Wallet
	if err := s.db.GetDB().WithContext(ctx).Preload("Accounts").Where("id = ?", walletID).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("wallet not found: %s", walletID)
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

func (s *GAuthService) ListWallets(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.Wallet, string, error) {
	s.logger.Debug("Listing wallets", "organization_id", organizationID)

	var wallets []models.Wallet
	query := s.db.GetDB().WithContext(ctx).Preload("Accounts").Where("organization_id = ?", organizationID)

	// Handle pagination
	if pageToken != "" {
		// Validate pageToken as UUID
		if _, err := uuid.Parse(pageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		query = query.Where("id > ?", pageToken)
	}

	query = query.Order("created_at ASC").Limit(pageSize + 1) // Get one extra to check if there's a next page

	if err := query.Find(&wallets).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list wallets: %w", err)
	}

	var nextToken string
	if len(wallets) > pageSize {
		// There are more items available, set nextToken to the last item of current page
		nextToken = wallets[pageSize-1].ID.String()
		wallets = wallets[:pageSize]
	}
	// If len(wallets) <= pageSize, nextToken remains empty string (no more pages)

	return wallets, nextToken, nil
}

func (s *GAuthService) DeleteWallet(ctx context.Context, walletID string, deleteWithoutExport bool) error {
	s.logger.Info("Deleting wallet", "wallet_id", walletID, "delete_without_export", deleteWithoutExport)

	// For production: Check if wallet has been exported unless deleteWithoutExport is true
	if !deleteWithoutExport {
		// Check export status - for now, we'll allow deletion
		s.logger.Warn("Wallet deletion without export check - implement export verification in production")
	}

	// Delete wallet accounts first (due to foreign key constraints)
	if err := s.db.GetDB().WithContext(ctx).Where("wallet_id = ?", walletID).Delete(&models.WalletAccount{}).Error; err != nil {
		return fmt.Errorf("failed to delete wallet accounts: %w", err)
	}

	// Delete the wallet
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", walletID).Delete(&models.Wallet{}).Error; err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	s.logger.Info("Wallet deleted successfully", "wallet_id", walletID)
	return nil
}

// Private key management methods

func (s *GAuthService) CreatePrivateKey(ctx context.Context, organizationID, name, curve string, privateKeyMaterial *string, tags []string) (*models.PrivateKey, error) {
	s.logger.Info("Creating private key", "organization_id", organizationID, "name", name, "curve", curve)

	// Validate organization ID
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Validate curve
	validCurves := []string{"CURVE_SECP256K1", "CURVE_ED25519"}
	valid := false
	for _, vc := range validCurves {
		if curve == vc {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("invalid curve: %s. Must be one of: CURVE_SECP256K1, CURVE_ED25519", curve)
	}

	privateKey := &models.PrivateKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Curve:          curve,
		Tags:           tags,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Generate public key - in production, integrate with renclave
	if privateKeyMaterial != nil {
		// Use provided private key material to derive public key
		privateKey.PublicKey = fmt.Sprintf("pub_from_material_%s", uuid.New().String()[:16])
	} else {
		// Generate new key pair
		privateKey.PublicKey = fmt.Sprintf("pub_%s", uuid.New().String()[:32])
	}

	// Save to database
	if err := s.db.GetDB().WithContext(ctx).Create(privateKey).Error; err != nil {
		s.logger.Error("Failed to create private key", "error", err)
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	s.logger.Info("Private key created successfully", "private_key_id", privateKey.ID.String())
	return privateKey, nil
}

func (s *GAuthService) GetPrivateKey(ctx context.Context, privateKeyID string) (*models.PrivateKey, error) {
	s.logger.Debug("Getting private key", "private_key_id", privateKeyID)

	var privateKey models.PrivateKey
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", privateKeyID).First(&privateKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("private key not found: %s", privateKeyID)
		}
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return &privateKey, nil
}

func (s *GAuthService) ListPrivateKeys(ctx context.Context, organizationID string, pageSize int, pageToken string) ([]models.PrivateKey, string, error) {
	s.logger.Debug("Listing private keys", "organization_id", organizationID)

	var privateKeys []models.PrivateKey
	query := s.db.GetDB().WithContext(ctx).Where("organization_id = ?", organizationID)

	// Handle pagination
	if pageToken != "" {
		// Validate pageToken as UUID
		if _, err := uuid.Parse(pageToken); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		query = query.Where("id > ?", pageToken)
	}

	query = query.Order("created_at ASC").Limit(pageSize + 1) // Get one extra to check if there's a next page

	if err := query.Find(&privateKeys).Error; err != nil {
		return nil, "", fmt.Errorf("failed to list private keys: %w", err)
	}

	var nextToken string
	if len(privateKeys) > pageSize {
		nextToken = privateKeys[pageSize-1].ID.String()
		privateKeys = privateKeys[:pageSize]
	}

	return privateKeys, nextToken, nil
}

func (s *GAuthService) DeletePrivateKey(ctx context.Context, privateKeyID string, deleteWithoutExport bool) error {
	s.logger.Info("Deleting private key", "private_key_id", privateKeyID, "delete_without_export", deleteWithoutExport)

	// For production: Check if private key has been exported unless deleteWithoutExport is true
	if !deleteWithoutExport {
		// Check export status - for now, we'll allow deletion
		s.logger.Warn("Private key deletion without export check - implement export verification in production")
	}

	// Delete the private key
	if err := s.db.GetDB().WithContext(ctx).Where("id = ?", privateKeyID).Delete(&models.PrivateKey{}).Error; err != nil {
		return fmt.Errorf("failed to delete private key: %w", err)
	}

	s.logger.Info("Private key deleted successfully", "private_key_id", privateKeyID)
	return nil
}
