# AGENTS.md - Core Tunneling Logic

## Overview
Core tunneling functionality with HTTP/HTTPS proxy setup, hosts file modification, and tunnel lifecycle management.

## Key Functions
```go
// Core operations
manager.StartTunnel(port, domain, https)  // Create new tunnel
manager.StopTunnel(domain)                // Stop specific tunnel
manager.ListTunnels()                     // Get active tunnels
manager.StopAllTunnels()                  // Stop all tunnels

// Testing
go test ./internal/tunnel -v               // Run tunnel tests
go test ./internal/tunnel -run TestStart  // Run specific test
```

## Code Patterns
- **Platform-Specific**: Separate `tunnel_unix.go` and `tunnel_windows.go` files
- **Hosts File**: Modifies `/etc/hosts` for domain resolution (requires privileges)
- **Certificate Integration**: Works with cert package for TLS certificates
- **DNS Integration**: Registers with dnsserver for mDNS discovery

## Dependencies
- cert: Certificate management
- dnsserver: mDNS service discovery
- proxy: HTTP/HTTPS proxy handling

## Testing
- Comprehensive test coverage in `tunnel_test.go`
- Platform-specific tests in respective `*_unix.go` and `*_windows.go` files
- Tests require filesystem permissions for hosts file modifications