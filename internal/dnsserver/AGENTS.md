# AGENTS.md - mDNS Service Discovery

## Overview
mDNS service discovery using hashicorp/mdns for network-wide access to local services.

## Key Functions
```go
dnsserver.StartDNSServer()                 // Initialize DNS server
dnsserver.RegisterService(domain, ip, port) // Register mDNS service
dnsserver.UnregisterService(domain)        // Remove service registration

// Testing
go test ./internal/dnsserver -v            // Run DNS server tests
go test ./internal/dnsserver -run TestRegister // Run specific test
```

## Code Patterns
- **Global Server**: Singleton pattern for DNS server instance
- **Service Entries**: Manages map of domain -> service mappings
- **Thread-Safe**: Uses RWMutex for concurrent access
- **Cleanup**: Proper service unregistration on shutdown

## Dependencies
- `github.com/hashicorp/mdns` - mDNS implementation

## Testing
- Tests in `dnsserver_test.go`
- Requires network access for mDNS functionality