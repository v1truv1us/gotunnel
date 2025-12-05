package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_RegisterProvider(t *testing.T) {
	manager := NewManager()
	
	// Test registering a valid provider
	provider := NewBasicProvider()
	err := manager.RegisterProvider(provider)
	require.NoError(t, err)
	
	// Test registering duplicate provider
	err = manager.RegisterProvider(provider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
	
	// Test registering nil provider
	err = manager.RegisterProvider(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestManager_GetProvider(t *testing.T) {
	manager := NewManager()
	provider := NewBasicProvider()
	
	// Test getting non-existent provider
	_, exists := manager.GetProvider("nonexistent")
	assert.False(t, exists)
	
	// Test getting existing provider
	err := manager.RegisterProvider(provider)
	require.NoError(t, err)
	
	retrieved, exists := manager.GetProvider("basic")
	assert.True(t, exists)
	assert.Equal(t, provider, retrieved)
}

func TestManager_ListProviders(t *testing.T) {
	manager := NewManager()
	
	// Initially empty
	providers := manager.ListProviders()
	assert.Empty(t, providers)
	
	// Add providers
	provider1 := NewBasicProvider()
	provider2 := NewJWTProvider()
	
	err := manager.RegisterProvider(provider1)
	require.NoError(t, err)
	err = manager.RegisterProvider(provider2)
	require.NoError(t, err)
	
	providers = manager.ListProviders()
	assert.Len(t, providers, 2)
}

func TestManager_Initialize(t *testing.T) {
	manager := NewManager()
	
	// Test with nil config
	err := manager.Initialize(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
	
	// Test with valid config
	config := &Config{
		Enabled: true,
		Providers: []ProviderConfig{
			{
				Type:    ProviderTypeBasic,
				Name:    "basic",
				Enabled: true,
			},
		},
	}
	
	err = manager.Initialize(config)
	assert.NoError(t, err)
}

func TestManager_Authenticate(t *testing.T) {
	manager := NewManager()
	
	// Test with disabled auth
	config := &Config{Enabled: false}
	err := manager.Initialize(config)
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/", nil)
	result, err := manager.Authenticate(context.Background(), req)
	
	assert.NoError(t, err)
	assert.True(t, result.Success)
}

func TestManager_ValidateToken(t *testing.T) {
	manager := NewManager()
	
	// Test with disabled auth
	config := &Config{Enabled: false}
	err := manager.Initialize(config)
	require.NoError(t, err)
	
	result, err := manager.ValidateToken(context.Background(), "")
	assert.NoError(t, err)
	assert.True(t, result.Success)
	
	// Test with empty token and enabled auth
	config.Enabled = true
	err = manager.Initialize(config)
	require.NoError(t, err)
	
	result, err = manager.ValidateToken(context.Background(), "")
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, ErrCodeMissingToken, result.Error)
}

func TestManager_UnregisterProvider(t *testing.T) {
	manager := NewManager()
	provider := NewBasicProvider()
	
	// Test unregistering non-existent provider
	err := manager.UnregisterProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Test unregistering existing provider
	err = manager.RegisterProvider(provider)
	require.NoError(t, err)
	
	err = manager.UnregisterProvider("basic")
	assert.NoError(t, err)
	
	// Verify provider is gone
	_, exists := manager.GetProvider("basic")
	assert.False(t, exists)
}

func TestManager_Shutdown(t *testing.T) {
	manager := NewManager()
	
	// Test shutdown with no providers
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := manager.Shutdown(ctx)
	assert.NoError(t, err)
	
	// Test shutdown with providers
	provider := NewBasicProvider()
	err = manager.RegisterProvider(provider)
	require.NoError(t, err)
	
	err = manager.Shutdown(ctx)
	assert.NoError(t, err)
	
	// Verify providers are cleared
	providers := manager.ListProviders()
	assert.Empty(t, providers)
}

func TestBasicProvider(t *testing.T) {
	provider := NewBasicProvider()
	
	assert.Equal(t, "basic", provider.Name())
	assert.Equal(t, ProviderTypeBasic, provider.Type())
	
	// Test initialization
	config := ProviderConfig{
		Type: ProviderTypeBasic,
		Name: "test",
	}
	
	err := provider.Initialize(config)
	assert.NoError(t, err)
	
	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = provider.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestJWTProvider(t *testing.T) {
	provider := NewJWTProvider()
	
	assert.Equal(t, "jwt", provider.Name())
	assert.Equal(t, ProviderTypeJWT, provider.Type())
}

func TestOAuthProvider(t *testing.T) {
	provider := NewOAuthProvider()
	
	assert.Equal(t, "oauth", provider.Name())
	assert.Equal(t, ProviderTypeOAuth, provider.Type())
}

func TestStaticProvider(t *testing.T) {
	provider := NewStaticProvider()
	
	assert.Equal(t, "static", provider.Name())
	assert.Equal(t, ProviderTypeStatic, provider.Type())
}

func TestMiddlewareOptions(t *testing.T) {
	config := DefaultMiddlewareConfig()
	
	// Test WithRequired
	WithRequired(false)(config)
	assert.False(t, config.Required)
	
	// Test WithProviders
	WithProviders("basic", "jwt")(config)
	assert.Contains(t, config.Providers, "basic")
	assert.Contains(t, config.Providers, "jwt")
	
	// Test WithBypassPaths
	WithBypassPaths("/health", "/metrics")(config)
	assert.Contains(t, config.BypassPaths, "/health")
	assert.Contains(t, config.BypassPaths, "/metrics")
	
	// Test WithTokenHeader
	WithTokenHeader("X-Auth-Token")(config)
	assert.Equal(t, "X-Auth-Token", config.TokenHeader)
	
	// Test WithSessionTimeout
	WithSessionTimeout(time.Hour * 2)(config)
	assert.Equal(t, time.Hour*2, config.SessionTimeout)
	
	// Test WithRequireHTTPS
	WithRequireHTTPS(true)(config)
	assert.True(t, config.RequireHTTPS)
	
	// Test WithRealm
	WithRealm("test-realm")(config)
	assert.Equal(t, "test-realm", config.Realm)
}

func TestErrorHandling(t *testing.T) {
	// Test NewError
	err := NewError(ErrCodeInvalidCredentials, "Invalid credentials", "User not found")
	
	assert.Equal(t, ErrCodeInvalidCredentials, err.Code)
	assert.Equal(t, "Invalid credentials", err.Message)
	assert.Equal(t, "User not found", err.Details)
	assert.Equal(t, "Invalid credentials", err.Error())
	
	// Test IsErrorCode
	authErr := NewError(ErrCodeTokenExpired, "Token expired", "")
	
	assert.True(t, IsErrorCode(authErr, ErrCodeTokenExpired))
	assert.False(t, IsErrorCode(authErr, ErrCodeInvalidCredentials))
	
	// Test IsErrorCode with non-auth error
	regularErr := assert.AnError
	assert.False(t, IsErrorCode(regularErr, ErrCodeTokenExpired))
}

func TestContextFunctions(t *testing.T) {
	// Test setAuthContext and GetAuthContext
	authCtx := &AuthContext{
		User: &User{
			ID:       "test-id",
			Username: "testuser",
			Email:    "test@example.com",
		},
		Method: "jwt",
	}
	
	ctx := setAuthContext(context.Background(), authCtx)
	
	// Create a request with the context
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx)
	
	// Test GetAuthContext
	retrievedCtx, ok := GetAuthContext(req)
	assert.True(t, ok)
	assert.Equal(t, authCtx.User.ID, retrievedCtx.User.ID)
	assert.Equal(t, authCtx.Method, retrievedCtx.Method)
	
	// Test GetUser
	user, ok := GetUser(req)
	assert.True(t, ok)
	assert.Equal(t, "test-id", user.ID)
	assert.Equal(t, "testuser", user.Username)
	
	// Test IsAuthenticated
	assert.True(t, IsAuthenticated(req))
	
	// Test HasRole
	assert.False(t, HasRole(req, "admin"))
	
	// Add role and test again
	authCtx.User.Roles = []string{"user", "admin"}
	ctx = setAuthContext(context.Background(), authCtx)
	req = req.WithContext(ctx)
	
	assert.True(t, HasRole(req, "admin"))
	assert.True(t, HasRole(req, "user"))
	assert.False(t, HasRole(req, "nonexistent"))
}

func TestExtractToken(t *testing.T) {
	config := DefaultMiddlewareConfig()
	
	// Test Bearer token in Authorization header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	token := extractToken(req, config)
	assert.Equal(t, "test-token", token)
	
	// Test Basic auth in Authorization header
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0") // "test:test"
	
	token = extractToken(req, config)
	assert.Equal(t, "Basic dGVzdDp0ZXN0", token)
	
	// Test token in query parameter
	req = httptest.NewRequest("GET", "/?token=test-token", nil)
	
	token = extractToken(req, config)
	assert.Equal(t, "test-token", token)
	
	// Test token in cookie
	req = httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: config.CookieName, Value: "test-token"})
	
	token = extractToken(req, config)
	assert.Equal(t, "test-token", token)
	
	// Test no token
	req = httptest.NewRequest("GET", "/", nil)
	
	token = extractToken(req, config)
	assert.Empty(t, token)
}

func TestDetermineAuthMethod(t *testing.T) {
	// Test Bearer token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	method := determineAuthMethod(req, "test-token")
	assert.Equal(t, "jwt", method)
	
	// Test Basic auth
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
	
	method = determineAuthMethod(req, "")
	assert.Equal(t, "basic", method)
	
	// Test token only
	req = httptest.NewRequest("GET", "/", nil)
	
	method = determineAuthMethod(req, "test-token")
	assert.Equal(t, "token", method)
	
	// Test unknown
	method = determineAuthMethod(req, "")
	assert.Equal(t, "unknown", method)
}