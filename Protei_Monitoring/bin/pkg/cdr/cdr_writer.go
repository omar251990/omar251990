package cdr

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CDRFormat represents the output format for CDR files
type CDRFormat string

const (
	FormatCSV  CDRFormat = "csv"
	FormatJSON CDRFormat = "json"
	FormatXML  CDRFormat = "xml"
)

// CDRRotation defines rotation strategy
type CDRRotation struct {
	MaxSizeMB   int           // Rotate when file reaches this size
	MaxDuration time.Duration // Rotate after this duration
	Compress    bool          // Compress rotated files
}

// CDRRecord is the interface that all CDR records must implement
type CDRRecord interface {
	ToCSV() []string
	ToJSON() map[string]interface{}
	GetTimestamp() time.Time
	GetProtocol() string
}

// CDRWriter handles writing CDR records to files with rotation
type CDRWriter struct {
	mu             sync.RWMutex
	baseDir        string
	protocol       string
	format         CDRFormat
	rotation       CDRRotation
	currentFile    *os.File
	currentWriter  io.Writer
	csvWriter      *csv.Writer
	bytesWritten   int64
	fileStartTime  time.Time
	recordCount    int64
	dbTracker      CDRDatabaseTracker
}

// CDRDatabaseTracker tracks CDR files in database
type CDRDatabaseTracker interface {
	TrackCDRFile(filename, protocol string, recordCount int64, fileSize int64, startTime, endTime time.Time) error
}

// CDRConfig holds configuration for CDR writer
type CDRConfig struct {
	BaseDir   string
	Protocol  string
	Format    CDRFormat
	Rotation  CDRRotation
	DBTracker CDRDatabaseTracker
}

// NewCDRWriter creates a new CDR writer for a specific protocol
func NewCDRWriter(config CDRConfig) (*CDRWriter, error) {
	// Create protocol-specific directory
	protocolDir := filepath.Join(config.BaseDir, config.Protocol)
	if err := os.MkdirAll(protocolDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create CDR directory: %w", err)
	}

	writer := &CDRWriter{
		baseDir:   protocolDir,
		protocol:  config.Protocol,
		format:    config.Format,
		rotation:  config.Rotation,
		dbTracker: config.DBTracker,
	}

	// Open initial file
	if err := writer.rotate(); err != nil {
		return nil, fmt.Errorf("failed to create initial CDR file: %w", err)
	}

	return writer, nil
}

// WriteRecord writes a single CDR record
func (w *CDRWriter) WriteRecord(record CDRRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed
	if w.needsRotation() {
		if err := w.rotate(); err != nil {
			return fmt.Errorf("failed to rotate CDR file: %w", err)
		}
	}

	// Write based on format
	var err error
	var bytesWritten int

	switch w.format {
	case FormatCSV:
		err = w.writeCSV(record)
		bytesWritten = len([]byte(fmt.Sprintf("%v", record.ToCSV())))
	case FormatJSON:
		err = w.writeJSON(record)
		bytesWritten = len([]byte(fmt.Sprintf("%v", record.ToJSON())))
	default:
		return fmt.Errorf("unsupported format: %s", w.format)
	}

	if err != nil {
		return err
	}

	w.bytesWritten += int64(bytesWritten)
	w.recordCount++

	return nil
}

// writeCSV writes a record in CSV format
func (w *CDRWriter) writeCSV(record CDRRecord) error {
	if w.csvWriter == nil {
		return fmt.Errorf("CSV writer not initialized")
	}

	return w.csvWriter.Write(record.ToCSV())
}

// writeJSON writes a record in JSON format
func (w *CDRWriter) writeJSON(record CDRRecord) error {
	data, err := json.Marshal(record.ToJSON())
	if err != nil {
		return err
	}

	_, err = w.currentWriter.Write(append(data, '\n'))
	return err
}

// needsRotation checks if file rotation is needed
func (w *CDRWriter) needsRotation() bool {
	// Check size-based rotation
	if w.rotation.MaxSizeMB > 0 {
		maxBytes := int64(w.rotation.MaxSizeMB) * 1024 * 1024
		if w.bytesWritten >= maxBytes {
			return true
		}
	}

	// Check time-based rotation
	if w.rotation.MaxDuration > 0 {
		if time.Since(w.fileStartTime) >= w.rotation.MaxDuration {
			return true
		}
	}

	return false
}

// rotate closes current file and opens a new one
func (w *CDRWriter) rotate() error {
	// Close current file
	if w.currentFile != nil {
		oldFilename := w.currentFile.Name()
		oldRecordCount := w.recordCount
		oldFileSize := w.bytesWritten
		startTime := w.fileStartTime
		endTime := time.Now()

		// Flush CSV writer
		if w.csvWriter != nil {
			w.csvWriter.Flush()
		}

		// Close file
		if err := w.currentFile.Close(); err != nil {
			return err
		}

		// Track in database
		if w.dbTracker != nil {
			if err := w.dbTracker.TrackCDRFile(oldFilename, w.protocol, oldRecordCount, oldFileSize, startTime, endTime); err != nil {
				// Log error but don't fail
				fmt.Printf("Warning: failed to track CDR file in database: %v\n", err)
			}
		}

		// Compress if needed
		if w.rotation.Compress {
			go w.compressFile(oldFilename)
		}
	}

	// Generate new filename
	timestamp := time.Now().Format("20060102_150405")
	var filename string

	switch w.format {
	case FormatCSV:
		filename = filepath.Join(w.baseDir, fmt.Sprintf("%s_%s.csv", w.protocol, timestamp))
	case FormatJSON:
		filename = filepath.Join(w.baseDir, fmt.Sprintf("%s_%s.json", w.protocol, timestamp))
	default:
		filename = filepath.Join(w.baseDir, fmt.Sprintf("%s_%s.cdr", w.protocol, timestamp))
	}

	// Open new file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	w.currentFile = file
	w.currentWriter = file
	w.bytesWritten = 0
	w.recordCount = 0
	w.fileStartTime = time.Now()

	// Initialize format-specific writer
	switch w.format {
	case FormatCSV:
		w.csvWriter = csv.NewWriter(file)
		// Write CSV header
		if err := w.writeCSVHeader(); err != nil {
			return err
		}
	case FormatJSON:
		w.csvWriter = nil
	}

	return nil
}

// writeCSVHeader writes CSV header based on protocol
func (w *CDRWriter) writeCSVHeader() error {
	header := w.getCSVHeader()
	return w.csvWriter.Write(header)
}

// getCSVHeader returns CSV header for the protocol
func (w *CDRWriter) getCSVHeader() []string {
	// Common headers
	common := []string{
		"Timestamp",
		"Protocol",
		"TransactionID",
		"IMSI",
		"MSISDN",
		"OperationType",
		"Result",
		"ResultCode",
		"DurationMs",
	}

	// Protocol-specific headers
	switch w.protocol {
	case "MAP":
		return append(common, "OperationCode", "InvokeID", "SCCP_Called", "SCCP_Calling", "MCC", "MNC", "LAC", "CellID")
	case "CAP":
		return append(common, "ServiceKey", "CallingParty", "CalledParty", "SCCP_Called", "SCCP_Calling", "EventType")
	case "INAP":
		return append(common, "ServiceKey", "CallingParty", "CalledParty", "TriggerType")
	case "Diameter":
		return append(common, "CommandCode", "ApplicationID", "SessionID", "OriginHost", "OriginRealm", "DestinationHost", "DestinationRealm")
	case "GTP":
		return append(common, "MessageType", "TEID", "SequenceNumber", "APN", "PDN_Type", "SourceIP", "DestIP", "BytesUplink", "BytesDownlink")
	case "PFCP":
		return append(common, "MessageType", "SEID", "NodeID", "FSE_ID", "UplinkBytes", "DownlinkBytes")
	case "HTTP2":
		return append(common, "Method", "URI", "StatusCode", "ServiceName", "APIVersion", "SourceNF", "TargetNF")
	case "NGAP":
		return append(common, "ProcedureCode", "AMF_UE_ID", "RAN_UE_ID", "GlobalRAN_ID", "GUAMI", "Cause")
	case "S1AP":
		return append(common, "ProcedureCode", "MME_UE_ID", "eNB_UE_ID", "TAI", "EUTRAN_CGI", "Cause")
	case "NAS":
		return append(common, "MessageType", "SecurityHeader", "ProtocolDiscriminator", "EPS_MobileIdentity", "EMM_Cause", "ESM_Cause")
	default:
		return common
	}
}

// compressFile compresses a CDR file using gzip
func (w *CDRWriter) compressFile(filename string) error {
	// Open source file
	srcFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create gzip file
	gzFilename := filename + ".gz"
	gzFile, err := os.Create(gzFilename)
	if err != nil {
		return err
	}
	defer gzFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	// Copy data
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return err
	}

	// Remove original file
	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}

// Close closes the CDR writer
func (w *CDRWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.csvWriter != nil {
		w.csvWriter.Flush()
	}

	if w.currentFile != nil {
		// Track final file
		if w.dbTracker != nil {
			w.dbTracker.TrackCDRFile(
				w.currentFile.Name(),
				w.protocol,
				w.recordCount,
				w.bytesWritten,
				w.fileStartTime,
				time.Now(),
			)
		}

		return w.currentFile.Close()
	}

	return nil
}

// Flush forces a flush of buffered data
func (w *CDRWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.csvWriter != nil {
		w.csvWriter.Flush()
		return w.csvWriter.Error()
	}

	if w.currentFile != nil {
		return w.currentFile.Sync()
	}

	return nil
}

// GetStats returns statistics about the current CDR file
func (w *CDRWriter) GetStats() (recordCount int64, bytesWritten int64, duration time.Duration) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.recordCount, w.bytesWritten, time.Since(w.fileStartTime)
}
