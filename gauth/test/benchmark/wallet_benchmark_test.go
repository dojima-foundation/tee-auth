package benchmark

import (
	"context"
	"fmt"
	"testing"

	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/test/testhelpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// BenchmarkWalletCreation measures wallet creation performance
func BenchmarkWalletCreation(b *testing.B) {
	// Setup
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	testLogger := logger.NewLogger(&logger.Config{
		Level:  "error", // Reduce logging for benchmarks
		Format: "text",
	})

	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	service := service.NewGAuthService(cfg, testLogger, testDB, nil)
	ctx := context.Background()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Benchmark Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	organizationID := org.ID.String()

	// Benchmark different wallet configurations
	benchmarks := []struct {
		name         string
		accountCount int
		mnemonicLen  int32
	}{
		{"SingleAccount_12Words", 1, 12},
		{"SingleAccount_24Words", 1, 24},
		{"ThreeAccounts_12Words", 3, 12},
		{"ThreeAccounts_24Words", 3, 24},
		{"FiveAccounts_12Words", 5, 12},
		{"TenAccounts_24Words", 10, 24},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			accounts := make([]models.WalletAccount, bm.accountCount)
			for i := 0; i < bm.accountCount; i++ {
				accounts[i] = models.WalletAccount{
					Curve:         "CURVE_SECP256K1",
					Path:          fmt.Sprintf("m/44'/60'/0'/0/%d", i),
					AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _, err := service.CreateWallet(
					ctx,
					organizationID,
					fmt.Sprintf("Benchmark Wallet %d", i),
					accounts,
					&bm.mnemonicLen,
					[]string{"benchmark"},
				)
				if err != nil {
					b.Fatalf("Wallet creation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPrivateKeyCreation measures private key creation performance
func BenchmarkPrivateKeyCreation(b *testing.B) {
	// Setup
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	testLogger := logger.NewLogger(&logger.Config{
		Level:  "error",
		Format: "text",
	})

	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	service := service.NewGAuthService(cfg, testLogger, testDB, nil)
	ctx := context.Background()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Benchmark Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	organizationID := org.ID.String()

	// Benchmark different curves
	curves := []string{"CURVE_SECP256K1", "CURVE_ED25519"}

	for _, curve := range curves {
		b.Run(curve, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := service.CreatePrivateKey(
					ctx,
					organizationID,
					fmt.Sprintf("Benchmark Key %d", i),
					curve,
					nil, // Generate new key
					[]string{"benchmark"},
				)
				if err != nil {
					b.Fatalf("Private key creation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkWalletRetrieval measures wallet retrieval performance
func BenchmarkWalletRetrieval(b *testing.B) {
	// Setup
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	testLogger := logger.NewLogger(&logger.Config{
		Level:  "error",
		Format: "text",
	})

	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	service := service.NewGAuthService(cfg, testLogger, testDB, nil)
	ctx := context.Background()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Benchmark Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	organizationID := org.ID.String()

	// Pre-create wallets for retrieval benchmarks
	walletCount := 100
	walletIDs := make([]string, walletCount)

	for i := 0; i < walletCount; i++ {
		accounts := []models.WalletAccount{
			{
				Curve:         "CURVE_SECP256K1",
				Path:          "m/44'/60'/0'/0/0",
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			},
		}

		wallet, _, err := service.CreateWallet(
			ctx,
			organizationID,
			fmt.Sprintf("Benchmark Wallet %d", i),
			accounts,
			nil,
			[]string{"benchmark"},
		)
		require.NoError(b, err)
		walletIDs[i] = wallet.ID.String()
	}

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark individual wallet retrieval
	b.Run("GetWallet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			walletID := walletIDs[i%len(walletIDs)]
			_, err := service.GetWallet(ctx, walletID)
			if err != nil {
				b.Fatalf("Wallet retrieval failed: %v", err)
			}
		}
	})

	// Benchmark wallet listing with pagination
	b.Run("ListWallets_Page10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := service.ListWallets(ctx, organizationID, 10, "")
			if err != nil {
				b.Fatalf("Wallet listing failed: %v", err)
			}
		}
	})

	b.Run("ListWallets_Page50", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := service.ListWallets(ctx, organizationID, 50, "")
			if err != nil {
				b.Fatalf("Wallet listing failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentWalletCreation measures concurrent wallet creation performance
func BenchmarkConcurrentWalletCreation(b *testing.B) {
	// Setup
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	testLogger := logger.NewLogger(&logger.Config{
		Level:  "error",
		Format: "text",
	})

	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	service := service.NewGAuthService(cfg, testLogger, testDB, nil)
	ctx := context.Background()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Benchmark Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	organizationID := org.ID.String()

	accounts := []models.WalletAccount{
		{
			Curve:         "CURVE_SECP256K1",
			Path:          "m/44'/60'/0'/0/0",
			AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
		},
	}

	mnemonicLen := int32(12)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _, err := service.CreateWallet(
				ctx,
				organizationID,
				fmt.Sprintf("Concurrent Wallet %d", i),
				accounts,
				&mnemonicLen,
				[]string{"concurrent"},
			)
			if err != nil {
				b.Fatalf("Concurrent wallet creation failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkMemoryUsage measures memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	// Setup
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	testLogger := logger.NewLogger(&logger.Config{
		Level:  "error",
		Format: "text",
	})

	cfg := &config.Config{
		Auth: config.AuthConfig{
			DefaultQuorumThreshold: 1,
		},
	}

	service := service.NewGAuthService(cfg, testLogger, testDB, nil)
	ctx := context.Background()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "Memory Test Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	organizationID := org.ID.String()

	b.Run("LargeWalletCreation", func(b *testing.B) {
		// Create wallet with many accounts
		accounts := make([]models.WalletAccount, 50)
		for i := 0; i < 50; i++ {
			accounts[i] = models.WalletAccount{
				Curve:         "CURVE_SECP256K1",
				Path:          fmt.Sprintf("m/44'/60'/0'/0/%d", i),
				AddressFormat: "ADDRESS_FORMAT_ETHEREUM",
			}
		}

		mnemonicLen := int32(24)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _, err := service.CreateWallet(
				ctx,
				organizationID,
				fmt.Sprintf("Large Wallet %d", i),
				accounts,
				&mnemonicLen,
				[]string{"large", "memory-test"},
			)
			if err != nil {
				b.Fatalf("Large wallet creation failed: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseOperations measures raw database performance
func BenchmarkDatabaseOperations(b *testing.B) {
	testDB := testhelpers.SetupTestDB(b)
	defer testDB.Cleanup()

	// Create test organization
	org := &models.Organization{
		ID:      uuid.New(),
		Name:    "DB Benchmark Org",
		Version: "1.0",
	}
	err := testDB.Create(org).Error
	require.NoError(b, err)

	b.Run("WalletInsert", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			wallet := &models.Wallet{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				Name:           fmt.Sprintf("DB Wallet %d", i),
				PublicKey:      fmt.Sprintf("pub_%d", i),
				Tags:           []string{"db-test"},
				IsActive:       true,
			}

			err := testDB.Create(wallet).Error
			if err != nil {
				b.Fatalf("Wallet insert failed: %v", err)
			}
		}
	})

	b.Run("PrivateKeyInsert", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			privateKey := &models.PrivateKey{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				Name:           fmt.Sprintf("DB Key %d", i),
				PublicKey:      fmt.Sprintf("pub_%d", i),
				Curve:          "CURVE_SECP256K1",
				Tags:           []string{"db-test"},
				IsActive:       true,
			}

			err := testDB.Create(privateKey).Error
			if err != nil {
				b.Fatalf("Private key insert failed: %v", err)
			}
		}
	})
}
