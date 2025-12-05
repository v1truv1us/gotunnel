package auth

import (
	"context"
	"net/http"
	"time"
)

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	Success   bool                   `json:"success"`
	User      *User                  `json:"user,omitempty"`
	Token     string                 `json:"token,omitempty"`
	ExpiresAt time.Time              `json:"expires_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// User represents an authenticated user
type User struct {
	ID       string                 `json:"id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Roles    []string               `json:"roles,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// AuthContext holds authentication context for a request
type AuthContext struct {
	User      *User                  `json:"user"`
	Token     string                 `json:"token,omitempty"`
	ExpiresAt time.Time              `json:"expires_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Method    string                 `json:"method"` // "basic", "jwt", "oauth", etc.
}

// Provider defines the interface for authentication providers
type Provider interface {
	// Name returns the provider name
	Name() string
	
	// Type returns the provider type
	Type() ProviderType
	
	// Authenticate authenticates a request
	Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error)
	
	// ValidateToken validates a token and returns user info
	ValidateToken(ctx context.Context, token string) (*AuthResult, error)
	
	// RefreshToken refreshes an existing token
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	
	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, token string) error
	
	// Initialize initializes the provider with configuration
	Initialize(config ProviderConfig) error
	
	// Shutdown gracefully shuts down the provider
	Shutdown(ctx context.Context) error
}

// ProviderType represents the type of authentication provider
type ProviderType string

const (
	ProviderTypeBasic   ProviderType = "basic"
	ProviderTypeJWT     ProviderType = "jwt"
	ProviderTypeOAuth   ProviderType = "oauth"
	ProviderTypeStatic  ProviderType = "static"
)

// ProviderConfig holds configuration for authentication providers
type ProviderConfig struct {
	Type        ProviderType          `yaml:"type" json:"type"`
	Name        string                `yaml:"name" json:"name"`
	Enabled     bool                  `yaml:"enabled" json:"enabled"`
	Priority    int                   `yaml:"priority" json:"priority"`
	Config      map[string]any `yaml:"config" json:"config"`
}

// Manager manages authentication providers and middleware
type Manager interface {
	// RegisterProvider registers an authentication provider
	RegisterProvider(provider Provider) error
	
	// UnregisterProvider unregisters an authentication provider
	UnregisterProvider(name string) error
	
	// GetProvider returns a provider by name
	GetProvider(name string) (Provider, bool)
	
	// ListProviders returns all registered providers
	ListProviders() []Provider
	
	// Authenticate authenticates a request using all registered providers
	Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error)
	
	// ValidateToken validates a token using the appropriate provider
	ValidateToken(ctx context.Context, token string) (*AuthResult, error)
	
	// GetMiddleware returns HTTP middleware for authentication
	GetMiddleware(options ...MiddlewareOption) http.Handler
	
	// Initialize initializes the authentication manager
	Initialize(config *Config) error
	
	// Shutdown gracefully shuts down the authentication manager
	Shutdown(ctx context.Context) error
}

// Config holds authentication configuration
type Config struct {
	// Global authentication settings
	Enabled          bool              `yaml:"enabled" json:"enabled"`
	DefaultProvider  string            `yaml:"default_provider" json:"default_provider"`
	RequireAuth      bool              `yaml:"require_auth" json:"require_auth"`
	BypassPaths      []string          `yaml:"bypass_paths" json:"bypass_paths"`
	SessionTimeout   string            `yaml:"session_timeout" json:"session_timeout"`
	TokenSecret      string            `yaml:"token_secret" json:"token_secret"`
	
	// Provider configurations
	Providers        []ProviderConfig  `yaml:"providers" json:"providers"`
	
	// JWT configuration
	JWT              *JWTConfig        `yaml:"jwt" json:"jwt"`
	
	// OAuth configuration
	OAuth            *OAuthConfig      `yaml:"oauth" json:"oauth"`
	
	// Static user configuration
	StaticUsers      []*StaticUser     `yaml:"static_users" json:"static_users"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret           string        `yaml:"secret" json:"secret"`
	Expiration       string        `yaml:"expiration" json:"expiration"`
	RefreshExpiration string       `yaml:"refresh_expiration" json:"refresh_expiration"`
	Issuer           string        `yaml:"issuer" json:"issuer"`
	Audience         string        `yaml:"audience" json:"audience"`
	Algorithm        string        `yaml:"algorithm" json:"algorithm"`
}

// OAuthConfig holds OAuth-specific configuration
type OAuthConfig struct {
	Providers        []OAuthProviderConfig `yaml:"providers" json:"providers"`
	SuccessRedirect  string               `yaml:"success_redirect" json:"success_redirect"`
	FailureRedirect  string               `yaml:"failure_redirect" json:"failure_redirect"`
}

// OAuthProviderConfig holds configuration for a specific OAuth provider
type OAuthProviderConfig struct {
	Name             string `yaml:"name" json:"name"`
	ClientID         string `yaml:"client_id" json:"client_id"`
	ClientSecret     string `yaml:"client_secret" json:"client_secret"`
	AuthURL          string `yaml:"auth_url" json:"auth_url"`
	TokenURL         string `yaml:"token_url" json:"token_url"`
	UserInfoURL      string `yaml:"user_info_url" json:"user_info_url"`
	Scopes           []string `yaml:"scopes" json:"scopes"`
	Enabled          bool     `yaml:"enabled" json:"enabled"`
}

// StaticUser holds static user configuration
type StaticUser struct {
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	Email    string   `yaml:"email" json:"email"`
	Name     string   `yaml:"name" json:"name"`
	Roles    []string `yaml:"roles" json:"roles"`
	Enabled  bool     `yaml:"enabled" json:"enabled"`
}

// MiddlewareOption configures authentication middleware
type MiddlewareOption func(*MiddlewareConfig)

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	// Authentication settings
	Required         bool     `json:"required"`
	Providers        []string `json:"providers"`
	BypassPaths      []string `json:"bypass_paths"`
	
	// Token settings
	TokenHeader      string `json:"token_header"`
	TokenQueryParam  string `json:"token_query_param"`
	CookieName       string `json:"cookie_name"`
	
	// Session settings
	SessionTimeout   time.Duration `json:"session_timeout"`
	RefreshThreshold time.Duration `json:"refresh_threshold"`
	
	// Security settings
	RequireHTTPS     bool `json:"require_https"`
	CSRFProtection   bool `json:"csrf_protection"`
	
	// Response settings
	UnauthorizedCode int    `json:"unauthorized_code"`
	ForbiddenCode    int    `json:"forbidden_code"`
	Realm            string `json:"realm"`
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		Required:          true,
		Providers:         []string{},
		BypassPaths:       []string{"/health", "/metrics", "/favicon.ico"},
		TokenHeader:       "Authorization",
		TokenQueryParam:   "token",
		CookieName:        "auth_token",
		SessionTimeout:    time.Hour * 24,
		RefreshThreshold:  time.Hour * 23,
		RequireHTTPS:      false,
		CSRFProtection:    false,
		UnauthorizedCode:  http.StatusUnauthorized,
		ForbiddenCode:     http.StatusForbidden,
		Realm:            "gotunnel",
	}
}

// Middleware options

// WithRequired sets whether authentication is required
func WithRequired(required bool) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Required = required
	}
}

// WithProviders sets the authentication providers to use
func WithProviders(providers ...string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Providers = providers
	}
}

// WithBypassPaths sets paths that bypass authentication
func WithBypassPaths(paths ...string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.BypassPaths = append(c.BypassPaths, paths...)
	}
}

// WithTokenHeader sets the token header name
func WithTokenHeader(header string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.TokenHeader = header
	}
}

// WithSessionTimeout sets the session timeout
func WithSessionTimeout(timeout time.Duration) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.SessionTimeout = timeout
	}
}

// WithRequireHTTPS sets whether HTTPS is required
func WithRequireHTTPS(require bool) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.RequireHTTPS = require
	}
}

// WithRealm sets the authentication realm
func WithRealm(realm string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Realm = realm
	}
}

// Error types

// Error represents an authentication error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeInvalidCredentials    = "invalid_credentials"
	ErrCodeTokenExpired         = "token_expired"
	ErrCodeTokenInvalid         = "token_invalid"
	ErrCodeProviderNotFound     = "provider_not_found"
	ErrCodeAuthenticationFailed = "authentication_failed"
	ErrCodeUnauthorized        = "unauthorized"
	ErrCodeForbidden           = "forbidden"
	ErrCodeMissingToken        = "missing_token"
	ErrCodeInvalidProvider     = "invalid_provider"
	ErrCodeConfigurationError  = "configuration_error"
)

// NewError creates a new authentication error
func NewError(code, message, details string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// IsErrorCode checks if an error matches a specific error code
func IsErrorCode(err error, code string) bool {
	if authErr, ok := err.(*Error); ok {
		return authErr.Code == code
	}
	return false
}