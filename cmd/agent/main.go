package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nilbyte-studios/network-infra/internal/config"
	"github.com/nilbyte-studios/network-infra/internal/metrics"
	"github.com/nilbyte-studios/network-infra/internal/server"
	"github.com/nilbyte-studios/network-infra/internal/storage"
	"github.com/nilbyte-studios/network-infra/internal/tailscale"
)

func main() {
	log.Println("Starting Tailscale Network Monitor Agent...")

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Using configuration from: config.yaml (or defaults if not found)")

	// Detect Tailscale IP
	tailscaleIP, err := tailscale.GetTailscaleIP()
	if err != nil {
		log.Fatalf("Failed to get Tailscale IP: %v", err)
	}
	log.Printf("Tailscale IP detected: %s", tailscaleIP)

	// Initialize storage
	store, err := storage.NewStorage(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()
	log.Printf("Storage initialized: %s", cfg.Storage.Path)

	// Initialize metrics collector
	collector, err := metrics.NewCollector()
	if err != nil {
		log.Fatalf("Failed to initialize metrics collector: %v", err)
	}

	// Start periodic metrics collection in background
	go periodicCollection(collector, store, cfg.Metrics.CollectionInterval)

	// Start periodic cleanup of old data
	go periodicCleanup(store, cfg.Storage.RetentionDays)

	// Configure HTTP server
	handler := server.NewHandler(collector, store, cfg)
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/status", handler.HandleStatus)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/api/peers", handler.HandlePeers)
	mux.HandleFunc("/metrics/history", handler.HandleHistory)

	// Web dashboard
	mux.Handle("/", http.RedirectHandler("/static/", http.StatusMovedPermanently))
	mux.Handle("/static/", http.StripPrefix("/static", server.ServeStatic()))

	// Determine bind address
	var addr string
	if cfg.Server.TailscaleOnly {
		addr = fmt.Sprintf("%s:%d", tailscaleIP, cfg.Server.Port)
	} else {
		addr = fmt.Sprintf(":%d", cfg.Server.Port)
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Graceful shutdown of HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Agent stopped")
}

// periodicCollection collects metrics periodically
func periodicCollection(collector *metrics.Collector, store *storage.Storage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m, err := collector.Collect()
		if err != nil {
			log.Printf("Error collecting metrics: %v", err)
			continue
		}

		if err := store.Save(m); err != nil {
			log.Printf("Error saving metrics: %v", err)
		}
	}
}

// periodicCleanup cleans up old data periodically
func periodicCleanup(store *storage.Storage, retentionDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := store.Cleanup(retentionDays); err != nil {
			log.Printf("Error cleaning up old metrics: %v", err)
		}
	}
}
