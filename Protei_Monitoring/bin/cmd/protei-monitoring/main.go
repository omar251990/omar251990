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
	"github.com/protei/monitoring/pkg/analysis"
	"github.com/protei/monitoring/pkg/analytics"
	"github.com/protei/monitoring/pkg/auth"
	"github.com/protei/monitoring/pkg/capture"
	"github.com/protei/monitoring/pkg/config"
	"github.com/protei/monitoring/pkg/correlation"
	"github.com/protei/monitoring/pkg/database"
	"github.com/protei/monitoring/pkg/decoder"
	"github.com/protei/monitoring/pkg/decoder/cap"
	"github.com/protei/monitoring/pkg/decoder/diameter"
	"github.com/protei/monitoring/pkg/decoder/gtp"
	"github.com/protei/monitoring/pkg/decoder/inap"
	map_decoder "github.com/protei/monitoring/pkg/decoder/map"
	"github.com/protei/monitoring/pkg/decoder/nas"
	"github.com/protei/monitoring/pkg/decoder/ngap"
	"github.com/protei/monitoring/pkg/decoder/pfcp"
	"github.com/protei/monitoring/pkg/decoder/s1ap"
	"github.com/protei/monitoring/pkg/dictionary"
	"github.com/protei/monitoring/pkg/flows"
	"github.com/protei/monitoring/pkg/health"
	"github.com/protei/monitoring/pkg/knowledge"
	"github.com/protei/monitoring/pkg/license"
	"github.com/protei/monitoring/pkg/storage"
	"github.com/protei/monitoring/pkg/visualization"
	"github.com/protei/monitoring/pkg/web"
)

const (
	appName    = "Protei_Monitoring"
	appVersion = "2.0.0"
)

var (
	configPath  = flag.String("config", "configs/config.yaml", "Path to configuration file")
	licensePath = flag.String("license", "configs/license.json", "Path to license file")
	version     = flag.Bool("version", false, "Print version and exit")
)

// Application holds all application components
type Application struct {
	config            *config.Config
	logger            *logger.Logger
	license           *license.Manager
	auth              *auth.Service
	db                *database.DB
	decoders          *decoder.DecoderRegistry
	dictionaries      *dictionary.Loader
	capture           *capture.Engine
	correlation       *correlation.Engine
	analytics         *analytics.KPIEngine
	storage           *storage.Storage
	visualization     *visualization.LadderDiagram
	health            *health.HealthCheck
	knowledgeBase     *knowledge.KnowledgeBase
	analysisEngine    *analysis.Analyzer
	flowReconstructor *flows.FlowReconstructor
	subscriberCorr    *correlation.SubscriberCorrelator
	webServer         *web.Server
	server            *http.Server
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	fmt.Printf("🚀 Starting %s v%s\n", appName, appVersion)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate license
	fmt.Println("🔐 Validating license...")
	licenseMgr, err := license.NewManager(*licensePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ License validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ License valid for: %s (expires: %s)\n",
		licenseMgr.GetLicense().CustomerName,
		licenseMgr.GetLicense().ExpiryDate)

	// Initialize application
	app, err := NewApplication(cfg, licenseMgr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Start application
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start application: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s started successfully\n", appName)
	fmt.Printf("📊 Dashboard: http://%s\n", cfg.GetAddr())
	fmt.Printf("🔍 Health: http://%s/health\n", cfg.GetAddr())

	// Wait for shutdown signal
	app.WaitForShutdown()

	// Graceful shutdown
	if err := app.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("👋 Application stopped gracefully")
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config, licenseMgr *license.Manager) (*Application, error) {
	app := &Application{
		config:  cfg,
		license: licenseMgr,
	}

	// Initialize logger
	fmt.Println("📝 Initializing logger...")
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
	app.logger.Info("Protei_Monitoring initializing", "version", appVersion)

	// Initialize authentication
	fmt.Println("🔐 Initializing authentication...")
	authCfg := &auth.Config{
		JWTSecret:      "CHANGE_THIS_SECRET_IN_PRODUCTION",
		TokenExpiry:    24 * time.Hour,
		PasswordMinLen: 8,
		AllowLocalAuth: true,
	}
	app.auth = auth.NewService(authCfg)

	// Initialize database (optional, can be disabled)
	if os.Getenv("DB_ENABLED") == "true" {
		fmt.Println("🗄️  Initializing database...")
		dbCfg := &database.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     5432,
			Database: os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  "disable",
			MaxConns: 50,
			MaxIdle:  10,
		}

		if dbCfg.Host != "" {
			db, err := database.New(dbCfg)
			if err != nil {
				app.logger.Warn("Database initialization failed, continuing without DB", "error", err)
			} else {
				app.db = db
				app.logger.Info("Database initialized successfully")
			}
		}
	}

	// Initialize decoder registry
	fmt.Println("🔧 Initializing protocol decoders...")
	app.decoders = decoder.NewRegistry()

	// Register protocol decoders based on license
	if licenseMgr.IsFeatureEnabled("map") && cfg.Protocols.MAP.Enabled {
		mapDecoder := map_decoder.NewMAPDecoder(cfg.Protocols.MAP.Version)
		app.decoders.Register(mapDecoder)
		app.logger.Info("Registered MAP decoder")
	}

	if licenseMgr.IsFeatureEnabled("cap") && cfg.Protocols.CAP.Enabled {
		capDecoder := cap.NewCAPDecoder(cfg.Protocols.CAP.Version)
		app.decoders.Register(capDecoder)
		app.logger.Info("Registered CAP decoder")
	}

	if licenseMgr.IsFeatureEnabled("inap") && cfg.Protocols.INAP.Enabled {
		inapDecoder := inap.NewINAPDecoder(cfg.Protocols.INAP.Version)
		app.decoders.Register(inapDecoder)
		app.logger.Info("Registered INAP decoder")
	}

	if licenseMgr.IsFeatureEnabled("diameter") && cfg.Protocols.Diameter.Enabled {
		diameterDecoder := diameter.NewDiameterDecoder(
			cfg.Protocols.Diameter.Applications,
			cfg.Protocols.Diameter.VendorSupport,
		)
		app.decoders.Register(diameterDecoder)
		app.logger.Info("Registered Diameter decoder")
	}

	if licenseMgr.IsFeatureEnabled("gtp") && cfg.Protocols.GTP.Enabled {
		gtpDecoder := gtp.NewGTPDecoder(cfg.Protocols.GTP.Versions)
		app.decoders.Register(gtpDecoder)
		app.logger.Info("Registered GTP decoder")
	}

	if cfg.Protocols.PFCP.Enabled {
		pfcpDecoder := pfcp.NewPFCPDecoder()
		app.decoders.Register(pfcpDecoder)
		app.logger.Info("Registered PFCP decoder")
	}

	if cfg.Protocols.NGAP.Enabled {
		ngapDecoder := ngap.NewNGAPDecoder()
		app.decoders.Register(ngapDecoder)
		app.logger.Info("Registered NGAP decoder")
	}

	if cfg.Protocols.S1AP.Enabled {
		s1apDecoder := s1ap.NewS1APDecoder()
		app.decoders.Register(s1apDecoder)
		app.logger.Info("Registered S1AP decoder")
	}

	if cfg.Protocols.NAS.Enabled {
		nasDecoder := nas.NewNASDecoder(cfg.Protocols.NAS.Generations)
		app.decoders.Register(nasDecoder)
		app.logger.Info("Registered NAS decoder")
	}

	// Initialize vendor dictionaries
	fmt.Println("📚 Loading vendor dictionaries...")
	dictCfg := &dictionary.Config{
		BasePath:    "/usr/protei/Protei_Monitoring/dictionaries",
		VendorPaths: cfg.Vendors.Dictionaries,
		AutoDetect:  cfg.Vendors.AutoDetect,
	}
	app.dictionaries = dictionary.NewLoader(dictCfg)
	if err := app.dictionaries.LoadAll(); err != nil {
		app.logger.Warn("Failed to load some dictionaries", "error", err)
	}

	// Initialize correlation engine
	fmt.Println("🔗 Initializing correlation engine...")
	corrCfg := &correlation.Config{
		CacheSize:         cfg.Correlation.TIDCacheSize,
		TIDTTL:            time.Duration(cfg.Correlation.TIDTTL) * time.Second,
		SessionTimeout:    time.Duration(cfg.Correlation.SessionTimeout) * time.Second,
		CorrelationFields: cfg.Correlation.CorrelationFields,
		E2ETracking:       cfg.Correlation.E2ETracking,
	}
	app.correlation = correlation.NewEngine(corrCfg)

	// Initialize analytics engine
	if cfg.Analytics.KPIs.Enabled {
		fmt.Println("📊 Initializing analytics engine...")
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
	}

	// Initialize storage
	fmt.Println("💾 Initializing storage layer...")
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

	// Initialize visualization
	if cfg.Visualization.LadderDiagrams.Enabled {
		vizCfg := &visualization.Config{
			Format:         cfg.Visualization.LadderDiagrams.Format,
			MaxMessages:    cfg.Visualization.LadderDiagrams.MaxMessagesPerDiagram,
			OutputPath:     cfg.Visualization.LadderDiagrams.OutputPath,
			AutoLabelNodes: cfg.Visualization.LadderDiagrams.AutoLabelNodes,
		}
		app.visualization = visualization.NewLadderDiagram(vizCfg)
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
	}

	// Initialize knowledge base (3GPP standards, error codes, procedures)
	fmt.Println("📚 Initializing knowledge base...")
	app.knowledgeBase = knowledge.NewKnowledgeBase()
	if err := app.knowledgeBase.LoadStandards(); err != nil {
		app.logger.Warn("Failed to load knowledge base standards", "error", err)
	} else {
		standards := app.knowledgeBase.ListStandards()
		protocols := app.knowledgeBase.ListProtocols()
		app.logger.Info("Knowledge base initialized",
			"standards", len(standards),
			"protocols", len(protocols))
		fmt.Printf("  ✅ Loaded %d standards and %d protocols\n", len(standards), len(protocols))
	}

	// Initialize AI analysis engine
	fmt.Println("🤖 Initializing AI analysis engine...")
	app.analysisEngine = analysis.NewAnalyzer()
	app.logger.Info("AI analysis engine initialized with detection rules")
	fmt.Println("  ✅ AI analysis engine ready (7 detection rules)")

	// Initialize flow reconstructor
	fmt.Println("🔄 Initializing flow reconstructor...")
	app.flowReconstructor = flows.NewFlowReconstructor()
	templates := app.flowReconstructor.ListTemplates()
	app.logger.Info("Flow reconstructor initialized", "templates", len(templates))
	fmt.Printf("  ✅ Flow reconstructor ready (%d procedure templates)\n", len(templates))

	// Initialize subscriber correlator
	fmt.Println("👤 Initializing subscriber correlator...")
	app.subscriberCorr = correlation.NewSubscriberCorrelator()
	app.logger.Info("Subscriber correlator initialized")
	fmt.Println("  ✅ Subscriber correlator ready (multi-identifier tracking)")

	// Initialize PCAP capture engine
	fmt.Println("📡 Initializing PCAP capture engine...")
	captureCfg := &capture.Config{
		Sources:       cfg.Ingestion.Sources,
		BufferSize:    cfg.Ingestion.BufferSize,
		Workers:       cfg.Ingestion.Workers,
		BatchSize:     cfg.Ingestion.BatchSize,
		OutputChannel: make(chan *capture.CapturedPacket, cfg.Ingestion.BufferSize),
	}
	app.capture = capture.NewEngine(captureCfg)

	// Register packet processor
	app.capture.RegisterProcessor(app)

	// Initialize web server with all services
	fmt.Println("🌐 Initializing web server...")
	webCfg := web.Config{
		Port:             cfg.Server.Port,
		AuthService:      app.auth,
		KnowledgeBase:    app.knowledgeBase,
		AnalysisEngine:   app.analysisEngine,
		FlowReconstructor: app.flowReconstructor,
		SubscriberCorr:   app.subscriberCorr,
		Logger:           app.logger.With().Str("component", "web").Logger(),
	}

	app.webServer = web.New(webCfg)

	app.logger.Info("Web server initialized", "port", cfg.Server.Port)
	fmt.Printf("  ✅ Web server ready on port %d\n", cfg.Server.Port)

	// Initialize HTTP server
	app.server = &http.Server{
		Addr:           cfg.GetAddr(),
		Handler:        app.webServer,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return app, nil
}

// Process implements capture.Processor interface
func (a *Application) Process(packet *capture.CapturedPacket) error {
	// Decode packet
	msg, err := a.decoders.Decode(packet.Data, packet.Metadata)
	if err != nil {
		return err
	}

	// Store event
	if a.storage != nil {
		a.storage.WriteEvent(msg)
	}

	// Correlate message
	if a.correlation != nil {
		session, err := a.correlation.Correlate(msg)
		if err == nil && session != nil {
			// Process session for analytics
			if a.analytics != nil {
				a.analytics.ProcessSession(session)
			}

			// Generate CDR if session complete
			if session.Result != decoder.ResultUnknown {
				if a.storage != nil {
					a.storage.WriteCDR(session)
				}
			}
		}
	}

	// Process message in subscriber correlator (for timeline tracking)
	if a.subscriberCorr != nil {
		a.subscriberCorr.ProcessMessage(msg)
	}

	// Analyze message for issues (AI analysis)
	if a.analysisEngine != nil {
		a.analysisEngine.AnalyzeMessage(msg)
	}

	// Update health metrics
	if a.health != nil {
		a.health.RecordMessage()
	}

	return nil
}

// Start starts the application
func (a *Application) Start() error {
	a.logger.Info("Starting Protei_Monitoring", "address", a.server.Addr)

	// Update health status
	if a.health != nil {
		a.health.UpdateComponentStatus("main", true, "Application started")
	}

	// Start PCAP capture
	if err := a.capture.Start(); err != nil {
		return fmt.Errorf("failed to start capture engine: %w", err)
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

	// Stop capture engine
	if a.capture != nil {
		a.capture.Stop()
	}

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

	// Close database
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Database close error", err)
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
	mux.HandleFunc("/api/license", a.handleLicense)

	// Authentication endpoints
	mux.HandleFunc("/api/auth/login", a.handleLogin)
	mux.HandleFunc("/api/auth/logout", a.handleLogout)

	// Dashboard
	mux.HandleFunc("/", a.handleDashboard)

	return mux
}

// HTTP Handlers (keeping existing ones and adding new)

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

		if i >= 99 {
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

func (a *Application) handleLicense(w http.ResponseWriter, r *http.Request) {
	lic := a.license.GetLicense()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"customer": "%s", "expiry": "%s", "max_subscribers": %d, "max_tps": %d, "features": {"2g": %v, "3g": %v, "4g": %v, "5g": %v}}`,
		lic.CustomerName, lic.ExpiryDate, lic.MaxSubscribers, lic.MaxTPS,
		lic.Enable2G, lic.Enable3G, lic.Enable4G, lic.Enable5G)
}

func (a *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Implement login logic
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Login endpoint - implement with auth service"}`)
}

func (a *Application) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Implement logout logic
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "Logout successful"}`)
}

func (a *Application) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
	<title>Protei Monitoring Dashboard v2.0</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 0; background: #f5f5f5; }
		.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; }
		.header h1 { margin: 0; font-size: 36px; }
		.header p { margin: 10px 0 0 0; opacity: 0.9; }
		.container { max-width: 1400px; margin: 20px auto; padding: 0 20px; }
		.card { background: white; padding: 25px; margin: 15px 0; border-radius: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
		.metric { display: inline-block; margin: 15px 30px; }
		.metric-value { font-size: 42px; font-weight: bold; color: #667eea; }
		.metric-label { font-size: 14px; color: #7f8c8d; margin-top: 5px; }
		.feature-badge { display: inline-block; background: #27ae60; color: white; padding: 5px 15px; border-radius: 20px; margin: 5px; font-size: 12px; }
		.feature-badge.disabled { background: #95a5a6; }
		.api-link { color: #667eea; text-decoration: none; padding: 5px 10px; display: inline-block; }
		.api-link:hover { background: #f0f0f0; border-radius: 5px; }
	</style>
</head>
<body>
	<div class="header">
		<h1>🌐 Protei_Monitoring v2.0</h1>
		<p>Advanced Multi-Protocol Telecom Monitoring & Analysis Platform</p>
		<p>Full 2G/3G/4G/5G Support | Real-time Analytics | Vendor-Agnostic</p>
	</div>
	<div class="container">
		<div class="card">
			<h2>📊 System Status</h2>
			<div id="status">Loading...</div>
		</div>
		<div class="card">
			<h2>📈 Key Performance Indicators</h2>
			<div id="kpi">Loading...</div>
		</div>
		<div class="card">
			<h2>🔐 License Information</h2>
			<div id="license">Loading...</div>
		</div>
		<div class="card">
			<h2>🔌 API Endpoints</h2>
			<ul style="list-style: none; padding: 0;">
				<li><a class="api-link" href="/health">GET /health</a> - Health check</li>
				<li><a class="api-link" href="/metrics">GET /metrics</a> - Prometheus metrics</li>
				<li><a class="api-link" href="/api/sessions">GET /api/sessions</a> - Active sessions</li>
				<li><a class="api-link" href="/api/kpi">GET /api/kpi</a> - KPI report</li>
				<li><a class="api-link" href="/api/roaming">GET /api/roaming</a> - Roaming heatmap</li>
				<li><a class="api-link" href="/api/license">GET /api/license</a> - License details</li>
			</ul>
		</div>
		<div class="card">
			<h2>✨ New Features in v2.0</h2>
			<ul>
				<li><strong>Enhanced Protocol Support:</strong> CAP, INAP, PFCP, NGAP, S1AP, NAS (4G/5G)</li>
				<li><strong>PCAP Capture Engine:</strong> Live capture and file monitoring</li>
				<li><strong>Database Integration:</strong> PostgreSQL with Liquibase migrations</li>
				<li><strong>Authentication:</strong> JWT-based auth with LDAP support</li>
				<li><strong>License Management:</strong> MAC-based validation with feature control</li>
				<li><strong>Vendor Dictionaries:</strong> Ericsson, Huawei, ZTE, Nokia support</li>
				<li><strong>Advanced Analytics:</strong> Real-time KPIs and roaming intelligence</li>
			</ul>
		</div>
	</div>
	<script>
		function updateStatus() {
			fetch('/health')
				.then(r => r.json())
				.then(data => {
					document.getElementById('status').innerHTML =
						'<div class="metric"><div class="metric-value">' + (data.healthy ? '✅' : '❌') + '</div><div class="metric-label">Health</div></div>' +
						'<div class="metric"><div class="metric-value">' + data.uptime + 's</div><div class="metric-label">Uptime</div></div>' +
						'<div class="metric"><div class="metric-value">' + data.messages.toLocaleString() + '</div><div class="metric-label">Messages</div></div>';
				})
				.catch(err => document.getElementById('status').innerHTML = '<p style="color: red;">Error loading status</p>');

			fetch('/api/kpi')
				.then(r => r.json())
				.then(data => {
					let html = '';
					for (let proc in data.procedures) {
						let m = data.procedures[proc];
						html += '<div class="metric"><div class="metric-value">' + m.success_rate.toFixed(1) + '%</div><div class="metric-label">' + proc + '</div></div>';
					}
					document.getElementById('kpi').innerHTML = html || '<p>No KPI data yet. Start processing PCAP files.</p>';
				})
				.catch(err => document.getElementById('kpi').innerHTML = '<p>No data yet</p>');

			fetch('/api/license')
				.then(r => r.json())
				.then(data => {
					document.getElementById('license').innerHTML =
						'<p><strong>Customer:</strong> ' + data.customer + '</p>' +
						'<p><strong>Expiry:</strong> ' + data.expiry + '</p>' +
						'<p><strong>Max Subscribers:</strong> ' + data.max_subscribers.toLocaleString() + '</p>' +
						'<p><strong>Max TPS:</strong> ' + data.max_tps.toLocaleString() + '</p>' +
						'<p><strong>Enabled Features:</strong></p>' +
						'<span class="feature-badge' + (data.features['2g'] ? '' : ' disabled') + '">2G</span>' +
						'<span class="feature-badge' + (data.features['3g'] ? '' : ' disabled') + '">3G</span>' +
						'<span class="feature-badge' + (data.features['4g'] ? '' : ' disabled') + '">4G</span>' +
						'<span class="feature-badge' + (data.features['5g'] ? '' : ' disabled') + '">5G</span>';
				})
				.catch(err => document.getElementById('license').innerHTML = '<p>License info not available</p>');
		}

		updateStatus();
		setInterval(updateStatus, 5000);
	</script>
</body>
</html>`)
}
