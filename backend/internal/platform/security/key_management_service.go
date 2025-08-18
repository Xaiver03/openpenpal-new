package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"
)

// ComprehensiveKeyManagementService implements enterprise-grade key management
type ComprehensiveKeyManagementService struct {
	config          *KeyManagementConfig
	keyStore        *KeyStore
	hsm             *HSMInterface
	policyEngine    *KeyPolicyEngine
	auditLogger     *KeyAuditLogger
	rotationManager *KeyRotationManager
	accessControl   *KeyAccessControl
	backupManager   *KeyBackupManager
	mutex           sync.RWMutex
}

// KeyManagementConfig defines key management parameters
type KeyManagementConfig struct {
	// Storage settings
	DefaultKeyStore      string        `json:"default_key_store"`
	EnableHSMStorage     bool          `json:"enable_hsm_storage"`
	EnableKeyBackup      bool          `json:"enable_key_backup"`
	
	// Key lifecycle
	DefaultKeyRotation   time.Duration `json:"default_key_rotation"`
	KeyRetentionPeriod   time.Duration `json:"key_retention_period"`
	AutoRotationEnabled  bool          `json:"auto_rotation_enabled"`
	
	// Security settings
	RequireKeyEscrow     bool          `json:"require_key_escrow"`
	EnableSplitKnowledge bool          `json:"enable_split_knowledge"`
	MinimumKeyStrength   int           `json:"minimum_key_strength"`
	
	// Compliance settings
	ComplianceMode       string        `json:"compliance_mode"`
	AuditRetentionPeriod time.Duration `json:"audit_retention_period"`
	EnableKeyProvenance  bool          `json:"enable_key_provenance"`
	
	// Performance settings
	KeyCacheEnabled      bool          `json:"key_cache_enabled"`
	KeyCacheSize         int           `json:"key_cache_size"`
	KeyCacheTTL          time.Duration `json:"key_cache_ttl"`
}

// NewComprehensiveKeyManagementService creates a new key management service
func NewComprehensiveKeyManagementService(config *KeyManagementConfig) *ComprehensiveKeyManagementService {
	if config == nil {
		config = getDefaultKeyManagementConfig()
	}

	return &ComprehensiveKeyManagementService{
		config:          config,
		keyStore:        NewKeyStore(config),
		hsm:             NewHSMInterface(nil),
		policyEngine:    NewKeyPolicyEngine(),
		auditLogger:     NewKeyAuditLogger(),
		rotationManager: NewKeyRotationManager(config),
		accessControl:   NewKeyAccessControl(),
		backupManager:   NewKeyBackupManager(),
	}
}

// GenerateKey generates a new cryptographic key
func (k *ComprehensiveKeyManagementService) GenerateKey(ctx context.Context, spec *KeySpec) (*Key, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// Validate key specification
	if err := k.validateKeySpec(spec); err != nil {
		k.auditLogger.LogKeyEvent(ctx, "key_generation_failed", map[string]interface{}{
			"error": err.Error(),
			"spec":  spec,
		})
		return nil, fmt.Errorf("invalid key specification: %w", err)
	}

	// Check policy compliance
	if allowed, err := k.policyEngine.CheckKeyGenerationPolicy(ctx, spec); !allowed || err != nil {
		return nil, fmt.Errorf("key generation policy violation: %w", err)
	}

	// Generate key based on type
	var keyData []byte
	var err error

	switch spec.Type {
	case KeyTypeSymmetric:
		keyData, err = k.generateSymmetricKey(spec.Size)
	case KeyTypeAsymmetric:
		keyData, err = k.generateAsymmetricKey(spec.Size)
	case KeyTypeHMAC:
		keyData, err = k.generateHMACKey(spec.Size)
	case KeyTypeDerivation:
		keyData, err = k.generateDerivationKey(spec.Size)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", spec.Type)
	}

	if err != nil {
		k.auditLogger.LogKeyEvent(ctx, "key_generation_failed", map[string]interface{}{
			"error": err.Error(),
			"type":  spec.Type,
			"size":  spec.Size,
		})
		return nil, fmt.Errorf("key generation failed: %w", err)
	}

	// Create key metadata
	key := &Key{
		ID:        generateKeyID(),
		Type:      spec.Type,
		Algorithm: spec.Algorithm,
		Size:      spec.Size,
		Usage:     spec.Usage,
		Status:    KeyStatusActive,
		CreatedAt: time.Now(),
		KeyData:   keyData,
		Metadata: map[string]interface{}{
			"generated_by": extractUserFromContext(ctx),
			"purpose":      spec.Purpose,
		},
	}

	// Set rotation policy if specified
	if spec.RotationPolicy != nil {
		key.RotationPolicy = spec.RotationPolicy
		key.ExpiresAt = &[]time.Time{time.Now().Add(spec.RotationPolicy.RotationInterval)}[0]
	}

	// Store key
	if err := k.storeKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	// Create backup if enabled
	if k.config.EnableKeyBackup {
		if err := k.backupManager.BackupKey(ctx, key); err != nil {
			k.auditLogger.LogKeyEvent(ctx, "key_backup_failed", map[string]interface{}{
				"key_id": key.ID,
				"error":  err.Error(),
			})
		}
	}

	// Schedule rotation if auto-rotation is enabled
	if k.config.AutoRotationEnabled && key.RotationPolicy != nil {
		k.rotationManager.ScheduleRotation(ctx, key.ID, key.RotationPolicy)
	}

	// Audit logging
	k.auditLogger.LogKeyEvent(ctx, "key_generated", map[string]interface{}{
		"key_id":    key.ID,
		"type":      key.Type,
		"algorithm": key.Algorithm,
		"size":      key.Size,
		"usage":     key.Usage,
	})

	return key, nil
}

// ImportKey imports an existing key
func (k *ComprehensiveKeyManagementService) ImportKey(ctx context.Context, keyData []byte, metadata *KeyMetadata) (*Key, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// Validate imported key
	if err := k.validateImportedKey(keyData, metadata); err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	// Check import policy
	if allowed, err := k.policyEngine.CheckKeyImportPolicy(ctx, metadata); !allowed || err != nil {
		return nil, fmt.Errorf("key import policy violation: %w", err)
	}

	// Create key object
	key := &Key{
		ID:        generateKeyID(),
		Type:      metadata.Type,
		Algorithm: metadata.Algorithm,
		Size:      metadata.Size,
		Usage:     metadata.Usage,
		Status:    KeyStatusActive,
		CreatedAt: time.Now(),
		KeyData:   keyData,
		Metadata: map[string]interface{}{
			"imported_by": extractUserFromContext(ctx),
			"source":      metadata.Source,
		},
	}

	// Store imported key
	if err := k.storeKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to store imported key: %w", err)
	}

	// Audit logging
	k.auditLogger.LogKeyEvent(ctx, "key_imported", map[string]interface{}{
		"key_id":    key.ID,
		"type":      key.Type,
		"algorithm": key.Algorithm,
		"source":    metadata.Source,
	})

	return key, nil
}

// ExportKey exports a key in specified format
func (k *ComprehensiveKeyManagementService) ExportKey(ctx context.Context, keyID string, format KeyFormat) ([]byte, error) {
	// Check export policy
	if allowed, err := k.policyEngine.CheckKeyExportPolicy(ctx, keyID); !allowed || err != nil {
		return nil, fmt.Errorf("key export policy violation: %w", err)
	}

	// Get key
	key, err := k.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Export in specified format
	var exportData []byte
	switch format {
	case KeyFormatPEM:
		exportData, err = k.exportKeyAsPEM(key)
	case KeyFormatJWK:
		exportData, err = k.exportKeyAsJWK(key)
	case KeyFormatPKCS12:
		exportData, err = k.exportKeyAsPKCS12(key)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("key export failed: %w", err)
	}

	// Audit logging
	k.auditLogger.LogKeyEvent(ctx, "key_exported", map[string]interface{}{
		"key_id": keyID,
		"format": format,
	})

	return exportData, nil
}

// DeleteKey deletes a key
func (k *ComprehensiveKeyManagementService) DeleteKey(ctx context.Context, keyID string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// Check deletion policy
	if allowed, err := k.policyEngine.CheckKeyDeletionPolicy(ctx, keyID); !allowed || err != nil {
		return fmt.Errorf("key deletion policy violation: %w", err)
	}

	// Get key for audit purposes
	key, err := k.GetKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get key for deletion: %w", err)
	}

	// Mark key as deleted (soft delete)
	key.Status = KeyStatusDeleted
	key.DeletedAt = &[]time.Time{time.Now()}[0]

	// Update key in store
	if err := k.keyStore.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("failed to mark key as deleted: %w", err)
	}

	// Schedule actual deletion after retention period
	k.scheduleKeyPurge(ctx, keyID, k.config.KeyRetentionPeriod)

	// Audit logging
	k.auditLogger.LogKeyEvent(ctx, "key_deleted", map[string]interface{}{
		"key_id":     keyID,
		"deleted_by": extractUserFromContext(ctx),
	})

	return nil
}

// RotateKey rotates an existing key
func (k *ComprehensiveKeyManagementService) RotateKey(ctx context.Context, keyID string) (*Key, error) {
	return k.rotationManager.RotateKey(ctx, keyID, k)
}

// ScheduleRotation schedules automatic key rotation
func (k *ComprehensiveKeyManagementService) ScheduleRotation(ctx context.Context, keyID string, policy *RotationPolicy) error {
	return k.rotationManager.ScheduleRotation(ctx, keyID, policy)
}

// GetRotationStatus gets the rotation status of a key
func (k *ComprehensiveKeyManagementService) GetRotationStatus(ctx context.Context, keyID string) (*RotationStatus, error) {
	return k.rotationManager.GetRotationStatus(ctx, keyID)
}

// UseKey uses a key for a cryptographic operation
func (k *ComprehensiveKeyManagementService) UseKey(ctx context.Context, keyID string, operation KeyOperation, data []byte) ([]byte, error) {
	// Check key usage policy
	if allowed, err := k.accessControl.CheckKeyUsage(ctx, keyID, operation); !allowed || err != nil {
		return nil, fmt.Errorf("key usage denied: %w", err)
	}

	// Get key
	key, err := k.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Perform operation based on type
	var result []byte
	switch operation {
	case KeyOperationEncrypt:
		result, err = k.performEncryption(key, data)
	case KeyOperationDecrypt:
		result, err = k.performDecryption(key, data)
	case KeyOperationSign:
		result, err = k.performSigning(key, data)
	case KeyOperationVerify:
		result, err = k.performVerification(key, data)
	default:
		return nil, fmt.Errorf("unsupported key operation: %s", operation)
	}

	if err != nil {
		return nil, fmt.Errorf("key operation failed: %w", err)
	}

	// Update key usage statistics
	k.updateKeyUsageStats(ctx, keyID, operation)

	// Audit logging
	k.auditLogger.LogKeyEvent(ctx, "key_used", map[string]interface{}{
		"key_id":    keyID,
		"operation": operation,
		"data_size": len(data),
	})

	return result, nil
}

// AuthorizeKeyAccess authorizes access to a key
func (k *ComprehensiveKeyManagementService) AuthorizeKeyAccess(ctx context.Context, keyID string, identity *Identity) (*KeyAccessResult, error) {
	return k.accessControl.AuthorizeAccess(ctx, keyID, identity)
}

// AuditKeyUsage audits key usage within a time range
func (k *ComprehensiveKeyManagementService) AuditKeyUsage(ctx context.Context, keyID string, timeRange *TimeRange) ([]*KeyUsageEvent, error) {
	return k.auditLogger.GetKeyUsageEvents(ctx, keyID, timeRange)
}

// GetKey retrieves a key by ID
func (k *ComprehensiveKeyManagementService) GetKey(ctx context.Context, keyID string) (*Key, error) {
	return k.keyStore.GetKey(ctx, keyID)
}

// UpdateKey updates key metadata
func (k *ComprehensiveKeyManagementService) UpdateKey(ctx context.Context, key *Key) error {
	return k.keyStore.UpdateKey(ctx, key)
}

// Private helper methods

func (k *ComprehensiveKeyManagementService) validateKeySpec(spec *KeySpec) error {
	if spec.Size < k.config.MinimumKeyStrength {
		return fmt.Errorf("key size %d below minimum strength %d", spec.Size, k.config.MinimumKeyStrength)
	}

	validUsages := []KeyUsage{KeyUsageEncrypt, KeyUsageDecrypt, KeyUsageSign, KeyUsageVerify}
	for _, usage := range spec.Usage {
		valid := false
		for _, validUsage := range validUsages {
			if usage == validUsage {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid key usage: %s", usage)
		}
	}

	return nil
}

func (k *ComprehensiveKeyManagementService) validateImportedKey(keyData []byte, metadata *KeyMetadata) error {
	if len(keyData) == 0 {
		return fmt.Errorf("empty key data")
	}

	if metadata.Size < k.config.MinimumKeyStrength {
		return fmt.Errorf("imported key size below minimum strength")
	}

	return nil
}

func (k *ComprehensiveKeyManagementService) generateSymmetricKey(size int) ([]byte, error) {
	keySize := size / 8 // Convert bits to bytes
	key := make([]byte, keySize)
	_, err := rand.Read(key)
	return key, err
}

func (k *ComprehensiveKeyManagementService) generateAsymmetricKey(size int) ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return nil, err
	}

	// Convert to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return pem.EncodeToMemory(privateKeyPEM), nil
}

func (k *ComprehensiveKeyManagementService) generateHMACKey(size int) ([]byte, error) {
	return k.generateSymmetricKey(size)
}

func (k *ComprehensiveKeyManagementService) generateDerivationKey(size int) ([]byte, error) {
	return k.generateSymmetricKey(size)
}

func (k *ComprehensiveKeyManagementService) storeKey(ctx context.Context, key *Key) error {
	// Store in HSM if enabled
	if k.config.EnableHSMStorage {
		if err := k.hsm.StoreKey(ctx, key); err != nil {
			return fmt.Errorf("HSM storage failed: %w", err)
		}
	}

	// Store in regular key store
	return k.keyStore.StoreKey(ctx, key)
}

func (k *ComprehensiveKeyManagementService) exportKeyAsPEM(key *Key) ([]byte, error) {
	// For demonstration, return the key data as-is
	// In production, this would format properly as PEM
	return key.KeyData, nil
}

func (k *ComprehensiveKeyManagementService) exportKeyAsJWK(key *Key) ([]byte, error) {
	// For demonstration, return JSON representation
	// In production, this would format as proper JWK
	return []byte(fmt.Sprintf(`{"kty":"RSA","kid":"%s"}`, key.ID)), nil
}

func (k *ComprehensiveKeyManagementService) exportKeyAsPKCS12(key *Key) ([]byte, error) {
	// For demonstration, return the key data
	// In production, this would format as PKCS#12
	return key.KeyData, nil
}

func (k *ComprehensiveKeyManagementService) performEncryption(key *Key, data []byte) ([]byte, error) {
	// Simplified encryption - in production would use proper crypto
	return append([]byte("encrypted:"), data...), nil
}

func (k *ComprehensiveKeyManagementService) performDecryption(key *Key, data []byte) ([]byte, error) {
	// Simplified decryption - remove prefix
	if len(data) > 10 && string(data[:10]) == "encrypted:" {
		return data[10:], nil
	}
	return nil, fmt.Errorf("invalid encrypted data")
}

func (k *ComprehensiveKeyManagementService) performSigning(key *Key, data []byte) ([]byte, error) {
	// Simplified signing
	return append([]byte("signature:"), data...), nil
}

func (k *ComprehensiveKeyManagementService) performVerification(key *Key, data []byte) ([]byte, error) {
	// Simplified verification
	return []byte("verified"), nil
}

func (k *ComprehensiveKeyManagementService) updateKeyUsageStats(ctx context.Context, keyID string, operation KeyOperation) {
	// Update usage statistics
	k.auditLogger.RecordKeyUsage(ctx, keyID, operation)
}

func (k *ComprehensiveKeyManagementService) scheduleKeyPurge(ctx context.Context, keyID string, delay time.Duration) {
	// Schedule key purge - simplified implementation
	go func() {
		time.Sleep(delay)
		k.keyStore.PurgeKey(ctx, keyID)
	}()
}

// Supporting types and configurations

// KeySpec defines key generation parameters
type KeySpec struct {
	Type           KeyType          `json:"type"`
	Algorithm      CryptoAlgorithm  `json:"algorithm"`
	Size           int              `json:"size"`
	Usage          []KeyUsage       `json:"usage"`
	Purpose        string           `json:"purpose"`
	RotationPolicy *RotationPolicy  `json:"rotation_policy,omitempty"`
}

// KeyMetadata contains key metadata for import/export
type KeyMetadata struct {
	Type      KeyType         `json:"type"`
	Algorithm CryptoAlgorithm `json:"algorithm"`
	Size      int             `json:"size"`
	Usage     []KeyUsage      `json:"usage"`
	Source    string          `json:"source"`
}

// KeyFormat represents key export formats
type KeyFormat string

const (
	KeyFormatPEM    KeyFormat = "pem"
	KeyFormatJWK    KeyFormat = "jwk"
	KeyFormatPKCS12 KeyFormat = "pkcs12"
)

// KeyOperation represents cryptographic operations
type KeyOperation string

const (
	KeyOperationEncrypt KeyOperation = "encrypt"
	KeyOperationDecrypt KeyOperation = "decrypt"
	KeyOperationSign    KeyOperation = "sign"
	KeyOperationVerify  KeyOperation = "verify"
)

// KeyUsage represents key usage types
type KeyUsage string

const (
	KeyUsageEncrypt KeyUsage = "encrypt"
	KeyUsageDecrypt KeyUsage = "decrypt"
	KeyUsageSign    KeyUsage = "sign"
	KeyUsageVerify  KeyUsage = "verify"
)

// RotationStatus represents key rotation status
type RotationStatus struct {
	KeyID           string    `json:"key_id"`
	NextRotation    time.Time `json:"next_rotation"`
	LastRotation    time.Time `json:"last_rotation"`
	RotationCount   int       `json:"rotation_count"`
	IsScheduled     bool      `json:"is_scheduled"`
}

// KeyAccessResult represents key access authorization result
type KeyAccessResult struct {
	Authorized  bool                   `json:"authorized"`
	Reason      string                 `json:"reason"`
	Permissions []string               `json:"permissions"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TimeRange represents a time range for auditing
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// KeyUsageEvent represents a key usage audit event
type KeyUsageEvent struct {
	KeyID       string       `json:"key_id"`
	Operation   KeyOperation `json:"operation"`
	UserID      string       `json:"user_id"`
	Timestamp   time.Time    `json:"timestamp"`
	Success     bool         `json:"success"`
	Error       string       `json:"error,omitempty"`
}

func getDefaultKeyManagementConfig() *KeyManagementConfig {
	return &KeyManagementConfig{
		DefaultKeyStore:      "database",
		EnableHSMStorage:     false,
		EnableKeyBackup:      true,
		DefaultKeyRotation:   time.Hour * 24 * 30, // 30 days
		KeyRetentionPeriod:   time.Hour * 24 * 365, // 1 year
		AutoRotationEnabled:  true,
		RequireKeyEscrow:     false,
		EnableSplitKnowledge: false,
		MinimumKeyStrength:   256,
		ComplianceMode:       "GDPR",
		AuditRetentionPeriod: time.Hour * 24 * 2555, // 7 years
		EnableKeyProvenance:  true,
		KeyCacheEnabled:      true,
		KeyCacheSize:         1000,
		KeyCacheTTL:          time.Hour,
	}
}

func generateKeyID() string {
	return fmt.Sprintf("key-%d", time.Now().UnixNano())
}

// Placeholder implementations for supporting components
type KeyStore struct{}
func NewKeyStore(config *KeyManagementConfig) *KeyStore { return &KeyStore{} }
func (k *KeyStore) StoreKey(ctx context.Context, key *Key) error { return nil }
func (k *KeyStore) GetKey(ctx context.Context, keyID string) (*Key, error) { return &Key{ID: keyID}, nil }
func (k *KeyStore) UpdateKey(ctx context.Context, key *Key) error { return nil }
func (k *KeyStore) PurgeKey(ctx context.Context, keyID string) error { return nil }

type KeyPolicyEngine struct{}
func NewKeyPolicyEngine() *KeyPolicyEngine { return &KeyPolicyEngine{} }
func (k *KeyPolicyEngine) CheckKeyGenerationPolicy(ctx context.Context, spec *KeySpec) (bool, error) { return true, nil }
func (k *KeyPolicyEngine) CheckKeyImportPolicy(ctx context.Context, metadata *KeyMetadata) (bool, error) { return true, nil }
func (k *KeyPolicyEngine) CheckKeyExportPolicy(ctx context.Context, keyID string) (bool, error) { return true, nil }
func (k *KeyPolicyEngine) CheckKeyDeletionPolicy(ctx context.Context, keyID string) (bool, error) { return true, nil }

type KeyAuditLogger struct{}
func NewKeyAuditLogger() *KeyAuditLogger { return &KeyAuditLogger{} }
func (k *KeyAuditLogger) LogKeyEvent(ctx context.Context, event string, data map[string]interface{}) {}
func (k *KeyAuditLogger) RecordKeyUsage(ctx context.Context, keyID string, operation KeyOperation) {}
func (k *KeyAuditLogger) GetKeyUsageEvents(ctx context.Context, keyID string, timeRange *TimeRange) ([]*KeyUsageEvent, error) { return []*KeyUsageEvent{}, nil }

type KeyRotationManager struct{}
func NewKeyRotationManager(config *KeyManagementConfig) *KeyRotationManager { return &KeyRotationManager{} }
func (k *KeyRotationManager) RotateKey(ctx context.Context, keyID string, kms *ComprehensiveKeyManagementService) (*Key, error) { return &Key{}, nil }
func (k *KeyRotationManager) ScheduleRotation(ctx context.Context, keyID string, policy *RotationPolicy) error { return nil }
func (k *KeyRotationManager) GetRotationStatus(ctx context.Context, keyID string) (*RotationStatus, error) { return &RotationStatus{}, nil }

type KeyAccessControl struct{}
func NewKeyAccessControl() *KeyAccessControl { return &KeyAccessControl{} }
func (k *KeyAccessControl) CheckKeyUsage(ctx context.Context, keyID string, operation KeyOperation) (bool, error) { return true, nil }
func (k *KeyAccessControl) AuthorizeAccess(ctx context.Context, keyID string, identity *Identity) (*KeyAccessResult, error) { return &KeyAccessResult{Authorized: true}, nil }

type KeyBackupManager struct{}
func NewKeyBackupManager() *KeyBackupManager { return &KeyBackupManager{} }
func (k *KeyBackupManager) BackupKey(ctx context.Context, key *Key) error { return nil }

func (h *HSMInterface) StoreKey(ctx context.Context, key *Key) error { return nil }