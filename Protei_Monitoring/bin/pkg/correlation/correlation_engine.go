package correlation

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// IdentifierType represents different types of subscriber identifiers
type IdentifierType string

const (
	IdentifierIMSI   IdentifierType = "IMSI"
	IdentifierMSISDN IdentifierType = "MSISDN"
	IdentifierIMEI   IdentifierType = "IMEI"
	IdentifierTEID   IdentifierType = "TEID"
	IdentifierSEID   IdentifierType = "SEID"
	IdentifierIP     IdentifierType = "IP"
	IdentifierAPN    IdentifierType = "APN"
	IdentifierMME_ID IdentifierType = "MME_UE_ID"
	IdentifierAMF_ID IdentifierType = "AMF_UE_ID"
	IdentifierENB_ID IdentifierType = "ENB_UE_ID"
	IdentifierRAN_ID IdentifierType = "RAN_UE_ID"
)

// Identifier represents a single identifier with metadata
type Identifier struct {
	Type      IdentifierType
	Value     string
	Protocol  string
	FirstSeen time.Time
	LastSeen  time.Time
	Confidence float64 // 0.0 to 1.0
}

// CorrelationSession represents a correlated session across multiple interfaces
type CorrelationSession struct {
	ID               string
	Identifiers      map[IdentifierType][]Identifier
	Protocols        []string
	Transactions     []string // Transaction IDs
	StartTime        time.Time
	EndTime          time.Time
	Status           string
	SessionType      string // "voice", "data", "sms", "location_update", etc.

	// Cross-interface references
	MapTransactionID string
	DiameterSessionID string
	GtpTEID          uint32
	PfcpSEID         uint64
	NgapUE_ID        uint64
	S1apMME_ID       uint32

	// Location tracking
	LocationHistory  []LocationUpdate
	CurrentLocation  *LocationUpdate

	// Data usage
	BytesUplink      uint64
	BytesDownlink    uint64

	// Quality metrics
	SuccessRate      float64
	AvgLatency       time.Duration
	ErrorCount       int

	mutex            sync.RWMutex
}

// LocationUpdate represents a location update event
type LocationUpdate struct {
	Timestamp     time.Time
	Protocol      string
	MCC           string
	MNC           string
	LAC           string
	CellID        string
	TAC           string // 4G/5G
	EUTRAN_CGI    string // 4G
	GlobalRAN_ID  string // 5G
	Latitude      float64
	Longitude     float64
}

// CorrelationEngine manages session correlation across protocols
type CorrelationEngine struct {
	mu                  sync.RWMutex
	sessions            map[string]*CorrelationSession // Key: Session ID
	identifierIndex     map[IdentifierType]map[string]*CorrelationSession
	transactionIndex    map[string]*CorrelationSession
	db                  *sql.DB
	sessionTimeout      time.Duration
	cleanupInterval     time.Duration
	stopChan            chan struct{}
}

// NewCorrelationEngine creates a new correlation engine
func NewCorrelationEngine(db *sql.DB, sessionTimeout time.Duration) *CorrelationEngine {
	engine := &CorrelationEngine{
		sessions:         make(map[string]*CorrelationSession),
		identifierIndex:  make(map[IdentifierType]map[string]*CorrelationSession),
		transactionIndex: make(map[string]*CorrelationSession),
		db:               db,
		sessionTimeout:   sessionTimeout,
		cleanupInterval:  5 * time.Minute,
		stopChan:         make(chan struct{}),
	}

	// Initialize identifier indexes
	identifierTypes := []IdentifierType{
		IdentifierIMSI, IdentifierMSISDN, IdentifierIMEI,
		IdentifierTEID, IdentifierSEID, IdentifierIP,
		IdentifierMME_ID, IdentifierAMF_ID, IdentifierENB_ID, IdentifierRAN_ID,
	}

	for _, idType := range identifierTypes {
		engine.identifierIndex[idType] = make(map[string]*CorrelationSession)
	}

	// Start cleanup goroutine
	go engine.cleanupExpiredSessions()

	return engine
}

// CorrelateTransaction correlates a transaction with existing or new session
func (e *CorrelationEngine) CorrelateTransaction(txn *TransactionEvent) (*CorrelationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Try to find existing session by identifiers
	session := e.findSessionByIdentifiers(txn.Identifiers)

	if session == nil {
		// Create new session
		session = e.createNewSession(txn)
	} else {
		// Update existing session
		e.updateSession(session, txn)
	}

	// Index transaction
	e.transactionIndex[txn.TransactionID] = session

	// Persist to database
	go e.persistSession(session)

	return session, nil
}

// findSessionByIdentifiers finds a session matching any of the identifiers
func (e *CorrelationEngine) findSessionByIdentifiers(identifiers []Identifier) *CorrelationSession {
	// Priority order: IMSI > MSISDN > TEID > SEID > IP
	priorityOrder := []IdentifierType{
		IdentifierIMSI,
		IdentifierMSISDN,
		IdentifierTEID,
		IdentifierSEID,
		IdentifierMME_ID,
		IdentifierAMF_ID,
		IdentifierIP,
	}

	for _, idType := range priorityOrder {
		for _, id := range identifiers {
			if id.Type == idType && id.Value != "" {
				if session, exists := e.identifierIndex[idType][id.Value]; exists {
					// Check if session is still active (within timeout)
					if time.Since(session.EndTime) < e.sessionTimeout {
						return session
					}
				}
			}
		}
	}

	return nil
}

// createNewSession creates a new correlation session
func (e *CorrelationEngine) createNewSession(txn *TransactionEvent) *CorrelationSession {
	sessionID := generateSessionID(txn)

	session := &CorrelationSession{
		ID:              sessionID,
		Identifiers:     make(map[IdentifierType][]Identifier),
		Protocols:       []string{txn.Protocol},
		Transactions:    []string{txn.TransactionID},
		StartTime:       txn.Timestamp,
		EndTime:         txn.Timestamp,
		Status:          "active",
		SessionType:     txn.SessionType,
		LocationHistory: make([]LocationUpdate, 0),
	}

	// Add identifiers
	for _, id := range txn.Identifiers {
		session.addIdentifier(id)
		e.identifierIndex[id.Type][id.Value] = session
	}

	// Add protocol-specific references
	session.updateProtocolReferences(txn)

	// Add location if available
	if txn.Location != nil {
		session.addLocation(*txn.Location)
	}

	e.sessions[sessionID] = session

	return session
}

// updateSession updates an existing session with new transaction data
func (e *CorrelationEngine) updateSession(session *CorrelationSession, txn *TransactionEvent) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	// Update identifiers
	for _, id := range txn.Identifiers {
		session.addIdentifier(id)
		e.identifierIndex[id.Type][id.Value] = session
	}

	// Update protocols
	if !contains(session.Protocols, txn.Protocol) {
		session.Protocols = append(session.Protocols, txn.Protocol)
	}

	// Update transactions
	session.Transactions = append(session.Transactions, txn.TransactionID)

	// Update timestamps
	session.EndTime = txn.Timestamp

	// Update protocol-specific references
	session.updateProtocolReferences(txn)

	// Update location
	if txn.Location != nil {
		session.addLocation(*txn.Location)
	}

	// Update data usage
	session.BytesUplink += txn.BytesUplink
	session.BytesDownlink += txn.BytesDownlink

	// Update quality metrics
	if txn.Success {
		successCount := int(session.SuccessRate * float64(len(session.Transactions)))
		session.SuccessRate = float64(successCount+1) / float64(len(session.Transactions))
	} else {
		session.ErrorCount++
		successCount := int(session.SuccessRate * float64(len(session.Transactions)-1))
		session.SuccessRate = float64(successCount) / float64(len(session.Transactions))
	}

	// Update average latency
	totalLatency := session.AvgLatency * time.Duration(len(session.Transactions)-1)
	session.AvgLatency = (totalLatency + txn.Latency) / time.Duration(len(session.Transactions))
}

// addIdentifier adds an identifier to the session
func (s *CorrelationSession) addIdentifier(id Identifier) {
	// Check if identifier already exists
	for _, existingID := range s.Identifiers[id.Type] {
		if existingID.Value == id.Value {
			// Update last seen
			existingID.LastSeen = id.LastSeen
			return
		}
	}

	// Add new identifier
	s.Identifiers[id.Type] = append(s.Identifiers[id.Type], id)
}

// updateProtocolReferences updates protocol-specific cross-references
func (s *CorrelationSession) updateProtocolReferences(txn *TransactionEvent) {
	switch txn.Protocol {
	case "MAP", "CAP", "INAP":
		if s.MapTransactionID == "" {
			s.MapTransactionID = txn.TransactionID
		}
	case "Diameter":
		if s.DiameterSessionID == "" && txn.DiameterSessionID != "" {
			s.DiameterSessionID = txn.DiameterSessionID
		}
	case "GTP":
		if s.GtpTEID == 0 && txn.GtpTEID != 0 {
			s.GtpTEID = txn.GtpTEID
		}
	case "PFCP":
		if s.PfcpSEID == 0 && txn.PfcpSEID != 0 {
			s.PfcpSEID = txn.PfcpSEID
		}
	case "NGAP":
		if s.NgapUE_ID == 0 && txn.NgapAMF_UE_ID != 0 {
			s.NgapUE_ID = txn.NgapAMF_UE_ID
		}
	case "S1AP":
		if s.S1apMME_ID == 0 && txn.S1apMME_UE_ID != 0 {
			s.S1apMME_ID = txn.S1apMME_UE_ID
		}
	}
}

// addLocation adds a location update to the session
func (s *CorrelationSession) addLocation(location LocationUpdate) {
	s.LocationHistory = append(s.LocationHistory, location)
	s.CurrentLocation = &location
}

// GetSessionByID retrieves a session by ID
func (e *CorrelationEngine) GetSessionByID(sessionID string) (*CorrelationSession, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[sessionID]
	return session, exists
}

// GetSessionByIdentifier retrieves a session by any identifier
func (e *CorrelationEngine) GetSessionByIdentifier(idType IdentifierType, value string) (*CorrelationSession, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if index, exists := e.identifierIndex[idType]; exists {
		if session, found := index[value]; found {
			return session, true
		}
	}

	return nil, false
}

// GetSessionByTransaction retrieves a session by transaction ID
func (e *CorrelationEngine) GetSessionByTransaction(transactionID string) (*CorrelationSession, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.transactionIndex[transactionID]
	return session, exists
}

// GetSubscriberTimeline gets all sessions for a subscriber
func (e *CorrelationEngine) GetSubscriberTimeline(imsi string, startTime, endTime time.Time) ([]*CorrelationSession, error) {
	// First check in-memory sessions
	e.mu.RLock()
	memSessions := make([]*CorrelationSession, 0)
	if index, exists := e.identifierIndex[IdentifierIMSI]; exists {
		if session, found := index[imsi]; found {
			memSessions = append(memSessions, session)
		}
	}
	e.mu.RUnlock()

	// Query database for historical sessions
	if e.db == nil {
		return memSessions, nil
	}

	query := `
		SELECT DISTINCT
			s.id, s.start_time, s.end_time, s.status, s.session_type,
			s.bytes_uplink, s.bytes_downlink, s.success_rate, s.error_count
		FROM correlation_sessions s
		JOIN correlation_identifiers i ON s.id = i.session_id
		WHERE i.identifier_type = 'IMSI'
		  AND i.identifier_value = $1
		  AND s.start_time >= $2
		  AND s.end_time <= $3
		ORDER BY s.start_time DESC
	`

	rows, err := e.db.Query(query, imsi, startTime, endTime)
	if err != nil {
		return memSessions, err
	}
	defer rows.Close()

	dbSessions := make([]*CorrelationSession, 0)
	for rows.Next() {
		session := &CorrelationSession{
			Identifiers: make(map[IdentifierType][]Identifier),
		}

		err := rows.Scan(
			&session.ID,
			&session.StartTime,
			&session.EndTime,
			&session.Status,
			&session.SessionType,
			&session.BytesUplink,
			&session.BytesDownlink,
			&session.SuccessRate,
			&session.ErrorCount,
		)
		if err != nil {
			continue
		}

		dbSessions = append(dbSessions, session)
	}

	// Combine in-memory and database sessions
	allSessions := append(memSessions, dbSessions...)
	return allSessions, nil
}

// persistSession persists a session to the database
func (e *CorrelationEngine) persistSession(session *CorrelationSession) error {
	if e.db == nil {
		return nil
	}

	session.mutex.RLock()
	defer session.mutex.RUnlock()

	// Upsert session
	sessionQuery := `
		INSERT INTO correlation_sessions (
			id, start_time, end_time, status, session_type,
			bytes_uplink, bytes_downlink, success_rate, error_count,
			map_transaction_id, diameter_session_id, gtp_teid,
			pfcp_seid, ngap_ue_id, s1ap_mme_id, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
		ON CONFLICT (id) DO UPDATE SET
			end_time = EXCLUDED.end_time,
			status = EXCLUDED.status,
			bytes_uplink = EXCLUDED.bytes_uplink,
			bytes_downlink = EXCLUDED.bytes_downlink,
			success_rate = EXCLUDED.success_rate,
			error_count = EXCLUDED.error_count,
			updated_at = NOW()
	`

	_, err := e.db.Exec(sessionQuery,
		session.ID, session.StartTime, session.EndTime, session.Status, session.SessionType,
		session.BytesUplink, session.BytesDownlink, session.SuccessRate, session.ErrorCount,
		session.MapTransactionID, session.DiameterSessionID, session.GtpTEID,
		session.PfcpSEID, session.NgapUE_ID, session.S1apMME_ID,
	)
	if err != nil {
		return fmt.Errorf("failed to persist session: %w", err)
	}

	// Persist identifiers
	for idType, identifiers := range session.Identifiers {
		for _, id := range identifiers {
			idQuery := `
				INSERT INTO correlation_identifiers (
					session_id, identifier_type, identifier_value,
					protocol, first_seen, last_seen, confidence
				) VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (session_id, identifier_type, identifier_value) DO UPDATE SET
					last_seen = EXCLUDED.last_seen,
					confidence = EXCLUDED.confidence
			`

			_, err := e.db.Exec(idQuery,
				session.ID, string(idType), id.Value,
				id.Protocol, id.FirstSeen, id.LastSeen, id.Confidence,
			)
			if err != nil {
				return fmt.Errorf("failed to persist identifier: %w", err)
			}
		}
	}

	return nil
}

// cleanupExpiredSessions removes expired sessions from memory
func (e *CorrelationEngine) cleanupExpiredSessions() {
	ticker := time.NewTicker(e.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.performCleanup()
		case <-e.stopChan:
			return
		}
	}
}

// performCleanup removes expired sessions
func (e *CorrelationEngine) performCleanup() {
	e.mu.Lock()
	defer e.mu.Unlock()

	expiredSessions := make([]string, 0)
	cutoff := time.Now().Add(-e.sessionTimeout)

	for sessionID, session := range e.sessions {
		if session.EndTime.Before(cutoff) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	// Remove expired sessions
	for _, sessionID := range expiredSessions {
		session := e.sessions[sessionID]

		// Remove from identifier indexes
		for idType, identifiers := range session.Identifiers {
			for _, id := range identifiers {
				delete(e.identifierIndex[idType], id.Value)
			}
		}

		// Remove from transaction index
		for _, txnID := range session.Transactions {
			delete(e.transactionIndex, txnID)
		}

		// Remove session
		delete(e.sessions, sessionID)
	}
}

// Stop stops the correlation engine
func (e *CorrelationEngine) Stop() {
	close(e.stopChan)
}

// GetStats returns correlation engine statistics
func (e *CorrelationEngine) GetStats() CorrelationStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := CorrelationStats{
		TotalSessions:    len(e.sessions),
		ActiveSessions:   0,
		TotalIdentifiers: 0,
	}

	cutoff := time.Now().Add(-5 * time.Minute)
	for _, session := range e.sessions {
		if session.EndTime.After(cutoff) {
			stats.ActiveSessions++
		}

		for _, identifiers := range session.Identifiers {
			stats.TotalIdentifiers += len(identifiers)
		}
	}

	return stats
}

// Helper functions

func generateSessionID(txn *TransactionEvent) string {
	// Use primary identifier as session ID prefix
	for _, id := range txn.Identifiers {
		if id.Type == IdentifierIMSI && id.Value != "" {
			return fmt.Sprintf("SESS_%s_%d", id.Value, txn.Timestamp.Unix())
		}
	}

	return fmt.Sprintf("SESS_%s_%d", txn.TransactionID, txn.Timestamp.Unix())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CorrelationStats holds correlation engine statistics
type CorrelationStats struct {
	TotalSessions    int
	ActiveSessions   int
	TotalIdentifiers int
}
