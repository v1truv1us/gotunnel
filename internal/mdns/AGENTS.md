# AGENTS.md - mDNS Implementation

## Overview
mDNS service implementation using grandcat/zeroconf for local network service discovery.

## Key Functions
```go
mdnsServer := mdns.New()                   // Create mDNS server
mdnsServer.RegisterDomain(domain)          // Register domain service
mdnsServer.UnregisterDomain(domain)        // Remove domain registration

// Testing
go test ./internal/mdns -v                  // Run mDNS tests
go test ./internal/mdns -run TestRegister  // Run specific test
```

## Code Patterns
- **Service Map**: Manages map of domain -> zeroconf.Server mappings
- **Thread-Safe**: Uses RWMutex for concurrent access
- **Zeroconf Integration**: Uses grandcat/zeroconf library
- **Cleanup**: Proper service shutdown and unregistration

## Dependencies
- `github.com/grandcat/zeroconf` - Zeroconf networking

## Testing
- Tests in `mdns_test.go`
- Requires network access for mDNS functionality