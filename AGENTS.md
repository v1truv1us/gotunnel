# AGENTS.md - gotunnel Project Guide

## Project Overview
**gotunnel** is a secure local tunneling tool that creates HTTPS-enabled local domains with mDNS discovery. Built in Go with cross-platform support.

## Core Architecture
- **CLI**: `cmd/gotunnel/` - Main entry point using urfave/cli/v2
- **Internal Packages**: `internal/` - Core functionality modules
- **Documentation**: `docs-site/` - Astro.js documentation site
- **Packaging**: `packaging/` - Cross-platform distribution packages

## Development Commands
```bash
# Core Go commands
go mod tidy                    # Update dependencies
go build ./cmd/gotunnel       # Build binary
go run ./cmd/gotunnel         # Run directly
go test ./...                 # Run all tests
go test -v ./internal/tunnel  # Test specific package
go test -run TestSpecific     # Run specific test

# Cross-platform builds
./build.bat                   # Multi-platform binaries to bin/

# Quality checks
go test ./... -cover          # Coverage report
go test -race ./...           # Race detection
go vet ./...                  # Static analysis
```

## Key Dependencies
- `github.com/urfave/cli/v2` - CLI framework
- `github.com/hashicorp/mdns` - mDNS discovery
- `github.com/grandcat/zeroconf` - Zero-conf networking
- `gopkg.in/yaml.v3` - Config parsing
- `github.com/stretchr/testify` - Testing

## Code Patterns
- **Error Handling**: `fmt.Errorf` with context wrapping
- **Testing**: testify assertions, comprehensive test coverage
- **Platform Code**: Separate `*_unix.go` and `*_windows.go` files
- **Privilege**: Elevated permissions required for system modifications

## Directory Index
- [cmd/gotunnel/](cmd/gotunnel/AGENTS.md) - CLI application and integration tests
- [internal/cert/](internal/cert/AGENTS.md) - Certificate management and mkcert integration
- [internal/dnsserver/](internal/dnsserver/AGENTS.md) - mDNS service discovery
- [internal/logging/](internal/logging/AGENTS.md) - Logging utilities
- [internal/mdns/](internal/mdns/AGENTS.md) - mDNS implementation
- [internal/observability/](internal/observability/AGENTS.md) - Metrics and tracing
- [internal/privilege/](internal/privilege/AGENTS.md) - Cross-platform privilege checking
- [internal/proxy/](internal/proxy/AGENTS.md) - HTTP/HTTPS proxy handling
- [internal/state/](internal/state/AGENTS.md) - Tunnel state management
- [internal/tunnel/](internal/tunnel/AGENTS.md) - Core tunneling logic
- [docs-site/](docs-site/AGENTS.md) - Astro.js documentation site
- [packaging/](packaging/AGENTS.md) - Distribution packaging

## Security Notes
- Requires elevated privileges for ports 80/443 and hosts file modifications
- Local CA certificates stored in `./certs`
- Safe hosts file backup/restore capabilities
- Network isolation through local-only certificates

## Testing Guidelines
- All packages have `*_test.go` files
- Integration tests in `cmd/gotunnel/main_test.go`
- Some tests require network access and filesystem permissions
- Use `go test -v` for verbose output, `-run TestName` for specific tests