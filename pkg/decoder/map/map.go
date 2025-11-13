package map_decoder

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// MAPDecoder handles Mobile Application Part (MAP) protocol
type MAPDecoder struct {
	version []int
}

// NewMAPDecoder creates a new MAP decoder
func NewMAPDecoder(versions []int) *MAPDecoder {
	return &MAPDecoder{
		version: versions,
	}
}

// Protocol returns the protocol type
func (d *MAPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolMAP
}

// CanDecode checks if the data is a MAP message
func (d *MAPDecoder) CanDecode(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// Check for TCAP tag (0x62 = Begin, 0x65 = Continue, 0x64 = End, 0x67 = Abort)
	tag := data[0]
	return tag == 0x62 || tag == 0x65 || tag == 0x64 || tag == 0x67
}

// Decode decodes a MAP message
func (d *MAPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 10 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolMAP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse TCAP layer
	tcapType := data[0]
	switch tcapType {
	case 0x62:
		msg.MessageType = "TCAP_Begin"
		msg.Direction = decoder.DirectionRequest
	case 0x65:
		msg.MessageType = "TCAP_Continue"
		msg.Direction = decoder.DirectionRequest
	case 0x64:
		msg.MessageType = "TCAP_End"
		msg.Direction = decoder.DirectionResponse
	case 0x67:
		msg.MessageType = "TCAP_Abort"
		msg.Direction = decoder.DirectionResponse
		msg.Result = decoder.ResultFailure
	default:
		msg.MessageType = "TCAP_Unknown"
	}

	// Parse MAP operation
	operation, err := d.parseOperation(data)
	if err == nil {
		msg.MessageName = operation.Name
		msg.Details["operation_code"] = operation.Code
		msg.Details["operation_type"] = operation.Type

		// Extract common fields
		if operation.IMSI != "" {
			msg.IMSI = operation.IMSI
		}
		if operation.MSISDN != "" {
			msg.MSISDN = operation.MSISDN
		}
		if operation.PLMN != "" {
			msg.PLMN = operation.PLMN
		}

		// Determine result
		if operation.ErrorCode != 0 {
			msg.Result = decoder.ResultFailure
			msg.CauseCode = operation.ErrorCode
			msg.CauseText = getMAPErrorText(operation.ErrorCode)
		} else if tcapType == 0x64 {
			msg.Result = decoder.ResultSuccess
		}

		msg.Details["parameters"] = operation.Parameters
	}

	// Set network elements based on operation
	d.identifyNetworkElements(msg, operation)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// MAPOperation represents a decoded MAP operation
type MAPOperation struct {
	Code       int
	Name       string
	Type       string // mobile, supplementary, network
	IMSI       string
	MSISDN     string
	PLMN       string
	ErrorCode  int
	Parameters map[string]interface{}
}

// parseOperation extracts MAP operation details
func (d *MAPDecoder) parseOperation(data []byte) (*MAPOperation, error) {
	op := &MAPOperation{
		Parameters: make(map[string]interface{}),
	}

	// Simple operation code extraction (real implementation would use ASN.1)
	if len(data) > 20 {
		// Look for invoke tag (0xa1)
		for i := 0; i < len(data)-5; i++ {
			if data[i] == 0xa1 {
				// Operation code is typically 2 bytes after invoke tag
				if i+4 < len(data) && data[i+2] == 0x02 { // INTEGER tag
					op.Code = int(data[i+4])
					op.Name = getMAPOperationName(op.Code)
					op.Type = getMAPOperationType(op.Code)

					// Extract IMSI if present (tag 0x04 or 0x80)
					imsi := d.extractIMSI(data)
					if imsi != "" {
						op.IMSI = imsi
					}

					break
				}
			}
		}
	}

	return op, nil
}

// extractIMSI tries to extract IMSI from the data
func (d *MAPDecoder) extractIMSI(data []byte) string {
	// Look for IMSI pattern (tag 0x04 or 0x80, length 7-8 bytes)
	for i := 0; i < len(data)-10; i++ {
		if (data[i] == 0x04 || data[i] == 0x80) && data[i+1] >= 7 && data[i+1] <= 8 {
			// Decode BCD digits
			imsi := decodeBCD(data[i+2 : i+2+int(data[i+1])])
			if len(imsi) == 15 {
				return imsi
			}
		}
	}
	return ""
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
		if high <= 9 {
			result += string('0' + high)
		}
	}
	return result
}

// identifyNetworkElements sets the network element types
func (d *MAPDecoder) identifyNetworkElements(msg *decoder.Message, op *MAPOperation) {
	// Based on operation type, identify source and destination
	switch op.Type {
	case "location":
		msg.Source.Type = "VLR"
		msg.Destination.Type = "HLR"
	case "subscriber_management":
		msg.Source.Type = "HLR"
		msg.Destination.Type = "VLR"
	case "authentication":
		msg.Source.Type = "MSC"
		msg.Destination.Type = "HLR"
	case "roaming":
		msg.Source.Type = "MSC_VPLMN"
		msg.Destination.Type = "HLR_HPLMN"
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// getMAPOperationName returns the operation name for a code
func getMAPOperationName(code int) string {
	operations := map[int]string{
		2:  "UpdateLocation",
		3:  "CancelLocation",
		4:  "ProvideRoamingNumber",
		5:  "InsertSubscriberData",
		6:  "DeleteSubscriberData",
		7:  "SendParameters",
		8:  "RegisterSS",
		9:  "EraseSS",
		10: "ActivateSS",
		11: "DeactivateSS",
		12: "InterrogateSS",
		13: "ProcessUnstructuredSSRequest",
		22: "SendRoutingInfo",
		23: "UpdateGprsLocation",
		24: "SendAuthenticationInfo",
		25: "RestoreData",
		44: "SendRoutingInfoForSM",
		45: "MoForwardSM",
		46: "MtForwardSM",
		54: "AnyTimeInterrogation",
		55: "AnyTimeSubscriptionInterrogation",
		56: "AnyTimeModification",
		59: "PrepareHandover",
		68: "ProcessAccessRequest",
		70: "SendIMSI",
	}

	if name, ok := operations[code]; ok {
		return name
	}
	return fmt.Sprintf("Unknown_%d", code)
}

// getMAPOperationType returns the operation category
func getMAPOperationType(code int) string {
	if code >= 2 && code <= 7 {
		return "location"
	} else if code >= 8 && code <= 19 {
		return "supplementary_services"
	} else if code >= 20 && code <= 30 {
		return "subscriber_management"
	} else if code >= 44 && code <= 46 {
		return "sms"
	} else if code >= 54 && code <= 59 {
		return "roaming"
	}
	return "other"
}

// getMAPErrorText returns error description
func getMAPErrorText(code int) string {
	errors := map[int]string{
		1:  "Unknown Subscriber",
		3:  "Unknown MSC",
		4:  "Unidentified Subscriber",
		5:  "Absent Subscriber SM",
		6:  "Unknown Equipment",
		7:  "Roaming Not Allowed",
		8:  "Illegal Subscriber",
		9:  "Bearer Service Not Provisioned",
		10: "Teleservice Not Provisioned",
		11: "Illegal Equipment",
		12: "Call Barred",
		21: "Facility Not Supported",
		27: "Absent Subscriber",
		34: "System Failure",
		35: "Data Missing",
		36: "Unexpected Data Value",
	}

	if text, ok := errors[code]; ok {
		return text
	}
	return fmt.Sprintf("Error_%d", code)
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
