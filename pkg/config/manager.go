package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Manager handles runtime configuration
type Manager struct {
	mu           sync.RWMutex
	configPath   string
	config       map[string]interface{}
	restartFunc  func() error
}

// NewManager creates a new configuration manager
func NewManager(configPath string, restartFunc func() error) (*Manager, error) {
	m := &Manager{
		configPath:  configPath,
		restartFunc: restartFunc,
	}

	// Load initial configuration
	if err := m.loadConfig(); err != nil {
		return nil, err
	}

	return m, nil
}

// loadConfig loads configuration from file
func (m *Manager) loadConfig() error {
	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	m.mu.Lock()
	m.config = config
	m.mu.Unlock()

	return nil
}

// saveConfig saves configuration to file
func (m *Manager) saveConfig() error {
	m.mu.RLock()
	data, err := yaml.Marshal(m.config)
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first
	tmpFile := m.configPath + ".tmp"
	if err := ioutil.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	// Rename to actual file (atomic operation)
	if err := os.Rename(tmpFile, m.configPath); err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}

	return nil
}

// GetConfig returns the entire configuration
func (m *Manager) GetConfig() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Deep copy to prevent external modifications
	return deepCopy(m.config), nil
}

// UpdateConfig updates configuration with provided values
func (m *Manager) UpdateConfig(updates map[string]interface{}) error {
	m.mu.Lock()
	// Merge updates into config
	for key, value := range updates {
		m.config[key] = value
	}
	m.mu.Unlock()

	// Save to file
	return m.saveConfig()
}

// RestartService restarts the service
func (m *Manager) RestartService() error {
	if m.restartFunc != nil {
		return m.restartFunc()
	}
	return fmt.Errorf("restart function not configured")
}

// GetProtocolConfig returns configuration for a specific protocol
func (m *Manager) GetProtocolConfig(protocol string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	protocols, ok := m.config["protocols"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("protocols configuration not found")
	}

	protocolConfig, ok := protocols[protocol].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("protocol %s not found", protocol)
	}

	return deepCopy(protocolConfig), nil
}

// UpdateProtocolConfig updates configuration for a specific protocol
func (m *Manager) UpdateProtocolConfig(protocol string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	protocols, ok := m.config["protocols"].(map[string]interface{})
	if !ok {
		protocols = make(map[string]interface{})
		m.config["protocols"] = protocols
	}

	protocols[protocol] = config

	return m.saveConfig()
}

// GetNetworkConfig returns configuration for a specific network
func (m *Manager) GetNetworkConfig(network string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	networks, ok := m.config["networks"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("networks configuration not found")
	}

	networkConfig, ok := networks[network].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("network %s not found", network)
	}

	return deepCopy(networkConfig), nil
}

// UpdateNetworkConfig updates configuration for a specific network
func (m *Manager) UpdateNetworkConfig(network string, config map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	networks, ok := m.config["networks"].(map[string]interface{})
	if !ok {
		networks = make(map[string]interface{})
		m.config["networks"] = networks
	}

	networks[network] = config

	return m.saveConfig()
}

// EnableProtocol enables a protocol
func (m *Manager) EnableProtocol(protocol string) error {
	config, err := m.GetProtocolConfig(protocol)
	if err != nil {
		return err
	}

	config["enabled"] = true
	return m.UpdateProtocolConfig(protocol, config)
}

// DisableProtocol disables a protocol
func (m *Manager) DisableProtocol(protocol string) error {
	config, err := m.GetProtocolConfig(protocol)
	if err != nil {
		return err
	}

	config["enabled"] = false
	return m.UpdateProtocolConfig(protocol, config)
}

// EnableNetwork enables a network
func (m *Manager) EnableNetwork(network string) error {
	config, err := m.GetNetworkConfig(network)
	if err != nil {
		return err
	}

	config["enabled"] = true
	return m.UpdateNetworkConfig(network, config)
}

// DisableNetwork disables a network
func (m *Manager) DisableNetwork(network string) error {
	config, err := m.GetNetworkConfig(network)
	if err != nil {
		return err
	}

	config["enabled"] = false
	return m.UpdateNetworkConfig(network, config)
}

// GetValue returns a configuration value by path
func (m *Manager) GetValue(path string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return getNestedValue(m.config, path)
}

// SetValue sets a configuration value by path
func (m *Manager) SetValue(path string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := setNestedValue(m.config, path, value); err != nil {
		return err
	}

	return m.saveConfig()
}

// Reload reloads configuration from file
func (m *Manager) Reload() error {
	return m.loadConfig()
}

// Helper: Deep copy map
func deepCopy(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})
	for k, v := range src {
		switch v := v.(type) {
		case map[string]interface{}:
			dst[k] = deepCopy(v)
		case []interface{}:
			dst[k] = deepCopySlice(v)
		default:
			dst[k] = v
		}
	}
	return dst
}

// Helper: Deep copy slice
func deepCopySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch v := v.(type) {
		case map[string]interface{}:
			dst[i] = deepCopy(v)
		case []interface{}:
			dst[i] = deepCopySlice(v)
		default:
			dst[i] = v
		}
	}
	return dst
}

// Helper: Get nested value by path (e.g., "server.port")
func getNestedValue(m map[string]interface{}, path string) (interface{}, error) {
	// TODO: Implement path parsing
	return nil, fmt.Errorf("not implemented")
}

// Helper: Set nested value by path
func setNestedValue(m map[string]interface{}, path string, value interface{}) error {
	// TODO: Implement path parsing
	return fmt.Errorf("not implemented")
}
