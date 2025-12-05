# AGENTS.md - Metrics and Tracing

## Overview
OpenTelemetry-based observability with metrics, tracing, and monitoring.

## Key Functions
```go
provider := observability.NewProvider()     // Create observability provider
metrics := observability.NewMetrics(provider) // Create metrics instance
metrics.RecordTunnelStart(domain)           // Record tunnel metrics

// Testing
go test ./internal/observability -v          // Run observability tests
go test ./internal/observability -run TestMetrics // Run specific test
```

## Code Patterns
- **OpenTelemetry**: Uses OTEL for metrics and tracing
- **Custom Metrics**: Tunnel count, duration, request metrics
- **Provider Pattern**: Separate provider and metrics structs
- **Attribute-Based**: Uses OTEL attributes for dimensions

## Dependencies
- `go.opentelemetry.io/otel/metric` - Metrics collection
- `go.opentelemetry.io/otel/trace` - Distributed tracing

## Testing
- Tests in `metrics_test.go` and `provider_test.go`
- Tests metric recording and provider initialization