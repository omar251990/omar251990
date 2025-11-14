package correlation

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// SubscriberProfile aggregates all information about a subscriber
type SubscriberProfile struct {
	mu                sync.RWMutex
	IMSI              string                   `json:"imsi"`
	MSISDN            string                   `json:"msisdn"`
	IMEI              string                   `json:"imei"`
	SUPI              string                   `json:"supi"`              // 5G
	CurrentLocation   *LocationInfo            `json:"current_location"`
	LocationHistory   []*LocationInfo          `json:"location_history"`
	ActiveSessions    []*SessionInfo           `json:"active_sessions"`
	SessionHistory    []*SessionInfo           `json:"session_history"`
	Procedures        []*ProcedureInstance     `json:"procedures"`
	Errors            []*ErrorOccurrence       `json:"errors"`
	Timeline          []*TimelineEvent         `json:"timeline"`
	Statistics        *SubscriberStatistics    `json:"statistics"`
	DeviceInfo        *DeviceInfo              `json:"device_info"`
	SubscriptionData  *SubscriptionData        `json:"subscription_data"`
	FirstSeen         time.Time                `json:"first_seen"`
	LastSeen          time.Time                `json:"last_seen"`
	Status            string                   `json:"status"`            // "active", "idle", "detached"
}

// LocationInfo tracks subscriber location
type LocationInfo struct {
	Timestamp     time.Time `json:"timestamp"`
	Generation    string    `json:"generation"`    // "2G", "3G", "4G", "5G"
	RAT           string    `json:"rat"`           // Radio Access Technology
	MCC           string    `json:"mcc"`           // Mobile Country Code
	MNC           string    `json:"mnc"`           // Mobile Network Code
	LAC           string    `json:"lac,omitempty"` // Location Area Code (2G/3G)
	TAC           string    `json:"tac,omitempty"` // Tracking Area Code (4G/5G)
	CellID        string    `json:"cell_id"`
	eNBID         string    `json:"enb_id,omitempty"`  // 4G
	gNBID         string    `json:"gnb_id,omitempty"`  // 5G
	Latitude      float64   `json:"latitude,omitempty"`
	Longitude     float64   `json:"longitude,omitempty"`
}

// SessionInfo tracks data sessions
type SessionInfo struct {
	SessionID     string                 `json:"session_id"`
	Type          string                 `json:"type"`          // "PDN", "PDU", "PDP"
	Protocol      string                 `json:"protocol"`      // "GTP", "PFCP"
	APN           string                 `json:"apn,omitempty"` // 4G
	DNN           string                 `json:"dnn,omitempty"` // 5G
	IPAddress     string                 `json:"ip_address"`
	IPv6Address   string                 `json:"ipv6_address,omitempty"`
	QoS           *QoSInfo               `json:"qos"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	BytesUplink   uint64                 `json:"bytes_uplink"`
	BytesDownlink uint64                 `json:"bytes_downlink"`
	Status        string                 `json:"status"`        // "active", "terminated"
	TermCause     string                 `json:"term_cause,omitempty"`
	Interfaces    map[string]string      `json:"interfaces"`    // Interface -> Node address
}

// QoSInfo quality of service information
type QoSInfo struct {
	QCI           int    `json:"qci,omitempty"`           // 4G
	5QI           int    `json:"5qi,omitempty"`           // 5G
	ARP           int    `json:"arp"`                     // Allocation and Retention Priority
	MBRUplink     uint64 `json:"mbr_uplink,omitempty"`    // Maximum Bit Rate
	MBRDownlink   uint64 `json:"mbr_downlink,omitempty"`
	GBRUplink     uint64 `json:"gbr_uplink,omitempty"`    // Guaranteed Bit Rate
	GBRDownlink   uint64 `json:"gbr_downlink,omitempty"`
}

// ProcedureInstance tracks individual procedure execution
type ProcedureInstance struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`          // "Attach", "TAU", "Handover"
	Protocol      string        `json:"protocol"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	Result        string        `json:"result"`        // "success", "failure", "ongoing"
	Cause         string        `json:"cause,omitempty"`
	MessageCount  int           `json:"message_count"`
	Location      *LocationInfo `json:"location"`
}

// ErrorOccurrence tracks errors for a subscriber
type ErrorOccurrence struct {
	Timestamp   time.Time              `json:"timestamp"`
	Protocol    string                 `json:"protocol"`
	Procedure   string                 `json:"procedure"`
	ErrorCode   int                    `json:"error_code"`
	ErrorName   string                 `json:"error_name"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Resolved    bool                   `json:"resolved"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TimelineEvent represents an event in subscriber timeline
type TimelineEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`        // "attach", "detach", "session_create", "handover", "error"
	Description string                 `json:"description"`
	Protocol    string                 `json:"protocol"`
	Location    *LocationInfo          `json:"location,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Icon        string                 `json:"icon"`        // For UI display
	Color       string                 `json:"color"`       // For UI display
}

// SubscriberStatistics aggregated statistics
type SubscriberStatistics struct {
	TotalSessions        int           `json:"total_sessions"`
	TotalProcedures      int           `json:"total_procedures"`
	SuccessfulProcedures int           `json:"successful_procedures"`
	FailedProcedures     int           `json:"failed_procedures"`
	SuccessRate          float64       `json:"success_rate"`
	TotalErrors          int           `json:"total_errors"`
	TotalDataUplink      uint64        `json:"total_data_uplink"`
	TotalDataDownlink    uint64        `json:"total_data_downlink"`
	AvgSessionDuration   time.Duration `json:"avg_session_duration"`
	AvgProcedureDuration time.Duration `json:"avg_procedure_duration"`
	LocationChanges      int           `json:"location_changes"`
	HandoverCount        int           `json:"handover_count"`
}

// DeviceInfo information about subscriber's device
type DeviceInfo struct {
	IMEI         string   `json:"imei"`
	TAC          string   `json:"tac"`           // Type Allocation Code (first 8 digits of IMEI)
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`  // LTE Cat, 5G bands, etc.
}

// SubscriptionData subscriber's plan and permissions
type SubscriptionData struct {
	SubscriberType    string   `json:"subscriber_type"`    // "prepaid", "postpaid"
	ServiceLevel      string   `json:"service_level"`      // "basic", "premium", "VIP"
	AllowedRATs       []string `json:"allowed_rats"`       // Allowed Radio Access Technologies
	AllowedAPNs       []string `json:"allowed_apns"`
	RoamingAllowed    bool     `json:"roaming_allowed"`
	MaxDataQuota      uint64   `json:"max_data_quota"`
	UsedDataQuota     uint64   `json:"used_data_quota"`
	VoiceEnabled      bool     `json:"voice_enabled"`
	SMSEnabled        bool     `json:"sms_enabled"`
	DataEnabled       bool     `json:"data_enabled"`
}

// SubscriberCorrelator correlates subscriber activity across interfaces
type SubscriberCorrelator struct {
	mu          sync.RWMutex
	subscribers map[string]*SubscriberProfile // Key: IMSI
	imsiToMSISDN map[string]string            // MSISDN -> IMSI mapping
	imeiToIMSI  map[string]string             // IMEI -> IMSI mapping
	teidToIMSI  map[string]string             // TEID -> IMSI mapping (GTP)
	seidToIMSI  map[string]string             // SEID -> IMSI mapping (PFCP)
}

// NewSubscriberCorrelator creates a new subscriber correlator
func NewSubscriberCorrelator() *SubscriberCorrelator {
	return &SubscriberCorrelator{
		subscribers:  make(map[string]*SubscriberProfile),
		imsiToMSISDN: make(map[string]string),
		imeiToIMSI:   make(map[string]string),
		teidToIMSI:   make(map[string]string),
		seidToIMSI:   make(map[string]string),
	}
}

// ProcessMessage processes a message and updates subscriber profile
func (sc *SubscriberCorrelator) ProcessMessage(msg *decoder.Message) {
	// Extract identifiers
	imsi, _ := msg.Attributes["imsi"].(string)
	msisdn, _ := msg.Attributes["msisdn"].(string)
	imei, _ := msg.Attributes["imei"].(string)
	teid, _ := msg.Attributes["teid"].(string)
	seid, _ := msg.Attributes["seid"].(string)

	if imsi == "" {
		// Try to resolve IMSI from other identifiers
		if msisdn != "" {
			sc.mu.RLock()
			imsi = sc.imsiToMSISDN[msisdn]
			sc.mu.RUnlock()
		} else if imei != "" {
			sc.mu.RLock()
			imsi = sc.imeiToIMSI[imei]
			sc.mu.RUnlock()
		} else if teid != "" {
			sc.mu.RLock()
			imsi = sc.teidToIMSI[teid]
			sc.mu.RUnlock()
		} else if seid != "" {
			sc.mu.RLock()
			imsi = sc.seidToIMSI[seid]
			sc.mu.RUnlock()
		}
	}

	if imsi == "" {
		return // Cannot correlate without IMSI
	}

	// Get or create subscriber profile
	profile := sc.getOrCreateProfile(imsi)

	profile.mu.Lock()
	defer profile.mu.Unlock()

	// Update identifiers
	if msisdn != "" && profile.MSISDN == "" {
		profile.MSISDN = msisdn
		sc.mu.Lock()
		sc.imsiToMSISDN[msisdn] = imsi
		sc.mu.Unlock()
	}
	if imei != "" && profile.IMEI == "" {
		profile.IMEI = imei
		sc.mu.Lock()
		sc.imeiToIMSI[imei] = imsi
		sc.mu.Unlock()

		// Update device info
		profile.DeviceInfo = &DeviceInfo{
			IMEI: imei,
			TAC:  imei[:8], // First 8 digits
			// Lookup manufacturer/model from TAC database (placeholder)
			Manufacturer: "Unknown",
			Model:        "Unknown",
		}
	}

	// Update location
	if mcc, ok := msg.Attributes["mcc"].(string); ok {
		location := &LocationInfo{
			Timestamp:  msg.Timestamp,
			Generation: sc.detectGeneration(msg.Protocol),
			RAT:        msg.Protocol,
			MCC:        mcc,
		}

		if mnc, ok := msg.Attributes["mnc"].(string); ok {
			location.MNC = mnc
		}
		if cellID, ok := msg.Attributes["cell_id"].(string); ok {
			location.CellID = cellID
		}
		if tac, ok := msg.Attributes["tac"].(string); ok {
			location.TAC = tac
		}

		profile.CurrentLocation = location
		profile.LocationHistory = append(profile.LocationHistory, location)
		if len(profile.LocationHistory) > 100 {
			profile.LocationHistory = profile.LocationHistory[1:]
		}
	}

	// Add timeline event
	event := &TimelineEvent{
		Timestamp:   msg.Timestamp,
		Type:        sc.classifyEvent(msg),
		Description: fmt.Sprintf("%s: %s", msg.Protocol, msg.Type),
		Protocol:    msg.Protocol,
		Icon:        sc.getEventIcon(msg),
		Color:       sc.getEventColor(msg),
	}
	profile.Timeline = append(profile.Timeline, event)

	// Track errors
	if msg.Result == "error" || msg.Result == "failure" {
		errorCode, _ := msg.Attributes["error_code"].(int)
		profile.Errors = append(profile.Errors, &ErrorOccurrence{
			Timestamp:   msg.Timestamp,
			Protocol:    msg.Protocol,
			Procedure:   msg.Type,
			ErrorCode:   errorCode,
			ErrorName:   msg.Attributes["error_name"].(string),
			Description: fmt.Sprintf("Error in %s procedure", msg.Type),
			Severity:    "major",
			Resolved:    false,
		})
		profile.Statistics.TotalErrors++
	}

	// Update statistics
	profile.Statistics.TotalProcedures++
	if msg.Result == "success" {
		profile.Statistics.SuccessfulProcedures++
	} else if msg.Result == "error" || msg.Result == "failure" {
		profile.Statistics.FailedProcedures++
	}

	if profile.Statistics.TotalProcedures > 0 {
		profile.Statistics.SuccessRate = float64(profile.Statistics.SuccessfulProcedures) / float64(profile.Statistics.TotalProcedures) * 100.0
	}

	profile.LastSeen = msg.Timestamp
	if profile.FirstSeen.IsZero() {
		profile.FirstSeen = msg.Timestamp
	}

	profile.Status = sc.determineStatus(msg)
}

// getOrCreateProfile gets or creates a subscriber profile
func (sc *SubscriberCorrelator) getOrCreateProfile(imsi string) *SubscriberProfile {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	profile, exists := sc.subscribers[imsi]
	if !exists {
		profile = &SubscriberProfile{
			IMSI:             imsi,
			LocationHistory:  make([]*LocationInfo, 0),
			ActiveSessions:   make([]*SessionInfo, 0),
			SessionHistory:   make([]*SessionInfo, 0),
			Procedures:       make([]*ProcedureInstance, 0),
			Errors:           make([]*ErrorOccurrence, 0),
			Timeline:         make([]*TimelineEvent, 0),
			Statistics:       &SubscriberStatistics{},
			FirstSeen:        time.Now(),
			Status:           "active",
		}
		sc.subscribers[imsi] = profile
	}

	return profile
}

// GetProfile returns subscriber profile by IMSI
func (sc *SubscriberCorrelator) GetProfile(imsi string) *SubscriberProfile {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.subscribers[imsi]
}

// GetProfileByMSISDN returns subscriber profile by MSISDN
func (sc *SubscriberCorrelator) GetProfileByMSISDN(msisdn string) *SubscriberProfile {
	sc.mu.RLock()
	imsi := sc.imsiToMSISDN[msisdn]
	sc.mu.RUnlock()

	if imsi == "" {
		return nil
	}

	return sc.GetProfile(imsi)
}

// GetTimeline returns subscriber timeline with filtering
func (sc *SubscriberCorrelator) GetTimeline(imsi string, startTime, endTime time.Time, eventTypes []string) []*TimelineEvent {
	profile := sc.GetProfile(imsi)
	if profile == nil {
		return nil
	}

	profile.mu.RLock()
	defer profile.mu.RUnlock()

	filtered := make([]*TimelineEvent, 0)
	for _, event := range profile.Timeline {
		// Time filter
		if !startTime.IsZero() && event.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && event.Timestamp.After(endTime) {
			continue
		}

		// Type filter
		if len(eventTypes) > 0 {
			found := false
			for _, t := range eventTypes {
				if event.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, event)
	}

	return filtered
}

// detectGeneration detects network generation from protocol
func (sc *SubscriberCorrelator) detectGeneration(protocol string) string {
	switch protocol {
	case "MAP", "CAP", "INAP":
		return "2G/3G"
	case "S1AP", "NAS", decoder.ProtocolGTP:
		return "4G"
	case "NGAP", "PFCP", "HTTP":
		return "5G"
	default:
		return "Unknown"
	}
}

// classifyEvent classifies message into event type
func (sc *SubscriberCorrelator) classifyEvent(msg *decoder.Message) string {
	msgType := strings.ToLower(msg.Type)

	if strings.Contains(msgType, "attach") {
		return "attach"
	} else if strings.Contains(msgType, "detach") {
		return "detach"
	} else if strings.Contains(msgType, "registration") {
		return "registration"
	} else if strings.Contains(msgType, "session") && strings.Contains(msgType, "create") {
		return "session_create"
	} else if strings.Contains(msgType, "session") && strings.Contains(msgType, "delete") {
		return "session_delete"
	} else if strings.Contains(msgType, "handover") {
		return "handover"
	} else if msg.Result == "error" || msg.Result == "failure" {
		return "error"
	} else if strings.Contains(msgType, "authentication") {
		return "authentication"
	} else if strings.Contains(msgType, "location") {
		return "location_update"
	}

	return "other"
}

// getEventIcon returns icon for event type
func (sc *SubscriberCorrelator) getEventIcon(msg *decoder.Message) string {
	eventType := sc.classifyEvent(msg)

	icons := map[string]string{
		"attach":          "📱",
		"detach":          "🚪",
		"registration":    "✅",
		"session_create":  "🔌",
		"session_delete":  "🔕",
		"handover":        "🔄",
		"error":           "⚠️",
		"authentication":  "🔐",
		"location_update": "📍",
		"other":           "📋",
	}

	return icons[eventType]
}

// getEventColor returns color for event type
func (sc *SubscriberCorrelator) getEventColor(msg *decoder.Message) string {
	eventType := sc.classifyEvent(msg)

	colors := map[string]string{
		"attach":          "#10b981",
		"detach":          "#6b7280",
		"registration":    "#3b82f6",
		"session_create":  "#8b5cf6",
		"session_delete":  "#f59e0b",
		"handover":        "#06b6d4",
		"error":           "#ef4444",
		"authentication":  "#ec4899",
		"location_update": "#14b8a6",
		"other":           "#9ca3af",
	}

	return colors[eventType]
}

// determineStatus determines subscriber status from message
func (sc *SubscriberCorrelator) determineStatus(msg *decoder.Message) string {
	msgType := strings.ToLower(msg.Type)

	if strings.Contains(msgType, "attach") || strings.Contains(msgType, "registration") {
		return "active"
	} else if strings.Contains(msgType, "detach") || strings.Contains(msgType, "deregistration") {
		return "detached"
	}

	return "active" // Default
}

// GetAllSubscribers returns all subscriber profiles
func (sc *SubscriberCorrelator) GetAllSubscribers() []*SubscriberProfile {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	profiles := make([]*SubscriberProfile, 0, len(sc.subscribers))
	for _, profile := range sc.subscribers {
		profiles = append(profiles, profile)
	}

	return profiles
}

// GetActiveSubscribers returns currently active subscribers
func (sc *SubscriberCorrelator) GetActiveSubscribers() []*SubscriberProfile {
	all := sc.GetAllSubscribers()
	active := make([]*SubscriberProfile, 0)

	for _, profile := range all {
		profile.mu.RLock()
		if profile.Status == "active" && time.Since(profile.LastSeen) < 5*time.Minute {
			active = append(active, profile)
		}
		profile.mu.RUnlock()
	}

	return active
}
