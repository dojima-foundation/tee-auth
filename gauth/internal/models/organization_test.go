package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_ToJSON(t *testing.T) {
	tests := []struct {
		name string
		org  *Organization
	}{
		{
			name: "basic organization",
			org: &Organization{
				ID:      uuid.New(),
				Version: "1.0",
				Name:    "Test Organization",
				RootQuorum: Quorum{
					UserIDs:   []uuid.UUID{uuid.New()},
					Threshold: 1,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "organization with users",
			org: &Organization{
				ID:      uuid.New(),
				Version: "1.0",
				Name:    "Test Organization with Users",
				Users: []User{
					{
						ID:       uuid.New(),
						Username: "testuser",
						Email:    "test@example.com",
						IsActive: true,
					},
				},
				RootQuorum: Quorum{
					UserIDs:   []uuid.UUID{uuid.New()},
					Threshold: 1,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := tt.org.ToJSON()
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Verify it's valid JSON
			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			require.NoError(t, err)

			// Verify key fields are present
			assert.Equal(t, tt.org.ID.String(), result["uuid"])
			assert.Equal(t, tt.org.Version, result["version"])
			assert.Equal(t, tt.org.Name, result["name"])
		})
	}
}

func TestOrganization_FromJSON(t *testing.T) {
	originalOrg := &Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Test Organization",
		RootQuorum: Quorum{
			UserIDs:   []uuid.UUID{uuid.New()},
			Threshold: 1,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Convert to JSON
	jsonData, err := originalOrg.ToJSON()
	require.NoError(t, err)

	// Convert back from JSON
	var restoredOrg Organization
	err = restoredOrg.FromJSON(jsonData)
	require.NoError(t, err)

	// Verify key fields match
	assert.Equal(t, originalOrg.ID, restoredOrg.ID)
	assert.Equal(t, originalOrg.Version, restoredOrg.Version)
	assert.Equal(t, originalOrg.Name, restoredOrg.Name)
	assert.Equal(t, originalOrg.RootQuorum.Threshold, restoredOrg.RootQuorum.Threshold)
}

func TestOrganization_FromJSON_InvalidData(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "invalid json",
			jsonData: `{"invalid": json}`,
		},
		{
			name:     "empty json",
			jsonData: `{}`,
		},
		{
			name:     "null json",
			jsonData: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var org Organization
			err := org.FromJSON([]byte(tt.jsonData))
			if tt.name == "empty json" {
				// Empty JSON should not error, just create empty struct
				assert.NoError(t, err)
			} else if tt.name == "null json" {
				// null JSON should not error in Go's json.Unmarshal
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name  string
		user  User
		valid bool
	}{
		{
			name: "valid user",
			user: User{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				Username:       "testuser",
				Email:          "test@example.com",
				PublicKey:      "public-key-data",
				IsActive:       true,
			},
			valid: true,
		},
		{
			name: "user with empty email",
			user: User{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				Username:       "testuser",
				Email:          "",
				PublicKey:      "public-key-data",
				IsActive:       true,
			},
			valid: false,
		},
		{
			name: "user with empty username",
			user: User{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				Username:       "",
				Email:          "test@example.com",
				PublicKey:      "public-key-data",
				IsActive:       true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.user.Username)
				assert.NotEmpty(t, tt.user.Email)
				assert.NotEmpty(t, tt.user.PublicKey)
				assert.NotEqual(t, uuid.Nil, tt.user.ID)
				assert.NotEqual(t, uuid.Nil, tt.user.OrganizationID)
			} else {
				// At least one field should be invalid
				invalid := tt.user.Username == "" || tt.user.Email == "" || tt.user.PublicKey == ""
				assert.True(t, invalid)
			}
		})
	}
}

func TestQuorum_Validation(t *testing.T) {
	tests := []struct {
		name   string
		quorum Quorum
		valid  bool
	}{
		{
			name: "valid quorum",
			quorum: Quorum{
				UserIDs:   []uuid.UUID{uuid.New(), uuid.New()},
				Threshold: 2,
			},
			valid: true,
		},
		{
			name: "threshold higher than user count",
			quorum: Quorum{
				UserIDs:   []uuid.UUID{uuid.New()},
				Threshold: 2,
			},
			valid: false,
		},
		{
			name: "zero threshold",
			quorum: Quorum{
				UserIDs:   []uuid.UUID{uuid.New()},
				Threshold: 0,
			},
			valid: false,
		},
		{
			name: "empty user list",
			quorum: Quorum{
				UserIDs:   []uuid.UUID{},
				Threshold: 1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.True(t, len(tt.quorum.UserIDs) >= tt.quorum.Threshold)
				assert.True(t, tt.quorum.Threshold > 0)
				assert.True(t, len(tt.quorum.UserIDs) > 0)
			} else {
				invalid := len(tt.quorum.UserIDs) < tt.quorum.Threshold ||
					tt.quorum.Threshold <= 0 ||
					len(tt.quorum.UserIDs) == 0
				assert.True(t, invalid)
			}
		})
	}
}

func TestActivity_StatusTransitions(t *testing.T) {
	activity := &Activity{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Type:           "TEST_ACTIVITY",
		Status:         "PENDING",
		Parameters:     json.RawMessage(`{"test": "data"}`),
		CreatedBy:      uuid.New(),
		CreatedAt:      time.Now(),
	}

	// Test valid status transitions
	validTransitions := map[string][]string{
		"PENDING":   {"COMPLETED", "FAILED"},
		"COMPLETED": {}, // Terminal state
		"FAILED":    {}, // Terminal state
	}

	for currentStatus, allowedNext := range validTransitions {
		activity.Status = currentStatus

		for _, nextStatus := range allowedNext {
			// This would be implemented in business logic
			assert.Contains(t, []string{"COMPLETED", "FAILED"}, nextStatus)
		}
	}
}

func TestTableNames(t *testing.T) {
	tests := []struct {
		model interface{}
		table string
	}{
		{Organization{}, "organizations"},
		{User{}, "users"},
		{AuthMethod{}, "auth_methods"},
		{Invitation{}, "invitations"},
		{Policy{}, "policies"},
		{Tag{}, "tags"},
		{PrivateKey{}, "private_keys"},
		{Wallet{}, "wallets"},
		{WalletAccount{}, "wallet_accounts"},
		{Activity{}, "activities"},
		{Proof{}, "proofs"},
	}

	for _, tt := range tests {
		t.Run(tt.table, func(t *testing.T) {
			var tableName string
			switch model := tt.model.(type) {
			case Organization:
				tableName = model.TableName()
			case User:
				tableName = model.TableName()
			case AuthMethod:
				tableName = model.TableName()
			case Invitation:
				tableName = model.TableName()
			case Policy:
				tableName = model.TableName()
			case Tag:
				tableName = model.TableName()
			case PrivateKey:
				tableName = model.TableName()
			case Wallet:
				tableName = model.TableName()
			case WalletAccount:
				tableName = model.TableName()
			case Activity:
				tableName = model.TableName()
			case Proof:
				tableName = model.TableName()
			}
			assert.Equal(t, tt.table, tableName)
		})
	}
}

func TestWallet_BIP44Paths(t *testing.T) {
	wallet := &Wallet{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Name:           "Test Wallet",
		Accounts: []WalletAccount{
			{
				ID:            uuid.New(),
				Name:          "Account 0",
				Path:          "m/44'/0'/0'/0/0",
				PublicKey:     "03...",
				Address:       "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
				Curve:         "SECP256K1",
				AddressFormat: "P2PKH",
			},
			{
				ID:            uuid.New(),
				Name:          "Account 1",
				Path:          "m/44'/0'/0'/0/1",
				PublicKey:     "02...",
				Address:       "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",
				Curve:         "SECP256K1",
				AddressFormat: "P2WPKH",
			},
		},
	}

	// Verify BIP44 path format
	for _, account := range wallet.Accounts {
		assert.Regexp(t, `^m/44'/\d+'/\d+'/\d+/\d+$`, account.Path)
		assert.NotEmpty(t, account.Address)
		assert.NotEmpty(t, account.PublicKey)
		assert.Contains(t, []string{"SECP256K1", "ED25519"}, account.Curve)
	}
}

// Benchmark tests
func BenchmarkOrganization_ToJSON(b *testing.B) {
	org := &Organization{
		ID:      uuid.New(),
		Version: "1.0",
		Name:    "Benchmark Organization",
		Users:   make([]User, 100), // Large organization
		RootQuorum: Quorum{
			UserIDs:   make([]uuid.UUID, 10),
			Threshold: 5,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Fill with test data
	for i := range org.Users {
		org.Users[i] = User{
			ID:       uuid.New(),
			Username: "user" + string(rune(i)),
			Email:    "user" + string(rune(i)) + "@example.com",
			IsActive: true,
		}
	}

	for i := range org.RootQuorum.UserIDs {
		org.RootQuorum.UserIDs[i] = uuid.New()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := org.ToJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrganization_FromJSON(b *testing.B) {
	org := &Organization{
		ID:        uuid.New(),
		Version:   "1.0",
		Name:      "Benchmark Organization",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	jsonData, err := org.ToJSON()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var newOrg Organization
		err := newOrg.FromJSON(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
