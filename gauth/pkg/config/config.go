package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the gauth service
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// Redis configuration
	Redis RedisConfig `yaml:"redis"`

	// gRPC configuration
	GRPC GRPCConfig `yaml:"grpc"`

	// Renclave configuration
	Renclave RenclaveConfig `yaml:"renclave"`

	// Authentication configuration
	Auth AuthConfig `yaml:"auth"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`

	// Security configuration
	Security SecurityConfig `yaml:"security"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type DatabaseConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Username     string        `yaml:"username"`
	Password     string        `yaml:"password"`
	Database     string        `yaml:"database"`
	SSLMode      string        `yaml:"ssl_mode"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
}

type RedisConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Password     string        `yaml:"password"`
	Database     int           `yaml:"database"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type GRPCConfig struct {
	Host                string        `yaml:"host"`
	Port                int           `yaml:"port"`
	MaxRecvMsgSize      int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize      int           `yaml:"max_send_msg_size"`
	ConnectionTimeout   time.Duration `yaml:"connection_timeout"`
	KeepAliveTime       time.Duration `yaml:"keep_alive_time"`
	KeepAliveTimeout    time.Duration `yaml:"keep_alive_timeout"`
	PermitWithoutStream bool          `yaml:"permit_without_stream"`
}

type RenclaveConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	UseTLS  bool          `yaml:"use_tls"`
	Timeout time.Duration `yaml:"timeout"`
}

type AuthConfig struct {
	JWTSecret              string        `yaml:"jwt_secret"`
	JWTExpiration          time.Duration `yaml:"jwt_expiration"`
	RefreshExpiration      time.Duration `yaml:"refresh_expiration"`
	SessionTimeout         time.Duration `yaml:"session_timeout"`
	MaxLoginAttempts       int           `yaml:"max_login_attempts"`
	LockoutDuration        time.Duration `yaml:"lockout_duration"`
	RequireQuorum          bool          `yaml:"require_quorum"`
	DefaultQuorumThreshold int           `yaml:"default_quorum_threshold"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"` // json, text
	Output     string `yaml:"output"` // stdout, stderr, file
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

type SecurityConfig struct {
	TLSEnabled       bool     `yaml:"tls_enabled"`
	TLSCertFile      string   `yaml:"tls_cert_file"`
	TLSKeyFile       string   `yaml:"tls_key_file"`
	CORSEnabled      bool     `yaml:"cors_enabled"`
	CORSOrigins      []string `yaml:"cors_origins"`
	RateLimitEnabled bool     `yaml:"rate_limit_enabled"`
	RateLimitRPS     int      `yaml:"rate_limit_rps"`
	RateLimitBurst   int      `yaml:"rate_limit_burst"`
	EncryptionKey    string   `yaml:"encryption_key"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			Username:     getEnv("DB_USERNAME", "gauth"),
			Password:     getEnv("DB_PASSWORD", "password"),
			Database:     getEnv("DB_DATABASE", "gauth"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvDuration("DB_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getEnvInt("REDIS_DATABASE", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		GRPC: GRPCConfig{
			Host:                getEnv("GRPC_HOST", "0.0.0.0"),
			Port:                getEnvInt("GRPC_PORT", 9090),
			MaxRecvMsgSize:      getEnvInt("GRPC_MAX_RECV_MSG_SIZE", 4*1024*1024), // 4MB
			MaxSendMsgSize:      getEnvInt("GRPC_MAX_SEND_MSG_SIZE", 4*1024*1024), // 4MB
			ConnectionTimeout:   getEnvDuration("GRPC_CONNECTION_TIMEOUT", 10*time.Second),
			KeepAliveTime:       getEnvDuration("GRPC_KEEP_ALIVE_TIME", 30*time.Second),
			KeepAliveTimeout:    getEnvDuration("GRPC_KEEP_ALIVE_TIMEOUT", 5*time.Second),
			PermitWithoutStream: getEnvBool("GRPC_PERMIT_WITHOUT_STREAM", true),
		},
		Renclave: RenclaveConfig{
			Host:    getEnv("RENCLAVE_HOST", "localhost"),
			Port:    getEnvInt("RENCLAVE_PORT", 3000),
			UseTLS:  getEnvBool("RENCLAVE_USE_TLS", false),
			Timeout: getEnvDuration("RENCLAVE_TIMEOUT", 30*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:              getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			JWTExpiration:          getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			RefreshExpiration:      getEnvDuration("REFRESH_EXPIRATION", 7*24*time.Hour),
			SessionTimeout:         getEnvDuration("SESSION_TIMEOUT", 30*time.Minute),
			MaxLoginAttempts:       getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:        getEnvDuration("LOCKOUT_DURATION", 15*time.Minute),
			RequireQuorum:          getEnvBool("REQUIRE_QUORUM", false),
			DefaultQuorumThreshold: getEnvInt("DEFAULT_QUORUM_THRESHOLD", 1),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			Filename:   getEnv("LOG_FILENAME", "gauth.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
		Security: SecurityConfig{
			TLSEnabled:       getEnvBool("TLS_ENABLED", false),
			TLSCertFile:      getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:       getEnv("TLS_KEY_FILE", ""),
			CORSEnabled:      getEnvBool("CORS_ENABLED", true),
			CORSOrigins:      getEnvStringSlice("CORS_ORIGINS", []string{"*"}),
			RateLimitEnabled: getEnvBool("RATE_LIMIT_ENABLED", true),
			RateLimitRPS:     getEnvInt("RATE_LIMIT_RPS", 100),
			RateLimitBurst:   getEnvInt("RATE_LIMIT_BURST", 200),
			EncryptionKey:    getEnv("ENCRYPTION_KEY", "your-encryption-key-32-bytes-long"),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret must be set to a secure value")
	}
	if len(c.Security.EncryptionKey) != 32 {
		return fmt.Errorf("encryption key must be exactly 32 bytes long")
	}
	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetGRPCAddr returns the gRPC server address
func (c *Config) GetGRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.GRPC.Host, c.GRPC.Port)
}

// GetRenclaveAddr returns the renclave service address
func (c *Config) GetRenclaveAddr() string {
	protocol := "http"
	if c.Renclave.UseTLS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, c.Renclave.Host, c.Renclave.Port)
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// For production, consider using a more robust parser
		return []string{value}
	}
	return defaultValue
}
