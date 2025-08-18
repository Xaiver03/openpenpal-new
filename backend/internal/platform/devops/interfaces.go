package devops

import (
	"context"
	"time"
)

// IntelligentCICDEngine defines the core CI/CD pipeline interface
type IntelligentCICDEngine interface {
	// Build operations
	RunBuild(ctx context.Context, config *BuildConfig) (*BuildResult, error)
	OptimizeBuild(ctx context.Context, project *Project) (*OptimizedBuildPlan, error)
	PredictBuildTime(ctx context.Context, changes []Change) (time.Duration, error)
	
	// Test operations
	SelectTests(ctx context.Context, changes []Change) (*TestSelectionResult, error)
	RunTests(ctx context.Context, tests []Test) (*TestExecutionResult, error)
	AnalyzeTestImpact(ctx context.Context, changes []Change) (*ImpactAnalysis, error)
	
	// Deployment operations
	DeployApplication(ctx context.Context, config *DeploymentConfig) (*DeploymentResult, error)
	RunCanaryDeployment(ctx context.Context, config *CanaryConfig) (*CanaryResult, error)
	RollbackDeployment(ctx context.Context, deploymentID string) error
	
	// Pipeline management
	CreatePipeline(ctx context.Context, config *PipelineConfig) (*Pipeline, error)
	ExecutePipeline(ctx context.Context, pipelineID string) (*PipelineExecution, error)
	GetPipelineStatus(ctx context.Context, executionID string) (*PipelineStatus, error)
}

// BuildOptimizer handles intelligent build optimization
type BuildOptimizer interface {
	AnalyzeDependencies(ctx context.Context, project *Project) (*DependencyGraph, error)
	OptimizeBuildOrder(ctx context.Context, graph *DependencyGraph) (*BuildPlan, error)
	PredictBuildTime(ctx context.Context, changes []Change) (time.Duration, error)
	RecommendParallelization(ctx context.Context, resources *Resources) (*ParallelPlan, error)
	CacheAnalysis(ctx context.Context, project *Project) (*CacheStrategy, error)
	OptimizeDockerLayers(ctx context.Context, dockerfile string) (*OptimizedDockerfile, error)
}

// TestIntelligence provides ML-based test selection and optimization
type TestIntelligence interface {
	AnalyzeCodeChanges(ctx context.Context, diff *CodeDiff) (*ImpactAnalysis, error)
	SelectRelevantTests(ctx context.Context, impact *ImpactAnalysis) ([]Test, error)
	PredictTestFailures(ctx context.Context, changes []Change) (*FailurePrediction, error)
	OptimizeTestExecution(ctx context.Context, tests []Test) (*TestPlan, error)
	GenerateTestInsights(ctx context.Context, results *TestResults) (*TestInsights, error)
	RecommendNewTests(ctx context.Context, uncoveredCode []CodeBlock) ([]*TestRecommendation, error)
}

// DeploymentOrchestrator manages intelligent deployments
type DeploymentOrchestrator interface {
	PlanDeployment(ctx context.Context, config *DeploymentConfig) (*DeploymentPlan, error)
	ValidateDeployment(ctx context.Context, plan *DeploymentPlan) (*ValidationResult, error)
	ExecuteDeployment(ctx context.Context, plan *DeploymentPlan) (*DeploymentResult, error)
	MonitorDeployment(ctx context.Context, deploymentID string) (*DeploymentMetrics, error)
	AnalyzeDeploymentRisk(ctx context.Context, plan *DeploymentPlan) (*RiskAssessment, error)
	RecommendRollbackStrategy(ctx context.Context, deployment *Deployment) (*RollbackStrategy, error)
}

// QualityGateEngine manages automated quality checks
type QualityGateEngine interface {
	DefineQualityGate(ctx context.Context, config *QualityGateConfig) (*QualityGate, error)
	EvaluateCodeQuality(ctx context.Context, code *CodeAnalysis) (*QualityScore, error)
	CheckSecurityVulnerabilities(ctx context.Context, project *Project) (*SecurityReport, error)
	AnalyzePerformanceRegression(ctx context.Context, metrics *PerformanceMetrics) (*RegressionAnalysis, error)
	GenerateQualityReport(ctx context.Context, evaluation *QualityEvaluation) (*QualityReport, error)
	RecommendImprovements(ctx context.Context, issues []*QualityIssue) ([]*Improvement, error)
}

// PipelineAnalytics provides insights and analytics
type PipelineAnalytics interface {
	AnalyzePipelinePerformance(ctx context.Context, pipelineID string, timeRange TimeRange) (*PerformanceAnalysis, error)
	IdentifyBottlenecks(ctx context.Context, executions []*PipelineExecution) ([]*Bottleneck, error)
	PredictPipelineDuration(ctx context.Context, pipeline *Pipeline) (time.Duration, error)
	GenerateTrendAnalysis(ctx context.Context, pipelineID string) (*TrendAnalysis, error)
	RecommendOptimizations(ctx context.Context, analysis *PerformanceAnalysis) ([]*Optimization, error)
	CalculateROI(ctx context.Context, improvements []*Improvement) (*ROIAnalysis, error)
}

// Core data models

// BuildConfig represents build configuration
type BuildConfig struct {
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"project_id"`
	Branch          string                 `json:"branch"`
	CommitHash      string                 `json:"commit_hash"`
	BuildType       BuildType              `json:"build_type"`
	Environment     string                 `json:"environment"`
	BuildArgs       map[string]string      `json:"build_args"`
	CacheEnabled    bool                   `json:"cache_enabled"`
	ParallelJobs    int                    `json:"parallel_jobs"`
	Timeout         time.Duration          `json:"timeout"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// BuildResult represents build execution result
type BuildResult struct {
	ID              string                 `json:"id"`
	BuildID         string                 `json:"build_id"`
	Status          BuildStatus            `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Artifacts       []*Artifact            `json:"artifacts"`
	Logs            string                 `json:"logs"`
	ExitCode        int                    `json:"exit_code"`
	CacheHitRate    float64                `json:"cache_hit_rate"`
	ResourceUsage   *ResourceUsage         `json:"resource_usage"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TestSelectionResult represents intelligent test selection
type TestSelectionResult struct {
	TotalTests      int                    `json:"total_tests"`
	SelectedTests   int                    `json:"selected_tests"`
	SkippedTests    int                    `json:"skipped_tests"`
	Tests           []*TestSelection       `json:"tests"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	CoverageImpact  float64                `json:"coverage_impact"`
	Confidence      float64                `json:"confidence"`
	SelectionReason string                 `json:"selection_reason"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DeploymentConfig represents deployment configuration
type DeploymentConfig struct {
	ID               string                 `json:"id"`
	ApplicationID    string                 `json:"application_id"`
	Version          string                 `json:"version"`
	Environment      string                 `json:"environment"`
	Strategy         DeploymentStrategy     `json:"strategy"`
	TargetInstances  int                    `json:"target_instances"`
	HealthCheckURL   string                 `json:"health_check_url"`
	RollbackOnFail   bool                   `json:"rollback_on_fail"`
	Timeout          time.Duration          `json:"timeout"`
	Configuration    map[string]interface{} `json:"configuration"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CanaryConfig represents canary deployment configuration
type CanaryConfig struct {
	DeploymentConfig
	TrafficPercentage   int                    `json:"traffic_percentage"`
	Duration            time.Duration          `json:"duration"`
	SuccessThreshold    float64                `json:"success_threshold"`
	MetricsToMonitor    []string               `json:"metrics_to_monitor"`
	AutoPromote         bool                   `json:"auto_promote"`
	ComparisonBaseline  string                 `json:"comparison_baseline"`
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ProjectID       string                 `json:"project_id"`
	Stages          []*PipelineStage       `json:"stages"`
	Triggers        []*PipelineTrigger     `json:"triggers"`
	Parameters      []*PipelineParameter   `json:"parameters"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Configuration   map[string]interface{} `json:"configuration"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PipelineExecution represents a pipeline run
type PipelineExecution struct {
	ID              string                 `json:"id"`
	PipelineID      string                 `json:"pipeline_id"`
	Status          PipelineStatus         `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty"`
	Duration        time.Duration          `json:"duration"`
	TriggerType     TriggerType            `json:"trigger_type"`
	TriggerUser     string                 `json:"trigger_user"`
	StageExecutions []*StageExecution      `json:"stage_executions"`
	Parameters      map[string]interface{} `json:"parameters"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// QualityGateConfig defines quality gate criteria
type QualityGateConfig struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	CodeCoverageMin     float64                `json:"code_coverage_min"`
	TestPassRateMin     float64                `json:"test_pass_rate_min"`
	SecuritySeverity    string                 `json:"security_severity"`
	PerformanceBaseline *PerformanceBaseline   `json:"performance_baseline"`
	CustomChecks        []*CustomCheck         `json:"custom_checks"`
	FailureAction       string                 `json:"failure_action"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// Enums and constants

type BuildType string
const (
	BuildTypeDebug      BuildType = "debug"
	BuildTypeRelease    BuildType = "release"
	BuildTypeProduction BuildType = "production"
)

type BuildStatus string
const (
	BuildStatusPending  BuildStatus = "pending"
	BuildStatusRunning  BuildStatus = "running"
	BuildStatusSuccess  BuildStatus = "success"
	BuildStatusFailed   BuildStatus = "failed"
	BuildStatusCanceled BuildStatus = "canceled"
)

type DeploymentStrategy string
const (
	DeploymentStrategyRolling     DeploymentStrategy = "rolling"
	DeploymentStrategyBlueGreen   DeploymentStrategy = "blue_green"
	DeploymentStrategyCanary      DeploymentStrategy = "canary"
	DeploymentStrategyRecreate    DeploymentStrategy = "recreate"
	DeploymentStrategyAB          DeploymentStrategy = "ab_testing"
)

type PipelineStatus string
const (
	PipelineStatusQueued    PipelineStatus = "queued"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSuccess   PipelineStatus = "success"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCanceled  PipelineStatus = "canceled"
	PipelineStatusPaused    PipelineStatus = "paused"
)

type TriggerType string
const (
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeCommit    TriggerType = "commit"
	TriggerTypeSchedule  TriggerType = "schedule"
	TriggerTypeWebhook   TriggerType = "webhook"
	TriggerTypeAPI       TriggerType = "api"
)