# AGENTS.md - Cross-Platform Privilege Checking

## Overview
Cross-platform privilege checking for system modifications (currently disabled).

## Key Functions
```go
err := privilege.CheckPrivileges()          // Check if running with privileges

// Testing
go test ./internal/privilege -v              // Run privilege tests
go test ./internal/privilege -run TestCheck // Run specific test
```

## Code Patterns
- **Platform-Specific**: Separate Unix and Windows implementations
- **Currently Disabled**: Privilege checks are bypassed for development
- **Extensible**: Framework in place for future privilege enforcement

## Dependencies
- Platform-specific privilege checking APIs

## Testing
- Tests in `privilege_test.go`
- Tests privilege checking logic (currently no-ops)