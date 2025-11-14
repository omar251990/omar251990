package diameter

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// DiameterDecoder handles Diameter protocol messages
type DiameterDecoder struct {
	applications  []string
	vendorSupport []string
}

// NewDiameterDecoder creates a new Diameter decoder
func NewDiameterDecoder(applications, vendorSupport []string) *DiameterDecoder {
	return &DiameterDecoder{
		applications:  applications,
		vendorSupport: vendorSupport,
	}
}

// Protocol returns the protocol type
func (d *DiameterDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolDiameter
}

// CanDecode checks if the data is a Diameter message
func (d *DiameterDecoder) CanDecode(data []byte) bool {
	if len(data) < 20 {
		return false
	}

	// Diameter messages start with version (0x01) and have specific length format
	version := data[0]
	if version != 0x01 {
		return false
	}

	// Check message length
	length := binary.BigEndian.Uint32(data[0:4]) & 0x00FFFFFF
	return length >= 20 && int(length) <= len(data)
}

// Decode decodes a Diameter message
func (d *DiameterDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 20 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolDiameter,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse Diameter header (20 bytes)
	header := d.parseHeader(data)
	msg.Details["version"] = header.Version
	msg.Details["command_code"] = header.CommandCode
	msg.Details["application_id"] = header.ApplicationID
	msg.Details["hop_by_hop_id"] = header.HopByHopID
	msg.Details["end_to_end_id"] = header.EndToEndID
	msg.Details["flags"] = header.Flags

	// Determine message type and direction
	msg.MessageName = getCommandName(header.CommandCode, header.ApplicationID)
	msg.MessageType = fmt.Sprintf("Diameter_%s", msg.MessageName)

	if header.Flags&0x80 != 0 { // Request flag
		msg.Direction = decoder.DirectionRequest
	} else {
		msg.Direction = decoder.DirectionResponse
	}

	// Parse AVPs
	avps, err := d.parseAVPs(data[20:])
	if err == nil {
		msg.Details["avps"] = avps

		// Extract key correlation fields
		d.extractCorrelationFields(msg, avps)

		// Extract result code for responses
		if msg.Direction == decoder.DirectionResponse {
			if resultCode, ok := avps["Result-Code"].(uint32); ok {
				msg.CauseCode = int(resultCode)
				if resultCode >= 2000 && resultCode < 3000 {
					msg.Result = decoder.ResultSuccess
				} else {
					msg.Result = decoder.ResultFailure
					msg.CauseText = getDiameterResultText(resultCode)
				}
			}
		}
	}

	// Identify network elements based on application
	d.identifyNetworkElements(msg, header)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// DiameterHeader represents the Diameter message header
type DiameterHeader struct {
	Version       uint8
	Length        uint32
	Flags         uint8
	CommandCode   uint32
	ApplicationID uint32
	HopByHopID    uint32
	EndToEndID    uint32
}

// parseHeader extracts Diameter header fields
func (d *DiameterDecoder) parseHeader(data []byte) *DiameterHeader {
	header := &DiameterHeader{}

	header.Version = data[0]
	header.Length = binary.BigEndian.Uint32(data[0:4]) & 0x00FFFFFF
	header.Flags = data[4]
	header.CommandCode = binary.BigEndian.Uint32(data[4:8]) & 0x00FFFFFF
	header.ApplicationID = binary.BigEndian.Uint32(data[8:12])
	header.HopByHopID = binary.BigEndian.Uint32(data[12:16])
	header.EndToEndID = binary.BigEndian.Uint32(data[16:20])

	return header
}

// parseAVPs parses Diameter AVPs (Attribute-Value Pairs)
func (d *DiameterDecoder) parseAVPs(data []byte) (map[string]interface{}, error) {
	avps := make(map[string]interface{})
	offset := 0

	for offset < len(data)-8 {
		// AVP header: code (4), flags (1), length (3)
		avpCode := binary.BigEndian.Uint32(data[offset : offset+4])
		flags := data[offset+4]
		avpLength := int(binary.BigEndian.Uint32(data[offset+4:offset+8]) & 0x00FFFFFF)

		if avpLength < 8 || offset+avpLength > len(data) {
			break
		}

		// Check if vendor-specific
		headerLength := 8
		var vendorID uint32
		if flags&0x80 != 0 { // Vendor-specific flag
			vendorID = binary.BigEndian.Uint32(data[offset+8 : offset+12])
			headerLength = 12
		}

		// Extract value
		valueLength := avpLength - headerLength
		if offset+headerLength+valueLength > len(data) {
			break
		}

		value := data[offset+headerLength : offset+headerLength+valueLength]
		avpName := getAVPName(avpCode, vendorID)

		// Parse common AVPs
		parsedValue := d.parseAVPValue(avpCode, value)
		avps[avpName] = parsedValue

		// Move to next AVP (aligned to 4 bytes)
		offset += avpLength
		padding := (4 - (avpLength % 4)) % 4
		offset += padding
	}

	return avps, nil
}

// parseAVPValue parses AVP value based on type
func (d *DiameterDecoder) parseAVPValue(code uint32, value []byte) interface{} {
	switch code {
	case 1: // User-Name (UTF8String)
		return string(value)
	case 25: // Class (OctetString)
		return value
	case 263, 264, 268: // Session-Id, Origin-Host, Result-Code (various)
		if code == 268 && len(value) == 4 {
			return binary.BigEndian.Uint32(value)
		}
		return string(value)
	case 283: // Destination-Realm (UTF8String)
		return string(value)
	case 293: // Destination-Host (UTF8String)
		return string(value)
	case 1: // User-Name
		return string(value)
	case 443: // Subscription-Id-Data (for IMSI/MSISDN)
		return string(value)
	default:
		// Try to parse as uint32 if 4 bytes, otherwise return raw
		if len(value) == 4 {
			return binary.BigEndian.Uint32(value)
		}
		return value
	}
}

// extractCorrelationFields extracts key fields for correlation
func (d *DiameterDecoder) extractCorrelationFields(msg *decoder.Message, avps map[string]interface{}) {
	// Session ID
	if sessionID, ok := avps["Session-Id"].(string); ok {
		msg.SessionID = sessionID
	}

	// IMSI (from User-Name or Subscription-Id-Data)
	if userName, ok := avps["User-Name"].(string); ok {
		if len(userName) == 15 && isNumeric(userName) {
			msg.IMSI = userName
		}
	}

	// Subscription-Id-Data (can be IMSI or MSISDN)
	if subData, ok := avps["Subscription-Id-Data"].(string); ok {
		if len(subData) == 15 {
			msg.IMSI = subData
		} else {
			msg.MSISDN = subData
		}
	}

	// Visited-PLMN-Id
	if plmn, ok := avps["Visited-PLMN-Id"].([]byte); ok && len(plmn) >= 3 {
		msg.PLMN = decodePLMN(plmn)
	}

	// APN
	if apn, ok := avps["Called-Station-Id"].(string); ok {
		msg.APN = apn
	}
}

// identifyNetworkElements identifies source and destination network elements
func (d *DiameterDecoder) identifyNetworkElements(msg *decoder.Message, header *DiameterHeader) {
	switch header.ApplicationID {
	case 16777251: // S6a/S6d
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "MME"
			msg.Destination.Type = "HSS"
		} else {
			msg.Source.Type = "HSS"
			msg.Destination.Type = "MME"
		}
	case 16777238: // Gx
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "PCEF"
			msg.Destination.Type = "PCRF"
		} else {
			msg.Source.Type = "PCRF"
			msg.Destination.Type = "PCEF"
		}
	case 16777238: // Gy
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "PGW"
			msg.Destination.Type = "OCS"
		} else {
			msg.Source.Type = "OCS"
			msg.Destination.Type = "PGW"
		}
	case 16777302: // S6t
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SCEF"
			msg.Destination.Type = "HSS"
		} else {
			msg.Source.Type = "HSS"
			msg.Destination.Type = "SCEF"
		}
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}

	// Set realm from AVPs if available
	if avps, ok := msg.Details["avps"].(map[string]interface{}); ok {
		if originHost, ok := avps["Origin-Host"].(string); ok {
			msg.Source.FQDN = originHost
		}
		if destHost, ok := avps["Destination-Host"].(string); ok {
			msg.Destination.FQDN = destHost
		}
		if originRealm, ok := avps["Origin-Realm"].(string); ok {
			msg.Source.Realm = originRealm
		}
		if destRealm, ok := avps["Destination-Realm"].(string); ok {
			msg.Destination.Realm = destRealm
		}
	}
}

// getCommandName returns the Diameter command name
func getCommandName(code, appID uint32) string {
	// Common commands
	commands := map[uint32]string{
		257: "CER", // Capabilities-Exchange-Request
		258: "CEA", // Capabilities-Exchange-Answer
		280: "DWR", // Device-Watchdog-Request
		282: "DWA", // Device-Watchdog-Answer
	}

	if name, ok := commands[code]; ok {
		return name
	}

	// Application-specific commands
	switch appID {
	case 16777251: // S6a/S6d
		s6aCommands := map[uint32]string{
			316: "ULR", // Update-Location-Request
			317: "ULA", // Update-Location-Answer
			318: "AIR", // Authentication-Information-Request
			319: "AIA", // Authentication-Information-Answer
			321: "PUR", // Purge-UE-Request
			322: "PUA", // Purge-UE-Answer
		}
		if name, ok := s6aCommands[code]; ok {
			return name
		}
	case 16777238: // Gx
		gxCommands := map[uint32]string{
			258: "RAR", // Re-Auth-Request
			265: "AAR", // AA-Request
			272: "CCR", // Credit-Control-Request
			273: "CCA", // Credit-Control-Answer
		}
		if name, ok := gxCommands[code]; ok {
			return name
		}
	}

	return fmt.Sprintf("CMD_%d", code)
}

// getAVPName returns the AVP name for a code
func getAVPName(code, vendorID uint32) string {
	if vendorID == 0 {
		// Standard AVPs
		standardAVPs := map[uint32]string{
			1:   "User-Name",
			25:  "Class",
			27:  "Session-Timeout",
			33:  "Proxy-State",
			44:  "Accounting-Session-Id",
			50:  "Acct-Multi-Session-Id",
			85:  "Acct-Interim-Interval",
			263: "Session-Id",
			264: "Origin-Host",
			268: "Result-Code",
			269: "Product-Name",
			283: "Destination-Realm",
			293: "Destination-Host",
			296: "Origin-State-Id",
			443: "Subscription-Id-Data",
			444: "Subscription-Id-Type",
			450: "Subscription-Id",
			1400: "User-Location-Info",
			1401: "MSISDN",
			1402: "IMSI",
			1405: "RAT-Type",
			1407: "Visited-PLMN-Id",
			1408: "QoS-Information",
			1409: "APN",
		}
		if name, ok := standardAVPs[code]; ok {
			return name
		}
	}

	return fmt.Sprintf("AVP_%d_Vendor_%d", code, vendorID)
}

// getDiameterResultText returns result code description
func getDiameterResultText(code uint32) string {
	results := map[uint32]string{
		2001: "Success",
		2002: "Limited Success",
		3001: "Command Unsupported",
		3002: "Unable to Deliver",
		3003: "Realm Not Served",
		3004: "Too Busy",
		3007: "Application Unsupported",
		4001: "Authentication Rejected",
		4010: "No Common Application",
		5001: "User Unknown",
		5002: "Rating Failed",
		5003: "Credit Control Not Applicable",
		5004: "Credit Limit Reached",
		5005: "User Unknown (Roaming)",
		5012: "Unknown Session Id",
		5030: "User Not Registered",
	}

	if text, ok := results[code]; ok {
		return text
	}
	return fmt.Sprintf("Result_%d", code)
}

// decodePLMN decodes PLMN from bytes
func decodePLMN(data []byte) string {
	if len(data) < 3 {
		return ""
	}
	mcc := fmt.Sprintf("%d%d%d", data[0]&0x0F, (data[0]>>4)&0x0F, data[1]&0x0F)
	mnc := fmt.Sprintf("%d%d", (data[1]>>4)&0x0F, data[2]&0x0F)
	return mcc + mnc
}

// isNumeric checks if string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
