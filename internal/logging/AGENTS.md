# AGENTS.md - Logging Utilities

## Overview
Structured logging utilities using slog with OpenTelemetry trace integration.

## Key Functions
```go
logger := logging.New(config)              // Create logger instance
logger.Info("message", "key", "value")     // Log info message
logger.Error("error occurred", "err", err) // Log error with context

// Testing
go test ./internal/logging -v               // Run logging tests
go test ./internal/logging -run TestNew    // Run specific test
```

## Code Patterns
- **Structured Logging**: Uses slog for structured log output
- **Configurable**: Level, format, output destination
- **Trace Integration**: Includes OpenTelemetry trace IDs
- **Multiple Formats**: Text and JSON output formats

## Dependencies
- `go.opentelemetry.io/otel/trace` - OpenTelemetry tracing

## Testing
- Tests in `logger_test.go`
- Tests logger creation and output formatting