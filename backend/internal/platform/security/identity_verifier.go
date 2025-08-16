package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/config"
	"go.uber.org/zap"
)

// IdentityVerifier implements advanced identity verification and authentication
// Integrates with existing JWT authentication while adding zero-trust verification
type IdentityVerifier struct {
	config               *config.Config
	logger               *zap.Logger
	deviceFingerprinting *DeviceFingerprinting
	behaviorAnalyzer     *BehaviorAnalyzer
	biometricValidator   *BiometricValidator
	mfaValidator         *MFAValidator
	sessionTracker       *SessionTracker
	geoValidator         *GeoValidator
	mu                   sync.RWMutex
	running              bool
	trustedDevices       map[string]*TrustedDevice
	activeSessions       map[string]*UserSession
}

type DeviceFingerprinting struct {
	fingerprintDatabase map[string]*DeviceFingerprint
	mu                  sync.RWMutex
}

type DeviceFingerprint struct {
	DeviceID          string            `json:"device_id"`
	UserAgent         string            `json:"user_agent"`
	ScreenResolution  string            `json:"screen_resolution"`
	Timezone          string            `json:"timezone"`
	Language          string            `json:"language"`
	Platform          string            `json:"platform"`
	Plugins           []string          `json:"plugins"`
	Fonts             []string          `json:"fonts"`
	Canvas            string            `json:"canvas"`
	WebGL             string            `json:"webgl"`
	Hash              string            `json:"hash"`
	TrustScore        float64           `json:"trust_score"`
	LastSeen          time.Time         `json:"last_seen"`
	CreatedAt         time.Time         `json:"created_at"`
	Metadata          map[string]string `json:"metadata"`
}

type TrustedDevice struct {
	DeviceID      string            `json:"device_id"`
	UserID        string            `json:"user_id"`
	DeviceName    string            `json:"device_name"`
	Fingerprint   *DeviceFingerprint `json:"fingerprint"`
	TrustLevel    TrustLevel        `json:"trust_level"`
	LastUsed      time.Time         `json:"last_used"`
	CreatedAt     time.Time         `json:"created_at"`
	IsCompromised bool              `json:"is_compromised"`
	Metadata      map[string]string `json:"metadata"`
}

type UserSession struct {
	SessionID         string                `json:"session_id"`
	UserID            string                `json:"user_id"`
	DeviceID          string                `json:"device_id"`
	IPAddress         string                `json:"ip_address"`
	UserAgent         string                `json:"user_agent"`
	GeoLocation       *GeoLocation          `json:"geo_location"`
	AuthMethod        AuthenticationMethod  `json:"auth_method"`
	TrustLevel        TrustLevel            `json:"trust_level"`
	RiskScore         float64               `json:"risk_score"`
	CreatedAt         time.Time             `json:"created_at"`
	LastActivity      time.Time             `json:"last_activity"`
	ExpiresAt         time.Time             `json:"expires_at"`
	BehaviorProfile   *BehaviorProfile      `json:"behavior_profile"`
	SecurityFlags     []SecurityFlag        `json:"security_flags"`
	IsActive          bool                  `json:"is_active"`
	Metadata          map[string]interface{} `json:"metadata"`
}

type BehaviorProfile struct {
	TypingPattern     *TypingPattern    `json:"typing_pattern"`
	MousePattern      *MousePattern     `json:"mouse_pattern"`
	NavigationPattern *NavigationPattern `json:"navigation_pattern"`
	TimePattern       *TimePattern      `json:"time_pattern"`
	LastUpdated       time.Time         `json:"last_updated"`
}

type TypingPattern struct {
	AverageSpeed      float64   `json:"average_speed"`
	Rhythm            []float64 `json:"rhythm"`
	PauseDuration     []float64 `json:"pause_duration"`
	KeyPressure       []float64 `json:"key_pressure"`
	Confidence        float64   `json:"confidence"`
}

type MousePattern struct {
	MovementSpeed     float64     `json:"movement_speed"`
	ClickPattern      []float64   `json:"click_pattern"`
	ScrollBehavior    []float64   `json:"scroll_behavior"`
	Trajectory        [][]float64 `json:"trajectory"`
	Confidence        float64     `json:"confidence"`
}

type NavigationPattern struct {
	PageSequence      []string  `json:"page_sequence"`
	TimeOnPage        []float64 `json:"time_on_page"`
	InteractionDepth  float64   `json:"interaction_depth"`
	PreferredFeatures []string  `json:"preferred_features"`
	Confidence        float64   `json:"confidence"`
}

type TimePattern struct {
	ActiveHours       []int     `json:"active_hours"`
	SessionDuration   []float64 `json:"session_duration"`
	FrequencyPattern  []float64 `json:"frequency_pattern"`
	Confidence        float64   `json:"confidence"`
}

type SecurityFlag struct {
	Type        SecurityFlagType `json:"type"`
	Severity    Severity         `json:"severity"`
	Description string           `json:"description"`
	Timestamp   time.Time        `json:"timestamp"`
	Evidence    interface{}      `json:"evidence"`
}

type SecurityFlagType string

const (
	FlagSuspiciousIP        SecurityFlagType = "suspicious_ip"
	FlagAnomalousLocation   SecurityFlagType = "anomalous_location"
	FlagDeviceChange        SecurityFlagType = "device_change"
	FlagBehaviorAnomaly     SecurityFlagType = "behavior_anomaly"
	FlagRapidRequests       SecurityFlagType = "rapid_requests"
	FlagUnusualTimeAccess   SecurityFlagType = "unusual_time_access"
	FlagMultipleFailedAuth  SecurityFlagType = "multiple_failed_auth"
)

type BehaviorAnalyzer struct {
	userProfiles map[string]*UserBehaviorProfile
	mu           sync.RWMutex
}

type UserBehaviorProfile struct {
	UserID          string           `json:"user_id"`
	BaselineProfile *BehaviorProfile `json:"baseline_profile"`
	RecentActivity  []*ActivityPoint `json:"recent_activity"`
	AnomalyScore    float64          `json:"anomaly_score"`
	LastAnalyzed    time.Time        `json:"last_analyzed"`
	Confidence      float64          `json:"confidence"`
}

type ActivityPoint struct {
	Timestamp   time.Time              `json:"timestamp"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Duration    time.Duration          `json:"duration"`
	Context     map[string]interface{} `json:"context"`
	RiskScore   float64                `json:"risk_score"`
}

type BiometricValidator struct {
	enabled           bool
	biometricProfiles map[string]*BiometricProfile
	mu                sync.RWMutex
}

type BiometricProfile struct {
	UserID        string                 `json:"user_id"`
	Modalities    []BiometricModality    `json:"modalities"`
	Templates     map[string]interface{} `json:"templates"`
	Confidence    float64                `json:"confidence"`
	LastUpdated   time.Time              `json:"last_updated"`
	IsActive      bool                   `json:"is_active"`
}

type BiometricModality struct {
	Type       BiometricType `json:"type"`
	Quality    float64       `json:"quality"`
	Template   interface{}   `json:"template"`
	Confidence float64       `json:"confidence"`
}

type BiometricType string

const (
	BiometricFingerprint BiometricType = "fingerprint"
	BiometricFace        BiometricType = "face"
	BiometricVoice       BiometricType = "voice"
	BiometricBehavioral  BiometricType = "behavioral"
)

type MFAValidator struct {
	mfaProviders map[string]MFAProvider
	mu           sync.RWMutex
}

type MFAProvider interface {
	ValidateToken(userID string, token string) (*MFAResult, error)
	GenerateChallenge(userID string) (*MFAChallenge, error)
	GetMethods() []MFAMethod
}

type MFAResult struct {
	Valid       bool      `json:"valid"`
	Method      MFAMethod `json:"method"`
	Confidence  float64   `json:"confidence"`
	ExpiresAt   time.Time `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type MFAChallenge struct {
	ChallengeID string    `json:"challenge_id"`
	Method      MFAMethod `json:"method"`
	Challenge   string    `json:"challenge"`
	ExpiresAt   time.Time `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type MFAMethod string

const (
	MFAMethodTOTP     MFAMethod = "totp"
	MFAMethodSMS      MFAMethod = "sms"
	MFAMethodEmail    MFAMethod = "email"
	MFAMethodPush     MFAMethod = "push"
	MFAMethodWebAuthn MFAMethod = "webauthn"
	MFAMethodBiometric MFAMethod = "biometric"
)

func NewIdentityVerifier(cfg *config.Config, logger *zap.Logger) *IdentityVerifier {
	iv := &IdentityVerifier{
		config:         cfg,
		logger:         logger,
		trustedDevices: make(map[string]*TrustedDevice),
		activeSessions: make(map[string]*UserSession),
	}

	// Initialize sub-components
	iv.deviceFingerprinting = NewDeviceFingerprinting(logger)
	iv.behaviorAnalyzer = NewBehaviorAnalyzer(logger)
	iv.biometricValidator = NewBiometricValidator(logger)
	iv.mfaValidator = NewMFAValidator(logger)
	iv.sessionTracker = NewSessionTracker(logger)
	iv.geoValidator = NewGeoValidator(logger)

	return iv
}

func (iv *IdentityVerifier) Start(ctx context.Context) error {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	if iv.running {
		return fmt.Errorf("identity verifier already running")
	}

	iv.logger.Info("Starting Identity Verifier")
	iv.running = true

	return nil
}

func (iv *IdentityVerifier) Stop(ctx context.Context) error {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	if !iv.running {
		return nil
	}

	iv.logger.Info("Stopping Identity Verifier")
	iv.running = false

	return nil
}

func (iv *IdentityVerifier) VerifyIdentity(ctx context.Context, request *SecurityRequest) (*IdentityResult, error) {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	if !iv.running {
		return nil, fmt.Errorf("identity verifier not running")
	}

	result := &IdentityResult{
		TrustLevel: TrustLevelLow,
		Verified:   false,
		AuthMethod: AuthenticationMethod{
			Primary: "password",
			Factors: 1,
			Strength: 3,
		},
	}

	iv.logger.Info("Verifying identity",
		zap.String("user_id", request.UserID),
		zap.String("ip_address", request.IPAddress))

	// Step 1: Device fingerprinting
	deviceFingerprint, err := iv.deviceFingerprinting.GenerateFingerprint(request)
	if err != nil {
		iv.logger.Error("Device fingerprinting failed", zap.Error(err))
		return result, nil
	}

	// Step 2: Check trusted devices
	trustedDevice := iv.getTrustedDevice(request.UserID, deviceFingerprint.DeviceID)
	if trustedDevice != nil && !trustedDevice.IsCompromised {
		result.TrustLevel = trustedDevice.TrustLevel
		iv.logger.Info("Trusted device found", zap.String("device_id", trustedDevice.DeviceID))
	}

	// Step 3: Behavior analysis
	behaviorScore, err := iv.behaviorAnalyzer.AnalyzeBehavior(request.UserID, request)
	if err != nil {
		iv.logger.Error("Behavior analysis failed", zap.Error(err))
		behaviorScore = 0.5 // Default moderate risk
	}

	// Step 4: Geolocation validation
	geoRisk, err := iv.geoValidator.ValidateLocation(request.IPAddress, request.UserID)
	if err != nil {
		iv.logger.Error("Geo validation failed", zap.Error(err))
		geoRisk = 0.5 // Default moderate risk
	}

	// Step 5: Calculate overall trust level
	trustScore := iv.calculateTrustScore(deviceFingerprint, behaviorScore, geoRisk, trustedDevice)
	result.TrustLevel = iv.scoresToTrustLevel(trustScore)

	// Step 6: Biometric validation (if available)
	if iv.biometricValidator.enabled {
		biometricResult, err := iv.biometricValidator.ValidateBiometric(request.UserID, request)
		if err == nil && biometricResult.Valid {
			result.TrustLevel = TrustLevelHigh
			result.AuthMethod.Secondary = append(result.AuthMethod.Secondary, "biometric")
			result.AuthMethod.Factors++
		}
	}

	// Step 7: Update session tracking
	session := &UserSession{
		SessionID:    iv.generateSessionID(),
		UserID:       request.UserID,
		DeviceID:     deviceFingerprint.DeviceID,
		IPAddress:    request.IPAddress,
		UserAgent:    request.UserAgent,
		TrustLevel:   result.TrustLevel,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		IsActive:     true,
		Metadata:     make(map[string]interface{}),
	}

	iv.activeSessions[session.SessionID] = session
	result.Verified = result.TrustLevel >= TrustLevelMedium

	iv.logger.Info("Identity verification completed",
		zap.String("user_id", request.UserID),
		zap.String("trust_level", iv.trustLevelToString(result.TrustLevel)),
		zap.Bool("verified", result.Verified))

	return result, nil
}

func (iv *IdentityVerifier) calculateTrustScore(fingerprint *DeviceFingerprint, behaviorScore, geoRisk float64, trustedDevice *TrustedDevice) float64 {
	score := 0.3 // Base score

	// Device trust
	if trustedDevice != nil {
		score += 0.4
	} else {
		score += fingerprint.TrustScore * 0.3
	}

	// Behavior score (inverse of risk)
	score += (1.0 - behaviorScore) * 0.2

	// Geographic risk (inverse)
	score += (1.0 - geoRisk) * 0.1

	// Clamp to valid range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

func (iv *IdentityVerifier) scoresToTrustLevel(score float64) TrustLevel {
	if score >= 0.9 {
		return TrustLevelComplete
	} else if score >= 0.7 {
		return TrustLevelHigh
	} else if score >= 0.5 {
		return TrustLevelMedium
	} else if score >= 0.3 {
		return TrustLevelLow
	}
	return TrustLevelDeny
}

func (iv *IdentityVerifier) trustLevelToString(level TrustLevel) string {
	switch level {
	case TrustLevelComplete:
		return "complete"
	case TrustLevelHigh:
		return "high"
	case TrustLevelMedium:
		return "medium"
	case TrustLevelLow:
		return "low"
	case TrustLevelDeny:
		return "deny"
	default:
		return "unknown"
	}
}

func (iv *IdentityVerifier) getTrustedDevice(userID, deviceID string) *TrustedDevice {
	key := fmt.Sprintf("%s:%s", userID, deviceID)
	return iv.trustedDevices[key]
}

func (iv *IdentityVerifier) generateSessionID() string {
	return fmt.Sprintf("sess_%d_%s", time.Now().Unix(), generateRandomString(16))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Stub implementations for sub-components
func NewDeviceFingerprinting(logger *zap.Logger) *DeviceFingerprinting {
	return &DeviceFingerprinting{
		fingerprintDatabase: make(map[string]*DeviceFingerprint),
	}
}

func (df *DeviceFingerprinting) GenerateFingerprint(request *SecurityRequest) (*DeviceFingerprint, error) {
	// Create a simplified fingerprint based on request data
	hash := sha256.Sum256([]byte(request.UserAgent + request.IPAddress))
	
	fingerprint := &DeviceFingerprint{
		DeviceID:     hex.EncodeToString(hash[:])[:16],
		UserAgent:    request.UserAgent,
		Platform:     extractPlatform(request.UserAgent),
		Hash:         hex.EncodeToString(hash[:]),
		TrustScore:   0.5, // Default medium trust
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]string),
	}

	return fingerprint, nil
}

func extractPlatform(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "windows") {
		return "windows"
	} else if strings.Contains(userAgent, "mac") {
		return "macos"
	} else if strings.Contains(userAgent, "linux") {
		return "linux"
	} else if strings.Contains(userAgent, "android") {
		return "android"
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "ios"
	}
	return "unknown"
}

func NewBehaviorAnalyzer(logger *zap.Logger) *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		userProfiles: make(map[string]*UserBehaviorProfile),
	}
}

func (ba *BehaviorAnalyzer) AnalyzeBehavior(userID string, request *SecurityRequest) (float64, error) {
	// Simplified behavior analysis
	// In production, this would analyze actual behavioral patterns
	return 0.3, nil // Default low risk
}

func NewBiometricValidator(logger *zap.Logger) *BiometricValidator {
	return &BiometricValidator{
		enabled:           false, // Disabled by default
		biometricProfiles: make(map[string]*BiometricProfile),
	}
}

func (bv *BiometricValidator) ValidateBiometric(userID string, request *SecurityRequest) (*BiometricResult, error) {
	return &BiometricResult{Valid: false}, nil
}

type BiometricResult struct {
	Valid      bool    `json:"valid"`
	Confidence float64 `json:"confidence"`
	Modality   BiometricType `json:"modality"`
}

func NewMFAValidator(logger *zap.Logger) *MFAValidator {
	return &MFAValidator{
		mfaProviders: make(map[string]MFAProvider),
	}
}

func NewSessionTracker(logger *zap.Logger) *SessionTracker {
	return &SessionTracker{}
}

type SessionTracker struct{}

func NewGeoValidator(logger *zap.Logger) *GeoValidator {
	return &GeoValidator{}
}

type GeoValidator struct{}

func (gv *GeoValidator) ValidateLocation(ipAddress, userID string) (float64, error) {
	// Simplified geo validation
	// Check if IP is local/private
	ip := net.ParseIP(ipAddress)
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		return 0.1, nil // Low risk for local IPs
	}
	
	// For demo purposes, return medium risk for external IPs
	return 0.4, nil
}