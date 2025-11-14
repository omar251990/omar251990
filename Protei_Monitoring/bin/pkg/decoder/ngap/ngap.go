package ngap

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// NGAPDecoder handles 5G NG Application Protocol
type NGAPDecoder struct{}

// NewNGAPDecoder creates a new NGAP decoder
func NewNGAPDecoder() *NGAPDecoder {
	return &NGAPDecoder{}
}

// Protocol returns the protocol type
func (d *NGAPDecoder) Protocol() decoder.Protocol {
	return decoder.ProtocolNGAP
}

// CanDecode checks if the data is an NGAP message
func (d *NGAPDecoder) CanDecode(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// NGAP uses SCTP, check for valid procedure code range
	// First byte is typically 0x00 for initiating message
	return data[0] == 0x00 || data[0] == 0x20 || data[0] == 0x40
}

// Decode decodes an NGAP message
func (d *NGAPDecoder) Decode(data []byte, metadata *decoder.Metadata) (*decoder.Message, error) {
	startTime := time.Now()

	if len(data) < 8 {
		return nil, decoder.ErrInsufficientData
	}

	msg := &decoder.Message{
		ID:          generateMessageID(),
		Timestamp:   metadata.CaptureTime,
		Protocol:    decoder.ProtocolNGAP,
		Details:     make(map[string]interface{}),
		Source:      decoder.NetworkElement{IP: metadata.SourceIP, Port: metadata.SourcePort},
		Destination: decoder.NetworkElement{IP: metadata.DestIP, Port: metadata.DestPort},
		RawPayload:  data,
		PayloadSize: len(data),
	}

	// Parse NGAP PDU choice
	pduChoice := data[0]
	var procedureCode int

	if len(data) > 2 {
		procedureCode = int(data[2])
	}

	// Determine message type and direction
	switch pduChoice {
	case 0x00: // initiatingMessage
		msg.Direction = decoder.DirectionRequest
		msg.MessageType = "NGAP_InitiatingMessage"
	case 0x20: // successfulOutcome
		msg.Direction = decoder.DirectionResponse
		msg.MessageType = "NGAP_SuccessfulOutcome"
		msg.Result = decoder.ResultSuccess
	case 0x40: // unsuccessfulOutcome
		msg.Direction = decoder.DirectionResponse
		msg.MessageType = "NGAP_UnsuccessfulOutcome"
		msg.Result = decoder.ResultFailure
	}

	msg.MessageName = getNGAPProcedureName(procedureCode)
	msg.Details["procedure_code"] = procedureCode

	// Extract IEs
	ies := d.parseIEs(data)
	msg.Details["ies"] = ies

	// Extract key fields
	d.extractCorrelationFields(msg, ies)

	// Identify network elements
	d.identifyNetworkElements(msg, procedureCode)

	msg.ProcessedAt = time.Now()
	msg.DecodeTimeUs = time.Since(startTime).Microseconds()

	return msg, nil
}

// parseIEs parses NGAP Information Elements
func (d *NGAPDecoder) parseIEs(data []byte) map[string]interface{} {
	ies := make(map[string]interface{})

	// Simplified IE parsing
	if len(data) > 10 {
		// Look for common IEs
		// AMF UE NGAP ID (id = 10)
		// RAN UE NGAP ID (id = 85)
		// NAS-PDU (id = 38)
		// These would require full ASN.1 PER decoder in production
		ies["parsed"] = "basic"
	}

	return ies
}

// extractCorrelationFields extracts key fields for correlation
func (d *NGAPDecoder) extractCorrelationFields(msg *decoder.Message, ies map[string]interface{}) {
	// In production, extract from actual IEs:
	// - AMF-UE-NGAP-ID
	// - RAN-UE-NGAP-ID
	// - GUAMI (AMF identifier)
	// - 5G-S-TMSI
	// - SUPI (IMSI)
}

// identifyNetworkElements identifies source and destination
func (d *NGAPDecoder) identifyNetworkElements(msg *decoder.Message, procedureCode int) {
	// Based on procedure code and direction
	switch procedureCode {
	case 21: // InitialUEMessage
		msg.Source.Type = "gNB"
		msg.Destination.Type = "AMF"
	case 46: // UplinkNASTransport
		msg.Source.Type = "gNB"
		msg.Destination.Type = "AMF"
	case 4: // DownlinkNASTransport
		msg.Source.Type = "AMF"
		msg.Destination.Type = "gNB"
	case 20: // InitialContextSetup
		msg.Source.Type = "AMF"
		msg.Destination.Type = "gNB"
	case 27: // PDUSessionResourceSetup
		msg.Source.Type = "AMF"
		msg.Destination.Type = "gNB"
	case 0: // AMFConfigurationUpdate
		msg.Source.Type = "AMF"
		msg.Destination.Type = "gNB"
	case 40: // HandoverRequired
		msg.Source.Type = "gNB"
		msg.Destination.Type = "AMF"
	case 41: // HandoverRequest
		msg.Source.Type = "AMF"
		msg.Destination.Type = "gNB"
	default:
		msg.Source.Type = "Unknown"
		msg.Destination.Type = "Unknown"
	}
}

// getNGAPProcedureName returns procedure name for code
func getNGAPProcedureName(code int) string {
	procedures := map[int]string{
		0:  "AMFConfigurationUpdate",
		1:  "AMFStatusIndication",
		2:  "CellTrafficTrace",
		3:  "DeactivateTrace",
		4:  "DownlinkNASTransport",
		5:  "DownlinkNonUEAssociatedNRPPaTransport",
		6:  "DownlinkRANConfigurationTransfer",
		7:  "DownlinkRANStatusTransfer",
		8:  "DownlinkUEAssociatedNRPPaTransport",
		9:  "ErrorIndication",
		10: "HandoverCancel",
		11: "HandoverNotification",
		12: "HandoverPreparation",
		13: "HandoverResourceAllocation",
		14: "InitialContextSetup",
		15: "InitialUEMessage",
		16: "LocationReportingControl",
		17: "LocationReportingFailureIndication",
		18: "LocationReport",
		19: "NASNonDeliveryIndication",
		20: "NGReset",
		21: "NGSetup",
		22: "OverloadStart",
		23: "OverloadStop",
		24: "Paging",
		25: "PathSwitchRequest",
		26: "PDUSessionResourceModify",
		27: "PDUSessionResourceModifyIndication",
		28: "PDUSessionResourceRelease",
		29: "PDUSessionResourceSetup",
		30: "PDUSessionResourceNotify",
		31: "PrivateMessage",
		32: "PWSCancel",
		33: "PWSFailureIndication",
		34: "PWSRestartIndication",
		35: "RANConfigurationUpdate",
		36: "RerouteNASRequest",
		37: "RRCInactiveTransitionReport",
		38: "TraceFailureIndication",
		39: "TraceStart",
		40: "UEContextModification",
		41: "UEContextRelease",
		42: "UEContextReleaseRequest",
		43: "UERadioCapabilityCheck",
		44: "UERadioCapabilityInfoIndication",
		45: "UETNLABindingRelease",
		46: "UplinkNASTransport",
		47: "UplinkNonUEAssociatedNRPPaTransport",
		48: "UplinkRANConfigurationTransfer",
		49: "UplinkRANStatusTransfer",
		50: "UplinkUEAssociatedNRPPaTransport",
		51: "WriteReplaceWarning",
	}

	if name, ok := procedures[code]; ok {
		return name
	}
	return fmt.Sprintf("NGAP_Procedure_%d", code)
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
