package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/johncferguson/gotunnel/internal/auth"
)

// Context key for auth context
type contextKey string

const authContextKey = contextKey("auth")

// SecurityConfig holds security configuration for proxy
type SecurityConfig struct {
	// Authentication settings
	Enabled          bool              `yaml:"enabled" json:"enabled"`
	Required         bool              `yaml:"required" json:"required"`
	Providers        []string          `yaml:"providers" json:"providers"`
	BypassPaths      []string          `yaml:"bypass_paths" json:"bypass_paths"`
	
	// Per-tunnel authentication
	TunnelAuth      map[string]bool   `yaml:"tunnel_auth" json:"tunnel_auth"`
	DefaultAuth     bool              `yaml:"default_auth" json:"default_auth"`
	
	// Security headers
	SecurityHeaders bool              `yaml:"security_headers" json:"security_headers"`
	CustomHeaders   map[string]string `yaml:"custom_headers" json:"custom_headers"`
	
	// Session settings
	SessionTimeout   string `yaml:"session_timeout" json:"session_timeout"`
	RefreshThreshold string `yaml:"refresh_threshold" json:"refresh_threshold"`
}

// SecurityMiddleware wraps HTTP handlers with security features
type SecurityMiddleware struct {
	authManager auth.Manager
	config      SecurityConfig
	next        http.Handler
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(authManager auth.Manager, config SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		authManager: authManager,
		config:      config,
	}
}

// ServeHTTP implements http.Handler interface
func (sm *SecurityMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Skip security if disabled
	if !sm.config.Enabled {
		sm.next.ServeHTTP(w, r)
		return
	}

	// Check bypass paths
	for _, path := range sm.config.BypassPaths {
		if r.URL.Path == path {
			sm.next.ServeHTTP(w, r)
			return
		}
	}

	// Perform authentication
	authResult, err := sm.authManager.Authenticate(r.Context(), r)
	if err != nil {
		sm.sendErrorResponse(w, http.StatusInternalServerError, "authentication error", err.Error())
		return
	}

	if !authResult.Success {
		sm.sendAuthErrorResponse(w, r)
		return
	}

	// Add authentication context to request
	authCtx := &auth.AuthContext{
		User:      authResult.User,
		Token:     authResult.Token,
		ExpiresAt: authResult.ExpiresAt,
		Metadata:  authResult.Metadata,
		Method:    sm.determineAuthMethod(r),
	}

	ctx := context.WithValue(r.Context(), authContextKey, authCtx)
	r = r.WithContext(ctx)

	// Add security headers
	if sm.config.SecurityHeaders {
		sm.addSecurityHeaders(w)
	}

	// Add custom headers
	for key, value := range sm.config.CustomHeaders {
		w.Header().Set(key, value)
	}

	// Continue to next handler
	sm.next.ServeHTTP(w, r)
}

// WrapHandler wraps an http.Handler with security middleware
func (sm *SecurityMiddleware) WrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm.ServeHTTP(w, r)
		handler.ServeHTTP(w, r)
	})
}

	// Check bypass paths
	for _, path := range sm.config.BypassPaths {
		if r.URL.Path == path {
			sm.next.ServeHTTP(w, r)
			return
		}
	}

	// Check if authentication is required for this tunnel
	tunnelDomain := sm.extractTunnelDomain(r)
	requiresAuth := sm.config.DefaultAuth
	
	if tunnelDomain != "" {
		if authRequired, exists := sm.config.TunnelAuth[tunnelDomain]; exists {
			requiresAuth = authRequired
		}
	}

	// Skip authentication if not required
	if !requiresAuth {
		sm.next.ServeHTTP(w, r)
		return
	}

	// Perform authentication
	authResult, err := sm.authManager.Authenticate(r.Context(), r)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "authentication error", err.Error())
		return
	}

	if !authResult.Success {
		sm.sendAuthErrorResponse(w, r)
		return
	}

	// Add authentication context to request
	authCtx := &auth.AuthContext{
		User:      authResult.User,
		Token:     authResult.Token,
		ExpiresAt: authResult.ExpiresAt,
		Metadata:  authResult.Metadata,
		Method:    sm.determineAuthMethod(r),
	}
	
	ctx := context.WithValue(r.Context(), authContextKey, authCtx)
	r = r.WithContext(ctx)

	// Add security headers
	if sm.config.SecurityHeaders {
		sm.addSecurityHeaders(w)
	}

	// Add custom headers
	for key, value := range sm.config.CustomHeaders {
		w.Header().Set(key, value)
	}

	// Continue to next handler
	sm.next.ServeHTTP(w, r)
}

// extractTunnelDomain extracts tunnel domain from request
func (sm *SecurityMiddleware) extractTunnelDomain(r *http.Request) string {
	host := strings.Split(r.Host, ":")[0] // Remove port from host header
	
	// Remove .local suffix if present
	if strings.HasSuffix(host, ".local") {
		return strings.TrimSuffix(host, ".local")
	}
	
	return host
}

// determineAuthMethod determines authentication method used
func (sm *SecurityMiddleware) determineAuthMethod(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return "jwt"
		}
		if strings.HasPrefix(authHeader, "Basic ") {
			return "basic"
		}
	}
	
	// Check for token in query or cookie
	if r.URL.Query().Get("token") != "" {
		return "token"
	}
	
	if cookie, err := r.Cookie("auth_token"); err == nil && cookie != nil {
		return "cookie"
	}
	
	return "unknown"
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	fmt.Fprintf(w, `{"error": true, "message": "%s", "details": "%s", "timestamp": "%s"}`,
		message, details, time.Now().UTC().Format(time.RFC3339))
}

// sendAuthErrorResponse sends authentication error response
func (sm *SecurityMiddleware) sendAuthErrorResponse(w http.ResponseWriter, r *http.Request) {
	// Determine appropriate response based on Accept header
	acceptHeader := r.Header.Get("Accept")
	
	if strings.Contains(acceptHeader, "text/html") {
		// HTML response for browsers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Required</title></head>
<body>
<h1>🔐 Authentication Required</h1>
<p>This tunnel requires authentication to access.</p>
<p>Please provide valid credentials to continue.</p>
<hr>
<em>gotunnel security proxy</em>
</body>
</html>`)
		return
	}
	
	// JSON response for API clients
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="gotunnel"`)
	w.WriteHeader(http.StatusUnauthorized)
	
	fmt.Fprintf(w, `{"error": true, "message": "Authentication required", "code": "auth_required", "timestamp": "%s", "realm": "gotunnel"}`,
		time.Now().UTC().Format(time.RFC3339))
}

// addSecurityHeaders adds security headers to response
func (sm *SecurityMiddleware) addSecurityHeaders(w http.ResponseWriter) {
	// Prevent clickjacking
	w.Header().Set("X-Frame-Options", "DENY")
	
	// Prevent MIME type sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	// Enable XSS protection
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	// Force HTTPS (if needed)
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	
	// Content Security Policy
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	
	// Referrer policy
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}



// WrapHandler wraps an http.Handler with security middleware
func (sm *SecurityMiddleware) WrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm.ServeHTTP(w, r)
		handler.ServeHTTP(w, r)
	})
}

// SecurityChain represents a chain of security middleware
type SecurityChain struct {
	middlewares []func(http.Handler) http.Handler
}

// NewSecurityChain creates a new security chain
func NewSecurityChain() *SecurityChain {
	return &SecurityChain{
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

// Add adds a middleware to the chain
func (sc *SecurityChain) Add(middleware func(http.Handler) http.Handler) {
	sc.middlewares = append(sc.middlewares, middleware)
}

// Wrap wraps a handler with the entire security chain
func (sc *SecurityChain) Wrap(handler http.Handler) http.Handler {
	for i := len(sc.middlewares) - 1; i >= 0; i-- {
		handler = sc.middlewares[i](handler)
	}
	return handler
}