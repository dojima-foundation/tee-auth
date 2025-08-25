package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Organization represents the main data container following Turnkey's architecture
// Organizations encapsulate signing resources (Wallets and Private keys) as well as
// data relevant for authentication and authorization of their usage
type Organization struct {
	ID          uuid.UUID    `json:"uuid" db:"id" gorm:"type:uuid;primary_key"`
	Version     string       `json:"version" db:"version" gorm:"not null"`
	Name        string       `json:"name" db:"name" gorm:"not null"`
	Users       []User       `json:"users" db:"-" gorm:"foreignKey:OrganizationID"`
	RootQuorum  Quorum       `json:"rootQuorum" db:"-" gorm:"embedded;embeddedPrefix:root_quorum_"`
	Invitations []Invitation `json:"invitations" db:"-" gorm:"foreignKey:OrganizationID"`
	Policies    []Policy     `json:"policies" db:"-" gorm:"foreignKey:OrganizationID"`
	Tags        []Tag        `json:"tags" db:"-" gorm:"foreignKey:OrganizationID"`
	PrivateKeys []PrivateKey `json:"privateKeys" db:"-" gorm:"foreignKey:OrganizationID"`
	Wallets     []Wallet     `json:"wallets" db:"-" gorm:"foreignKey:OrganizationID"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// User represents an authenticated entity within an organization
type User struct {
	ID             uuid.UUID    `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID    `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Username       string       `json:"username" db:"username" gorm:"not null"`
	Email          string       `json:"email" db:"email" gorm:"not null"`
	PublicKey      string       `json:"publicKey" db:"public_key" gorm:"not null"`
	AuthMethods    []AuthMethod `json:"authMethods" db:"-" gorm:"foreignKey:UserID"`
	Tags           []string     `json:"tags" db:"-" gorm:"-"`
	IsActive       bool         `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt      time.Time    `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// AuthMethod represents authentication methods for users
type AuthMethod struct {
	ID        uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"userId" db:"user_id" gorm:"type:uuid;not null"`
	Type      string    `json:"type" db:"type" gorm:"not null"` // API_KEY, PASSKEY, OAUTH
	Name      string    `json:"name" db:"name" gorm:"not null"`
	Data      string    `json:"data" db:"data" gorm:"not null"` // JSON data specific to auth type
	IsActive  bool      `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// Quorum defines the quorum requirements for critical operations
type Quorum struct {
	UserIDs   []uuid.UUID `json:"userIds" db:"-" gorm:"-"`
	Threshold int         `json:"threshold" db:"threshold" gorm:"not null"`
}

// Invitation represents pending invitations to join an organization
type Invitation struct {
	ID             uuid.UUID  `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID  `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Email          string     `json:"email" db:"email" gorm:"not null"`
	Role           string     `json:"role" db:"role" gorm:"not null"`
	Token          string     `json:"token" db:"token" gorm:"not null;unique"`
	ExpiresAt      time.Time  `json:"expiresAt" db:"expires_at" gorm:"not null"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty" db:"accepted_at"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
}

// Policy represents authorization policies within an organization
type Policy struct {
	ID             uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Name           string    `json:"name" db:"name" gorm:"not null"`
	Description    string    `json:"description" db:"description"`
	Rules          string    `json:"rules" db:"rules" gorm:"type:text;not null"` // JSON policy rules
	IsActive       bool      `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// Tag represents categorization labels for resources
type Tag struct {
	ID             uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Name           string    `json:"name" db:"name" gorm:"not null"`
	Description    string    `json:"description" db:"description"`
	Color          string    `json:"color" db:"color"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
}

// PrivateKey represents cryptographic private keys managed by the system
type PrivateKey struct {
	ID             uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	WalletID       uuid.UUID `json:"walletId" db:"wallet_id" gorm:"type:uuid;not null"` // Link to wallet
	Name           string    `json:"name" db:"name" gorm:"not null"`
	PublicKey      string    `json:"publicKey" db:"public_key" gorm:"not null"`
	Curve          string    `json:"curve" db:"curve" gorm:"not null"` // SECP256K1, ED25519, etc.
	Path           string    `json:"path" db:"path" gorm:"not null"`   // BIP32 derivation path
	Tags           []string  `json:"tags" db:"-" gorm:"-"`
	IsActive       bool      `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// Wallet represents HD wallets compliant with BIP39
type Wallet struct {
	ID             uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID       `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Name           string          `json:"name" db:"name" gorm:"not null"`
	SeedPhrase     string          `json:"seedPhrase" db:"seed_phrase" gorm:"type:text;not null"` // Encrypted seed phrase
	PublicKey      string          `json:"publicKey" db:"public_key" gorm:"not null"`
	Accounts       []WalletAccount `json:"accounts" db:"-" gorm:"foreignKey:WalletID"`
	Tags           []string        `json:"tags" db:"-" gorm:"-"`
	IsActive       bool            `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt      time.Time       `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// WalletAccount represents individual accounts within an HD wallet
type WalletAccount struct {
	ID            uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	WalletID      uuid.UUID `json:"walletId" db:"wallet_id" gorm:"type:uuid;not null"`
	Name          string    `json:"name" db:"name" gorm:"not null"`
	Path          string    `json:"path" db:"path" gorm:"not null"` // BIP44 derivation path
	PublicKey     string    `json:"publicKey" db:"public_key" gorm:"not null"`
	Address       string    `json:"address" db:"address" gorm:"not null"`
	Curve         string    `json:"curve" db:"curve" gorm:"not null"`
	AddressFormat string    `json:"addressFormat" db:"address_format" gorm:"not null"`
	IsActive      bool      `json:"isActive" db:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// Activity represents critical operations within the system
type Activity struct {
	ID             uuid.UUID       `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	OrganizationID uuid.UUID       `json:"organizationId" db:"organization_id" gorm:"type:uuid;not null"`
	Type           string          `json:"type" db:"type" gorm:"not null"`
	Status         string          `json:"status" db:"status" gorm:"not null"` // PENDING, COMPLETED, FAILED
	Parameters     json.RawMessage `json:"parameters" db:"parameters" gorm:"type:jsonb"`
	Result         json.RawMessage `json:"result,omitempty" db:"result" gorm:"type:jsonb"`
	Intent         ActivityIntent  `json:"intent" db:"-" gorm:"embedded;embeddedPrefix:intent_"`
	CreatedBy      uuid.UUID       `json:"createdBy" db:"created_by" gorm:"type:uuid;not null"`
	CreatedAt      time.Time       `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// ActivityIntent represents the parsed intent of an activity request
type ActivityIntent struct {
	Fingerprint string `json:"fingerprint" db:"fingerprint"`
	Summary     string `json:"summary" db:"summary"`
}

// Proof represents cryptographic proofs generated by the system
type Proof struct {
	ID         uuid.UUID `json:"id" db:"id" gorm:"type:uuid;primary_key"`
	ActivityID uuid.UUID `json:"activityId" db:"activity_id" gorm:"type:uuid;not null"`
	Type       string    `json:"type" db:"type" gorm:"not null"` // SIGNATURE, POLICY_OUTCOME, etc.
	Data       string    `json:"data" db:"data" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
}

// ToJSON serializes the organization to JSON for storage/transmission
func (o *Organization) ToJSON() ([]byte, error) {
	return json.Marshal(o)
}

// FromJSON deserializes JSON data into an organization
func (o *Organization) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

// TableName returns the table name for GORM
func (Organization) TableName() string {
	return "organizations"
}

func (User) TableName() string {
	return "users"
}

func (AuthMethod) TableName() string {
	return "auth_methods"
}

func (Invitation) TableName() string {
	return "invitations"
}

func (Policy) TableName() string {
	return "policies"
}

func (Tag) TableName() string {
	return "tags"
}

func (PrivateKey) TableName() string {
	return "private_keys"
}

func (Wallet) TableName() string {
	return "wallets"
}

func (WalletAccount) TableName() string {
	return "wallet_accounts"
}

func (Activity) TableName() string {
	return "activities"
}

func (Proof) TableName() string {
	return "proofs"
}
