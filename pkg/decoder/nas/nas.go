package nas

import (
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// NASDecoder handles Non-Access Stratum messages (4G/5G)
type NASDecoder struct {
	generations []string // 4G, 5G
}

// NewNASDecoder creates a new NAS decoder
func NewNASDecoder(generations []string) *NASDecoder {
	return &NASDecoder{
		generations: generations,
	}
}

// Protocol returns the protocol type
func (d *NASDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolNAS5G // Will detect 4G vs 5G
}

// CanDecode checks if the data is a NAS message
func (d *NASDecoder) CanDecode(data []byte) bool {
	if len(data) < 3 {
		return false
	}

	// Check for EPS/5GS security header type
	secHeaderType := (data[0] >> 4) & 0x0F
	return secHeaderType <= 0x0F
}

// Decode decodes a NAS message
func (d *NASDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 3 {
		return nil, decoder.ErrInsufficientData
	}

	// Determine generation from protocol discriminator
	epd := data[0] & 0x0F
	var generation string
	var protocol decoder.Protocol

	if epd == 0x07 { // EPS Mobility Management
		generation = "4G"
		protocol = decoder.ProtocolNAS4G
	} else if epd == 0x0F { // 5GS Mobility Management
		generation = "5G"
		protocol = decoder.ProtocolNAS5G
	} else if epd == 0x02 { // EPS Session Management
		generation = "4G"
		protocol = decoder.ProtocolNAS4G
	} else {
		generation = "Unknown"
		protocol = decoder.ProtocolUnknown
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    protocol,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse security header
	secHeaderType := (data[0] >> 4) & 0x0F
	msg.Details["security_header_type"] = secHeaderType
	msg.Details["generation"] = generation

	// Skip security header if present
	offset := 0
	if secHeaderType != 0 {
		// Plain NAS message, security header present
		offset = 7 // Skip MAC + sequence number
	}

	if offset+2 > len(data) {
		return msg, nil
	}

	// Parse message type
	messageType := data[offset+1]
	msg.MessageName = getNASMessageName(epd, messageType)
	msg.MessageType = fmt.Sprintf("NAS_%s_%s", generation, msg.MessageName)
	msg.Details["message_type"] = messageType

	// Determine direction based on message type
	if isUplink(epd, messageType) {
		msg.Direction = decoder.DirectionRequest
		msg.Source.Type = "UE"
		msg.Destination.Type = getDestinationNode(generation)
	} else {
		msg.Direction = decoder.DirectionResponse
		msg.Source.Type = getDestinationNode(generation)
		msg.Destination.Type = "UE"
	}

	// Parse IEs
	ies := d.parseIEs(data[offset+2:], epd, messageType)
	msg.Details["ies"] = ies

	// Extract correlation fields
	d.extractCorrelationFields(msg, ies, generation)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// parseIEs parses NAS Information Elements
func (d *NASDecoder) parseIEs(data []byte, epd, messageType uint8) map[string]interface{} {
	ies := make(map[string]interface{})

	// Simplified IE parsing
	// Full implementation would decode each IE based on message type
	offset := 0
	for offset < len(data)-2 {
		ieType := data[offset]
		ieLen := int(data[offset+1])

		if offset+2+ieLen > len(data) {
			break
		}

		ieName := getNASIEName(ieType)
		ies[ieName] = data[offset+2 : offset+2+ieLen]

		offset += 2 + ieLen
	}

	return ies
}

// extractCorrelationFields extracts key fields for correlation
func (d *NASDecoder) extractCorrelationFields(msg *decoder.Message, ies map[string]interface{}, generation string) {
	// Extract IMSI/SUPI, GUTI/5G-GUTI, etc. from IEs
	// Simplified - would parse actual IE structures in production
}

// getNASMessageName returns message name
func getNASMessageName(epd, messageType uint8) string {
	if epd == 0x07 { // 4G MM
		messages4GMM := map[uint8]string{
			0x41: "AttachRequest",
			0x42: "AttachAccept",
			0x43: "AttachComplete",
			0x44: "AttachReject",
			0x45: "DetachRequest",
			0x46: "DetachAccept",
			0x48: "TrackingAreaUpdateRequest",
			0x49: "TrackingAreaUpdateAccept",
			0x4a: "TrackingAreaUpdateComplete",
			0x4b: "TrackingAreaUpdateReject",
			0x4c: "ExtendedServiceRequest",
			0x4e: "ServiceReject",
			0x50: "GUTIReallocationCommand",
			0x51: "GUTIReallocationComplete",
			0x52: "AuthenticationRequest",
			0x53: "AuthenticationResponse",
			0x54: "AuthenticationReject",
			0x55: "AuthenticationFailure",
			0x5c: "SecurityModeCommand",
			0x5d: "SecurityModeComplete",
			0x5e: "SecurityModeReject",
		}
		if name, ok := messages4GMM[messageType]; ok {
			return name
		}
	} else if epd == 0x0F { // 5G MM
		messages5GMM := map[uint8]string{
			0x41: "RegistrationRequest",
			0x42: "RegistrationAccept",
			0x43: "RegistrationComplete",
			0x44: "RegistrationReject",
			0x45: "DeregistrationRequest",
			0x46: "DeregistrationAccept",
			0x4c: "ServiceRequest",
			0x4d: "ServiceReject",
			0x4e: "ServiceAccept",
			0x54: "ConfigurationUpdateCommand",
			0x55: "ConfigurationUpdateComplete",
			0x56: "AuthenticationRequest",
			0x57: "AuthenticationResponse",
			0x58: "AuthenticationReject",
			0x59: "AuthenticationFailure",
			0x5a: "AuthenticationResult",
			0x5b: "IdentityRequest",
			0x5c: "IdentityResponse",
			0x5d: "SecurityModeCommand",
			0x5e: "SecurityModeComplete",
			0x5f: "SecurityModeReject",
		}
		if name, ok := messages5GMM[messageType]; ok {
			return name
		}
	} else if epd == 0x02 { // 4G SM
		messages4GSM := map[uint8]string{
			0xc1: "ActivateDefaultEPSBearerContextRequest",
			0xc2: "ActivateDefaultEPSBearerContextAccept",
			0xc3: "ActivateDefaultEPSBearerContextReject",
			0xc5: "ActivateDedicatedEPSBearerContextRequest",
			0xc6: "ActivateDedicatedEPSBearerContextAccept",
			0xc7: "ActivateDedicatedEPSBearerContextReject",
			0xcd: "PDNConnectivityRequest",
			0xce: "PDNConnectivityReject",
			0xd0: "PDNDisconnectRequest",
			0xd1: "PDNDisconnectReject",
		}
		if name, ok := messages4GSM[messageType]; ok {
			return name
		}
	}

	return fmt.Sprintf("Unknown_%02x", messageType)
}

// getNASIEName returns IE name for type
func getNASIEName(ieType uint8) string {
	ies := map[uint8]string{
		0x50: "GUTI",
		0x52: "MobileIdentity",
		0x53: "TAI",
		0x58: "UENetworkCapability",
		0x59: "ESMMessageContainer",
		0x5a: "NASKeySetIdentifier",
		0x5c: "EMM Cause",
		0x5d: "ESM Cause",
	}

	if name, ok := ies[ieType]; ok {
		return name
	}
	return fmt.Sprintf("IE_%02x", ieType)
}

// isUplink determines if message is from UE
func isUplink(epd, messageType uint8) bool {
	uplinkMessages := map[uint8]bool{
		0x41: true, // AttachRequest / RegistrationRequest
		0x43: true, // AttachComplete / RegistrationComplete
		0x45: true, // DetachRequest / DeregistrationRequest
		0x48: true, // TAU Request
		0x4a: true, // TAU Complete
		0x4c: true, // Service Request
		0x51: true, // GUTI Reallocation Complete
		0x53: true, // Authentication Response
		0x55: true, // Authentication Failure
		0x5d: true, // Security Mode Complete
		0x5e: true, // Security Mode Reject
	}

	return uplinkMessages[messageType]
}

// getDestinationNode returns core node name
func getDestinationNode(generation string) string {
	if generation == "4G" {
		return "MME"
	} else if generation == "5G" {
		return "AMF"
	}
	return "Unknown"
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
