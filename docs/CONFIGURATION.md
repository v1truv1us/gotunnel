# Configuration Guide

This guide covers all configuration options for gotunnel, including configuration files, environment variables, and CLI flags.

## Configuration Precedence

gotunnel uses a hierarchical configuration system with the following precedence (highest to lowest):

1. **CLI Flags** - Command line arguments
2. **Environment Variables** - `GOTUNNEL_*` prefixed variables
3. **Configuration File** - YAML configuration
4. **Default Values** - Built-in defaults

## Configuration Files

### File Locations

gotunnel searches for configuration files in these locations (in order):

1. `./gotunnel.yaml` or `./gotunnel.yml`
2. `./config/gotunnel.yaml` or `./config/gotunnel.yml`
3. `~/.gotunnel/gotunnel.yaml` or `~/.gotunnel/gotunnel.yml`
4. `/etc/gotunnel/gotunnel.yaml` or `/etc/gotunnel/gotunnel.yml`

### File Format

Configuration files use YAML format. See the [complete example](#complete-configuration-example) below.

## Configuration Sections

### Global Settings

Controls overall application behavior.

```yaml
global:
  environment: "development"          # Runtime environment
  debug: false                       # Enable debug mode
  no_privilege_check: false           # Skip privilege requirements
  default_http_port: 80              # Default HTTP port
  default_https_port: 443             # Default HTTPS port
  certs_dir: "./certs"               # Certificate storage
```

**Field Details:**

- `environment`: `development`, `staging`, or `production`
- `debug`: Enables verbose logging and debugging features
- `no_privilege_check`: Bypasses admin privilege requirements
- `default_http_port`: Fallback HTTP port when privileges unavailable
- `default_https_port`: Fallback HTTPS port when privileges unavailable
- `certs_dir`: Directory for storing generated certificates

### Proxy Configuration

Controls how gotunnel handles HTTP/HTTPS proxying.

```yaml
proxy:
  mode: "auto"                      # Proxy mode
  http_port: 80                     # HTTP listening port
  https_port: 443                   # HTTPS listening port
  external:                          # External proxy settings
    nginx:
      binary_path: "nginx"           # Nginx binary path
      config_path: "/etc/nginx/nginx.conf"
      pid_path: "/var/run/nginx.pid"
    caddy:
      binary_path: "caddy"           # Caddy binary path
      config_path: "/etc/caddy/Caddyfile"
      pid_path: "/var/run/caddy.pid"
```

**Proxy Modes:**

- `auto`: Automatically detect and use best available proxy
- `builtin`: Use gotunnel's built-in Go proxy (recommended)
- `nginx`: Use external nginx server
- `caddy`: Use external caddy server
- `config`: Generate configuration files only
- `none`: No proxy, manual routing

### Logging Configuration

Controls application logging behavior.

```yaml
logging:
  level: "info"                      # Log level
  format: "text"                     # Log format
  file: ""                           # Log file path
  rotate: false                      # Enable log rotation
  max_size: "100MB"                 # Maximum file size
  max_files: 3                      # Number of backup files
```

**Log Levels:**

- `debug`: Detailed debugging information
- `info`: General information messages
- `warn`: Warning messages
- `error`: Error messages only

**Log Formats:**

- `text`: Human-readable text format
- `json`: Structured JSON format (recommended for production)

### Observability Configuration

Controls monitoring, tracing, and error tracking.

```yaml
observability:
  sentry:                            # Sentry error tracking
    dsn: ""                          # Sentry DSN
    environment: ""                    # Environment tag
    sample_rate: 1.0                  # Error sampling rate
    enable_performance_monitoring: true
  prometheus:                         # Prometheus metrics
    enabled: false                     # Enable metrics
    port: 9090                       # Metrics port
    path: "/metrics"                  # Metrics endpoint
  opentelemetry:                      # OpenTelemetry tracing
    enabled: false                     # Enable tracing
    service_name: "gotunnel"          # Service name
    service_version: ""                # Service version
    endpoint: ""                      # OTLP endpoint
    sample_rate: 1.0                  # Trace sampling rate
```

### Tunnel Definitions

Predefined tunnels that can be automatically started.

```yaml
tunnels:
  - domain: "web.local"              # Domain name
    backend: "http://localhost:3000"  # Backend URL
    https: true                       # Enable HTTPS
    http_port: 80                     # Custom HTTP port
    https_port: 443                   # Custom HTTPS port
    health_check:                     # Health check settings
      path: "/health"                 # Health check path
      interval: "30s"                 # Check interval
      timeout: "5s"                   # Request timeout
      failure_threshold: 3             # Failure threshold
    labels:                           # Metadata labels
      environment: "dev"
      service: "frontend"
```

## Environment Variables

All configuration options can be overridden using environment variables with the `GOTUNNEL_` prefix.

### Variable Mapping

Environment variables map to configuration fields using double underscores (`__`) for nested fields:

```bash
# Global settings
GOTUNNEL_GLOBAL_DEBUG=true
GOTUNNEL_GLOBAL_ENVIRONMENT=production
GOTUNNEL_GLOBAL_NO_PRIVILEGE_CHECK=true

# Proxy settings
GOTUNNEL_PROXY_MODE=builtin
GOTUNNEL_PROXY_HTTP_PORT=8080
GOTUNNEL_PROXY_HTTPS_PORT=8443

# Logging settings
GOTUNNEL_LOGGING_LEVEL=debug
GOTUNNEL_LOGGING_FORMAT=json
GOTUNNEL_LOGGING_FILE=/var/log/gotunnel.log

# Observability settings
GOTUNNEL_OBSERVABILITY_SENTRY_DSN=https://your-sentry-dsn
GOTUNNEL_OBSERVABILITY_PROMETHEUS_ENABLED=true
```

## Complete Configuration Example

```yaml
# Production-ready configuration
global:
  environment: "production"
  debug: false
  no_privilege_check: false
  default_http_port: 80
  default_https_port: 443
  certs_dir: "/etc/gotunnel/certs"

proxy:
  mode: "nginx"
  http_port: 80
  https_port: 443
  external:
    nginx:
      binary_path: "/usr/sbin/nginx"
      config_path: "/etc/nginx/sites-available/gotunnel"
      pid_path: "/var/run/nginx.pid"

logging:
  level: "info"
  format: "json"
  file: "/var/log/gotunnel/app.log"
  rotate: true
  max_size: "100MB"
  max_files: 10

observability:
  sentry:
    dsn: "${SENTRY_DSN}"
    environment: "production"
    sample_rate: 0.1
    enable_performance_monitoring: true
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  opentelemetry:
    enabled: true
    service_name: "gotunnel"
    service_version: "1.0.0"
    endpoint: "https://your-otlp-endpoint:4317"
    sample_rate: 0.1

tunnels:
  - domain: "app.local"
    backend: "http://localhost:3000"
    https: true
    health_check:
      path: "/health"
      interval: "30s"
      timeout: "5s"
      failure_threshold: 3
    labels:
      environment: "production"
      service: "frontend"
      team: "platform"
```

## Configuration Validation

gotunnel validates configuration on startup and provides helpful error messages:

```bash
# Invalid proxy mode
Error: invalid proxy mode "invalid". Valid options: auto, builtin, nginx, caddy, config, none

# Invalid log level
Error: invalid log level "invalid". Valid options: debug, info, warn, error

# Missing required tunnel fields
Error: tunnel configuration missing required field: domain
```

## Best Practices

### Development Environment

```yaml
global:
  debug: true
  no_privilege_check: true

proxy:
  mode: "builtin"
  http_port: 8080
  https_port: 8443

logging:
  level: "debug"
  format: "text"
```

### Production Environment

```yaml
global:
  environment: "production"
  debug: false

proxy:
  mode: "nginx"
  http_port: 80
  https_port: 443

logging:
  level: "info"
  format: "json"
  file: "/var/log/gotunnel/app.log"
  rotate: true

observability:
  sentry:
    dsn: "${SENTRY_DSN}"
    environment: "production"
    sample_rate: 0.1
  prometheus:
    enabled: true
```

### Corporate Environment

```yaml
global:
  no_privilege_check: true
  default_http_port: 8080
  default_https_port: 8443

proxy:
  mode: "builtin"  # Avoid external dependencies
  http_port: 8080
  https_port: 8443

logging:
  level: "info"
  format: "json"
  file: "./gotunnel.log"
```

## Troubleshooting

### Configuration Not Loading

1. Check file location matches one of the [search paths](#file-locations)
2. Verify YAML syntax: `python -c "import yaml; yaml.safe_load(open('gotunnel.yaml'))"`
3. Check file permissions: `ls -la gotunnel.yaml`

### Environment Variables Not Working

1. Verify prefix: `env | grep GOTUNNEL_`
2. Check for typos in variable names
3. Use double underscores for nested fields: `GOTUNNEL_PROXY__MODE`

### Validation Errors

1. Check required fields are present
2. Verify values are within allowed ranges
3. Ensure URLs and paths are valid

For more help, see the [main README](../README.md) or run `gotunnel --help`.
## Troubleshooting

### Common Issues and Solutions

#### Privilege Issues

**Problem**: `insufficient privileges: cannot bind to port 80`

**Solutions**:
1. **Run with elevated privileges**:
   ```bash
   # Linux/macOS
   sudo gotunnel start --port 3000 --domain app
   
   # Windows
   # Right-click and "Run as administrator"
   ```

2. **Skip privilege check and use high ports**:
   ```bash
   gotunnel --no-privilege-check start --port 3000 --domain app
   ```

3. **Configure proxy mode for non-privileged usage**:
   ```yaml
   global:
     no_privilege_check: true
   
   proxy:
     mode: "builtin"
     http_port: 8080
     https_port: 8443
   ```

#### Port Conflicts

**Problem**: `bind: address already in use`

**Solutions**:
1. **Find and stop the conflicting process**:
   ```bash
   # Find process using port 8080
   lsof -i :8080
   
   # Kill the process
   kill -9 <PID>
   ```

2. **Use different ports**:
   ```bash
   gotunnel --proxy-http-port 8888 --proxy-https-port 9443 start --port 3000 --domain app
   ```

3. **Configure in YAML**:
   ```yaml
   proxy:
     http_port: 8888
     https_port: 9443
   ```

#### Certificate Issues

**Problem**: `mkcert is not available for HTTPS certificate generation`

**Solutions**:
1. **Install mkcert**:
   ```bash
   # macOS
   brew install mkcert && mkcert -install
   
   # Linux
   # Follow: https://github.com/FiloSottile/mkcert#linux
   
   # Windows
   choco install mkcert && mkcert -install
   ```

2. **Use HTTP only**:
   ```bash
   gotunnel start --port 3000 --domain app --https=false
   ```

#### DNS Resolution Issues

**Problem**: Tunnel domain not resolving

**Solutions**:
1. **Check if DNS server is running**:
   ```bash
   # Look for gotunnel DNS process
   ps aux | grep gotunnel
   ```

2. **Verify /etc/hosts entry** (non-proxy mode):
   ```bash
   cat /etc/hosts | grep app.local
   # Should contain: 127.0.0.1	app.local
   ```

3. **Test mDNS discovery**:
   ```bash
   # Browse available services
   dns-sd -B _http._tcp local.
   ```

#### Configuration Errors

**Problem**: `Failed to load configuration`

**Solutions**:
1. **Validate YAML syntax**:
   ```bash
   # Install yamllint or use online validator
   python3 -c "import yaml; yaml.safe_load(open('gotunnel.yaml'))"
   ```

2. **Check file permissions**:
   ```bash
   ls -la gotunnel.yaml
   chmod 644 gotunnel.yaml
   ```

3. **Use minimal config**:
   ```yaml
   global:
     no_privilege_check: true
   
   proxy:
     mode: "builtin"
   ```

#### Proxy Mode Issues

**Problem**: `No route found for domain.local`

**Solutions**:
1. **Check proxy is running**:
   ```bash
   curl http://localhost:8080
   # Should show gotunnel proxy page
   ```

2. **Verify route registration**:
   ```bash
   gotunnel list
   # Should show your tunnel
   ```

3. **Test direct tunnel access**:
   ```bash
   # Find actual tunnel port (usually 9080+)
   curl http://localhost:9080
   ```

### Getting Help

1. **Enable debug logging**:
   ```bash
   gotunnel --debug start --port 3000 --domain app
   ```

2. **Check logs**:
   ```bash
   # With file logging configured
   tail -f /var/log/gotunnel/app.log
   ```

3. **Validate configuration**:
   ```bash
   gotunnel --help
   gotunnel list
   ```

4. **Report issues**:
   - Include: OS, gotunnel version, configuration, error message
   - Use `gotunnel --version` to get version info
   - Enable debug logging for detailed traces
