package oam

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ConfigManager manages application configuration
type ConfigManager struct {
	mu                sync.RWMutex
	configDir         string
	configs           map[string]*ConfigFile
	versionHistory    []ConfigVersion
	maxVersionHistory int
}

// ConfigFile represents a configuration file
type ConfigFile struct {
	Name         string
	Path         string
	Content      map[string]string
	LastModified time.Time
	Version      int
	ChecksumMD5  string
}

// ConfigVersion represents a version of configuration
type ConfigVersion struct {
	Timestamp   time.Time
	Files       []string
	Description string
	CreatedBy   string
}

// ConfigChange represents a configuration change
type ConfigChange struct {
	File      string
	Key       string
	OldValue  string
	NewValue  string
	Timestamp time.Time
	ChangedBy string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configDir string) (*ConfigManager, error) {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("config directory does not exist: %s", configDir)
	}

	manager := &ConfigManager{
		configDir:         configDir,
		configs:           make(map[string]*ConfigFile),
		versionHistory:    make([]ConfigVersion, 0),
		maxVersionHistory: 50,
	}

	// Load all configuration files
	if err := manager.loadAllConfigs(); err != nil {
		return nil, err
	}

	return manager, nil
}

// loadAllConfigs loads all configuration files from the config directory
func (m *ConfigManager) loadAllConfigs() error {
	configFiles := []string{
		"db.cfg",
		"license.cfg",
		"protocols.cfg",
		"system.cfg",
		"trace.cfg",
		"paths.cfg",
		"security.cfg",
	}

	for _, filename := range configFiles {
		if err := m.loadConfig(filename); err != nil {
			return fmt.Errorf("failed to load %s: %w", filename, err)
		}
	}

	return nil
}

// loadConfig loads a single configuration file
func (m *ConfigManager) loadConfig(filename string) error {
	filepath := filepath.Join(m.configDir, filename)

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	config := &ConfigFile{
		Name:    filename,
		Path:    filepath,
		Content: make(map[string]string),
	}

	// Parse configuration file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			config.Content[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Get file stats
	stat, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	config.LastModified = stat.ModTime()

	m.mu.Lock()
	m.configs[filename] = config
	m.mu.Unlock()

	return nil
}

// GetConfig retrieves a configuration file
func (m *ConfigManager) GetConfig(filename string) (*ConfigFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[filename]
	if !exists {
		return nil, fmt.Errorf("configuration file not found: %s", filename)
	}

	return config, nil
}

// GetConfigValue retrieves a specific configuration value
func (m *ConfigManager) GetConfigValue(filename, key string) (string, error) {
	config, err := m.GetConfig(filename)
	if err != nil {
		return "", err
	}

	value, exists := config.Content[key]
	if !exists {
		return "", fmt.Errorf("configuration key not found: %s in %s", key, filename)
	}

	return value, nil
}

// SetConfigValue sets a configuration value
func (m *ConfigManager) SetConfigValue(filename, key, value, changedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, exists := m.configs[filename]
	if !exists {
		return fmt.Errorf("configuration file not found: %s", filename)
	}

	// Record old value
	oldValue := config.Content[key]

	// Set new value
	config.Content[key] = value

	// Save to file
	if err := m.saveConfigFile(config); err != nil {
		// Rollback
		config.Content[key] = oldValue
		return err
	}

	// Record change
	change := ConfigChange{
		File:      filename,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
		ChangedBy: changedBy,
	}

	m.recordChange(change)

	return nil
}

// SetMultipleValues sets multiple configuration values atomically
func (m *ConfigManager) SetMultipleValues(filename string, values map[string]string, changedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, exists := m.configs[filename]
	if !exists {
		return fmt.Errorf("configuration file not found: %s", filename)
	}

	// Save old values for rollback
	oldValues := make(map[string]string)
	for key := range values {
		oldValues[key] = config.Content[key]
	}

	// Set new values
	for key, value := range values {
		config.Content[key] = value
	}

	// Save to file
	if err := m.saveConfigFile(config); err != nil {
		// Rollback all changes
		for key, oldValue := range oldValues {
			config.Content[key] = oldValue
		}
		return err
	}

	// Record changes
	for key, value := range values {
		change := ConfigChange{
			File:      filename,
			Key:       key,
			OldValue:  oldValues[key],
			NewValue:  value,
			Timestamp: time.Now(),
			ChangedBy: changedBy,
		}
		m.recordChange(change)
	}

	return nil
}

// saveConfigFile writes configuration to disk
func (m *ConfigManager) saveConfigFile(config *ConfigFile) error {
	// Create backup first
	backupPath := config.Path + ".bak"
	if err := copyFile(config.Path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write new configuration
	file, err := os.Create(config.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	fmt.Fprintf(writer, "# Configuration file: %s\n", config.Name)
	fmt.Fprintf(writer, "# Last modified: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(writer, "#\n\n")

	// Group related settings
	groups := m.groupConfigKeys(config.Name, config.Content)

	for groupName, keys := range groups {
		if groupName != "" {
			fmt.Fprintf(writer, "[%s]\n", groupName)
		}

		for _, key := range keys {
			value := config.Content[key]
			// Quote values with spaces
			if strings.Contains(value, " ") {
				fmt.Fprintf(writer, "%s = \"%s\"\n", key, value)
			} else {
				fmt.Fprintf(writer, "%s = %s\n", key, value)
			}
		}
		fmt.Fprintln(writer, "")
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	// Update metadata
	config.LastModified = time.Now()
	config.Version++

	return nil
}

// groupConfigKeys groups configuration keys by category
func (m *ConfigManager) groupConfigKeys(filename string, content map[string]string) map[string][]string {
	groups := make(map[string][]string)

	for key := range content {
		var group string

		switch filename {
		case "db.cfg":
			if strings.HasPrefix(key, "DB_") {
				group = "database"
			} else if strings.HasPrefix(key, "POOL_") {
				group = "connection_pool"
			}

		case "system.cfg":
			if strings.HasPrefix(key, "LOG_") {
				group = "logging"
			} else if strings.HasPrefix(key, "CDR_") {
				group = "cdr"
			} else if strings.HasPrefix(key, "REDIS_") {
				group = "redis"
			} else if strings.HasPrefix(key, "WEB_") {
				group = "web_server"
			}

		case "protocols.cfg":
			if strings.Contains(key, "ENABLE_") {
				group = "enabled_protocols"
			} else if strings.Contains(key, "PORT_") {
				group = "protocol_ports"
			}

		default:
			group = ""
		}

		groups[group] = append(groups[group], key)
	}

	return groups
}

// ValidateConfig validates configuration values
func (m *ConfigManager) ValidateConfig(filename string, values map[string]string) []ValidationError {
	errors := make([]ValidationError, 0)

	for key, value := range values {
		if err := m.validateConfigValue(filename, key, value); err != nil {
			errors = append(errors, ValidationError{
				File:    filename,
				Key:     key,
				Value:   value,
				Message: err.Error(),
			})
		}
	}

	return errors
}

// validateConfigValue validates a single configuration value
func (m *ConfigManager) validateConfigValue(filename, key, value string) error {
	switch filename {
	case "db.cfg":
		return m.validateDatabaseConfig(key, value)
	case "system.cfg":
		return m.validateSystemConfig(key, value)
	case "protocols.cfg":
		return m.validateProtocolConfig(key, value)
	case "security.cfg":
		return m.validateSecurityConfig(key, value)
	}

	return nil
}

// validateDatabaseConfig validates database configuration
func (m *ConfigManager) validateDatabaseConfig(key, value string) error {
	switch key {
	case "DB_PORT":
		return validatePort(value)
	case "DB_MAX_CONNECTIONS":
		return validatePositiveInteger(value, 1, 1000)
	case "DB_TIMEOUT_SECONDS":
		return validatePositiveInteger(value, 1, 3600)
	}
	return nil
}

// validateSystemConfig validates system configuration
func (m *ConfigManager) validateSystemConfig(key, value string) error {
	switch key {
	case "WEB_PORT":
		return validatePort(value)
	case "LOG_LEVEL":
		return validateEnum(value, []string{"debug", "info", "warning", "error"})
	case "LOG_MAX_SIZE_MB":
		return validatePositiveInteger(value, 1, 10000)
	case "CDR_ROTATION_SIZE_MB":
		return validatePositiveInteger(value, 1, 10000)
	case "REDIS_PORT":
		return validatePort(value)
	}
	return nil
}

// validateProtocolConfig validates protocol configuration
func (m *ConfigManager) validateProtocolConfig(key, value string) error {
	if strings.HasPrefix(key, "ENABLE_") {
		return validateBoolean(value)
	}
	if strings.Contains(key, "PORT") {
		return validatePort(value)
	}
	return nil
}

// validateSecurityConfig validates security configuration
func (m *ConfigManager) validateSecurityConfig(key, value string) error {
	switch key {
	case "SESSION_TIMEOUT_MINUTES":
		return validatePositiveInteger(value, 1, 1440)
	case "PASSWORD_MIN_LENGTH":
		return validatePositiveInteger(value, 6, 32)
	case "MAX_LOGIN_ATTEMPTS":
		return validatePositiveInteger(value, 1, 10)
	}
	return nil
}

// CreateBackup creates a backup of all configuration files
func (m *ConfigManager) CreateBackup(description, createdBy string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backupDir := filepath.Join(m.configDir, ".backups", time.Now().Format("20060102_150405"))
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return err
	}

	backedUpFiles := make([]string, 0)

	for _, config := range m.configs {
		backupPath := filepath.Join(backupDir, config.Name)
		if err := copyFile(config.Path, backupPath); err != nil {
			return fmt.Errorf("failed to backup %s: %w", config.Name, err)
		}
		backedUpFiles = append(backedUpFiles, config.Name)
	}

	// Record version
	version := ConfigVersion{
		Timestamp:   time.Now(),
		Files:       backedUpFiles,
		Description: description,
		CreatedBy:   createdBy,
	}

	m.versionHistory = append(m.versionHistory, version)

	// Limit history size
	if len(m.versionHistory) > m.maxVersionHistory {
		m.versionHistory = m.versionHistory[1:]
	}

	return nil
}

// RestoreBackup restores configuration from a backup
func (m *ConfigManager) RestoreBackup(backupTimestamp time.Time) error {
	backupDir := filepath.Join(m.configDir, ".backups", backupTimestamp.Format("20060102_150405"))

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupTimestamp.Format("2006-01-02 15:04:05"))
	}

	// Restore each file
	for filename := range m.configs {
		backupPath := filepath.Join(backupDir, filename)
		destPath := filepath.Join(m.configDir, filename)

		if err := copyFile(backupPath, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", filename, err)
		}
	}

	// Reload configurations
	return m.loadAllConfigs()
}

// GetBackupHistory returns the backup history
func (m *ConfigManager) GetBackupHistory() []ConfigVersion {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]ConfigVersion{}, m.versionHistory...)
}

// recordChange records a configuration change (stub for audit log)
func (m *ConfigManager) recordChange(change ConfigChange) {
	// TODO: Implement audit logging
	// For now, just log to stdout
	fmt.Printf("[CONFIG] %s: %s.%s changed from '%s' to '%s' by %s\n",
		change.Timestamp.Format(time.RFC3339),
		change.File,
		change.Key,
		change.OldValue,
		change.NewValue,
		change.ChangedBy,
	)
}

// Helper functions

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func validatePort(value string) error {
	return validatePositiveInteger(value, 1, 65535)
}

func validatePositiveInteger(value string, min, max int) error {
	var num int
	if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
		return fmt.Errorf("invalid integer: %s", value)
	}
	if num < min || num > max {
		return fmt.Errorf("value must be between %d and %d", min, max)
	}
	return nil
}

func validateBoolean(value string) error {
	lower := strings.ToLower(value)
	if lower != "true" && lower != "false" && lower != "1" && lower != "0" {
		return fmt.Errorf("invalid boolean value: %s", value)
	}
	return nil
}

func validateEnum(value string, validValues []string) error {
	for _, valid := range validValues {
		if value == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid value: %s (must be one of: %v)", value, validValues)
}

func validateIPAddress(value string) error {
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if !ipRegex.MatchString(value) {
		return fmt.Errorf("invalid IP address: %s", value)
	}
	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	File    string
	Key     string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s.%s: %s (value: %s)", e.File, e.Key, e.Message, e.Value)
}
