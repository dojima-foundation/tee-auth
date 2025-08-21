package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	grpcServer "github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
)

const (
	serviceName    = "gauth"
	serviceVersion = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	lgr, err := logger.New(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	lgr.Info("Starting gauth service",
		"service", serviceName,
		"version", serviceVersion,
		"environment", os.Getenv("ENVIRONMENT"),
	)

	// Initialize database
	database, err := db.NewPostgresDB(&cfg.Database)
	if err != nil {
		lgr.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			lgr.Error("Failed to close database connection", "error", err)
		}
	}()

	lgr.Info("Database connection established",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.Database,
	)

	// Initialize Redis
	redis, err := db.NewRedisClient(&cfg.Redis)
	if err != nil {
		lgr.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := redis.Close(); err != nil {
			lgr.Error("Failed to close Redis connection", "error", err)
		}
	}()

	lgr.Info("Redis connection established",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"database", cfg.Redis.Database,
	)

	// Initialize service layer
	svc := service.NewGAuthService(cfg, lgr, database, redis)

	// Initialize gRPC server
	grpcSrv := grpcServer.NewServer(cfg, lgr, svc)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	var wg sync.WaitGroup

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		lgr.Info("Starting gRPC server", "address", cfg.GetGRPCAddr())

		if err := grpcSrv.Start(); err != nil {
			lgr.Error("gRPC server failed", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		lgr.Info("Received shutdown signal", "signal", sig.String())
	case <-ctx.Done():
		lgr.Info("Context cancelled, shutting down")
	}

	// Graceful shutdown
	lgr.Info("Starting graceful shutdown...")

	// Stop gRPC server
	grpcSrv.Stop()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lgr.Info("Graceful shutdown completed")
	case <-time.After(30 * time.Second):
		lgr.Warn("Shutdown timeout exceeded, forcing exit")
	}

	lgr.Info("gauth service stopped")
}

// healthCheck performs a basic health check of all dependencies
func healthCheck(ctx context.Context, database db.DatabaseInterface, redis db.RedisInterface, lgr *logger.Logger) error {
	// Check database
	if err := database.Health(ctx); err != nil {
		lgr.Error("Database health check failed", "error", err)
		return fmt.Errorf("database unhealthy: %w", err)
	}

	// Check Redis
	if err := redis.Health(ctx); err != nil {
		lgr.Error("Redis health check failed", "error", err)
		return fmt.Errorf("redis unhealthy: %w", err)
	}

	lgr.Info("All health checks passed")
	return nil
}

// printBanner prints the service banner
func printBanner() {
	banner := `
   ____                 _   _     
  / ___| __ _ _   _ _ __| |_| |__  
 | |  _ / _` + "`" + ` | | | | '__| __| '_ \ 
 | |_| | (_| | |_| | |  | |_| | | |
  \____|\__,_|\__,_|_|   \__|_| |_|
                                   
  Go Authentication Service v%s
  Turnkey-inspired Architecture
  
`
	fmt.Printf(banner, serviceVersion)
}

func init() {
	printBanner()
}
