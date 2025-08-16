package security

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"openpenpal-backend/internal/config"
	"go.uber.org/zap"
)

// PolicyEngine implements a sophisticated rule-based security policy system
// with support for dynamic policies, machine learning insights, and compliance frameworks
type PolicyEngine struct {
	config           *config.Config
	logger           *zap.Logger
	policies         map[string]*SecurityPolicy
	policyGroups     map[string]*PolicyGroup
	ruleEvaluator    *RuleEvaluator
	policyOptimizer  *PolicyOptimizer
	complianceFrameworks map[string]*ComplianceFramework
	dynamicPolicies  *DynamicPolicyManager
	mu               sync.RWMutex
	running          bool
	evaluationCache  map[string]*CachedDecision
	cacheTimeout     time.Duration
}

type PolicyGroup struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Policies    []string           `json:"policies"`
	Priority    int                `json:"priority"`
	Enabled     bool               `json:"enabled"`
	Conditions  []GroupCondition   `json:"conditions"`
	Metadata    map[string]string  `json:"metadata"`
}

type GroupCondition struct {
	Type     ConditionType `json:"type"`
	Value    interface{}   `json:"value"`
	Operator string        `json:"operator"`
}

type ConditionType string

const (
	ConditionTimeRange    ConditionType = "time_range"
	ConditionUserRole     ConditionType = "user_role"
	ConditionIPRange      ConditionType = "ip_range"
	ConditionGeoLocation  ConditionType = "geo_location"
	ConditionRiskScore    ConditionType = "risk_score"
	ConditionTrustLevel   ConditionType = "trust_level"
	ConditionResource     ConditionType = "resource"
	ConditionAction       ConditionType = "action"
)

type RuleEvaluator struct {
	evaluationEngine *ExpressionEngine
	contextProvider  *ContextProvider
	operatorRegistry map[string]Operator
	mu               sync.RWMutex
}

type ExpressionEngine struct {
	builtinFunctions map[string]Function
	operators        map[string]Operator
}

type Function interface {
	Name() string
	Execute(args []interface{}) (interface{}, error)
	Validate(args []interface{}) error
}

type Operator interface {
	Name() string
	Evaluate(left, right interface{}) (bool, error)
	SupportedTypes() []string
}

type ContextProvider struct {
	contextSources map[string]ContextSource
	mu             sync.RWMutex
}

type ContextSource interface {
	GetContext(key string) (interface{}, error)
	GetType() string
}

type PolicyOptimizer struct {
	optimizationRules []OptimizationRule
	performanceMetrics *PerformanceMetrics
	mlInsights        *MLInsights
	mu                sync.RWMutex
}

type OptimizationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Condition   OptimizationCondition  `json:"condition"`
	Action      OptimizationAction     `json:"action"`
	Enabled     bool                   `json:"enabled"`
	Metrics     map[string]float64     `json:"metrics"`
}

type OptimizationCondition struct {
	Type      string      `json:"type"`
	Threshold float64     `json:"threshold"`
	Metric    string      `json:"metric"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
}

type OptimizationAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Impact     string                 `json:"impact"`
}

type PerformanceMetrics struct {
	EvaluationTime     map[string]time.Duration `json:"evaluation_time"`
	ThroughputRPS      float64                  `json:"throughput_rps"`
	CacheHitRate       float64                  `json:"cache_hit_rate"`
	PolicyEffectiveness map[string]float64      `json:"policy_effectiveness"`
	ErrorRate          float64                  `json:"error_rate"`
	LastUpdated        time.Time                `json:"last_updated"`
}

type MLInsights struct {
	ThreatPredictions   []ThreatPrediction   `json:"threat_predictions"`
	PolicySuggestions   []PolicySuggestion   `json:"policy_suggestions"`
	AnomalyDetection    []AnomalyAlert       `json:"anomaly_detection"`
	OptimizationInsights []OptimizationInsight `json:"optimization_insights"`
	LastAnalysis        time.Time            `json:"last_analysis"`
	Confidence          float64              `json:"confidence"`
}

type ThreatPrediction struct {
	ThreatType   string    `json:"threat_type"`
	Probability  float64   `json:"probability"`
	Impact       string    `json:"impact"`
	Indicators   []string  `json:"indicators"`
	Timestamp    time.Time `json:"timestamp"`
	Confidence   float64   `json:"confidence"`
}

type PolicySuggestion struct {
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Priority     int                    `json:"priority"`
	Impact       string                 `json:"impact"`
	PolicyDraft  *SecurityPolicy        `json:"policy_draft"`
	Evidence     map[string]interface{} `json:"evidence"`
	Confidence   float64                `json:"confidence"`
}

type AnomalyAlert struct {
	AlertID     string                 `json:"alert_id"`
	Type        string                 `json:"type"`
	Severity    Severity               `json:"severity"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
}

type OptimizationInsight struct {
	InsightID   string                 `json:"insight_id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Benefit     string                 `json:"benefit"`
	Effort      string                 `json:"effort"`
	Data        map[string]interface{} `json:"data"`
	Confidence  float64                `json:"confidence"`
}

type ComplianceFramework struct {
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	Version      string                    `json:"version"`
	Requirements []ComplianceRequirement   `json:"requirements"`
	Policies     []string                  `json:"policies"`
	ValidationRules []ValidationRule       `json:"validation_rules"`
	Enabled      bool                      `json:"enabled"`
	LastUpdated  time.Time                 `json:"last_updated"`
}

type ComplianceRequirement struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Controls    []string `json:"controls"`
	Mandatory   bool     `json:"mandatory"`
	Evidence    []string `json:"evidence"`
}

type ValidationRule struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Condition interface{} `json:"condition"`
	Action    string      `json:"action"`
	Severity  Severity    `json:"severity"`
}

type DynamicPolicyManager struct {
	activePolicies    map[string]*DynamicPolicy
	policyTemplates   map[string]*PolicyTemplate
	triggerEngine     *TriggerEngine
	adaptationEngine  *AdaptationEngine
	mu                sync.RWMutex
}

type DynamicPolicy struct {
	ID              string            `json:"id"`
	BasePolicy      *SecurityPolicy   `json:"base_policy"`
	Adaptations     []Adaptation      `json:"adaptations"`
	Triggers        []Trigger         `json:"triggers"`
	TTL             time.Duration     `json:"ttl"`
	CreatedAt       time.Time         `json:"created_at"`
	LastAdapted     time.Time         `json:"last_adapted"`
	EffectivenessScore float64        `json:"effectiveness_score"`
	Metadata        map[string]string `json:"metadata"`
}

type Adaptation struct {
	Type        AdaptationType         `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Reason      string                 `json:"reason"`
	Timestamp   time.Time              `json:"timestamp"`
	Effectiveness float64              `json:"effectiveness"`
}

type AdaptationType string

const (
	AdaptationRuleModification AdaptationType = "rule_modification"
	AdaptationThresholdAdjust  AdaptationType = "threshold_adjust"
	AdaptationActionChange     AdaptationType = "action_change"
	AdaptationConditionAdd     AdaptationType = "condition_add"
	AdaptationPriorityAdjust   AdaptationType = "priority_adjust"
)

type Trigger struct {
	ID        string                 `json:"id"`
	Type      TriggerType            `json:"type"`
	Condition TriggerCondition       `json:"condition"`
	Action    TriggerAction          `json:"action"`
	Enabled   bool                   `json:"enabled"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type TriggerType string

const (
	TriggerThreatLevel    TriggerType = "threat_level"
	TriggerAttackPattern  TriggerType = "attack_pattern"
	TriggerPerformance    TriggerType = "performance"
	TriggerCompliance     TriggerType = "compliance"
	TriggerUserBehavior   TriggerType = "user_behavior"
)

type TriggerCondition struct {
	Metric    string      `json:"metric"`
	Operator  string      `json:"operator"`
	Threshold interface{} `json:"threshold"`
	Duration  time.Duration `json:"duration"`
}

type TriggerAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
}

type PolicyTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Template    *SecurityPolicy        `json:"template"`
	Variables   []TemplateVariable     `json:"variables"`
	UseCase     string                 `json:"use_case"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Description  string      `json:"description"`
	Validation   string      `json:"validation"`
}

type CachedDecision struct {
	Decision    *PolicyDecision `json:"decision"`
	CachedAt    time.Time       `json:"cached_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	HitCount    int             `json:"hit_count"`
	LastAccess  time.Time       `json:"last_access"`
}

type TriggerEngine struct{}
type AdaptationEngine struct{}

func NewPolicyEngine(cfg *config.Config, logger *zap.Logger) *PolicyEngine {
	pe := &PolicyEngine{
		config:               cfg,
		logger:               logger,
		policies:             make(map[string]*SecurityPolicy),
		policyGroups:         make(map[string]*PolicyGroup),
		complianceFrameworks: make(map[string]*ComplianceFramework),
		evaluationCache:      make(map[string]*CachedDecision),
		cacheTimeout:         5 * time.Minute,
	}

	// Initialize sub-components
	pe.ruleEvaluator = NewRuleEvaluator(logger)
	pe.policyOptimizer = NewPolicyOptimizer(logger)
	pe.dynamicPolicies = NewDynamicPolicyManager(logger)

	// Load default policies
	pe.loadDefaultPolicies()
	pe.loadComplianceFrameworks()

	return pe
}

func (pe *PolicyEngine) Start(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.running {
		return fmt.Errorf("policy engine already running")
	}

	pe.logger.Info("Starting Security Policy Engine")
	pe.running = true

	// Start background optimization
	go pe.backgroundOptimization(ctx)

	return nil
}

func (pe *PolicyEngine) Stop(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.running {
		return nil
	}

	pe.logger.Info("Stopping Security Policy Engine")
	pe.running = false

	return nil
}

func (pe *PolicyEngine) EvaluatePolicies(ctx context.Context, securityCtx *SecurityContext, request *SecurityRequest) (*PolicyDecision, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	if !pe.running {
		return nil, fmt.Errorf("policy engine not running")
	}

	// Generate cache key
	cacheKey := pe.generateCacheKey(securityCtx, request)

	// Check cache first
	if cached := pe.getCachedDecision(cacheKey); cached != nil {
		cached.HitCount++
		cached.LastAccess = time.Now()
		pe.logger.Debug("Using cached policy decision", zap.String("cache_key", cacheKey))
		return cached.Decision, nil
	}

	pe.logger.Info("Evaluating security policies",
		zap.String("user_id", securityCtx.UserID),
		zap.String("resource", request.Resource),
		zap.String("action", request.Action))

	// Get applicable policies
	applicablePolicies := pe.getApplicablePolicies(securityCtx, request)

	// Evaluate policies in priority order
	sort.Slice(applicablePolicies, func(i, j int) bool {
		return applicablePolicies[i].Priority > applicablePolicies[j].Priority
	})

	decision := &PolicyDecision{
		Allow:       false,
		Permissions: []Permission{},
		Actions:     []PolicyAction{},
	}

	var evaluationErrors []string

	for _, policy := range applicablePolicies {
		policyResult, err := pe.evaluatePolicy(policy, securityCtx, request)
		if err != nil {
			evaluationErrors = append(evaluationErrors, fmt.Sprintf("Policy %s: %v", policy.ID, err))
			continue
		}

		// Apply policy result
		if policyResult.Allow {
			decision.Allow = true
			decision.Permissions = append(decision.Permissions, policyResult.Permissions...)
		}

		decision.Actions = append(decision.Actions, policyResult.Actions...)

		pe.logger.Debug("Policy evaluated",
			zap.String("policy_id", policy.ID),
			zap.Bool("allow", policyResult.Allow))
	}

	// Cache the decision
	pe.cacheDecision(cacheKey, decision)

	if len(evaluationErrors) > 0 {
		pe.logger.Warn("Some policies failed to evaluate", zap.Strings("errors", evaluationErrors))
	}

	pe.logger.Info("Policy evaluation completed",
		zap.String("user_id", securityCtx.UserID),
		zap.Bool("allow", decision.Allow),
		zap.Int("applicable_policies", len(applicablePolicies)))

	return decision, nil
}

func (pe *PolicyEngine) loadDefaultPolicies() {
	// Default administrative policy
	adminPolicy := &SecurityPolicy{
		ID:          "admin_policy",
		Name:        "Administrative Access Policy",
		Description: "Controls administrative access to the system",
		Priority:    100,
		Enabled:     true,
		CreatedAt:   time.Now(),
		Rules: []PolicyRule{
			{
				Field:    "user_role",
				Operator: "equals",
				Value:    "admin",
				Weight:   1.0,
			},
			{
				Field:    "trust_level",
				Operator: "gte",
				Value:    TrustLevelHigh,
				Weight:   1.0,
			},
		},
		Actions: []PolicyAction{
			{
				Type: ActionAllow,
				Parameters: map[string]interface{}{
					"scope": "administrative",
				},
			},
		},
	}

	// Default user policy
	userPolicy := &SecurityPolicy{
		ID:          "user_policy",
		Name:        "Standard User Policy",
		Description: "Controls standard user access",
		Priority:    50,
		Enabled:     true,
		CreatedAt:   time.Now(),
		Rules: []PolicyRule{
			{
				Field:    "trust_level",
				Operator: "gte",
				Value:    TrustLevelMedium,
				Weight:   1.0,
			},
			{
				Field:    "risk_score",
				Operator: "lt",
				Value:    0.7,
				Weight:   0.8,
			},
		},
		Actions: []PolicyAction{
			{
				Type: ActionAllow,
				Parameters: map[string]interface{}{
					"scope": "user",
				},
			},
		},
	}

	// High-risk prevention policy
	riskPolicy := &SecurityPolicy{
		ID:          "risk_policy",
		Name:        "High Risk Prevention Policy",
		Description: "Blocks high-risk requests",
		Priority:    200,
		Enabled:     true,
		CreatedAt:   time.Now(),
		Rules: []PolicyRule{
			{
				Field:    "risk_score",
				Operator: "gte",
				Value:    0.8,
				Weight:   1.0,
			},
		},
		Actions: []PolicyAction{
			{
				Type: ActionDeny,
				Parameters: map[string]interface{}{
					"reason": "High risk score detected",
				},
			},
		},
	}

	pe.policies[adminPolicy.ID] = adminPolicy
	pe.policies[userPolicy.ID] = userPolicy
	pe.policies[riskPolicy.ID] = riskPolicy

	pe.logger.Info("Loaded default security policies", zap.Int("count", len(pe.policies)))
}

func (pe *PolicyEngine) loadComplianceFrameworks() {
	// GDPR compliance framework
	gdpr := &ComplianceFramework{
		ID:      "gdpr",
		Name:    "General Data Protection Regulation",
		Version: "2018",
		Enabled: true,
		Requirements: []ComplianceRequirement{
			{
				ID:          "gdpr_data_protection",
				Category:    "data_protection",
				Description: "Ensure data protection by design and by default",
				Mandatory:   true,
				Controls:    []string{"encryption", "access_control", "audit_logging"},
			},
		},
		LastUpdated: time.Now(),
	}

	pe.complianceFrameworks[gdpr.ID] = gdpr
	pe.logger.Info("Loaded compliance frameworks", zap.Int("count", len(pe.complianceFrameworks)))
}

func (pe *PolicyEngine) getApplicablePolicies(securityCtx *SecurityContext, request *SecurityRequest) []*SecurityPolicy {
	var applicable []*SecurityPolicy

	for _, policy := range pe.policies {
		if !policy.Enabled {
			continue
		}

		if pe.isPolicyApplicable(policy, securityCtx, request) {
			applicable = append(applicable, policy)
		}
	}

	return applicable
}

func (pe *PolicyEngine) isPolicyApplicable(policy *SecurityPolicy, securityCtx *SecurityContext, request *SecurityRequest) bool {
	// Check policy conditions
	for _, condition := range policy.Conditions {
		if !pe.evaluateCondition(condition, securityCtx, request) {
			return false
		}
	}
	return true
}

func (pe *PolicyEngine) evaluateCondition(condition PolicyCondition, securityCtx *SecurityContext, request *SecurityRequest) bool {
	// Simplified condition evaluation
	switch condition.Field {
	case "resource":
		return pe.matchesPattern(request.Resource, condition.Value.(string))
	case "action":
		return request.Action == condition.Value.(string)
	case "trust_level":
		return pe.compareTrustLevel(securityCtx.TrustLevel, condition.Operator, condition.Value)
	default:
		return true
	}
}

func (pe *PolicyEngine) matchesPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

func (pe *PolicyEngine) compareTrustLevel(level TrustLevel, operator string, value interface{}) bool {
	targetLevel, ok := value.(TrustLevel)
	if !ok {
		return false
	}

	switch operator {
	case "equals":
		return level == targetLevel
	case "gte":
		return level >= targetLevel
	case "gt":
		return level > targetLevel
	case "lte":
		return level <= targetLevel
	case "lt":
		return level < targetLevel
	default:
		return false
	}
}

func (pe *PolicyEngine) evaluatePolicy(policy *SecurityPolicy, securityCtx *SecurityContext, request *SecurityRequest) (*PolicyDecision, error) {
	decision := &PolicyDecision{
		Allow:       false,
		Permissions: []Permission{},
		Actions:     []PolicyAction{},
	}

	totalWeight := 0.0
	passedWeight := 0.0

	// Evaluate all rules
	for _, rule := range policy.Rules {
		totalWeight += rule.Weight
		
		if pe.evaluateRule(rule, securityCtx, request) {
			passedWeight += rule.Weight
		}
	}

	// Determine if policy passes (simple majority for now)
	if totalWeight > 0 && passedWeight/totalWeight >= 0.5 {
		decision.Allow = true
		decision.Actions = policy.Actions
		
		// Generate permissions based on policy
		permission := Permission{
			Resource: request.Resource,
			Action:   request.Action,
			Scope:    "default",
			TTL:      time.Hour,
		}
		decision.Permissions = append(decision.Permissions, permission)
	}

	return decision, nil
}

func (pe *PolicyEngine) evaluateRule(rule PolicyRule, securityCtx *SecurityContext, request *SecurityRequest) bool {
	switch rule.Field {
	case "user_role":
		// This would integrate with actual user role system
		return true // Simplified for demo
	case "trust_level":
		return pe.compareTrustLevel(securityCtx.TrustLevel, rule.Operator, rule.Value)
	case "risk_score":
		return pe.compareFloat(securityCtx.RiskScore, rule.Operator, rule.Value)
	default:
		return false
	}
}

func (pe *PolicyEngine) compareFloat(value float64, operator string, target interface{}) bool {
	targetFloat, ok := target.(float64)
	if !ok {
		return false
	}

	switch operator {
	case "equals":
		return value == targetFloat
	case "gte":
		return value >= targetFloat
	case "gt":
		return value > targetFloat
	case "lte":
		return value <= targetFloat
	case "lt":
		return value < targetFloat
	default:
		return false
	}
}

func (pe *PolicyEngine) generateCacheKey(securityCtx *SecurityContext, request *SecurityRequest) string {
	key := fmt.Sprintf("%s:%s:%s:%s:%f",
		securityCtx.UserID,
		request.Resource,
		request.Action,
		pe.trustLevelToString(securityCtx.TrustLevel),
		securityCtx.RiskScore)
	return key
}

func (pe *PolicyEngine) getCachedDecision(key string) *CachedDecision {
	cached, exists := pe.evaluationCache[key]
	if !exists {
		return nil
	}

	if time.Now().After(cached.ExpiresAt) {
		delete(pe.evaluationCache, key)
		return nil
	}

	return cached
}

func (pe *PolicyEngine) cacheDecision(key string, decision *PolicyDecision) {
	pe.evaluationCache[key] = &CachedDecision{
		Decision:   decision,
		CachedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(pe.cacheTimeout),
		HitCount:   0,
		LastAccess: time.Now(),
	}
}

func (pe *PolicyEngine) trustLevelToString(level TrustLevel) string {
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

func (pe *PolicyEngine) backgroundOptimization(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pe.optimizePolicies()
		}
	}
}

func (pe *PolicyEngine) optimizePolicies() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.running {
		return
	}

	pe.logger.Info("Running policy optimization")

	// Clean expired cache entries
	now := time.Now()
	for key, cached := range pe.evaluationCache {
		if now.After(cached.ExpiresAt) {
			delete(pe.evaluationCache, key)
		}
	}

	// Additional optimization logic would go here
	pe.logger.Info("Policy optimization completed")
}

// Stub implementations for sub-components
func NewRuleEvaluator(logger *zap.Logger) *RuleEvaluator {
	return &RuleEvaluator{
		evaluationEngine: &ExpressionEngine{
			builtinFunctions: make(map[string]Function),
			operators:        make(map[string]Operator),
		},
		contextProvider:  &ContextProvider{
			contextSources: make(map[string]ContextSource),
		},
		operatorRegistry: make(map[string]Operator),
	}
}

func NewPolicyOptimizer(logger *zap.Logger) *PolicyOptimizer {
	return &PolicyOptimizer{
		optimizationRules: []OptimizationRule{},
		performanceMetrics: &PerformanceMetrics{
			EvaluationTime:      make(map[string]time.Duration),
			PolicyEffectiveness: make(map[string]float64),
			LastUpdated:         time.Now(),
		},
		mlInsights: &MLInsights{
			ThreatPredictions:    []ThreatPrediction{},
			PolicySuggestions:    []PolicySuggestion{},
			AnomalyDetection:     []AnomalyAlert{},
			OptimizationInsights: []OptimizationInsight{},
			LastAnalysis:         time.Now(),
		},
	}
}

func NewDynamicPolicyManager(logger *zap.Logger) *DynamicPolicyManager {
	return &DynamicPolicyManager{
		activePolicies:   make(map[string]*DynamicPolicy),
		policyTemplates:  make(map[string]*PolicyTemplate),
		triggerEngine:    &TriggerEngine{},
		adaptationEngine: &AdaptationEngine{},
	}
}