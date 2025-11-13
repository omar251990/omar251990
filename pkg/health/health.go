package health

import (
	"sync"
	"time"
)

// HealthCheck monitors application health
type HealthCheck struct {
	config     *Config
	status     *Status
	lastCheck  time.Time
	mu         sync.RWMutex
}

// Config holds health check configuration
type Config struct {
	Enabled          bool
	CheckInterval    time.Duration
	WatchdogEnabled  bool
	WatchdogTimeout  time.Duration
	RestartOnFailure bool
}

// Status represents the health status
type Status struct {
	Healthy            bool
	Timestamp          time.Time
	UptimeSeconds      int64
	MessagesProcessed  int64
	SessionsActive     int64
	ErrorCount         int64
	LastError          string
	ComponentStatus    map[string]ComponentStatus
}

// ComponentStatus represents the status of a component
type ComponentStatus struct {
	Name      string
	Healthy   bool
	Message   string
	LastCheck time.Time
}

// NewHealthCheck creates a new health check instance
func NewHealthCheck(config *Config) *HealthCheck {
	hc := &HealthCheck{
		config: config,
		status: &Status{
			Healthy:         true,
			Timestamp:       time.Now(),
			ComponentStatus: make(map[string]ComponentStatus),
		},
		lastCheck: time.Now(),
	}

	if config.Enabled {
		go hc.checkLoop()
	}

	if config.WatchdogEnabled {
		go hc.watchdogLoop()
	}

	return hc
}

// GetStatus returns the current health status
func (h *HealthCheck) GetStatus() *Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Create a copy
	statusCopy := *h.status
	statusCopy.ComponentStatus = make(map[string]ComponentStatus)
	for k, v := range h.status.ComponentStatus {
		statusCopy.ComponentStatus[k] = v
	}

	return &statusCopy
}

// UpdateComponentStatus updates the status of a component
func (h *HealthCheck) UpdateComponentStatus(name string, healthy bool, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.status.ComponentStatus[name] = ComponentStatus{
		Name:      name,
		Healthy:   healthy,
		Message:   message,
		LastCheck: time.Now(),
	}

	// Update overall health
	h.updateOverallHealth()
}

// RecordMessage increments the message counter
func (h *HealthCheck) RecordMessage() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status.MessagesProcessed++
}

// RecordError increments the error counter
func (h *HealthCheck) RecordError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status.ErrorCount++
	h.status.LastError = err.Error()
}

// UpdateSessionCount updates the active session count
func (h *HealthCheck) UpdateSessionCount(count int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status.SessionsActive = count
}

// checkLoop performs periodic health checks
func (h *HealthCheck) checkLoop() {
	ticker := time.NewTicker(h.config.CheckInterval)
	defer ticker.Stop()

	startTime := time.Now()

	for range ticker.C {
		h.mu.Lock()
		h.status.Timestamp = time.Now()
		h.status.UptimeSeconds = int64(time.Since(startTime).Seconds())
		h.lastCheck = time.Now()
		h.updateOverallHealth()
		h.mu.Unlock()
	}
}

// watchdogLoop monitors for application hangs
func (h *HealthCheck) watchdogLoop() {
	ticker := time.NewTicker(h.config.WatchdogTimeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		timeSinceLastCheck := time.Since(h.lastCheck)
		h.mu.RUnlock()

		if timeSinceLastCheck > h.config.WatchdogTimeout {
			// Application may be hanging
			if h.config.RestartOnFailure {
				// Trigger restart (in production, this would signal supervisor)
				panic("Watchdog timeout - application not responding")
			}
		}
	}
}

// updateOverallHealth determines overall health from component statuses
func (h *HealthCheck) updateOverallHealth() {
	h.status.Healthy = true

	for _, component := range h.status.ComponentStatus {
		if !component.Healthy {
			h.status.Healthy = false
			break
		}
	}
}

// IsHealthy returns true if the application is healthy
func (h *HealthCheck) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status.Healthy
}
