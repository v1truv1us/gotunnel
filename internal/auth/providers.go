package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// manager implements the Manager interface
type manager struct {
	providers map[string]Provider
	config    *Config
	mu        sync.RWMutex
	
	// Middleware configuration
	middlewareConfig *MiddlewareConfig
}

// NewManager creates a new authentication manager
func NewManager() Manager {
	return &manager{
		providers:      make(map[string]Provider),
		middlewareConfig: DefaultMiddlewareConfig(),
	}
}

// RegisterProvider registers an authentication provider
func (m *manager) RegisterProvider(provider Provider) error {
	if provider == nil {
		return NewError(ErrCodeInvalidProvider, "provider cannot be nil", "")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	name := provider.Name()
	if _, exists := m.providers[name]; exists {
		return NewError(ErrCodeInvalidProvider, 
			fmt.Sprintf("provider '%s' already registered", name), "")
	}
	
	m.providers[name] = provider
	return nil
}

// UnregisterProvider unregisters an authentication provider
func (m *manager) UnregisterProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.providers[name]; !exists {
		return NewError(ErrCodeProviderNotFound, 
			fmt.Sprintf("provider '%s' not found", name), "")
	}
	
	// Shutdown provider before removing
	if provider, exists := m.providers[name]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = provider.Shutdown(ctx) // Best effort, ignore error for unregistration
	}
	
	delete(m.providers, name)
	return nil
}

// GetProvider returns a provider by name
func (m *manager) GetProvider(name string) (Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	provider, exists := m.providers[name]
	return provider, exists
}

// ListProviders returns all registered providers
func (m *manager) ListProviders() []Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	providers := make([]Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	
	return providers
}

// Authenticate authenticates a request using all registered providers
func (m *manager) Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error) {
	if !m.config.Enabled {
		return &AuthResult{Success: true}, nil
	}
	
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	m.mu.RUnlock()
	
	// Try each provider in order
	for _, provider := range providers {
		result, err := provider.Authenticate(ctx, req)
		if err != nil {
			// Log error but continue to next provider
			continue
		}
		
		if result.Success {
			return result, nil
		}
	}
	
	// No provider succeeded
	return &AuthResult{
		Success: false,
		Error:   ErrCodeAuthenticationFailed,
	}, NewError(ErrCodeAuthenticationFailed, "authentication failed", "no provider could authenticate the request")
}

// ValidateToken validates a token using the appropriate provider
func (m *manager) ValidateToken(ctx context.Context, token string) (*AuthResult, error) {
	if !m.config.Enabled {
		return &AuthResult{Success: true}, nil
	}
	
	if token == "" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeMissingToken,
		}, NewError(ErrCodeMissingToken, "missing authentication token", "")
	}
	
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	m.mu.RUnlock()
	
	// Try each provider that supports token validation
	for _, provider := range providers {
		result, err := provider.ValidateToken(ctx, token)
		if err != nil {
			continue
		}
		
		if result.Success {
			return result, nil
		}
	}
	
	return &AuthResult{
		Success: false,
		Error:   ErrCodeTokenInvalid,
	}, NewError(ErrCodeTokenInvalid, "invalid authentication token", "")
}

// GetMiddleware returns HTTP middleware for authentication
func (m *manager) GetMiddleware(options ...MiddlewareOption) http.Handler {
	config := *m.middlewareConfig // Copy default config
	
	// Apply options
	for _, option := range options {
		option(&config)
	}
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required
		if !config.Required {
			// Add context with empty auth info and continue
			ctx := setAuthContext(r.Context(), &AuthContext{})
			r = r.WithContext(ctx)
			http.DefaultServeMux.ServeHTTP(w, r)
			return
		}
		
		// Check bypass paths
		for _, path := range config.BypassPaths {
			if r.URL.Path == path {
				ctx := setAuthContext(r.Context(), &AuthContext{})
				r = r.WithContext(ctx)
				http.DefaultServeMux.ServeHTTP(w, r)
				return
			}
		}
		
		// Check HTTPS requirement
		if config.RequireHTTPS && r.URL.Scheme != "https" {
			w.WriteHeader(config.UnauthorizedCode)
			fmt.Fprintf(w, `{"error": "https_required", "message": "HTTPS is required for authentication"}`)
			return
		}
		
		// Extract token from various sources
		token := extractToken(r, &config)
		
		var authResult *AuthResult
		var err error
		
		if token != "" {
			// Validate existing token
			authResult, err = m.ValidateToken(r.Context(), token)
		} else {
			// Try to authenticate with request
			authResult, err = m.Authenticate(r.Context(), r)
		}
		
		if err != nil || !authResult.Success {
			// Authentication failed
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, config.Realm))
			w.WriteHeader(config.UnauthorizedCode)
			fmt.Fprintf(w, `{"error": "unauthorized", "message": "Authentication required"}`)
			return
		}
		
		// Create auth context
		authCtx := &AuthContext{
			User:      authResult.User,
			Token:     authResult.Token,
			ExpiresAt: authResult.ExpiresAt,
			Metadata:  authResult.Metadata,
			Method:    determineAuthMethod(r, token),
		}
		
		// Add auth context to request
		ctx := setAuthContext(r.Context(), authCtx)
		r = r.WithContext(ctx)
		
		// Continue with request
		http.DefaultServeMux.ServeHTTP(w, r)
	})
}

// Initialize initializes the authentication manager
func (m *manager) Initialize(config *Config) error {
	if config == nil {
		return NewError(ErrCodeConfigurationError, "configuration cannot be nil", "")
	}
	
	m.config = config
	
	// Apply defaults
	if config.SessionTimeout == "" {
		config.SessionTimeout = "24h"
	}
	if config.TokenSecret == "" {
		config.TokenSecret = "default-secret-change-in-production"
	}
	
	// Initialize providers based on configuration
	for _, providerConfig := range config.Providers {
		if !providerConfig.Enabled {
			continue
		}
		
		provider, err := createProvider(providerConfig)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", providerConfig.Name, err)
		}
		
		if err := provider.Initialize(providerConfig); err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", providerConfig.Name, err)
		}
		
		if err := m.RegisterProvider(provider); err != nil {
			return fmt.Errorf("failed to register provider %s: %w", providerConfig.Name, err)
		}
	}
	
	return nil
}

// Shutdown gracefully shuts down the authentication manager
func (m *manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var errors []error
	
	for name, provider := range m.providers {
		if err := provider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown provider %s: %w", name, err))
		}
	}
	
	m.providers = make(map[string]Provider)
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	return nil
}

// Helper functions

// createProvider creates a provider based on configuration
func createProvider(config ProviderConfig) (Provider, error) {
	switch config.Type {
	case ProviderTypeBasic:
		return NewBasicProvider(), nil
	case ProviderTypeJWT:
		return NewJWTProvider(), nil
	case ProviderTypeOAuth:
		return NewOAuthProvider(), nil
	case ProviderTypeStatic:
		return NewStaticProvider(), nil
	default:
		return nil, NewError(ErrCodeInvalidProvider, 
			fmt.Sprintf("unknown provider type: %s", config.Type), "")
	}
}

// Context key for auth context
type contextKey string

const authContextKey = contextKey("auth")

// setAuthContext sets authentication context in request context
func setAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// GetAuthContext retrieves authentication context from request context
func GetAuthContext(req *http.Request) (*AuthContext, bool) {
	authCtx, ok := req.Context().Value(authContextKey).(*AuthContext)
	return authCtx, ok
}

// GetUser retrieves authenticated user from request context
func GetUser(req *http.Request) (*User, bool) {
	if authCtx, ok := GetAuthContext(req); ok {
		return authCtx.User, authCtx.User != nil
	}
	return nil, false
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(req *http.Request) bool {
	_, ok := GetAuthContext(req)
	return ok
}

// HasRole checks if authenticated user has a specific role
func HasRole(req *http.Request, role string) bool {
	if user, ok := GetUser(req); ok {
		for _, userRole := range user.Roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

