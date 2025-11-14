package dictionary

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Loader loads and manages vendor-specific dictionaries
type Loader struct {
	config       *Config
	dictionaries map[string]*VendorDictionary
	mu           sync.RWMutex
}

// Config holds dictionary configuration
type Config struct {
	BasePath     string
	VendorPaths  map[string]string
	AutoDetect   bool
}

// VendorDictionary represents a vendor-specific dictionary
type VendorDictionary struct {
	Vendor       string
	DiameterAVPs map[uint32]*AVPDefinition
	GTPIES       map[uint8]*IEDefinition
	MessageTypes map[string]string
	CauseCodes   map[int]string
	Features     map[string]interface{}
}

// AVPDefinition defines a Diameter AVP
type AVPDefinition struct {
	Code        uint32
	VendorID    uint32
	Name        string
	Type        string
	Description string
}

// IEDefinition defines a GTP Information Element
type IEDefinition struct {
	Type        uint8
	Name        string
	Description string
}

// NewLoader creates a new dictionary loader
func NewLoader(config *Config) *Loader {
	return &Loader{
		config:       config,
		dictionaries: make(map[string]*VendorDictionary),
	}
}

// LoadAll loads all vendor dictionaries
func (l *Loader) LoadAll() error {
	for vendor, path := range l.config.VendorPaths {
		fullPath := filepath.Join(l.config.BasePath, path)
		if err := l.LoadVendor(vendor, fullPath); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to load %s dictionary: %v\n", vendor, err)
			continue
		}
	}

	return nil
}

// LoadVendor loads a vendor-specific dictionary
func (l *Loader) LoadVendor(vendor, path string) error {
	dict := &VendorDictionary{
		Vendor:       vendor,
		DiameterAVPs: make(map[uint32]*AVPDefinition),
		GTPIES:       make(map[uint8]*IEDefinition),
		MessageTypes: make(map[string]string),
		CauseCodes:   make(map[int]string),
		Features:     make(map[string]interface{}),
	}

	// Load Diameter AVPs if file exists
	diameterFile := filepath.Join(path, "diameter.yaml")
	if _, err := os.Stat(diameterFile); err == nil {
		if err := l.loadDiameterAVPs(diameterFile, dict); err != nil {
			return fmt.Errorf("failed to load Diameter AVPs: %w", err)
		}
	}

	// Load GTP IEs if file exists
	gtpFile := filepath.Join(path, "gtp.yaml")
	if _, err := os.Stat(gtpFile); err == nil {
		if err := l.loadGTPIEs(gtpFile, dict); err != nil {
			return fmt.Errorf("failed to load GTP IEs: %w", err)
		}
	}

	// Load cause codes if file exists
	causesFile := filepath.Join(path, "causes.yaml")
	if _, err := os.Stat(causesFile); err == nil {
		if err := l.loadCauseCodes(causesFile, dict); err != nil {
			return fmt.Errorf("failed to load cause codes: %w", err)
		}
	}

	l.mu.Lock()
	l.dictionaries[vendor] = dict
	l.mu.Unlock()

	return nil
}

// loadDiameterAVPs loads Diameter AVP definitions
func (l *Loader) loadDiameterAVPs(filename string, dict *VendorDictionary) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var avps struct {
		AVPs []struct {
			Code        uint32 `yaml:"code"`
			VendorID    uint32 `yaml:"vendor_id"`
			Name        string `yaml:"name"`
			Type        string `yaml:"type"`
			Description string `yaml:"description"`
		} `yaml:"avps"`
	}

	if err := yaml.Unmarshal(data, &avps); err != nil {
		return err
	}

	for _, avp := range avps.AVPs {
		dict.DiameterAVPs[avp.Code] = &AVPDefinition{
			Code:        avp.Code,
			VendorID:    avp.VendorID,
			Name:        avp.Name,
			Type:        avp.Type,
			Description: avp.Description,
		}
	}

	return nil
}

// loadGTPIEs loads GTP IE definitions
func (l *Loader) loadGTPIEs(filename string, dict *VendorDictionary) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var ies struct {
		IEs []struct {
			Type        uint8  `yaml:"type"`
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		} `yaml:"ies"`
	}

	if err := yaml.Unmarshal(data, &ies); err != nil {
		return err
	}

	for _, ie := range ies.IEs {
		dict.GTPIES[ie.Type] = &IEDefinition{
			Type:        ie.Type,
			Name:        ie.Name,
			Description: ie.Description,
		}
	}

	return nil
}

// loadCauseCodes loads cause code definitions
func (l *Loader) loadCauseCodes(filename string, dict *VendorDictionary) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var causes struct {
		Causes map[int]string `yaml:"causes"`
	}

	if err := yaml.Unmarshal(data, &causes); err != nil {
		return err
	}

	dict.CauseCodes = causes.Causes

	return nil
}

// GetVendorDictionary retrieves a vendor dictionary
func (l *Loader) GetVendorDictionary(vendor string) (*VendorDictionary, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	dict, ok := l.dictionaries[vendor]
	return dict, ok
}

// GetAVPName retrieves AVP name for vendor and code
func (l *Loader) GetAVPName(vendor string, code uint32) string {
	dict, ok := l.GetVendorDictionary(vendor)
	if !ok {
		return fmt.Sprintf("AVP_%d", code)
	}

	if avp, ok := dict.DiameterAVPs[code]; ok {
		return avp.Name
	}

	return fmt.Sprintf("AVP_%d", code)
}

// GetIEName retrieves IE name for vendor and type
func (l *Loader) GetIEName(vendor string, ieType uint8) string {
	dict, ok := l.GetVendorDictionary(vendor)
	if !ok {
		return fmt.Sprintf("IE_%d", ieType)
	}

	if ie, ok := dict.GTPIES[ieType]; ok {
		return ie.Name
	}

	return fmt.Sprintf("IE_%d", ieType)
}

// GetCauseText retrieves cause text for vendor and code
func (l *Loader) GetCauseText(vendor string, code int) string {
	dict, ok := l.GetVendorDictionary(vendor)
	if !ok {
		return fmt.Sprintf("Cause_%d", code)
	}

	if text, ok := dict.CauseCodes[code]; ok {
		return text
	}

	return fmt.Sprintf("Cause_%d", code)
}
