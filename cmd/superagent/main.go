package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/router"
	"github.com/superagent/superagent/pkg/metrics"
)

var (
	configFile = flag.String("config", "", "Path to configuration file (YAML)")
	version    = flag.Bool("version", false, "Show version information")
	help       = flag.Bool("help", false, "Show help information")
)

const (
	AppName    = "SuperAgent"
	AppVersion = "1.0.0"
	AppDesc    = "Advanced AI Agent Platform with Multi-Provider Support"
)

func main() {
	flag.Parse()

	if *version {
		printVersion()
		return
	}

	if *help {
		printHelp()
		return
	}

	// Initialize logger
	logger := logrus.New()

	// Load configuration
	cfg, err := loadConfiguration(*configFile, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Configure logger based on config
	configureLogger(logger, cfg)

	logger.WithFields(logrus.Fields{
		"version": AppVersion,
		"port":    cfg.Server.Port,
		"mode":    cfg.Server.Mode,
	}).Info("Starting SuperAgent")

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize metrics if enabled
	var metricsServer *http.Server
	if cfg.Monitoring.Enabled {
		metricsServer = startMetricsServer(cfg, logger)
	}

	// Setup main router with all services
	appRouter := router.SetupRouter(cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      appRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Performance.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		}).Info("SuperAgent HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down SuperAgent...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown metrics server if running
	if metricsServer != nil {
		logger.Info("Shutting down metrics server...")
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("Error shutting down metrics server")
		}
	}

	// Shutdown main HTTP server
	logger.Info("Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error shutting down HTTP server")
	}

	logger.Info("SuperAgent shutdown complete")
}

// loadConfiguration loads configuration from environment variables
func loadConfiguration(configFile string, logger *logrus.Logger) (*config.Config, error) {
	// Load configuration from environment variables
	cfg := config.Load()

	// Log configuration loading
	if configFile != "" {
		logger.WithField("file", configFile).Info("Configuration file specified but YAML parsing not available, using environment variables")
	} else {
		logger.Info("Loading configuration from environment variables")
	}

	// Log key configuration values (without sensitive data)
	logger.WithFields(logrus.Fields{
		"server_port":     cfg.Server.Port,
		"server_host":     cfg.Server.Host,
		"server_mode":     cfg.Server.Mode,
		"database_host":   cfg.Database.Host,
		"database_port":   cfg.Database.Port,
		"database_name":   cfg.Database.Name,
		"redis_host":      cfg.Redis.Host,
		"redis_port":      cfg.Redis.Port,
		"metrics_enabled": cfg.Monitoring.Enabled,
		"debug_enabled":   cfg.Server.DebugEnabled,
	}).Info("Configuration loaded")

	// Validate configuration
	if err := validateConfiguration(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validateConfiguration validates the loaded configuration
func validateConfiguration(cfg *config.Config) error {
	// Validate server configuration
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.Server.JWTSecret == "" || cfg.Server.JWTSecret == "development-jwt-secret-key-change-in-production" {
		if cfg.Server.Mode == "release" {
			return fmt.Errorf("JWT secret must be set in production mode")
		}
	}

	// Validate database configuration
	if cfg.Database.Host == "" || cfg.Database.Port == "" || cfg.Database.User == "" || cfg.Database.Name == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	// Validate at least one LLM provider is configured
	if len(cfg.LLM.Providers) == 0 {
		return fmt.Errorf("at least one LLM provider must be configured")
	}

	// Check if any provider is enabled
	hasEnabledProvider := false
	for _, provider := range cfg.LLM.Providers {
		if provider.Enabled {
			hasEnabledProvider = true
			break
		}
	}
	if !hasEnabledProvider {
		return fmt.Errorf("at least one LLM provider must be enabled")
	}

	return nil
}

// configureLogger configures the logger based on configuration
func configureLogger(logger *logrus.Logger, cfg *config.Config) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Monitoring.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter based on mode
	if cfg.Server.Mode == "debug" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// Enable debug mode if requested
	if cfg.Server.DebugEnabled {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Debug mode enabled")
	}
}

// startMetricsServer starts the Prometheus metrics server
func startMetricsServer(cfg *config.Config, logger *logrus.Logger) *http.Server {
	metricsAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Monitoring.Prometheus.Port)

	mux := http.NewServeMux()
	mux.Handle(cfg.Monitoring.Prometheus.Path, metrics.Handler())

	server := &http.Server{
		Addr:         metricsAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.WithFields(logrus.Fields{
			"port": cfg.Monitoring.Prometheus.Port,
			"path": cfg.Monitoring.Prometheus.Path,
		}).Info("Starting metrics server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Failed to start metrics server")
		}
	}()

	return server
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("%s v%s\n", AppName, AppVersion)
	fmt.Printf("%s\n", AppDesc)
	fmt.Printf("Built with Go %s\n", "1.23+")
}

// printHelp prints help information
func printHelp() {
	fmt.Printf("%s - %s\n\n", AppName, AppDesc)
	fmt.Printf("Usage: %s [options]\n\n", filepath.Base(os.Args[0]))
	fmt.Printf("Options:\n")
	fmt.Printf("  -config string    Path to configuration file (YAML)\n")
	fmt.Printf("  -version          Show version information\n")
	fmt.Printf("  -help             Show this help message\n\n")
	fmt.Printf("Environment Variables:\n")
	fmt.Printf("  PORT              Server port (default: 8080)\n")
	fmt.Printf("  SERVER_HOST       Server host (default: 0.0.0.0)\n")
	fmt.Printf("  GIN_MODE          Gin mode (debug/release, default: release)\n")
	fmt.Printf("  LOG_LEVEL         Log level (debug/info/warn/error, default: info)\n")
	fmt.Printf("  DB_HOST           Database host (default: localhost)\n")
	fmt.Printf("  DB_PORT           Database port (default: 5432)\n")
	fmt.Printf("  DB_USER           Database user (default: superagent)\n")
	fmt.Printf("  DB_PASSWORD       Database password\n")
	fmt.Printf("  DB_NAME           Database name (default: superagent_db)\n")
	fmt.Printf("  REDIS_HOST        Redis host (default: localhost)\n")
	fmt.Printf("  REDIS_PORT        Redis port (default: 6379)\n")
	fmt.Printf("  JWT_SECRET        JWT secret key\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Configuration is loaded from environment variables\n")
	fmt.Printf("  Copy .env.example to .env and configure as needed\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  %s                           # Start with default configuration\n", filepath.Base(os.Args[0]))
	fmt.Printf("  %s -version                 # Show version\n", filepath.Base(os.Args[0]))
	fmt.Printf("  PORT=9000 %s                # Start on port 9000\n", filepath.Base(os.Args[0]))
}
