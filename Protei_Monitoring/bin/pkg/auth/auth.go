package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication and authorization
type Service struct {
	config     *Config
	jwtSecret  []byte
	users      map[string]*User
	sessions   map[string]*Session
}

// Config holds authentication configuration
type Config struct {
	JWTSecret       string
	TokenExpiry     time.Duration
	PasswordMinLen  int
	LDAPEnabled     bool
	LDAPServer      string
	LDAPBaseDN      string
	AllowLocalAuth  bool
}

// User represents an authenticated user
type User struct {
	ID          int
	Username    string
	PasswordHash string
	FullName    string
	Email       string
	Role        Role
	Permissions map[string]bool
	Enabled     bool
	LastLogin   time.Time
}

// Session represents an active user session
type Session struct {
	Token      string
	Username   string
	Role       Role
	CreatedAt  time.Time
	ExpiresAt  time.Time
	IP         string
}

// Role represents user roles
type Role string

const (
	RoleAdmin         Role = "admin"
	RoleEngineer      Role = "engineer"
	RoleNOCViewer     Role = "noc_viewer"
	RoleSecurityAudit Role = "security_auditor"
)

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	Role     Role   `json:"role"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user account disabled")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrPermissionDenied   = errors.New("permission denied")
)

// NewService creates a new authentication service
func NewService(config *Config) *Service {
	return &Service{
		config:    config,
		jwtSecret: []byte(config.JWTSecret),
		users:     make(map[string]*User),
		sessions:  make(map[string]*Session),
	}
}

// Authenticate authenticates a user with username and password
func (s *Service) Authenticate(username, password, ip string) (*Session, error) {
	// Try local authentication first
	if s.config.AllowLocalAuth {
		user, err := s.authenticateLocal(username, password)
		if err == nil {
			return s.createSession(user, ip)
		}
	}

	// Try LDAP if enabled
	if s.config.LDAPEnabled {
		user, err := s.authenticateLDAP(username, password)
		if err == nil {
			return s.createSession(user, ip)
		}
	}

	return nil, ErrInvalidCredentials
}

// authenticateLocal authenticates against local user database
func (s *Service) authenticateLocal(username, password string) (*User, error) {
	user, ok := s.users[username]
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if !user.Enabled {
		return nil, ErrUserDisabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	user.LastLogin = time.Now()
	return user, nil
}

// authenticateLDAP authenticates against LDAP server (placeholder)
func (s *Service) authenticateLDAP(username, password string) (*User, error) {
	// LDAP authentication would go here
	// For now, return not implemented
	return nil, fmt.Errorf("LDAP authentication not implemented")
}

// createSession creates a new session with JWT token
func (s *Service) createSession(user *User, ip string) (*Session, error) {
	// Generate JWT token
	expiresAt := time.Now().Add(s.config.TokenExpiry)

	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	session := &Session{
		Token:     tokenString,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IP:        ip,
	}

	s.sessions[tokenString] = session

	return session, nil
}

// ValidateToken validates a JWT token and returns the session
func (s *Service) ValidateToken(tokenString string) (*Session, error) {
	// Check session cache first
	if session, ok := s.sessions[tokenString]; ok {
		if time.Now().After(session.ExpiresAt) {
			delete(s.sessions, tokenString)
			return nil, ErrTokenExpired
		}
		return session, nil
	}

	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		session := &Session{
			Token:     tokenString,
			Username:  claims.Username,
			Role:      claims.Role,
			ExpiresAt: claims.ExpiresAt.Time,
		}
		s.sessions[tokenString] = session
		return session, nil
	}

	return nil, ErrInvalidToken
}

// CheckPermission checks if a user has permission for an action
func (s *Service) CheckPermission(session *Session, permission string) error {
	user, ok := s.users[session.Username]
	if !ok {
		return ErrPermissionDenied
	}

	// Admin has all permissions
	if user.Role == RoleAdmin {
		return nil
	}

	// Check specific permission
	if allowed, ok := user.Permissions[permission]; ok && allowed {
		return nil
	}

	// Check role-based permissions
	if s.checkRolePermission(user.Role, permission) {
		return nil
	}

	return ErrPermissionDenied
}

// checkRolePermission checks role-based permissions
func (s *Service) checkRolePermission(role Role, permission string) bool {
	rolePermissions := map[Role][]string{
		RoleEngineer: {
			"view_sessions",
			"view_kpi",
			"view_traces",
			"download_pcap",
			"create_filters",
		},
		RoleNOCViewer: {
			"view_sessions",
			"view_kpi",
			"view_dashboard",
		},
		RoleSecurityAudit: {
			"view_audit_log",
			"view_alarms",
		},
	}

	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}

	return false
}

// Logout invalidates a session
func (s *Service) Logout(token string) {
	delete(s.sessions, token)
}

// RegisterUser registers a new user
func (s *Service) RegisterUser(user *User) error {
	if _, exists := s.users[user.Username]; exists {
		return fmt.Errorf("user already exists")
	}

	s.users[user.Username] = user
	return nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// GenerateAPIKey generates a random API key
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
