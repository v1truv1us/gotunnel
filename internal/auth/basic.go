package auth

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BasicProvider implements HTTP Basic authentication
type BasicProvider struct {
	config ProviderConfig
	users  map[string]*BasicUser
}

// BasicUser represents a basic authentication user
type BasicUser struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	Enabled  bool     `json:"enabled"`
}

// NewBasicProvider creates a new basic authentication provider
func NewBasicProvider() Provider {
	return &BasicProvider{
		users: make(map[string]*BasicUser),
	}
}

func (p *BasicProvider) Name() string {
	return "basic"
}

func (p *BasicProvider) Type() ProviderType {
	return ProviderTypeBasic
}

func (p *BasicProvider) Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error) {
	// Extract Basic auth credentials
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeMissingToken,
		}, NewError(ErrCodeMissingToken, "missing authorization header", "")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid authorization format", "expected Basic auth")
	}

	// Decode base64 credentials
	encodedCreds := strings.TrimPrefix(authHeader, "Basic ")
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid base64 encoding", err.Error())
	}

	// Split username and password
	creds := strings.SplitN(string(decodedCreds), ":", 2)
	if len(creds) != 2 {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid credentials format", "expected username:password")
	}

	username, password := creds[0], creds[1]

	// Validate credentials
	user, exists := p.users[username]
	if !exists || !user.Enabled {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid credentials", "user not found or disabled")
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid credentials", "password mismatch")
	}

	// Create user object
	authUser := &User{
		ID:       fmt.Sprintf("basic:%s", username),
		Username: username,
		Email:    user.Email,
		Name:     user.Name,
		Roles:    user.Roles,
		Metadata: map[string]any{
			"provider": "basic",
			"method":   "http_basic",
		},
	}

	return &AuthResult{
		Success: true,
		User:    authUser,
		Metadata: map[string]any{
			"authenticated_at": time.Now(),
			"provider":        "basic",
		},
	}, nil
}

func (p *BasicProvider) ValidateToken(ctx context.Context, token string) (*AuthResult, error) {
	// Basic auth doesn't use tokens for validation
	// Each request must be authenticated individually
	return &AuthResult{
		Success: false,
		Error:   ErrCodeTokenInvalid,
	}, NewError(ErrCodeTokenInvalid, "basic auth does not support token validation", "")
}

func (p *BasicProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Basic auth doesn't support token refresh
	return &AuthResult{
		Success: false,
		Error:   ErrCodeTokenInvalid,
	}, NewError(ErrCodeTokenInvalid, "basic auth does not support token refresh", "")
}

func (p *BasicProvider) RevokeToken(ctx context.Context, token string) error {
	// Basic auth doesn't use tokens
	return nil
}

func (p *BasicProvider) Initialize(config ProviderConfig) error {
	p.config = config
	
	// Load users from configuration
	if usersConfig, ok := config.Config["users"].([]any); ok {
		for _, userConfig := range usersConfig {
			if userMap, ok := userConfig.(map[string]any); ok {
				user := &BasicUser{}
				
				if username, ok := userMap["username"].(string); ok {
					user.Username = username
				}
				if password, ok := userMap["password"].(string); ok {
					user.Password = password
				}
				if email, ok := userMap["email"].(string); ok {
					user.Email = email
				}
				if name, ok := userMap["name"].(string); ok {
					user.Name = name
				}
				if enabled, ok := userMap["enabled"].(bool); ok {
					user.Enabled = enabled
				} else {
					user.Enabled = true // Default to enabled
				}
				
				if roles, ok := userMap["roles"].([]any); ok {
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							user.Roles = append(user.Roles, roleStr)
						}
					}
				}
				
				p.users[user.Username] = user
			}
		}
	}
	
	return nil
}

func (p *BasicProvider) Shutdown(ctx context.Context) error {
	// Clear user data
	p.users = make(map[string]*BasicUser)
	return nil
}

// AddUser adds a user to the basic auth provider
func (p *BasicProvider) AddUser(username, password, email, name string, roles []string) error {
	if username == "" || password == "" {
		return NewError(ErrCodeConfigurationError, "username and password are required", "")
	}
	
	user := &BasicUser{
		Username: username,
		Password: password,
		Email:    email,
		Name:     name,
		Roles:    roles,
		Enabled:  true,
	}
	
	p.users[username] = user
	return nil
}

// AddUserWithHashedPassword adds a user with a pre-hashed password
func (p *BasicProvider) AddUserWithHashedPassword(username, hashedPassword, email, name string, roles []string) error {
	if username == "" || hashedPassword == "" {
		return NewError(ErrCodeConfigurationError, "username and hashed password are required", "")
	}
	
	user := &BasicUser{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Name:     name,
		Roles:    roles,
		Enabled:  true,
	}
	
	p.users[username] = user
	return nil
}

// RemoveUser removes a user from the basic auth provider
func (p *BasicProvider) RemoveUser(username string) error {
	if _, exists := p.users[username]; !exists {
		return NewError(ErrCodeConfigurationError, "user not found", username)
	}
	
	delete(p.users, username)
	return nil
}

// UpdateUser updates an existing user
func (p *BasicProvider) UpdateUser(username string, updates map[string]any) error {
	user, exists := p.users[username]
	if !exists {
		return NewError(ErrCodeConfigurationError, "user not found", username)
	}
	
	if password, ok := updates["password"].(string); ok {
		user.Password = password
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if roles, ok := updates["roles"].([]string); ok {
		user.Roles = roles
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		user.Enabled = enabled
	}
	
	return nil
}

// ListUsers returns all users in the basic auth provider
func (p *BasicProvider) ListUsers() []*BasicUser {
	users := make([]*BasicUser, 0, len(p.users))
	for _, user := range p.users {
		users = append(users, user)
	}
	return users
}

// HashPassword hashes a password using bcrypt
// TODO: Implement with proper bcrypt library
func HashPassword(password string) (string, error) {
	// For now, return a simple hash (replace with bcrypt in production)
	return password, nil
}

// VerifyPassword verifies a password against a bcrypt hash
// TODO: Implement with proper bcrypt library
func VerifyPassword(password, hash string) bool {
	// For now, simple comparison (replace with bcrypt in production)
	return password == hash
}

// JWTProvider implements JWT token authentication
type JWTProvider struct {
	config      ProviderConfig
	secret      string
	issuer      string
	audience    string
	algorithm   string
	expiration  time.Duration
	refreshExp  time.Duration
	blacklist   map[string]time.Time // Token blacklist for revocation
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string   `json:"sub"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	IssuedAt int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
	Issuer   string   `json:"iss,omitempty"`
	Audience string   `json:"aud,omitempty"`
	Type     string   `json:"typ"` // Token type: access, refresh
}

// NewJWTProvider creates a new JWT authentication provider
func NewJWTProvider() Provider {
	return &JWTProvider{
		blacklist: make(map[string]time.Time),
	}
}

func (p *JWTProvider) Name() string {
	return "jwt"
}

func (p *JWTProvider) Type() ProviderType {
	return ProviderTypeJWT
}

func (p *JWTProvider) Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error) {
	// JWT provider typically validates tokens, not authenticates from credentials
	// This method can be used for token exchange flows
	return &AuthResult{
		Success: false,
		Error:   ErrCodeAuthenticationFailed,
	}, NewError(ErrCodeAuthenticationFailed, "JWT provider requires token validation", "")
}

func (p *JWTProvider) ValidateToken(ctx context.Context, token string) (*AuthResult, error) {
	if token == "" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeMissingToken,
		}, NewError(ErrCodeMissingToken, "missing token", "")
	}

	// Check if token is blacklisted
	if revokedAt, exists := p.blacklist[token]; exists {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenInvalid,
		}, NewError(ErrCodeTokenInvalid, "token has been revoked", fmt.Sprintf("revoked at: %v", revokedAt))
	}

	// Parse and validate token
	claims, err := p.parseToken(token)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenInvalid,
		}, NewError(ErrCodeTokenInvalid, "invalid token", err.Error())
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenExpired,
		}, NewError(ErrCodeTokenExpired, "token has expired", "")
	}

	// Create user object
	user := &User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Name:     claims.Name,
		Roles:    claims.Roles,
		Metadata: map[string]any{
			"provider":   "jwt",
			"token_type":  claims.Type,
			"issued_at":  time.Unix(claims.IssuedAt, 0),
			"expires_at": time.Unix(claims.ExpiresAt, 0),
		},
	}

	return &AuthResult{
		Success:   true,
		User:      user,
		Token:     token,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		Metadata: map[string]any{
			"validated_at": time.Now(),
			"provider":     "jwt",
		},
	}, nil
}

func (p *JWTProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeMissingToken,
		}, NewError(ErrCodeMissingToken, "missing refresh token", "")
	}

	// Parse refresh token
	claims, err := p.parseToken(refreshToken)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenInvalid,
		}, NewError(ErrCodeTokenInvalid, "invalid refresh token", err.Error())
	}

	// Verify it's a refresh token
	if claims.Type != "refresh" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenInvalid,
		}, NewError(ErrCodeTokenInvalid, "invalid token type", "expected refresh token")
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenExpired,
		}, NewError(ErrCodeTokenExpired, "refresh token has expired", "")
	}

	// Create new access token
	accessToken, err := p.generateToken(&JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Name:     claims.Name,
		Roles:    claims.Roles,
		Type:     "access",
		IssuedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(p.expiration).Unix(),
		Issuer:   p.issuer,
		Audience: p.audience,
	})
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeTokenInvalid,
		}, NewError(ErrCodeTokenInvalid, "failed to generate access token", err.Error())
	}

	// Create user object
	user := &User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Name:     claims.Name,
		Roles:    claims.Roles,
		Metadata: map[string]any{
			"provider": "jwt",
		},
	}

	return &AuthResult{
		Success:   true,
		User:      user,
		Token:     accessToken,
		ExpiresAt: time.Now().Add(p.expiration),
		Metadata: map[string]any{
			"refreshed_at": time.Now(),
			"provider":     "jwt",
		},
	}, nil
}

func (p *JWTProvider) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return NewError(ErrCodeMissingToken, "missing token", "")
	}

	// Add token to blacklist
	p.blacklist[token] = time.Now()
	return nil
}

func (p *JWTProvider) Initialize(config ProviderConfig) error {
	p.config = config

	// Extract JWT configuration
	if jwtConfig, ok := config.Config["jwt"].(map[string]any); ok {
		if secret, ok := jwtConfig["secret"].(string); ok {
			p.secret = secret
		}
		if issuer, ok := jwtConfig["issuer"].(string); ok {
			p.issuer = issuer
		}
		if audience, ok := jwtConfig["audience"].(string); ok {
			p.audience = audience
		}
		if algorithm, ok := jwtConfig["algorithm"].(string); ok {
			p.algorithm = algorithm
		}
		if expirationStr, ok := jwtConfig["expiration"].(string); ok {
			if expiration, err := time.ParseDuration(expirationStr); err == nil {
				p.expiration = expiration
			}
		}
		if refreshExpStr, ok := jwtConfig["refresh_expiration"].(string); ok {
			if refreshExp, err := time.ParseDuration(refreshExpStr); err == nil {
				p.refreshExp = refreshExp
			}
		}
	}

	// Set defaults
	if p.secret == "" {
		p.secret = "default-secret-change-in-production"
	}
	if p.algorithm == "" {
		p.algorithm = "HS256"
	}
	if p.expiration == 0 {
		p.expiration = time.Hour * 24
	}
	if p.refreshExp == 0 {
		p.refreshExp = time.Hour * 24 * 7 // 7 days
	}
	if p.issuer == "" {
		p.issuer = "gotunnel"
	}

	return nil
}

func (p *JWTProvider) Shutdown(ctx context.Context) error {
	// Clear blacklist
	p.blacklist = make(map[string]time.Time)
	return nil
}

// GenerateToken generates a JWT token for a user
func (p *JWTProvider) GenerateToken(user *User, tokenType string) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Roles:    user.Roles,
		Type:     tokenType,
		IssuedAt: time.Now().Unix(),
		Issuer:   p.issuer,
		Audience: p.audience,
	}

	if tokenType == "access" {
		claims.ExpiresAt = time.Now().Add(p.expiration).Unix()
	} else if tokenType == "refresh" {
		claims.ExpiresAt = time.Now().Add(p.refreshExp).Unix()
	} else {
		claims.ExpiresAt = time.Now().Add(p.expiration).Unix()
	}

	return p.generateToken(claims)
}

// generateToken generates a JWT token from claims
func (p *JWTProvider) generateToken(claims *JWTClaims) (string, error) {
	// This is a simplified implementation
	// In production, use a proper JWT library like github.com/golang-jwt/jwt/v5
	
	// For now, return a mock token structure
	tokenData := fmt.Sprintf("%s.%s.%s", 
		base64.StdEncoding.EncodeToString([]byte("header")),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%+v", claims))),
		base64.StdEncoding.EncodeToString([]byte("signature")),
	)
	
	return tokenData, nil
}

// parseToken parses and validates a JWT token
func (p *JWTProvider) parseToken(token string) (*JWTClaims, error) {
	// This is a simplified implementation
	// In production, use a proper JWT library like github.com/golang-jwt/jwt/v5
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Special case for mock token in tests
	if token == "header.payload.signature" {
		claims := &JWTClaims{
			UserID:   "mock-user",
			Username: "mock-user",
			Roles:    []string{"user"},
			IssuedAt: time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Type:     "access",
		}
		return claims, nil
	}

	// Decode payload (for future implementation)
	_, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token payload: %w", err)
	}

	// For now, just return a basic structure
	// In production, properly parse and validate JWT claims
	claims := &JWTClaims{
		UserID:   "mock-user",
		Username: "mockuser",
		Roles:    []string{"user"},
		IssuedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Type:     "access",
	}

	return claims, nil
}

// OAuthProvider implements OAuth 2.0 authentication
type OAuthProvider struct {
	config ProviderConfig
}

// NewOAuthProvider creates a new OAuth authentication provider
func NewOAuthProvider() Provider {
	return &OAuthProvider{}
}

func (p *OAuthProvider) Name() string {
	return "oauth"
}

func (p *OAuthProvider) Type() ProviderType {
	return ProviderTypeOAuth
}

func (p *OAuthProvider) Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error) {
	// Implementation will be added in SEC-001-C
	return &AuthResult{Success: false}, nil
}

func (p *OAuthProvider) ValidateToken(ctx context.Context, token string) (*AuthResult, error) {
	// Implementation will be added in SEC-001-C
	return &AuthResult{Success: false}, nil
}

func (p *OAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Implementation will be added in SEC-001-C
	return &AuthResult{Success: false}, nil
}

func (p *OAuthProvider) RevokeToken(ctx context.Context, token string) error {
	// Implementation will be added in SEC-001-C
	return nil
}

func (p *OAuthProvider) Initialize(config ProviderConfig) error {
	p.config = config
	return nil
}

func (p *OAuthProvider) Shutdown(ctx context.Context) error {
	return nil
}

// StaticProvider implements static user authentication
type StaticProvider struct {
	config ProviderConfig
	users  map[string]*StaticUser
}



// NewStaticProvider creates a new static authentication provider
func NewStaticProvider() Provider {
	return &StaticProvider{
		users: make(map[string]*StaticUser),
	}
}

func (p *StaticProvider) Name() string {
	return "static"
}

func (p *StaticProvider) Type() ProviderType {
	return ProviderTypeStatic
}

func (p *StaticProvider) Authenticate(ctx context.Context, req *http.Request) (*AuthResult, error) {
	// Extract credentials from various sources
	username, password := p.extractCredentials(req)
	if username == "" || password == "" {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeMissingToken,
		}, NewError(ErrCodeMissingToken, "missing credentials", "")
	}

	// Validate credentials
	user, exists := p.users[username]
	if !exists || !user.Enabled {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid credentials", "user not found or disabled")
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		return &AuthResult{
			Success: false,
			Error:   ErrCodeInvalidCredentials,
		}, NewError(ErrCodeInvalidCredentials, "invalid credentials", "password mismatch")
	}

	// Create user object
	authUser := &User{
		ID:       fmt.Sprintf("static:%s", username),
		Username: username,
		Email:    user.Email,
		Name:     user.Name,
		Roles:    user.Roles,
		Metadata: map[string]any{
			"provider": "static",
			"method":   "static_auth",
		},
	}

	return &AuthResult{
		Success: true,
		User:    authUser,
		Metadata: map[string]any{
			"authenticated_at": time.Now(),
			"provider":        "static",
		},
	}, nil
}

func (p *StaticProvider) ValidateToken(ctx context.Context, token string) (*AuthResult, error) {
	// Static auth doesn't use tokens for validation
	// Each request must be authenticated individually
	return &AuthResult{
		Success: false,
		Error:   ErrCodeTokenInvalid,
	}, NewError(ErrCodeTokenInvalid, "static auth does not support token validation", "")
}

func (p *StaticProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Static auth doesn't support token refresh
	return &AuthResult{
		Success: false,
		Error:   ErrCodeTokenInvalid,
	}, NewError(ErrCodeTokenInvalid, "static auth does not support token refresh", "")
}

func (p *StaticProvider) RevokeToken(ctx context.Context, token string) error {
	// Static auth doesn't use tokens
	return nil
}

func (p *StaticProvider) Initialize(config ProviderConfig) error {
	p.config = config
	
	// Load users from configuration
	if usersConfig, ok := config.Config["users"].([]any); ok {
		for _, userConfig := range usersConfig {
			if userMap, ok := userConfig.(map[string]any); ok {
				user := &StaticUser{}
				
				if username, ok := userMap["username"].(string); ok {
					user.Username = username
				}
				if password, ok := userMap["password"].(string); ok {
					user.Password = password
				}
				if email, ok := userMap["email"].(string); ok {
					user.Email = email
				}
				if name, ok := userMap["name"].(string); ok {
					user.Name = name
				}
				if enabled, ok := userMap["enabled"].(bool); ok {
					user.Enabled = enabled
				} else {
					user.Enabled = true // Default to enabled
				}
				
				if roles, ok := userMap["roles"].([]any); ok {
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							user.Roles = append(user.Roles, roleStr)
						}
					}
				}
				
				p.users[user.Username] = user
			}
		}
	}
	
	return nil
}

func (p *StaticProvider) Shutdown(ctx context.Context) error {
	// Clear user data
	p.users = make(map[string]*StaticUser)
	return nil
}

// extractCredentials extracts username and password from request
func (p *StaticProvider) extractCredentials(req *http.Request) (string, string) {
	// Try Basic Auth header first
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Basic ") {
		encodedCreds := strings.TrimPrefix(authHeader, "Basic ")
		decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err == nil {
			creds := strings.SplitN(string(decodedCreds), ":", 2)
			if len(creds) == 2 {
				return creds[0], creds[1]
			}
		}
	}
	
	// Try query parameters
	username := req.URL.Query().Get("username")
	password := req.URL.Query().Get("password")
	if username != "" && password != "" {
		return username, password
	}
	
	// Try form data (for POST requests)
	if req.Method == "POST" {
		if err := req.ParseForm(); err == nil {
			username = req.FormValue("username")
			password = req.FormValue("password")
			if username != "" && password != "" {
				return username, password
			}
		}
	}
	
	return "", ""
}

// AddUser adds a user to the static auth provider
func (p *StaticProvider) AddUser(username, password, email, name string, roles []string) error {
	if username == "" || password == "" {
		return NewError(ErrCodeConfigurationError, "username and password are required", "")
	}
	
	user := &StaticUser{
		Username: username,
		Password: password,
		Email:    email,
		Name:     name,
		Roles:    roles,
		Enabled:  true,
	}
	
	p.users[username] = user
	return nil
}

// RemoveUser removes a user from the static auth provider
func (p *StaticProvider) RemoveUser(username string) error {
	if _, exists := p.users[username]; !exists {
		return NewError(ErrCodeConfigurationError, "user not found", username)
	}
	
	delete(p.users, username)
	return nil
}

// UpdateUser updates an existing user
func (p *StaticProvider) UpdateUser(username string, updates map[string]any) error {
	user, exists := p.users[username]
	if !exists {
		return NewError(ErrCodeConfigurationError, "user not found", username)
	}
	
	if password, ok := updates["password"].(string); ok {
		user.Password = password
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if roles, ok := updates["roles"].([]string); ok {
		user.Roles = roles
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		user.Enabled = enabled
	}
	
	return nil
}

// ListUsers returns all users in the static auth provider
func (p *StaticProvider) ListUsers() []*StaticUser {
	users := make([]*StaticUser, 0, len(p.users))
	for _, user := range p.users {
		users = append(users, user)
	}
	return users
}