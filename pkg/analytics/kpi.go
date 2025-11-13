package analytics

import (
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/correlation"
	"github.com/protei/monitoring/pkg/decoder"
)

// KPIEngine calculates Key Performance Indicators
type KPIEngine struct {
	config     *Config
	metrics    map[string]*ProcedureMetrics
	metricsMu  sync.RWMutex
	roaming    *RoamingAnalytics
}

// Config holds analytics configuration
type Config struct {
	Enabled             bool
	CalculationInterval time.Duration
	Procedures          []string
	Metrics             []string
	RoamingEnabled      bool
	FailureThreshold    float64
	LatencyThreshold    int
}

// ProcedureMetrics holds metrics for a specific procedure
type ProcedureMetrics struct {
	Procedure      string
	TotalCount     int64
	SuccessCount   int64
	FailureCount   int64
	TimeoutCount   int64
	SuccessRate    float64
	FailureRate    float64
	Latencies      []int64 // microseconds
	LatencyAvg     int64
	LatencyP95     int64
	LatencyP99     int64
	CauseCodes     map[int]int
	LastUpdate     time.Time
	mu             sync.Mutex
}

// RoamingAnalytics tracks roaming-specific KPIs
type RoamingAnalytics struct {
	InboundRoamers  map[string]*RoamerMetrics // PLMN -> metrics
	OutboundRoamers map[string]*RoamerMetrics // PLMN -> metrics
	CellMetrics     map[string]*CellMetrics   // CellID -> metrics
	mu              sync.RWMutex
}

// RoamerMetrics holds roaming metrics per PLMN
type RoamerMetrics struct {
	PLMN              string
	RoamerCount       int64
	SessionCount      int64
	SuccessfulSessions int64
	FailedSessions     int64
	SuccessRate        float64
	TotalDataVolume    int64
	APNUsage           map[string]int64
}

// CellMetrics holds per-cell metrics
type CellMetrics struct {
	CellID        string
	PLMN          string
	RoamerCount   int
	Procedures    map[string]int
	SuccessRate   float64
	LastUpdate    time.Time
}

// KPIReport represents a KPI calculation result
type KPIReport struct {
	Timestamp      time.Time
	Period         time.Duration
	Procedures     map[string]*ProcedureMetrics
	RoamingMetrics *RoamingMetrics
	Alerts         []Alert
}

// Alert represents a KPI alert
type Alert struct {
	Severity   string // critical, high, medium, low
	Procedure  string
	Message    string
	Value      float64
	Threshold  float64
	Timestamp  time.Time
}

// NewKPIEngine creates a new KPI engine
func NewKPIEngine(config *Config) *KPIEngine {
	engine := &KPIEngine{
		config:  config,
		metrics: make(map[string]*ProcedureMetrics),
		roaming: &RoamingAnalytics{
			InboundRoamers:  make(map[string]*RoamerMetrics),
			OutboundRoamers: make(map[string]*RoamerMetrics),
			CellMetrics:     make(map[string]*CellMetrics),
		},
	}

	// Initialize metrics for configured procedures
	for _, procedure := range config.Procedures {
		engine.metrics[procedure] = &ProcedureMetrics{
			Procedure:  procedure,
			CauseCodes: make(map[int]int),
			Latencies:  make([]int64, 0, 10000),
		}
	}

	// Start periodic calculation
	go engine.periodicCalculation()

	return engine
}

// ProcessSession processes a completed session for KPI calculation
func (e *KPIEngine) ProcessSession(session *correlation.Session) {
	if session.Procedure == "" {
		return
	}

	// Get or create metrics for this procedure
	e.metricsMu.RLock()
	metrics, exists := e.metrics[session.Procedure]
	e.metricsMu.RUnlock()

	if !exists {
		e.metricsMu.Lock()
		metrics = &ProcedureMetrics{
			Procedure:  session.Procedure,
			CauseCodes: make(map[int]int),
			Latencies:  make([]int64, 0, 10000),
		}
		e.metrics[session.Procedure] = metrics
		e.metricsMu.Unlock()
	}

	// Update metrics
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.TotalCount++

	switch session.Result {
	case decoder.ResultSuccess:
		metrics.SuccessCount++
	case decoder.ResultFailure:
		metrics.FailureCount++
		if session.FailureCause != 0 {
			metrics.CauseCodes[session.FailureCause]++
		}
	case decoder.ResultTimeout:
		metrics.TimeoutCount++
	}

	// Record latency
	if session.Duration > 0 {
		latencyUs := session.Duration.Microseconds()
		metrics.Latencies = append(metrics.Latencies, latencyUs)

		// Limit latency array size (keep last 10000)
		if len(metrics.Latencies) > 10000 {
			metrics.Latencies = metrics.Latencies[len(metrics.Latencies)-10000:]
		}
	}

	metrics.LastUpdate = time.Now()

	// Process roaming if enabled
	if e.config.RoamingEnabled && session.PLMN != "" {
		e.processRoaming(session)
	}
}

// processRoaming processes roaming-specific metrics
func (e *KPIEngine) processRoaming(session *correlation.Session) {
	e.roaming.mu.Lock()
	defer e.roaming.mu.Unlock()

	// Determine roaming direction (simplified - should compare with home PLMN)
	isInbound := e.isInboundRoamer(session.PLMN)

	var roamerMap map[string]*RoamerMetrics
	if isInbound {
		roamerMap = e.roaming.InboundRoamers
	} else {
		roamerMap = e.roaming.OutboundRoamers
	}

	// Get or create roamer metrics
	metrics, exists := roamerMap[session.PLMN]
	if !exists {
		metrics = &RoamerMetrics{
			PLMN:     session.PLMN,
			APNUsage: make(map[string]int64),
		}
		roamerMap[session.PLMN] = metrics
	}

	// Update metrics
	metrics.SessionCount++
	if session.Result == decoder.ResultSuccess {
		metrics.SuccessfulSessions++
	} else if session.Result == decoder.ResultFailure {
		metrics.FailedSessions++
	}

	// Update APN usage
	if session.APN != "" {
		metrics.APNUsage[session.APN]++
	}

	// Update cell metrics
	if session.CellID != "" {
		cellMetrics, exists := e.roaming.CellMetrics[session.CellID]
		if !exists {
			cellMetrics = &CellMetrics{
				CellID:     session.CellID,
				PLMN:       session.PLMN,
				Procedures: make(map[string]int),
			}
			e.roaming.CellMetrics[session.CellID] = cellMetrics
		}

		cellMetrics.RoamerCount++
		cellMetrics.Procedures[session.Procedure]++
		cellMetrics.LastUpdate = time.Now()
	}
}

// Calculate performs KPI calculation and returns a report
func (e *KPIEngine) Calculate() *KPIReport {
	report := &KPIReport{
		Timestamp:  time.Now(),
		Period:     e.config.CalculationInterval,
		Procedures: make(map[string]*ProcedureMetrics),
		Alerts:     make([]Alert, 0),
	}

	e.metricsMu.RLock()
	defer e.metricsMu.RUnlock()

	// Calculate metrics for each procedure
	for procedure, metrics := range e.metrics {
		metrics.mu.Lock()

		// Calculate success/failure rates
		if metrics.TotalCount > 0 {
			metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.TotalCount) * 100
			metrics.FailureRate = float64(metrics.FailureCount) / float64(metrics.TotalCount) * 100
		}

		// Calculate latency percentiles
		if len(metrics.Latencies) > 0 {
			metrics.LatencyAvg = e.calculateAverage(metrics.Latencies)
			metrics.LatencyP95 = e.calculatePercentile(metrics.Latencies, 95)
			metrics.LatencyP99 = e.calculatePercentile(metrics.Latencies, 99)
		}

		// Create a copy for the report
		reportMetrics := &ProcedureMetrics{
			Procedure:    metrics.Procedure,
			TotalCount:   metrics.TotalCount,
			SuccessCount: metrics.SuccessCount,
			FailureCount: metrics.FailureCount,
			TimeoutCount: metrics.TimeoutCount,
			SuccessRate:  metrics.SuccessRate,
			FailureRate:  metrics.FailureRate,
			LatencyAvg:   metrics.LatencyAvg,
			LatencyP95:   metrics.LatencyP95,
			LatencyP99:   metrics.LatencyP99,
			CauseCodes:   make(map[int]int),
			LastUpdate:   metrics.LastUpdate,
		}

		// Copy cause codes
		for code, count := range metrics.CauseCodes {
			reportMetrics.CauseCodes[code] = count
		}

		report.Procedures[procedure] = reportMetrics

		// Check for alerts
		if metrics.FailureRate > e.config.FailureThreshold {
			report.Alerts = append(report.Alerts, Alert{
				Severity:  "high",
				Procedure: procedure,
				Message:   "High failure rate detected",
				Value:     metrics.FailureRate,
				Threshold: e.config.FailureThreshold,
				Timestamp: time.Now(),
			})
		}

		if metrics.LatencyP95 > int64(e.config.LatencyThreshold)*1000 {
			report.Alerts = append(report.Alerts, Alert{
				Severity:  "medium",
				Procedure: procedure,
				Message:   "High latency detected (P95)",
				Value:     float64(metrics.LatencyP95) / 1000,
				Threshold: float64(e.config.LatencyThreshold),
				Timestamp: time.Now(),
			})
		}

		metrics.mu.Unlock()
	}

	// Calculate roaming metrics
	e.calculateRoamingMetrics()

	return report
}

// calculateRoamingMetrics calculates roaming KPIs
func (e *KPIEngine) calculateRoamingMetrics() {
	e.roaming.mu.Lock()
	defer e.roaming.mu.Unlock()

	// Calculate success rates for inbound roamers
	for _, metrics := range e.roaming.InboundRoamers {
		if metrics.SessionCount > 0 {
			metrics.SuccessRate = float64(metrics.SuccessfulSessions) / float64(metrics.SessionCount) * 100
		}
	}

	// Calculate success rates for outbound roamers
	for _, metrics := range e.roaming.OutboundRoamers {
		if metrics.SessionCount > 0 {
			metrics.SuccessRate = float64(metrics.SuccessfulSessions) / float64(metrics.SessionCount) * 100
		}
	}

	// Calculate cell success rates
	for _, cell := range e.roaming.CellMetrics {
		totalProcedures := 0
		for _, count := range cell.Procedures {
			totalProcedures += count
		}
		// Simplified - should track successes separately
		cell.SuccessRate = 95.0 // Placeholder
	}
}

// calculateAverage calculates the average of a slice
func (e *KPIEngine) calculateAverage(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	var sum int64
	for _, v := range values {
		sum += v
	}

	return sum / int64(len(values))
}

// calculatePercentile calculates the Nth percentile
func (e *KPIEngine) calculatePercentile(values []int64, percentile int) int64 {
	if len(values) == 0 {
		return 0
	}

	// Simple percentile calculation (should use proper algorithm)
	sorted := make([]int64, len(values))
	copy(sorted, values)

	// Bubble sort (for simplicity - use quicksort for production)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (len(sorted) * percentile) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// periodicCalculation performs periodic KPI calculation
func (e *KPIEngine) periodicCalculation() {
	ticker := time.NewTicker(e.config.CalculationInterval)
	defer ticker.Stop()

	for range ticker.C {
		report := e.Calculate()
		// Report would be published to subscribers/dashboard
		_ = report
	}
}

// GetMetrics returns current metrics for a procedure
func (e *KPIEngine) GetMetrics(procedure string) *ProcedureMetrics {
	e.metricsMu.RLock()
	defer e.metricsMu.RUnlock()
	return e.metrics[procedure]
}

// GetRoamingMetrics returns roaming analytics
func (e *KPIEngine) GetRoamingMetrics() *RoamingAnalytics {
	return e.roaming
}

// GetCellHeatmap returns cell-based heatmap data
func (e *KPIEngine) GetCellHeatmap() map[string]*CellMetrics {
	e.roaming.mu.RLock()
	defer e.roaming.mu.RUnlock()

	// Return a copy
	heatmap := make(map[string]*CellMetrics)
	for cellID, metrics := range e.roaming.CellMetrics {
		heatmap[cellID] = metrics
	}
	return heatmap
}

// isInboundRoamer determines if a PLMN is inbound (simplified)
func (e *KPIEngine) isInboundRoamer(plmn string) bool {
	// Simplified logic - should compare against home PLMN list
	// For now, assume all foreign PLMNs are inbound
	return plmn != "310410" // Example home PLMN
}

// Reset resets all metrics (for testing or periodic reset)
func (e *KPIEngine) Reset() {
	e.metricsMu.Lock()
	defer e.metricsMu.Unlock()

	for _, metrics := range e.metrics {
		metrics.mu.Lock()
		metrics.TotalCount = 0
		metrics.SuccessCount = 0
		metrics.FailureCount = 0
		metrics.TimeoutCount = 0
		metrics.Latencies = metrics.Latencies[:0]
		metrics.CauseCodes = make(map[int]int)
		metrics.mu.Unlock()
	}
}
