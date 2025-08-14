package env

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// SecureConfig provides encrypted configuration management
type SecureConfig struct {
	encryptionKey []byte
	values        map[string]string
	encrypted     map[string]bool
}

// SensitiveKeys defines which configuration keys should be encrypted
var SensitiveKeys = []string{
	"JWT_SECRET",
	"DATABASE_PASSWORD",
	"DB_PASSWORD",
	"SMTP_PASSWORD",
	"EMAIL_API_KEY",
	"OPENAI_API_KEY",
	"CLAUDE_API_KEY",
	"SILICONFLOW_API_KEY",
	"MOONSHOT_API_KEY",
}

// NewSecureConfig creates a new secure configuration manager
func NewSecureConfig() (*SecureConfig, error) {
	// Generate or load encryption key
	key, err := getOrCreateEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	sc := &SecureConfig{
		encryptionKey: key,
		values:        make(map[string]string),
		encrypted:     make(map[string]bool),
	}

	// Load and decrypt existing values
	if err := sc.loadFromEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	return sc, nil
}

// Get retrieves a configuration value, decrypting if necessary
func (sc *SecureConfig) Get(key string) string {
	value, exists := sc.values[key]
	if !exists {
		return ""
	}

	// Decrypt if it's an encrypted value
	if sc.encrypted[key] {
		decrypted, err := sc.decrypt(value)
		if err != nil {
			// Log error but return empty string to avoid exposing encrypted data
			fmt.Fprintf(os.Stderr, "Failed to decrypt %s: %v\n", key, err)
			return ""
		}
		return decrypted
	}

	return value
}

// Set stores a configuration value, encrypting if it's sensitive
func (sc *SecureConfig) Set(key, value string) error {
	if sc.isSensitive(key) {
		encrypted, err := sc.encrypt(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt %s: %w", key, err)
		}
		sc.values[key] = encrypted
		sc.encrypted[key] = true
	} else {
		sc.values[key] = value
		sc.encrypted[key] = false
	}
	return nil
}

// GetOrDefault retrieves a value or returns a default
func (sc *SecureConfig) GetOrDefault(key, defaultValue string) string {
	if value := sc.Get(key); value != "" {
		return value
	}
	return defaultValue
}

// MustGet retrieves a value or panics if not found
func (sc *SecureConfig) MustGet(key string) string {
	value := sc.Get(key)
	if value == "" {
		panic(fmt.Sprintf("required configuration key %s not found", key))
	}
	return value
}

// ValidateRequired ensures all required keys are present
func (sc *SecureConfig) ValidateRequired(keys []string) error {
	var missing []string
	for _, key := range keys {
		if sc.Get(key) == "" {
			missing = append(missing, key)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration keys: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// loadFromEnvironment loads configuration from environment variables
func (sc *SecureConfig) loadFromEnvironment() error {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := parts[0]
		value := parts[1]
		
		// Check if value is already encrypted (prefixed with "enc:")
		if strings.HasPrefix(value, "enc:") {
			sc.values[key] = strings.TrimPrefix(value, "enc:")
			sc.encrypted[key] = true
		} else {
			if err := sc.Set(key, value); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// isSensitive checks if a key should be encrypted
func (sc *SecureConfig) isSensitive(key string) bool {
	for _, sensitive := range SensitiveKeys {
		if key == sensitive {
			return true
		}
	}
	return false
}

// encrypt encrypts a value using AES-GCM
func (sc *SecureConfig) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sc.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a value using AES-GCM
func (sc *SecureConfig) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sc.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// getOrCreateEncryptionKey gets or creates the master encryption key
func getOrCreateEncryptionKey() ([]byte, error) {
	// Try to get from environment first
	if keyStr := os.Getenv("CONFIG_ENCRYPTION_KEY"); keyStr != "" {
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid encryption key format: %w", err)
		}
		if len(key) != 32 {
			return nil, errors.New("encryption key must be 32 bytes")
		}
		return key, nil
	}

	// Try to load from file
	keyFile := ".encryption_key"
	if data, err := os.ReadFile(keyFile); err == nil {
		key, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}

	// Generate new key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save to file (development only)
	if os.Getenv("ENVIRONMENT") == "development" {
		keyStr := base64.StdEncoding.EncodeToString(key)
		if err := os.WriteFile(keyFile, []byte(keyStr), 0600); err != nil {
			// Non-fatal: just log the error
			fmt.Fprintf(os.Stderr, "Warning: failed to save encryption key: %v\n", err)
		}
	}

	return key, nil
}

// ExportSecure exports configuration with encrypted sensitive values
func (sc *SecureConfig) ExportSecure() map[string]string {
	export := make(map[string]string)
	
	for key, value := range sc.values {
		if sc.encrypted[key] {
			export[key] = "enc:" + value
		} else {
			export[key] = value
		}
	}
	
	return export
}

// GenerateSecureDefaults generates secure random values for sensitive configs
func GenerateSecureDefaults() map[string]string {
	defaults := make(map[string]string)
	
	// Generate secure JWT secret
	jwtSecret := make([]byte, 32)
	rand.Read(jwtSecret)
	defaults["JWT_SECRET"] = base64.StdEncoding.EncodeToString(jwtSecret)
	
	// Generate secure database password if not set
	if os.Getenv("DATABASE_PASSWORD") == "" {
		dbPass := make([]byte, 16)
		rand.Read(dbPass)
		defaults["DATABASE_PASSWORD"] = base64.StdEncoding.EncodeToString(dbPass)
	}
	
	return defaults
}