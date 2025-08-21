package db

import (
	"context"
	"fmt"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	// Configure GORM logger based on environment
	var gormLogger logger.Interface
	if cfg.Database == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresDB{
		db:     db,
		config: cfg,
	}, nil
}

// GetDB returns the GORM database instance
func (p *PostgresDB) GetDB() *gorm.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Health checks the database connection health
func (p *PostgresDB) Health(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// GetStats returns database connection statistics
func (p *PostgresDB) GetStats() map[string]interface{} {
	sqlDB, err := p.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction executes a function within a database transaction
func (p *PostgresDB) Transaction(fn func(*gorm.DB) error) error {
	return p.db.Transaction(fn)
}

// BeginTx starts a new transaction with context
func (p *PostgresDB) BeginTx(ctx context.Context) *gorm.DB {
	return p.db.WithContext(ctx).Begin()
}

// Migrate runs database migrations
func (p *PostgresDB) Migrate() error {
	// Auto-migrate would be handled by migration files
	// This is a placeholder for any additional migration logic
	return nil
}

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetDB() *gorm.DB
	Close() error
	Health(ctx context.Context) error
	GetStats() map[string]interface{}
	Transaction(fn func(*gorm.DB) error) error
	BeginTx(ctx context.Context) *gorm.DB
}
