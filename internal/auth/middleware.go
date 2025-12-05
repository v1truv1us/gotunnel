package auth

import (
	"net/http"
)

// Middleware creates HTTP authentication middleware
func Middleware(manager Manager, options ...MiddlewareOption) func(http.Handler) http.Handler {
	config := DefaultMiddlewareConfig()
	
	// Apply options
	for _, option := range options {
		option(config)
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if authentication is required
			if !config.Required {
				// Add context with empty auth info and continue
				ctx := setAuthContext(r.Context(), &AuthContext{})
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
			
			// Check bypass paths
			for _, path := range config.BypassPaths {
				if r.URL.Path == path {
					ctx := setAuthContext(r.Context(), &AuthContext{})
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
			}
			
			// Check HTTPS requirement
			if config.RequireHTTPS && r.URL.Scheme != "https" {
				w.WriteHeader(config.UnauthorizedCode)
				w.Write([]byte(`{"error": "https_required", "message": "HTTPS is required for authentication"}`))
				return
			}
			
			// Extract token from various sources
			token := extractToken(r, config)
			
			var authResult *AuthResult
			var err error
			
			if token != "" {
				// Validate existing token
				authResult, err = manager.ValidateToken(r.Context(), token)
			} else {
				// Try to authenticate with request
				authResult, err = manager.Authenticate(r.Context(), r)
			}
			
			if err != nil || !authResult.Success {
				// Authentication failed
				w.Header().Set("WWW-Authenticate", `Bearer realm="`+config.Realm+`"`)
				w.WriteHeader(config.UnauthorizedCode)
				w.Write([]byte(`{"error": "unauthorized", "message": "Authentication required"}`))
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
			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts authentication token from request
func extractToken(req *http.Request, config *MiddlewareConfig) string {
	// Try Authorization header first
	if authHeader := req.Header.Get(config.TokenHeader); authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
		// Support Basic auth as well
		if len(authHeader) > 6 && authHeader[:6] == "Basic " {
			return authHeader
		}
	}
	
	// Try query parameter
	if token := req.URL.Query().Get(config.TokenQueryParam); token != "" {
		return token
	}
	
	// Try cookie
	if cookie, err := req.Cookie(config.CookieName); err == nil {
		return cookie.Value
	}
	
	return ""
}

// determineAuthMethod determines the authentication method used
func determineAuthMethod(req *http.Request, token string) string {
	authHeader := req.Header.Get("Authorization")
	
	if authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return "jwt"
		}
		if len(authHeader) > 6 && authHeader[:6] == "Basic " {
			return "basic"
		}
	}
	
	if token != "" {
		return "token"
	}
	
	return "unknown"
}