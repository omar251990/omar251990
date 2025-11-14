package flows

import (
	"fmt"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// ProcedureTemplate represents a standard 3GPP procedure flow
type ProcedureTemplate struct {
	Name         string                `json:"name"`           // e.g., "4G Attach"
	Description  string                `json:"description"`    // What this procedure does
	Standard     string                `json:"standard"`       // e.g., "TS 23.401"
	Section      string                `json:"section"`        // e.g., "5.3.2.1"
	Generation   string                `json:"generation"`     // "2G", "3G", "4G", "5G"
	Steps        []*ProcedureStep      `json:"steps"`          // Expected message sequence
	Interfaces   []string              `json:"interfaces"`     // S6a, S11, S1-MME, etc.
	Duration     time.Duration         `json:"duration"`       // Expected duration
	Variants     []*ProcedureVariant   `json:"variants"`       // Different paths (success, failure)
}

// ProcedureStep represents one step in a procedure
type ProcedureStep struct {
	Number      int      `json:"number"`       // Step number
	Message     string   `json:"message"`      // Message name (e.g., "Attach Request")
	Direction   string   `json:"direction"`    // "UE->eNB", "MME->HSS"
	Interface   string   `json:"interface"`    // "S1-MME", "S6a", etc.
	Protocol    string   `json:"protocol"`     // "NAS", "S1AP", "Diameter"
	Mandatory   bool     `json:"mandatory"`    // Is this step required?
	Expected    bool     `json:"expected"`     // Expected in normal flow?
	IEs         []string `json:"ies"`          // Expected Information Elements
	Description string   `json:"description"`  // What happens in this step
}

// ProcedureVariant represents different execution paths
type ProcedureVariant struct {
	Name        string            `json:"name"`        // "Success", "IMSI Unknown", "Roaming Rejected"
	Probability float64           `json:"probability"` // Expected occurrence %
	Steps       []*ProcedureStep  `json:"steps"`       // Steps for this variant
	Outcome     string            `json:"outcome"`     // "success", "failure"
	Cause       string            `json:"cause"`       // Cause if failure
}

// CapturedFlow represents actual captured traffic flow
type CapturedFlow struct {
	ID           string                   `json:"id"`
	Procedure    string                   `json:"procedure"`     // Detected procedure name
	IMSI         string                   `json:"imsi"`
	MSISDN       string                   `json:"msisdn"`
	StartTime    time.Time                `json:"start_time"`
	EndTime      time.Time                `json:"end_time"`
	Duration     time.Duration            `json:"duration"`
	Messages     []*decoder.Message       `json:"messages"`      // Actual messages
	Steps        []*CapturedStep          `json:"steps"`         // Mapped to template steps
	Result       string                   `json:"result"`        // "success", "failure", "partial"
	Deviations   []*FlowDeviation         `json:"deviations"`    // Deviations from standard
	Completeness float64                  `json:"completeness"`  // % of expected steps seen
}

// CapturedStep maps a real message to a template step
type CapturedStep struct {
	TemplateStep *ProcedureStep     `json:"template_step"`
	ActualMsg    *decoder.Message   `json:"actual_msg"`
	Matched      bool               `json:"matched"`       // Does it match template?
	Latency      time.Duration      `json:"latency"`       // Time from previous step
	Missing      bool               `json:"missing"`       // Expected but not found
}

// FlowDeviation represents a deviation from standard flow
type FlowDeviation struct {
	Type        string   `json:"type"`        // "missing_step", "unexpected_msg", "wrong_order", "timeout"
	Severity    string   `json:"severity"`    // "critical", "major", "minor"
	Step        int      `json:"step"`        // Step number where deviation occurred
	Expected    string   `json:"expected"`    // What was expected
	Actual      string   `json:"actual"`      // What was seen
	Impact      string   `json:"impact"`      // Impact description
	Standard    string   `json:"standard"`    // 3GPP reference
	Explanation string   `json:"explanation"` // Human-readable explanation
}

// FlowReconstructor reconstructs signaling flows from captured messages
type FlowReconstructor struct {
	templates map[string]*ProcedureTemplate
}

// NewFlowReconstructor creates a new flow reconstructor
func NewFlowReconstructor() *FlowReconstructor {
	fr := &FlowReconstructor{
		templates: make(map[string]*ProcedureTemplate),
	}
	fr.loadStandardProcedures()
	return fr
}

// Load standard 3GPP procedures
func (fr *FlowReconstructor) loadStandardProcedures() {
	// 4G Attach Procedure
	fr.templates["4G_Attach"] = &ProcedureTemplate{
		Name:        "4G Attach Procedure",
		Description: "Initial attachment of UE to LTE/EPS network",
		Standard:    "TS 23.401",
		Section:     "5.3.2.1",
		Generation:  "4G",
		Duration:    2 * time.Second,
		Interfaces:  []string{"S1-MME", "S6a", "S11", "S5/S8"},
		Steps: []*ProcedureStep{
			{
				Number:      1,
				Message:     "Attach Request",
				Direction:   "UE->MME",
				Interface:   "S1-MME",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "UE Network Capability", "PDN Type"},
				Description: "UE initiates attach with IMSI and capabilities",
			},
			{
				Number:      2,
				Message:     "Authentication Information Request",
				Direction:   "MME->HSS",
				Interface:   "S6a",
				Protocol:    "Diameter",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "Visited PLMN ID", "Number of Requested Vectors"},
				Description: "MME requests authentication vectors from HSS",
			},
			{
				Number:      3,
				Message:     "Authentication Information Answer",
				Direction:   "HSS->MME",
				Interface:   "S6a",
				Protocol:    "Diameter",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Authentication Vectors (RAND, AUTN, XRES, KASME)"},
				Description: "HSS provides authentication vectors",
			},
			{
				Number:      4,
				Message:     "Authentication Request",
				Direction:   "MME->UE",
				Interface:   "S1-MME",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"RAND", "AUTN"},
				Description: "MME challenges UE with authentication parameters",
			},
			{
				Number:      5,
				Message:     "Authentication Response",
				Direction:   "UE->MME",
				Interface:   "S1-MME",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"RES"},
				Description: "UE responds with authentication result",
			},
			{
				Number:      6,
				Message:     "Update Location Request",
				Direction:   "MME->HSS",
				Interface:   "S6a",
				Protocol:    "Diameter",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "Visited PLMN ID", "RAT Type", "ULR Flags"},
				Description: "MME updates subscriber location in HSS",
			},
			{
				Number:      7,
				Message:     "Update Location Answer",
				Direction:   "HSS->MME",
				Interface:   "S6a",
				Protocol:    "Diameter",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Subscription Data (APN, QoS, etc.)"},
				Description: "HSS provides subscription data",
			},
			{
				Number:      8,
				Message:     "Create Session Request",
				Direction:   "MME->SGW->PGW",
				Interface:   "S11/S5",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "APN", "RAT Type", "Bearer Contexts"},
				Description: "MME requests session creation with default bearer",
			},
			{
				Number:      9,
				Message:     "Create Session Response",
				Direction:   "PGW->SGW->MME",
				Interface:   "S5/S11",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Cause", "PDN Address", "Bearer Contexts"},
				Description: "PGW confirms session creation and assigns IP",
			},
			{
				Number:      10,
				Message:     "Initial Context Setup Request",
				Direction:   "MME->eNB",
				Interface:   "S1-MME",
				Protocol:    "S1AP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"E-RAB to be Setup", "Security Context", "UE Aggregate Max Bitrate"},
				Description: "MME requests eNB to setup radio resources",
			},
			{
				Number:      11,
				Message:     "Initial Context Setup Response",
				Direction:   "eNB->MME",
				Interface:   "S1-MME",
				Protocol:    "S1AP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"E-RAB Setup List"},
				Description: "eNB confirms radio bearer establishment",
			},
			{
				Number:      12,
				Message:     "Attach Accept",
				Direction:   "MME->UE",
				Interface:   "S1-MME",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"GUTI", "TAI List"},
				Description: "MME accepts attach and provides GUTI",
			},
			{
				Number:      13,
				Message:     "Attach Complete",
				Direction:   "UE->MME",
				Interface:   "S1-MME",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"ESM Message Container"},
				Description: "UE confirms attach completion",
			},
		},
	}

	// 5G Registration Procedure
	fr.templates["5G_Registration"] = &ProcedureTemplate{
		Name:        "5G Registration Procedure",
		Description: "Initial registration of UE to 5G network",
		Standard:    "TS 23.502",
		Section:     "4.2.2.2.2",
		Generation:  "5G",
		Duration:    2 * time.Second,
		Interfaces:  []string{"N1", "N2", "Namf", "Nudm"},
		Steps: []*ProcedureStep{
			{
				Number:      1,
				Message:     "Registration Request",
				Direction:   "UE->AMF",
				Interface:   "N1",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"SUCI/SUPI", "Registration Type", "5G Capabilities"},
				Description: "UE initiates registration with identity",
			},
			{
				Number:      2,
				Message:     "Nudm_UECM_Registration",
				Direction:   "AMF->UDM",
				Interface:   "Nudm",
				Protocol:    "HTTP/2",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"SUPI", "AMF Address"},
				Description: "AMF registers with UDM",
			},
			{
				Number:      3,
				Message:     "Nudm_SDM_Get",
				Direction:   "AMF->UDM",
				Interface:   "Nudm",
				Protocol:    "HTTP/2",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"SUPI", "Data Set"},
				Description: "AMF retrieves subscription data",
			},
			{
				Number:      4,
				Message:     "Registration Accept",
				Direction:   "AMF->UE",
				Interface:   "N1",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"5G-GUTI", "TAI List"},
				Description: "AMF accepts registration",
			},
		},
	}

	// GTP Create Session Procedure
	fr.templates["GTP_Create_Session"] = &ProcedureTemplate{
		Name:        "GTP Create Session Procedure",
		Description: "Establishment of GTP tunnel for data session",
		Standard:    "TS 29.274",
		Section:     "7.2.1",
		Generation:  "4G",
		Duration:    500 * time.Millisecond,
		Interfaces:  []string{"S11", "S5/S8"},
		Steps: []*ProcedureStep{
			{
				Number:      1,
				Message:     "Create Session Request",
				Direction:   "MME/SGSN->SGW",
				Interface:   "S11/S4",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "APN", "Bearer Contexts", "PDN Type"},
				Description: "Request to create GTP session",
			},
			{
				Number:      2,
				Message:     "Create Session Request",
				Direction:   "SGW->PGW",
				Interface:   "S5/S8",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "APN", "Bearer Contexts"},
				Description: "SGW forwards request to PGW",
			},
			{
				Number:      3,
				Message:     "Create Session Response",
				Direction:   "PGW->SGW",
				Interface:   "S5/S8",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Cause", "PDN Address", "Bearer Contexts"},
				Description: "PGW responds with session details",
			},
			{
				Number:      4,
				Message:     "Create Session Response",
				Direction:   "SGW->MME/SGSN",
				Interface:   "S11/S4",
				Protocol:    "GTPv2-C",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Cause", "Bearer Contexts"},
				Description: "SGW forwards response to MME",
			},
		},
	}

	// MAP Update Location
	fr.templates["MAP_Update_Location"] = &ProcedureTemplate{
		Name:        "MAP Update Location",
		Description: "Location update in 2G/3G network",
		Standard:    "TS 29.002",
		Section:     "7.3",
		Generation:  "2G/3G",
		Duration:    1 * time.Second,
		Interfaces:  []string{"D", "C"},
		Steps: []*ProcedureStep{
			{
				Number:      1,
				Message:     "Update Location (invoke)",
				Direction:   "VLR->HLR",
				Interface:   "D",
				Protocol:    "MAP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"IMSI", "MSC Number", "VLR Number"},
				Description: "VLR sends location update to HLR",
			},
			{
				Number:      2,
				Message:     "Update Location (return result)",
				Direction:   "HLR->VLR",
				Interface:   "D",
				Protocol:    "MAP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"HLR Number", "Subscription Data"},
				Description: "HLR confirms and sends subscriber data",
			},
		},
	}

	// PDU Session Establishment (5G)
	fr.templates["5G_PDU_Session"] = &ProcedureTemplate{
		Name:        "5G PDU Session Establishment",
		Description: "PDU session creation in 5G network",
		Standard:    "TS 23.502",
		Section:     "4.3.2.2.1",
		Generation:  "5G",
		Duration:    1 * time.Second,
		Interfaces:  []string{"N1", "N2", "N4", "N11"},
		Steps: []*ProcedureStep{
			{
				Number:      1,
				Message:     "PDU Session Establishment Request",
				Direction:   "UE->AMF->SMF",
				Interface:   "N1",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"PDU Session ID", "DNN", "S-NSSAI"},
				Description: "UE requests PDU session establishment",
			},
			{
				Number:      2,
				Message:     "Nsmf_PDUSession_CreateSMContext",
				Direction:   "AMF->SMF",
				Interface:   "N11",
				Protocol:    "HTTP/2",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"SUPI", "DNN", "S-NSSAI"},
				Description: "AMF requests SMF to create SM context",
			},
			{
				Number:      3,
				Message:     "PFCP Session Establishment Request",
				Direction:   "SMF->UPF",
				Interface:   "N4",
				Protocol:    "PFCP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Node ID", "PDR", "FAR", "QER"},
				Description: "SMF creates forwarding rules in UPF",
			},
			{
				Number:      4,
				Message:     "PFCP Session Establishment Response",
				Direction:   "UPF->SMF",
				Interface:   "N4",
				Protocol:    "PFCP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"Cause", "F-SEID"},
				Description: "UPF confirms session establishment",
			},
			{
				Number:      5,
				Message:     "PDU Session Resource Setup Request",
				Direction:   "AMF->gNB",
				Interface:   "N2",
				Protocol:    "NGAP",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"PDU Session Resource Setup List", "QoS Flows"},
				Description: "AMF requests gNB to setup radio resources",
			},
			{
				Number:      6,
				Message:     "PDU Session Establishment Accept",
				Direction:   "SMF->AMF->UE",
				Interface:   "N1",
				Protocol:    "NAS",
				Mandatory:   true,
				Expected:    true,
				IEs:         []string{"PDU Session ID", "QoS Rules"},
				Description: "SMF accepts PDU session",
			},
		},
	}
}

// ReconstructFlow attempts to match captured messages to a procedure template
func (fr *FlowReconstructor) ReconstructFlow(messages []*decoder.Message) *CapturedFlow {
	if len(messages) == 0 {
		return nil
	}

	// Detect which procedure this is
	procedureName := fr.detectProcedure(messages)
	template := fr.templates[procedureName]

	if template == nil {
		// Unknown procedure
		return &CapturedFlow{
			ID:        fmt.Sprintf("FLOW_%d", time.Now().Unix()),
			Procedure: "Unknown",
			Messages:  messages,
			Result:    "unknown",
		}
	}

	flow := &CapturedFlow{
		ID:         fmt.Sprintf("FLOW_%s_%d", procedureName, time.Now().Unix()),
		Procedure:  template.Name,
		Messages:   messages,
		Steps:      make([]*CapturedStep, 0),
		Deviations: make([]*FlowDeviation, 0),
	}

	// Extract identifiers
	for _, msg := range messages {
		if imsi, ok := msg.Attributes["imsi"].(string); ok && flow.IMSI == "" {
			flow.IMSI = imsi
		}
		if msisdn, ok := msg.Attributes["msisdn"].(string); ok && flow.MSISDN == "" {
			flow.MSISDN = msisdn
		}
	}

	flow.StartTime = messages[0].Timestamp
	flow.EndTime = messages[len(messages)-1].Timestamp
	flow.Duration = flow.EndTime.Sub(flow.StartTime)

	// Match messages to template steps
	flow.Steps = fr.matchSteps(messages, template)

	// Detect deviations
	flow.Deviations = fr.detectDeviations(flow.Steps, template)

	// Calculate completeness
	matchedSteps := 0
	for _, step := range flow.Steps {
		if step.Matched && !step.Missing {
			matchedSteps++
		}
	}
	flow.Completeness = float64(matchedSteps) / float64(len(template.Steps)) * 100.0

	// Determine result
	if flow.Completeness >= 90.0 && len(flow.Deviations) == 0 {
		flow.Result = "success"
	} else if flow.Completeness < 50.0 {
		flow.Result = "failure"
	} else {
		flow.Result = "partial"
	}

	return flow
}

// detectProcedure detects which standard procedure is being executed
func (fr *FlowReconstructor) detectProcedure(messages []*decoder.Message) string {
	// Simple detection based on message patterns
	protocols := make(map[string]bool)
	messageTypes := make(map[string]bool)

	for _, msg := range messages {
		protocols[msg.Protocol] = true
		messageTypes[msg.Type] = true
	}

	// 4G Attach: NAS + S1AP + Diameter S6a + GTP
	if protocols["NAS"] && protocols["S1AP"] && protocols["Diameter"] && protocols["GTP"] {
		if messageTypes["Attach Request"] || messageTypes["Attach Accept"] {
			return "4G_Attach"
		}
	}

	// 5G Registration: NAS 5G + NGAP + HTTP/2
	if protocols["NAS"] && (protocols["NGAP"] || protocols["HTTP"]) {
		if messageTypes["Registration Request"] || messageTypes["Registration Accept"] {
			return "5G_Registration"
		}
	}

	// GTP Create Session
	if protocols["GTP"] && (messageTypes["Create Session Request"] || messageTypes["Create Session Response"]) {
		return "GTP_Create_Session"
	}

	// MAP Update Location
	if protocols["MAP"] && messageTypes["Update Location"] {
		return "MAP_Update_Location"
	}

	// 5G PDU Session
	if protocols["NAS"] && protocols["PFCP"] {
		if messageTypes["PDU Session Establishment Request"] {
			return "5G_PDU_Session"
		}
	}

	return "Unknown"
}

// matchSteps matches captured messages to template steps
func (fr *FlowReconstructor) matchSteps(messages []*decoder.Message, template *ProcedureTemplate) []*CapturedStep {
	steps := make([]*CapturedStep, 0)
	usedMessages := make(map[int]bool)

	// Try to match each template step with a message
	for _, templateStep := range template.Steps {
		matched := false
		var prevTime time.Time
		if len(steps) > 0 && steps[len(steps)-1].ActualMsg != nil {
			prevTime = steps[len(steps)-1].ActualMsg.Timestamp
		}

		for i, msg := range messages {
			if usedMessages[i] {
				continue
			}

			// Simple matching: protocol and message type
			if msg.Protocol == templateStep.Protocol && msg.Type == templateStep.Message {
				latency := time.Duration(0)
				if !prevTime.IsZero() {
					latency = msg.Timestamp.Sub(prevTime)
				}

				steps = append(steps, &CapturedStep{
					TemplateStep: templateStep,
					ActualMsg:    msg,
					Matched:      true,
					Latency:      latency,
					Missing:      false,
				})

				usedMessages[i] = true
				matched = true
				break
			}
		}

		// If not matched and mandatory, mark as missing
		if !matched && templateStep.Mandatory {
			steps = append(steps, &CapturedStep{
				TemplateStep: templateStep,
				ActualMsg:    nil,
				Matched:      false,
				Missing:      true,
			})
		}
	}

	return steps
}

// detectDeviations detects deviations from standard flow
func (fr *FlowReconstructor) detectDeviations(steps []*CapturedStep, template *ProcedureTemplate) []*FlowDeviation {
	deviations := make([]*FlowDeviation, 0)

	for i, step := range steps {
		// Missing mandatory step
		if step.Missing && step.TemplateStep.Mandatory {
			deviations = append(deviations, &FlowDeviation{
				Type:        "missing_step",
				Severity:    "critical",
				Step:        i + 1,
				Expected:    step.TemplateStep.Message,
				Actual:      "Not received",
				Impact:      "Procedure cannot complete successfully",
				Standard:    template.Standard + " Section " + template.Section,
				Explanation: fmt.Sprintf("Mandatory step %d (%s) is missing. This violates 3GPP %s.", i+1, step.TemplateStep.Message, template.Standard),
			})
		}

		// High latency
		if step.Latency > 5*time.Second {
			deviations = append(deviations, &FlowDeviation{
				Type:        "timeout",
				Severity:    "major",
				Step:        i + 1,
				Expected:    "< 5s",
				Actual:      fmt.Sprintf("%.2fs", step.Latency.Seconds()),
				Impact:      "Slow procedure execution may cause UE timeout",
				Standard:    template.Standard,
				Explanation: fmt.Sprintf("Step %d took %.2fs which exceeds recommended timeout.", i+1, step.Latency.Seconds()),
			})
		}
	}

	return deviations
}

// GetTemplate returns a procedure template by name
func (fr *FlowReconstructor) GetTemplate(name string) *ProcedureTemplate {
	return fr.templates[name]
}

// ListTemplates returns all available templates
func (fr *FlowReconstructor) ListTemplates() []*ProcedureTemplate {
	templates := make([]*ProcedureTemplate, 0, len(fr.templates))
	for _, template := range fr.templates {
		templates = append(templates, template)
	}
	return templates
}
