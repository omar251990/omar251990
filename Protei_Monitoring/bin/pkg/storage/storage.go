package storage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/protei/monitoring/pkg/correlation"
	"github.com/protei/monitoring/pkg/decoder"
)

// Storage handles all data persistence
type Storage struct {
	config       *Config
	eventWriter  *EventWriter
	cdrWriter    *CDRWriter
	mu           sync.Mutex
}

// Config holds storage configuration
type Config struct {
	EventsEnabled     bool
	EventsPath        string
	EventsFormat      string
	CDREnabled        bool
	CDRPath           string
	CDRFormat         string
	CDRFields         []string
	RetentionDays     int
}

// EventWriter writes decoded messages to JSONL files
type EventWriter struct {
	basePath string
	file     *os.File
	encoder  *json.Encoder
	lastRotate time.Time
	mu       sync.Mutex
}

// CDRWriter writes Call Detail Records to CSV files
type CDRWriter struct {
	basePath  string
	file      *os.File
	writer    *csv.Writer
	fields    []string
	lastRotate time.Time
	mu        sync.Mutex
}

// NewStorage creates a new storage instance
func NewStorage(config *Config) (*Storage, error) {
	storage := &Storage{
		config: config,
	}

	// Initialize event writer
	if config.EventsEnabled {
		eventWriter, err := NewEventWriter(config.EventsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create event writer: %w", err)
		}
		storage.eventWriter = eventWriter
	}

	// Initialize CDR writer
	if config.CDREnabled {
		cdrWriter, err := NewCDRWriter(config.CDRPath, config.CDRFields)
		if err != nil {
			return nil, fmt.Errorf("failed to create CDR writer: %w", err)
		}
		storage.cdrWriter = cdrWriter
	}

	// Start cleanup routine
	go storage.cleanupRoutine()

	return storage, nil
}

// NewEventWriter creates a new event writer
func NewEventWriter(basePath string) (*EventWriter, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	writer := &EventWriter{
		basePath:   basePath,
		lastRotate: time.Now(),
	}

	if err := writer.rotate(); err != nil {
		return nil, err
	}

	return writer, nil
}

// NewCDRWriter creates a new CDR writer
func NewCDRWriter(basePath string, fields []string) (*CDRWriter, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	writer := &CDRWriter{
		basePath:   basePath,
		fields:     fields,
		lastRotate: time.Now(),
	}

	if err := writer.rotate(); err != nil {
		return nil, err
	}

	return writer, nil
}

// WriteEvent writes a decoded message as an event
func (s *Storage) WriteEvent(msg *decoder.Message) error {
	if !s.config.EventsEnabled || s.eventWriter == nil {
		return nil
	}

	return s.eventWriter.Write(msg)
}

// WriteCDR writes a CDR from a completed session
func (s *Storage) WriteCDR(session *correlation.Session) error {
	if !s.config.CDREnabled || s.cdrWriter == nil {
		return nil
	}

	return s.cdrWriter.Write(session)
}

// EventWriter methods

// Write writes a message to the events file
func (w *EventWriter) Write(msg *decoder.Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed (daily rotation)
	if time.Since(w.lastRotate) > 24*time.Hour {
		if err := w.rotate(); err != nil {
			return err
		}
	}

	// Write the event
	return w.encoder.Encode(msg)
}

// rotate rotates the events file
func (w *EventWriter) rotate() error {
	// Close current file if open
	if w.file != nil {
		w.file.Close()
	}

	// Create new file with timestamp
	filename := fmt.Sprintf("events_%s.jsonl", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(w.basePath, filename)

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.encoder = json.NewEncoder(file)
	w.lastRotate = time.Now()

	return nil
}

// Close closes the event writer
func (w *EventWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// CDRWriter methods

// Write writes a CDR from a session
func (w *CDRWriter) Write(session *correlation.Session) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation is needed (hourly rotation)
	if time.Since(w.lastRotate) > time.Hour {
		if err := w.rotate(); err != nil {
			return err
		}
	}

	// Build CDR record
	record := w.buildRecord(session)

	// Write the record
	if err := w.writer.Write(record); err != nil {
		return err
	}

	w.writer.Flush()
	return w.writer.Error()
}

// buildRecord builds a CSV record from a session
func (w *CDRWriter) buildRecord(session *correlation.Session) []string {
	record := make([]string, len(w.fields))

	for i, field := range w.fields {
		switch field {
		case "tid":
			record[i] = session.TID
		case "imsi":
			record[i] = session.IMSI
		case "msisdn":
			record[i] = session.MSISDN
		case "procedure":
			record[i] = session.Procedure
		case "start_time":
			record[i] = session.StartTime.Format(time.RFC3339)
		case "end_time":
			record[i] = session.LastActivity.Format(time.RFC3339)
		case "duration_ms":
			record[i] = fmt.Sprintf("%d", session.Duration.Milliseconds())
		case "result":
			record[i] = string(session.Result)
		case "cause":
			record[i] = fmt.Sprintf("%d", session.FailureCause)
		case "plmn":
			record[i] = session.PLMN
		case "cell_id":
			record[i] = session.CellID
		case "apn":
			record[i] = session.APN
		case "vendor":
			// Extract vendor from first message
			if len(session.Messages) > 0 {
				record[i] = session.Messages[0].VendorName
			}
		default:
			record[i] = ""
		}
	}

	return record
}

// rotate rotates the CDR file
func (w *CDRWriter) rotate() error {
	// Close and flush current file if open
	if w.writer != nil {
		w.writer.Flush()
	}
	if w.file != nil {
		w.file.Close()
	}

	// Create new file with timestamp
	filename := fmt.Sprintf("cdr_%s.csv", time.Now().Format("2006-01-02_15"))
	filepath := filepath.Join(w.basePath, filename)

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.writer = csv.NewWriter(file)

	// Write header if new file
	if stat, err := file.Stat(); err == nil && stat.Size() == 0 {
		if err := w.writer.Write(w.fields); err != nil {
			return err
		}
		w.writer.Flush()
	}

	w.lastRotate = time.Now()

	return nil
}

// Close closes the CDR writer
func (w *CDRWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		w.writer.Flush()
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Storage cleanup methods

// cleanupRoutine periodically cleans up old files
func (s *Storage) cleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup removes files older than retention period
func (s *Storage) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	// Cleanup events
	if s.config.EventsEnabled {
		s.cleanupDirectory(s.config.EventsPath, cutoff)
	}

	// Cleanup CDRs
	if s.config.CDREnabled {
		s.cleanupDirectory(s.config.CDRPath, cutoff)
	}
}

// cleanupDirectory removes old files from a directory
func (s *Storage) cleanupDirectory(dirPath string, cutoff time.Time) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := filepath.Join(dirPath, entry.Name())
			os.Remove(filepath)
		}
	}
}

// Close closes all writers
func (s *Storage) Close() error {
	if s.eventWriter != nil {
		s.eventWriter.Close()
	}
	if s.cdrWriter != nil {
		s.cdrWriter.Close()
	}
	return nil
}
