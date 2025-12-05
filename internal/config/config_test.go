package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gotunnelErrors "github.com/johncferguson/gotunnel/internal/errors"
)

func TestConfigDefaults(t *testing.T) {
	config := &Config{}
	applyDefaults(config)

	assert.Equal(t, "development", config.Global.Environment)
	assert.Equal(t, 80, config.Global.DefaultHTTPPort)
	assert.Equal(t, 443, config.Global.DefaultHTTPSPort)
	assert.Equal(t, "./certs", config.Global.CertsDir)
	assert.Equal(t, "auto", config.Proxy.Mode)
	assert.Equal(t, 80, config.Proxy.HTTPPort)
	assert.Equal(t, 443, config.Proxy.HTTPSPort)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
}

func TestFileSource(t *testing.T) {
	t.Run("load existing file", func(t *testing.T) {
		// Create temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test.yaml")
		
		configContent := `
global:
  environment: "test"
  debug: true
  certs_dir: "/tmp/certs"

tunnels:
  - domain: "test.local"
    backend: "http://localhost:3000"
    https: true
    labels:
      env: "test"

proxy:
  mode: "nginx"
  http_port: 8080

logging:
  level: "debug"
  format: "json"
  file: "/tmp/gotunnel.log"
`
		
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		source := &FileSource{Path: configPath, priority: 100}
		config, err := source.Load()
		
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "test", config.Global.Environment)
		assert.True(t, config.Global.Debug)
		assert.Equal(t, "/tmp/certs", config.Global.CertsDir)
		assert.Len(t, config.Tunnels, 1)
		assert.Equal(t, "test.local", config.Tunnels[0].Domain)
		assert.Equal(t, "http://localhost:3000", config.Tunnels[0].Backend)
		assert.True(t, *config.Tunnels[0].HTTPS)
		assert.Equal(t, "nginx", config.Proxy.Mode)
		assert.Equal(t, 8080, config.Proxy.HTTPPort)
		assert.Equal(t, "debug", config.Logging.Level)
		assert.Equal(t, "json", config.Logging.Format)
		assert.Equal(t, "/tmp/gotunnel.log", config.Logging.File)
	})

	t.Run("non-existent file", func(t *testing.T) {
		source := &FileSource{Path: "/non/existent/file.yaml", priority: 100}
		config, err := source.Load()
		
		assert.NoError(t, err)
		assert.Nil(t, config)
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		
		err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
		require.NoError(t, err)

		source := &FileSource{Path: configPath, priority: 100}
		config, err := source.Load()
		
		assert.Error(t, err)
		assert.Nil(t, config)
		
		gotunnelErr, ok := gotunnelErrors.IsGotunnelError(err)
		assert.True(t, ok)
		assert.NotNil(t, gotunnelErr)
		assert.Equal(t, gotunnelErrors.ErrCodeConfigParse, gotunnelErr.Code)
	})

	t.Run("empty path", func(t *testing.T) {
		source := &FileSource{Path: "", priority: 100}
		config, err := source.Load()
		
		assert.NoError(t, err)
		assert.Nil(t, config)
	})
}

func TestEnvSource(t *testing.T) {
	// Save original environment variables
	originalEnv := make(map[string]string)
	envVars := []string{
		"GOTUNNEL_ENVIRONMENT",
		"GOTUNNEL_DEBUG",
		"GOTUNNEL_NO_PRIVILEGE_CHECK",
		"GOTUNNEL_CERTS_DIR",
		"GOTUNNEL_PROXY_MODE",
		"GOTUNNEL_PROXY_HTTP_PORT",
		"GOTUNNEL_PROXY_HTTPS_PORT",
		"GOTUNNEL_LOG_LEVEL",
		"GOTUNNEL_LOG_FORMAT",
		"GOTUNNEL_LOG_FILE",
		"GOTUNNEL_SENTRY_DSN",
		"GOTUNNEL_PROMETHEUS_ENABLED",
	}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}

	// Restore environment variables after test
	defer func() {
		for env, val := range originalEnv {
			os.Setenv(env, val)
		}
		for _, env := range envVars {
			if _, exists := originalEnv[env]; !exists {
				os.Unsetenv(env)
			}
		}
	}()

	t.Run("load from environment", func(t *testing.T) {
		os.Setenv("GOTUNNEL_ENVIRONMENT", "production")
		os.Setenv("GOTUNNEL_DEBUG", "true")
		os.Setenv("GOTUNNEL_CERTS_DIR", "/app/certs")
		os.Setenv("GOTUNNEL_PROXY_MODE", "caddy")
		os.Setenv("GOTUNNEL_PROXY_HTTP_PORT", "8080")
		os.Setenv("GOTUNNEL_LOG_LEVEL", "warn")
		os.Setenv("GOTUNNEL_LOG_FORMAT", "json")
		os.Setenv("GOTUNNEL_SENTRY_DSN", "https://test@sentry.io/123")
		os.Setenv("GOTUNNEL_PROMETHEUS_ENABLED", "true")

		source := &EnvSource{Prefix: "GOTUNNEL_"}
		config, err := source.Load()
		
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "production", config.Global.Environment)
		assert.True(t, config.Global.Debug)
		assert.Equal(t, "/app/certs", config.Global.CertsDir)
		assert.Equal(t, "caddy", config.Proxy.Mode)
		assert.Equal(t, 8080, config.Proxy.HTTPPort)
		assert.Equal(t, "warn", config.Logging.Level)
		assert.Equal(t, "json", config.Logging.Format)
		assert.NotNil(t, config.Observability.Sentry)
		assert.Equal(t, "https://test@sentry.io/123", config.Observability.Sentry.DSN)
		assert.NotNil(t, config.Observability.Prometheus)
		assert.True(t, config.Observability.Prometheus.Enabled)
	})

	t.Run("boolean parsing", func(t *testing.T) {
		testCases := []struct {
			value    string
			expected bool
		}{
			{"true", true},
			{"1", true},
			{"false", false},
			{"0", false},
			{"", false},
		}

		for _, tc := range testCases {
			os.Setenv("GOTUNNEL_DEBUG", tc.value)
			source := &EnvSource{Prefix: "GOTUNNEL_"}
			config, err := source.Load()
			
			require.NoError(t, err)
			assert.Equal(t, tc.expected, config.Global.Debug, "value: %s", tc.value)
		}
	})
}

func TestManager(t *testing.T) {
	t.Run("load and merge configurations", func(t *testing.T) {
		// Create temporary config files
		tmpDir := t.TempDir()
		
		// Base config
		baseConfigPath := filepath.Join(tmpDir, "base.yaml")
		baseContent := `
global:
  environment: "development"
  debug: false
  certs_dir: "./certs"

tunnels:
  - domain: "app.local"
    backend: "http://localhost:3000"

proxy:
  mode: "auto"
  http_port: 80
`
		err := os.WriteFile(baseConfigPath, []byte(baseContent), 0644)
		require.NoError(t, err)

		// Override config
		overrideConfigPath := filepath.Join(tmpDir, "override.yaml")
		overrideContent := `
global:
  environment: "production"
  debug: true
  certs_dir: "./certs"  # Keep this from base

tunnels:
  - domain: "api.local"
    backend: "http://localhost:8080"

proxy:
  mode: "auto"  # Keep this from base
  http_port: 8080
  https_port: 443  # Keep this from base
`
		err = os.WriteFile(overrideConfigPath, []byte(overrideContent), 0644)
		require.NoError(t, err)

			// Create manager and add sources
		manager := NewManager()
		manager.AddSource(&FileSource{Path: baseConfigPath, priority: 50})
		manager.AddSource(&FileSource{Path: overrideConfigPath, priority: 100}) // Higher priority

		config, err := manager.Load()
		require.NoError(t, err)
		assert.NotNil(t, config)

	// Check that overrides were applied
		assert.Equal(t, "production", config.Global.Environment) // Overridden
		assert.True(t, config.Global.Debug)                        // Overridden
		assert.Equal(t, "./certs", config.Global.CertsDir)         // From base
		assert.Equal(t, 8080, config.Proxy.HTTPPort)               // Overridden
		
		// Check that tunnels were merged
		assert.Len(t, config.Tunnels, 2)
		domains := make(map[string]bool)
		for _, tunnel := range config.Tunnels {
			domains[tunnel.Domain] = true
		}
		assert.True(t, domains["app.local"])
		assert.True(t, domains["api.local"])
	})

	t.Run("no sources", func(t *testing.T) {
		manager := NewManager()
		config, err := manager.Load()
		
		require.NoError(t, err)
		assert.NotNil(t, config)
		// Should have defaults applied
		assert.Equal(t, "development", config.Global.Environment)
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			Global: GlobalConfig{
				Environment: "production",
			},
			Tunnels: []TunnelConfig{
				{
					Domain:  "test.local",
					Backend: "http://localhost:3000",
				},
			},
			Proxy: ProxyConfig{
				Mode: "auto",
			},
			Logging: LoggingConfig{
				Level: "info",
			},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid tunnel", func(t *testing.T) {
		config := &Config{
			Tunnels: []TunnelConfig{
				{
					Domain:  "", // Invalid
					Backend: "http://localhost:3000",
				},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		
		gotunnelErr, ok := gotunnelErrors.IsGotunnelError(err)
		assert.True(t, ok)
		assert.NotNil(t, gotunnelErr)
		assert.Equal(t, gotunnelErrors.ErrCodeValidationDomain, gotunnelErr.Code)
	})

	t.Run("invalid proxy mode", func(t *testing.T) {
		config := &Config{
			Proxy: ProxyConfig{
				Mode: "invalid",
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		
		gotunnelErr, ok := gotunnelErrors.IsGotunnelError(err)
		assert.True(t, ok)
		assert.NotNil(t, gotunnelErr)
		assert.Equal(t, gotunnelErrors.ErrorCode("VALIDATION_PROXY_MODE"), gotunnelErr.Code)
	})

	t.Run("invalid log level", func(t *testing.T) {
		config := &Config{
			Logging: LoggingConfig{
				Level: "invalid",
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		
		gotunnelErr, ok := gotunnelErrors.IsGotunnelError(err)
		assert.True(t, ok)
		assert.NotNil(t, gotunnelErr)
	})
}

func TestConfigSave(t *testing.T) {
	config := &Config{
		Global: GlobalConfig{
			Environment: "test",
			Debug:      true,
		},
		Tunnels: []TunnelConfig{
			{
				Domain:  "test.local",
				Backend: "http://localhost:3000",
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "saved.yaml")

	err := config.Save(configPath)
	require.NoError(t, err)

	// Verify file was created and contains valid YAML
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify
	loadedConfig, err := LoadFromFile(configPath)
	require.NoError(t, err)
	
	assert.Equal(t, config.Global.Environment, loadedConfig.Global.Environment)
	assert.Equal(t, config.Global.Debug, loadedConfig.Global.Debug)
	assert.Len(t, loadedConfig.Tunnels, 1)
	assert.Equal(t, config.Tunnels[0].Domain, loadedConfig.Tunnels[0].Domain)
	assert.Equal(t, config.Tunnels[0].Backend, loadedConfig.Tunnels[0].Backend)
}

func TestFindConfigFile(t *testing.T) {
	t.Run("find existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create config file
		configPath := filepath.Join(tmpDir, "gotunnel.yaml")
		err := os.WriteFile(configPath, []byte("test: config"), 0644)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		found := FindConfigFile()
		assert.NotEmpty(t, found)
		assert.Contains(t, found, "gotunnel.yaml")
	})

	t.Run("no config found", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Change to empty temp directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		found := FindConfigFile()
		assert.Empty(t, found)
	})
}

func TestLoadFromEnv(t *testing.T) {
	// Save and restore environment
	originalEnv := make(map[string]string)
	envVars := []string{"TEST_ENVIRONMENT", "TEST_DEBUG", "TEST_PROXY_MODE"}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}

	defer func() {
		for env, val := range originalEnv {
			os.Setenv(env, val)
		}
		for _, env := range envVars {
			if _, exists := originalEnv[env]; !exists {
				os.Unsetenv(env)
			}
		}
	}()

	os.Setenv("TEST_ENVIRONMENT", "test")
	os.Setenv("TEST_DEBUG", "true")
	os.Setenv("TEST_PROXY_MODE", "nginx")

	config, err := LoadFromEnv("TEST_")
	require.NoError(t, err)
	
	assert.Equal(t, "test", config.Global.Environment)
	assert.True(t, config.Global.Debug)
	assert.Equal(t, "nginx", config.Proxy.Mode)
}

func TestMergeConfigs(t *testing.T) {
	config1 := &Config{
		Global: GlobalConfig{
			Environment: "development",
			Debug:      false,
		},
		Tunnels: []TunnelConfig{
			{Domain: "app1.local", Backend: "http://localhost:3000"},
		},
	}

	config2 := &Config{
		Global: GlobalConfig{
			Environment: "production", // Override
			Debug:      true,          // Override
		},
		Tunnels: []TunnelConfig{
			{Domain: "app2.local", Backend: "http://localhost:8080"},
		},
	}

	merged := mergeConfigs(config1, config2)

	assert.Equal(t, "production", merged.Global.Environment)
	assert.True(t, merged.Global.Debug)
	assert.Len(t, merged.Tunnels, 2)
}