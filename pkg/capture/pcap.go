package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/decoder"
)

// Engine handles packet capture from files and live interfaces
type Engine struct {
	config     *Config
	processors []Processor
	running    bool
	wg         sync.WaitGroup
	stopCh     chan struct{}
	mu         sync.Mutex
}

// Config holds capture configuration
type Config struct {
	Sources          []SourceConfig
	BufferSize       int
	Workers          int
	BatchSize        int
	OutputChannel    chan *CapturedPacket
}

// SourceConfig represents a capture source
type SourceConfig struct {
	Type      string // pcap_file, pcap_live
	Path      string
	Watch     bool
	Recursive bool
	Pattern   string
	Interface string
	Snaplen   int
	Promisc   bool
}

// Processor processes captured packets
type Processor interface {
	Process(*CapturedPacket) error
}

// CapturedPacket represents a captured packet with metadata
type CapturedPacket struct {
	Timestamp   time.Time
	Data        []byte
	Length      int
	SourceIP    string
	DestIP      string
	SourcePort  uint16
	DestPort    uint16
	Protocol    string
	InterfaceName string
	Metadata    *decoder.Metadata
}

// NewEngine creates a new capture engine
func NewEngine(config *Config) *Engine {
	return &Engine{
		config:     config,
		processors: make([]Processor, 0),
		stopCh:     make(chan struct{}),
	}
}

// RegisterProcessor registers a packet processor
func (e *Engine) RegisterProcessor(p Processor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processors = append(e.processors, p)
}

// Start starts the capture engine
func (e *Engine) Start() error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return fmt.Errorf("capture engine already running")
	}
	e.running = true
	e.mu.Unlock()

	// Start workers for each source
	for _, source := range e.config.Sources {
		if source.Type == "pcap_file" {
			e.wg.Add(1)
			go e.captureFromFiles(source)
		} else if source.Type == "pcap_live" {
			e.wg.Add(1)
			go e.captureLive(source)
		}
	}

	return nil
}

// Stop stops the capture engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return fmt.Errorf("capture engine not running")
	}
	e.running = false
	e.mu.Unlock()

	close(e.stopCh)
	e.wg.Wait()

	return nil
}

// captureFromFiles captures packets from PCAP files
func (e *Engine) captureFromFiles(source SourceConfig) {
	defer e.wg.Done()

	if source.Watch {
		// Watch directory for new files
		e.watchDirectory(source)
	} else {
		// Process existing files once
		e.processDirectory(source)
	}
}

// watchDirectory monitors a directory for new PCAP files
func (e *Engine) watchDirectory(source SourceConfig) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	processedFiles := make(map[string]bool)

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			files, err := e.findPCAPFiles(source)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !processedFiles[file] {
					e.processPCAPFile(file, source)
					processedFiles[file] = true
				}
			}
		}
	}
}

// processDirectory processes all PCAP files in directory once
func (e *Engine) processDirectory(source SourceConfig) {
	files, err := e.findPCAPFiles(source)
	if err != nil {
		return
	}

	for _, file := range files {
		select {
		case <-e.stopCh:
			return
		default:
			e.processPCAPFile(file, source)
		}
	}
}

// findPCAPFiles finds PCAP files matching the pattern
func (e *Engine) findPCAPFiles(source SourceConfig) ([]string, error) {
	var files []string

	if source.Recursive {
		err := filepath.Walk(source.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && matchPattern(info.Name(), source.Pattern) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(source.Path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() && matchPattern(entry.Name(), source.Pattern) {
				files = append(files, filepath.Join(source.Path, entry.Name()))
			}
		}
	}

	return files, nil
}

// processPCAPFile processes a single PCAP file
func (e *Engine) processPCAPFile(filename string, source SourceConfig) error {
	// Open PCAP file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open PCAP file: %w", err)
	}
	defer file.Close()

	// Read PCAP header (24 bytes)
	header := make([]byte, 24)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read PCAP header: %w", err)
	}

	// Verify PCAP magic number
	magicNumber := uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16 | uint32(header[3])<<24
	if magicNumber != 0xa1b2c3d4 && magicNumber != 0xd4c3b2a1 {
		return fmt.Errorf("invalid PCAP magic number: %x", magicNumber)
	}

	// Read packets
	for {
		select {
		case <-e.stopCh:
			return nil
		default:
			packet, err := e.readPCAPPacket(file)
			if err != nil {
				return nil // End of file or error
			}

			// Process packet
			e.processPacket(packet)
		}
	}
}

// readPCAPPacket reads a single packet from PCAP file
func (e *Engine) readPCAPPacket(file *os.File) (*CapturedPacket, error) {
	// Read packet header (16 bytes)
	packetHeader := make([]byte, 16)
	if _, err := file.Read(packetHeader); err != nil {
		return nil, err
	}

	// Extract timestamp and length
	tsSec := uint32(packetHeader[0]) | uint32(packetHeader[1])<<8 | uint32(packetHeader[2])<<16 | uint32(packetHeader[3])<<24
	tsUsec := uint32(packetHeader[4]) | uint32(packetHeader[5])<<8 | uint32(packetHeader[6])<<16 | uint32(packetHeader[7])<<24
	capturedLen := uint32(packetHeader[8]) | uint32(packetHeader[9])<<8 | uint32(packetHeader[10])<<16 | uint32(packetHeader[11])<<24

	// Read packet data
	data := make([]byte, capturedLen)
	if _, err := file.Read(data); err != nil {
		return nil, err
	}

	// Parse Ethernet frame to extract IP/TCP/UDP headers
	packet := &CapturedPacket{
		Timestamp: time.Unix(int64(tsSec), int64(tsUsec)*1000),
		Data:      data,
		Length:    int(capturedLen),
	}

	e.parsePacketHeaders(packet)

	return packet, nil
}

// parsePacketHeaders parses Ethernet/IP/TCP/UDP headers
func (e *Engine) parsePacketHeaders(packet *CapturedPacket) {
	data := packet.Data

	// Skip Ethernet header (14 bytes)
	if len(data) < 14 {
		return
	}

	etherType := uint16(data[12])<<8 | uint16(data[13])
	if etherType != 0x0800 { // IPv4
		return
	}

	// Parse IPv4 header
	if len(data) < 34 {
		return
	}

	ipHeader := data[14:]
	ipVersion := (ipHeader[0] >> 4) & 0x0F
	if ipVersion != 4 {
		return
	}

	ipHeaderLen := int(ipHeader[0]&0x0F) * 4
	protocol := ipHeader[9]

	packet.SourceIP = fmt.Sprintf("%d.%d.%d.%d", ipHeader[12], ipHeader[13], ipHeader[14], ipHeader[15])
	packet.DestIP = fmt.Sprintf("%d.%d.%d.%d", ipHeader[16], ipHeader[17], ipHeader[18], ipHeader[19])

	// Parse transport layer
	if len(ipHeader) < ipHeaderLen+8 {
		return
	}

	transportHeader := ipHeader[ipHeaderLen:]

	switch protocol {
	case 6: // TCP
		packet.Protocol = "TCP"
		packet.SourcePort = uint16(transportHeader[0])<<8 | uint16(transportHeader[1])
		packet.DestPort = uint16(transportHeader[2])<<8 | uint16(transportHeader[3])
	case 17: // UDP
		packet.Protocol = "UDP"
		packet.SourcePort = uint16(transportHeader[0])<<8 | uint16(transportHeader[1])
		packet.DestPort = uint16(transportHeader[2])<<8 | uint16(transportHeader[3])
	case 132: // SCTP
		packet.Protocol = "SCTP"
		packet.SourcePort = uint16(transportHeader[0])<<8 | uint16(transportHeader[1])
		packet.DestPort = uint16(transportHeader[2])<<8 | uint16(transportHeader[3])
	}

	// Build decoder metadata
	packet.Metadata = &decoder.Metadata{
		CaptureTime:    packet.Timestamp,
		SourceIP:       packet.SourceIP,
		DestIP:         packet.DestIP,
		SourcePort:     packet.SourcePort,
		DestPort:       packet.DestPort,
		TransportProto: packet.Protocol,
	}
}

// processPacket sends packet to all registered processors
func (e *Engine) processPacket(packet *CapturedPacket) {
	e.mu.Lock()
	processors := e.processors
	e.mu.Unlock()

	for _, processor := range processors {
		processor.Process(packet)
	}

	// Send to output channel if configured
	if e.config.OutputChannel != nil {
		select {
		case e.config.OutputChannel <- packet:
		default:
			// Channel full, drop packet
		}
	}
}

// captureLive captures packets from live interface (placeholder)
func (e *Engine) captureLive(source SourceConfig) {
	defer e.wg.Done()

	// Live capture would use libpcap/gopacket or raw sockets
	// Placeholder implementation
	<-e.stopCh
}

// matchPattern checks if filename matches pattern
func matchPattern(filename, pattern string) bool {
	matched, _ := filepath.Match(pattern, filename)
	return matched
}
