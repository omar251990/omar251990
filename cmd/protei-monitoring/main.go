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

	"github.com/protei/monitoring/internal/logger"
	"github.com/protei/monitoring/pkg/analytics"
	"github.com/protei/monitoring/pkg/config"
	"github.com/protei/monitoring/pkg/correlation"
	"github.com/protei/monitoring/pkg/decoder"
	"github.com/protei/monitoring/pkg/decoder/diameter"
	"github.com/protei/monitoring/pkg/decoder/gtp"
	map_decoder "github.com/protei/monitoring/pkg/decoder/map"
	"github.com/protei/monitoring/pkg/health"
	"github.com/protei/monitoring/pkg/storage"
	"github.com/protei/monitoring/pkg/visualization"
)

const (
	appName    = "Protei_Monitoring"
	appVersion = "1.0.0"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
	version    = flag.Bool("version", false, "Print version and exit")
)

// Application holds all application components
type Application struct {
	config      *config.Config
	logger      *logger.Logger
	decoders    *decoder.DecoderRegistry
	correlation *correlation.Engine
	analytics   *analytics.KPIEngine
	storage     *storage.Storage
	visualization *visualization.LadderDiagram
	health      *health.HealthCheck
	server      *http.Server
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize application
	app, err := NewApplication(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Start application
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	app.WaitForShutdown()

	// Graceful shutdown
	if err := app.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application stopped gracefully")
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	// Initialize logger
	logCfg := logger.Config{
		Path:       cfg.Storage.Logs.Path + "/app.log",
		Level:      cfg.Storage.Logs.Level,
		Format:     cfg.Storage.Logs.Format,
		MaxSizeMB:  cfg.Storage.Logs.MaxSizeMB,
		MaxBackups: cfg.Storage.Logs.MaxBackups,
		MaxAgeDays: cfg.Storage.Logs.MaxAgeDays,
		Compress:   cfg.Storage.Logs.Compress,
	}

	log, err := logger.New(logCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	app.logger = log

	app.logger.Info("Initializing Protei_Monitoring",
		"version", appVersion,
		"config", *configPath,
	)

	// Initialize decoder registry
	app.decoders = decoder.NewRegistry()

	// Register protocol decoders
	if cfg.Protocols.MAP.Enabled {
		mapDecoder := map_decoder.NewMAPDecoder(cfg.Protocols.MAP.Version)
		app.decoders.Register(mapDecoder)
		app.logger.Info("Registered MAP decoder")
	}

	if cfg.Protocols.Diameter.Enabled {
		diameterDecoder := diameter.NewDiameterDecoder(
			cfg.Protocols.Diameter.Applications,
			cfg.Protocols.Diameter.VendorSupport,
		)
		app.decoders.Register(diameterDecoder)
		app.logger.Info("Registered Diameter decoder")
	}

	if cfg.Protocols.GTP.Enabled {
		gtpDecoder := gtp.NewGTPDecoder(cfg.Protocols.GTP.Versions)
		app.decoders.Register(gtpDecoder)
		app.logger.Info("Registered GTP decoder")
	}

	// Initialize correlation engine
	corrCfg := &correlation.Config{
		CacheSize:         cfg.Correlation.TIDCacheSize,
		TIDTTL:            time.Duration(cfg.Correlation.TIDTTL) * time.Second,
		SessionTimeout:    time.Duration(cfg.Correlation.SessionTimeout) * time.Second,
		CorrelationFields: cfg.Correlation.CorrelationFields,
		E2ETracking:       cfg.Correlation.E2ETracking,
	}
	app.correlation = correlation.NewEngine(corrCfg)
	app.logger.Info("Initialized correlation engine")

	// Initialize analytics engine
	if cfg.Analytics.KPIs.Enabled {
		analyticsCfg := &analytics.Config{
			Enabled:             true,
			CalculationInterval: time.Duration(cfg.Analytics.KPIs.CalculationInterval) * time.Second,
			Procedures:          cfg.Analytics.KPIs.Procedures,
			Metrics:             cfg.Analytics.KPIs.Metrics,
			RoamingEnabled:      cfg.Analytics.Roaming.Enabled,
			FailureThreshold:    cfg.Analytics.FailureAnalysis.ThresholdFailureRate,
			LatencyThreshold:    cfg.Analytics.FailureAnalysis.ThresholdLatencyMs,
		}
		app.analytics = analytics.NewKPIEngine(analyticsCfg)
		app.logger.Info("Initialized analytics engine")
	}

	// Initialize storage
	storageCfg := &storage.Config{
		EventsEnabled: cfg.Storage.Events.Enabled,
		EventsPath:    cfg.Storage.Events.Path,
		EventsFormat:  cfg.Storage.Events.Format,
		CDREnabled:    cfg.Storage.CDR.Enabled,
		CDRPath:       cfg.Storage.CDR.Path,
		CDRFormat:     cfg.Storage.CDR.Format,
		CDRFields:     cfg.Storage.CDR.Fields,
		RetentionDays: cfg.Storage.CDR.RetentionDays,
	}
	app.storage, err = storage.NewStorage(storageCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	app.logger.Info("Initialized storage layer")

	// Initialize visualization
	if cfg.Visualization.LadderDiagrams.Enabled {
		vizCfg := &visualization.Config{
			Format:         cfg.Visualization.LadderDiagrams.Format,
			MaxMessages:    cfg.Visualization.LadderDiagrams.MaxMessagesPerDiagram,
			OutputPath:     cfg.Visualization.LadderDiagrams.OutputPath,
			AutoLabelNodes: cfg.Visualization.LadderDiagrams.AutoLabelNodes,
		}
		app.visualization = visualization.NewLadderDiagram(vizCfg)
		app.logger.Info("Initialized visualization engine")
	}

	// Initialize health check
	if cfg.Health.Enabled {
		healthCfg := &health.Config{
			Enabled:          true,
			CheckInterval:    time.Duration(cfg.Health.CheckInterval) * time.Second,
			WatchdogEnabled:  cfg.Health.Watchdog.Enabled,
			WatchdogTimeout:  time.Duration(cfg.Health.Watchdog.Timeout) * time.Second,
			RestartOnFailure: cfg.Health.Watchdog.RestartOnFailure,
		}
		app.health = health.NewHealthCheck(healthCfg)
		app.logger.Info("Initialized health check system")
	}

	// Initialize HTTP server
	app.server = &http.Server{
		Addr:           cfg.GetAddr(),
		Handler:        app.setupRoutes(),
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return app, nil
}

// Start starts the application
func (a *Application) Start() error {
	a.logger.Info("Starting Protei_Monitoring", "address", a.server.Addr)

	// Update health status
	if a.health != nil {
		a.health.UpdateComponentStatus("main", true, "Application started")
	}

	// Start HTTP server in goroutine
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server error", err)
		}
	}()

	a.logger.Info("Application started successfully")
	return nil
}

// Stop gracefully stops the application
func (a *Application) Stop() error {
	a.logger.Info("Stopping application...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", err)
	}

	// Close storage
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			a.logger.Error("Storage close error", err)
		}
	}

	a.logger.Info("Application stopped")
	return nil
}

// WaitForShutdown waits for termination signal
func (a *Application) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	a.logger.Info("Received shutdown signal", "signal", sig.String())
}

// setupRoutes configures HTTP routes
func (a *Application) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/ready", a.handleReady)
	mux.HandleFunc("/metrics", a.handleMetrics)

	// API endpoints
	mux.HandleFunc("/api/sessions", a.handleSessions)
	mux.HandleFunc("/api/kpi", a.handleKPI)
	mux.HandleFunc("/api/roaming", a.handleRoaming)

	// Dashboard
	mux.HandleFunc("/", a.handleDashboard)

	return mux
}

// HTTP Handlers

func (a *Application) handleHealth(w http.ResponseWriter, r *http.Request) {
	if a.health == nil {
		http.Error(w, "Health check not enabled", http.StatusServiceUnavailable)
		return
	}

	status := a.health.GetStatus()
	w.Header().Set("Content-Type", "application/json")

	if status.Healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{"healthy": %v, "uptime": %d, "messages": %d}`,
		status.Healthy, status.UptimeSeconds, status.MessagesProcessed)
}

func (a *Application) handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"ready": true}`)
}

func (a *Application) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if a.health == nil {
		http.Error(w, "Metrics not available", http.StatusServiceUnavailable)
		return
	}

	status := a.health.GetStatus()
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "# HELP protei_uptime_seconds Application uptime\n")
	fmt.Fprintf(w, "protei_uptime_seconds %d\n", status.UptimeSeconds)
	fmt.Fprintf(w, "# HELP protei_messages_processed_total Total messages processed\n")
	fmt.Fprintf(w, "protei_messages_processed_total %d\n", status.MessagesProcessed)
	fmt.Fprintf(w, "# HELP protei_sessions_active Active sessions\n")
	fmt.Fprintf(w, "protei_sessions_active %d\n", status.SessionsActive)
}

func (a *Application) handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := a.correlation.GetAllSessions()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"count": %d, "sessions": [`, len(sessions))

	for i, session := range sessions {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `{"tid": "%s", "imsi": "%s", "procedure": "%s", "result": "%s"}`,
			session.TID, session.IMSI, session.Procedure, session.Result)

		if i >= 99 { // Limit to 100 sessions
			break
		}
	}

	fmt.Fprint(w, "]}")
}

func (a *Application) handleKPI(w http.ResponseWriter, r *http.Request) {
	if a.analytics == nil {
		http.Error(w, "Analytics not enabled", http.StatusNotFound)
		return
	}

	report := a.analytics.Calculate()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"timestamp": "%s", "procedures": {`, report.Timestamp.Format(time.RFC3339))

	first := true
	for proc, metrics := range report.Procedures {
		if !first {
			fmt.Fprint(w, ",")
		}
		first = false

		fmt.Fprintf(w, `"%s": {"total": %d, "success": %d, "failure": %d, "success_rate": %.2f}`,
			proc, metrics.TotalCount, metrics.SuccessCount, metrics.FailureCount, metrics.SuccessRate)
	}

	fmt.Fprint(w, "}}")
}

func (a *Application) handleRoaming(w http.ResponseWriter, r *http.Request) {
	if a.analytics == nil {
		http.Error(w, "Analytics not enabled", http.StatusNotFound)
		return
	}

	heatmap := a.analytics.GetCellHeatmap()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"cells": [`)

	first := true
	for cellID, metrics := range heatmap {
		if !first {
			fmt.Fprint(w, ",")
		}
		first = false

		fmt.Fprintf(w, `{"cell_id": "%s", "plmn": "%s", "roamers": %d}`,
			cellID, metrics.PLMN, metrics.RoamerCount)
	}

	fmt.Fprint(w, "]}")
}

func (a *Application) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
	<title>Protei Monitoring Dashboard</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
		.header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
		.container { margin-top: 20px; }
		.card { background: white; padding: 20px; margin: 10px 0; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.metric { display: inline-block; margin: 10px 20px; }
		.metric-value { font-size: 32px; font-weight: bold; color: #3498db; }
		.metric-label { font-size: 14px; color: #7f8c8d; }
	</style>
</head>
<body>
	<div class="header">
		<h1>🌐 Protei_Monitoring Dashboard</h1>
		<p>Multi-Protocol Telecom Monitoring & Analysis Platform</p>
	</div>
	<div class="container">
		<div class="card">
			<h2>System Status</h2>
			<div id="status">Loading...</div>
		</div>
		<div class="card">
			<h2>Key Performance Indicators</h2>
			<div id="kpi">Loading...</div>
		</div>
		<div class="card">
			<h2>API Endpoints</h2>
			<ul>
				<li><a href="/health">/health</a> - Health check</li>
				<li><a href="/metrics">/metrics</a> - Prometheus metrics</li>
				<li><a href="/api/sessions">/api/sessions</a> - Active sessions</li>
				<li><a href="/api/kpi">/api/kpi</a> - KPI report</li>
				<li><a href="/api/roaming">/api/roaming</a> - Roaming heatmap</li>
			</ul>
		</div>
	</div>
	<script>
		function updateStatus() {
			fetch('/health')
				.then(r => r.json())
				.then(data => {
					document.getElementById('status').innerHTML =
						'<div class="metric"><div class="metric-value">' + (data.healthy ? '✓' : '✗') + '</div><div class="metric-label">Health</div></div>' +
						'<div class="metric"><div class="metric-value">' + data.uptime + 's</div><div class="metric-label">Uptime</div></div>' +
						'<div class="metric"><div class="metric-value">' + data.messages + '</div><div class="metric-label">Messages</div></div>';
				});

			fetch('/api/kpi')
				.then(r => r.json())
				.then(data => {
					let html = '';
					for (let proc in data.procedures) {
						let m = data.procedures[proc];
						html += '<div class="metric"><div class="metric-value">' + m.success_rate.toFixed(1) + '%</div><div class="metric-label">' + proc + '</div></div>';
					}
					document.getElementById('kpi').innerHTML = html || 'No data yet';
				});
		}

		updateStatus();
		setInterval(updateStatus, 5000);
	</script>
</body>
</html>`)
}
