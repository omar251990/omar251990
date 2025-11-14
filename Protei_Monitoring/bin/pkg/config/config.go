package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Application    ApplicationConfig    `yaml:"application"`
	Server         ServerConfig         `yaml:"server"`
	Ingestion      IngestionConfig      `yaml:"ingestion"`
	Protocols      ProtocolsConfig      `yaml:"protocols"`
	Correlation    CorrelationConfig    `yaml:"correlation"`
	Analytics      AnalyticsConfig      `yaml:"analytics"`
	Visualization  VisualizationConfig  `yaml:"visualization"`
	Storage        StorageConfig        `yaml:"storage"`
	Recommendations RecommendationsConfig `yaml:"recommendations"`
	Security       SecurityConfig       `yaml:"security"`
	Health         HealthConfig         `yaml:"health"`
	Vendors        VendorsConfig        `yaml:"vendors"`
	Performance    PerformanceConfig    `yaml:"performance"`
	Features       FeaturesConfig       `yaml:"features"`

	mu sync.RWMutex
}

// ApplicationConfig holds application identity
type ApplicationConfig struct {
	Name           string `yaml:"name"`
	Version        string `yaml:"version"`
	DeploymentPath string `yaml:"deployment_path"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

// IngestionConfig holds input source configuration
type IngestionConfig struct {
	Sources    []SourceConfig `yaml:"sources"`
	BufferSize int            `yaml:"buffer_size"`
	Workers    int            `yaml:"workers"`
	BatchSize  int            `yaml:"batch_size"`
}

// SourceConfig represents an input source
type SourceConfig struct {
	Type      string `yaml:"type"`
	Path      string `yaml:"path,omitempty"`
	Watch     bool   `yaml:"watch,omitempty"`
	Recursive bool   `yaml:"recursive,omitempty"`
	Pattern   string `yaml:"pattern,omitempty"`
	Enabled   bool   `yaml:"enabled,omitempty"`
	Interface string `yaml:"interface,omitempty"`
	Snaplen   int    `yaml:"snaplen,omitempty"`
	Promisc   bool   `yaml:"promisc,omitempty"`
}

// ProtocolsConfig holds protocol decoder settings
type ProtocolsConfig struct {
	MAP      MAPConfig      `yaml:"map"`
	CAP      CAPConfig      `yaml:"cap"`
	INAP     INAPConfig     `yaml:"inap"`
	Diameter DiameterConfig `yaml:"diameter"`
	GTP      GTPConfig      `yaml:"gtp"`
	PFCP     PFCPConfig     `yaml:"pfcp"`
	HTTP     HTTPConfig     `yaml:"http"`
	NGAP     NGAPConfig     `yaml:"ngap"`
	S1AP     S1APConfig     `yaml:"s1ap"`
	NAS      NASConfig      `yaml:"nas"`
}

type MAPConfig struct {
	Enabled bool  `yaml:"enabled"`
	Version []int `yaml:"version"`
}

type CAPConfig struct {
	Enabled bool  `yaml:"enabled"`
	Version []int `yaml:"version"`
}

type INAPConfig struct {
	Enabled bool  `yaml:"enabled"`
	Version []int `yaml:"version"`
}

type DiameterConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Applications  []string `yaml:"applications"`
	VendorSupport []string `yaml:"vendor_support"`
}

type GTPConfig struct {
	Enabled  bool  `yaml:"enabled"`
	Versions []int `yaml:"versions"`
}

type PFCPConfig struct {
	Enabled bool `yaml:"enabled"`
}

type HTTPConfig struct {
	Enabled  bool      `yaml:"enabled"`
	Versions []float64 `yaml:"versions"`
}

type NGAPConfig struct {
	Enabled bool `yaml:"enabled"`
}

type S1APConfig struct {
	Enabled bool `yaml:"enabled"`
}

type NASConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Generations []string `yaml:"generations"`
}

// CorrelationConfig holds correlation engine settings
type CorrelationConfig struct {
	TIDCacheSize      int      `yaml:"tid_cache_size"`
	TIDTTL            int      `yaml:"tid_ttl"`
	CorrelationFields []string `yaml:"correlation_fields"`
	SessionTimeout    int      `yaml:"session_timeout"`
	E2ETracking       bool     `yaml:"e2e_tracking"`
}

// AnalyticsConfig holds analytics and KPI settings
type AnalyticsConfig struct {
	KPIs            KPIConfig            `yaml:"kpis"`
	Roaming         RoamingConfig        `yaml:"roaming"`
	FailureAnalysis FailureAnalysisConfig `yaml:"failure_analysis"`
}

type KPIConfig struct {
	Enabled              bool     `yaml:"enabled"`
	CalculationInterval  int      `yaml:"calculation_interval"`
	Procedures           []string `yaml:"procedures"`
	Metrics              []string `yaml:"metrics"`
}

type RoamingConfig struct {
	Enabled          bool   `yaml:"enabled"`
	TrackInbound     bool   `yaml:"track_inbound"`
	TrackOutbound    bool   `yaml:"track_outbound"`
	HeatmapResolution string `yaml:"heatmap_resolution"`
}

type FailureAnalysisConfig struct {
	Enabled               bool    `yaml:"enabled"`
	ThresholdFailureRate  float64 `yaml:"threshold_failure_rate"`
	ThresholdLatencyMs    int     `yaml:"threshold_latency_ms"`
}

// VisualizationConfig holds visualization settings
type VisualizationConfig struct {
	LadderDiagrams LadderDiagramConfig `yaml:"ladder_diagrams"`
	Heatmaps       HeatmapConfig       `yaml:"heatmaps"`
	Dashboard      DashboardConfig     `yaml:"dashboard"`
}

type LadderDiagramConfig struct {
	Enabled              bool   `yaml:"enabled"`
	Format               string `yaml:"format"`
	MaxMessagesPerDiagram int   `yaml:"max_messages_per_diagram"`
	OutputPath           string `yaml:"output_path"`
	AutoLabelNodes       bool   `yaml:"auto_label_nodes"`
}

type HeatmapConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
}

type DashboardConfig struct {
	Enabled         bool `yaml:"enabled"`
	RefreshInterval int  `yaml:"refresh_interval"`
	RealTime        bool `yaml:"real_time"`
}

// StorageConfig holds storage and output settings
type StorageConfig struct {
	Logs   LogConfig   `yaml:"logs"`
	Events EventConfig `yaml:"events"`
	CDR    CDRConfig   `yaml:"cdr"`
}

type LogConfig struct {
	Path         string `yaml:"path"`
	Format       string `yaml:"format"`
	Level        string `yaml:"level"`
	MaxSizeMB    int    `yaml:"max_size_mb"`
	MaxBackups   int    `yaml:"max_backups"`
	MaxAgeDays   int    `yaml:"max_age_days"`
	Compress     bool   `yaml:"compress"`
}

type EventConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Path          string `yaml:"path"`
	Format        string `yaml:"format"`
	Rotation      string `yaml:"rotation"`
	RetentionDays int    `yaml:"retention_days"`
}

type CDRConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Path          string   `yaml:"path"`
	Format        string   `yaml:"format"`
	Fields        []string `yaml:"fields"`
	Rotation      string   `yaml:"rotation"`
	RetentionDays int      `yaml:"retention_days"`
}

// RecommendationsConfig holds recommendation engine settings
type RecommendationsConfig struct {
	Enabled       bool   `yaml:"enabled"`
	RulesFile     string `yaml:"rules_file"`
	MLEnabled     bool   `yaml:"ml_enabled"`
	MLModelPath   string `yaml:"ml_model_path"`
}

// SecurityConfig holds security settings
type SecurityConfig struct {
	AuthEnabled bool     `yaml:"auth_enabled"`
	AuthType    string   `yaml:"auth_type"`
	RBACEnabled bool     `yaml:"rbac_enabled"`
	LocalOnly   bool     `yaml:"local_only"`
	AllowedIPs  []string `yaml:"allowed_ips"`
}

// HealthConfig holds health check settings
type HealthConfig struct {
	Enabled       bool            `yaml:"enabled"`
	CheckInterval int             `yaml:"check_interval"`
	Endpoints     []string        `yaml:"endpoints"`
	Watchdog      WatchdogConfig  `yaml:"watchdog"`
}

type WatchdogConfig struct {
	Enabled          bool `yaml:"enabled"`
	Timeout          int  `yaml:"timeout"`
	RestartOnFailure bool `yaml:"restart_on_failure"`
}

// VendorsConfig holds vendor dictionary settings
type VendorsConfig struct {
	Dictionaries map[string]string `yaml:"dictionaries"`
	AutoDetect   bool              `yaml:"auto_detect"`
}

// PerformanceConfig holds performance tuning settings
type PerformanceConfig struct {
	MaxGoroutines    int  `yaml:"max_goroutines"`
	GCPercent        int  `yaml:"gc_percent"`
	NUMAAware        bool `yaml:"numa_aware"`
	MaxMemoryMB      int  `yaml:"max_memory_mb"`
	AsyncProcessing  bool `yaml:"async_processing"`
	PipelineDepth    int  `yaml:"pipeline_depth"`
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	HotReload             bool `yaml:"hot_reload"`
	GracefulRestart       bool `yaml:"graceful_restart"`
	AutoCleanup           bool `yaml:"auto_cleanup"`
	MetricsExport         bool `yaml:"metrics_export"`
	PrometheusCompatible  bool `yaml:"prometheus_compatible"`
}

// Global config instance
var globalConfig *Config
var configMu sync.RWMutex

// Load reads configuration from a YAML file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set global config
	configMu.Lock()
	globalConfig = &cfg
	configMu.Unlock()

	return &cfg, nil
}

// Get returns the global configuration instance
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// Reload reloads configuration from disk (hot reload)
func Reload(configPath string) error {
	newConfig, err := Load(configPath)
	if err != nil {
		return err
	}

	configMu.Lock()
	globalConfig = newConfig
	configMu.Unlock()

	return nil
}

// Validate performs configuration validation
func (c *Config) Validate() error {
	if c.Application.Name == "" {
		return fmt.Errorf("application name is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Ingestion.Workers < 1 {
		return fmt.Errorf("at least 1 ingestion worker is required")
	}

	return nil
}

// GetAddr returns the server address in host:port format
func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
