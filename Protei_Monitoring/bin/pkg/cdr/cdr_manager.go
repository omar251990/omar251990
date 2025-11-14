package cdr

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// CDRManager manages CDR writers for all protocols
type CDRManager struct {
	mu        sync.RWMutex
	writers   map[string]*CDRWriter
	baseDir   string
	format    CDRFormat
	rotation  CDRRotation
	dbTracker CDRDatabaseTracker
}

// NewCDRManager creates a new CDR manager
func NewCDRManager(baseDir string, format CDRFormat, rotation CDRRotation, db *sql.DB) (*CDRManager, error) {
	dbTracker := NewDatabaseTracker(db)

	manager := &CDRManager{
		writers:   make(map[string]*CDRWriter),
		baseDir:   baseDir,
		format:    format,
		rotation:  rotation,
		dbTracker: dbTracker,
	}

	// Initialize writers for all supported protocols
	protocols := []string{
		"MAP", "CAP", "INAP", "Diameter", "GTP", "PFCP",
		"HTTP2", "NGAP", "S1AP", "NAS",
	}

	for _, protocol := range protocols {
		if err := manager.initializeWriter(protocol); err != nil {
			return nil, fmt.Errorf("failed to initialize %s CDR writer: %w", protocol, err)
		}
	}

	return manager, nil
}

// initializeWriter creates a CDR writer for a specific protocol
func (m *CDRManager) initializeWriter(protocol string) error {
	config := CDRConfig{
		BaseDir:   m.baseDir,
		Protocol:  protocol,
		Format:    m.format,
		Rotation:  m.rotation,
		DBTracker: m.dbTracker,
	}

	writer, err := NewCDRWriter(config)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.writers[protocol] = writer
	m.mu.Unlock()

	return nil
}

// WriteMapCDR writes a MAP CDR record
func (m *CDRManager) WriteMapCDR(cdr *MapCDR) error {
	return m.writeRecord("MAP", cdr)
}

// WriteCapCDR writes a CAP CDR record
func (m *CDRManager) WriteCapCDR(cdr *CapCDR) error {
	return m.writeRecord("CAP", cdr)
}

// WriteInapCDR writes an INAP CDR record
func (m *CDRManager) WriteInapCDR(cdr *InapCDR) error {
	return m.writeRecord("INAP", cdr)
}

// WriteDiameterCDR writes a Diameter CDR record
func (m *CDRManager) WriteDiameterCDR(cdr *DiameterCDR) error {
	return m.writeRecord("Diameter", cdr)
}

// WriteGtpCDR writes a GTP CDR record
func (m *CDRManager) WriteGtpCDR(cdr *GtpCDR) error {
	return m.writeRecord("GTP", cdr)
}

// WritePfcpCDR writes a PFCP CDR record
func (m *CDRManager) WritePfcpCDR(cdr *PfcpCDR) error {
	return m.writeRecord("PFCP", cdr)
}

// WriteHttp2CDR writes an HTTP/2 CDR record
func (m *CDRManager) WriteHttp2CDR(cdr *Http2CDR) error {
	return m.writeRecord("HTTP2", cdr)
}

// WriteNgapCDR writes an NGAP CDR record
func (m *CDRManager) WriteNgapCDR(cdr *NgapCDR) error {
	return m.writeRecord("NGAP", cdr)
}

// WriteS1apCDR writes an S1AP CDR record
func (m *CDRManager) WriteS1apCDR(cdr *S1apCDR) error {
	return m.writeRecord("S1AP", cdr)
}

// WriteNasCDR writes a NAS CDR record
func (m *CDRManager) WriteNasCDR(cdr *NasCDR) error {
	return m.writeRecord("NAS", cdr)
}

// writeRecord writes a CDR record using the appropriate writer
func (m *CDRManager) writeRecord(protocol string, record CDRRecord) error {
	m.mu.RLock()
	writer, exists := m.writers[protocol]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no CDR writer found for protocol: %s", protocol)
	}

	return writer.WriteRecord(record)
}

// Flush flushes all CDR writers
func (m *CDRManager) Flush() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for protocol, writer := range m.writers {
		if err := writer.Flush(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", protocol, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("flush errors: %v", errs)
	}

	return nil
}

// Close closes all CDR writers
func (m *CDRManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for protocol, writer := range m.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", protocol, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// GetStats returns statistics for all CDR writers
func (m *CDRManager) GetStats() map[string]CDRStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]CDRStats)
	for protocol, writer := range m.writers {
		recordCount, bytesWritten, duration := writer.GetStats()
		stats[protocol] = CDRStats{
			Protocol:      protocol,
			RecordCount:   recordCount,
			BytesWritten:  bytesWritten,
			Duration:      duration,
		}
	}

	return stats
}

// CDRStats holds statistics for a CDR writer
type CDRStats struct {
	Protocol     string
	RecordCount  int64
	BytesWritten int64
	Duration     time.Duration
}

// DatabaseTracker implements CDRDatabaseTracker interface
type DatabaseTracker struct {
	db *sql.DB
}

// NewDatabaseTracker creates a new database tracker
func NewDatabaseTracker(db *sql.DB) *DatabaseTracker {
	return &DatabaseTracker{db: db}
}

// TrackCDRFile records CDR file metadata in the database
func (t *DatabaseTracker) TrackCDRFile(filename, protocol string, recordCount int64, fileSize int64, startTime, endTime time.Time) error {
	if t.db == nil {
		return nil // Database tracking is optional
	}

	query := `
		INSERT INTO cdr_files (
			filename, protocol, record_count, file_size_bytes,
			start_time, end_time, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err := t.db.Exec(query, filename, protocol, recordCount, fileSize, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to track CDR file: %w", err)
	}

	return nil
}

// GetCDRFileList retrieves CDR file list from database
func (t *DatabaseTracker) GetCDRFileList(protocol string, startTime, endTime time.Time) ([]CDRFileInfo, error) {
	if t.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT
			id, filename, protocol, record_count, file_size_bytes,
			start_time, end_time, compressed, created_at
		FROM cdr_files
		WHERE protocol = $1
		  AND start_time >= $2
		  AND end_time <= $3
		ORDER BY start_time DESC
	`

	rows, err := t.db.Query(query, protocol, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []CDRFileInfo
	for rows.Next() {
		var file CDRFileInfo
		err := rows.Scan(
			&file.ID,
			&file.Filename,
			&file.Protocol,
			&file.RecordCount,
			&file.FileSizeBytes,
			&file.StartTime,
			&file.EndTime,
			&file.Compressed,
			&file.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

// CDRFileInfo holds information about a CDR file
type CDRFileInfo struct {
	ID            int64
	Filename      string
	Protocol      string
	RecordCount   int64
	FileSizeBytes int64
	StartTime     time.Time
	EndTime       time.Time
	Compressed    bool
	CreatedAt     time.Time
}

// CleanupOldCDRFiles deletes old CDR file records from database
func (t *DatabaseTracker) CleanupOldCDRFiles(retentionDays int) (int64, error) {
	if t.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	query := `
		DELETE FROM cdr_files
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := t.db.Exec(query, retentionDays)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetCDRStatsByProtocol returns aggregated statistics per protocol
func (t *DatabaseTracker) GetCDRStatsByProtocol(startTime, endTime time.Time) (map[string]ProtocolCDRStats, error) {
	if t.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT
			protocol,
			COUNT(*) as file_count,
			SUM(record_count) as total_records,
			SUM(file_size_bytes) as total_bytes
		FROM cdr_files
		WHERE start_time >= $1 AND end_time <= $2
		GROUP BY protocol
		ORDER BY protocol
	`

	rows, err := t.db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]ProtocolCDRStats)
	for rows.Next() {
		var protocol string
		var stat ProtocolCDRStats
		err := rows.Scan(
			&protocol,
			&stat.FileCount,
			&stat.TotalRecords,
			&stat.TotalBytes,
		)
		if err != nil {
			return nil, err
		}
		stats[protocol] = stat
	}

	return stats, rows.Err()
}

// ProtocolCDRStats holds aggregated statistics for a protocol
type ProtocolCDRStats struct {
	FileCount    int64
	TotalRecords int64
	TotalBytes   int64
}
