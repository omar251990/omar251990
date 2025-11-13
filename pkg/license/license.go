package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

// Manager handles license validation and enforcement
type Manager struct {
	license *License
	secret  []byte
}

// License represents the application license
type License struct {
	CustomerName   string            `json:"customer_name"`
	ExpiryDate     string            `json:"expiry_date"`
	LicensedMAC    string            `json:"licensed_mac"`
	Enable2G       bool              `json:"enable_2g"`
	Enable3G       bool              `json:"enable_3g"`
	Enable4G       bool              `json:"enable_4g"`
	Enable5G       bool              `json:"enable_5g"`
	EnableMAP      bool              `json:"enable_map"`
	EnableCAP      bool              `json:"enable_cap"`
	EnableINAP     bool              `json:"enable_inap"`
	EnableDiameter bool              `json:"enable_diameter"`
	EnableHTTP     bool              `json:"enable_http"`
	EnableGTP      bool              `json:"enable_gtp"`
	MaxSubscribers int               `json:"max_subscribers"`
	MaxTPS         int               `json:"max_tps"`
	Features       map[string]bool   `json:"features,omitempty"`
	Signature      string            `json:"signature"`
}

var (
	ErrInvalidLicense   = errors.New("invalid license")
	ErrExpiredLicense   = errors.New("license expired")
	ErrMACMismatch      = errors.New("MAC address mismatch")
	ErrInvalidSignature = errors.New("invalid license signature")
	ErrFeatureDisabled  = errors.New("feature not enabled in license")
)

// Secret vendor key (should be compiled into binary, not in config)
const vendorSecretKey = "PROTEI_MONITORING_VENDOR_KEY_2025_CHANGE_THIS_IN_PRODUCTION"

// NewManager creates a new license manager
func NewManager(licensePath string) (*Manager, error) {
	mgr := &Manager{
		secret: []byte(vendorSecretKey),
	}

	if err := mgr.LoadLicense(licensePath); err != nil {
		return nil, err
	}

	if err := mgr.ValidateLicense(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// LoadLicense loads license from file
func (m *Manager) LoadLicense(licensePath string) error {
	data, err := os.ReadFile(licensePath)
	if err != nil {
		return fmt.Errorf("failed to read license file: %w", err)
	}

	var license License
	if err := json.Unmarshal(data, &license); err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}

	m.license = &license
	return nil
}

// ValidateLicense validates the license
func (m *Manager) ValidateLicense() error {
	if m.license == nil {
		return ErrInvalidLicense
	}

	// Check expiry date
	expiryDate, err := time.Parse("2006-01-02", m.license.ExpiryDate)
	if err != nil {
		return fmt.Errorf("invalid expiry date: %w", err)
	}

	if time.Now().After(expiryDate) {
		return ErrExpiredLicense
	}

	// Check MAC address
	currentMAC, err := getMACAddress()
	if err != nil {
		return fmt.Errorf("failed to get MAC address: %w", err)
	}

	if !strings.EqualFold(m.license.LicensedMAC, currentMAC) {
		return fmt.Errorf("%w: expected %s, got %s", ErrMACMismatch, m.license.LicensedMAC, currentMAC)
	}

	// Verify signature
	expectedSignature := m.computeSignature()
	if m.license.Signature != expectedSignature {
		return ErrInvalidSignature
	}

	return nil
}

// computeSignature computes HMAC-SHA256 signature of license fields
func (m *Manager) computeSignature() string {
	// Build canonical representation
	data := map[string]interface{}{
		"customer_name":   m.license.CustomerName,
		"expiry_date":     m.license.ExpiryDate,
		"licensed_mac":    m.license.LicensedMAC,
		"enable_2g":       m.license.Enable2G,
		"enable_3g":       m.license.Enable3G,
		"enable_4g":       m.license.Enable4G,
		"enable_5g":       m.license.Enable5G,
		"enable_map":      m.license.EnableMAP,
		"enable_cap":      m.license.EnableCAP,
		"enable_inap":     m.license.EnableINAP,
		"enable_diameter": m.license.EnableDiameter,
		"enable_http":     m.license.EnableHTTP,
		"enable_gtp":      m.license.EnableGTP,
		"max_subscribers": m.license.MaxSubscribers,
		"max_tps":         m.license.MaxTPS,
	}

	canonical := canonicalize(data)
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil))
}

// canonicalize creates a canonical string representation
func canonicalize(data map[string]interface{}) string {
	// Sort keys for deterministic representation
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build JSON array of [key, value] pairs
	pairs := make([][2]interface{}, 0, len(data))
	for _, k := range keys {
		pairs = append(pairs, [2]interface{}{k, data[k]})
	}

	jsonData, _ := json.Marshal(pairs)
	return string(jsonData)
}

// getMACAddress retrieves the system MAC address
func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get hardware address
		if len(iface.HardwareAddr) > 0 {
			mac := iface.HardwareAddr.String()
			// Normalize to lowercase with colons
			mac = strings.ToLower(strings.ReplaceAll(mac, "-", ":"))
			return mac, nil
		}
	}

	return "", errors.New("no valid MAC address found")
}

// IsFeatureEnabled checks if a feature is enabled
func (m *Manager) IsFeatureEnabled(feature string) bool {
	switch feature {
	case "2g":
		return m.license.Enable2G
	case "3g":
		return m.license.Enable3G
	case "4g":
		return m.license.Enable4G
	case "5g":
		return m.license.Enable5G
	case "map":
		return m.license.EnableMAP
	case "cap":
		return m.license.EnableCAP
	case "inap":
		return m.license.EnableINAP
	case "diameter":
		return m.license.EnableDiameter
	case "http":
		return m.license.EnableHTTP
	case "gtp":
		return m.license.EnableGTP
	default:
		if enabled, ok := m.license.Features[feature]; ok {
			return enabled
		}
		return false
	}
}

// CheckFeature checks if a feature is enabled, returns error if not
func (m *Manager) CheckFeature(feature string) error {
	if !m.IsFeatureEnabled(feature) {
		return fmt.Errorf("%w: %s", ErrFeatureDisabled, feature)
	}
	return nil
}

// GetLicense returns the current license
func (m *Manager) GetLicense() *License {
	return m.license
}

// GetMaxSubscribers returns max subscribers allowed
func (m *Manager) GetMaxSubscribers() int {
	return m.license.MaxSubscribers
}

// GetMaxTPS returns max transactions per second allowed
func (m *Manager) GetMaxTPS() int {
	return m.license.MaxTPS
}

// GenerateLicense generates a license file (vendor tool)
func GenerateLicense(license *License, secret string) (string, error) {
	// Compute signature
	mgr := &Manager{
		license: license,
		secret:  []byte(secret),
	}

	license.Signature = mgr.computeSignature()

	// Marshal to JSON
	data, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
