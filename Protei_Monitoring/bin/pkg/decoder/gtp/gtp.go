package gtp

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// GTPDecoder handles GTP-C (Control Plane) protocol
type GTPDecoder struct {
	versions []int
}

// NewGTPDecoder creates a new GTP decoder
func NewGTPDecoder(versions []int) *GTPDecoder {
	return &GTPDecoder{
		versions: versions,
	}
}

// Protocol returns the protocol type (will be refined based on version)
func (d *GTPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolGTPv2C // Default, will be set during decode
}

// CanDecode checks if the data is a GTP-C message
func (d *GTPDecoder) CanDecode(data []byte) bool {
	if len(data) < 8 {
		return false
	}

	// Check GTP version in first 3 bits
	version := (data[0] >> 5) & 0x07

	// GTPv1-C or GTPv2-C
	return version == 1 || version == 2
}

// Decode decodes a GTP-C message
func (d *GTPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 8 {
		return nil, decoder.ErrInsufficientData
	}

	version := (data[0] >> 5) & 0x07

	var msg *decoder.Message
	var err error

	if version == 1 {
		msg, err = d.decodeGTPv1(data, metadata)
	} else if version == 2 {
		msg, err = d.decodeGTPv2(data, metadata)
	} else {
		return nil, decoder.ErrUnsupportedVersion
	}

	if err != nil {
		return nil, err
	}

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// decodeGTPv1 handles GTPv1-C messages
func (d *GTPDecoder) decodeGTPv1(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolGTPv1C,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse GTPv1 header
	flags := data[0]
	msgType := data[1]
	length := binary.BigEndian.Uint16(data[2:4])
	teid := binary.BigEndian.Uint32(data[4:8])

	msg.MessageType = fmt.Sprintf("GTPv1_%s", getGTPv1MessageName(msgType))
	msg.MessageName = getGTPv1MessageName(msgType)
	msg.TEID = teid
	msg.Details["version"] = 1
	msg.Details["flags"] = flags
	msg.Details["length"] = length

	// Sequence number (if present)
	headerLen := 8
	if flags&0x02 != 0 { // Sequence number flag
		if len(data) >= 10 {
			seqNum := binary.BigEndian.Uint16(data[8:10])
			msg.SequenceNum = uint32(seqNum)
			headerLen = 12
		}
	}

	// Determine direction and result
	if msgType%2 == 0 { // Even = response
		msg.Direction = decoder.DirectionResponse
	} else { // Odd = request
		msg.Direction = decoder.DirectionRequest
	}

	// Parse IEs (Information Elements)
	if len(data) > headerLen {
		ies, cause := d.parseGTPv1IEs(data[headerLen:])
		msg.Details["ies"] = ies

		// Extract correlation fields
		d.extractGTPv1CorrelationFields(msg, ies)

		// Set result based on cause
		if cause != 0 {
			msg.CauseCode = cause
			if cause == 128 { // Request accepted
				msg.Result = decoder.ResultSuccess
			} else {
				msg.Result = decoder.ResultFailure
				msg.CauseText = getGTPv1CauseText(cause)
			}
		}
	}

	// Identify network elements
	d.identifyGTPv1NetworkElements(msg, msgType)

	return msg, nil
}

// decodeGTPv2 handles GTPv2-C messages
func (d *GTPDecoder) decodeGTPv2(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolGTPv2C,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse GTPv2 header (minimum 8 bytes)
	flags := data[0]
	msgType := data[1]
	length := binary.BigEndian.Uint16(data[2:4])

	msg.MessageType = fmt.Sprintf("GTPv2_%s", getGTPv2MessageName(msgType))
	msg.MessageName = getGTPv2MessageName(msgType)
	msg.Details["version"] = 2
	msg.Details["flags"] = flags
	msg.Details["length"] = length

	// TEID flag
	headerLen := 8
	if flags&0x08 != 0 { // TEID present
		if len(data) >= 12 {
			teid := binary.BigEndian.Uint32(data[4:8])
			msg.TEID = teid
			seqNum := binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
			msg.SequenceNum = seqNum
			headerLen = 12
		}
	} else {
		if len(data) >= 12 {
			seqNum := binary.BigEndian.Uint32(data[4:8]) & 0x00FFFFFF
			msg.SequenceNum = seqNum
		}
	}

	// Determine direction
	if msgType%2 == 0 { // Even = response
		msg.Direction = decoder.DirectionResponse
	} else { // Odd = request
		msg.Direction = decoder.DirectionRequest
	}

	// Parse IEs
	if len(data) > headerLen {
		ies, cause := d.parseGTPv2IEs(data[headerLen:])
		msg.Details["ies"] = ies

		// Extract correlation fields
		d.extractGTPv2CorrelationFields(msg, ies)

		// Set result based on cause
		if cause != 0 {
			msg.CauseCode = cause
			if cause == 16 { // Request accepted
				msg.Result = decoder.ResultSuccess
			} else {
				msg.Result = decoder.ResultFailure
				msg.CauseText = getGTPv2CauseText(cause)
			}
		}
	}

	// Identify network elements
	d.identifyGTPv2NetworkElements(msg, msgType)

	return msg, nil
}

// parseGTPv1IEs parses GTPv1 Information Elements
func (d *GTPDecoder) parseGTPv1IEs(data []byte) (map[string]interface{}, int) {
	ies := make(map[string]interface{})
	cause := 0
	offset := 0

	for offset < len(data)-3 {
		ieType := data[offset]
		ieLen := int(binary.BigEndian.Uint16(data[offset+1 : offset+3]))

		if offset+3+ieLen > len(data) {
			break
		}

		ieData := data[offset+3 : offset+3+ieLen]
		ieName := getGTPv1IEName(ieType)

		// Parse specific IEs
		switch ieType {
		case 1: // Cause
			if len(ieData) > 0 {
				cause = int(ieData[0])
				ies[ieName] = cause
			}
		case 2: // IMSI
			ies[ieName] = decodeBCD(ieData)
		case 14: // Recovery
			ies[ieName] = ieData[0]
		default:
			ies[ieName] = ieData
		}

		offset += 3 + ieLen
	}

	return ies, cause
}

// parseGTPv2IEs parses GTPv2 Information Elements
func (d *GTPDecoder) parseGTPv2IEs(data []byte) (map[string]interface{}, int) {
	ies := make(map[string]interface{})
	cause := 0
	offset := 0

	for offset < len(data)-4 {
		ieType := data[offset]
		ieLen := int(binary.BigEndian.Uint16(data[offset+1 : offset+3]))
		// instance := data[offset+3] & 0x0F

		if offset+4+ieLen > len(data) {
			break
		}

		ieData := data[offset+4 : offset+4+ieLen]
		ieName := getGTPv2IEName(ieType)

		// Parse specific IEs
		switch ieType {
		case 2: // Cause
			if len(ieData) > 0 {
				cause = int(ieData[0])
				ies[ieName] = cause
			}
		case 1: // IMSI
			ies[ieName] = decodeBCD(ieData)
		case 71: // APN
			ies[ieName] = decodeAPN(ieData)
		case 75: // MSISDN
			ies[ieName] = decodeBCD(ieData)
		case 87: // F-TEID
			if len(ieData) >= 9 {
				teid := binary.BigEndian.Uint32(ieData[1:5])
				ies[ieName] = map[string]interface{}{
					"teid": teid,
					"ipv4": fmt.Sprintf("%d.%d.%d.%d", ieData[5], ieData[6], ieData[7], ieData[8]),
				}
			}
		default:
			ies[ieName] = ieData
		}

		offset += 4 + ieLen
	}

	return ies, cause
}

// extractGTPv1CorrelationFields extracts correlation fields from IEs
func (d *GTPDecoder) extractGTPv1CorrelationFields(msg *decoder.Message, ies map[string]interface{}) {
	if imsi, ok := ies["IMSI"].(string); ok {
		msg.IMSI = imsi
	}
}

// extractGTPv2CorrelationFields extracts correlation fields from IEs
func (d *GTPDecoder) extractGTPv2CorrelationFields(msg *decoder.Message, ies map[string]interface{}) {
	if imsi, ok := ies["IMSI"].(string); ok {
		msg.IMSI = imsi
	}
	if msisdn, ok := ies["MSISDN"].(string); ok {
		msg.MSISDN = msisdn
	}
	if apn, ok := ies["APN"].(string); ok {
		msg.APN = apn
	}
}

// identifyGTPv1NetworkElements identifies network elements for GTPv1
func (d *GTPDecoder) identifyGTPv1NetworkElements(msg *decoder.Message, msgType uint8) {
	switch msgType {
	case 16, 17: // Create PDP Context Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SGSN"
			msg.Destination.Type = "GGSN"
		} else {
			msg.Source.Type = "GGSN"
			msg.Destination.Type = "SGSN"
		}
	case 32, 33: // Create Session Request/Response (on S5/S8)
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "SGW"
			msg.Destination.Type = "PGW"
		} else {
			msg.Source.Type = "PGW"
			msg.Destination.Type = "SGW"
		}
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// identifyGTPv2NetworkElements identifies network elements for GTPv2
func (d *GTPDecoder) identifyGTPv2NetworkElements(msg *decoder.Message, msgType uint8) {
	switch msgType {
	case 32, 33: // Create Session Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "MME"
			msg.Destination.Type = "SGW"
		} else {
			msg.Source.Type = "SGW"
			msg.Destination.Type = "MME"
		}
	case 34, 35: // Modify Bearer Request/Response
		if msg.Direction == decoder.DirectionRequest {
			msg.Source.Type = "MME"
			msg.Destination.Type = "SGW"
		} else {
			msg.Source.Type = "SGW"
			msg.Destination.Type = "MME"
		}
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// Helper functions for message and IE names
func getGTPv1MessageName(msgType uint8) string {
	messages := map[uint8]string{
		16: "CreatePDPContextRequest",
		17: "CreatePDPContextResponse",
		18: "UpdatePDPContextRequest",
		19: "UpdatePDPContextResponse",
		20: "DeletePDPContextRequest",
		21: "DeletePDPContextResponse",
	}
	if name, ok := messages[msgType]; ok {
		return name
	}
	return fmt.Sprintf("GTPv1_Type_%d", msgType)
}

func getGTPv2MessageName(msgType uint8) string {
	messages := map[uint8]string{
		32: "CreateSessionRequest",
		33: "CreateSessionResponse",
		34: "ModifyBearerRequest",
		35: "ModifyBearerResponse",
		36: "DeleteSessionRequest",
		37: "DeleteSessionResponse",
	}
	if name, ok := messages[msgType]; ok {
		return name
	}
	return fmt.Sprintf("GTPv2_Type_%d", msgType)
}

func getGTPv1IEName(ieType uint8) string {
	ies := map[uint8]string{
		1:  "Cause",
		2:  "IMSI",
		14: "Recovery",
		16: "TEID-Data-I",
		127: "ChargingID",
	}
	if name, ok := ies[ieType]; ok {
		return name
	}
	return fmt.Sprintf("IE_%d", ieType)
}

func getGTPv2IEName(ieType uint8) string {
	ies := map[uint8]string{
		1:  "IMSI",
		2:  "Cause",
		71: "APN",
		74: "AMBR",
		75: "MSISDN",
		87: "F-TEID",
		93: "Bearer-Context",
	}
	if name, ok := ies[ieType]; ok {
		return name
	}
	return fmt.Sprintf("IE_%d", ieType)
}

func getGTPv1CauseText(cause int) string {
	causes := map[int]string{
		128: "Request Accepted",
		192: "Non-existent",
		193: "Invalid Message Format",
		194: "IMSI Not Known",
		195: "MS is GPRS Detached",
		199: "No Resources Available",
	}
	if text, ok := causes[cause]; ok {
		return text
	}
	return fmt.Sprintf("Cause_%d", cause)
}

func getGTPv2CauseText(cause int) string {
	causes := map[int]string{
		16: "Request Accepted",
		64: "Context Not Found",
		65: "Invalid Message Format",
		66: "Version Not Supported",
		72: "Semantic Error in TFT",
		73: "Syntactic Error in TFT",
	}
	if text, ok := causes[cause]; ok {
		return text
	}
	return fmt.Sprintf("Cause_%d", cause)
}

// decodeBCD decodes BCD (Binary Coded Decimal) to string
func decodeBCD(data []byte) string {
	result := ""
	for _, b := range data {
		low := b & 0x0F
		high := (b >> 4) & 0x0F
		if low <= 9 {
			result += string('0' + low)
		}
		if high <= 9 && high != 0x0F {
			result += string('0' + high)
		}
	}
	return result
}

// decodeAPN decodes APN from encoded format
func decodeAPN(data []byte) string {
	apn := ""
	offset := 0
	for offset < len(data) {
		labelLen := int(data[offset])
		if labelLen == 0 || offset+1+labelLen > len(data) {
			break
		}
		if apn != "" {
			apn += "."
		}
		apn += string(data[offset+1 : offset+1+labelLen])
		offset += 1 + labelLen
	}
	return apn
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
