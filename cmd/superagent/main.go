package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/router"
)

var (
	configFile = flag.String("config", "", "Path to configuration file (YAML)")
	version    = flag.Bool("version", false, "Show version information")
	help       = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: "8080",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	r := router.SetupRouter(cfg)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		}).Info("Starting SuperAgent server with Models.dev integration")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Use r variable to avoid unused import
	_ = r
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown complete")
	}
}

func showHelp() {
	fmt.Printf(`SuperAgent - Advanced LLM Gateway with Models.dev Integration

Usage:
  superagent [options]

Options:
  -config string
        Path to configuration file (YAML)
  -version
        Show version information  
  -help
        Show this help message

Features:
  - Models.dev integration for comprehensive model metadata
  - Multi-layer caching with Redis and in-memory
  - Circuit breaker for API resilience
  - Auto-refresh with configurable intervals
  - Model comparison and capability filtering
  - Comprehensive monitoring and health checks

Examples:
  superagent
  superagent -config /path/to/config.yaml
  superagent -version

For more information, visit: https://github.com/superagent/superagent
`)
}

func showVersion() {
	fmt.Printf("SuperAgent v%s - Models.dev Enhanced Edition\n", "1.0.0")
}
