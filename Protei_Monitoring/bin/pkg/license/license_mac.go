package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

// LicenseConfig represents the complete license configuration
type LicenseConfig struct {
	CustomerName   string `yaml:"customer_name"`
	ExpiryDate     string `yaml:"expiry_date"` // YYYY-MM-DD format
	LicensedMAC    string `yaml:"licensed_mac"` // XX:XX:XX:XX:XX:XX format

	// Feature flags
	Enable2G bool `yaml:"enable_2g"`
	Enable3G bool `yaml:"enable_3g"`
	Enable4G bool `yaml:"enable_4g"`
	Enable5G bool `yaml:"enable_5g"`

	// Protocol flags
	EnableMAP      bool `yaml:"enable_map"`
	EnableCAP      bool `yaml:"enable_cap"`
	EnableINAP     bool `yaml:"enable_inap"`
	EnableDiameter bool `yaml:"enable_diameter"`
	EnableHTTP     bool `yaml:"enable_http"`
	EnableGTP      bool `yaml:"enable_gtp"`

	// Capacity limits
	MaxSubscribers int `yaml:"max_subscribers"`
	MaxTPS         int `yaml:"max_tps"` // Transactions Per Second

	// HMAC signature (calculated over all above fields)
	Signature string `yaml:"signature"`
}

const (
	// VendorSecret is the secret key for HMAC signature
	// In production, this should be:
	// 1. Stored securely (not in source code)
	// 2. Different for each customer/deployment
	// 3. Loaded from environment variable or encrypted config
	VendorSecret = "PROTEI_MONITORING_VENDOR_SECRET_KEY_2025"
)

// ValidateLicense validates the license against system MAC and signature
func ValidateLicense(cfg *LicenseConfig) error {
	// 1. Validate expiry date
	if err := validateExpiryDate(cfg.ExpiryDate); err != nil {
		return fmt.Errorf("license expired: %w", err)
	}

	// 2. Get system MAC addresses
	systemMACs, err := getSystemMACAddresses()
	if err != nil {
		return fmt.Errorf("failed to get system MAC addresses: %w", err)
	}

	// 3. Check if any system MAC matches licensed MAC
	licensedMAC := normalizeMACAddress(cfg.LicensedMAC)
	if !containsMAC(systemMACs, licensedMAC) {
		return fmt.Errorf("license MAC mismatch: licensed=%s, system=%v",
			licensedMAC, systemMACs)
	}

	// 4. Verify HMAC signature
	expectedSignature := calculateSignature(cfg)
	if cfg.Signature != expectedSignature {
		return fmt.Errorf("license signature verification failed: license may be tampered")
	}

	return nil
}

// validateExpiryDate checks if license has not expired
func validateExpiryDate(expiryDate string) error {
	expiry, err := time.Parse("2006-01-02", expiryDate)
	if err != nil {
		return fmt.Errorf("invalid expiry date format (expected YYYY-MM-DD): %w", err)
	}

	if time.Now().After(expiry) {
		daysExpired := int(time.Since(expiry).Hours() / 24)
		return fmt.Errorf("license expired %d days ago (expiry: %s)",
			daysExpired, expiryDate)
	}

	return nil
}

// getSystemMACAddresses returns all non-loopback MAC addresses on the system
func getSystemMACAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var macs []string
	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip interfaces without MAC address
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		mac := normalizeMACAddress(iface.HardwareAddr.String())
		macs = append(macs, mac)
	}

	if len(macs) == 0 {
		return nil, fmt.Errorf("no valid network interfaces found")
	}

	return macs, nil
}

// normalizeMACAddress normalizes MAC address to lowercase with colons
func normalizeMACAddress(mac string) string {
	// Remove all non-hex characters
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ReplaceAll(mac, ".", ":")
	return mac
}

// containsMAC checks if MAC address exists in list
func containsMAC(macs []string, targetMAC string) bool {
	targetMAC = normalizeMACAddress(targetMAC)
	for _, mac := range macs {
		if normalizeMACAddress(mac) == targetMAC {
			return true
		}
	}
	return false
}

// calculateSignature computes HMAC-SHA256 signature for license config
func calculateSignature(cfg *LicenseConfig) string {
	// Build canonical string from all license fields (excluding signature itself)
	// Fields are sorted alphabetically for consistency
	canonicalData := buildCanonicalString(cfg)

	// Compute HMAC-SHA256
	h := hmac.New(sha256.New, []byte(VendorSecret))
	h.Write([]byte(canonicalData))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// buildCanonicalString creates a canonical representation of license data
func buildCanonicalString(cfg *LicenseConfig) string {
	// Create map of all fields
	fields := map[string]string{
		"customer_name":   cfg.CustomerName,
		"expiry_date":     cfg.ExpiryDate,
		"licensed_mac":    normalizeMACAddress(cfg.LicensedMAC),
		"enable_2g":       boolToString(cfg.Enable2G),
		"enable_3g":       boolToString(cfg.Enable3G),
		"enable_4g":       boolToString(cfg.Enable4G),
		"enable_5g":       boolToString(cfg.Enable5G),
		"enable_map":      boolToString(cfg.EnableMAP),
		"enable_cap":      boolToString(cfg.EnableCAP),
		"enable_inap":     boolToString(cfg.EnableINAP),
		"enable_diameter": boolToString(cfg.EnableDiameter),
		"enable_http":     boolToString(cfg.EnableHTTP),
		"enable_gtp":      boolToString(cfg.EnableGTP),
		"max_subscribers": fmt.Sprintf("%d", cfg.MaxSubscribers),
		"max_tps":         fmt.Sprintf("%d", cfg.MaxTPS),
	}

	// Sort keys alphabetically for consistent ordering
	var keys []string
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build canonical string: key1=value1|key2=value2|...
	var parts []string
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, fields[key]))
	}

	return strings.Join(parts, "|")
}

// boolToString converts boolean to "1" or "0" string
func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// GenerateLicense generates a new license with signature
// This function would typically be used by the vendor's license generation tool
func GenerateLicense(cfg *LicenseConfig) *LicenseConfig {
	// Normalize MAC address
	cfg.LicensedMAC = normalizeMACAddress(cfg.LicensedMAC)

	// Calculate and set signature
	cfg.Signature = calculateSignature(cfg)

	return cfg
}

// PrintLicenseInfo prints human-readable license information
func PrintLicenseInfo(cfg *LicenseConfig) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  License Information")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Customer:       %s\n", cfg.CustomerName)
	fmt.Printf("Expiry Date:    %s\n", cfg.ExpiryDate)
	fmt.Printf("Licensed MAC:   %s\n", cfg.LicensedMAC)
	fmt.Println()
	fmt.Println("Generation Support:")
	fmt.Printf("  2G: %s\n", enabledStatus(cfg.Enable2G))
	fmt.Printf("  3G: %s\n", enabledStatus(cfg.Enable3G))
	fmt.Printf("  4G: %s\n", enabledStatus(cfg.Enable4G))
	fmt.Printf("  5G: %s\n", enabledStatus(cfg.Enable5G))
	fmt.Println()
	fmt.Println("Protocol Support:")
	fmt.Printf("  MAP:      %s\n", enabledStatus(cfg.EnableMAP))
	fmt.Printf("  CAP:      %s\n", enabledStatus(cfg.EnableCAP))
	fmt.Printf("  INAP:     %s\n", enabledStatus(cfg.EnableINAP))
	fmt.Printf("  Diameter: %s\n", enabledStatus(cfg.EnableDiameter))
	fmt.Printf("  HTTP:     %s\n", enabledStatus(cfg.EnableHTTP))
	fmt.Printf("  GTP:      %s\n", enabledStatus(cfg.EnableGTP))
	fmt.Println()
	fmt.Printf("Max Subscribers: %d\n", cfg.MaxSubscribers)
	fmt.Printf("Max TPS:         %d\n", cfg.MaxTPS)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func enabledStatus(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

// CheckLicenseFile validates license file and returns license config
func CheckLicenseFile(licensePath string) (*LicenseConfig, error) {
	// Read license file
	data, err := os.ReadFile(licensePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}

	// Parse license file (assuming simple key=value format)
	cfg, err := parseLicenseFile(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse license file: %w", err)
	}

	// Validate license
	if err := ValidateLicense(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// parseLicenseFile parses simple key=value license file
func parseLicenseFile(content string) (*LicenseConfig, error) {
	cfg := &LicenseConfig{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Parse each field
		switch key {
		case "customer_name":
			cfg.CustomerName = value
		case "expiry_date":
			cfg.ExpiryDate = value
		case "licensed_mac":
			cfg.LicensedMAC = value
		case "enable_2g":
			cfg.Enable2G = value == "1" || strings.ToLower(value) == "true"
		case "enable_3g":
			cfg.Enable3G = value == "1" || strings.ToLower(value) == "true"
		case "enable_4g":
			cfg.Enable4G = value == "1" || strings.ToLower(value) == "true"
		case "enable_5g":
			cfg.Enable5G = value == "1" || strings.ToLower(value) == "true"
		case "enable_map":
			cfg.EnableMAP = value == "1" || strings.ToLower(value) == "true"
		case "enable_cap":
			cfg.EnableCAP = value == "1" || strings.ToLower(value) == "true"
		case "enable_inap":
			cfg.EnableINAP = value == "1" || strings.ToLower(value) == "true"
		case "enable_diameter":
			cfg.EnableDiameter = value == "1" || strings.ToLower(value) == "true"
		case "enable_http":
			cfg.EnableHTTP = value == "1" || strings.ToLower(value) == "true"
		case "enable_gtp":
			cfg.EnableGTP = value == "1" || strings.ToLower(value) == "true"
		case "max_subscribers":
			fmt.Sscanf(value, "%d", &cfg.MaxSubscribers)
		case "max_tps":
			fmt.Sscanf(value, "%d", &cfg.MaxTPS)
		case "signature":
			cfg.Signature = value
		}
	}

	return cfg, nil
}

// GetDaysUntilExpiry returns number of days until license expires
func GetDaysUntilExpiry(expiryDate string) (int, error) {
	expiry, err := time.Parse("2006-01-02", expiryDate)
	if err != nil {
		return 0, err
	}

	duration := time.Until(expiry)
	days := int(duration.Hours() / 24)

	return days, nil
}

// IsFeatureEnabled checks if a specific feature is enabled in license
func IsFeatureEnabled(cfg *LicenseConfig, feature string) bool {
	switch strings.ToLower(feature) {
	case "2g":
		return cfg.Enable2G
	case "3g":
		return cfg.Enable3G
	case "4g", "lte":
		return cfg.Enable4G
	case "5g", "nr":
		return cfg.Enable5G
	case "map":
		return cfg.EnableMAP
	case "cap":
		return cfg.EnableCAP
	case "inap":
		return cfg.EnableINAP
	case "diameter":
		return cfg.EnableDiameter
	case "http", "http2":
		return cfg.EnableHTTP
	case "gtp":
		return cfg.EnableGTP
	default:
		return false
	}
}
