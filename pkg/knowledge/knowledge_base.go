package knowledge

import (
	"fmt"
	"strings"
)

// ProtocolStandard represents a telecom standard document
type ProtocolStandard struct {
	ID          string   `json:"id"`           // e.g., "TS 29.272"
	Title       string   `json:"title"`        // e.g., "Evolved Packet System; MME and SGSN related interfaces based on Diameter protocol"
	Version     string   `json:"version"`      // e.g., "Release 16"
	URL         string   `json:"url"`          // Link to standard document
	Organization string  `json:"organization"` // "3GPP", "IETF", "ETSI"
	Protocols   []string `json:"protocols"`    // Associated protocols: ["S6a", "S6d", "Diameter"]
	Description string   `json:"description"`  // Brief description
}

// ProcedureReference represents a specific procedure in a standard
type ProcedureReference struct {
	StandardID  string   `json:"standard_id"`   // e.g., "TS 29.272"
	Section     string   `json:"section"`       // e.g., "7.2.3"
	Protocol    string   `json:"protocol"`      // e.g., "S6a"
	Procedure   string   `json:"procedure"`     // e.g., "Update Location Request"
	MessageType string   `json:"message_type"`  // e.g., "ULR"
	Description string   `json:"description"`   // Detailed description
	Purpose     string   `json:"purpose"`       // What this procedure does
	IEs         []string `json:"ies"`           // Information Elements
	Flows       []string `json:"flows"`         // Message flow steps
}

// ErrorCodeReference represents an error or cause code
type ErrorCodeReference struct {
	Protocol    string `json:"protocol"`     // e.g., "Diameter", "GTP", "MAP"
	Code        int    `json:"code"`         // Numeric code
	Name        string `json:"name"`         // e.g., "DIAMETER_ERROR_USER_UNKNOWN"
	Description string `json:"description"`  // What this error means
	Causes      string `json:"causes"`       // Common causes
	Solutions   string `json:"solutions"`    // How to fix
	StandardRef string `json:"standard_ref"` // e.g., "TS 29.272 Section 7.4.3"
	Severity    string `json:"severity"`     // "critical", "major", "minor", "warning"
}

// KnowledgeBase holds all protocol standards and references
type KnowledgeBase struct {
	standards        map[string]*ProtocolStandard
	procedures       map[string][]*ProcedureReference // Key: protocol name
	errorCodes       map[string]map[int]*ErrorCodeReference // Key: protocol, subkey: code
	searchIndex      map[string][]interface{} // Simple search index
}

// NewKnowledgeBase creates a new knowledge base with all standards
func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{
		standards:   make(map[string]*ProtocolStandard),
		procedures:  make(map[string][]*ProcedureReference),
		errorCodes:  make(map[string]map[int]*ErrorCodeReference),
		searchIndex: make(map[string][]interface{}),
	}

	// Load all standards
	kb.load3GPPStandards()
	kb.loadIETFStandards()
	kb.loadProcedures()
	kb.loadErrorCodes()
	kb.buildSearchIndex()

	return kb
}

// Load 3GPP standards
func (kb *KnowledgeBase) load3GPPStandards() {
	standards := []*ProtocolStandard{
		{
			ID:           "TS 29.272",
			Title:        "Evolved Packet System (EPS); Mobility Management Entity (MME) and Serving GPRS Support Node (SGSN) related interfaces based on Diameter protocol",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.272/",
			Organization: "3GPP",
			Protocols:    []string{"S6a", "S6d", "Diameter"},
			Description:  "Defines Diameter-based S6a (MME-HSS) and S6d (SGSN-HSS) interfaces for LTE/EPS subscriber data management.",
		},
		{
			ID:           "TS 29.274",
			Title:        "3GPP Evolved Packet System (EPS); Evolved General Packet Radio Service (GPRS) Tunnelling Protocol for Control plane (GTPv2-C)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.274/",
			Organization: "3GPP",
			Protocols:    []string{"GTPv2-C", "GTP"},
			Description:  "Defines GTPv2-C protocol for control plane communication on S4, S5, S8, S11 interfaces.",
		},
		{
			ID:           "TS 29.244",
			Title:        "Interface between the Control Plane and the User Plane nodes (PFCP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.244/",
			Organization: "3GPP",
			Protocols:    []string{"PFCP", "N4"},
			Description:  "Defines Packet Forwarding Control Protocol (PFCP) for SMF-UPF communication on N4 interface.",
		},
		{
			ID:           "TS 29.002",
			Title:        "Mobile Application Part (MAP) specification",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.002/",
			Organization: "3GPP",
			Protocols:    []string{"MAP", "SS7"},
			Description:  "Defines MAP protocol for 2G/3G core network signaling (HLR, VLR, MSC, SGSN).",
		},
		{
			ID:           "TS 29.078",
			Title:        "Customised Applications for Mobile network Enhanced Logic (CAMEL) Phase 4; CAMEL Application Part (CAP) specification",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.078/",
			Organization: "3GPP",
			Protocols:    []string{"CAP", "CAMEL", "IN"},
			Description:  "Defines CAP protocol for intelligent network services and prepaid charging.",
		},
		{
			ID:           "TS 38.413",
			Title:        "NG-RAN; NG Application Protocol (NGAP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/38_series/38.413/",
			Organization: "3GPP",
			Protocols:    []string{"NGAP", "N2", "5G"},
			Description:  "Defines NGAP for 5G NG-RAN (gNB) to AMF communication on N2 interface.",
		},
		{
			ID:           "TS 36.413",
			Title:        "Evolved Universal Terrestrial Radio Access Network (E-UTRAN); S1 Application Protocol (S1AP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/36_series/36.413/",
			Organization: "3GPP",
			Protocols:    []string{"S1AP", "S1", "4G"},
			Description:  "Defines S1AP for 4G E-UTRAN (eNB) to MME communication on S1-MME interface.",
		},
		{
			ID:           "TS 24.301",
			Title:        "Non-Access-Stratum (NAS) protocol for Evolved Packet System (EPS)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/24_series/24.301/",
			Organization: "3GPP",
			Protocols:    []string{"NAS", "EPS", "4G"},
			Description:  "Defines NAS protocol for UE-MME signaling in LTE/EPS networks.",
		},
		{
			ID:           "TS 24.501",
			Title:        "Non-Access-Stratum (NAS) protocol for 5G System (5GS)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/24_series/24.501/",
			Organization: "3GPP",
			Protocols:    []string{"NAS", "5GS", "5G"},
			Description:  "Defines NAS protocol for UE-AMF signaling in 5G networks.",
		},
		{
			ID:           "TS 29.228",
			Title:        "IP Multimedia (IM) Subsystem Cx and Dx interfaces; Signalling flows and message contents",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.228/",
			Organization: "3GPP",
			Protocols:    []string{"Cx", "Dx", "Diameter", "IMS"},
			Description:  "Defines Cx/Dx interfaces between I-CSCF/S-CSCF and HSS for IMS.",
		},
		{
			ID:           "TS 29.213",
			Title:        "Policy and Charging Control signalling flows and Quality of Service (QoS) parameter mapping",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/29_series/29.213/",
			Organization: "3GPP",
			Protocols:    []string{"Gx", "Diameter", "PCC"},
			Description:  "Defines Gx interface between PCEF (PGW) and PCRF for policy control.",
		},
		{
			ID:           "TS 32.299",
			Title:        "Telecommunication management; Charging management; Diameter charging applications",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/ftp/Specs/archive/32_series/32.299/",
			Organization: "3GPP",
			Protocols:    []string{"Gy", "Ro", "Diameter", "Charging"},
			Description:  "Defines online charging interfaces (Gy, Ro) using Diameter protocol.",
		},
	}

	for _, std := range standards {
		kb.standards[std.ID] = std
	}
}

// Load IETF/RFC standards
func (kb *KnowledgeBase) loadIETFStandards() {
	standards := []*ProtocolStandard{
		{
			ID:           "RFC 6733",
			Title:        "Diameter Base Protocol",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc6733",
			Organization: "IETF",
			Protocols:    []string{"Diameter"},
			Description:  "Defines the base Diameter protocol (AAA framework).",
		},
		{
			ID:           "RFC 793",
			Title:        "Transmission Control Protocol (TCP)",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc793",
			Organization: "IETF",
			Protocols:    []string{"TCP"},
			Description:  "Defines TCP protocol for reliable stream-oriented communication.",
		},
		{
			ID:           "RFC 768",
			Title:        "User Datagram Protocol (UDP)",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc768",
			Organization: "IETF",
			Protocols:    []string{"UDP"},
			Description:  "Defines UDP protocol for datagram communication.",
		},
		{
			ID:           "RFC 4960",
			Title:        "Stream Control Transmission Protocol (SCTP)",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc4960",
			Organization: "IETF",
			Protocols:    []string{"SCTP"},
			Description:  "Defines SCTP protocol used by many telecom signaling protocols.",
		},
		{
			ID:           "RFC 2616",
			Title:        "Hypertext Transfer Protocol -- HTTP/1.1",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc2616",
			Organization: "IETF",
			Protocols:    []string{"HTTP"},
			Description:  "Defines HTTP/1.1 protocol.",
		},
		{
			ID:           "RFC 5246",
			Title:        "The Transport Layer Security (TLS) Protocol Version 1.2",
			Version:      "",
			URL:          "https://tools.ietf.org/html/rfc5246",
			Organization: "IETF",
			Protocols:    []string{"TLS", "SSL"},
			Description:  "Defines TLS 1.2 for secure communication.",
		},
	}

	for _, std := range standards {
		kb.standards[std.ID] = std
	}
}

// Load procedures
func (kb *KnowledgeBase) loadProcedures() {
	procedures := []*ProcedureReference{
		// S6a Procedures
		{
			StandardID:  "TS 29.272",
			Section:     "7.2.3",
			Protocol:    "S6a",
			Procedure:   "Update Location Request/Answer",
			MessageType: "ULR/ULA",
			Description: "Sent by MME to HSS to update subscriber location and retrieve subscriber data when UE attaches or moves to a new tracking area.",
			Purpose:     "Informs HSS of subscriber's current MME and retrieves subscription data including APN configurations, QoS profiles, and roaming restrictions.",
			IEs:         []string{"User-Name (IMSI)", "Visited-PLMN-Id", "ULR-Flags", "RAT-Type"},
			Flows:       []string{"1. MME sends ULR to HSS", "2. HSS validates IMSI and location", "3. HSS sends ULA with subscription data", "4. MME stores subscription data"},
		},
		{
			StandardID:  "TS 29.272",
			Section:     "7.2.5",
			Protocol:    "S6a",
			Procedure:   "Authentication Information Request/Answer",
			MessageType: "AIR/AIA",
			Description: "Sent by MME to HSS to request authentication vectors for UE authentication.",
			Purpose:     "Retrieves authentication vectors (RAND, AUTN, XRES, KASME) for EPS-AKA authentication.",
			IEs:         []string{"User-Name (IMSI)", "Requested-EUTRAN-Authentication-Info"},
			Flows:       []string{"1. MME sends AIR to HSS", "2. HSS generates authentication vectors", "3. HSS sends AIA with authentication vectors", "4. MME performs authentication"},
		},
		// GTPv2-C Procedures
		{
			StandardID:  "TS 29.274",
			Section:     "7.2.1",
			Protocol:    "GTPv2-C",
			Procedure:   "Create Session Request/Response",
			MessageType: "CSReq/CSResp",
			Description: "Sent by MME/SGSN to SGW/PGW to create a new session for a UE.",
			Purpose:     "Establishes GTP tunnels (bearers) for user plane traffic between eNB, SGW, and PGW.",
			IEs:         []string{"IMSI", "MSISDN", "APN", "RAT-Type", "Bearer-Contexts", "PDN-Type"},
			Flows:       []string{"1. MME sends CSReq to SGW", "2. SGW forwards to PGW", "3. PGW allocates IP address", "4. PGW responds with CSResp", "5. SGW forwards to MME"},
		},
		{
			StandardID:  "TS 29.274",
			Section:     "7.2.7",
			Protocol:    "GTPv2-C",
			Procedure:   "Delete Session Request/Response",
			MessageType: "DSReq/DSResp",
			Description: "Sent to delete an existing PDN connection.",
			Purpose:     "Releases GTP tunnels and resources when UE detaches or PDN connection is deleted.",
			IEs:         []string{"Cause", "EBI (EPS Bearer Identity)"},
			Flows:       []string{"1. MME sends DSReq to SGW", "2. SGW forwards to PGW", "3. PGW releases resources", "4. PGW sends DSResp", "5. SGW forwards to MME"},
		},
		// PFCP Procedures
		{
			StandardID:  "TS 29.244",
			Section:     "7.4.2",
			Protocol:    "PFCP",
			Procedure:   "Session Establishment Request/Response",
			MessageType: "PFCP Session Establishment",
			Description: "Sent by SMF to UPF to establish a PFCP session for 5G PDU session.",
			Purpose:     "Creates packet forwarding rules in UPF for user plane traffic.",
			IEs:         []string{"Node ID", "F-SEID", "PDR", "FAR", "QER"},
			Flows:       []string{"1. SMF sends Session Establishment Request to UPF", "2. UPF creates forwarding rules", "3. UPF allocates F-TEID", "4. UPF responds with Session Establishment Response"},
		},
		// MAP Procedures
		{
			StandardID:  "TS 29.002",
			Section:     "7.3",
			Protocol:    "MAP",
			Procedure:   "Update Location",
			MessageType: "UpdateLocationArg/UpdateLocationRes",
			Description: "Sent by VLR to HLR when subscriber enters a new location area.",
			Purpose:     "Updates subscriber location in HLR and retrieves subscriber data.",
			IEs:         []string{"IMSI", "MSC Number", "VLR Number", "LMSI"},
			Flows:       []string{"1. VLR sends UpdateLocationArg to HLR", "2. HLR validates and updates location", "3. HLR cancels old VLR if needed", "4. HLR sends UpdateLocationRes with subscriber data"},
		},
		// NGAP Procedures
		{
			StandardID:  "TS 38.413",
			Section:     "8.3.1",
			Protocol:    "NGAP",
			Procedure:   "Initial Context Setup",
			MessageType: "InitialContextSetupRequest/Response",
			Description: "Sent by AMF to gNB to establish initial UE context after registration.",
			Purpose:     "Configures radio resources, security, and QoS for UE in 5G network.",
			IEs:         []string{"AMF-UE-NGAP-ID", "RAN-UE-NGAP-ID", "PDU-Session-Resource-Setup-List", "Security-Key"},
			Flows:       []string{"1. AMF sends InitialContextSetupRequest to gNB", "2. gNB configures radio resources", "3. gNB establishes security", "4. gNB responds with InitialContextSetupResponse"},
		},
		// S1AP Procedures
		{
			StandardID:  "TS 36.413",
			Section:     "8.3.1",
			Protocol:    "S1AP",
			Procedure:   "Initial Context Setup",
			MessageType: "InitialContextSetupRequest/Response",
			Description: "Sent by MME to eNB to establish initial UE context after attach.",
			Purpose:     "Configures radio resources, security, and bearers for UE in LTE network.",
			IEs:         []string{"MME-UE-S1AP-ID", "eNB-UE-S1AP-ID", "E-RABToBeSetupList", "Security-Key"},
			Flows:       []string{"1. MME sends InitialContextSetupRequest to eNB", "2. eNB configures radio resources", "3. eNB establishes security and bearers", "4. eNB responds with InitialContextSetupResponse"},
		},
	}

	for _, proc := range procedures {
		kb.procedures[proc.Protocol] = append(kb.procedures[proc.Protocol], proc)
	}
}

// Load error codes
func (kb *KnowledgeBase) loadErrorCodes() {
	// Diameter error codes
	diameterErrors := []*ErrorCodeReference{
		{
			Protocol:    "Diameter",
			Code:        5001,
			Name:        "DIAMETER_ERROR_USER_UNKNOWN",
			Description: "The specified user (IMSI) is not known in the HSS/HLR.",
			Causes:      "Subscriber not provisioned in HSS, IMSI typo, database synchronization issue, subscriber deleted but still in network cache.",
			Solutions:   "1. Verify IMSI is correctly provisioned in HSS. 2. Check HSS database for subscriber record. 3. Verify no recent provisioning changes. 4. Clear stale cache in MME if applicable. Reference: TS 29.272 Section 7.4.3",
			StandardRef: "TS 29.272 Section 7.4.3",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        5004,
			Name:        "DIAMETER_ERROR_ROAMING_NOT_ALLOWED",
			Description: "Roaming is not allowed for this subscriber in the visited network.",
			Causes:      "Roaming agreement not in place, subscriber barred from roaming, visited PLMN not in allowed PLMN list, roaming blacklist/whitelist mismatch.",
			Solutions:   "1. Verify roaming agreements between operators. 2. Check subscriber roaming permissions in HSS. 3. Validate VPLMN-ID in ULR message. 4. Review roaming restrictions in subscription profile. Reference: TS 29.272 Section 7.4.3",
			StandardRef: "TS 29.272 Section 7.4.3",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        5420,
			Name:        "DIAMETER_ERROR_UNKNOWN_EPS_SUBSCRIPTION",
			Description: "The HSS does not have EPS subscription data for this user.",
			Causes:      "Subscriber has only 2G/3G subscription, EPS not activated, migration to LTE not completed, provisioning error.",
			Solutions:   "1. Verify EPS subscription is provisioned in HSS. 2. Check if subscriber has LTE/4G service activated. 3. Review subscription plan. 4. Ensure APN configurations are present. Reference: TS 29.272 Section 7.4.4",
			StandardRef: "TS 29.272 Section 7.4.4",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        5421,
			Name:        "DIAMETER_ERROR_RAT_NOT_ALLOWED",
			Description: "The requested Radio Access Type (RAT) is not allowed for this subscriber.",
			Causes:      "RAT restrictions in subscription (e.g., only 2G/3G allowed, LTE barred), network policy, roaming RAT restrictions.",
			Solutions:   "1. Check RAT-Type restrictions in HSS subscription data. 2. Verify ULR RAT-Type value matches subscription. 3. Review roaming RAT policies. 4. Check if LTE/5G service is activated for subscriber. Reference: TS 29.272 Section 7.4.4",
			StandardRef: "TS 29.272 Section 7.4.4",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        3002,
			Name:        "DIAMETER_UNABLE_TO_DELIVER",
			Description: "The message could not be delivered to the destination host.",
			Causes:      "Destination unreachable, network connectivity issue, routing problem, peer down, congestion.",
			Solutions:   "1. Verify network connectivity to HSS. 2. Check Diameter routing configuration. 3. Verify destination host is operational. 4. Review DRA/DEA routing tables. 5. Check for network congestion. Reference: RFC 6733 Section 7.1.3",
			StandardRef: "RFC 6733 Section 7.1.3",
			Severity:    "critical",
		},
		{
			Protocol:    "Diameter",
			Code:        3003,
			Name:        "DIAMETER_REALM_NOT_SERVED",
			Description:   "The realm specified in the Destination-Realm AVP is not served by any peer.",
			Causes:      "Incorrect realm configuration, DRA misconfiguration, realm routing not configured, typo in realm name.",
			Solutions:   "1. Verify Destination-Realm AVP value. 2. Check DRA routing configuration for realm. 3. Ensure peer connections are established. 4. Review Diameter peer table. Reference: RFC 6733 Section 7.1.3",
			StandardRef: "RFC 6733 Section 7.1.3",
			Severity:    "critical",
		},
	}

	// GTP error causes
	gtpErrors := []*ErrorCodeReference{
		{
			Protocol:    "GTP",
			Code:        64,
			Name:        "Context Not Found",
			Description: "The requested GTP context (bearer/session) does not exist on the receiving node.",
			Causes:      "Session already deleted, timeout, restart without state sync, mismatch in TEID, SGW/PGW restart, network element failure.",
			Solutions:   "1. Check session lifecycle - may have been legitimately deleted. 2. Verify TEID values in messages. 3. Check for recent SGW/PGW restarts. 4. Enable GTP echo procedure for peer monitoring. 5. Review session timeout configurations. Reference: TS 29.274 Section 8.4",
			StandardRef: "TS 29.274 Section 8.4",
			Severity:    "major",
		},
		{
			Protocol:    "GTP",
			Code:        67,
			Name:        "Missing or Unknown APN",
			Description: "The Access Point Name is not configured or not recognized by PGW.",
			Causes:      "APN not provisioned in PGW, typo in APN name, APN not in HSS subscription, default APN not configured.",
			Solutions:   "1. Verify APN is configured in PGW. 2. Check APN name in Create Session Request. 3. Validate APN in HSS subscription profile. 4. Ensure default APN is configured if UE doesn't specify. Reference: TS 29.274 Section 8.4",
			StandardRef: "TS 29.274 Section 8.4",
			Severity:    "major",
		},
		{
			Protocol:    "GTP",
			Code:        72,
			Name:        "Semantic Error in TFT Operation",
			Description: "Traffic Flow Template (TFT) has semantic errors.",
			Causes:      "Invalid packet filter, conflicting TFT rules, incorrect precedence, malformed TFT IE.",
			Solutions:   "1. Validate TFT packet filters syntax. 2. Check for conflicting filter rules. 3. Review precedence values (must be unique). 4. Ensure TFT IE format is correct. Reference: TS 29.274 Section 8.4",
			StandardRef: "TS 29.274 Section 8.4",
			Severity:    "minor",
		},
		{
			Protocol:    "GTP",
			Code:        91,
			Name:        "No Resources Available",
			Description: "The node has insufficient resources to handle the request.",
			Causes:      "Memory exhaustion, CPU overload, license limit reached, maximum bearers/sessions exceeded, network congestion.",
			Solutions:   "1. Check node resource utilization (CPU, memory). 2. Verify license limits and current usage. 3. Review maximum session/bearer configuration. 4. Check for memory leaks. 5. Consider scaling or load balancing. Reference: TS 29.274 Section 8.4",
			StandardRef: "TS 29.274 Section 8.4",
			Severity:    "critical",
		},
		{
			Protocol:    "GTP",
			Code:        93,
			Name:        "Request Rejected",
			Description: "The request was rejected by the receiving node (generic rejection).",
			Causes:      "Policy violation, feature not supported, administrative restrictions, temporary overload, misconfiguration.",
			Solutions:   "1. Check detailed cause in message. 2. Review policy rules. 3. Verify feature support on both sides. 4. Check administrative configurations. 5. Review logs for specific rejection reason. Reference: TS 29.274 Section 8.4",
			StandardRef: "TS 29.274 Section 8.4",
			Severity:    "major",
		},
	}

	// MAP error causes
	mapErrors := []*ErrorCodeReference{
		{
			Protocol:    "MAP",
			Code:        1,
			Name:        "Unknown Subscriber",
			Description: "The subscriber identity (IMSI) is not known in the HLR.",
			Causes:      "IMSI not provisioned, subscriber deleted, database corruption, incorrect IMSI value in message.",
			Solutions:   "1. Verify IMSI exists in HLR. 2. Check for recent deletions. 3. Validate IMSI format (MCC-MNC-MSIN). 4. Check HLR database integrity. Reference: TS 29.002 Section 17.7.1",
			StandardRef: "TS 29.002 Section 17.7.1",
			Severity:    "major",
		},
		{
			Protocol:    "MAP",
			Code:        8,
			Name:        "Roaming Not Allowed",
			Description: "The subscriber is not allowed to roam in the requested network.",
			Causes:      "No roaming agreement, PLMN not in allowed list, subscriber roaming barred, national roaming issue.",
			Solutions:   "1. Verify roaming agreements. 2. Check subscriber roaming permissions. 3. Validate VPLMN in subscription. 4. Review ODB (Operator Determined Barring) settings. Reference: TS 29.002 Section 17.7.1",
			StandardRef: "TS 29.002 Section 17.7.1",
			Severity:    "major",
		},
		{
			Protocol:    "MAP",
			Code:        21,
			Name:        "Facility Not Supported",
			Description: "The requested supplementary service or feature is not supported.",
			Causes:      "Service not subscribed, feature not supported by network, incompatible versions, service barred.",
			Solutions:   "1. Check supplementary service subscription. 2. Verify feature support in network. 3. Review service provisioning. 4. Check MAP protocol version compatibility. Reference: TS 29.002 Section 17.7.1",
			StandardRef: "TS 29.002 Section 17.7.1",
			Severity:    "minor",
		},
	}

	// Initialize map for each protocol
	kb.errorCodes["Diameter"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["GTP"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["MAP"] = make(map[int]*ErrorCodeReference)

	// Add all error codes
	for _, err := range diameterErrors {
		kb.errorCodes["Diameter"][err.Code] = err
	}
	for _, err := range gtpErrors {
		kb.errorCodes["GTP"][err.Code] = err
	}
	for _, err := range mapErrors {
		kb.errorCodes["MAP"][err.Code] = err
	}
}

// Build search index for fast lookups
func (kb *KnowledgeBase) buildSearchIndex() {
	// Index standards
	for _, std := range kb.standards {
		keywords := []string{
			strings.ToLower(std.ID),
			strings.ToLower(std.Title),
		}
		for _, proto := range std.Protocols {
			keywords = append(keywords, strings.ToLower(proto))
		}

		for _, keyword := range keywords {
			kb.searchIndex[keyword] = append(kb.searchIndex[keyword], std)
		}
	}

	// Index procedures
	for _, procs := range kb.procedures {
		for _, proc := range procs {
			keywords := []string{
				strings.ToLower(proc.Protocol),
				strings.ToLower(proc.Procedure),
				strings.ToLower(proc.MessageType),
			}

			for _, keyword := range keywords {
				kb.searchIndex[keyword] = append(kb.searchIndex[keyword], proc)
			}
		}
	}

	// Index error codes
	for protocol, errors := range kb.errorCodes {
		for _, err := range errors {
			keywords := []string{
				strings.ToLower(protocol),
				strings.ToLower(err.Name),
				fmt.Sprintf("%d", err.Code),
			}

			for _, keyword := range keywords {
				kb.searchIndex[keyword] = append(kb.searchIndex[keyword], err)
			}
		}
	}
}

// GetStandard returns a standard by ID
func (kb *KnowledgeBase) GetStandard(id string) (*ProtocolStandard, error) {
	std, ok := kb.standards[id]
	if !ok {
		return nil, fmt.Errorf("standard %s not found", id)
	}
	return std, nil
}

// GetProceduresByProtocol returns all procedures for a protocol
func (kb *KnowledgeBase) GetProceduresByProtocol(protocol string) []*ProcedureReference {
	return kb.procedures[protocol]
}

// GetErrorCode returns error code information
func (kb *KnowledgeBase) GetErrorCode(protocol string, code int) (*ErrorCodeReference, error) {
	protocolErrors, ok := kb.errorCodes[protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s not found", protocol)
	}

	errRef, ok := protocolErrors[code]
	if !ok {
		return nil, fmt.Errorf("error code %d not found for protocol %s", code, protocol)
	}

	return errRef, nil
}

// Search performs a search across all knowledge base
func (kb *KnowledgeBase) Search(query string) []interface{} {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	// Direct lookup
	if results, ok := kb.searchIndex[query]; ok {
		return results
	}

	// Partial match
	var results []interface{}
	seen := make(map[interface{}]bool)

	for keyword, items := range kb.searchIndex {
		if strings.Contains(keyword, query) || strings.Contains(query, keyword) {
			for _, item := range items {
				if !seen[item] {
					results = append(results, item)
					seen[item] = true
				}
			}
		}
	}

	return results
}

// ListAllStandards returns all standards
func (kb *KnowledgeBase) ListAllStandards() []*ProtocolStandard {
	standards := make([]*ProtocolStandard, 0, len(kb.standards))
	for _, std := range kb.standards {
		standards = append(standards, std)
	}
	return standards
}

// ListAllProtocols returns list of all protocols
func (kb *KnowledgeBase) ListAllProtocols() []string {
	protocols := make(map[string]bool)
	for _, procs := range kb.procedures {
		for _, proc := range procs {
			protocols[proc.Protocol] = true
		}
	}

	result := make([]string, 0, len(protocols))
	for proto := range protocols {
		result = append(result, proto)
	}
	return result
}
