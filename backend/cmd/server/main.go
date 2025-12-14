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

	"home-run-backend/internal/api"
	"home-run-backend/internal/config"
	"home-run-backend/internal/logger"
	"home-run-backend/internal/services"
	"home-run-backend/internal/services/federation"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	logger.Log.Infof("Loading configuration from %s", *configPath)
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize service manager
	logger.Log.Info("Initializing service manager...")
	manager, err := services.NewManager(cfg)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize service manager: %v", err)
	}

	// Start background processes
	manager.Start(ctx)
	defer manager.Stop()

	// Initialize federation aggregator
	aggregator := federation.NewAggregator(manager, cfg.RemoteHosts)

	// Setup router
	router := api.SetupRouter(cfg, manager, aggregator)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Log.Infof("Starting server on %s", addr)
		logger.Log.Infof("CORS allowed origin: %s", cfg.Server.CORSAllowOrigin)
		logger.Log.Infof("Monitoring %d services", len(cfg.Services))
		logger.Log.Infof("Federated with %d remote hosts", len(cfg.RemoteHosts))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Log.Info("Server exited")
}
