package correlation

import "time"

// TransactionEvent represents a transaction event from any protocol
type TransactionEvent struct {
	TransactionID string
	Protocol      string
	Timestamp     time.Time
	SessionType   string // "voice", "data", "sms", "location_update", etc.

	// Identifiers
	Identifiers []Identifier

	// Protocol-specific IDs
	DiameterSessionID string
	GtpTEID           uint32
	PfcpSEID          uint64
	NgapAMF_UE_ID     uint64
	NgapRAN_UE_ID     uint32
	S1apMME_UE_ID     uint32
	S1apeNB_UE_ID     uint32

	// Location information
	Location *LocationUpdate

	// Data usage
	BytesUplink   uint64
	BytesDownlink uint64

	// Quality metrics
	Success bool
	Latency time.Duration

	// Additional metadata
	Metadata map[string]interface{}
}

// NewTransactionEvent creates a new transaction event
func NewTransactionEvent(protocol, transactionID string) *TransactionEvent {
	return &TransactionEvent{
		Protocol:      protocol,
		TransactionID: transactionID,
		Timestamp:     time.Now(),
		Identifiers:   make([]Identifier, 0),
		Metadata:      make(map[string]interface{}),
	}
}

// AddIdentifier adds an identifier to the transaction
func (t *TransactionEvent) AddIdentifier(idType IdentifierType, value, protocol string) {
	if value == "" {
		return
	}

	identifier := Identifier{
		Type:       idType,
		Value:      value,
		Protocol:   protocol,
		FirstSeen:  t.Timestamp,
		LastSeen:   t.Timestamp,
		Confidence: 1.0,
	}

	t.Identifiers = append(t.Identifiers, identifier)
}

// AddIMSI adds an IMSI identifier
func (t *TransactionEvent) AddIMSI(imsi string) {
	t.AddIdentifier(IdentifierIMSI, imsi, t.Protocol)
}

// AddMSISDN adds an MSISDN identifier
func (t *TransactionEvent) AddMSISDN(msisdn string) {
	t.AddIdentifier(IdentifierMSISDN, msisdn, t.Protocol)
}

// AddIMEI adds an IMEI identifier
func (t *TransactionEvent) AddIMEI(imei string) {
	t.AddIdentifier(IdentifierIMEI, imei, t.Protocol)
}

// AddTEID adds a GTP TEID identifier
func (t *TransactionEvent) AddTEID(teid uint32) {
	t.GtpTEID = teid
	t.AddIdentifier(IdentifierTEID, formatUint32(teid), t.Protocol)
}

// AddSEID adds a PFCP SEID identifier
func (t *TransactionEvent) AddSEID(seid uint64) {
	t.PfcpSEID = seid
	t.AddIdentifier(IdentifierSEID, formatUint64(seid), t.Protocol)
}

// AddIP adds an IP address identifier
func (t *TransactionEvent) AddIP(ip string) {
	t.AddIdentifier(IdentifierIP, ip, t.Protocol)
}

// AddLocation adds location information
func (t *TransactionEvent) AddLocation(mcc, mnc, lac, cellID string) {
	t.Location = &LocationUpdate{
		Timestamp: t.Timestamp,
		Protocol:  t.Protocol,
		MCC:       mcc,
		MNC:       mnc,
		LAC:       lac,
		CellID:    cellID,
	}
}

// Add5GLocation adds 5G-specific location information
func (t *TransactionEvent) Add5GLocation(mcc, mnc, tac, globalRAN_ID string) {
	t.Location = &LocationUpdate{
		Timestamp:    t.Timestamp,
		Protocol:     t.Protocol,
		MCC:          mcc,
		MNC:          mnc,
		TAC:          tac,
		GlobalRAN_ID: globalRAN_ID,
	}
}

// Add4GLocation adds 4G-specific location information
func (t *TransactionEvent) Add4GLocation(mcc, mnc, tac, eutran_cgi string) {
	t.Location = &LocationUpdate{
		Timestamp:  t.Timestamp,
		Protocol:   t.Protocol,
		MCC:        mcc,
		MNC:        mnc,
		TAC:        tac,
		EUTRAN_CGI: eutran_cgi,
	}
}

// SetDataUsage sets data usage statistics
func (t *TransactionEvent) SetDataUsage(uplink, downlink uint64) {
	t.BytesUplink = uplink
	t.BytesDownlink = downlink
}

// SetQuality sets quality metrics
func (t *TransactionEvent) SetQuality(success bool, latency time.Duration) {
	t.Success = success
	t.Latency = latency
}

// Helper functions
func formatUint32(value uint32) string {
	return string(rune(value))
}

func formatUint64(value uint64) string {
	return string(rune(value))
}
