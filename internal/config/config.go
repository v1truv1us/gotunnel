package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/johncferguson/gotunnel/internal/auth"
	gotunnelErrors "github.com/johncferguson/gotunnel/internal/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for gotunnel
type Config struct {
	// Global settings
	Global GlobalConfig `yaml:"global"`
	
	// Tunnel definitions
	Tunnels []TunnelConfig `yaml:"tunnels"`
	
	// Proxy configuration
	Proxy ProxyConfig `yaml:"proxy"`
	
	// Logging configuration
	Logging LoggingConfig `yaml:"logging"`
	
	// Observability configuration
	Observability ObservabilityConfig `yaml:"observability"`
	
	// Authentication configuration
	Auth auth.Config `yaml:"auth"`
}

// GlobalConfig contains global settings
type GlobalConfig struct {
	// Environment (development, staging, production)
	Environment string `yaml:"environment" default:"development"`
	
	// Enable debug mode
	Debug bool `yaml:"debug" default:"false"`
	
	// Skip privilege check
	NoPrivilegeCheck bool `yaml:"no_privilege_check" default:"false"`
	
	// Default ports
	DefaultHTTPPort  int `yaml:"default_http_port" default:"80"`
	DefaultHTTPSPort int `yaml:"default_https_port" default:"443"`
	
	// Certificate directory
	CertsDir string `yaml:"certs_dir" default:"./certs"`
}

// TunnelConfig represents a single tunnel configuration
type TunnelConfig struct {
	// Domain name for the tunnel
	Domain string `yaml:"domain" validate:"required"`
	
	// Backend target
	Backend string `yaml:"backend" validate:"required"`
	
	// Enable HTTPS
	HTTPS *bool `yaml:"https"`
	
	// Custom ports (override global defaults)
	HTTPPort  *int `yaml:"http_port"`
	HTTPSPort *int `yaml:"https_port"`
	
	// Health check configuration
	HealthCheck *HealthCheckConfig `yaml:"health_check"`
	
	// Labels for metadata
	Labels map[string]string `yaml:"labels"`
}

// HealthCheckConfig defines health check settings
type HealthCheckConfig struct {
	// Health check path
	Path string `yaml:"path" default:"/health"`
	
	// Health check interval
	Interval string `yaml:"interval" default:"30s"`
	
	// Timeout for health check
	Timeout string `yaml:"timeout" default:"5s"`
	
	// Number of failures before marking as unhealthy
	FailureThreshold int `yaml:"failure_threshold" default:"3"`
}

// ProxyConfig contains proxy settings
type ProxyConfig struct {
	// Proxy mode (builtin, nginx, caddy, auto, config, none)
	Mode string `yaml:"mode" default:"auto"`
	
	// HTTP port for proxy
	HTTPPort int `yaml:"http_port" default:"80"`
	
	// HTTPS port for proxy
	HTTPSPort int `yaml:"https_port" default:"443"`
	
	// External proxy configuration
	External *ExternalProxyConfig `yaml:"external"`
}

// ExternalProxyConfig contains external proxy settings
type ExternalProxyConfig struct {
	// Nginx configuration
	Nginx *NginxConfig `yaml:"nginx"`
	
	// Caddy configuration
	Caddy *CaddyConfig `yaml:"caddy"`
}

// NginxConfig contains nginx-specific settings
type NginxConfig struct {
	// Path to nginx binary
	BinaryPath string `yaml:"binary_path" default:"nginx"`
	
	// Configuration file path
	ConfigPath string `yaml:"config_path" default:"/etc/nginx/nginx.conf"`
	
	// PID file path
	PIDPath string `yaml:"pid_path" default:"/var/run/nginx.pid"`
}

// CaddyConfig contains caddy-specific settings
type CaddyConfig struct {
	// Path to caddy binary
	BinaryPath string `yaml:"binary_path" default:"caddy"`
	
	// Configuration file path
	ConfigPath string `yaml:"config_path" default:"/etc/caddy/Caddyfile"`
	
	// PID file path
	PIDPath string `yaml:"pid_path" default:"/var/run/caddy.pid"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string `yaml:"level" default:"info"`
	
	// Log format (json, text)
	Format string `yaml:"format" default:"text"`
	
	// Log file path (empty for stdout)
	File string `yaml:"file"`
	
	// Enable file rotation
	Rotate bool `yaml:"rotate" default:"false"`
	
	// Maximum log file size
	MaxSize string `yaml:"max_size" default:"100MB"`
	
	// Maximum number of old log files
	MaxFiles int `yaml:"max_files" default:"3"`
}

// ObservabilityConfig contains observability settings
type ObservabilityConfig struct {
	// Sentry configuration
	Sentry *SentryConfig `yaml:"sentry"`
	
	// Prometheus configuration
	Prometheus *PrometheusConfig `yaml:"prometheus"`
	
	// OpenTelemetry configuration
	OpenTelemetry *OpenTelemetryConfig `yaml:"opentelemetry"`
}

// SentryConfig contains Sentry error tracking settings
type SentryConfig struct {
	// Sentry DSN
	DSN string `yaml:"dsn"`
	
	// Environment tag
	Environment string `yaml:"environment"`
	
	// Sample rate (0.0 to 1.0)
	SampleRate float64 `yaml:"sample_rate" default:"1.0"`
	
	// Enable performance monitoring
	EnablePerformanceMonitoring bool `yaml:"enable_performance_monitoring" default:"true"`
}

// PrometheusConfig contains Prometheus metrics settings
type PrometheusConfig struct {
	// Enable metrics collection
	Enabled bool `yaml:"enabled" default:"false"`
	
	// Metrics port
	Port int `yaml:"port" default:"9090"`
	
	// Metrics path
	Path string `yaml:"path" default:"/metrics"`
}

// OpenTelemetryConfig contains OpenTelemetry settings
type OpenTelemetryConfig struct {
	// Enable tracing
	Enabled bool `yaml:"enabled" default:"false"`
	
	// Service name
	ServiceName string `yaml:"service_name" default:"gotunnel"`
	
	// Service version
	ServiceVersion string `yaml:"service_version"`
	
	// Endpoint URL
	Endpoint string `yaml:"endpoint"`
	
	// Sample rate
	SampleRate float64 `yaml:"sample_rate" default:"1.0"`
}

// ConfigSource represents a source of configuration
type ConfigSource interface {
	Load() (*Config, error)
	Priority() int
}

// FileSource loads configuration from a file
type FileSource struct {
	Path     string
	priority int
}

// NewFileSource creates a new file source with the given priority
func NewFileSource(path string, priority int) *FileSource {
	return &FileSource{
		Path:     path,
		priority: priority,
	}
}

func (fs *FileSource) Load() (*Config, error) {
	if fs.Path == "" {
		return nil, nil
	}

	// Check if file exists
	if _, err := os.Stat(fs.Path); os.IsNotExist(err) {
		return nil, nil // File not found is not an error
	}

	data, err := os.ReadFile(fs.Path)
	if err != nil {
		return nil, gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeConfigLoad, 
			fmt.Sprintf("Failed to read config file: %s", fs.Path))
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeConfigParse,
			fmt.Sprintf("Failed to parse config file: %s", fs.Path))
	}

	return &config, nil
}

func (fs *FileSource) Priority() int {
	return fs.priority
}

// EnvSource loads configuration from environment variables
type EnvSource struct {
	Prefix string
}

func (es *EnvSource) Load() (*Config, error) {
	config := &Config{}
	
	// Global settings
	if env := os.Getenv(es.Prefix + "ENVIRONMENT"); env != "" {
		config.Global.Environment = env
	}
	if env := os.Getenv(es.Prefix + "DEBUG"); env != "" {
		config.Global.Debug = env == "true" || env == "1"
	}
	if env := os.Getenv(es.Prefix + "NO_PRIVILEGE_CHECK"); env != "" {
		config.Global.NoPrivilegeCheck = env == "true" || env == "1"
	}
	if env := os.Getenv(es.Prefix + "CERTS_DIR"); env != "" {
		config.Global.CertsDir = env
	}
	
	// Proxy settings
	if env := os.Getenv(es.Prefix + "PROXY_MODE"); env != "" {
		config.Proxy.Mode = env
	}
	if env := os.Getenv(es.Prefix + "PROXY_HTTP_PORT"); env != "" {
		if port, err := parseInt(env); err == nil {
			config.Proxy.HTTPPort = port
		}
	}
	if env := os.Getenv(es.Prefix + "PROXY_HTTPS_PORT"); env != "" {
		if port, err := parseInt(env); err == nil {
			config.Proxy.HTTPSPort = port
		}
	}
	
	// Logging settings
	if env := os.Getenv(es.Prefix + "LOG_LEVEL"); env != "" {
		config.Logging.Level = env
	}
	if env := os.Getenv(es.Prefix + "LOG_FORMAT"); env != "" {
		config.Logging.Format = env
	}
	if env := os.Getenv(es.Prefix + "LOG_FILE"); env != "" {
		config.Logging.File = env
	}
	
	// Observability settings
	if env := os.Getenv(es.Prefix + "SENTRY_DSN"); env != "" {
		if config.Observability.Sentry == nil {
			config.Observability.Sentry = &SentryConfig{}
		}
		config.Observability.Sentry.DSN = env
	}
	if env := os.Getenv(es.Prefix + "PROMETHEUS_ENABLED"); env != "" {
		if config.Observability.Prometheus == nil {
			config.Observability.Prometheus = &PrometheusConfig{}
		}
		config.Observability.Prometheus.Enabled = env == "true" || env == "1"
	}
	
	return config, nil
}

func (es *EnvSource) Priority() int {
	return 50 // Medium priority
}

// Manager manages configuration loading and merging
type Manager struct {
	sources []ConfigSource
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		sources: []ConfigSource{},
	}
}

// AddSource adds a configuration source
func (m *Manager) AddSource(source ConfigSource) {
	m.sources = append(m.sources, source)
}

// Load loads and merges configuration from all sources
func (m *Manager) Load() (*Config, error) {
	// Sort sources by priority (lowest first)
	sortedSources := make([]ConfigSource, len(m.sources))
	copy(sortedSources, m.sources)
	
	// Simple bubble sort by priority (ascending)
	for i := 0; i < len(sortedSources)-1; i++ {
		for j := 0; j < len(sortedSources)-i-1; j++ {
			if sortedSources[j].Priority() > sortedSources[j+1].Priority() {
				sortedSources[j], sortedSources[j+1] = sortedSources[j+1], sortedSources[j]
			}
		}
	}

	var merged *Config
	
	// Load and merge configurations
	for _, source := range sortedSources {
		config, err := source.Load()
		if err != nil {
			return nil, err
		}
		
		if config != nil {
			if merged == nil {
				merged = config
			} else {
				merged = mergeConfigs(merged, config)
			}
		}
	}

	if merged == nil {
		// No configuration loaded, use defaults
		merged = &Config{}
	}

	// Apply defaults at the end
	applyDefaults(merged)

	return merged, nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	source := NewFileSource(path, 100)
	return source.Load()
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv(prefix string) (*Config, error) {
	source := &EnvSource{Prefix: prefix}
	return source.Load()
}

// FindConfigFile finds the configuration file in standard locations
func FindConfigFile() string {
	locations := []string{
		"./gotunnel.yaml",
		"./gotunnel.yml",
		"./config/gotunnel.yaml",
		"./config/gotunnel.yml",
		"~/.gotunnel/gotunnel.yaml",
		"~/.gotunnel/gotunnel.yml",
		"/etc/gotunnel/gotunnel.yaml",
		"/etc/gotunnel/gotunnel.yml",
	}

	for _, location := range locations {
		// Expand ~ to home directory
		path := strings.Replace(location, "~", os.Getenv("HOME"), 1)
		
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// applyDefaults applies default values to configuration
func applyDefaults(config *Config) {
	// Global defaults
	if config.Global.Environment == "" {
		config.Global.Environment = "development"
	}
	if config.Global.DefaultHTTPPort == 0 {
		config.Global.DefaultHTTPPort = 80
	}
	if config.Global.DefaultHTTPSPort == 0 {
		config.Global.DefaultHTTPSPort = 443
	}
	if config.Global.CertsDir == "" {
		config.Global.CertsDir = "./certs"
	}
	
	// Proxy defaults
	if config.Proxy.Mode == "" {
		config.Proxy.Mode = "auto"
	}
	if config.Proxy.HTTPPort == 0 {
		config.Proxy.HTTPPort = 80
	}
	if config.Proxy.HTTPSPort == 0 {
		config.Proxy.HTTPSPort = 443
	}
	
	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	
	// Observability defaults
	if config.Observability.Sentry != nil && config.Observability.Sentry.SampleRate == 0 {
		config.Observability.Sentry.SampleRate = 1.0
	}
	if config.Observability.Prometheus != nil && config.Observability.Prometheus.Port == 0 {
		config.Observability.Prometheus.Port = 9090
	}
	if config.Observability.Prometheus != nil && config.Observability.Prometheus.Path == "" {
		config.Observability.Prometheus.Path = "/metrics"
	}
	if config.Observability.OpenTelemetry != nil {
		if config.Observability.OpenTelemetry.ServiceName == "" {
			config.Observability.OpenTelemetry.ServiceName = "gotunnel"
		}
		if config.Observability.OpenTelemetry.SampleRate == 0 {
			config.Observability.OpenTelemetry.SampleRate = 1.0
		}
	}
}

// mergeConfigs merges two configurations, with config2 overriding config1
func mergeConfigs(config1, config2 *Config) *Config {
	merged := *config1 // Copy config1
	
	// Merge global settings
	if config2.Global.Environment != "" {
		merged.Global.Environment = config2.Global.Environment
	}
	merged.Global.Debug = config2.Global.Debug || merged.Global.Debug
	merged.Global.NoPrivilegeCheck = config2.Global.NoPrivilegeCheck || merged.Global.NoPrivilegeCheck
	if config2.Global.DefaultHTTPPort != 0 {
		merged.Global.DefaultHTTPPort = config2.Global.DefaultHTTPPort
	}
	if config2.Global.DefaultHTTPSPort != 0 {
		merged.Global.DefaultHTTPSPort = config2.Global.DefaultHTTPSPort
	}
	if config2.Global.CertsDir != "" {
		merged.Global.CertsDir = config2.Global.CertsDir
	}
	
	// Merge tunnels (append)
	merged.Tunnels = append(merged.Tunnels, config2.Tunnels...)
	
	// Merge proxy settings
	if config2.Proxy.Mode != "" {
		merged.Proxy.Mode = config2.Proxy.Mode
	}
	if config2.Proxy.HTTPPort != 0 {
		merged.Proxy.HTTPPort = config2.Proxy.HTTPPort
	}
	if config2.Proxy.HTTPSPort != 0 {
		merged.Proxy.HTTPSPort = config2.Proxy.HTTPSPort
	}
	if config2.Proxy.External != nil {
		merged.Proxy.External = config2.Proxy.External
	}
	
	// Merge logging settings
	if config2.Logging.Level != "" {
		merged.Logging.Level = config2.Logging.Level
	}
	if config2.Logging.Format != "" {
		merged.Logging.Format = config2.Logging.Format
	}
	if config2.Logging.File != "" {
		merged.Logging.File = config2.Logging.File
	}
	merged.Logging.Rotate = config2.Logging.Rotate || merged.Logging.Rotate
	if config2.Logging.MaxSize != "" {
		merged.Logging.MaxSize = config2.Logging.MaxSize
	}
	if config2.Logging.MaxFiles != 0 {
		merged.Logging.MaxFiles = config2.Logging.MaxFiles
	}
	
	// Merge observability settings
	if config2.Observability.Sentry != nil {
		if merged.Observability.Sentry == nil {
			merged.Observability.Sentry = config2.Observability.Sentry
		} else {
			if config2.Observability.Sentry.DSN != "" {
				merged.Observability.Sentry.DSN = config2.Observability.Sentry.DSN
			}
			if config2.Observability.Sentry.Environment != "" {
				merged.Observability.Sentry.Environment = config2.Observability.Sentry.Environment
			}
			merged.Observability.Sentry.SampleRate = config2.Observability.Sentry.SampleRate
			merged.Observability.Sentry.EnablePerformanceMonitoring = config2.Observability.Sentry.EnablePerformanceMonitoring
		}
	}
	
	if config2.Observability.Prometheus != nil {
		if merged.Observability.Prometheus == nil {
			merged.Observability.Prometheus = config2.Observability.Prometheus
		} else {
			merged.Observability.Prometheus.Enabled = config2.Observability.Prometheus.Enabled
			if config2.Observability.Prometheus.Port != 0 {
				merged.Observability.Prometheus.Port = config2.Observability.Prometheus.Port
			}
			if config2.Observability.Prometheus.Path != "" {
				merged.Observability.Prometheus.Path = config2.Observability.Prometheus.Path
			}
		}
	}
	
	if config2.Observability.OpenTelemetry != nil {
		if merged.Observability.OpenTelemetry == nil {
			merged.Observability.OpenTelemetry = config2.Observability.OpenTelemetry
		} else {
			if config2.Observability.OpenTelemetry.ServiceName != "" {
				merged.Observability.OpenTelemetry.ServiceName = config2.Observability.OpenTelemetry.ServiceName
			}
			if config2.Observability.OpenTelemetry.ServiceVersion != "" {
				merged.Observability.OpenTelemetry.ServiceVersion = config2.Observability.OpenTelemetry.ServiceVersion
			}
			if config2.Observability.OpenTelemetry.Endpoint != "" {
				merged.Observability.OpenTelemetry.Endpoint = config2.Observability.OpenTelemetry.Endpoint
			}
			merged.Observability.OpenTelemetry.SampleRate = config2.Observability.OpenTelemetry.SampleRate
		}
	}
	
	return &merged
}

// Helper function to parse integers
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate tunnel configurations
	for i, tunnel := range c.Tunnels {
		if tunnel.Domain == "" {
			return gotunnelErrors.ValidationError("domain", "", "domain is required").WithContext("tunnel_index", i)
		}
		if tunnel.Backend == "" {
			return gotunnelErrors.ValidationError("backend", "", "backend is required").WithContext("tunnel_index", i)
		}
	}
	
	// Validate proxy mode
	validModes := []string{"builtin", "nginx", "caddy", "auto", "config", "none"}
	if !slices.Contains(validModes, c.Proxy.Mode) {
		return gotunnelErrors.ValidationError("proxy_mode", c.Proxy.Mode, 
			fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")))
	}
	
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !slices.Contains(validLevels, c.Logging.Level) {
		return gotunnelErrors.ValidationError("log_level", c.Logging.Level,
			fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")))
	}
	
	return nil
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeConfigLoad, "Failed to marshal configuration")
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeFilesystem, "Failed to create config directory")
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return gotunnelErrors.Wrap(err, gotunnelErrors.ErrCodeFilesystem, "Failed to write config file")
	}
	
	return nil
}