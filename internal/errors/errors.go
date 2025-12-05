package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// ErrorCode represents different types of errors that can occur
type ErrorCode string

const (
	// Tunnel errors
	ErrCodeTunnelCreate    ErrorCode = "TUNNEL_CREATE"
	ErrCodeTunnelStart     ErrorCode = "TUNNEL_START"
	ErrCodeTunnelStop      ErrorCode = "TUNNEL_STOP"
	ErrCodeTunnelNotFound  ErrorCode = "TUNNEL_NOT_FOUND"
	ErrCodePortUnavailable ErrorCode = "PORT_UNAVAILABLE"

	// Certificate errors
	ErrCodeCertGenerate    ErrorCode = "CERT_GENERATE"
	ErrCodeCertInstall     ErrorCode = "CERT_INSTALL"
	ErrCodeCertLoad        ErrorCode = "CERT_LOAD"
	ErrCodeMkcertMissing   ErrorCode = "MKCERT_MISSING"

	// DNS/Hosts errors
	ErrCodeDNSRegister     ErrorCode = "DNS_REGISTER"
	ErrCodeDNSUnregister   ErrorCode = "DNS_UNREGISTER"
	ErrCodeHostsUpdate     ErrorCode = "HOSTS_UPDATE"
	ErrCodeHostsPermission ErrorCode = "HOSTS_PERMISSION"

	// Privilege errors
	ErrCodePrivilegeRequired ErrorCode = "PRIVILEGE_REQUIRED"
	ErrCodePrivilegeEscalate ErrorCode = "PRIVILEGE_ESCALATE"

	// Configuration errors
	ErrCodeConfigLoad    ErrorCode = "CONFIG_LOAD"
	ErrCodeConfigParse   ErrorCode = "CONFIG_PARSE"
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// Network errors
	ErrCodeNetworkConnect ErrorCode = "NETWORK_CONNECT"
	ErrCodeNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"
	ErrCodeNetworkResolve ErrorCode = "NETWORK_RESOLVE"

	// System errors
	ErrCodeFilesystem    ErrorCode = "FILESYSTEM"
	ErrCodePermission    ErrorCode = "PERMISSION"
	ErrCodeCommandFailed ErrorCode = "COMMAND_FAILED"

	// Validation errors
	ErrCodeValidationDomain ErrorCode = "VALIDATION_DOMAIN"
	ErrCodeValidationPort   ErrorCode = "VALIDATION_PORT"
	ErrCodeValidationURL    ErrorCode = "VALIDATION_URL"
)

// GotunnelError represents a structured error with context and helpful guidance
type GotunnelError struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Operation string                 `json:"operation,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
	Help      string                 `json:"help,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// Error implements the error interface
func (e *GotunnelError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *GotunnelError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *GotunnelError) Is(target error) bool {
	if t, ok := target.(*GotunnelError); ok {
		return e.Code == t.Code
	}
	return false
}

// HTTPStatus returns the appropriate HTTP status code for this error
func (e *GotunnelError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeTunnelNotFound:
		return http.StatusNotFound
	case ErrCodePortUnavailable:
		return http.StatusConflict
	case ErrCodeValidationDomain, ErrCodeValidationPort, ErrCodeValidationURL, ErrCodeConfigInvalid:
		return http.StatusBadRequest
	case ErrCodePrivilegeRequired, ErrCodeHostsPermission, ErrCodePermission:
		return http.StatusForbidden
	case ErrCodeNetworkTimeout:
		return http.StatusRequestTimeout
	case ErrCodeNetworkConnect, ErrCodeNetworkResolve:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new GotunnelError with the given parameters
func New(code ErrorCode, message string) *GotunnelError {
	return &GotunnelError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WithCause adds an underlying cause to the error
func (e *GotunnelError) WithCause(cause error) *GotunnelError {
	e.Cause = cause
	return e
}

// WithOperation adds the operation context to the error
func (e *GotunnelError) WithOperation(operation string) *GotunnelError {
	e.Operation = operation
	return e
}

// WithContext adds key-value context to the error
func (e *GotunnelError) WithContext(key string, value interface{}) *GotunnelError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithHelp adds helpful guidance for resolving the error
func (e *GotunnelError) WithHelp(help string) *GotunnelError {
	e.Help = help
	return e
}

// WithCaller adds file and line information
func (e *GotunnelError) WithCaller(skip int) *GotunnelError {
	_, file, line, ok := runtime.Caller(skip + 1)
	if ok {
		e.File = file
		e.Line = line
	}
	return e
}

// Convenience functions for common error types

// TunnelCreateError creates an error for tunnel creation failures
func TunnelCreateError(domain string, port int, cause error) *GotunnelError {
	return New(ErrCodeTunnelCreate, "Failed to create tunnel").
		WithCause(cause).
		WithContext("domain", domain).
		WithContext("port", port).
		WithHelp("Check if the port is available and mkcert is installed for HTTPS").
		WithCaller(1)
}

// TunnelStartError creates an error for tunnel start failures
func TunnelStartError(domain string, cause error) *GotunnelError {
	return New(ErrCodeTunnelStart, "Failed to start tunnel").
		WithCause(cause).
		WithContext("domain", domain).
		WithHelp("Check if ports 80/443 are available and you have necessary privileges").
		WithCaller(1)
}

// CertificateError creates an error for certificate-related failures
func CertificateError(operation string, domain string, cause error) *GotunnelError {
	var code ErrorCode
	var help string

	switch operation {
	case "generate":
		code = ErrCodeCertGenerate
		help = "Install mkcert: https://github.com/FiloSottile/mkcert#installation"
	case "install":
		code = ErrCodeCertInstall
		help = "Run 'mkcert -install' to install the local CA"
	case "load":
		code = ErrCodeCertLoad
		help = "Check if certificate files exist and are readable"
	default:
		code = ErrCodeCertLoad
		help = "Check certificate configuration and permissions"
	}

	return New(code, fmt.Sprintf("Failed to %s certificate", operation)).
		WithCause(cause).
		WithContext("domain", domain).
		WithContext("operation", operation).
		WithHelp(help).
		WithCaller(1)
}

// MkcertMissingError creates an error when mkcert is not available
func MkcertMissingError() *GotunnelError {
	return New(ErrCodeMkcertMissing, "mkcert is not available for HTTPS certificate generation").
		WithHelp("Install mkcert for HTTPS support: https://github.com/FiloSottile/mkcert#installation\nAlternatively, use --https=false to use HTTP only").
		WithCaller(1)
}

// HostsFileError creates an error for hosts file operations
func HostsFileError(operation string, cause error) *GotunnelError {
	var help string
	if operation == "update" || operation == "write" {
		help = "Try running with elevated privileges or check file permissions"
	} else {
		help = "Check if the hosts file exists and is readable"
	}

	return New(ErrCodeHostsUpdate, fmt.Sprintf("Failed to %s hosts file", operation)).
		WithCause(cause).
		WithContext("operation", operation).
		WithHelp(help).
		WithCaller(1)
}

// PrivilegeError creates an error for privilege-related failures
func PrivilegeError(operation string, cause error) *GotunnelError {
	return New(ErrCodePrivilegeRequired, fmt.Sprintf("Elevated privileges required for %s", operation)).
		WithCause(cause).
		WithHelp("Try running with sudo/administrator privileges or use non-privileged ports (>1024)").
		WithCaller(1)
}

// ValidationError creates an error for input validation failures
func ValidationError(field string, value string, reason string) *GotunnelError {
	var help string
	switch field {
	case "domain":
		help = "Use a valid domain name (e.g., app.local, test.dev)"
	case "port":
		help = "Use a valid port number (1-65535)"
	case "url":
		help = "Use a valid URL format (e.g., http://localhost:3000)"
	default:
		help = "Check the input format and requirements"
	}

	return New(ErrorCode("VALIDATION_"+strings.ToUpper(field)), 
		fmt.Sprintf("Invalid %s: %s", field, reason)).
		WithContext("field", field).
		WithContext("value", value).
		WithHelp(help).
		WithCaller(1)
}

// NetworkError creates an error for network-related failures
func NetworkError(operation string, target string, cause error) *GotunnelError {
	var code ErrorCode
	switch operation {
	case "connect":
		code = ErrCodeNetworkConnect
	case "resolve":
		code = ErrCodeNetworkResolve
	case "timeout":
		code = ErrCodeNetworkTimeout
	default:
		code = ErrCodeNetworkConnect
	}

	return New(code, fmt.Sprintf("Network %s failed", operation)).
		WithCause(cause).
		WithContext("operation", operation).
		WithContext("target", target).
		WithHelp("Check network connectivity and DNS configuration").
		WithCaller(1)
}

// IsGotunnelError checks if an error is a GotunnelError
func IsGotunnelError(err error) (*GotunnelError, bool) {
	if err == nil {
		return nil, false
	}
	
	var gotunnelErr *GotunnelError
	if errors.As(err, &gotunnelErr) {
		return gotunnelErr, true
	}
	
	return nil, false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if gotunnelErr, ok := IsGotunnelError(err); ok {
		return gotunnelErr.Code
	}
	return ""
}

// GetHelpText extracts help text from an error
func GetHelpText(err error) string {
	if gotunnelErr, ok := IsGotunnelError(err); ok && gotunnelErr.Help != "" {
		return gotunnelErr.Help
	}
	return ""
}

// Wrap wraps a standard error with additional context
func Wrap(err error, code ErrorCode, message string) *GotunnelError {
	return New(code, message).WithCause(err).WithCaller(1)
}

// WrapOperation wraps an error with operation context
func WrapOperation(err error, operation string) *GotunnelError {
	return New(ErrorCode("OPERATION_FAILED"), fmt.Sprintf("Operation %s failed", operation)).
		WithCause(err).
		WithOperation(operation).
		WithCaller(1)
}