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

	restServer "github.com/dojima-foundation/tee-auth/gauth/api/rest"
	"github.com/dojima-foundation/tee-auth/gauth/internal/db"
	grpcServer "github.com/dojima-foundation/tee-auth/gauth/internal/grpc"
	"github.com/dojima-foundation/tee-auth/gauth/internal/service"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/config"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/logger"
	"github.com/dojima-foundation/tee-auth/gauth/pkg/telemetry"
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

	// Initialize telemetry
	ctx := context.Background()
	tel, err := telemetry.New(ctx, telemetry.Config{
		ServiceName:            serviceName,
		ServiceVersion:         serviceVersion,
		Environment:            os.Getenv("ENVIRONMENT"),
		TracingEnabled:         cfg.Telemetry.TracingEnabled,
		MetricsEnabled:         cfg.Telemetry.MetricsEnabled,
		OTLPEndpoint:           cfg.Telemetry.OTLPEndpoint,
		OTLPInsecure:           cfg.Telemetry.OTLPInsecure,
		TraceSamplingRatio:     cfg.Telemetry.TraceSamplingRatio,
		MetricsReportingPeriod: time.Second * 30,
		MetricsPort:            cfg.Telemetry.MetricsPort,
	})
	if err != nil {
		lgr.Error("Failed to initialize telemetry", "error", err)
		os.Exit(1)
	}

	lgr.Info("Telemetry initialized",
		"tracing_enabled", cfg.Telemetry.TracingEnabled,
		"metrics_enabled", cfg.Telemetry.MetricsEnabled,
		"otlp_endpoint", cfg.Telemetry.OTLPEndpoint,
	)

	// Initialize service layer
	svc := service.NewGAuthService(cfg, lgr, database, redis)

	// Initialize servers
	grpcSrv := grpcServer.NewServer(cfg, lgr, svc, tel)
	restSrv := restServer.NewServer(cfg, lgr, tel)

	// Set Redis interface for session management
	restSrv.SetRedis(redis)

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

	// Start REST API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Wait a moment for gRPC server to start first
		time.Sleep(1 * time.Second)

		lgr.Info("Starting REST API server", "address", cfg.GetServerAddr())

		if err := restSrv.Start(); err != nil {
			lgr.Error("REST API server failed", "error", err)
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

	// Stop servers
	grpcSrv.Stop()
	if err := restSrv.Stop(); err != nil {
		lgr.Error("Failed to stop REST API server", "error", err)
	}

	// Shutdown telemetry
	if tel != nil && tel.Shutdown != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := tel.Shutdown(shutdownCtx); err != nil {
			lgr.Error("Failed to shutdown telemetry", "error", err)
		} else {
			lgr.Info("Telemetry shutdown completed")
		}
	}

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
