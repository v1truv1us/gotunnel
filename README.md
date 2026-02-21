# gotunnel 🚇 

[![Go Version](https://img.shields.io/github/go-mod/go-version/johncferguson/gotunnel)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/johncferguson/gotunnel?include_prereleases)](https://github.com/johncferguson/gotunnel/releases)
[![CI/CD](https://github.com/johncferguson/gotunnel/workflows/CI%2FCD%20Pipeline/badge.svg)](https://github.com/johncferguson/gotunnel/actions)
[![Docker](https://img.shields.io/badge/docker-available-blue.svg)](https://ghcr.io/johncferguson/gotunnel)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/johncferguson/gotunnel)](https://goreportcard.com/report/github.com/johncferguson/gotunnel)

**Create secure local tunnels for development without root privileges**

gotunnel provides secure HTTP/HTTPS tunnels for local development with built-in proxy capabilities, OpenTelemetry observability, and enterprise-friendly configuration options.

## ✨ Features

- 🔐 **No Root Required**: Works without administrator privileges
- 🌐 **Built-in HTTP Proxy**: No external dependencies needed
- 🏢 **Enterprise Ready**: Works with corporate firewalls and proxy settings
- 🔄 **Multiple Backends**: Support for nginx, Caddy, or built-in proxy
- 📊 **OpenTelemetry**: Full observability with traces, metrics, and logs
- 🖥️ **Cross-Platform**: Native support for macOS, Linux, and Windows
- 🐳 **Docker Ready**: Full containerization support with Compose
- 🔍 **Auto-Discovery**: mDNS support for network-wide access

## 🚀 Quick Start

### Installation

#### Package Managers

**Homebrew (macOS/Linux):**
```bash
brew tap ferg-cod3s/gotunnel
brew install gotunnel
```

**Scoop (Windows):**
```powershell
scoop bucket add ferg-cod3s https://github.com/ferg-cod3s/scoop-bucket
scoop install gotunnel
```

**APT (Debian/Ubuntu):**
```bash
curl -fsSL https://github.com/johncferguson/gotunnel/releases/latest/download/gotunnel_0.1.0-beta_amd64.deb -o gotunnel.deb
sudo dpkg -i gotunnel.deb
```

**AUR (Arch Linux):**
```bash
yay -S gotunnel
# or: paru -S gotunnel
```

#### Direct Installation

**Install Script (Unix):**
```bash
curl -sSL https://raw.githubusercontent.com/johncferguson/gotunnel/main/scripts/install.sh | bash
```

**Go Install:**
```bash
go install github.com/johncferguson/gotunnel/cmd/gotunnel@latest
```

**Docker:**
```bash
docker run --rm -p 80:80 -p 443:443 ghcr.io/johncferguson/gotunnel:latest
```

### Basic Usage

**Start a tunnel (no privileges required):**
```bash
# Tunnel your app running on port 3000
gotunnel --proxy=builtin --no-privilege-check start \
  --port 3000 --domain myapp --https=false
```

**Access your app:**
- Local: `http://localhost:3000`
- Tunnel: `http://myapp.local:8080` (with non-privileged ports)

**With HTTPS (default):**
```bash
gotunnel start --port 3000 --domain myapp
# Access at: https://myapp.local
```

**Multiple tunnels:**
```bash
# Terminal 1: Frontend
gotunnel start --port 3000 --domain frontend

# Terminal 2: API  
gotunnel start --port 8080 --domain api

# Terminal 3: Database Admin
gotunnel start --port 5432 --domain pgadmin
```

## 🏢 Enterprise Usage

### Custom Proxy Ports
```bash
# Use non-standard ports for corporate environments
gotunnel --proxy=builtin --proxy-http-port 8080 --proxy-https-port 8443 \
  start --port 3000 --domain myapp
```

### Configuration File
```bash
# Use configuration file (recommended for teams)
cp configs/gotunnel.example.yaml ~/.config/gotunnel/config.yaml
gotunnel start --port 3000 --domain myapp
```

### Generate Proxy Config Only
```bash
# Generate nginx/Caddy configuration without running proxy
gotunnel --proxy=config start --port 3000 --domain myapp
```

## 🐳 Docker Deployment

### Docker Compose (Recommended)

**Quick Start:**
```yaml
version: '3.8'
services:
  gotunnel:
    image: ghcr.io/johncferguson/gotunnel:latest
    ports:
      - "80:80"
      - "443:443"
    environment:
      - ENVIRONMENT=production
      - SENTRY_DSN=${SENTRY_DSN}
    volumes:
      - ./certs:/app/certs
      - ./config:/app/config
    restart: unless-stopped
```

**With Monitoring Stack:**
```bash
# Copy the provided docker-compose.yml
docker-compose --profile monitoring up -d

# Access services
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gotunnel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gotunnel
  template:
    metadata:
      labels:
        app: gotunnel
    spec:
      containers:
      - name: gotunnel
        image: ghcr.io/johncferguson/gotunnel:latest
        ports:
        - containerPort: 80
        - containerPort: 443
        env:
        - name: ENVIRONMENT
          value: "production"
```

## ⚙️ Configuration

### Environment Variables

Environment variables override configuration file settings and use the `GOTUNNEL_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Runtime environment | `development` |
| `SENTRY_DSN` | Sentry error tracking DSN | - |
| `DEBUG` | Enable debug mode | `false` |
| `GOTUNNEL_PROXY` | Proxy mode | `auto` |
| `GOTUNNEL_PROXY_HTTP_PORT` | HTTP proxy port | `80` |
| `GOTUNNEL_PROXY_HTTPS_PORT` | HTTPS proxy port | `443` |
| `GOTUNNEL_GLOBAL_DEBUG` | Global debug flag | `false` |
| `GOTUNNEL_GLOBAL_NO_PRIVILEGE_CHECK` | Skip privilege check | `false` |
| `GOTUNNEL_GLOBAL_CERTS_DIR` | Certificate directory | `./certs` |
| `GOTUNNEL_LOGGING_LEVEL` | Log level | `info` |
| `GOTUNNEL_LOGGING_FORMAT` | Log format | `text` |
| `GOTUNNEL_LOGGING_FILE` | Log file path | `stdout` |
| `GOTUNNEL_OBSERVABILITY_SENTRY_DSN` | Sentry DSN | - |
| `GOTUNNEL_OBSERVABILITY_SENTRY_ENVIRONMENT` | Sentry environment | - |
| `GOTUNNEL_OBSERVABILITY_PROMETHEUS_ENABLED` | Enable Prometheus | `false` |
| `GOTUNNEL_OBSERVABILITY_PROMETHEUS_PORT` | Prometheus port | `9090` |

**Example with Environment Variables:**
```bash
export GOTUNNEL_GLOBAL_DEBUG=true
export GOTUNNEL_GLOBAL_NO_PRIVILEGE_CHECK=true
export GOTUNNEL_PROXY=builtin
export GOTUNNEL_LOGGING_LEVEL=debug
export GOTUNNEL_LOGGING_FORMAT=json

gotunnel start --port 3000 --domain myapp
```

### Configuration File

gotunnel looks for configuration files in these locations (in order):
- `./gotunnel.yaml` or `./gotunnel.yml`
- `./config/gotunnel.yaml` or `./config/gotunnel.yml`
- `~/.gotunnel/gotunnel.yaml` or `~/.gotunnel/gotunnel.yml`
- `/etc/gotunnel/gotunnel.yaml` or `/etc/gotunnel/gotunnel.yml`

#### Complete Configuration Example

```yaml
# Global settings
global:
  environment: "development"          # development, staging, production
  debug: true                       # Enable debug mode
  no_privilege_check: true           # Skip privilege check
  default_http_port: 8080           # Default HTTP port
  default_https_port: 8443          # Default HTTPS port
  certs_dir: "./certs"              # Certificate storage directory

# Proxy configuration
proxy:
  mode: "builtin"                   # builtin, nginx, caddy, auto, config, none
  http_port: 8080                   # HTTP proxy port
  https_port: 8443                  # HTTPS proxy port
  external:
    nginx:
      binary_path: "/usr/sbin/nginx"
      config_path: "/etc/nginx/nginx.conf"
      pid_path: "/var/run/nginx.pid"
    caddy:
      binary_path: "/usr/bin/caddy"
      config_path: "/etc/caddy/Caddyfile"
      pid_path: "/var/run/caddy.pid"

# Logging configuration
logging:
  level: "debug"                    # debug, info, warn, error
  format: "json"                    # json, text
  file: "gotunnel.log"              # Log file path (empty = stdout)
  rotate: true                      # Enable log rotation
  max_size: "100MB"                 # Max file size before rotation
  max_files: 3                      # Number of old log files to keep

# Observability configuration
observability:
  sentry:
    dsn: "${SENTRY_DSN}"            # Sentry DSN from environment
    environment: "development"         # Sentry environment tag
    sample_rate: 1.0                # Error sampling rate (0.0-1.0)
    enable_performance_monitoring: true
  prometheus:
    enabled: true                    # Enable Prometheus metrics
    port: 9090                      # Metrics port
    path: "/metrics"                 # Metrics endpoint
  opentelemetry:
    enabled: false                   # Enable OpenTelemetry tracing
    service_name: "gotunnel"
    service_version: "1.0.0"
    endpoint: ""                     # OTLP endpoint
    sample_rate: 1.0                # Trace sampling rate

# Predefined tunnels (optional)
tunnels:
  - domain: "web.local"
    backend: "http://localhost:3000"
    https: true
    http_port: 8080
    https_port: 8443
    health_check:
      path: "/health"
      interval: "30s"
      timeout: "5s"
      failure_threshold: 3
    labels:
      environment: "dev"
      service: "frontend"
  
  - domain: "api.local"
    backend: "http://localhost:8080"
    https: true
    labels:
      environment: "dev"
      service: "backend"
```

#### Quick Configuration Examples

**Development Setup:**
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

**Production Setup:**
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
    port: 9090
```

**Corporate Environment:**
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
  file: "gotunnel.log"
```

📖 **For complete configuration reference, see [Configuration Guide](docs/CONFIGURATION.md)**

## 📊 Observability

### Metrics

gotunnel exposes Prometheus metrics at `:9090/metrics`:

- `gotunnel_tunnels_active` - Number of active tunnels
- `gotunnel_requests_total` - Total HTTP requests processed
- `gotunnel_request_duration_seconds` - Request processing time
- `gotunnel_errors_total` - Total errors by type

### Tracing

Distributed tracing via OpenTelemetry:

```bash
# With OTLP endpoint
gotunnel --debug start --port 3000 --domain myapp
```

### Monitoring Stack

```bash
# Start with Prometheus + Grafana
docker-compose --profile monitoring up -d

# Access dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
```

## 📚 CLI Reference

### Global Flags

```bash
gotunnel [global options] command [command options] [arguments...]

GLOBAL OPTIONS:
   --no-privilege-check         Skip privilege check
   --sentry-dsn value           Sentry DSN for error tracking [$SENTRY_DSN]
   --environment value          Environment (development, staging, production) [$ENVIRONMENT]
   --debug                      Enable debug logging and tracing [$DEBUG]
   --proxy value                Proxy mode: builtin, nginx, caddy, auto, config, none [$GOTUNNEL_PROXY]
   --proxy-http-port value      HTTP port for proxy (default: 80) [$GOTUNNEL_PROXY_HTTP_PORT]
   --proxy-https-port value     HTTPS port for proxy (default: 443) [$GOTUNNEL_PROXY_HTTPS_PORT]
```

### Commands

```bash
gotunnel start --port 3000 --domain myapp    # Start tunnel
gotunnel stop myapp                           # Stop specific tunnel  
gotunnel list                                 # List active tunnels
gotunnel stop-all                            # Stop all tunnels
```

## 🛠️ Troubleshooting

### Common Issues

**❌ "insufficient privileges: cannot bind to port 80"**
```
💡 Solutions:
   • Run with sudo: sudo gotunnel ...
   • Use --no-privilege-check to skip privilege checks (limited functionality)
   • Configure proxy mode for non-root usage: gotunnel --proxy=builtin --no-privilege-check
   • Use custom ports: gotunnel --proxy-http-port 8080 --proxy-https-port 8443
```

**❌ "TUNNEL_START: Failed to start tunnel (caused by: failed to update hosts file)"**
```
💡 Solutions:
   • Run with elevated privileges: sudo gotunnel ...
   • Use proxy mode: gotunnel --proxy=builtin --no-privilege-check
   • Check file permissions: ls -la /etc/hosts
   • Use configuration file instead of CLI flags
```

**❌ "unsupported proxy mode: invalid"**
```
💡 Valid proxy modes: builtin, nginx, caddy, auto, config, none
   • Use --proxy=builtin for built-in proxy (recommended)
   • Use --proxy=auto for automatic detection
   • Check configuration file syntax
```

**❌ "mkcert is not available. Falling back to HTTP"**
```
💡 HTTPS requires mkcert for certificate generation:
   • macOS: brew install mkcert && mkcert -install
   • Linux: Follow https://github.com/FiloSottile/mkcert#linux
   • Windows: choco install mkcert
   • Or use HTTP only: --https=false
```

**❌ "Domain not accessible"**
```bash
# Check /etc/hosts (requires root)
cat /etc/hosts | grep myapp.local

# Check DNS resolution
dig myapp.local
nslookup myapp.local

# Check proxy routes (if using proxy mode)
curl http://localhost:8080  # Default proxy port
```

**❌ "Configuration file not found"**
```bash
# Check standard locations:
ls -la gotunnel.yaml gotunnel.yml
ls -la config/gotunnel.yaml
ls -la ~/.gotunnel/gotunnel.yaml
ls -la /etc/gotunnel/gotunnel.yaml

# Validate YAML syntax:
python3 -c "import yaml; yaml.safe_load(open('gotunnel.yaml'))"
```

**❌ "Connection refused" or "No route to host"**
```bash
# Check if backend service is running
curl http://localhost:3000  # Or your backend port

# Check firewall settings
sudo ufw status  # Linux
sudo pfctl -s rules  # macOS

# Verify tunnel is active
gotunnel list
```

### Configuration Troubleshooting

**Invalid configuration file:**
```bash
# Test configuration loading
gotunnel --debug --no-privilege-check list 2>&1 | head -10

# Validate specific config file
go run -c 'package main; import "github.com/johncferguson/gotunnel/internal/config"; func main() { cfg, err := config.LoadFromFile("gotunnel.yaml"); if err != nil { panic(err) }; println("Config loaded successfully") }'
```

**Environment variables not working:**
```bash
# Check environment variables
env | grep GOTUNNEL_

# Test with explicit values
GOTUNNEL_LOGGING_LEVEL=debug GOTUNNEL_PROXY=builtin gotunnel --no-privilege-check list
```

### Debug Mode & Observability

```bash
# Enable debug logging
gotunnel --debug start --port 3000 --domain myapp

# Check metrics (if enabled)
curl http://localhost:9090/metrics

# View structured logs
tail -f gotunnel.log  # If file logging is configured

# Test connectivity
curl -v http://myapp.local  # Should route through tunnel
curl -v https://myapp.local  # HTTPS (if certificates available)
```

### Network Diagnostics

**Check tunnel routing:**
```bash
# Test direct tunnel access
curl http://localhost:9080  # Default tunnel port in proxy mode

# Test DNS resolution
nslookup myapp.local

# Check active connections
netstat -tlnp | grep :80
netstat -tlnp | grep :443
```

**Firewall issues:**
```bash
# Linux (ufw)
sudo ufw status
sudo ufw allow 80
sudo ufw allow 443

# macOS (pf)
sudo pfctl -s rules

# Windows (firewall)
netsh advfirewall firewall show rule name=all
```

### Platform-Specific Issues

**macOS:**
```bash
# Certificate trust issues
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ~/.gotunnel/certs/rootCA.pem

# DNS resolution issues
sudo killall -HUP mDNSResponder
```

**Windows:**
```bash
# Run as Administrator
# Right-click > Run as administrator

# Certificate installation
certlm.msc  # Manage computer certificates
# Import rootCA.pem to Trusted Root Certification Authorities
```

**Linux:**
```bash
# SELinux issues
sudo setsebool -P httpd_can_network_connect 1

# AppArmor issues
sudo apparmor_status
sudo aa-complain /usr/bin/gotunnel  # If confined
```

### Advanced Configuration

**Multiple tunnels:**
```yaml
# gotunnel.yaml
proxy:
  mode: "builtin"
  http_port: 80
  https_port: 443

tunnels:
  - domain: "frontend"
    backend: "http://localhost:3000"
    https: true
  - domain: "api"
    backend: "http://localhost:8080"
    https: true
  - domain: "admin"
    backend: "http://localhost:4000"
    https: false
```

**Enterprise setup:**
```yaml
# gotunnel.yaml
global:
  environment: "production"
  no_privilege_check: true

proxy:
  mode: "builtin"
  http_port: 8080
  https_port: 8443

logging:
  level: "info"
  format: "json"
  file: "/var/log/gotunnel/app.log"

observability:
  sentry:
    dsn: "${SENTRY_DSN}"
    environment: "production"
```

### System Service

**systemd (Linux):**
```bash
# Install as system service
sudo ./scripts/install.sh --service

# Control service
sudo systemctl start gotunnel
sudo systemctl enable gotunnel
sudo journalctl -u gotunnel -f
```

**launchd (macOS):**
```bash
# Install via Homebrew (includes service)
brew services start gotunnel
brew services stop gotunnel
```

**Windows Service:**
```powershell
# Install as Windows service (requires NSSM or similar)
nssm install gotunnel "C:\path\to\gotunnel.exe"
nssm set gotunnel AppParameters "start --config C:\path\to\config.yaml"
nssm start gotunnel
```

## 🧪 Development

### Building from Source

```bash
git clone https://github.com/johncferguson/gotunnel.git
cd gotunnel
go mod tidy
go build ./cmd/gotunnel
```

### Running Tests

```bash
go test ./...                    # All tests
go test ./internal/tunnel -v     # Specific package
go test -race ./...              # Race detection
```

### Quality Checks

```bash
golangci-lint run               # Full linting
go fmt ./...                    # Format code
go vet ./...                    # Static analysis
```

### CI/CD Pipeline

The project uses GitHub Actions for:
- ✅ Multi-platform testing (Go 1.22, 1.23)
- ✅ Comprehensive linting with golangci-lint
- ✅ Security scanning with gosec and Trivy
- ✅ Multi-architecture builds (Linux, macOS, Windows - amd64/arm64)
- ✅ Docker image building and publishing
- ✅ Automated releases and package distribution
- ✅ Code signing for macOS binaries

## 🔒 Security

### Reporting Vulnerabilities

Please report security vulnerabilities via [GitHub Security Advisories](https://github.com/johncferguson/gotunnel/security/advisories/new).

### Security Features

- **Code Signing**: macOS binaries are signed with Apple Developer ID and notarized
- **No Root Required**: Core functionality works without administrator privileges
- **Automatic Certificate Management**: Self-signed certificates for HTTPS tunnels
- **Host File Safety**: Automatic backup and restoration of system hosts file
- **Network Isolation**: Docker containers with proper security boundaries
- **Security Scanning**: Automated vulnerability scanning in CI/CD pipeline
- **Encrypted Secrets**: All sensitive data handled through encrypted GitHub secrets

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation for user-facing changes
- Use conventional commits for clear history
- Ensure all CI checks pass

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📚 [Documentation](https://gotunnel.dev)
- 🐛 [Issue Tracker](https://github.com/johncferguson/gotunnel/issues)
- 💬 [Discussions](https://github.com/johncferguson/gotunnel/discussions)
- 🔐 [Security](https://github.com/johncferguson/gotunnel/security)

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/) and love ❤️
- Observability powered by [OpenTelemetry](https://opentelemetry.io/)
- Error tracking via [Sentry](https://sentry.io/)
- CLI framework by [urfave/cli](https://github.com/urfave/cli)

---

**gotunnel** - Making local development tunnels simple and secure 🚇