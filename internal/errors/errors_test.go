package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGotunnelError(t *testing.T) {
	tests := []struct {
		name     string
		err      *GotunnelError
		expected string
	}{
		{
			name: "error without cause",
			err: New(ErrCodeTunnelCreate, "Failed to create tunnel"),
			expected: "TUNNEL_CREATE: Failed to create tunnel",
		},
		{
			name: "error with cause",
			err: New(ErrCodeTunnelCreate, "Failed to create tunnel").
				WithCause(errors.New("port already in use")),
			expected: "TUNNEL_CREATE: Failed to create tunnel (caused by: port already in use)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestGotunnelError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := New(ErrCodeTunnelCreate, "Failed").WithCause(cause)
	
	assert.Equal(t, cause, err.Unwrap())
}

func TestGotunnelError_Is(t *testing.T) {
	err1 := New(ErrCodeTunnelCreate, "Failed")
	err2 := New(ErrCodeTunnelCreate, "Also failed")
	err3 := New(ErrCodeTunnelStart, "Failed")
	
	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
}

func TestGotunnelError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      *GotunnelError
		expected int
	}{
		{
			name:     "not found",
			err:      New(ErrCodeTunnelNotFound, "Not found"),
			expected: 404,
		},
		{
			name:     "conflict",
			err:      New(ErrCodePortUnavailable, "Port in use"),
			expected: 409,
		},
		{
			name:     "bad request",
			err:      New(ErrCodeValidationDomain, "Invalid domain"),
			expected: 400,
		},
		{
			name:     "forbidden",
			err:      New(ErrCodePrivilegeRequired, "Need admin"),
			expected: 403,
		},
		{
			name:     "timeout",
			err:      New(ErrCodeNetworkTimeout, "Timeout"),
			expected: 408,
		},
		{
			name:     "bad gateway",
			err:      New(ErrCodeNetworkConnect, "Connection failed"),
			expected: 502,
		},
		{
			name:     "default internal error",
			err:      New(ErrCodeTunnelCreate, "Generic error"),
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.HTTPStatus())
		})
	}
}

func TestGotunnelError_WithMethods(t *testing.T) {
	cause := errors.New("underlying")
	
	err := New(ErrCodeTunnelCreate, "Base error").
		WithCause(cause).
		WithOperation("create_tunnel").
		WithContext("domain", "test.local").
		WithContext("port", 8080).
		WithHelp("Check port availability").
		WithCaller(0)

	assert.Equal(t, cause, err.Cause)
	assert.Equal(t, "create_tunnel", err.Operation)
	assert.Equal(t, "test.local", err.Context["domain"])
	assert.Equal(t, 8080, err.Context["port"])
	assert.Equal(t, "Check port availability", err.Help)
	assert.NotEmpty(t, err.File)
	assert.Greater(t, err.Line, 0)
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("TunnelCreateError", func(t *testing.T) {
		cause := errors.New("port in use")
		err := TunnelCreateError("test.local", 8080, cause)
		
		assert.Equal(t, ErrCodeTunnelCreate, err.Code)
		assert.Equal(t, "Failed to create tunnel", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "test.local", err.Context["domain"])
		assert.Equal(t, 8080, err.Context["port"])
		assert.NotEmpty(t, err.Help)
	})

	t.Run("TunnelStartError", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := TunnelStartError("test.local", cause)
		
		assert.Equal(t, ErrCodeTunnelStart, err.Code)
		assert.Equal(t, "Failed to start tunnel", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "test.local", err.Context["domain"])
		assert.NotEmpty(t, err.Help)
	})

	t.Run("CertificateError", func(t *testing.T) {
		cause := errors.New("mkcert not found")
		err := CertificateError("generate", "test.local", cause)
		
		assert.Equal(t, ErrCodeCertGenerate, err.Code)
		assert.Contains(t, err.Message, "generate certificate")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "test.local", err.Context["domain"])
		assert.Equal(t, "generate", err.Context["operation"])
		assert.NotEmpty(t, err.Help)
	})

	t.Run("MkcertMissingError", func(t *testing.T) {
		err := MkcertMissingError()
		
		assert.Equal(t, ErrCodeMkcertMissing, err.Code)
		assert.Contains(t, err.Message, "mkcert is not available")
		assert.NotEmpty(t, err.Help)
	})

	t.Run("HostsFileError", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := HostsFileError("update", cause)
		
		assert.Equal(t, ErrCodeHostsUpdate, err.Code)
		assert.Contains(t, err.Message, "update hosts file")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "update", err.Context["operation"])
		assert.NotEmpty(t, err.Help)
	})

	t.Run("PrivilegeError", func(t *testing.T) {
		cause := errors.New("not root")
		err := PrivilegeError("bind to port 80", cause)
		
		assert.Equal(t, ErrCodePrivilegeRequired, err.Code)
		assert.Contains(t, err.Message, "Elevated privileges required")
		assert.Equal(t, cause, err.Cause)
		assert.NotEmpty(t, err.Help)
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := ValidationError("domain", "invalid..domain", "invalid format")
		
		assert.Equal(t, ErrorCode("VALIDATION_DOMAIN"), err.Code)
		assert.Contains(t, err.Message, "Invalid domain")
		assert.Equal(t, "domain", err.Context["field"])
		assert.Equal(t, "invalid..domain", err.Context["value"])
		assert.NotEmpty(t, err.Help)
	})

	t.Run("NetworkError", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := NetworkError("connect", "localhost:8080", cause)
		
		assert.Equal(t, ErrCodeNetworkConnect, err.Code)
		assert.Contains(t, err.Message, "Network connect failed")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "connect", err.Context["operation"])
		assert.Equal(t, "localhost:8080", err.Context["target"])
		assert.NotEmpty(t, err.Help)
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("IsGotunnelError", func(t *testing.T) {
		gotunnelErr := New(ErrCodeTunnelCreate, "Test")
		standardErr := errors.New("standard error")
		
		_, ok := IsGotunnelError(gotunnelErr)
		assert.True(t, ok)
		
		_, ok = IsGotunnelError(standardErr)
		assert.False(t, ok)
		
		_, ok = IsGotunnelError(nil)
		assert.False(t, ok)
	})

	t.Run("GetErrorCode", func(t *testing.T) {
		gotunnelErr := New(ErrCodeTunnelCreate, "Test")
		standardErr := errors.New("standard error")
		
		assert.Equal(t, ErrCodeTunnelCreate, GetErrorCode(gotunnelErr))
		assert.Equal(t, ErrorCode(""), GetErrorCode(standardErr))
		assert.Equal(t, ErrorCode(""), GetErrorCode(nil))
	})

	t.Run("GetHelpText", func(t *testing.T) {
		errWithHelp := New(ErrCodeTunnelCreate, "Test").WithHelp("Help text")
		errWithoutHelp := New(ErrCodeTunnelCreate, "Test")
		standardErr := errors.New("standard error")
		
		assert.Equal(t, "Help text", GetHelpText(errWithHelp))
		assert.Equal(t, "", GetHelpText(errWithoutHelp))
		assert.Equal(t, "", GetHelpText(standardErr))
	})

	t.Run("Wrap", func(t *testing.T) {
		cause := errors.New("underlying")
		err := Wrap(cause, ErrCodeTunnelCreate, "Wrapped error")
		
		assert.Equal(t, ErrCodeTunnelCreate, err.Code)
		assert.Equal(t, "Wrapped error", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("WrapOperation", func(t *testing.T) {
		cause := errors.New("underlying")
		err := WrapOperation(cause, "test_operation")
		
		assert.Equal(t, ErrorCode("OPERATION_FAILED"), err.Code)
		assert.Contains(t, err.Message, "test_operation")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "test_operation", err.Operation)
	})
}

func TestErrorContext(t *testing.T) {
	err := New(ErrCodeTunnelCreate, "Test").
		WithContext("string", "value").
		WithContext("number", 42).
		WithContext("bool", true).
		WithContext("slice", []string{"a", "b"})

	assert.Equal(t, "value", err.Context["string"])
	assert.Equal(t, 42, err.Context["number"])
	assert.Equal(t, true, err.Context["bool"])
	assert.Equal(t, []string{"a", "b"}, err.Context["slice"])
}

func TestErrorChaining(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, ErrCodeTunnelCreate, "wrapped")
	doubleWrapped := Wrap(wrapped, ErrCodeTunnelStart, "double wrapped")

	assert.Equal(t, original, errors.Unwrap(wrapped))
	assert.Equal(t, wrapped, errors.Unwrap(doubleWrapped))
	assert.True(t, errors.Is(doubleWrapped, wrapped))
	assert.True(t, errors.Is(doubleWrapped, original))
}