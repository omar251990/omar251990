package inap

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// INAPDecoder handles Intelligent Network Application Part protocol
type INAPDecoder struct {
	version []int // INAP CS-1, CS-2, CS-3
}

// NewINAPDecoder creates a new INAP decoder
func NewINAPDecoder(versions []int) *INAPDecoder {
	return &INAPDecoder{
		version: versions,
	}
}

// Protocol returns the protocol type
func (d *INAPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolINAP
}

// CanDecode checks if the data is an INAP message
func (d *INAPDecoder) CanDecode(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// INAP uses TCAP, check for TCAP tags
	tag := data[0]
	return tag == 0x62 || tag == 0x65 || tag == 0x64 || tag == 0x67
}

// Decode decodes an INAP message
func (d *INAPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 10 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolINAP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort, Type: "SSP"},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort, Type: "SCP"},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse TCAP layer
	tcapType := data[0]
	switch tcapType {
	case 0x62:
		msg.MessageType = "INAP_Begin"
		msg.Direction = decoder.DirectionRequest
	case 0x65:
		msg.MessageType = "INAP_Continue"
		msg.Direction = decoder.DirectionRequest
	case 0x64:
		msg.MessageType = "INAP_End"
		msg.Direction = decoder.DirectionResponse
	case 0x67:
		msg.MessageType = "INAP_Abort"
		msg.Direction = decoder.DirectionResponse
		msg.Result = decoder.ResultFailure
	}

	// Parse INAP operation
	operation, err := d.parseOperation(data)
	if err == nil {
		msg.MessageName = operation.Name
		msg.Details["operation_code"] = operation.Code
		msg.Details["inap_version"] = operation.Version
		msg.Details["service_key"] = operation.ServiceKey

		// Extract identifiers
		if operation.CallingParty != "" {
			msg.MSISDN = operation.CallingParty
		}
		if operation.CalledParty != "" {
			msg.Details["called_party"] = operation.CalledParty
		}

		// Determine result
		if operation.ErrorCode != 0 {
			msg.Result = decoder.ResultFailure
			msg.CauseCode = operation.ErrorCode
			msg.CauseText = getINAPErrorText(operation.ErrorCode)
		} else if tcapType == 0x64 {
			msg.Result = decoder.ResultSuccess
		}

		msg.Details["parameters"] = operation.Parameters
	}

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// INAPOperation represents a decoded INAP operation
type INAPOperation struct {
	Code          int
	Name          string
	Version       int // CS-1, CS-2, CS-3
	ServiceKey    int
	CallingParty  string
	CalledParty   string
	ErrorCode     int
	Parameters    map[string]interface{}
}

// parseOperation extracts INAP operation details
func (d *INAPDecoder) parseOperation(data []byte) (*INAPOperation, error) {
	op := &INAPOperation{
		Parameters: make(map[string]interface{}),
	}

	// Look for invoke component
	for i := 0; i < len(data)-5; i++ {
		if data[i] == 0xa1 { // Invoke tag
			// Operation code
			if i+4 < len(data) && data[i+2] == 0x02 {
				op.Code = int(data[i+4])
				op.Name = getINAPOperationName(op.Code)
				op.Version = getINAPVersion(op.Code)

				// Extract service key
				serviceKey := d.extractServiceKey(data[i:])
				if serviceKey != -1 {
					op.ServiceKey = serviceKey
				}

				// Extract calling party
				callingParty := d.extractCallingParty(data[i:])
				if callingParty != "" {
					op.CallingParty = callingParty
				}

				// Extract called party
				calledParty := d.extractCalledParty(data[i:])
				if calledParty != "" {
					op.CalledParty = calledParty
				}

				break
			}
		}
	}

	return op, nil
}

// extractServiceKey extracts the service key from INAP data
func (d *INAPDecoder) extractServiceKey(data []byte) int {
	for i := 0; i < len(data)-3; i++ {
		if data[i] == 0x80 && data[i+1] <= 4 {
			length := int(data[i+1])
			if i+2+length <= len(data) {
				var key int
				for j := 0; j < length; j++ {
					key = (key << 8) | int(data[i+2+j])
				}
				return key
			}
		}
	}
	return -1
}

// extractCallingParty extracts calling party number
func (d *INAPDecoder) extractCallingParty(data []byte) string {
	for i := 0; i < len(data)-10; i++ {
		if data[i] == 0x81 && data[i+1] >= 6 && data[i+1] <= 15 {
			return decodeBCD(data[i+2 : i+2+int(data[i+1])])
		}
	}
	return ""
}

// extractCalledParty extracts called party number
func (d *INAPDecoder) extractCalledParty(data []byte) string {
	for i := 0; i < len(data)-10; i++ {
		if data[i] == 0x82 && data[i+1] >= 6 && data[i+1] <= 15 {
			return decodeBCD(data[i+2 : i+2+int(data[i+1])])
		}
	}
	return ""
}

// getINAPOperationName returns the operation name for a code
func getINAPOperationName(code int) string {
	operations := map[int]string{
		// INAP CS-1
		0:  "InitialDP",
		1:  "OriginationAttemptAuthorized",
		2:  "CollectedInformation",
		3:  "AnalyzedInformation",
		4:  "RouteSelectFailure",
		5:  "oCalledPartyBusy",
		6:  "oNoAnswer",
		7:  "oAnswer",
		8:  "oMidCall",
		9:  "oDisconnect",
		10: "oAbandon",
		11: "TermAttemptAuthorized",
		12: "tBusy",
		13: "tNoAnswer",
		14: "tAnswer",
		15: "tMidCall",
		16: "tDisconnect",
		17: "tAbandon",
		// INAP CS-2
		18: "Connect",
		19: "ConnectToResource",
		20: "EstablishTemporaryConnection",
		21: "DisconnectForwardConnection",
		22: "ContinueWithArgument",
		23: "ReleaseCall",
		24: "RequestReportBCSMEvent",
		25: "EventReportBCSM",
		// INAP CS-3
		27: "PlayAnnouncement",
		28: "PromptAndCollectUserInformation",
		29: "SpecializedResourceReport",
		30: "Cancel",
		31: "ActivityTest",
		32: "ServiceFilteringResponse",
		33: "CallGap",
		34: "CallInformationRequest",
		35: "CallInformationReport",
	}

	if name, ok := operations[code]; ok {
		return name
	}
	return fmt.Sprintf("INAP_Unknown_%d", code)
}

// getINAPVersion determines INAP version from operation code
func getINAPVersion(code int) int {
	if code >= 0 && code <= 17 {
		return 1 // CS-1
	} else if code >= 18 && code <= 26 {
		return 2 // CS-2
	} else if code >= 27 {
		return 3 // CS-3
	}
	return 1
}

// getINAPErrorText returns error description
func getINAPErrorText(code int) string {
	errors := map[int]string{
		0:  "Canceled",
		1:  "CancelFailed",
		2:  "ETCFailed",
		3:  "ImproperCallerResponse",
		4:  "MissingCustomerRecord",
		5:  "MissingParameter",
		6:  "ParameterOutOfRange",
		7:  "RequestedInfoError",
		8:  "SystemFailure",
		9:  "TaskRefused",
		10: "UnavailableResource",
		11: "UnexpectedComponentSequence",
		12: "UnexpectedDataValue",
		13: "UnexpectedParameter",
		14: "UnknownLegID",
		15: "UnknownCSID",
	}

	if text, ok := errors[code]; ok {
		return text
	}
	return fmt.Sprintf("INAP_Error_%d", code)
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

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
