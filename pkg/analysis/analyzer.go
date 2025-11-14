package analysis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
	"github.com/protei/monitoring/pkg/knowledge"
	"github.com/rs/zerolog"
)

// IssueDetected represents a detected issue
type IssueDetected struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	Severity        string                 `json:"severity"` // critical, major, minor, warning
	Category        string                 `json:"category"` // protocol_error, timeout, abnormal_pattern, config_issue, performance
	Protocol        string                 `json:"protocol"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	RootCause       string                 `json:"root_cause"`
	Recommendations []string               `json:"recommendations"`
	StandardRef     string                 `json:"standard_ref"`
	AffectedIMSI    string                 `json:"affected_imsi,omitempty"`
	AffectedMSISDN  string                 `json:"affected_msisdn,omitempty"`
	ErrorCode       int                    `json:"error_code,omitempty"`
	RelatedMessages []string               `json:"related_messages,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisEngine performs intelligent traffic analysis
type AnalysisEngine struct {
	mu              sync.RWMutex
	kb              *knowledge.KnowledgeBase
	logger          zerolog.Logger
	rules           []*DetectionRule
	issueHistory    map[string]*IssueDetected
	statistics      *Statistics
	anomalyDetector *AnomalyDetector
}

// DetectionRule represents a detection rule
type DetectionRule struct {
	ID          string
	Name        string
	Description string
	Protocol    string
	Condition   func(*decoder.Message, *Statistics) bool
	Severity    string
	Category    string
	Action      func(*decoder.Message, *knowledge.KnowledgeBase) *IssueDetected
}

// Statistics holds traffic statistics
type Statistics struct {
	mu                sync.RWMutex
	TotalMessages     int64
	ErrorsByProtocol  map[string]int64
	ErrorsByCode      map[string]map[int]int64 // Protocol -> Code -> Count
	SuccessRate       map[string]float64       // Protocol -> Success %
	AvgLatency        map[string]float64       // Procedure -> Latency ms
	TimeoutCount      int64
	RecentErrors      []*ErrorOccurrence
	ProcedureCounts   map[string]int64
	ProcedureFailures map[string]int64
}

// ErrorOccurrence tracks an error occurrence
type ErrorOccurrence struct {
	Timestamp time.Time
	Protocol  string
	ErrorCode int
	ErrorName string
	IMSI      string
	Count     int
}

// AnomalyDetector detects abnormal patterns
type AnomalyDetector struct {
	mu               sync.RWMutex
	baselineRates    map[string]float64 // Protocol -> Normal rate
	baselineLatency  map[string]float64 // Procedure -> Normal latency
	recentSamples    map[string][]float64
	anomalyThreshold float64 // Deviation threshold (e.g., 2.0 = 2 standard deviations)
}

// NewAnalysisEngine creates a new analysis engine
func NewAnalysisEngine(kb *knowledge.KnowledgeBase, logger zerolog.Logger) *AnalysisEngine {
	ae := &AnalysisEngine{
		kb:           kb,
		logger:       logger,
		issueHistory: make(map[string]*IssueDetected),
		statistics: &Statistics{
			ErrorsByProtocol:  make(map[string]int64),
			ErrorsByCode:      make(map[string]map[int]int64),
			SuccessRate:       make(map[string]float64),
			AvgLatency:        make(map[string]float64),
			RecentErrors:      make([]*ErrorOccurrence, 0),
			ProcedureCounts:   make(map[string]int64),
			ProcedureFailures: make(map[string]int64),
		},
		anomalyDetector: &AnomalyDetector{
			baselineRates:    make(map[string]float64),
			baselineLatency:  make(map[string]float64),
			recentSamples:    make(map[string][]float64),
			anomalyThreshold: 2.0,
		},
	}

	ae.initializeRules()
	return ae
}

// Initialize detection rules
func (ae *AnalysisEngine) initializeRules() {
	ae.rules = []*DetectionRule{
		// Diameter: DIAMETER_ERROR_USER_UNKNOWN
		{
			ID:          "DIAMETER_USER_UNKNOWN",
			Name:        "Diameter User Unknown Error",
			Description: "HSS reports DIAMETER_ERROR_USER_UNKNOWN (5001)",
			Protocol:    decoder.ProtocolDiameter,
			Severity:    "major",
			Category:    "protocol_error",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				if msg.Protocol != decoder.ProtocolDiameter {
					return false
				}
				// Check for result code 5001
				if resultCode, ok := msg.Attributes["result_code"].(int); ok {
					return resultCode == 5001
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				errRef, _ := kb.GetErrorCode("Diameter", 5001)
				issue := &IssueDetected{
					ID:          fmt.Sprintf("DIAM_5001_%d", time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "major",
					Category:    "protocol_error",
					Protocol:    "Diameter",
					Title:       "Subscriber Not Found in HSS",
					Description: "HSS returned DIAMETER_ERROR_USER_UNKNOWN indicating the IMSI is not provisioned.",
					ErrorCode:   5001,
				}

				if errRef != nil {
					issue.RootCause = errRef.Causes
					issue.Recommendations = strings.Split(errRef.Solutions, ". ")
					issue.StandardRef = errRef.StandardRef
				}

				if imsi, ok := msg.Attributes["imsi"].(string); ok {
					issue.AffectedIMSI = imsi
					issue.Description = fmt.Sprintf("HSS returned DIAMETER_ERROR_USER_UNKNOWN for IMSI %s. Subscriber not provisioned in HSS.", imsi)
				}

				return issue
			},
		},

		// GTP: Context Not Found
		{
			ID:          "GTP_CONTEXT_NOT_FOUND",
			Name:        "GTP Context Not Found",
			Description: "High rate of 'Context Not Found' errors",
			Protocol:    decoder.ProtocolGTP,
			Severity:    "major",
			Category:    "protocol_error",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				if msg.Protocol != decoder.ProtocolGTP {
					return false
				}
				if cause, ok := msg.Attributes["cause"].(int); ok {
					return cause == 64
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				errRef, _ := kb.GetErrorCode("GTP", 64)
				issue := &IssueDetected{
					ID:          fmt.Sprintf("GTP_64_%d", time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "major",
					Category:    "protocol_error",
					Protocol:    "GTP",
					Title:       "GTP Session Context Not Found",
					Description: "Receiving node cannot find the requested GTP context. This may indicate session synchronization issues.",
					ErrorCode:   64,
				}

				if errRef != nil {
					issue.RootCause = errRef.Causes
					issue.Recommendations = strings.Split(errRef.Solutions, ". ")
					issue.StandardRef = errRef.StandardRef
				}

				return issue
			},
		},

		// High Error Rate Detection
		{
			ID:          "HIGH_ERROR_RATE",
			Name:        "High Error Rate Detected",
			Description: "Error rate exceeds threshold",
			Protocol:    "", // Any protocol
			Severity:    "major",
			Category:    "abnormal_pattern",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				// Calculate current error rate
				stats.mu.RLock()
				defer stats.mu.RUnlock()

				if successRate, ok := stats.SuccessRate[msg.Protocol]; ok {
					return successRate < 95.0 // Alert if below 95%
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				return &IssueDetected{
					ID:          fmt.Sprintf("HIGH_ERR_RATE_%s_%d", msg.Protocol, time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "major",
					Category:    "abnormal_pattern",
					Protocol:    msg.Protocol,
					Title:       fmt.Sprintf("High Error Rate on %s", msg.Protocol),
					Description: fmt.Sprintf("Success rate for %s protocol has dropped below 95%%. This indicates a widespread issue.", msg.Protocol),
					RootCause:   "Multiple failures detected. Possible causes: network congestion, configuration issue, backend system overload, recent configuration change.",
					Recommendations: []string{
						"Review recent configuration changes",
						"Check backend system (HSS/PGW/SMF) health and logs",
						"Analyze error code distribution to identify specific failures",
						"Monitor resource utilization (CPU, memory, disk)",
						"Verify network connectivity and routing",
					},
				}
			},
		},

		// Roaming Not Allowed
		{
			ID:          "ROAMING_NOT_ALLOWED",
			Name:        "Roaming Rejection",
			Description: "Subscriber roaming is blocked",
			Protocol:    decoder.ProtocolDiameter,
			Severity:    "major",
			Category:    "protocol_error",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				if msg.Protocol != decoder.ProtocolDiameter {
					return false
				}
				if resultCode, ok := msg.Attributes["result_code"].(int); ok {
					return resultCode == 5004
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				errRef, _ := kb.GetErrorCode("Diameter", 5004)
				issue := &IssueDetected{
					ID:          fmt.Sprintf("ROAM_BLOCKED_%d", time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "major",
					Category:    "protocol_error",
					Protocol:    "Diameter",
					Title:       "Roaming Not Allowed",
					Description: "HSS rejected roaming attempt. Subscriber is not permitted to roam in visited network.",
					ErrorCode:   5004,
				}

				if errRef != nil {
					issue.RootCause = errRef.Causes
					issue.Recommendations = strings.Split(errRef.Solutions, ". ")
					issue.StandardRef = errRef.StandardRef
				}

				if imsi, ok := msg.Attributes["imsi"].(string); ok {
					issue.AffectedIMSI = imsi
				}
				if vplmn, ok := msg.Attributes["visited_plmn"].(string); ok {
					issue.Metadata = map[string]interface{}{
						"visited_plmn": vplmn,
					}
				}

				return issue
			},
		},

		// No Resources Available
		{
			ID:          "NO_RESOURCES",
			Name:        "Resource Exhaustion",
			Description: "Node reports insufficient resources",
			Protocol:    decoder.ProtocolGTP,
			Severity:    "critical",
			Category:    "performance",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				if msg.Protocol != decoder.ProtocolGTP {
					return false
				}
				if cause, ok := msg.Attributes["cause"].(int); ok {
					return cause == 91
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				errRef, _ := kb.GetErrorCode("GTP", 91)
				issue := &IssueDetected{
					ID:          fmt.Sprintf("NO_RES_%d", time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "critical",
					Category:    "performance",
					Protocol:    "GTP",
					Title:       "Network Node Resource Exhaustion",
					Description: "SGW/PGW reports insufficient resources. This is a critical issue affecting new session creation.",
					ErrorCode:   91,
				}

				if errRef != nil {
					issue.RootCause = errRef.Causes
					issue.Recommendations = strings.Split(errRef.Solutions, ". ")
					issue.StandardRef = errRef.StandardRef
				}

				return issue
			},
		},

		// Missing APN
		{
			ID:          "MISSING_APN",
			Name:        "Missing or Unknown APN",
			Description: "PGW does not recognize the APN",
			Protocol:    decoder.ProtocolGTP,
			Severity:    "major",
			Category:    "config_issue",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				if msg.Protocol != decoder.ProtocolGTP {
					return false
				}
				if cause, ok := msg.Attributes["cause"].(int); ok {
					return cause == 67
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				errRef, _ := kb.GetErrorCode("GTP", 67)
				issue := &IssueDetected{
					ID:          fmt.Sprintf("MISS_APN_%d", time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "major",
					Category:    "config_issue",
					Protocol:    "GTP",
					Title:       "APN Configuration Issue",
					Description: "PGW rejected Create Session Request due to missing or unknown APN.",
					ErrorCode:   67,
				}

				if errRef != nil {
					issue.RootCause = errRef.Causes
					issue.Recommendations = strings.Split(errRef.Solutions, ". ")
					issue.StandardRef = errRef.StandardRef
				}

				if apn, ok := msg.Attributes["apn"].(string); ok {
					issue.Metadata = map[string]interface{}{
						"apn": apn,
					}
					issue.Description = fmt.Sprintf("PGW does not recognize APN '%s'. APN may not be configured in PGW.", apn)
				}

				return issue
			},
		},

		// High Latency Detection
		{
			ID:          "HIGH_LATENCY",
			Name:        "High Procedure Latency",
			Description: "Procedure taking longer than expected",
			Protocol:    "", // Any protocol
			Severity:    "warning",
			Category:    "performance",
			Condition: func(msg *decoder.Message, stats *Statistics) bool {
				// Check if latency exceeds baseline
				stats.mu.RLock()
				defer stats.mu.RUnlock()

				if latency, ok := msg.Attributes["latency_ms"].(float64); ok {
					procedure := msg.Attributes["procedure"].(string)
					if baseline, exists := stats.AvgLatency[procedure]; exists {
						return latency > baseline*2.0 // Alert if 2x normal
					}
				}
				return false
			},
			Action: func(msg *decoder.Message, kb *knowledge.KnowledgeBase) *IssueDetected {
				procedure := msg.Attributes["procedure"].(string)
				latency := msg.Attributes["latency_ms"].(float64)

				return &IssueDetected{
					ID:          fmt.Sprintf("HIGH_LAT_%s_%d", procedure, time.Now().Unix()),
					Timestamp:   time.Now(),
					Severity:    "warning",
					Category:    "performance",
					Protocol:    msg.Protocol,
					Title:       fmt.Sprintf("High Latency for %s", procedure),
					Description: fmt.Sprintf("Procedure %s took %.2f ms, which is significantly higher than baseline.", procedure, latency),
					RootCause:   "Possible causes: network congestion, backend system slow response, database query performance, overload, increased processing time.",
					Recommendations: []string{
						"Check network latency between nodes",
						"Review backend system performance",
						"Analyze database query performance",
						"Check for resource contention (CPU, memory, I/O)",
						"Review recent configuration or code changes",
					},
				}
			},
		},
	}
}

// AnalyzeMessage analyzes a decoded message
func (ae *AnalysisEngine) AnalyzeMessage(msg *decoder.Message) []*IssueDetected {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Update statistics
	ae.updateStatistics(msg)

	// Run detection rules
	var issues []*IssueDetected
	for _, rule := range ae.rules {
		if rule.Condition(msg, ae.statistics) {
			issue := rule.Action(msg, ae.kb)
			if issue != nil {
				issues = append(issues, issue)
				ae.issueHistory[issue.ID] = issue
				ae.logger.Warn().
					Str("issue_id", issue.ID).
					Str("severity", issue.Severity).
					Str("title", issue.Title).
					Msg("Issue detected")
			}
		}
	}

	// Check for anomalies
	anomalies := ae.anomalyDetector.Detect(msg, ae.statistics)
	issues = append(issues, anomalies...)

	return issues
}

// updateStatistics updates internal statistics
func (ae *AnalysisEngine) updateStatistics(msg *decoder.Message) {
	ae.statistics.mu.Lock()
	defer ae.statistics.mu.Unlock()

	ae.statistics.TotalMessages++

	// Track by protocol
	if msg.Result == "error" || msg.Result == "failure" {
		ae.statistics.ErrorsByProtocol[msg.Protocol]++

		// Track by error code
		if ae.statistics.ErrorsByCode[msg.Protocol] == nil {
			ae.statistics.ErrorsByCode[msg.Protocol] = make(map[int]int64)
		}
		if errorCode, ok := msg.Attributes["error_code"].(int); ok {
			ae.statistics.ErrorsByCode[msg.Protocol][errorCode]++
		}
	}

	// Update success rate
	procedure := fmt.Sprintf("%s_%s", msg.Protocol, msg.Type)
	ae.statistics.ProcedureCounts[procedure]++
	if msg.Result == "error" || msg.Result == "failure" {
		ae.statistics.ProcedureFailures[procedure]++
	}

	successCount := ae.statistics.ProcedureCounts[procedure] - ae.statistics.ProcedureFailures[procedure]
	ae.statistics.SuccessRate[msg.Protocol] = float64(successCount) / float64(ae.statistics.ProcedureCounts[procedure]) * 100.0
}

// GetStatistics returns current statistics
func (ae *AnalysisEngine) GetStatistics() *Statistics {
	ae.statistics.mu.RLock()
	defer ae.statistics.mu.RUnlock()

	// Return copy
	stats := &Statistics{
		TotalMessages:     ae.statistics.TotalMessages,
		ErrorsByProtocol:  make(map[string]int64),
		ErrorsByCode:      make(map[string]map[int]int64),
		SuccessRate:       make(map[string]float64),
		AvgLatency:        make(map[string]float64),
		TimeoutCount:      ae.statistics.TimeoutCount,
		ProcedureCounts:   make(map[string]int64),
		ProcedureFailures: make(map[string]int64),
	}

	for k, v := range ae.statistics.ErrorsByProtocol {
		stats.ErrorsByProtocol[k] = v
	}
	for k, v := range ae.statistics.SuccessRate {
		stats.SuccessRate[k] = v
	}

	return stats
}

// GetRecentIssues returns recent detected issues
func (ae *AnalysisEngine) GetRecentIssues(limit int) []*IssueDetected {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	issues := make([]*IssueDetected, 0, len(ae.issueHistory))
	for _, issue := range ae.issueHistory {
		issues = append(issues, issue)
	}

	// Sort by timestamp (most recent first)
	// In production, use proper sorting

	if len(issues) > limit {
		issues = issues[:limit]
	}

	return issues
}

// Detect anomalies
func (ad *AnomalyDetector) Detect(msg *decoder.Message, stats *Statistics) []*IssueDetected {
	// Implement anomaly detection logic
	// For now, return empty
	return nil
}
