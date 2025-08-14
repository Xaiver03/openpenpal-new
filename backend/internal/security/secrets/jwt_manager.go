package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoSecretAvailable = errors.New("no JWT secret available")
	ErrInvalidSecretVersion = errors.New("invalid secret version")
)

// JWTSecret represents a versioned JWT secret
type JWTSecret struct {
	Version   string
	Secret    []byte
	CreatedAt time.Time
	ExpiresAt time.Time
	IsActive  bool
}

// JWTManager manages JWT secrets with rotation support
type JWTManager struct {
	mu            sync.RWMutex
	secrets       map[string]*JWTSecret
	activeVersion string
	rotationDays  int
}

// NewJWTManager creates a new JWT secret manager
func NewJWTManager(rotationDays int) *JWTManager {
	if rotationDays <= 0 {
		rotationDays = 90 // Default to 90 days rotation
	}
	
	return &JWTManager{
		secrets:      make(map[string]*JWTSecret),
		rotationDays: rotationDays,
	}
}

// Initialize loads or generates JWT secrets
func (m *JWTManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try to load from environment first
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		// For backward compatibility
		secret := &JWTSecret{
			Version:   "v1",
			Secret:    []byte(envSecret),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(m.rotationDays) * 24 * time.Hour),
			IsActive:  true,
		}
		m.secrets[secret.Version] = secret
		m.activeVersion = secret.Version
		return nil
	}

	// Generate a new secure secret
	secret, err := m.generateNewSecret()
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	m.secrets[secret.Version] = secret
	m.activeVersion = secret.Version
	
	return nil
}

// generateNewSecret creates a new cryptographically secure JWT secret
func (m *JWTManager) generateNewSecret() (*JWTSecret, error) {
	// Generate 32 bytes (256 bits) of random data
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	version := fmt.Sprintf("v%d", time.Now().Unix())
	
	return &JWTSecret{
		Version:   version,
		Secret:    randomBytes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(m.rotationDays) * 24 * time.Hour),
		IsActive:  true,
	}, nil
}

// GetActiveSecret returns the currently active JWT secret
func (m *JWTManager) GetActiveSecret() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeVersion == "" {
		return nil, ErrNoSecretAvailable
	}

	secret, exists := m.secrets[m.activeVersion]
	if !exists || !secret.IsActive {
		return nil, ErrNoSecretAvailable
	}

	return secret.Secret, nil
}

// GetSecretByVersion returns a specific version of the JWT secret
func (m *JWTManager) GetSecretByVersion(version string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret, exists := m.secrets[version]
	if !exists {
		return nil, ErrInvalidSecretVersion
	}

	return secret.Secret, nil
}

// RotateSecret creates a new secret and marks the old one for expiration
func (m *JWTManager) RotateSecret() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate new secret
	newSecret, err := m.generateNewSecret()
	if err != nil {
		return fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Mark old secret as inactive after grace period (1 hour)
	if oldSecret, exists := m.secrets[m.activeVersion]; exists {
		oldSecret.ExpiresAt = time.Now().Add(1 * time.Hour)
	}

	// Add new secret and make it active
	m.secrets[newSecret.Version] = newSecret
	m.activeVersion = newSecret.Version

	// Clean up expired secrets
	m.cleanupExpiredSecrets()

	return nil
}

// cleanupExpiredSecrets removes secrets that have expired
func (m *JWTManager) cleanupExpiredSecrets() {
	now := time.Now()
	for version, secret := range m.secrets {
		if secret.ExpiresAt.Before(now) && version != m.activeVersion {
			delete(m.secrets, version)
		}
	}
}

// ValidateToken validates a JWT token using appropriate secret version
func (m *JWTManager) ValidateToken(tokenString string) (*jwt.Token, error) {
	// Try to parse without validation first to get the version
	unverified, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := unverified.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Get the version from claims or use active version
	version := m.activeVersion
	if v, exists := claims["ver"].(string); exists {
		version = v
	}

	// Get the appropriate secret
	secret, err := m.GetSecretByVersion(version)
	if err != nil {
		// Fallback to active secret
		secret, err = m.GetActiveSecret()
		if err != nil {
			return nil, err
		}
	}

	// Parse and validate with the correct secret
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
}

// GetSecretString returns the active secret as a base64 encoded string
func (m *JWTManager) GetSecretString() (string, error) {
	secret, err := m.GetActiveSecret()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(secret), nil
}