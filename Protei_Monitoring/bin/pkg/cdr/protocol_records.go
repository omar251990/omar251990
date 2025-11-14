package cdr

import (
	"fmt"
	"time"
)

// MapCDR represents a MAP protocol CDR
type MapCDR struct {
	Timestamp      time.Time
	TransactionID  string
	IMSI           string
	MSISDN         string
	OperationType  string
	OperationCode  int
	InvokeID       int
	Result         string
	ResultCode     int
	DurationMs     int64
	SCCP_Called    string
	SCCP_Calling   string
	MCC            string
	MNC            string
	LAC            string
	CellID         string
}

func (c *MapCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"MAP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.OperationType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.OperationCode),
		fmt.Sprintf("%d", c.InvokeID),
		c.SCCP_Called,
		c.SCCP_Calling,
		c.MCC,
		c.MNC,
		c.LAC,
		c.CellID,
	}
}

func (c *MapCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":      c.Timestamp.Format(time.RFC3339),
		"protocol":       "MAP",
		"transaction_id": c.TransactionID,
		"imsi":           c.IMSI,
		"msisdn":         c.MSISDN,
		"operation_type": c.OperationType,
		"operation_code": c.OperationCode,
		"invoke_id":      c.InvokeID,
		"result":         c.Result,
		"result_code":    c.ResultCode,
		"duration_ms":    c.DurationMs,
		"sccp_called":    c.SCCP_Called,
		"sccp_calling":   c.SCCP_Calling,
		"mcc":            c.MCC,
		"mnc":            c.MNC,
		"lac":            c.LAC,
		"cell_id":        c.CellID,
	}
}

func (c *MapCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *MapCDR) GetProtocol() string     { return "MAP" }

// CapCDR represents a CAP protocol CDR
type CapCDR struct {
	Timestamp      time.Time
	TransactionID  string
	IMSI           string
	MSISDN         string
	ServiceKey     int
	CallingParty   string
	CalledParty    string
	OperationType  string
	EventType      string
	Result         string
	ResultCode     int
	DurationMs     int64
	SCCP_Called    string
	SCCP_Calling   string
}

func (c *CapCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"CAP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.OperationType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.ServiceKey),
		c.CallingParty,
		c.CalledParty,
		c.SCCP_Called,
		c.SCCP_Calling,
		c.EventType,
	}
}

func (c *CapCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":      c.Timestamp.Format(time.RFC3339),
		"protocol":       "CAP",
		"transaction_id": c.TransactionID,
		"imsi":           c.IMSI,
		"msisdn":         c.MSISDN,
		"service_key":    c.ServiceKey,
		"calling_party":  c.CallingParty,
		"called_party":   c.CalledParty,
		"operation_type": c.OperationType,
		"event_type":     c.EventType,
		"result":         c.Result,
		"result_code":    c.ResultCode,
		"duration_ms":    c.DurationMs,
		"sccp_called":    c.SCCP_Called,
		"sccp_calling":   c.SCCP_Calling,
	}
}

func (c *CapCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *CapCDR) GetProtocol() string     { return "CAP" }

// DiameterCDR represents a Diameter protocol CDR
type DiameterCDR struct {
	Timestamp        time.Time
	TransactionID    string
	IMSI             string
	MSISDN           string
	SessionID        string
	CommandCode      int
	ApplicationID    int
	OperationType    string
	Result           string
	ResultCode       int
	DurationMs       int64
	OriginHost       string
	OriginRealm      string
	DestinationHost  string
	DestinationRealm string
}

func (c *DiameterCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"Diameter",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.OperationType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.CommandCode),
		fmt.Sprintf("%d", c.ApplicationID),
		c.SessionID,
		c.OriginHost,
		c.OriginRealm,
		c.DestinationHost,
		c.DestinationRealm,
	}
}

func (c *DiameterCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":         c.Timestamp.Format(time.RFC3339),
		"protocol":          "Diameter",
		"transaction_id":    c.TransactionID,
		"imsi":              c.IMSI,
		"msisdn":            c.MSISDN,
		"session_id":        c.SessionID,
		"command_code":      c.CommandCode,
		"application_id":    c.ApplicationID,
		"operation_type":    c.OperationType,
		"result":            c.Result,
		"result_code":       c.ResultCode,
		"duration_ms":       c.DurationMs,
		"origin_host":       c.OriginHost,
		"origin_realm":      c.OriginRealm,
		"destination_host":  c.DestinationHost,
		"destination_realm": c.DestinationRealm,
	}
}

func (c *DiameterCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *DiameterCDR) GetProtocol() string     { return "Diameter" }

// GtpCDR represents a GTP protocol CDR
type GtpCDR struct {
	Timestamp      time.Time
	TransactionID  string
	IMSI           string
	MSISDN         string
	MessageType    string
	TEID           uint32
	SequenceNumber uint32
	APN            string
	PDN_Type       string
	Result         string
	ResultCode     int
	DurationMs     int64
	SourceIP       string
	DestIP         string
	BytesUplink    uint64
	BytesDownlink  uint64
}

func (c *GtpCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"GTP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.MessageType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("0x%08X", c.TEID),
		fmt.Sprintf("%d", c.SequenceNumber),
		c.APN,
		c.PDN_Type,
		c.SourceIP,
		c.DestIP,
		fmt.Sprintf("%d", c.BytesUplink),
		fmt.Sprintf("%d", c.BytesDownlink),
	}
}

func (c *GtpCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":       c.Timestamp.Format(time.RFC3339),
		"protocol":        "GTP",
		"transaction_id":  c.TransactionID,
		"imsi":            c.IMSI,
		"msisdn":          c.MSISDN,
		"message_type":    c.MessageType,
		"teid":            c.TEID,
		"sequence_number": c.SequenceNumber,
		"apn":             c.APN,
		"pdn_type":        c.PDN_Type,
		"result":          c.Result,
		"result_code":     c.ResultCode,
		"duration_ms":     c.DurationMs,
		"source_ip":       c.SourceIP,
		"dest_ip":         c.DestIP,
		"bytes_uplink":    c.BytesUplink,
		"bytes_downlink":  c.BytesDownlink,
	}
}

func (c *GtpCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *GtpCDR) GetProtocol() string     { return "GTP" }

// PfcpCDR represents a PFCP protocol CDR
type PfcpCDR struct {
	Timestamp     time.Time
	TransactionID string
	IMSI          string
	MSISDN        string
	MessageType   string
	SEID          uint64
	NodeID        string
	FSE_ID        uint32
	Result        string
	ResultCode    int
	DurationMs    int64
	UplinkBytes   uint64
	DownlinkBytes uint64
}

func (c *PfcpCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"PFCP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.MessageType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("0x%016X", c.SEID),
		c.NodeID,
		fmt.Sprintf("%d", c.FSE_ID),
		fmt.Sprintf("%d", c.UplinkBytes),
		fmt.Sprintf("%d", c.DownlinkBytes),
	}
}

func (c *PfcpCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":       c.Timestamp.Format(time.RFC3339),
		"protocol":        "PFCP",
		"transaction_id":  c.TransactionID,
		"imsi":            c.IMSI,
		"msisdn":          c.MSISDN,
		"message_type":    c.MessageType,
		"seid":            c.SEID,
		"node_id":         c.NodeID,
		"fse_id":          c.FSE_ID,
		"result":          c.Result,
		"result_code":     c.ResultCode,
		"duration_ms":     c.DurationMs,
		"uplink_bytes":    c.UplinkBytes,
		"downlink_bytes":  c.DownlinkBytes,
	}
}

func (c *PfcpCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *PfcpCDR) GetProtocol() string     { return "PFCP" }

// Http2CDR represents an HTTP/2 protocol CDR (5G SBA)
type Http2CDR struct {
	Timestamp      time.Time
	TransactionID  string
	IMSI           string
	MSISDN         string
	Method         string
	URI            string
	StatusCode     int
	ServiceName    string
	APIVersion     string
	SourceNF       string
	TargetNF       string
	Result         string
	DurationMs     int64
}

func (c *Http2CDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"HTTP2",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.Method,
		c.Result,
		fmt.Sprintf("%d", c.StatusCode),
		fmt.Sprintf("%d", c.DurationMs),
		c.URI,
		c.ServiceName,
		c.APIVersion,
		c.SourceNF,
		c.TargetNF,
	}
}

func (c *Http2CDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":      c.Timestamp.Format(time.RFC3339),
		"protocol":       "HTTP2",
		"transaction_id": c.TransactionID,
		"imsi":           c.IMSI,
		"msisdn":         c.MSISDN,
		"method":         c.Method,
		"uri":            c.URI,
		"status_code":    c.StatusCode,
		"service_name":   c.ServiceName,
		"api_version":    c.APIVersion,
		"source_nf":      c.SourceNF,
		"target_nf":      c.TargetNF,
		"result":         c.Result,
		"duration_ms":    c.DurationMs,
	}
}

func (c *Http2CDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *Http2CDR) GetProtocol() string     { return "HTTP2" }

// NgapCDR represents an NGAP protocol CDR (5G RAN)
type NgapCDR struct {
	Timestamp     time.Time
	TransactionID string
	IMSI          string
	MSISDN        string
	ProcedureCode int
	AMF_UE_ID     uint64
	RAN_UE_ID     uint32
	GlobalRAN_ID  string
	GUAMI         string
	Cause         string
	Result        string
	ResultCode    int
	DurationMs    int64
}

func (c *NgapCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"NGAP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		fmt.Sprintf("Procedure_%d", c.ProcedureCode),
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.ProcedureCode),
		fmt.Sprintf("%d", c.AMF_UE_ID),
		fmt.Sprintf("%d", c.RAN_UE_ID),
		c.GlobalRAN_ID,
		c.GUAMI,
		c.Cause,
	}
}

func (c *NgapCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":       c.Timestamp.Format(time.RFC3339),
		"protocol":        "NGAP",
		"transaction_id":  c.TransactionID,
		"imsi":            c.IMSI,
		"msisdn":          c.MSISDN,
		"procedure_code":  c.ProcedureCode,
		"amf_ue_id":       c.AMF_UE_ID,
		"ran_ue_id":       c.RAN_UE_ID,
		"global_ran_id":   c.GlobalRAN_ID,
		"guami":           c.GUAMI,
		"cause":           c.Cause,
		"result":          c.Result,
		"result_code":     c.ResultCode,
		"duration_ms":     c.DurationMs,
	}
}

func (c *NgapCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *NgapCDR) GetProtocol() string     { return "NGAP" }

// S1apCDR represents an S1AP protocol CDR (4G RAN)
type S1apCDR struct {
	Timestamp     time.Time
	TransactionID string
	IMSI          string
	MSISDN        string
	ProcedureCode int
	MME_UE_ID     uint32
	eNB_UE_ID     uint32
	TAI           string
	EUTRAN_CGI    string
	Cause         string
	Result        string
	ResultCode    int
	DurationMs    int64
}

func (c *S1apCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"S1AP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		fmt.Sprintf("Procedure_%d", c.ProcedureCode),
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.ProcedureCode),
		fmt.Sprintf("%d", c.MME_UE_ID),
		fmt.Sprintf("%d", c.eNB_UE_ID),
		c.TAI,
		c.EUTRAN_CGI,
		c.Cause,
	}
}

func (c *S1apCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":      c.Timestamp.Format(time.RFC3339),
		"protocol":       "S1AP",
		"transaction_id": c.TransactionID,
		"imsi":           c.IMSI,
		"msisdn":         c.MSISDN,
		"procedure_code": c.ProcedureCode,
		"mme_ue_id":      c.MME_UE_ID,
		"enb_ue_id":      c.eNB_UE_ID,
		"tai":            c.TAI,
		"eutran_cgi":     c.EUTRAN_CGI,
		"cause":          c.Cause,
		"result":         c.Result,
		"result_code":    c.ResultCode,
		"duration_ms":    c.DurationMs,
	}
}

func (c *S1apCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *S1apCDR) GetProtocol() string     { return "S1AP" }

// NasCDR represents a NAS protocol CDR (4G/5G mobile)
type NasCDR struct {
	Timestamp            time.Time
	TransactionID        string
	IMSI                 string
	MSISDN               string
	MessageType          string
	SecurityHeader       string
	ProtocolDiscriminator string
	EPS_MobileIdentity   string
	EMM_Cause            int
	ESM_Cause            int
	Result               string
	DurationMs           int64
}

func (c *NasCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"NAS",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.MessageType,
		c.Result,
		fmt.Sprintf("%d", c.EMM_Cause),
		fmt.Sprintf("%d", c.DurationMs),
		c.SecurityHeader,
		c.ProtocolDiscriminator,
		c.EPS_MobileIdentity,
		fmt.Sprintf("%d", c.EMM_Cause),
		fmt.Sprintf("%d", c.ESM_Cause),
	}
}

func (c *NasCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":              c.Timestamp.Format(time.RFC3339),
		"protocol":               "NAS",
		"transaction_id":         c.TransactionID,
		"imsi":                   c.IMSI,
		"msisdn":                 c.MSISDN,
		"message_type":           c.MessageType,
		"security_header":        c.SecurityHeader,
		"protocol_discriminator": c.ProtocolDiscriminator,
		"eps_mobile_identity":    c.EPS_MobileIdentity,
		"emm_cause":              c.EMM_Cause,
		"esm_cause":              c.ESM_Cause,
		"result":                 c.Result,
		"duration_ms":            c.DurationMs,
	}
}

func (c *NasCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *NasCDR) GetProtocol() string     { return "NAS" }

// InapCDR represents an INAP protocol CDR
type InapCDR struct {
	Timestamp     time.Time
	TransactionID string
	IMSI          string
	MSISDN        string
	ServiceKey    int
	CallingParty  string
	CalledParty   string
	OperationType string
	TriggerType   string
	Result        string
	ResultCode    int
	DurationMs    int64
}

func (c *InapCDR) ToCSV() []string {
	return []string{
		c.Timestamp.Format(time.RFC3339),
		"INAP",
		c.TransactionID,
		c.IMSI,
		c.MSISDN,
		c.OperationType,
		c.Result,
		fmt.Sprintf("%d", c.ResultCode),
		fmt.Sprintf("%d", c.DurationMs),
		fmt.Sprintf("%d", c.ServiceKey),
		c.CallingParty,
		c.CalledParty,
		c.TriggerType,
	}
}

func (c *InapCDR) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":      c.Timestamp.Format(time.RFC3339),
		"protocol":       "INAP",
		"transaction_id": c.TransactionID,
		"imsi":           c.IMSI,
		"msisdn":         c.MSISDN,
		"service_key":    c.ServiceKey,
		"calling_party":  c.CallingParty,
		"called_party":   c.CalledParty,
		"operation_type": c.OperationType,
		"trigger_type":   c.TriggerType,
		"result":         c.Result,
		"result_code":    c.ResultCode,
		"duration_ms":    c.DurationMs,
	}
}

func (c *InapCDR) GetTimestamp() time.Time { return c.Timestamp }
func (c *InapCDR) GetProtocol() string     { return "INAP" }
