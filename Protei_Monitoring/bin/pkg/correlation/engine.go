package correlation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// Engine handles message correlation and session tracking
type Engine struct {
	cache      *TIDCache
	config     *Config
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
}

// Config holds correlation engine configuration
type Config struct {
	CacheSize         int
	TIDTTL            time.Duration
	SessionTimeout    time.Duration
	CorrelationFields []string
	E2ETracking       bool
}

// Session represents a correlated transaction/session
type Session struct {
	TID           string
	IMSI          string
	MSISDN        string
	SUPI          string
	SessionID     string
	PLMN          string
	CellID        string
	APN           string
	DNN           string
	Procedure     string
	StartTime     time.Time
	LastActivity  time.Time
	Messages      []*decoder.Message
	NetworkPath   []string
	Result        decoder.Result
	Duration      time.Duration
	MessageCount  int
	FailureCause  int
	FailureText   string
	mu            sync.Mutex
}

// TIDCache is a cache for Transaction IDs
type TIDCache struct {
	cache map[string]*CacheEntry
	ttl   time.Duration
	mu    sync.RWMutex
}

// CacheEntry represents a cached TID
type CacheEntry struct {
	TID       string
	Keys      map[string]string
	CreatedAt time.Time
	LastSeen  time.Time
}

// NewEngine creates a new correlation engine
func NewEngine(config *Config) *Engine {
	return &Engine{
		cache:    NewTIDCache(config.CacheSize, config.TIDTTL),
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// NewTIDCache creates a new TID cache
func NewTIDCache(size int, ttl time.Duration) *TIDCache {
	cache := &TIDCache{
		cache: make(map[string]*CacheEntry, size),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Correlate processes a message and assigns it to a session
func (e *Engine) Correlate(msg *decoder.Message) (*Session, error) {
	// Generate or retrieve TID
	tid := e.getTID(msg)
	if tid == "" {
		// If no correlation possible, create a unique TID
		tid = e.generateUniqueTID(msg)
	}

	msg.TransactionID = tid

	// Get or create session
	e.sessionsMu.Lock()
	session, exists := e.sessions[tid]
	if !exists {
		session = e.createSession(tid, msg)
		e.sessions[tid] = session
	}
	e.sessionsMu.Unlock()

	// Add message to session
	session.mu.Lock()
	session.Messages = append(session.Messages, msg)
	session.MessageCount++
	session.LastActivity = msg.Timestamp

	// Update session metadata
	e.updateSessionMetadata(session, msg)

	// Update network path
	e.updateNetworkPath(session, msg)

	// Check if session is complete
	if e.isSessionComplete(session) {
		session.Duration = session.LastActivity.Sub(session.StartTime)
		session.Result = e.determineSessionResult(session)
	}

	session.mu.Unlock()

	return session, nil
}

// getTID retrieves or generates a Transaction ID for a message
func (e *Engine) getTID(msg *decoder.Message) string {
	// Build correlation key from message fields
	key := e.buildCorrelationKey(msg)
	if key == "" {
		return ""
	}

	// Check cache
	entry := e.cache.Get(key)
	if entry != nil {
		entry.LastSeen = time.Now()
		return entry.TID
	}

	// Generate new TID
	tid := e.generateTID(msg)

	// Store in cache
	e.cache.Put(key, tid, e.extractCorrelationFields(msg))

	return tid
}

// buildCorrelationKey builds a cache key from correlation fields
func (e *Engine) buildCorrelationKey(msg *decoder.Message) string {
	keys := []string{}

	for _, field := range e.config.CorrelationFields {
		var value string
		switch field {
		case "imsi":
			value = msg.IMSI
		case "msisdn":
			value = msg.MSISDN
		case "supi":
			value = msg.SUPI
		case "teid":
			if msg.TEID != 0 {
				value = fmt.Sprintf("%d", msg.TEID)
			}
		case "seid":
			if msg.SEID != 0 {
				value = fmt.Sprintf("%d", msg.SEID)
			}
		case "session_id":
			value = msg.SessionID
		case "apn":
			value = msg.APN
		case "dnn":
			value = msg.DNN
		case "plmn":
			value = msg.PLMN
		case "cell_id":
			value = msg.CellID
		}

		if value != "" {
			keys = append(keys, value)
		}
	}

	if len(keys) == 0 {
		return ""
	}

	// Concatenate keys
	keyStr := ""
	for _, k := range keys {
		keyStr += k + "|"
	}

	return keyStr
}

// generateTID generates a unique Transaction ID
func (e *Engine) generateTID(msg *decoder.Message) string {
	// Use primary identifiers for TID
	tid := ""

	if msg.IMSI != "" {
		tid = "imsi_" + msg.IMSI
	} else if msg.SUPI != "" {
		tid = "supi_" + msg.SUPI
	} else if msg.MSISDN != "" {
		tid = "msisdn_" + msg.MSISDN
	} else if msg.SessionID != "" {
		tid = "session_" + msg.SessionID
	} else if msg.TEID != 0 {
		tid = fmt.Sprintf("teid_%d", msg.TEID)
	}

	// Add timestamp and procedure for uniqueness
	if tid != "" {
		procedure := e.detectProcedure(msg)
		tid = fmt.Sprintf("%s_%s_%d", tid, procedure, msg.Timestamp.Unix())
	}

	return tid
}

// generateUniqueTID creates a unique TID when no correlation is possible
func (e *Engine) generateUniqueTID(msg *decoder.Message) string {
	data := fmt.Sprintf("%s_%s_%s_%d_%d",
		msg.Protocol,
		msg.MessageName,
		msg.Source.IP,
		msg.Timestamp.UnixNano(),
		msg.SequenceNum,
	)

	hash := sha256.Sum256([]byte(data))
	return "tid_" + hex.EncodeToString(hash[:8])
}

// extractCorrelationFields extracts key fields from message
func (e *Engine) extractCorrelationFields(msg *decoder.Message) map[string]string {
	fields := make(map[string]string)

	if msg.IMSI != "" {
		fields["imsi"] = msg.IMSI
	}
	if msg.MSISDN != "" {
		fields["msisdn"] = msg.MSISDN
	}
	if msg.SUPI != "" {
		fields["supi"] = msg.SUPI
	}
	if msg.SessionID != "" {
		fields["session_id"] = msg.SessionID
	}
	if msg.APN != "" {
		fields["apn"] = msg.APN
	}
	if msg.DNN != "" {
		fields["dnn"] = msg.DNN
	}

	return fields
}

// createSession creates a new session
func (e *Engine) createSession(tid string, msg *decoder.Message) *Session {
	return &Session{
		TID:          tid,
		IMSI:         msg.IMSI,
		MSISDN:       msg.MSISDN,
		SUPI:         msg.SUPI,
		SessionID:    msg.SessionID,
		PLMN:         msg.PLMN,
		CellID:       msg.CellID,
		APN:          msg.APN,
		DNN:          msg.DNN,
		Procedure:    e.detectProcedure(msg),
		StartTime:    msg.Timestamp,
		LastActivity: msg.Timestamp,
		Messages:     make([]*decoder.Message, 0, 10),
		NetworkPath:  make([]string, 0, 5),
		Result:       decoder.ResultUnknown,
	}
}

// updateSessionMetadata updates session metadata from message
func (e *Engine) updateSessionMetadata(session *Session, msg *decoder.Message) {
	if session.IMSI == "" && msg.IMSI != "" {
		session.IMSI = msg.IMSI
	}
	if session.MSISDN == "" && msg.MSISDN != "" {
		session.MSISDN = msg.MSISDN
	}
	if session.SUPI == "" && msg.SUPI != "" {
		session.SUPI = msg.SUPI
	}
	if session.PLMN == "" && msg.PLMN != "" {
		session.PLMN = msg.PLMN
	}
	if session.CellID == "" && msg.CellID != "" {
		session.CellID = msg.CellID
	}
	if session.APN == "" && msg.APN != "" {
		session.APN = msg.APN
	}
	if session.DNN == "" && msg.DNN != "" {
		session.DNN = msg.DNN
	}

	// Update failure information
	if msg.Result == decoder.ResultFailure {
		session.FailureCause = msg.CauseCode
		session.FailureText = msg.CauseText
	}
}

// updateNetworkPath tracks the path of network elements
func (e *Engine) updateNetworkPath(session *Session, msg *decoder.Message) {
	// Add source and destination to path if not already present
	srcNode := fmt.Sprintf("%s(%s)", msg.Source.Type, msg.Source.IP)
	dstNode := fmt.Sprintf("%s(%s)", msg.Destination.Type, msg.Destination.IP)

	if !contains(session.NetworkPath, srcNode) {
		session.NetworkPath = append(session.NetworkPath, srcNode)
	}
	if !contains(session.NetworkPath, dstNode) {
		session.NetworkPath = append(session.NetworkPath, dstNode)
	}
}

// detectProcedure detects the procedure type from the message
func (e *Engine) detectProcedure(msg *decoder.Message) string {
	// Map message types to procedures
	procedureMap := map[string]string{
		"UpdateLocation":               "attach_4g",
		"ULR":                          "attach_4g",
		"AIR":                          "authentication",
		"CreateSessionRequest":         "pdu_session_establishment",
		"CreatePDPContextRequest":      "pdp_activation",
		"Diameter_Registration":        "registration_5g",
		"NGAP_InitialUEMessage":        "registration_5g",
		"S1AP_InitialUEMessage":        "attach_4g",
		"NGAP_HandoverRequired":        "handover",
		"S1AP_HandoverRequired":        "handover",
	}

	if procedure, ok := procedureMap[msg.MessageName]; ok {
		return procedure
	}

	// Fallback: use protocol as procedure
	return string(msg.Protocol)
}

// isSessionComplete checks if a session is complete
func (e *Engine) isSessionComplete(session *Session) bool {
	// Simple heuristic: session is complete if we have both request and response
	hasRequest := false
	hasResponse := false

	for _, msg := range session.Messages {
		if msg.Direction == decoder.DirectionRequest {
			hasRequest = true
		}
		if msg.Direction == decoder.DirectionResponse {
			hasResponse = true
		}
	}

	return hasRequest && hasResponse
}

// determineSessionResult determines the overall result of a session
func (e *Engine) determineSessionResult(session *Session) decoder.Result {
	// If any message failed, the session failed
	for _, msg := range session.Messages {
		if msg.Result == decoder.ResultFailure {
			return decoder.ResultFailure
		}
		if msg.Result == decoder.ResultTimeout {
			return decoder.ResultTimeout
		}
	}

	// If we have a response with success, it's successful
	for _, msg := range session.Messages {
		if msg.Direction == decoder.DirectionResponse && msg.Result == decoder.ResultSuccess {
			return decoder.ResultSuccess
		}
	}

	return decoder.ResultUnknown
}

// GetSession retrieves a session by TID
func (e *Engine) GetSession(tid string) (*Session, bool) {
	e.sessionsMu.RLock()
	defer e.sessionsMu.RUnlock()
	session, ok := e.sessions[tid]
	return session, ok
}

// GetAllSessions returns all active sessions
func (e *Engine) GetAllSessions() []*Session {
	e.sessionsMu.RLock()
	defer e.sessionsMu.RUnlock()

	sessions := make([]*Session, 0, len(e.sessions))
	for _, session := range e.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CleanupSessions removes old sessions
func (e *Engine) CleanupSessions() int {
	now := time.Now()
	timeout := e.config.SessionTimeout

	e.sessionsMu.Lock()
	defer e.sessionsMu.Unlock()

	count := 0
	for tid, session := range e.sessions {
		if now.Sub(session.LastActivity) > timeout {
			delete(e.sessions, tid)
			count++
		}
	}

	return count
}

// TIDCache methods

// Get retrieves an entry from cache
func (c *TIDCache) Get(key string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[key]
}

// Put stores an entry in cache
func (c *TIDCache) Put(key, tid string, fields map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = &CacheEntry{
		TID:       tid,
		Keys:      fields,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
}

// cleanupLoop periodically removes expired entries
func (c *TIDCache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries
func (c *TIDCache) cleanup() {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.cache {
		if now.Sub(entry.LastSeen) > c.ttl {
			delete(c.cache, key)
		}
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
