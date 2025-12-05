# AGENTS.md - Tunnel State Management

## Overview
Persistent tunnel state management using YAML configuration files.

## Key Functions
```go
state.SaveTunnelState(domain, state)       // Save tunnel configuration
state.LoadTunnelStates()                   // Load all saved states
state.RemoveTunnelState(domain)            // Remove tunnel state

// Testing
go test ./internal/state -v                 // Run state management tests
go test ./internal/state -run TestSave     // Run specific test
```

## Code Patterns
- **YAML Storage**: Uses gopkg.in/yaml.v3 for serialization
- **File-Based**: Stores state in local files
- **State Struct**: TunnelState with Port, Domain, HTTPS fields

## Dependencies
- `gopkg.in/yaml.v3` - YAML parsing and serialization

## Testing
- Tests in `state_test.go`
- Tests state save/load/remove operations