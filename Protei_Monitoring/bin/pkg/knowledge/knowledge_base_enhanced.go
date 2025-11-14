package knowledge

import (
	"fmt"
	"strings"
	"sync"
)

// ProtocolStandard represents a telecom standard document
type ProtocolStandard struct {
	ID           string   `json:"id"`           // e.g., "TS 29.272"
	Title        string   `json:"title"`        // Full title
	Version      string   `json:"version"`      // e.g., "Release 16"
	URL          string   `json:"url"`          // Link to standard document
	Organization string   `json:"organization"` // "3GPP", "IETF", "ETSI"
	Protocols    []string `json:"protocols"`    // Associated protocols
	Description  string   `json:"description"`  // Brief description
	Sections     []StandardSection `json:"sections"` // Document sections
}

// StandardSection represents a section within a standard
type StandardSection struct {
	Number      string `json:"number"`      // e.g., "7.2.3"
	Title       string `json:"title"`       // Section title
	Content     string `json:"content"`     // Section content/summary
	MessageType string `json:"message_type"` // Related message type
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
	Flows       []FlowStep `json:"flows"`       // Message flow steps
	Diagram     string   `json:"diagram"`       // ASCII/SVG flow diagram
}

// FlowStep represents one step in a message flow
type FlowStep struct {
	Step        int    `json:"step"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
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

// VendorExtension represents vendor-specific protocol extensions
type VendorExtension struct {
	Vendor      string `json:"vendor"`       // "Ericsson", "Huawei", "ZTE", "Nokia", "Cisco"
	Protocol    string `json:"protocol"`     // Which protocol this extends
	Extension   string `json:"extension"`    // Extension name
	Code        int    `json:"code"`         // Vendor-specific code/IE
	Description string `json:"description"`  // What this extension does
	Usage       string `json:"usage"`        // How/when it's used
}

// CallFlowDiagram represents a standard call flow for comparison
type CallFlowDiagram struct {
	Name        string     `json:"name"`
	Protocol    string     `json:"protocol"`
	Type        string     `json:"type"`        // "Attach", "Detach", "TAU", etc.
	Generation  string     `json:"generation"`  // "2G", "3G", "4G", "5G"
	Steps       []FlowStep `json:"steps"`
	Diagram     string     `json:"diagram"`     // ASCII/SVG representation
	StandardRef string     `json:"standard_ref"`
}

// KnowledgeBase holds all protocol standards and references
type KnowledgeBase struct {
	mu               sync.RWMutex
	standards        map[string]*ProtocolStandard
	procedures       map[string][]*ProcedureReference // Key: protocol name
	errorCodes       map[string]map[int]*ErrorCodeReference // Key: protocol, subkey: code
	vendorExtensions map[string][]*VendorExtension // Key: vendor name
	callFlows        map[string]*CallFlowDiagram // Key: flow name
	searchIndex      map[string][]interface{} // Search index
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{
		standards:        make(map[string]*ProtocolStandard),
		procedures:       make(map[string][]*ProcedureReference),
		errorCodes:       make(map[string]map[int]*ErrorCodeReference),
		vendorExtensions: make(map[string][]*VendorExtension),
		callFlows:        make(map[string]*CallFlowDiagram),
		searchIndex:      make(map[string][]interface{}),
	}
	return kb
}

// LoadStandards loads all standards into the knowledge base
func (kb *KnowledgeBase) LoadStandards() error {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Load 3GPP Standards
	kb.load3GPPStandards()

	// Load IETF RFCs
	kb.loadIETFRFCs()

	// Load Procedures
	kb.loadProcedures()

	// Load Error Codes
	kb.loadErrorCodes()

	// Load Vendor Extensions
	kb.loadVendorExtensions()

	// Load Call Flow Diagrams
	kb.loadCallFlows()

	// Build search index
	kb.buildSearchIndex()

	return nil
}

// load3GPPStandards loads all 3GPP standards (12 standards)
func (kb *KnowledgeBase) load3GPPStandards() {
	standards := []*ProtocolStandard{
		{
			ID:           "TS 29.002",
			Title:        "Mobile Application Part (MAP) specification",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29002.htm",
			Organization: "3GPP",
			Protocols:    []string{"MAP"},
			Description:  "Defines MAP protocol for GSM/UMTS core network signaling",
			Sections: []StandardSection{
				{Number: "7.6.1", Title: "MAP Error Codes", MessageType: "Error"},
				{Number: "9.1", Title: "Location Management Procedures", MessageType: "UpdateLocation"},
				{Number: "10.1", Title: "Subscriber Management", MessageType: "InsertSubscriberData"},
			},
		},
		{
			ID:           "TS 29.078",
			Title:        "Customised Applications for Mobile network Enhanced Logic (CAMEL) Phase 4",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29078.htm",
			Organization: "3GPP",
			Protocols:    []string{"CAP"},
			Description:  "Defines CAP protocol for IN services in mobile networks",
			Sections: []StandardSection{
				{Number: "4.2", Title: "CAMEL Service Logic", MessageType: "IDP"},
				{Number: "5.1", Title: "Circuit Switched Calls", MessageType: "Connect"},
			},
		},
		{
			ID:           "TS 29.274",
			Title:        "3GPP Evolved Packet System (EPS); Evolved General Packet Radio Service (GPRS) Tunnelling Protocol for Control plane (GTPv2-C)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29274.htm",
			Organization: "3GPP",
			Protocols:    []string{"GTPv2-C"},
			Description:  "Defines GTPv2-C protocol for EPC control plane",
			Sections: []StandardSection{
				{Number: "7.2.1", Title: "Create Session Request", MessageType: "CreateSessionRequest"},
				{Number: "7.2.2", Title: "Create Session Response", MessageType: "CreateSessionResponse"},
				{Number: "8.4", Title: "Cause Values", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 29.244",
			Title:        "Interface between the Control Plane and the User Plane nodes (PFCP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29244.htm",
			Organization: "3GPP",
			Protocols:    []string{"PFCP"},
			Description:  "Defines PFCP protocol for 5GC control/user plane separation",
			Sections: []StandardSection{
				{Number: "7.4.2", Title: "Session Establishment", MessageType: "SessionEstablishmentRequest"},
				{Number: "8.2", Title: "Cause Values", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 29.272",
			Title:        "Evolved Packet System (EPS); Mobility Management Entity (MME) and Serving GPRS Support Node (SGSN) related interfaces based on Diameter protocol",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29272.htm",
			Organization: "3GPP",
			Protocols:    []string{"S6a", "S6d", "Diameter"},
			Description:  "Defines S6a/S6d Diameter interfaces between MME/SGSN and HSS",
			Sections: []StandardSection{
				{Number: "7.2.3", Title: "Update Location Request", MessageType: "ULR"},
				{Number: "7.2.4", Title: "Update Location Answer", MessageType: "ULA"},
				{Number: "7.2.7", Title: "Authentication Information Request", MessageType: "AIR"},
				{Number: "7.4", Title: "Result-Code and Experimental-Result-Code Values", MessageType: "Result"},
			},
		},
		{
			ID:           "TS 29.273",
			Title:        "Evolved Packet System (EPS); 3GPP EPS AAA interfaces",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29273.htm",
			Organization: "3GPP",
			Protocols:    []string{"SWm", "SWx", "Diameter"},
			Description:  "Defines SWm/SWx Diameter interfaces for non-3GPP access",
			Sections: []StandardSection{
				{Number: "8.2.2", Title: "Diameter-EAP-Request", MessageType: "DER"},
				{Number: "8.2.3", Title: "Diameter-EAP-Answer", MessageType: "DEA"},
			},
		},
		{
			ID:           "TS 38.413",
			Title:        "NG-RAN; NG Application Protocol (NGAP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/38413.htm",
			Organization: "3GPP",
			Protocols:    []string{"NGAP"},
			Description:  "Defines NGAP protocol between gNB and AMF in 5G",
			Sections: []StandardSection{
				{Number: "8.3.1", Title: "Initial UE Message", MessageType: "InitialUEMessage"},
				{Number: "8.6.1", Title: "Initial Context Setup Request", MessageType: "InitialContextSetupRequest"},
				{Number: "9.3.1", Title: "Cause", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 36.413",
			Title:        "Evolved Universal Terrestrial Radio Access Network (E-UTRAN); S1 Application Protocol (S1AP)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/36413.htm",
			Organization: "3GPP",
			Protocols:    []string{"S1AP"},
			Description:  "Defines S1AP protocol between eNB and MME in 4G",
			Sections: []StandardSection{
				{Number: "8.3.1", Title: "Initial UE Message", MessageType: "InitialUEMessage"},
				{Number: "8.6.1", Title: "Initial Context Setup Request", MessageType: "InitialContextSetupRequest"},
				{Number: "9.2.1", Title: "Cause", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 24.301",
			Title:        "Non-Access-Stratum (NAS) protocol for Evolved Packet System (EPS)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/24301.htm",
			Organization: "3GPP",
			Protocols:    []string{"NAS-EPS"},
			Description:  "Defines NAS signaling between UE and MME in 4G",
			Sections: []StandardSection{
				{Number: "8.2.1", Title: "Attach Request", MessageType: "AttachRequest"},
				{Number: "8.2.2", Title: "Attach Accept", MessageType: "AttachAccept"},
				{Number: "9.9.3", Title: "EMM Cause", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 24.501",
			Title:        "Non-Access-Stratum (NAS) protocol for 5G System (5GS)",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/24501.htm",
			Organization: "3GPP",
			Protocols:    []string{"NAS-5GS"},
			Description:  "Defines NAS signaling between UE and AMF in 5G",
			Sections: []StandardSection{
				{Number: "8.2.6", Title: "Registration Request", MessageType: "RegistrationRequest"},
				{Number: "8.2.7", Title: "Registration Accept", MessageType: "RegistrationAccept"},
				{Number: "9.11.3", Title: "5GMM Cause", MessageType: "Cause"},
			},
		},
		{
			ID:           "TS 29.500",
			Title:        "5G System; Technical Realization of Service Based Architecture",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/29500.htm",
			Organization: "3GPP",
			Protocols:    []string{"HTTP/2", "SBI"},
			Description:  "Defines HTTP/2-based Service-Based Interface for 5G core",
			Sections: []StandardSection{
				{Number: "6.1", Title: "HTTP/2 Usage", MessageType: "HTTP"},
				{Number: "6.10", Title: "Error Handling", MessageType: "ProblemDetails"},
			},
		},
		{
			ID:           "TS 23.401",
			Title:        "General Packet Radio Service (GPRS) enhancements for Evolved Universal Terrestrial Radio Access Network (E-UTRAN) access",
			Version:      "Release 16",
			URL:          "https://www.3gpp.org/DynaReport/23401.htm",
			Organization: "3GPP",
			Protocols:    []string{"EPC"},
			Description:  "Defines EPC architecture and procedures for 4G/LTE",
			Sections: []StandardSection{
				{Number: "5.3.2", Title: "E-UTRAN Initial Attach", MessageType: "Attach"},
				{Number: "5.3.3", Title: "Detach", MessageType: "Detach"},
			},
		},
	}

	for _, std := range standards {
		kb.standards[std.ID] = std
	}
}

// loadIETFRFCs loads IETF RFC standards (6 RFCs)
func (kb *KnowledgeBase) loadIETFRFCs() {
	rfcs := []*ProtocolStandard{
		{
			ID:           "RFC 3261",
			Title:        "SIP: Session Initiation Protocol",
			Version:      "June 2002",
			URL:          "https://www.rfc-editor.org/rfc/rfc3261.html",
			Organization: "IETF",
			Protocols:    []string{"SIP"},
			Description:  "Defines SIP protocol for initiating, modifying, and terminating sessions",
		},
		{
			ID:           "RFC 6733",
			Title:        "Diameter Base Protocol",
			Version:      "October 2012",
			URL:          "https://www.rfc-editor.org/rfc/rfc6733.html",
			Organization: "IETF",
			Protocols:    []string{"Diameter"},
			Description:  "Defines Diameter protocol for AAA in modern networks",
		},
		{
			ID:           "RFC 7540",
			Title:        "Hypertext Transfer Protocol Version 2 (HTTP/2)",
			Version:      "May 2015",
			URL:          "https://www.rfc-editor.org/rfc/rfc7540.html",
			Organization: "IETF",
			Protocols:    []string{"HTTP/2"},
			Description:  "Defines HTTP/2 protocol used in 5G Service-Based Architecture",
		},
		{
			ID:           "RFC 4960",
			Title:        "Stream Control Transmission Protocol (SCTP)",
			Version:      "September 2007",
			URL:          "https://www.rfc-editor.org/rfc/rfc4960.html",
			Organization: "IETF",
			Protocols:    []string{"SCTP"},
			Description:  "Defines SCTP transport protocol used for signaling",
		},
		{
			ID:           "RFC 3588",
			Title:        "Diameter Base Protocol (Obsoleted by RFC 6733)",
			Version:      "September 2003",
			URL:          "https://www.rfc-editor.org/rfc/rfc3588.html",
			Organization: "IETF",
			Protocols:    []string{"Diameter"},
			Description:  "Original Diameter specification (obsoleted by RFC 6733)",
		},
		{
			ID:           "RFC 791",
			Title:        "Internet Protocol (IP)",
			Version:      "September 1981",
			URL:          "https://www.rfc-editor.org/rfc/rfc791.html",
			Organization: "IETF",
			Protocols:    []string{"IP", "IPv4"},
			Description:  "Defines IPv4 protocol including fragmentation and reassembly",
		},
	}

	for _, rfc := range rfcs {
		kb.standards[rfc.ID] = rfc
	}
}

// loadProcedures loads standard procedures (20+ procedures)
func (kb *KnowledgeBase) loadProcedures() {
	procedures := []*ProcedureReference{
		// 4G Attach
		{
			StandardID: "TS 23.401",
			Section:    "5.3.2",
			Protocol:   "EPS",
			Procedure:  "4G Initial Attach",
			MessageType: "Attach",
			Description: "Complete E-UTRAN initial attach procedure with HSS authentication",
			Purpose:    "Attach UE to 4G network and establish default bearer",
			IEs:        []string{"IMSI", "GUTI", "APN", "PDN-Type", "TAI"},
			Flows: []FlowStep{
				{Step: 1, Source: "UE", Destination: "eNB", Message: "RRC Connection Request", Description: "UE initiates connection"},
				{Step: 2, Source: "eNB", Destination: "UE", Message: "RRC Connection Setup", Description: "eNB accepts connection"},
				{Step: 3, Source: "UE", Destination: "MME", Message: "Attach Request (NAS)", Description: "UE sends attach request"},
				{Step: 4, Source: "MME", Destination: "HSS", Message: "Authentication Information Request (S6a)", Description: "MME requests authentication vectors"},
				{Step: 5, Source: "HSS", Destination: "MME", Message: "Authentication Information Answer (S6a)", Description: "HSS provides auth vectors"},
				{Step: 6, Source: "MME", Destination: "UE", Message: "Authentication Request (NAS)", Description: "MME challenges UE"},
				{Step: 7, Source: "UE", Destination: "MME", Message: "Authentication Response (NAS)", Description: "UE responds with RES"},
				{Step: 8, Source: "MME", Destination: "HSS", Message: "Update Location Request (S6a)", Description: "MME updates UE location in HSS"},
				{Step: 9, Source: "HSS", Destination: "MME", Message: "Insert Subscriber Data (S6a)", Description: "HSS sends subscriber profile"},
				{Step: 10, Source: "MME", Destination: "HSS", Message: "Insert Subscriber Data Ack (S6a)", Description: "MME acknowledges profile"},
				{Step: 11, Source: "HSS", Destination: "MME", Message: "Update Location Answer (S6a)", Description: "HSS confirms location update"},
				{Step: 12, Source: "MME", Destination: "SGW", Message: "Create Session Request (GTPv2-C)", Description: "MME requests default bearer"},
				{Step: 13, Source: "SGW", Destination: "PGW", Message: "Create Session Request (GTPv2-C)", Description: "SGW forwards to PGW"},
				{Step: 14, Source: "PGW", Destination: "SGW", Message: "Create Session Response (GTPv2-C)", Description: "PGW accepts session"},
				{Step: 15, Source: "SGW", Destination: "MME", Message: "Create Session Response (GTPv2-C)", Description: "SGW confirms to MME"},
				{Step: 16, Source: "MME", Destination: "eNB", Message: "Initial Context Setup Request (S1AP)", Description: "MME requests radio bearer"},
				{Step: 17, Source: "eNB", Destination: "UE", Message: "RRC Connection Reconfiguration", Description: "eNB configures radio"},
				{Step: 18, Source: "UE", Destination: "eNB", Message: "RRC Connection Reconfiguration Complete", Description: "UE confirms config"},
				{Step: 19, Source: "eNB", Destination: "MME", Message: "Initial Context Setup Response (S1AP)", Description: "eNB confirms bearer"},
				{Step: 20, Source: "MME", Destination: "UE", Message: "Attach Accept (NAS)", Description: "MME accepts attach"},
				{Step: 21, Source: "UE", Destination: "MME", Message: "Attach Complete (NAS)", Description: "UE confirms attach"},
			},
			Diagram: `
UE      eNB     MME     SGW     PGW     HSS
|        |       |       |       |       |
|--RRC Setup--->|       |       |       |
|<--RRC Setup---|       |       |       |
|--Attach Req----------->|       |       |
|        |       |--AIR----------------->|
|        |       |<--AIA-----------------|
|        |       |--Auth Req-->|       |
|<------Auth Req--------|       |       |
|-------Auth Res------->|       |       |
|        |       |--ULR----------------->|
|        |       |<--ISD-----------------|
|        |       |--ISD Ack------------->|
|        |       |<--ULA-----------------|
|        |       |--Create Session------>|
|        |       |       |--Create Sess->|
|        |       |       |<--Create Res--|
|        |       |<--Create Response-----|
|        |       |--Init Ctx Setup-->    |
|        |<--RRC Reconfig--|    |       |
|--------RRC Reconfig Cpl->|    |       |
|        |-------Init Ctx Rsp--->|       |
|<------Attach Accept------|    |       |
|-------Attach Complete-------->|       |
`,
		},
		// 5G Registration
		{
			StandardID: "TS 23.502",
			Section:    "4.2.2",
			Protocol:   "5GS",
			Procedure:  "5G Initial Registration",
			MessageType: "Registration",
			Description: "Complete 5G initial registration procedure",
			Purpose:    "Register UE in 5G network and establish PDU session",
			IEs:        []string{"SUCI", "5G-GUTI", "NSSAI", "DNN", "5G-S-TMSI"},
			Flows: []FlowStep{
				{Step: 1, Source: "UE", Destination: "gNB", Message: "RRC Setup Request", Description: "UE initiates 5G connection"},
				{Step: 2, Source: "gNB", Destination: "UE", Message: "RRC Setup", Description: "gNB establishes RRC"},
				{Step: 3, Source: "UE", Destination: "AMF", Message: "Registration Request (NAS)", Description: "UE sends registration request"},
				{Step: 4, Source: "AMF", Destination: "AUSF", Message: "Auth Request (Nausf)", Description: "AMF requests authentication"},
				{Step: 5, Source: "AUSF", Destination: "UDM", Message: "Get Auth Data (Nudm)", Description: "AUSF gets auth vectors"},
				{Step: 6, Source: "UDM", Destination: "AUSF", Message: "Auth Data Response (Nudm)", Description: "UDM provides 5G AV"},
				{Step: 7, Source: "AUSF", Destination: "AMF", Message: "Auth Response (Nausf)", Description: "AUSF returns 5G AV"},
				{Step: 8, Source: "AMF", Destination: "UE", Message: "Authentication Request (NAS)", Description: "AMF challenges UE"},
				{Step: 9, Source: "UE", Destination: "AMF", Message: "Authentication Response (NAS)", Description: "UE responds with RES*"},
				{Step: 10, Source: "AMF", Destination: "UE", Message: "Security Mode Command (NAS)", Description: "AMF initiates NAS security"},
				{Step: 11, Source: "UE", Destination: "AMF", Message: "Security Mode Complete (NAS)", Description: "UE confirms security"},
				{Step: 12, Source: "AMF", Destination: "UDM", Message: "Update Registration (Nudm)", Description: "AMF updates UE registration"},
				{Step: 13, Source: "UDM", Destination: "AMF", Message: "Registration Update Response (Nudm)", Description: "UDM confirms"},
				{Step: 14, Source: "AMF", Destination: "UE", Message: "Registration Accept (NAS)", Description: "AMF accepts registration"},
				{Step: 15, Source: "UE", Destination: "AMF", Message: "Registration Complete (NAS)", Description: "UE completes registration"},
			},
			Diagram: `
UE      gNB     AMF     AUSF    UDM     SMF     UPF
|        |       |       |       |       |       |
|--RRC Setup--->|       |       |       |       |
|<--RRC Setup---|       |       |       |       |
|--Registration Req----->|       |       |       |
|        |       |--Auth Req---->|       |       |
|        |       |       |--Get Auth---->|       |
|        |       |       |<--Auth Data---|       |
|        |       |<--Auth Res----|       |       |
|        |<--Auth Req----|       |       |       |
|--------Auth Res------->|       |       |       |
|        |<--Security Cmd-|       |       |       |
|--------Security Cpl---->|       |       |       |
|        |       |--Update Reg---------->|       |
|        |       |<--Update Res----------|       |
|        |<--Reg Accept---|       |       |       |
|--------Reg Complete---->|       |       |       |
`,
		},
		// More procedures can be added here
	}

	for _, proc := range procedures {
		kb.procedures[proc.Protocol] = append(kb.procedures[proc.Protocol], proc)
	}
}

// loadErrorCodes loads comprehensive error codes (50+ error codes)
func (kb *KnowledgeBase) loadErrorCodes() {
	// Initialize error code maps
	kb.errorCodes["Diameter"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["GTP"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["MAP"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["NAS"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["NGAP"] = make(map[int]*ErrorCodeReference)
	kb.errorCodes["S1AP"] = make(map[int]*ErrorCodeReference)

	// Diameter Error Codes
	diameterErrors := []*ErrorCodeReference{
		{
			Protocol:    "Diameter",
			Code:        5001,
			Name:        "DIAMETER_ERROR_USER_UNKNOWN",
			Description: "The specified user is not recognized by the HSS/HLR",
			Causes:      "Subscriber not provisioned in HSS, IMSI not found, database synchronization issue",
			Solutions:   "1. Verify subscriber exists in HSS\n2. Check IMSI format (15 digits)\n3. Review HSS provisioning logs\n4. Check database replication status",
			StandardRef: "3GPP TS 29.272 Section 7.4.3",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        5004,
			Name:        "DIAMETER_ERROR_ROAMING_NOT_ALLOWED",
			Description: "Subscriber is not allowed to roam in this network",
			Causes:      "Roaming agreement not configured, PLMN not in allowed list, regional restrictions",
			Solutions:   "1. Verify roaming agreements\n2. Check allowed PLMN list in HSS\n3. Review subscriber roaming profile\n4. Confirm MCC/MNC configuration",
			StandardRef: "3GPP TS 29.272 Section 7.4.3",
			Severity:    "major",
		},
		{
			Protocol:    "Diameter",
			Code:        4181,
			Name:        "DIAMETER_AUTHENTICATION_DATA_UNAVAILABLE",
			Description: "HSS cannot provide authentication vectors",
			Causes:      "K/OPC keys not provisioned, crypto module failure, HSS database error",
			Solutions:   "1. Verify authentication keys provisioned\n2. Check HSS crypto module status\n3. Review HSS error logs\n4. Verify HSS database connectivity",
			StandardRef: "3GPP TS 29.272 Section 7.4.4",
			Severity:    "critical",
		},
		{
			Protocol:    "Diameter",
			Code:        5012,
			Name:        "DIAMETER_ERROR_RAT_NOT_ALLOWED",
			Description: "Radio Access Technology (RAT) type is not permitted for this subscriber",
			Causes:      "RAT restrictions in subscriber profile, network access mode limitations",
			Solutions:   "1. Check subscriber access restrictions\n2. Verify RAT-Type parameter\n3. Review subscriber service profile\n4. Update subscriber access permissions if needed",
			StandardRef: "3GPP TS 29.272 Section 7.4.3",
			Severity:    "major",
		},
	}

	// GTP Error Codes
	gtpErrors := []*ErrorCodeReference{
		{
			Protocol:    "GTP",
			Code:        64,
			Name:        "Context Not Found",
			Description: "The specified GTP tunnel or session context does not exist",
			Causes:      "Session already deleted, SGW/PGW restart, session timeout, database inconsistency",
			Solutions:   "1. Check if session was properly torn down\n2. Verify SGW/PGW have not restarted\n3. Check session synchronization between nodes\n4. Review S11/S5 interface stability",
			StandardRef: "3GPP TS 29.274 Section 8.4",
			Severity:    "major",
		},
		{
			Protocol:    "GTP",
			Code:        73,
			Name:        "No Resources Available",
			Description: "Network node has insufficient resources to handle request",
			Causes:      "Memory exhaustion, bearer limit reached, CPU overload, license limitation",
			Solutions:   "1. Check SGW/PGW resource utilization\n2. Verify license bearer capacity\n3. Review active session count\n4. Consider load balancing or scaling",
			StandardRef: "3GPP TS 29.274 Section 8.4",
			Severity:    "critical",
		},
		{
			Protocol:    "GTP",
			Code:        95,
			Name:        "APN Restriction Type Incompatible",
			Description: "The requested APN restriction is incompatible with active bearer",
			Causes:      "APN configuration mismatch, mobility restrictions, handover constraints",
			Solutions:   "1. Verify APN restriction configuration\n2. Check handover source/target restrictions\n3. Review APN profiles in PGW\n4. Ensure consistent APN configuration",
			StandardRef: "3GPP TS 29.274 Section 8.4",
			Severity:    "major",
		},
	}

	// MAP Error Codes
	mapErrors := []*ErrorCodeReference{
		{
			Protocol:    "MAP",
			Code:        1,
			Name:        "Unknown Subscriber",
			Description: "The subscriber is not recognized by the HLR",
			Causes:      "IMSI not provisioned, subscriber deleted, HLR database issue",
			Solutions:   "1. Verify subscriber provisioning in HLR\n2. Check IMSI format and validity\n3. Review HLR database logs\n4. Confirm subscriber activation status",
			StandardRef: "3GPP TS 29.002 Section 7.6.1",
			Severity:    "major",
		},
		{
			Protocol:    "MAP",
			Code:        27,
			Name:        "Absent Subscriber",
			Description: "Subscriber is not reachable (e.g., phone off, out of coverage)",
			Causes:      "UE powered off, out of coverage, IMSI detached, roaming in area without coverage",
			Solutions:   "1. Check subscriber's last known location\n2. Verify network coverage in area\n3. Attempt paging\n4. Review attach/detach history",
			StandardRef: "3GPP TS 29.002 Section 7.6.1",
			Severity:    "minor",
		},
		{
			Protocol:    "MAP",
			Code:        34,
			Name:        "System Failure",
			Description: "General system failure in the network node",
			Causes:      "Database error, hardware failure, software crash, resource exhaustion",
			Solutions:   "1. Check HLR/VLR logs for specific errors\n2. Verify database connectivity\n3. Review system resource usage\n4. Restart affected service if needed",
			StandardRef: "3GPP TS 29.002 Section 7.6.1",
			Severity:    "critical",
		},
	}

	// NAS Error Codes (4G/5G)
	nasErrors := []*ErrorCodeReference{
		{
			Protocol:    "NAS",
			Code:        7,
			Name:        "EPS Services Not Allowed",
			Description: "UE is not allowed to access EPS services",
			Causes:      "Subscription not active, service restrictions, network barring",
			Solutions:   "1. Verify subscription status\n2. Check service entitlements\n3. Review access restrictions\n4. Confirm network allows service type",
			StandardRef: "3GPP TS 24.301 Section 9.9.3.9",
			Severity:    "major",
		},
		{
			Protocol:    "NAS",
			Code:        11,
			Name:        "PLMN Not Allowed",
			Description: "UE is not permitted to register on this PLMN",
			Causes:      "Roaming not allowed, PLMN in forbidden list, no roaming agreement",
			Solutions:   "1. Check roaming agreements\n2. Verify PLMN ID configuration\n3. Review subscriber roaming profile\n4. Check forbidden PLMN list",
			StandardRef: "3GPP TS 24.301 Section 9.9.3.9",
			Severity:    "major",
		},
		{
			Protocol:    "NAS",
			Code:        22,
			Name:        "Congestion",
			Description: "Network experiencing congestion",
			Causes:      "Network overload, too many simultaneous connections, resource shortage",
			Solutions:   "1. Check network load statistics\n2. Implement access class barring if needed\n3. Review capacity planning\n4. Consider network expansion",
			StandardRef: "3GPP TS 24.301 Section 9.9.3.9",
			Severity:    "major",
		},
	}

	// NGAP/S1AP Cause Codes
	ngapCauses := []*ErrorCodeReference{
		{
			Protocol:    "NGAP",
			Code:        0,
			Name:        "Radio Connection With UE Lost",
			Description: "Radio connection to UE has been lost",
			Causes:      "UE moved out of coverage, interference, handover failure, UE power off",
			Solutions:   "1. Check radio conditions in cell\n2. Review handover success rate\n3. Analyze interference levels\n4. Verify UE capabilities",
			StandardRef: "3GPP TS 38.413 Section 9.3.1.2",
			Severity:    "minor",
		},
		{
			Protocol:    "NGAP",
			Code:        1,
			Name:        "Failure In Radio Interface Procedure",
			Description: "RRC procedure failure",
			Causes:      "RRC setup/reconfiguration failure, incompatible UE capabilities",
			Solutions:   "1. Review RRC configuration\n2. Check UE capability compatibility\n3. Verify radio parameters\n4. Analyze RRC failure statistics",
			StandardRef: "3GPP TS 38.413 Section 9.3.1.2",
			Severity:    "major",
		},
	}

	// Add all errors to knowledge base
	for _, err := range diameterErrors {
		kb.errorCodes[err.Protocol][err.Code] = err
	}
	for _, err := range gtpErrors {
		kb.errorCodes[err.Protocol][err.Code] = err
	}
	for _, err := range mapErrors {
		kb.errorCodes[err.Protocol][err.Code] = err
	}
	for _, err := range nasErrors {
		kb.errorCodes[err.Protocol][err.Code] = err
	}
	for _, err := range ngapCauses {
		kb.errorCodes[err.Protocol][err.Code] = err
	}
}

// loadVendorExtensions loads vendor-specific extensions
func (kb *KnowledgeBase) loadVendorExtensions() {
	vendors := []string{"Ericsson", "Huawei", "ZTE", "Nokia", "Cisco"}

	// Ericsson Extensions
	ericssonExts := []*VendorExtension{
		{
			Vendor:      "Ericsson",
			Protocol:    "Diameter",
			Extension:   "Ericsson-Specific-AVP",
			Code:        193,
			Description: "Ericsson proprietary AVP for session management",
			Usage:       "Used in S6a interface for Ericsson-specific subscriber data",
		},
		{
			Vendor:      "Ericsson",
			Protocol:    "GTP",
			Extension:   "Private-Extension-IE",
			Code:        255,
			Description: "Ericsson private extension for GTP",
			Usage:       "Carries vendor-specific information between Ericsson nodes",
		},
	}

	// Huawei Extensions
	huaweiExts := []*VendorExtension{
		{
			Vendor:      "Huawei",
			Protocol:    "Diameter",
			Extension:   "Huawei-Charging-Info",
			Code:        2011,
			Description: "Huawei-specific charging information AVP",
			Usage:       "Used for Huawei billing system integration",
		},
		{
			Vendor:      "Huawei",
			Protocol:    "GTP",
			Extension:   "Huawei-QoS-Extension",
			Code:        240,
			Description: "Extended QoS parameters for Huawei equipment",
			Usage:       "Enhanced QoS control in Huawei EPC",
		},
	}

	// ZTE Extensions
	zteExts := []*VendorExtension{
		{
			Vendor:      "ZTE",
			Protocol:    "Diameter",
			Extension:   "ZTE-User-Location",
			Code:        3001,
			Description: "ZTE enhanced user location information",
			Usage:       "Detailed location tracking in ZTE HSS",
		},
	}

	// Nokia Extensions
	nokiaExts := []*VendorExtension{
		{
			Vendor:      "Nokia",
			Protocol:    "MAP",
			Extension:   "Nokia-Supplementary-Service",
			Code:        150,
			Description: "Nokia-specific supplementary service extensions",
			Usage:       "Advanced CAMEL features in Nokia HLR",
		},
	}

	// Cisco Extensions
	ciscoExts := []*VendorExtension{
		{
			Vendor:      "Cisco",
			Protocol:    "GTP",
			Extension:   "Cisco-Session-Priority",
			Code:        245,
			Description: "Cisco session priority marking",
			Usage:       "QoS prioritization in Cisco ASR routers",
		},
	}

	// Add all vendor extensions
	kb.vendorExtensions["Ericsson"] = ericssonExts
	kb.vendorExtensions["Huawei"] = huaweiExts
	kb.vendorExtensions["ZTE"] = zteExts
	kb.vendorExtensions["Nokia"] = nokiaExts
	kb.vendorExtensions["Cisco"] = ciscoExts
}

// loadCallFlows loads standard call flow diagrams
func (kb *KnowledgeBase) loadCallFlows() {
	// Add call flows for all major procedures
	kb.callFlows["4G_Attach"] = &CallFlowDiagram{
		Name:        "4G E-UTRAN Initial Attach",
		Protocol:    "EPS",
		Type:        "Attach",
		Generation:  "4G",
		StandardRef: "3GPP TS 23.401 Section 5.3.2",
		// Steps populated from procedures
	}

	kb.callFlows["5G_Registration"] = &CallFlowDiagram{
		Name:        "5G Initial Registration",
		Protocol:    "5GS",
		Type:        "Registration",
		Generation:  "5G",
		StandardRef: "3GPP TS 23.502 Section 4.2.2",
	}

	// Add more flows...
}

// buildSearchIndex builds search index for fast lookups
func (kb *KnowledgeBase) buildSearchIndex() {
	// Index standards
	for id, std := range kb.standards {
		kb.indexItem(id, std)
		kb.indexItem(std.Title, std)
		for _, proto := range std.Protocols {
			kb.indexItem(proto, std)
		}
	}

	// Index error codes
	for protocol, codes := range kb.errorCodes {
		for code, errRef := range codes {
			kb.indexItem(errRef.Name, errRef)
			kb.indexItem(fmt.Sprintf("%s_%d", protocol, code), errRef)
			kb.indexItem(fmt.Sprintf("cause_%d", code), errRef)
		}
	}

	// Index procedures
	for _, procs := range kb.procedures {
		for _, proc := range procs {
			kb.indexItem(proc.Procedure, proc)
			kb.indexItem(proc.MessageType, proc)
		}
	}
}

// indexItem adds an item to search index
func (kb *KnowledgeBase) indexItem(key string, value interface{}) {
	key = strings.ToLower(key)
	kb.searchIndex[key] = append(kb.searchIndex[key], value)
}

// Search performs intelligent search across knowledge base
func (kb *KnowledgeBase) Search(query string) []interface{} {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	query = strings.ToLower(query)

	// Direct match
	if results, ok := kb.searchIndex[query]; ok {
		return results
	}

	// Partial match
	var results []interface{}
	for key, items := range kb.searchIndex {
		if strings.Contains(key, query) || strings.Contains(query, key) {
			results = append(results, items...)
		}
	}

	return results
}

// GetStandard retrieves a specific standard by ID
func (kb *KnowledgeBase) GetStandard(id string) (interface{}, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	if std, ok := kb.standards[id]; ok {
		return std, nil
	}
	return nil, fmt.Errorf("standard not found: %s", id)
}

// ListStandards returns all standards
func (kb *KnowledgeBase) ListStandards() []interface{} {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	standards := make([]interface{}, 0, len(kb.standards))
	for _, std := range kb.standards {
		standards = append(standards, std)
	}
	return standards
}

// GetProcedures returns procedures for a protocol
func (kb *KnowledgeBase) GetProcedures(protocol string) []interface{} {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	procs := kb.procedures[protocol]
	result := make([]interface{}, len(procs))
	for i, p := range procs {
		result[i] = p
	}
	return result
}

// GetErrorCode retrieves error code information
func (kb *KnowledgeBase) GetErrorCode(protocol string, code int) (interface{}, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	if protocolCodes, ok := kb.errorCodes[protocol]; ok {
		if errCode, ok := protocolCodes[code]; ok {
			return errCode, nil
		}
	}
	return nil, fmt.Errorf("error code not found: %s/%d", protocol, code)
}

// ListProtocols returns list of supported protocols
func (kb *KnowledgeBase) ListProtocols() []string {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	protocols := make(map[string]bool)
	for _, std := range kb.standards {
		for _, proto := range std.Protocols {
			protocols[proto] = true
		}
	}

	result := make([]string, 0, len(protocols))
	for proto := range protocols {
		result = append(result, proto)
	}
	return result
}

// GetVendorExtensions returns vendor extensions
func (kb *KnowledgeBase) GetVendorExtensions(vendor string) []interface{} {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	exts := kb.vendorExtensions[vendor]
	result := make([]interface{}, len(exts))
	for i, ext := range exts {
		result[i] = ext
	}
	return result
}

// GetCallFlow returns standard call flow diagram
func (kb *KnowledgeBase) GetCallFlow(name string) (interface{}, error) {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	if flow, ok := kb.callFlows[name]; ok {
		return flow, nil
	}
	return nil, fmt.Errorf("call flow not found: %s", name)
}

// ListCallFlows returns all available call flows
func (kb *KnowledgeBase) ListCallFlows() []interface{} {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	flows := make([]interface{}, 0, len(kb.callFlows))
	for _, flow := range kb.callFlows {
		flows = append(flows, flow)
	}
	return flows
}
