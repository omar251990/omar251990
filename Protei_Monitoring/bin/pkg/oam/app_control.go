package oam

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// ApplicationStatus represents the application status
type ApplicationStatus string

const (
	StatusRunning ApplicationStatus = "running"
	StatusStopped ApplicationStatus = "stopped"
	StatusStarting ApplicationStatus = "starting"
	StatusStopping ApplicationStatus = "stopping"
	StatusError    ApplicationStatus = "error"
	StatusUnknown  ApplicationStatus = "unknown"
)

// AppController controls the application lifecycle
type AppController struct {
	mu            sync.RWMutex
	installDir    string
	pidFile       string
	status        ApplicationStatus
	startTime     time.Time
	stopTime      time.Time
	restartCount  int
	lastError     error
}

// AppInfo holds application information
type AppInfo struct {
	Status       ApplicationStatus
	PID          int
	Uptime       time.Duration
	StartTime    time.Time
	Version      string
	BuildDate    string
	GitCommit    string
	RestartCount int
	LastError    string
}

// NewAppController creates a new application controller
func NewAppController(installDir string) *AppController {
	return &AppController{
		installDir: installDir,
		pidFile:    filepath.Join(installDir, "tmp", "protei-monitoring.pid"),
		status:     StatusUnknown,
	}
}

// GetStatus returns the current application status
func (c *AppController) GetStatus() (*AppInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := &AppInfo{
		Status:       c.status,
		RestartCount: c.restartCount,
	}

	if c.lastError != nil {
		info.LastError = c.lastError.Error()
	}

	// Check if process is running
	if pid, running := c.isProcessRunning(); running {
		info.PID = pid
		info.Status = StatusRunning
		info.StartTime = c.startTime
		if !c.startTime.IsZero() {
			info.Uptime = time.Since(c.startTime)
		}
	} else {
		info.Status = StatusStopped
	}

	// Get version info
	version, buildDate, gitCommit := c.getVersionInfo()
	info.Version = version
	info.BuildDate = buildDate
	info.GitCommit = gitCommit

	return info, nil
}

// Start starts the application
func (c *AppController) Start() error {
	c.mu.Lock()
	c.status = StatusStarting
	c.mu.Unlock()

	// Check if already running
	if pid, running := c.isProcessRunning(); running {
		c.mu.Lock()
		c.status = StatusRunning
		c.mu.Unlock()
		return fmt.Errorf("application is already running (PID: %d)", pid)
	}

	// Execute start script
	startScript := filepath.Join(c.installDir, "scripts", "start")
	cmd := exec.Command(startScript)
	cmd.Dir = c.installDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.mu.Lock()
		c.status = StatusError
		c.lastError = fmt.Errorf("start failed: %w\nOutput: %s", err, string(output))
		c.mu.Unlock()
		return c.lastError
	}

	// Wait for process to start
	time.Sleep(2 * time.Second)

	// Verify startup
	if _, running := c.isProcessRunning(); !running {
		c.mu.Lock()
		c.status = StatusError
		c.lastError = fmt.Errorf("application failed to start")
		c.mu.Unlock()
		return c.lastError
	}

	c.mu.Lock()
	c.status = StatusRunning
	c.startTime = time.Now()
	c.lastError = nil
	c.mu.Unlock()

	return nil
}

// Stop stops the application
func (c *AppController) Stop() error {
	c.mu.Lock()
	c.status = StatusStopping
	c.mu.Unlock()

	// Check if already stopped
	if _, running := c.isProcessRunning(); !running {
		c.mu.Lock()
		c.status = StatusStopped
		c.mu.Unlock()
		return fmt.Errorf("application is not running")
	}

	// Execute stop script
	stopScript := filepath.Join(c.installDir, "scripts", "stop")
	cmd := exec.Command(stopScript)
	cmd.Dir = c.installDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.mu.Lock()
		c.status = StatusError
		c.lastError = fmt.Errorf("stop failed: %w\nOutput: %s", err, string(output))
		c.mu.Unlock()
		return c.lastError
	}

	// Wait for process to stop
	time.Sleep(2 * time.Second)

	// Verify shutdown
	if _, running := c.isProcessRunning(); running {
		c.mu.Lock()
		c.status = StatusError
		c.lastError = fmt.Errorf("application failed to stop gracefully")
		c.mu.Unlock()
		return c.lastError
	}

	c.mu.Lock()
	c.status = StatusStopped
	c.stopTime = time.Now()
	c.lastError = nil
	c.mu.Unlock()

	return nil
}

// Restart restarts the application
func (c *AppController) Restart() error {
	// Stop first
	if err := c.Stop(); err != nil && !contains(err.Error(), "not running") {
		return err
	}

	// Wait a bit
	time.Sleep(3 * time.Second)

	// Start
	if err := c.Start(); err != nil {
		return err
	}

	c.mu.Lock()
	c.restartCount++
	c.mu.Unlock()

	return nil
}

// Reload reloads the application configuration without restart
func (c *AppController) Reload() error {
	// Check if running
	pid, running := c.isProcessRunning()
	if !running {
		return fmt.Errorf("application is not running")
	}

	// Execute reload script
	reloadScript := filepath.Join(c.installDir, "scripts", "reload")
	cmd := exec.Command(reloadScript)
	cmd.Dir = c.installDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.mu.Lock()
		c.lastError = fmt.Errorf("reload failed: %w\nOutput: %s", err, string(output))
		c.mu.Unlock()
		return c.lastError
	}

	// Alternative: Send SIGHUP signal
	if err := c.sendSignal(pid, "HUP"); err != nil {
		c.mu.Lock()
		c.lastError = err
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.lastError = nil
	c.mu.Unlock()

	return nil
}

// isProcessRunning checks if the application process is running
func (c *AppController) isProcessRunning() (int, bool) {
	// Read PID file
	pidData, err := os.ReadFile(c.pidFile)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return 0, false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(os.Signal(nil))
	if err != nil {
		return 0, false
	}

	return pid, true
}

// sendSignal sends a signal to the application process
func (c *AppController) sendSignal(pid int, signal string) error {
	cmd := exec.Command("kill", "-"+signal, strconv.Itoa(pid))
	return cmd.Run()
}

// getVersionInfo retrieves version information
func (c *AppController) getVersionInfo() (version, buildDate, gitCommit string) {
	// Try to execute binary with --version flag
	binaryPath := filepath.Join(c.installDir, "bin", "protei-monitoring")
	cmd := exec.Command(binaryPath, "--version")

	output, err := cmd.Output()
	if err == nil {
		// Parse version output
		// Expected format: "Protei Monitoring v2.0.0 (build: 2024-01-15, commit: abc123)"
		versionStr := string(output)
		version = extractValue(versionStr, "v", " ")
		buildDate = extractValue(versionStr, "build: ", ",")
		gitCommit = extractValue(versionStr, "commit: ", ")")
	}

	// Fallback to default values
	if version == "" {
		version = "2.0.0"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}
	if gitCommit == "" {
		gitCommit = "unknown"
	}

	return
}

// GetLogs retrieves application logs
func (c *AppController) GetLogs(logType string, lines int) ([]string, error) {
	var logFile string

	switch logType {
	case "application":
		logFile = filepath.Join(c.installDir, "logs", "application", "protei-monitoring.log")
	case "system":
		logFile = filepath.Join(c.installDir, "logs", "system", "system.log")
	case "error":
		logFile = filepath.Join(c.installDir, "logs", "error", "error.log")
	case "access":
		logFile = filepath.Join(c.installDir, "logs", "access", "access.log")
	default:
		return nil, fmt.Errorf("unknown log type: %s", logType)
	}

	// Use tail to get last N lines
	cmd := exec.Command("tail", "-n", strconv.Itoa(lines), logFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	// Split into lines
	logLines := splitLines(string(output))

	return logLines, nil
}

// GetSystemMetrics retrieves system metrics
func (c *AppController) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{}

	// CPU usage
	cpuCmd := exec.Command("ps", "-p", "-1", "-o", "%cpu")
	if output, err := cpuCmd.Output(); err == nil {
		if _, err := fmt.Sscanf(string(output), "%f", &metrics.CPUPercent); err == nil {
			// Successfully parsed
		}
	}

	// Memory usage
	memCmd := exec.Command("ps", "-p", "-1", "-o", "rss")
	if output, err := memCmd.Output(); err == nil {
		var rssKB int64
		if _, err := fmt.Sscanf(string(output), "%d", &rssKB); err == nil {
			metrics.MemoryMB = float64(rssKB) / 1024.0
		}
	}

	// Disk usage
	duCmd := exec.Command("df", "-h", c.installDir)
	if output, err := duCmd.Output(); err == nil {
		// Parse df output
		lines := splitLines(string(output))
		if len(lines) > 1 {
			fields := splitFields(lines[1])
			if len(fields) >= 5 {
				metrics.DiskUsage = fields[4] // Percentage
			}
		}
	}

	// Network connections
	netCmd := exec.Command("netstat", "-tn")
	if output, err := netCmd.Output(); err == nil {
		lines := splitLines(string(output))
		metrics.NetworkConnections = len(lines) - 2 // Exclude header lines
	}

	// Goroutines, threads (would need to query app API)
	metrics.Goroutines = 0 // TODO: Implement via app API
	metrics.Threads = 0    // TODO: Implement via app API

	return metrics, nil
}

// ExecuteHealthCheck performs application health check
func (c *AppController) ExecuteHealthCheck() (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckStatus),
	}

	// Check 1: Process running
	if _, running := c.isProcessRunning(); running {
		result.Checks["process"] = CheckStatus{Healthy: true, Message: "Process is running"}
	} else {
		result.Checks["process"] = CheckStatus{Healthy: false, Message: "Process is not running"}
		result.Healthy = false
		return result, nil
	}

	// Check 2: Web server responding
	webCheck := c.checkWebServer()
	result.Checks["web_server"] = webCheck
	if !webCheck.Healthy {
		result.Healthy = false
	}

	// Check 3: Database connectivity
	dbCheck := c.checkDatabase()
	result.Checks["database"] = dbCheck
	if !dbCheck.Healthy {
		result.Healthy = false
	}

	// Check 4: Disk space
	diskCheck := c.checkDiskSpace()
	result.Checks["disk_space"] = diskCheck
	if !diskCheck.Healthy {
		result.Healthy = false
	}

	// Overall health
	result.Healthy = true
	for _, check := range result.Checks {
		if !check.Healthy {
			result.Healthy = false
			break
		}
	}

	return result, nil
}

// checkWebServer checks if web server is responding
func (c *AppController) checkWebServer() CheckStatus {
	// Try to curl the health endpoint
	cmd := exec.Command("curl", "-f", "-s", "http://localhost:8080/health")
	if err := cmd.Run(); err != nil {
		return CheckStatus{Healthy: false, Message: "Web server not responding"}
	}
	return CheckStatus{Healthy: true, Message: "Web server is healthy"}
}

// checkDatabase checks database connectivity
func (c *AppController) checkDatabase() CheckStatus {
	// Execute database check script
	checkScript := filepath.Join(c.installDir, "scripts", "utils", "check_db.sh")
	cmd := exec.Command(checkScript)
	if err := cmd.Run(); err != nil {
		return CheckStatus{Healthy: false, Message: "Database connection failed"}
	}
	return CheckStatus{Healthy: true, Message: "Database is accessible"}
}

// checkDiskSpace checks available disk space
func (c *AppController) checkDiskSpace() CheckStatus {
	cmd := exec.Command("df", "-h", c.installDir)
	output, err := cmd.Output();
	if err != nil {
		return CheckStatus{Healthy: false, Message: "Failed to check disk space"}
	}

	lines := splitLines(string(output))
	if len(lines) > 1 {
		fields := splitFields(lines[1])
		if len(fields) >= 5 {
			usageStr := fields[4]
			// Remove % sign
			usageStr = usageStr[:len(usageStr)-1]
			usage, _ := strconv.Atoi(usageStr)

			if usage > 90 {
				return CheckStatus{
					Healthy: false,
					Message: fmt.Sprintf("Disk usage critical: %d%%", usage),
				}
			} else if usage > 80 {
				return CheckStatus{
					Healthy: true,
					Message: fmt.Sprintf("Disk usage warning: %d%%", usage),
				}
			}

			return CheckStatus{
				Healthy: true,
				Message: fmt.Sprintf("Disk usage normal: %d%%", usage),
			}
		}
	}

	return CheckStatus{Healthy: true, Message: "Disk space check inconclusive"}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func extractValue(text, start, end string) string {
	startIdx := len(start)
	if idx := indexOf(text, start); idx >= 0 {
		text = text[idx+startIdx:]
		if idx := indexOf(text, end); idx >= 0 {
			return text[:idx]
		}
		return text
	}
	return ""
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitLines(s string) []string {
	lines := make([]string, 0)
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitFields(s string) []string {
	fields := make([]string, 0)
	current := ""
	inSpace := false

	for _, ch := range s {
		if ch == ' ' || ch == '\t' {
			if !inSpace && current != "" {
				fields = append(fields, current)
				current = ""
			}
			inSpace = true
		} else {
			inSpace = false
			current += string(ch)
		}
	}

	if current != "" {
		fields = append(fields, current)
	}

	return fields
}

// Data structures

type SystemMetrics struct {
	CPUPercent          float64
	MemoryMB            float64
	DiskUsage           string
	NetworkConnections  int
	Goroutines          int
	Threads             int
}

type HealthCheckResult struct {
	Timestamp time.Time
	Healthy   bool
	Checks    map[string]CheckStatus
}

type CheckStatus struct {
	Healthy bool
	Message string
}
