package decoder

import (
	"time"
)

// Protocol represents a telecom protocol type
type Protocol string

const (
	ProtocolMAP      Protocol = "MAP"
	ProtocolCAP      Protocol = "CAP"
	ProtocolINAP     Protocol = "INAP"
	ProtocolDiameter Protocol = "Diameter"
	ProtocolGTPv1C   Protocol = "GTPv1-C"
	ProtocolGTPv2C   Protocol = "GTPv2-C"
	ProtocolPFCP     Protocol = "PFCP"
	ProtocolHTTP1    Protocol = "HTTP/1.1"
	ProtocolHTTP2    Protocol = "HTTP/2"
	ProtocolNGAP     Protocol = "NGAP"
	ProtocolS1AP     Protocol = "S1AP"
	ProtocolNAS4G    Protocol = "NAS-4G"
	ProtocolNAS5G    Protocol = "NAS-5G"
	ProtocolSCTP     Protocol = "SCTP"
	ProtocolTCP      Protocol = "TCP"
	ProtocolUDP      Protocol = "UDP"
	ProtocolUnknown  Protocol = "Unknown"
)

// Message represents a decoded telecom protocol message
type Message struct {
	// Basic identification
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Protocol     Protocol               `json:"protocol"`
	MessageType  string                 `json:"message_type"`
	MessageName  string                 `json:"message_name"`
	Direction    Direction              `json:"direction"`

	// Network elements
	Source       NetworkElement         `json:"source"`
	Destination  NetworkElement         `json:"destination"`

	// Protocol-specific details
	Details      map[string]interface{} `json:"details"`

	// Correlation fields
	IMSI         string                 `json:"imsi,omitempty"`
	MSISDN       string                 `json:"msisdn,omitempty"`
	SUPI         string                 `json:"supi,omitempty"`
	TEID         uint32                 `json:"teid,omitempty"`
	SEID         uint64                 `json:"seid,omitempty"`
	PLMN         string                 `json:"plmn,omitempty"`
	CellID       string                 `json:"cell_id,omitempty"`
	APN          string                 `json:"apn,omitempty"`
	DNN          string                 `json:"dnn,omitempty"`

	// Session correlation
	SessionID    string                 `json:"session_id,omitempty"`
	TransactionID string                `json:"transaction_id,omitempty"`
	SequenceNum  uint32                 `json:"sequence_num,omitempty"`

	// Result and cause
	Result       Result                 `json:"result"`
	CauseCode    int                    `json:"cause_code,omitempty"`
	CauseText    string                 `json:"cause_text,omitempty"`

	// Raw data
	RawPayload   []byte                 `json:"-"`
	PayloadSize  int                    `json:"payload_size"`

	// Metadata
	VendorID     string                 `json:"vendor_id,omitempty"`
	VendorName   string                 `json:"vendor_name,omitempty"`
	InterfaceType string                `json:"interface_type,omitempty"`

	// Performance
	ProcessedAt  time.Time              `json:"processed_at"`
	DecodeTimeUs int64                  `json:"decode_time_us"`
}

// Direction represents message direction
type Direction string

const (
	DirectionRequest  Direction = "request"
	DirectionResponse Direction = "response"
	DirectionUnknown  Direction = "unknown"
)

// Result represents the outcome of a procedure
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
	ResultTimeout Result = "timeout"
	ResultUnknown Result = "unknown"
)

// NetworkElement represents a network node
type NetworkElement struct {
	Type    string `json:"type"`    // MME, SGW, PGW, HSS, AMF, SMF, UPF, etc.
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Port    uint16 `json:"port,omitempty"`
	FQDN    string `json:"fqdn,omitempty"`
	Realm   string `json:"realm,omitempty"`
	GT      string `json:"gt,omitempty"`      // Global Title for SS7
	PC      string `json:"pc,omitempty"`      // Point Code for SS7
}

// Decoder is the interface that all protocol decoders must implement
type Decoder interface {
	// Decode decodes a raw message into a structured Message
	Decode(data []byte, metadata *Metadata) (*Message, error)

	// Protocol returns the protocol this decoder handles
	Protocol() Protocol

	// CanDecode checks if this decoder can handle the given data
	CanDecode(data []byte) bool
}

// Metadata provides context for decoding
type Metadata struct {
	CaptureTime    time.Time
	SourceIP       string
	DestIP         string
	SourcePort     uint16
	DestPort       uint16
	TransportProto string
	InterfaceName  string
	VendorHint     string
}

// DecoderRegistry manages all protocol decoders
type DecoderRegistry struct {
	decoders map[Protocol]Decoder
}

// NewRegistry creates a new decoder registry
func NewRegistry() *DecoderRegistry {
	return &DecoderRegistry{
		decoders: make(map[Protocol]Decoder),
	}
}

// Register registers a decoder for a protocol
func (r *DecoderRegistry) Register(decoder Decoder) {
	r.decoders[decoder.Protocol()] = decoder
}

// Get returns a decoder for the specified protocol
func (r *DecoderRegistry) Get(protocol Protocol) (Decoder, bool) {
	decoder, ok := r.decoders[protocol]
	return decoder, ok
}

// Decode attempts to decode data using all registered decoders
func (r *DecoderRegistry) Decode(data []byte, metadata *Metadata) (*Message, error) {
	for _, decoder := range r.decoders {
		if decoder.CanDecode(data) {
			return decoder.Decode(data, metadata)
		}
	}
	return nil, ErrNoDecoderFound
}

// Error types
type DecoderError struct {
	Protocol Protocol
	Message  string
	Err      error
}

func (e *DecoderError) Error() string {
	if e.Err != nil {
		return e.Protocol + " decoder error: " + e.Message + ": " + e.Err.Error()
	}
	return string(e.Protocol) + " decoder error: " + e.Message
}

func (e *DecoderError) Unwrap() error {
	return e.Err
}

var (
	ErrNoDecoderFound     = &DecoderError{Message: "no suitable decoder found"}
	ErrInvalidData        = &DecoderError{Message: "invalid data"}
	ErrInsufficientData   = &DecoderError{Message: "insufficient data"}
	ErrUnsupportedVersion = &DecoderError{Message: "unsupported version"}
)
