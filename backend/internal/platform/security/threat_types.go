package security

import (
	"fmt"
	"time"
)

// Enhanced types for AI-driven threat detection (Phase 4.3)

// ThreatAssessment represents the result of threat analysis
type ThreatAssessment struct {
	EventID      string                 `json:"event_id"`
	Timestamp    time.Time              `json:"timestamp"`
	ThreatScore  float64                `json:"threat_score"`
	Confidence   float64                `json:"confidence"`
	RiskLevel    RiskLevel              `json:"risk_level"`
	Indicators   []string               `json:"indicators"`
	Mitigations  []string               `json:"mitigations"`
	Evidence     []*Evidence            `json:"evidence"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// RiskLevel represents the assessed risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ThreatCategory represents different categories of threats
type ThreatCategory string

const (
	ThreatCategoryMaliciousActivity ThreatCategory = "malicious_activity"
	ThreatCategoryAnomalousAccess   ThreatCategory = "anomalous_access"
	ThreatCategoryDataBreach        ThreatCategory = "data_breach"
	ThreatCategorySystemIntrusion   ThreatCategory = "system_intrusion"
	ThreatCategoryPrivilegeAbuse    ThreatCategory = "privilege_abuse"
	ThreatCategoryNetworkAttack     ThreatCategory = "network_attack"
)

// ThreatStatus represents the current status of a threat
type ThreatStatus string

const (
	ThreatStatusActive      ThreatStatus = "active"
	ThreatStatusInvestigating ThreatStatus = "investigating"
	ThreatStatusMitigated   ThreatStatus = "mitigated"
	ThreatStatusResolved    ThreatStatus = "resolved"
	ThreatStatusFalsePositive ThreatStatus = "false_positive"
)

// EnrichedSecurityEvent extends SecurityEvent with additional context
type EnrichedSecurityEvent struct {
	SecurityEvent
	GeoLocation   *GeoLocation               `json:"geo_location"`
	DeviceInfo    *DeviceInfo                `json:"device_info"`
	ThreatIntel   *ThreatIntelligenceData    `json:"threat_intel"`
	Context       map[string]interface{}     `json:"context"`
}

// GeoLocation represents geographical information
type GeoLocation struct {
	Country     string  `json:"country"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ASN         string  `json:"asn"`
	ISP         string  `json:"isp"`
	Timezone    string  `json:"timezone"`
}

// DeviceInfo represents device-specific information
type DeviceInfo struct {
	DeviceID     string                 `json:"device_id"`
	DeviceType   string                 `json:"device_type"`
	OS           string                 `json:"os"`
	Browser      string                 `json:"browser"`
	UserAgent    string                 `json:"user_agent"`
	Fingerprint  string                 `json:"fingerprint"`
	IsKnown      bool                   `json:"is_known"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThreatIntelligenceData represents threat intelligence information
type ThreatIntelligenceData struct {
	Reputation   float64                `json:"reputation"`
	ThreatTypes  []string               `json:"threat_types"`
	Sources      []string               `json:"sources"`
	IOCs         []*IOC                 `json:"iocs"`
	Attribution  *AttackAttribution     `json:"attribution"`
	Context      *ThreatContext         `json:"context"`
	LastUpdated  time.Time              `json:"last_updated"`
}

// IOC represents an Indicator of Compromise
type IOC struct {
	Type        string    `json:"type"`
	Value       string    `json:"value"`
	Confidence  float64   `json:"confidence"`
	Source      string    `json:"source"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Tags        []string  `json:"tags"`
}

// AttackAttribution represents attribution information
type AttackAttribution struct {
	ThreatActor   string                 `json:"threat_actor"`
	Campaign      string                 `json:"campaign"`
	Country       string                 `json:"country"`
	Motivation    string                 `json:"motivation"`
	Confidence    float64                `json:"confidence"`
	Sources       []string               `json:"sources"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ThreatContext provides additional context about threats
type ThreatContext struct {
	Industry      string                 `json:"industry"`
	Region        string                 `json:"region"`
	TTPSUsed      []string               `json:"ttps_used"`
	KillChain     *KillChainMapping      `json:"kill_chain"`
	MitreMapping  *MitreAttackMapping    `json:"mitre_mapping"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// KillChainMapping maps to cyber kill chain phases
type KillChainMapping struct {
	Phase         string                 `json:"phase"`
	Description   string                 `json:"description"`
	Techniques    []string               `json:"techniques"`
	Indicators    []string               `json:"indicators"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// MitreAttackMapping maps to MITRE ATT&CK framework
type MitreAttackMapping struct {
	TacticID      string                 `json:"tactic_id"`
	TacticName    string                 `json:"tactic_name"`
	TechniqueID   string                 `json:"technique_id"`
	TechniqueName string                 `json:"technique_name"`
	SubTechnique  string                 `json:"sub_technique"`
	DataSources   []string               `json:"data_sources"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AnalysisResult represents the result of a specific analysis
type AnalysisResult struct {
	AnalysisType string                 `json:"analysis_type"`
	Score        float64                `json:"score"`
	Confidence   float64                `json:"confidence"`
	Indicators   []string               `json:"indicators"`
	Metadata     map[string]interface{} `json:"metadata"`
	Error        error                  `json:"error,omitempty"`
}

// BehaviorAnalysis represents behavioral analysis results
type BehaviorAnalysis struct {
	EntityID       string                 `json:"entity_id"`
	AnalysisWindow time.Duration          `json:"analysis_window"`
	Patterns       []*BehaviorPattern     `json:"patterns"`
	Anomalies      []*BehaviorAnomaly     `json:"anomalies"`
	RiskScore      float64                `json:"risk_score"`
	Confidence     float64                `json:"confidence"`
	Recommendations []string              `json:"recommendations"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// BehaviorPattern represents identified behavioral patterns
type BehaviorPattern struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Frequency    float64                `json:"frequency"`
	Confidence   float64                `json:"confidence"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// BehaviorAnomaly represents detected behavioral anomalies
type BehaviorAnomaly struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Severity     float64                `json:"severity"`
	Confidence   float64                `json:"confidence"`
	Timestamp    time.Time              `json:"timestamp"`
	Context      map[string]interface{} `json:"context"`
}

// BehaviorData represents raw behavioral data points
type BehaviorData struct {
	EntityID     string                 `json:"entity_id"`
	Timestamp    time.Time              `json:"timestamp"`
	ActionType   string                 `json:"action_type"`
	Resource     string                 `json:"resource"`
	Value        float64                `json:"value"`
	Context      map[string]interface{} `json:"context"`
}

// BehaviorModel represents a trained behavioral model
type BehaviorModel struct {
	EntityID       string                 `json:"entity_id"`
	ModelVersion   string                 `json:"model_version"`
	Baseline       *BehaviorBaseline      `json:"baseline"`
	Patterns       []*BehaviorPattern     `json:"patterns"`
	Features       []float64              `json:"features"`
	TrainingData   int                    `json:"training_data"`
	Accuracy       float64                `json:"accuracy"`
	CreatedAt      time.Time              `json:"created_at"`
	LastUpdated    time.Time              `json:"last_updated"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Activity represents user/entity activity
type Activity struct {
	ID           string                 `json:"id"`
	EntityID     string                 `json:"entity_id"`
	Type         string                 `json:"type"`
	Timestamp    time.Time              `json:"timestamp"`
	Resource     string                 `json:"resource"`
	Action       string                 `json:"action"`
	Result       string                 `json:"result"`
	Context      map[string]interface{} `json:"context"`
}

// Anomaly represents detected anomalies
type Anomaly struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Score        float64                `json:"score"`
	Severity     float64                `json:"severity"`
	Confidence   float64                `json:"confidence"`
	Timestamp    time.Time              `json:"timestamp"`
	Source       string                 `json:"source"`
	Evidence     map[string]interface{} `json:"evidence"`
}

// SecurityMetrics represents various security metrics
type SecurityMetrics struct {
	AuthenticationMetrics *AuthenticationMetrics `json:"authentication_metrics"`
	NetworkMetrics        *NetworkMetrics        `json:"network_metrics"`
	AccessMetrics         *AccessMetrics         `json:"access_metrics"`
	DataAccessMetrics     *DataAccessMetrics     `json:"data_access_metrics"`
	Timestamp             time.Time              `json:"timestamp"`
}

// AuthenticationMetrics represents authentication-related metrics
type AuthenticationMetrics struct {
	LoginAttempts       int       `json:"login_attempts"`
	FailedLogins        int       `json:"failed_logins"`
	SuccessfulLogins    int       `json:"successful_logins"`
	UnknownDevices      int       `json:"unknown_devices"`
	SuspiciousLocations int       `json:"suspicious_locations"`
	MFAFailures         int       `json:"mfa_failures"`
	AverageSessionTime  float64   `json:"average_session_time"`
	Timestamp           time.Time `json:"timestamp"`
}

// AccessMetrics represents access-related metrics
type AccessMetrics struct {
	ResourceAccesses    int       `json:"resource_accesses"`
	UnauthorizedAttempts int      `json:"unauthorized_attempts"`
	PrivilegeEscalations int      `json:"privilege_escalations"`
	DataExfiltration    int       `json:"data_exfiltration"`
	AdminActions        int       `json:"admin_actions"`
	APIUsage            int       `json:"api_usage"`
	Timestamp           time.Time `json:"timestamp"`
}

// DataAccessMetrics represents data access metrics
type DataAccessMetrics struct {
	SensitiveDataAccess int       `json:"sensitive_data_access"`
	BulkDataDownloads   int       `json:"bulk_data_downloads"`
	UnusualDataPatterns int       `json:"unusual_data_patterns"`
	DataModifications   int       `json:"data_modifications"`
	ExportOperations    int       `json:"export_operations"`
	Timestamp           time.Time `json:"timestamp"`
}

// ThreatFeatures represents features for ML threat detection
type ThreatFeatures struct {
	Category     string                 `json:"category"`
	Features     map[string]float64     `json:"features"`
	Timestamp    time.Time              `json:"timestamp"`
	Context      map[string]interface{} `json:"context"`
}

// MLThreatModel represents a machine learning threat detection model
type MLThreatModel struct {
	ID           string                 `json:"id"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Accuracy     float64                `json:"accuracy"`
	TrainedAt    time.Time              `json:"trained_at"`
	Features     []string               `json:"features"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// MLThreatResult represents ML threat detection result
type MLThreatResult struct {
	ThreatProbability float64                `json:"threat_probability"`
	ThreatLevel       string                 `json:"threat_level"`
	Confidence        float64                `json:"confidence"`
	ModelVersion      string                 `json:"model_version"`
	FeatureImportance map[string]float64     `json:"feature_importance"`
	PredictionTime    time.Time              `json:"prediction_time"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ThreatPattern represents identified threat patterns
type ThreatPattern struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	ThreatScore  float64                `json:"threat_score"`
	Confidence   float64                `json:"confidence"`
	Frequency    float64                `json:"frequency"`
	Events       []*SecurityEvent       `json:"events"`
	Indicators   []string               `json:"indicators"`
	FirstSeen    time.Time              `json:"first_seen"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// EnrichedThreat represents a threat enriched with intelligence
type EnrichedThreat struct {
	DetectedThreat
	ThreatIntel  *ThreatIntelligenceData `json:"threat_intel"`
	Attribution  *AttackAttribution      `json:"attribution"`
	Context      *ThreatContext          `json:"context"`
	Enrichments  map[string]interface{}  `json:"enrichments"`
}

// ResponseAction represents an automated security response
type ResponseAction struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Target       string                 `json:"target"`
	Parameters   map[string]interface{} `json:"parameters"`
	Status       string                 `json:"status"`
	ExecutedAt   time.Time              `json:"executed_at"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ThreatEscalation represents threat escalation information
type ThreatEscalation struct {
	ThreatID        string    `json:"threat_id"`
	EscalationLevel int       `json:"escalation_level"`
	EscalatedAt     time.Time `json:"escalated_at"`
	EscalatedBy     string    `json:"escalated_by"`
	Reason          string    `json:"reason"`
	PreviousLevel   string    `json:"previous_level"`
	NewLevel        string    `json:"new_level"`
}

// Entity represents a security entity (user, device, service, etc.)
type Entity struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
}

// Evidence represents security evidence
type Evidence struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Hash        string                 `json:"hash"`
}

// EventSource represents the source of a security event
type EventSource struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// RequestContext represents context for security requests
type RequestContext struct {
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Timestamp   time.Time              `json:"timestamp"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// AttackVector represents an attack vector
type AttackVector struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Method      string                 `json:"method"`
	Indicators  []string               `json:"indicators"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ThreatEvent represents an event in a threat timeline
type ThreatEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Evidence    map[string]interface{} `json:"evidence"`
}

// Helper methods for MLThreatModel
func (m *MLThreatModel) Predict(features []float64) (float64, error) {
	// Simplified prediction - in production would use actual ML model
	if len(features) == 0 {
		return 0.0, nil
	}
	
	// Simple weighted average as placeholder
	var sum float64
	for _, feature := range features {
		sum += feature
	}
	
	probability := sum / float64(len(features))
	if probability > 1.0 {
		probability = 1.0
	}
	if probability < 0.0 {
		probability = 0.0
	}
	
	return probability, nil
}

func (m *MLThreatModel) GetConfidence(features []float64) float64 {
	// Simplified confidence calculation
	if len(features) == 0 {
		return 0.0
	}
	return m.Accuracy
}

func (m *MLThreatModel) GetFeatureImportance(features []float64) map[string]float64 {
	// Simplified feature importance
	importance := make(map[string]float64)
	for i, feature := range features {
		importance[fmt.Sprintf("feature_%d", i)] = feature
	}
	return importance
}

// Helper function to create default security metrics
func NewSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		AuthenticationMetrics: &AuthenticationMetrics{
			Timestamp: time.Now(),
		},
		NetworkMetrics: &NetworkMetrics{
			Timestamp: time.Now(),
		},
		AccessMetrics: &AccessMetrics{
			Timestamp: time.Now(),
		},
		DataAccessMetrics: &DataAccessMetrics{
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}
}