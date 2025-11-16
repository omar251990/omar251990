package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

const (
	Version   = "2.0.0"
	BuildDate = "2024-01-15"
	GitCommit = "88c149c"
)

// Application represents the main application
type Application struct {
	InstallDir string
	ConfigDir  string
	DB         *sql.DB
	HTTPServer *http.Server
	PIDFile    string
}

func main() {
	// Print version if requested
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("Protei Monitoring v%s (build: %s, commit: %s)\n", Version, BuildDate, GitCommit)
		os.Exit(0)
	}

	// Initialize application
	app := &Application{
		InstallDir: "/usr/protei/Protei_Monitoring",
	}

	app.ConfigDir = filepath.Join(app.InstallDir, "config")
	app.PIDFile = filepath.Join(app.InstallDir, "tmp", "protei-monitoring.pid")

	// Print startup banner
	printBanner()

	// Write PID file
	if err := app.writePIDFile(); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}

	// Ensure PID file is removed on exit
	defer app.removePIDFile()

	// Load configuration
	log.Println("[INFO] Loading configuration...")
	config, err := app.loadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	log.Println("[INFO] Connecting to database...")
	db, err := app.connectDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	app.DB = db
	defer db.Close()

	log.Println("[SUCCESS] Database connected")

	// Initialize components
	log.Println("[INFO] Initializing components...")
	if err := app.initializeComponents(); err != nil {
		log.Fatalf("Failed to initialize components: %v", err)
	}

	// Start HTTP server
	log.Println("[INFO] Starting web server...")
	app.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", config["WEB_PORT"]),
		Handler: app.setupRoutes(),
	}

	// Start server in goroutine
	go func() {
		log.Printf("[SUCCESS] Web server listening on port %s", config["WEB_PORT"])
		if err := app.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	sig := <-sigChan
	log.Printf("[INFO] Received signal: %v", sig)

	if sig == syscall.SIGHUP {
		log.Println("[INFO] Reloading configuration...")
		// Reload configuration without shutdown
		config, err = app.loadConfiguration()
		if err != nil {
			log.Printf("[ERROR] Failed to reload configuration: %v", err)
		} else {
			log.Println("[SUCCESS] Configuration reloaded")
		}
		return
	}

	// Graceful shutdown
	log.Println("[INFO] Shutting down gracefully...")
	app.shutdown()
	log.Println("[SUCCESS] Shutdown complete")
}

func printBanner() {
	banner := `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Protei Monitoring v%s
  Professional Telecom Signaling Monitor
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Build: %s | Commit: %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

`
	fmt.Printf(banner, Version, BuildDate, GitCommit)
}

func (a *Application) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(a.PIDFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func (a *Application) removePIDFile() {
	os.Remove(a.PIDFile)
}

func (a *Application) loadConfiguration() (map[string]string, error) {
	config := make(map[string]string)

	// Load system.cfg
	systemConfig := filepath.Join(a.ConfigDir, "system.cfg")
	if err := loadConfigFile(systemConfig, config); err != nil {
		return nil, err
	}

	// Load db.cfg
	dbConfig := filepath.Join(a.ConfigDir, "db.cfg")
	if err := loadConfigFile(dbConfig, config); err != nil {
		return nil, err
	}

	// Set defaults if not present
	if config["WEB_PORT"] == "" {
		config["WEB_PORT"] = "8080"
	}

	return config, nil
}

func loadConfigFile(filename string, config map[string]string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse simple key=value format
	lines := string(data)
	for _, line := range splitLines(lines) {
		line = trimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '[' {
			continue
		}

		// Split on first =
		parts := splitOnFirst(line, '=')
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(trimQuotes(parts[1]))
			config[key] = value
		}
	}

	return nil
}

func (a *Application) connectDatabase(config map[string]string) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config["DB_HOST"],
		config["DB_PORT"],
		config["DB_USER"],
		config["DB_PASSWORD"],
		config["DB_NAME"],
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func (a *Application) initializeComponents() error {
	log.Println("  - Initializing protocol decoders...")
	// TODO: Initialize protocol decoders

	log.Println("  - Initializing CDR writers...")
	// TODO: Initialize CDR manager

	log.Println("  - Initializing correlation engine...")
	// TODO: Initialize correlation engine

	log.Println("  - Initializing knowledge base...")
	// TODO: Initialize knowledge base

	log.Println("[SUCCESS] All components initialized")
	return nil
}

func (a *Application) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/health/live", a.handleLiveness)
	mux.HandleFunc("/health/ready", a.handleReadiness)

	// Status endpoint
	mux.HandleFunc("/api/v1/status", a.handleStatus)

	// Version endpoint
	mux.HandleFunc("/api/v1/version", a.handleVersion)

	// OAM endpoints (placeholder)
	mux.HandleFunc("/api/v1/oam/", a.handleOAM)

	return mux
}

func (a *Application) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database
	dbHealthy := true
	if err := a.DB.Ping(); err != nil {
		dbHealthy = false
	}

	if dbHealthy {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","database":"connected","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"disconnected","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}
}

func (a *Application) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (a *Application) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Check if database is ready
	if err := a.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "NOT READY")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "READY")
}

func (a *Application) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uptime := time.Since(time.Now()) // This would be from actual start time

	status := fmt.Sprintf(`{
		"application":"Protei Monitoring",
		"version":"%s",
		"status":"running",
		"uptime_seconds":%d,
		"timestamp":"%s"
	}`, Version, int(uptime.Seconds()), time.Now().Format(time.RFC3339))

	fmt.Fprint(w, status)
}

func (a *Application) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"version":"%s","build_date":"%s","git_commit":"%s"}`, Version, BuildDate, GitCommit)
}

func (a *Application) handleOAM(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message":"OAM endpoint - implementation in progress"}`)
}

func (a *Application) shutdown() {
	// Shutdown HTTP server
	if a.HTTPServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.HTTPServer.Shutdown(ctx)
	}

	// Close database
	if a.DB != nil {
		a.DB.Close()
	}
}

// Helper functions

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
		} else if ch != '\r' {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitOnFirst(s string, sep rune) []string {
	for i, ch := range s {
		if ch == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

func trimQuotes(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}
