package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

//go:embed templates/* static/*
var embeddedFS embed.FS

// Server represents the web server
type Server struct {
	port          int
	server        *http.Server
	logger        zerolog.Logger
	authService   AuthService
	configManager ConfigManager
	systemMonitor SystemMonitor
	dataProvider  DataProvider
	wsClients     map[*websocket.Conn]bool
	wsClientsMux  sync.RWMutex
	upgrader      websocket.Upgrader
}

// AuthService interface for authentication
type AuthService interface {
	ValidateToken(token string) (string, string, error) // Returns username, role, error
	Login(username, password string) (string, error)    // Returns token, error
	Logout(token string) error
}

// ConfigManager interface for configuration management
type ConfigManager interface {
	GetConfig() (map[string]interface{}, error)
	UpdateConfig(updates map[string]interface{}) error
	RestartService() error
	GetProtocolConfig(protocol string) (map[string]interface{}, error)
	UpdateProtocolConfig(protocol string, config map[string]interface{}) error
	GetNetworkConfig(network string) (map[string]interface{}, error)
	UpdateNetworkConfig(network string, config map[string]interface{}) error
}

// SystemMonitor interface for resource monitoring
type SystemMonitor interface {
	GetCPUUsage() float64
	GetMemoryUsage() float64
	GetDiskUsage() (float64, error)
	GetNetworkStats() (map[string]interface{}, error)
	GetProcessStats() (map[string]interface{}, error)
}

// DataProvider interface for monitoring data
type DataProvider interface {
	GetKPIs() (map[string]interface{}, error)
	GetSessions(limit int, offset int) ([]map[string]interface{}, error)
	GetSession(tid string) (map[string]interface{}, error)
	GetAlarms(status string) ([]map[string]interface{}, error)
	GetAlarm(id string) (map[string]interface{}, error)
	AcknowledgeAlarm(id string, username string) error
	GetLicenseInfo() (map[string]interface{}, error)
	GetTopology() (map[string]interface{}, error)
	GetUsers() ([]map[string]interface{}, error)
	CreateUser(user map[string]interface{}) error
	UpdateUser(username string, updates map[string]interface{}) error
	DeleteUser(username string) error
	GetLogs(logType string, limit int) ([]map[string]interface{}, error)
	SearchSessions(filters map[string]interface{}) ([]map[string]interface{}, error)
}

// Config for web server
type Config struct {
	Port          int
	AuthService   AuthService
	ConfigManager ConfigManager
	SystemMonitor SystemMonitor
	DataProvider  DataProvider
	Logger        zerolog.Logger
}

// New creates a new web server
func New(cfg Config) *Server {
	return &Server{
		port:          cfg.Port,
		logger:        cfg.Logger,
		authService:   cfg.AuthService,
		configManager: cfg.ConfigManager,
		systemMonitor: cfg.SystemMonitor,
		dataProvider:  cfg.DataProvider,
		wsClients:     make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files from embedded FS
	staticFS, err := fs.Sub(embeddedFS, "static")
	if err != nil {
		return fmt.Errorf("failed to get static FS: %w", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Serve HTML templates
	templatesFS, err := fs.Sub(embeddedFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to get templates FS: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(templatesFS)))

	// API routes
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/logout", s.requireAuth(s.handleLogout))
	mux.HandleFunc("/api/kpi", s.requireAuth(s.handleKPIs))
	mux.HandleFunc("/api/sessions", s.requireAuth(s.handleSessions))
	mux.HandleFunc("/api/sessions/", s.requireAuth(s.handleSessionDetail))
	mux.HandleFunc("/api/alarms", s.requireAuth(s.handleAlarms))
	mux.HandleFunc("/api/alarms/", s.requireAuth(s.handleAlarmActions))
	mux.HandleFunc("/api/resources", s.requireAuth(s.handleResources))
	mux.HandleFunc("/api/license", s.requireAuth(s.handleLicense))
	mux.HandleFunc("/api/topology", s.requireAuth(s.handleTopology))
	mux.HandleFunc("/api/configuration", s.requireAuth(s.requireRole("admin", s.handleConfiguration)))
	mux.HandleFunc("/api/configuration/protocols/", s.requireAuth(s.requireRole("admin", s.handleProtocolConfig)))
	mux.HandleFunc("/api/configuration/networks/", s.requireAuth(s.requireRole("admin", s.handleNetworkConfig)))
	mux.HandleFunc("/api/system/restart", s.requireAuth(s.requireRole("admin", s.handleSystemRestart)))
	mux.HandleFunc("/api/users", s.requireAuth(s.requireRole("admin", s.handleUsers)))
	mux.HandleFunc("/api/logs", s.requireAuth(s.handleLogs))
	mux.HandleFunc("/api/search", s.requireAuth(s.handleSearch))

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info().Int("port", s.port).Msg("Starting web server")

	// Start WebSocket broadcast routine
	go s.broadcastLoop()

	return s.server.ListenAndServe()
}

// Stop stops the web server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Stopping web server")

	// Close all WebSocket connections
	s.wsClientsMux.Lock()
	for client := range s.wsClients {
		client.Close()
	}
	s.wsClientsMux.Unlock()

	return s.server.Shutdown(ctx)
}

// Middleware: CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Middleware: Require authentication
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.sendError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.sendError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]
		username, role, err := s.authService.ValidateToken(token)
		if err != nil {
			s.sendError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "username", username)
		ctx = context.WithValue(ctx, "role", role)
		next(w, r.WithContext(ctx))
	}
}

// Middleware: Require specific role
func (s *Server) requireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)

		// Admin can access everything
		if role == "admin" {
			next(w, r)
			return
		}

		if role != requiredRole {
			s.sendError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}

		next(w, r)
	}
}

// Handler: Login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		s.sendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"token":   token,
		"message": "Login successful",
	})
}

// Handler: Logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := s.authService.Logout(token); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to logout")
	}

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "Logout successful"})
}

// Handler: KPIs
func (s *Server) handleKPIs(w http.ResponseWriter, r *http.Request) {
	kpis, err := s.dataProvider.GetKPIs()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get KPIs")
		return
	}

	s.sendJSON(w, http.StatusOK, kpis)
}

// Handler: Sessions
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 100
	offset := 0

	sessions, err := s.dataProvider.GetSessions(limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get sessions")
		return
	}

	s.sendJSON(w, http.StatusOK, sessions)
}

// Handler: Session detail
func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	tid := strings.TrimPrefix(r.URL.Path, "/api/sessions/")

	session, err := s.dataProvider.GetSession(tid)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Session not found")
		return
	}

	s.sendJSON(w, http.StatusOK, session)
}

// Handler: Alarms
func (s *Server) handleAlarms(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}

	alarms, err := s.dataProvider.GetAlarms(status)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get alarms")
		return
	}

	s.sendJSON(w, http.StatusOK, alarms)
}

// Handler: Alarm actions
func (s *Server) handleAlarmActions(w http.ResponseWriter, r *http.Request) {
	alarmID := strings.TrimPrefix(r.URL.Path, "/api/alarms/")
	username := r.Context().Value("username").(string)

	if r.Method == http.MethodPost {
		// Acknowledge alarm
		if err := s.dataProvider.AcknowledgeAlarm(alarmID, username); err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to acknowledge alarm")
			return
		}
		s.sendJSON(w, http.StatusOK, map[string]string{"message": "Alarm acknowledged"})
		return
	}

	// Get alarm details
	alarm, err := s.dataProvider.GetAlarm(alarmID)
	if err != nil {
		s.sendError(w, http.StatusNotFound, "Alarm not found")
		return
	}

	s.sendJSON(w, http.StatusOK, alarm)
}

// Handler: Resources
func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	resources := map[string]interface{}{
		"cpu":    s.systemMonitor.GetCPUUsage(),
		"memory": s.systemMonitor.GetMemoryUsage(),
	}

	disk, err := s.systemMonitor.GetDiskUsage()
	if err == nil {
		resources["disk"] = disk
	}

	network, err := s.systemMonitor.GetNetworkStats()
	if err == nil {
		resources["network"] = network
	}

	process, err := s.systemMonitor.GetProcessStats()
	if err == nil {
		resources["process"] = process
	}

	s.sendJSON(w, http.StatusOK, resources)
}

// Handler: License
func (s *Server) handleLicense(w http.ResponseWriter, r *http.Request) {
	license, err := s.dataProvider.GetLicenseInfo()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get license info")
		return
	}

	s.sendJSON(w, http.StatusOK, license)
}

// Handler: Topology
func (s *Server) handleTopology(w http.ResponseWriter, r *http.Request) {
	topology, err := s.dataProvider.GetTopology()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get topology")
		return
	}

	s.sendJSON(w, http.StatusOK, topology)
}

// Handler: Configuration
func (s *Server) handleConfiguration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config, err := s.configManager.GetConfig()
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get configuration")
			return
		}
		s.sendJSON(w, http.StatusOK, config)

	case http.MethodPost, http.MethodPut:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := s.configManager.UpdateConfig(updates); err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to update configuration")
			return
		}

		s.sendJSON(w, http.StatusOK, map[string]string{"message": "Configuration updated successfully"})

	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handler: Protocol configuration
func (s *Server) handleProtocolConfig(w http.ResponseWriter, r *http.Request) {
	protocol := strings.TrimPrefix(r.URL.Path, "/api/configuration/protocols/")

	switch r.Method {
	case http.MethodGet:
		config, err := s.configManager.GetProtocolConfig(protocol)
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get protocol config")
			return
		}
		s.sendJSON(w, http.StatusOK, config)

	case http.MethodPost, http.MethodPUT:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := s.configManager.UpdateProtocolConfig(protocol, updates); err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to update protocol config")
			return
		}

		s.sendJSON(w, http.StatusOK, map[string]string{"message": "Protocol config updated"})

	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handler: Network configuration
func (s *Server) handleNetworkConfig(w http.ResponseWriter, r *http.Request) {
	network := strings.TrimPrefix(r.URL.Path, "/api/configuration/networks/")

	switch r.Method {
	case http.MethodGet:
		config, err := s.configManager.GetNetworkConfig(network)
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get network config")
			return
		}
		s.sendJSON(w, http.StatusOK, config)

	case http.MethodPost, http.MethodPUT:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := s.configManager.UpdateNetworkConfig(network, updates); err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to update network config")
			return
		}

		s.sendJSON(w, http.StatusOK, map[string]string{"message": "Network config updated"})

	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handler: System restart
func (s *Server) handleSystemRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{"message": "System restart initiated"})

	// Restart in background
	go func() {
		time.Sleep(2 * time.Second)
		if err := s.configManager.RestartService(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to restart service")
		}
	}()
}

// Handler: Users
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := s.dataProvider.GetUsers()
		if err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to get users")
			return
		}
		s.sendJSON(w, http.StatusOK, users)

	case http.MethodPost:
		var user map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			s.sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := s.dataProvider.CreateUser(user); err != nil {
			s.sendError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		s.sendJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})

	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Handler: Logs
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("type")
	if logType == "" {
		logType = "application"
	}

	limit := 1000

	logs, err := s.dataProvider.GetLogs(logType, limit)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get logs")
		return
	}

	s.sendJSON(w, http.StatusOK, logs)
}

// Handler: Search
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var filters map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessions, err := s.dataProvider.SearchSessions(filters)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to search sessions")
		return
	}

	s.sendJSON(w, http.StatusOK, sessions)
}

// Handler: WebSocket
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Validate token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		s.logger.Warn().Msg("WebSocket connection without token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, _, err := s.authService.ValidateToken(token)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Invalid WebSocket token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Register client
	s.wsClientsMux.Lock()
	s.wsClients[conn] = true
	s.wsClientsMux.Unlock()

	s.logger.Info().Msg("New WebSocket client connected")

	// Handle disconnect
	defer func() {
		s.wsClientsMux.Lock()
		delete(s.wsClients, conn)
		s.wsClientsMux.Unlock()
		conn.Close()
		s.logger.Info().Msg("WebSocket client disconnected")
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Broadcast to all WebSocket clients
func (s *Server) Broadcast(messageType string, payload interface{}) {
	message := map[string]interface{}{
		"type":      messageType,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	s.wsClientsMux.RLock()
	defer s.wsClientsMux.RUnlock()

	for client := range s.wsClients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			s.logger.Warn().Err(err).Msg("Failed to send WebSocket message")
		}
	}
}

// Broadcast loop for periodic updates
func (s *Server) broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Broadcast resource updates
		resources := map[string]interface{}{
			"cpu":    s.systemMonitor.GetCPUUsage(),
			"memory": s.systemMonitor.GetMemoryUsage(),
		}
		s.Broadcast("resource_update", resources)

		// Broadcast KPI updates
		if kpis, err := s.dataProvider.GetKPIs(); err == nil {
			s.Broadcast("kpi_update", kpis)
		}
	}
}

// Handler: Health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"version": "2.0.0",
		"uptime":  time.Since(time.Now()).Seconds(),
		"go_version": runtime.Version(),
		"hostname": getHostname(),
	}

	s.sendJSON(w, http.StatusOK, health)
}

// Helper: Send JSON response
func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

// Helper: Send error response
func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, map[string]string{"error": message})
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
