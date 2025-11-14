package cap

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// CAPDecoder handles CAMEL Application Part protocol
type CAPDecoder struct {
	version []int // CAP phases 1-4
}

// NewCAPDecoder creates a new CAP decoder
func NewCAPDecoder(versions []int) *CAPDecoder {
	return &CAPDecoder{
		version: versions,
	}
}

// Protocol returns the protocol type
func (d *CAPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolCAP
}

// CanDecode checks if the data is a CAP message
func (d *CAPDecoder) CanDecode(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// CAP uses TCAP, check for TCAP tags
	tag := data[0]
	return tag == 0x62 || tag == 0x65 || tag == 0x64 || tag == 0x67
}

// Decode decodes a CAP message
func (d *CAPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 10 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolCAP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort, Type: "MSC/SSF"},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort, Type: "gsmSCF"},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse TCAP layer
	tcapType := data[0]
	switch tcapType {
	case 0x62:
		msg.MessageType = "CAP_Begin"
		msg.Direction = decoder.DirectionRequest
	case 0x65:
		msg.MessageType = "CAP_Continue"
		msg.Direction = decoder.DirectionRequest
	case 0x64:
		msg.MessageType = "CAP_End"
		msg.Direction = decoder.DirectionResponse
	case 0x67:
		msg.MessageType = "CAP_Abort"
		msg.Direction = decoder.DirectionResponse
		msg.Result = decoder.ResultFailure
	}

	// Parse CAP operation
	operation, err := d.parseOperation(data)
	if err == nil {
		msg.MessageName = operation.Name
		msg.Details["operation_code"] = operation.Code
		msg.Details["cap_phase"] = operation.Phase
		msg.Details["service_key"] = operation.ServiceKey

		// Extract subscriber identifiers
		if operation.IMSI != "" {
			msg.IMSI = operation.IMSI
		}
		if operation.MSISDN != "" {
			msg.MSISDN = operation.MSISDN
		}

		// Determine result
		if operation.ErrorCode != 0 {
			msg.Result = decoder.ResultFailure
			msg.CauseCode = operation.ErrorCode
			msg.CauseText = getCAPErrorText(operation.ErrorCode)
		} else if tcapType == 0x64 {
			msg.Result = decoder.ResultSuccess
		}

		msg.Details["parameters"] = operation.Parameters
	}

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// CAPOperation represents a decoded CAP operation
type CAPOperation struct {
	Code       int
	Name       string
	Phase      int // CAP phase 1-4
	ServiceKey int
	IMSI       string
	MSISDN     string
	ErrorCode  int
	Parameters map[string]interface{}
}

// parseOperation extracts CAP operation details
func (d *CAPDecoder) parseOperation(data []byte) (*CAPOperation, error) {
	op := &CAPOperation{
		Parameters: make(map[string]interface{}),
	}

	// Look for invoke component
	for i := 0; i < len(data)-5; i++ {
		if data[i] == 0xa1 { // Invoke tag
			// Operation code
			if i+4 < len(data) && data[i+2] == 0x02 {
				op.Code = int(data[i+4])
				op.Name = getCAPOperationName(op.Code)
				op.Phase = getCAPPhase(op.Code)

				// Extract service key if present
				serviceKey := d.extractServiceKey(data[i:])
				if serviceKey != -1 {
					op.ServiceKey = serviceKey
				}

				// Extract IMSI
				imsi := d.extractIMSI(data[i:])
				if imsi != "" {
					op.IMSI = imsi
				}

				// Extract MSISDN
				msisdn := d.extractMSISDN(data[i:])
				if msisdn != "" {
					op.MSISDN = msisdn
				}

				break
			}
		}
	}

	return op, nil
}

// extractServiceKey extracts the service key from CAP data
func (d *CAPDecoder) extractServiceKey(data []byte) int {
	// Service key tag is typically 0x80
	for i := 0; i < len(data)-3; i++ {
		if data[i] == 0x80 && data[i+1] <= 4 {
			length := int(data[i+1])
			if i+2+length <= len(data) {
				// Convert bytes to integer
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

// extractIMSI tries to extract IMSI from CAP data
func (d *CAPDecoder) extractIMSI(data []byte) string {
	for i := 0; i < len(data)-10; i++ {
		if (data[i] == 0x80 || data[i] == 0x04) && data[i+1] >= 7 && data[i+1] <= 8 {
			imsi := decodeBCD(data[i+2 : i+2+int(data[i+1])])
			if len(imsi) == 15 {
				return imsi
			}
		}
	}
	return ""
}

// extractMSISDN tries to extract MSISDN from CAP data
func (d *CAPDecoder) extractMSISDN(data []byte) string {
	for i := 0; i < len(data)-10; i++ {
		if data[i] == 0x81 && data[i+1] >= 6 && data[i+1] <= 10 {
			msisdn := decodeBCD(data[i+2 : i+2+int(data[i+1])])
			if len(msisdn) >= 10 && len(msisdn) <= 15 {
				return msisdn
			}
		}
	}
	return ""
}

// getCAPOperationName returns the operation name for a code
func getCAPOperationName(code int) string {
	operations := map[int]string{
		// CAP Phase 1
		0:  "InitialDP",
		1:  "AssistRequestInstructions",
		2:  "EstablishTemporaryConnection",
		3:  "DisconnectForwardConnection",
		4:  "ConnectToResource",
		5:  "Connect",
		6:  "ReleaseCall",
		7:  "RequestReportBCSMEvent",
		8:  "EventReportBCSM",
		// CAP Phase 2
		9:  "CollectInformation",
		10: "Continue",
		11: "InitiateCallAttempt",
		12: "ApplyCharging",
		13: "ApplyChargingReport",
		// CAP Phase 3
		14: "CallInformationRequest",
		15: "CallInformationReport",
		16: "PlayAnnouncement",
		17: "PromptAndCollectUserInformation",
		// CAP Phase 4
		18: "SpecializedResourceReport",
		19: "Cancel",
		20: "ActivityTest",
		22: "InitialDPSMS",
		23: "FurnishChargingInformation",
		24: "ConnectSMS",
		25: "RequestReportSMSEvent",
		26: "EventReportSMS",
		27: "ContinueSMS",
		28: "ReleaseSMS",
		31: "CallGap",
		32: "ActivateServiceFiltering",
		33: "ServiceFilteringResponse",
	}

	if name, ok := operations[code]; ok {
		return name
	}
	return fmt.Sprintf("CAP_Unknown_%d", code)
}

// getCAPPhase determines CAP phase from operation code
func getCAPPhase(code int) int {
	if code >= 0 && code <= 8 {
		return 1
	} else if code >= 9 && code <= 13 {
		return 2
	} else if code >= 14 && code <= 17 {
		return 3
	} else if code >= 18 {
		return 4
	}
	return 1
}

// getCAPErrorText returns error description
func getCAPErrorText(code int) string {
	errors := map[int]string{
		0:  "Canceled",
		1:  "CancelFailed",
		3:  "RequestedInfoError",
		4:  "SystemFailure",
		5:  "TaskRefused",
		6:  "UnavailableResource",
		7:  "UnexpectedComponentSequence",
		8:  "UnexpectedDataValue",
		9:  "UnexpectedParameter",
		10: "UnknownLegID",
		11: "UnknownPDPID",
		12: "UnknownCSID",
	}

	if text, ok := errors[code]; ok {
		return text
	}
	return fmt.Sprintf("CAP_Error_%d", code)
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
