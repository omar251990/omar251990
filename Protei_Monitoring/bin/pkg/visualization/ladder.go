package visualization

import (
	"fmt"
	"strings"

	"github.com/protei/monitoring/pkg/correlation"
	"github.com/protei/monitoring/pkg/decoder"
)

// LadderDiagram generates ladder diagrams for message flows
type LadderDiagram struct {
	config *Config
}

// Config holds visualization configuration
type Config struct {
	Format             string
	MaxMessages        int
	OutputPath         string
	AutoLabelNodes     bool
}

// NewLadderDiagram creates a new ladder diagram generator
func NewLadderDiagram(config *Config) *LadderDiagram {
	return &LadderDiagram{
		config: config,
	}
}

// Generate generates a ladder diagram for a session
func (l *LadderDiagram) Generate(session *correlation.Session) (string, error) {
	if l.config.Format == "svg" {
		return l.generateSVG(session)
	}
	return l.generateText(session)
}

// generateSVG generates an SVG ladder diagram
func (l *LadderDiagram) generateSVG(session *correlation.Session) (string, error) {
	// Identify unique network elements
	nodes := l.identifyNodes(session)

	// Build SVG
	width := len(nodes) * 200 + 100
	height := len(session.Messages)*80 + 200

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, width, height)
	svg += `<style>
		.node { fill: #4a90e2; }
		.node-text { fill: white; font-family: Arial; font-size: 12px; }
		.message-line { stroke: #333; stroke-width: 2; marker-end: url(#arrowhead); }
		.message-text { font-family: Arial; font-size: 11px; fill: #333; }
		.success { stroke: #27ae60; }
		.failure { stroke: #e74c3c; }
	</style>
	<defs>
		<marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
			<polygon points="0 0, 10 3, 0 6" fill="#333" />
		</marker>
	</defs>`

	// Draw nodes (network elements)
	y := 50
	nodePositions := make(map[string]int)
	for i, node := range nodes {
		x := 100 + i*200
		nodePositions[node] = x

		svg += fmt.Sprintf(`<rect x="%d" y="%d" width="120" height="40" class="node" rx="5"/>`, x-60, y-20)
		svg += fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="node-text">%s</text>`, x, y+5, truncate(node, 15))

		// Draw vertical line
		svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ccc" stroke-width="1" stroke-dasharray="5,5"/>`,
			x, y+30, x, height-50)
	}

	// Draw messages
	msgY := y + 80
	for _, msg := range session.Messages {
		srcNode := fmt.Sprintf("%s(%s)", msg.Source.Type, msg.Source.IP)
		dstNode := fmt.Sprintf("%s(%s)", msg.Destination.Type, msg.Destination.IP)

		srcX, srcExists := nodePositions[srcNode]
		dstX, dstExists := nodePositions[dstNode]

		if !srcExists || !dstExists {
			continue
		}

		// Determine line style based on result
		lineClass := "message-line"
		if msg.Result == decoder.ResultSuccess {
			lineClass += " success"
		} else if msg.Result == decoder.ResultFailure {
			lineClass += " failure"
		}

		// Draw message line
		svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" class="%s"/>`,
			srcX, msgY, dstX, msgY, lineClass)

		// Draw message label
		labelX := (srcX + dstX) / 2
		labelY := msgY - 5
		svg += fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="message-text">%s</text>`,
			labelX, labelY, truncate(msg.MessageName, 25))

		msgY += 60
		if msgY > height-100 {
			break
		}
	}

	svg += `</svg>`

	return svg, nil
}

// generateText generates a text-based ladder diagram
func (l *LadderDiagram) generateText(session *correlation.Session) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Ladder Diagram for TID: %s\n", session.TID))
	output.WriteString(fmt.Sprintf("Procedure: %s\n", session.Procedure))
	output.WriteString(strings.Repeat("=", 80) + "\n\n")

	nodes := l.identifyNodes(session)

	// Header
	for _, node := range nodes {
		output.WriteString(fmt.Sprintf("%-25s  ", truncate(node, 23)))
	}
	output.WriteString("\n")

	for _, node := range nodes {
		output.WriteString(fmt.Sprintf("%-25s  ", strings.Repeat("|", 1)))
	}
	output.WriteString("\n")

	// Messages
	for i, msg := range session.Messages {
		if i >= l.config.MaxMessages {
			output.WriteString("\n... (truncated) ...\n")
			break
		}

		srcNode := fmt.Sprintf("%s(%s)", msg.Source.Type, msg.Source.IP)
		dstNode := fmt.Sprintf("%s(%s)", msg.Destination.Type, msg.Destination.IP)

		srcIdx := indexOf(nodes, srcNode)
		dstIdx := indexOf(nodes, dstNode)

		if srcIdx == -1 || dstIdx == -1 {
			continue
		}

		// Draw arrow
		for j := range nodes {
			if j == srcIdx {
				if srcIdx < dstIdx {
					output.WriteString("o" + strings.Repeat("-", 24) + "  ")
				} else {
					output.WriteString(strings.Repeat("-", 24) + "o  ")
				}
			} else if j == dstIdx {
				if srcIdx < dstIdx {
					output.WriteString(">  ")
				} else {
					output.WriteString("<  ")
				}
			} else if (j > srcIdx && j < dstIdx) || (j < srcIdx && j > dstIdx) {
				output.WriteString(strings.Repeat("-", 25) + "  ")
			} else {
				output.WriteString(strings.Repeat(" ", 25) + "  ")
			}
		}

		result := "✓"
		if msg.Result == decoder.ResultFailure {
			result = "✗"
		}

		output.WriteString(fmt.Sprintf(" %s %s\n", result, truncate(msg.MessageName, 30)))

		// Node lines
		for range nodes {
			output.WriteString("|" + strings.Repeat(" ", 24) + "  ")
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

// identifyNodes identifies unique network elements in the session
func (l *LadderDiagram) identifyNodes(session *correlation.Session) []string {
	nodeMap := make(map[string]bool)
	var nodes []string

	for _, msg := range session.Messages {
		srcNode := fmt.Sprintf("%s(%s)", msg.Source.Type, msg.Source.IP)
		dstNode := fmt.Sprintf("%s(%s)", msg.Destination.Type, msg.Destination.IP)

		if !nodeMap[srcNode] {
			nodeMap[srcNode] = true
			nodes = append(nodes, srcNode)
		}
		if !nodeMap[dstNode] {
			nodeMap[dstNode] = true
			nodes = append(nodes, dstNode)
		}
	}

	return nodes
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
