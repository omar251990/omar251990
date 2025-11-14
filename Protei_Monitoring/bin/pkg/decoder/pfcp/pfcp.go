package pfcp

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// PFCPDecoder handles Packet Forwarding Control Protocol (5G N4 interface)
type PFCPDecoder struct{}

// NewPFCPDecoder creates a new PFCP decoder
func NewPFCPDecoder() *PFCPDecoder {
	return &PFCPDecoder{}
}

// Protocol returns the protocol type
func (d *PFCPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolPFCP
}

// CanDecode checks if the data is a PFCP message
func (d *PFCPDecoder) CanDecode(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// PFCP version should be 1 (bits 5-7 of first byte)
	version := (data[0] >> 5) & 0x07
	return version == 1
}

// Decode decodes a PFCP message
func (d *PFCPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 8 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolPFCP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse PFCP header
	flags := data[0]
	version := (flags >> 5) & 0x07
	messageType := data[1]
	length := binary.BigEndian.Uint16(data[2:4])

	msg.Details["version"] = version
	msg.Details["flags"] = flags
	msg.Details["length"] = length
	msg.MessageType = fmt.Sprintf("PFCP_%s", getPFCPMessageName(messageType))
	msg.MessageName = getPFCPMessageName(messageType)

	// Determine direction
	if messageType%2 == 1 { // Odd = request
		msg.Direction = decoder.DirectionRequest
	} else { // Even = response
		msg.Direction = decoder.DirectionResponse
	}

	// Extract SEID if present
	if flags&0x01 != 0 { // SEID present
		if len(data) >= 16 {
			seid := binary.BigEndian.Uint64(data[4:12])
			msg.SEID = seid
			seqNum := binary.BigEndian.Uint32(data[12:16]) & 0x00FFFFFF
			msg.SequenceNum = seqNum
		}
	} else {
		if len(data) >= 8 {
			seqNum := binary.BigEndian.Uint32(data[4:8]) & 0x00FFFFFF
			msg.SequenceNum = seqNum
		}
	}

	// Parse IEs
	headerLen := 8
	if flags&0x01 != 0 {
		headerLen = 16
	}

	if len(data) > headerLen {
		ies := d.parseIEs(data[headerLen:])
		msg.Details["ies"] = ies

		// Extract correlation fields
		d.extractCorrelationFields(msg, ies)

		// Determine result for responses
		if msg.Direction == decoder.DirectionResponse {
			if cause, ok := ies["Cause"].(int); ok {
				msg.CauseCode = cause
				if cause == 1 { // Request accepted
					msg.Result = decoder.ResultSuccess
				} else {
					msg.Result = decoder.ResultFailure
					msg.CauseText = getPFCPCauseText(cause)
				}
			}
		}
	}

	// Identify network elements
	d.identifyNetworkElements(msg, messageType)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// parseIEs parses PFCP Information Elements
func (d *PFCPDecoder) parseIEs(data []byte) map[string]interface{} {
	ies := make(map[string]interface{})
	offset := 0

	for offset < len(data)-4 {
		ieType := binary.BigEndian.Uint16(data[offset : offset+2])
		ieLen := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))

		if offset+4+ieLen > len(data) {
			break
		}

		ieData := data[offset+4 : offset+4+ieLen]
		ieName := getPFCPIEName(ieType)

		// Parse specific IEs
		switch ieType {
		case 19: // Cause
			if len(ieData) > 0 {
				ies[ieName] = int(ieData[0])
			}
		case 21: // F-SEID
			if len(ieData) >= 13 {
				seid := binary.BigEndian.Uint64(ieData[1:9])
				ipv4 := fmt.Sprintf("%d.%d.%d.%d", ieData[9], ieData[10], ieData[11], ieData[12])
				ies[ieName] = map[string]interface{}{
					"seid": seid,
					"ipv4": ipv4,
				}
			}
		case 56: // PDR ID
			if len(ieData) >= 2 {
				ies[ieName] = binary.BigEndian.Uint16(ieData[0:2])
			}
		default:
			ies[ieName] = ieData
		}

		offset += 4 + ieLen
	}

	return ies
}

// extractCorrelationFields extracts key fields
func (d *PFCPDecoder) extractCorrelationFields(msg *decoder.Message, ies map[string]interface{}) {
	// Extract F-SEID, Node ID, etc. from IEs
	if fseid, ok := ies["F-SEID"].(map[string]interface{}); ok {
		if seid, ok := fseid["seid"].(uint64); ok {
			msg.SEID = seid
		}
	}
}

// identifyNetworkElements identifies source and destination
func (d *PFCPDecoder) identifyNetworkElements(msg *decoder.Message, messageType uint8) {
	switch messageType {
	case 50, 51: // Session Establishment Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SMF"
			msg.Destination.Type = "UPF"
		} else {
			msg.Source.Type = "UPF"
			msg.Destination.Type = "SMF"
		}
	case 52, 53: // Session Modification Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SMF"
			msg.Destination.Type = "UPF"
		} else {
			msg.Source.Type = "UPF"
			msg.Destination.Type = "SMF"
		}
	case 54, 55: // Session Deletion Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SMF"
			msg.Destination.Type = "UPF"
		} else {
			msg.Source.Type = "UPF"
			msg.Destination.Type = "SMF"
		}
	case 56, 57: // Session Report Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "UPF"
			msg.Destination.Type = "SMF"
		} else {
			msg.Source.Type = "SMF"
			msg.Destination.Type = "UPF"
		}
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// getPFCPMessageName returns message name
func getPFCPMessageName(messageType uint8) string {
	messages := map[uint8]string{
		// Node related messages
		1:  "HeartbeatRequest",
		2:  "HeartbeatResponse",
		3:  "PFDManagementRequest",
		4:  "PFDManagementResponse",
		5:  "AssociationSetupRequest",
		6:  "AssociationSetupResponse",
		7:  "AssociationUpdateRequest",
		8:  "AssociationUpdateResponse",
		9:  "AssociationReleaseRequest",
		10: "AssociationReleaseResponse",
		11: "VersionNotSupportedResponse",
		12: "NodeReportRequest",
		13: "NodeReportResponse",
		14: "SessionSetDeletionRequest",
		15: "SessionSetDeletionResponse",
		// Session related messages
		50: "SessionEstablishmentRequest",
		51: "SessionEstablishmentResponse",
		52: "SessionModificationRequest",
		53: "SessionModificationResponse",
		54: "SessionDeletionRequest",
		55: "SessionDeletionResponse",
		56: "SessionReportRequest",
		57: "SessionReportResponse",
	}

	if name, ok := messages[messageType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown_%d", messageType)
}

// getPFCPIEName returns IE name
func getPFCPIEName(ieType uint16) string {
	ies := map[uint16]string{
		2:  "Create-PDR",
		7:  "Create-FAR",
		19: "Cause",
		21: "F-SEID",
		22: "Node-ID",
		44: "UE-IP-Address",
		56: "PDR-ID",
		108: "FAR-ID",
		109: "QER-ID",
		137: "Activate-Predefined-Rules",
	}

	if name, ok := ies[ieType]; ok {
		return name
	}
	return fmt.Sprintf("IE_%d", ieType)
}

// getPFCPCauseText returns cause description
func getPFCPCauseText(cause int) string {
	causes := map[int]string{
		1:   "Request accepted",
		64:  "Request rejected",
		65:  "Session context not found",
		66:  "Mandatory IE missing",
		67:  "Conditional IE missing",
		68:  "Invalid length",
		69:  "Mandatory IE incorrect",
		70:  "Invalid Forwarding Policy",
		71:  "Invalid F-TEID allocation option",
		72:  "No established PFCP Association",
		73:  "Rule creation/modification Failure",
		74:  "PFCP entity in congestion",
		75:  "No resources available",
		76:  "Service not supported",
		77:  "System failure",
	}

	if text, ok := causes[cause]; ok {
		return text
	}
	return fmt.Sprintf("Cause_%d", cause)
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
