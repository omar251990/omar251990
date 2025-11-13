package s1ap

import (
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// S1APDecoder handles 4G S1 Application Protocol
type S1APDecoder struct{}

// NewS1APDecoder creates a new S1AP decoder
func NewS1APDecoder() *S1APDecoder {
	return &S1APDecoder{}
}

// Protocol returns the protocol type
func (d *S1APDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolS1AP
}

// CanDecode checks if the data is an S1AP message
func (d *S1APDecoder) CanDecode(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// S1AP uses SCTP, check for valid PDU choice
	return data[0] == 0x00 || data[0] == 0x20 || data[0] == 0x40
}

// Decode decodes an S1AP message
func (d *S1APDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 8 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolS1AP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse S1AP PDU choice
	pduChoice := data[0]
	var procedureCode int

	if len(data) > 2 {
		procedureCode = int(data[2])
	}

	// Determine message type and direction
	switch pduChoice {
	case 0x00: // initiatingMessage
		msg.Direction = decoder.DirectionRequest
		msg.MessageType = "S1AP_InitiatingMessage"
	case 0x20: // successfulOutcome
		msg.Direction = decoder.DirectionResponse
		msg.MessageType = "S1AP_SuccessfulOutcome"
		msg.Result = decoder.ResultSuccess
	case 0x40: // unsuccessfulOutcome
		msg.Direction = decoder.DirectionResponse
		msg.MessageType = "S1AP_UnsuccessfulOutcome"
		msg.Result = decoder.ResultFailure
	}

	msg.MessageName = getS1APProcedureName(procedureCode)
	msg.Details["procedure_code"] = procedureCode

	// Parse IEs
	ies := d.parseIEs(data)
	msg.Details["ies"] = ies

	// Extract correlation fields
	d.extractCorrelationFields(msg, ies)

	// Identify network elements
	d.identifyNetworkElements(msg, procedureCode)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// parseIEs parses S1AP Information Elements
func (d *S1APDecoder) parseIEs(data []byte) map[string]interface{} {
	ies := make(map[string]interface{})

	// Simplified IE parsing
	// In production, would need full ASN.1 PER decoder
	if len(data) > 10 {
		ies["parsed"] = "basic"
	}

	return ies
}

// extractCorrelationFields extracts key fields
func (d *S1APDecoder) extractCorrelationFields(msg *decoder.Message, ies map[string]interface{}) {
	// Extract from IEs:
	// - MME-UE-S1AP-ID
	// - ENB-UE-S1AP-ID
	// - GUMMEI
	// - IMSI
	// - E-RABs
}

// identifyNetworkElements identifies source and destination
func (d *S1APDecoder) identifyNetworkElements(msg *decoder.Message, procedureCode int) {
	switch procedureCode {
	case 12: // InitialUEMessage
		msg.Source.Type = "eNB"
		msg.Destination.Type = "MME"
	case 13: // UplinkNASTransport
		msg.Source.Type = "eNB"
		msg.Destination.Type = "MME"
	case 11: // DownlinkNASTransport
		msg.Source.Type = "MME"
		msg.Destination.Type = "eNB"
	case 9: // InitialContextSetup
		msg.Source.Type = "MME"
		msg.Destination.Type = "eNB"
	case 5: // E-RABSetup
		msg.Source.Type = "MME"
		msg.Destination.Type = "eNB"
	case 0: // HandoverPreparation
		msg.Source.Type = "eNB"
		msg.Destination.Type = "MME"
	case 1: // HandoverResourceAllocation
		msg.Source.Type = "MME"
		msg.Destination.Type = "eNB"
	case 23: // UEContextRelease
		msg.Source.Type = "MME"
		msg.Destination.Type = "eNB"
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// getS1APProcedureName returns procedure name for code
func getS1APProcedureName(code int) string {
	procedures := map[int]string{
		0:  "HandoverPreparation",
		1:  "HandoverResourceAllocation",
		2:  "HandoverNotification",
		3:  "PathSwitchRequest",
		4:  "HandoverCancel",
		5:  "E-RABSetup",
		6:  "E-RABModify",
		7:  "E-RABRelease",
		8:  "E-RABReleaseIndication",
		9:  "InitialContextSetup",
		10: "Paging",
		11: "DownlinkNASTransport",
		12: "InitialUEMessage",
		13: "UplinkNASTransport",
		14: "Reset",
		15: "ErrorIndication",
		16: "NASNonDeliveryIndication",
		17: "S1Setup",
		18: "UEContextReleaseRequest",
		19: "DownlinkS1cdma2000tunnelling",
		20: "UplinkS1cdma2000tunnelling",
		21: "UEContextModification",
		22: "UECapabilityInfoIndication",
		23: "UEContextRelease",
		24: "eNBStatusTransfer",
		25: "MMEStatusTransfer",
		26: "DeactivateTrace",
		27: "TraceStart",
		28: "TraceFailureIndication",
		29: "ENBConfigurationUpdate",
		30: "MMEConfigurationUpdate",
		31: "LocationReportingControl",
		32: "LocationReportingFailureIndication",
		33: "LocationReport",
		34: "OverloadStart",
		35: "OverloadStop",
		36: "WriteReplaceWarning",
		37: "eNBDirectInformationTransfer",
		38: "MMEDirectInformationTransfer",
		39: "PrivateMessage",
		40: "eNBConfigurationTransfer",
		41: "MMEConfigurationTransfer",
		42: "CellTrafficTrace",
		43: "Kill",
		44: "DownlinkUEAssociatedLPPaTransport",
		45: "UplinkUEAssociatedLPPaTransport",
		46: "DownlinkNonUEAssociatedLPPaTransport",
		47: "UplinkNonUEAssociatedLPPaTransport",
	}

	if name, ok := procedures[code]; ok {
		return name
	}
	return fmt.Sprintf("S1AP_Procedure_%d", code)
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
