# AGENTS.md - Certificate Management

## Overview
Platform-specific certificate handling with mkcert integration for local CA certificates.

## Key Functions
```go
certManager := cert.New("./certs")         // Create cert manager
cert, err := certManager.GetCert(domain)   // Get TLS certificate
certManager.InstallCA()                    // Install local CA

// Testing
go test ./internal/cert -v                  // Run certificate tests
go test ./internal/cert -run TestGetCert   // Run specific test
```

## Code Patterns
- **Platform-Specific**: Separate `cert_unix.go` and `cert_windows.go` files
- **mkcert Integration**: Uses external mkcert tool for certificate generation
- **Directory Management**: Stores certificates in `./certs` directory
- **CA Installation**: Installs local certificate authority system-wide

## Dependencies
- External: mkcert command-line tool
- Platform-specific certificate stores (Keychain on macOS, cert store on Windows)

## Testing
- Mock tests in `cert_mock_test.go`
- Platform-specific tests in respective files
- Requires mkcert installation for full functionality