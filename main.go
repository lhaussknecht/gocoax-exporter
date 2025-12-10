package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/louispool/gocoax-exporter/collector"
	"github.com/louispool/gocoax-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "0.1.0"

var (
	configFile = flag.String("config", "config.yaml", "Path to configuration file")
	showVer    = flag.Bool("version", false, "Show version and exit")
)

func main() {
	flag.Parse()

	if *showVer {
		fmt.Printf("goCoax Prometheus Exporter v%s\n", version)
		os.Exit(0)
	}

	log.Printf("goCoax Prometheus Exporter v%s", version)
	log.Printf("Loading configuration from: %s", *configFile)

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: %d device(s), listen on %s", len(cfg.Devices), cfg.ListenAddress)

	// Create Prometheus registry
	registry := prometheus.NewRegistry()

	// Create multi-device collector registry
	multiCollector, err := collector.NewMultiDeviceRegistry(cfg)
	if err != nil {
		log.Fatalf("Failed to create collectors: %v", err)
	}
	defer multiCollector.Close()

	// Register collectors with Prometheus
	if err := multiCollector.Register(registry); err != nil {
		log.Fatalf("Failed to register collectors: %v", err)
	}

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.Default(),
		ErrorHandling: promhttp.ContinueOnError,
	}))

	// Health endpoint
	mux.HandleFunc("/health", healthHandler)

	// Index page
	mux.HandleFunc("/", indexHandler(cfg, multiCollector))

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.ListenAddress)
		log.Printf("Metrics available at http://%s/metrics", cfg.ListenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, stopping...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Exporter stopped")
}

// healthHandler handles the /health endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

// indexHandler creates a handler for the index page
func indexHandler(cfg *config.Config, registry *collector.MultiDeviceRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>goCoax Prometheus Exporter</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
        }
        h1 {
            color: #333;
        }
        .info {
            background-color: #f0f0f0;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .devices {
            margin: 20px 0;
        }
        .device {
            background-color: #e8f4f8;
            padding: 10px;
            margin: 5px 0;
            border-left: 4px solid #0066cc;
        }
        a {
            color: #0066cc;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>goCoax Prometheus Exporter</h1>
    <p>Version: %s</p>

    <div class="info">
        <h2>Status</h2>
        <p><strong>Monitoring:</strong> %d device(s)</p>
        <p><strong>Active collectors:</strong> %d</p>
    </div>

    <div class="devices">
        <h2>Configured Devices</h2>
`, version, len(cfg.Devices), registry.GetCollectorCount())

		for i, device := range cfg.Devices {
			fmt.Fprintf(w, `        <div class="device">
            <strong>%d.</strong> %s (%s)
        </div>
`, i+1, device.Name, device.Address)
		}

		fmt.Fprintf(w, `    </div>

    <div class="info">
        <h2>Endpoints</h2>
        <ul>
            <li><a href="/metrics">/metrics</a> - Prometheus metrics</li>
            <li><a href="/health">/health</a> - Health check</li>
        </ul>
    </div>

    <div class="info">
        <h2>Prometheus Configuration</h2>
        <p>Add this to your <code>prometheus.yml</code>:</p>
        <pre>
scrape_configs:
  - job_name: 'gocoax'
    static_configs:
      - targets: ['%s']
        </pre>
    </div>
</body>
</html>
`, cfg.ListenAddress)
	}
}
