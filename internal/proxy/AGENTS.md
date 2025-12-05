# AGENTS.md - HTTP/HTTPS Proxy Handling

## Overview
HTTP/HTTPS proxy handling with support for multiple proxy modes (builtin, nginx, caddy, auto).

## Key Functions
```go
proxyManager := proxy.NewManager(mode)     // Create proxy manager
proxyManager.Start(port)                   // Start proxy server
proxyManager.Stop()                        // Stop proxy server

// Testing
go test ./internal/proxy -v                 // Run proxy tests
go test ./internal/proxy -run TestBuiltIn  // Run specific test
```

## Code Patterns
- **Multiple Modes**: Support for builtin, nginx, caddy, auto-detection
- **Reverse Proxy**: Uses httputil.ReverseProxy for request forwarding
- **Configuration Generation**: Can generate nginx/caddy config files
- **Privilege Integration**: Works with privilege package for port binding

## Dependencies
- privilege: For checking port binding permissions

## Testing
- Tests in `proxy_test.go`
- Tests different proxy modes and configurations