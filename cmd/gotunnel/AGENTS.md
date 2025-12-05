# AGENTS.md - CLI Application

## Overview
Main CLI entry point for gotunnel using urfave/cli/v2 framework. Handles command parsing, configuration, and orchestrates all internal components.

## Key Commands
```bash
# Core tunnel operations
go run ./cmd/gotunnel start --port 3000 --domain myapp  # Start HTTPS tunnel
go run ./cmd/gotunnel stop myapp                        # Stop specific tunnel
go run ./cmd/gotunnel list                              # List active tunnels
go run ./cmd/gotunnel stop-all                          # Stop all tunnels

# Testing
go test ./cmd/gotunnel -v                                # Run integration tests
go test ./cmd/gotunnel -run TestStartTunnel             # Run specific test
```

## Code Patterns
- **CLI Framework**: Uses urfave/cli/v2 for command structure
- **Global State**: Manager instances stored in global variables
- **Graceful Shutdown**: Signal handling with context cancellation
- **Error Handling**: fmt.Errorf with context wrapping

## Integration Points
- Imports all internal packages (cert, dnsserver, logging, observability, privilege, proxy, tunnel)
- Initializes observability provider and metrics
- Manages proxy server lifecycle

## Testing
- Integration tests in `main_test.go`
- Tests full CLI command execution
- Requires network access for some tests