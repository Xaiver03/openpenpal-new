// Package core provides the foundational interfaces for the SOTA Testing Infrastructure
package core

import (
	"context"
	"time"
)

// TestEngine is the main interface for the SOTA Testing Infrastructure
type TestEngine interface {
	// Initialize initializes the testing engine
	Initialize(ctx context.Context, config *TestingConfig) error
	
	// GenerateTests generates test cases using AI
	GenerateTests(ctx context.Context, target *TestTarget) (*TestSuite, error)
	
	// ExecuteTests executes a test suite
	ExecuteTests(ctx context.Context, suite *TestSuite) (*TestResults, error)
	
	// AnalyzeResults analyzes test results using ML
	AnalyzeResults(ctx context.Context, results *TestResults) (*TestAnalysis, error)
	
	// GetMetrics returns testing metrics
	GetMetrics(ctx context.Context) (*TestingMetrics, error)
}

// AITestGenerator generates intelligent test cases
type AITestGenerator interface {
	// AnalyzeCode analyzes source code to understand testing requirements
	AnalyzeCode(ctx context.Context, codebase *Codebase) (*CodeAnalysis, error)
	
	// GenerateTestCases generates test cases based on code analysis
	GenerateTestCases(ctx context.Context, analysis *CodeAnalysis) ([]*TestCase, error)
	
	// OptimizeCoverage optimizes test cases for maximum coverage
	OptimizeCoverage(ctx context.Context, testCases []*TestCase) ([]*TestCase, error)
	
	// LearnFromResults learns from test execution results
	LearnFromResults(ctx context.Context, results *TestResults) error
}

// SmartDataGenerator generates realistic test data
type SmartDataGenerator interface {
	// AnalyzeSchema analyzes database schema and data patterns
	AnalyzeSchema(ctx context.Context, schema *DatabaseSchema) (*DataProfile, error)
	
	// GenerateTestData generates synthetic test data
	GenerateTestData(ctx context.Context, profile *DataProfile, volume int) (*TestDataSet, error)
	
	// PreserveRelationships maintains referential integrity
	PreserveRelationships(ctx context.Context, dataset *TestDataSet) error
	
	// AnonymizeData anonymizes sensitive data for testing
	AnonymizeData(ctx context.Context, data interface{}) (interface{}, error)
}

// PerformanceEngine provides performance and load testing capabilities
type PerformanceEngine interface {
	// CreateBaseline establishes performance baselines
	CreateBaseline(ctx context.Context, config *BaselineConfig) (*PerformanceBaseline, error)
	
	// ExecuteLoadTest runs adaptive load tests
	ExecuteLoadTest(ctx context.Context, config *LoadTestConfig) (*LoadTestResults, error)
	
	// MonitorResources monitors system resources during testing
	MonitorResources(ctx context.Context, duration time.Duration) (*ResourceMetrics, error)
	
	// DetectBottlenecks identifies performance bottlenecks
	DetectBottlenecks(ctx context.Context, metrics *ResourceMetrics) ([]*Bottleneck, error)
}

// EnvironmentManager manages test environments
type EnvironmentManager interface {
	// CreateEnvironment creates an isolated test environment
	CreateEnvironment(ctx context.Context, config *EnvironmentConfig) (*TestEnvironment, error)
	
	// DestroyEnvironment cleans up a test environment
	DestroyEnvironment(ctx context.Context, envID string) error
	
	// ListEnvironments returns all available test environments
	ListEnvironments(ctx context.Context) ([]*TestEnvironment, error)
	
	// GetEnvironmentStatus returns the status of a test environment
	GetEnvironmentStatus(ctx context.Context, envID string) (*EnvironmentStatus, error)
}

// IntelligentAnalyzer provides ML-based test result analysis
type IntelligentAnalyzer interface {
	// ClassifyResults classifies test results using ML
	ClassifyResults(ctx context.Context, results *TestResults) (*ResultClassification, error)
	
	// AnalyzeTrends analyzes long-term testing trends
	AnalyzeTrends(ctx context.Context, timeRange TimeRange) (*TrendAnalysis, error)
	
	// AssessRisk assesses production risks based on test results
	AssessRisk(ctx context.Context, results *TestResults) (*RiskAssessment, error)
	
	// GenerateReport generates comprehensive test reports
	GenerateReport(ctx context.Context, analysis *TestAnalysis) (*TestReport, error)
}

// Data Models

// TestingConfig represents the configuration for the testing engine
type TestingConfig struct {
	AIConfig          *AIConfig          `json:"ai_config"`
	DataConfig        *DataConfig        `json:"data_config"`
	PerformanceConfig *PerformanceConfig `json:"performance_config"`
	EnvironmentConfig *EnvironmentConfig `json:"environment_config"`
	AnalysisConfig    *AnalysisConfig    `json:"analysis_config"`
}

// EnvironmentConfig represents test environment configuration
type EnvironmentConfig struct {
	Type        EnvironmentType        `json:"type"`
	Resources   *EnvironmentResources  `json:"resources"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TestTarget represents a target for test generation
type TestTarget struct {
	Type         TestTargetType `json:"type"`
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Dependencies []string       `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TestCases   []*TestCase `json:"test_cases"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   string      `json:"created_by"`
	Tags        []string    `json:"tags"`
}

// TestCase represents a single test case
type TestCase struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         TestCaseType           `json:"type"`
	Setup        []string               `json:"setup"`
	Actions      []string               `json:"actions"`
	Assertions   []string               `json:"assertions"`
	Teardown     []string               `json:"teardown"`
	ExpectedData interface{}            `json:"expected_data"`
	TestData     interface{}            `json:"test_data"`
	Metadata     map[string]interface{} `json:"metadata"`
	Priority     TestPriority           `json:"priority"`
	Tags         []string               `json:"tags"`
}

// TestResults represents the results of test execution
type TestResults struct {
	SuiteID         string             `json:"suite_id"`
	ExecutionID     string             `json:"execution_id"`
	StartTime       time.Time          `json:"start_time"`
	EndTime         time.Time          `json:"end_time"`
	Duration        time.Duration      `json:"duration"`
	TotalTests      int                `json:"total_tests"`
	PassedTests     int                `json:"passed_tests"`
	FailedTests     int                `json:"failed_tests"`
	SkippedTests    int                `json:"skipped_tests"`
	CoveragePercent float64            `json:"coverage_percent"`
	TestCaseResults []*TestCaseResult  `json:"test_case_results"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TestCaseResult represents the result of a single test case
type TestCaseResult struct {
	TestCaseID   string        `json:"test_case_id"`
	Status       TestStatus    `json:"status"`
	Duration     time.Duration `json:"duration"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Output       string        `json:"output,omitempty"`
	Screenshots  []string      `json:"screenshots,omitempty"`
	Metrics      *TestMetrics  `json:"metrics,omitempty"`
}

// TestAnalysis represents the analysis of test results
type TestAnalysis struct {
	ExecutionID      string                 `json:"execution_id"`
	QualityScore     float64                `json:"quality_score"`
	RiskLevel        RiskLevel              `json:"risk_level"`
	Classification   *ResultClassification  `json:"classification"`
	TrendAnalysis    *TrendAnalysis         `json:"trend_analysis"`
	RiskAssessment   *RiskAssessment        `json:"risk_assessment"`
	Recommendations  []string               `json:"recommendations"`
	Insights         []string               `json:"insights"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Codebase represents source code for analysis
type Codebase struct {
	Path         string            `json:"path"`
	Language     string            `json:"language"`
	Framework    string            `json:"framework"`
	Dependencies []string          `json:"dependencies"`
	Files        []*SourceFile     `json:"files"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SourceFile represents a single source code file
type SourceFile struct {
	Path         string   `json:"path"`
	Language     string   `json:"language"`
	Content      string   `json:"content"`
	Functions    []string `json:"functions"`
	Classes      []string `json:"classes"`
	Imports      []string `json:"imports"`
	LineCount    int      `json:"line_count"`
	Complexity   int      `json:"complexity"`
}

// CodeAnalysis represents the result of code analysis
type CodeAnalysis struct {
	CodebaseID       string                 `json:"codebase_id"`
	TotalFiles       int                    `json:"total_files"`
	TotalLines       int                    `json:"total_lines"`
	Complexity       int                    `json:"complexity"`
	TestableUnits    []*TestableUnit        `json:"testable_units"`
	Dependencies     []*Dependency          `json:"dependencies"`
	RiskAreas        []*RiskArea            `json:"risk_areas"`
	CoverageGaps     []*CoverageGap         `json:"coverage_gaps"`
	Patterns         []*CodePattern         `json:"patterns"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// TestableUnit represents a unit of code that can be tested
type TestableUnit struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	Complexity   int        `json:"complexity"`
	Dependencies []string   `json:"dependencies"`
	Parameters   []string   `json:"parameters"`
	ReturnTypes  []string   `json:"return_types"`
	Examples     []string   `json:"examples"`
	Priority     TestPriority `json:"priority"`
}

// DatabaseSchema represents database schema information
type DatabaseSchema struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Tables       []*Table      `json:"tables"`
	Relationships []*Relationship `json:"relationships"`
	Indexes      []*Index      `json:"indexes"`
	Constraints  []*Constraint `json:"constraints"`
}

// Table represents a database table
type Table struct {
	Name     string    `json:"name"`
	Schema   string    `json:"schema"`
	Columns  []*Column `json:"columns"`
	RowCount int64     `json:"row_count"`
}

// Column represents a database column
type Column struct {
	Name         string      `json:"name"`
	DataType     string      `json:"data_type"`
	Nullable     bool        `json:"nullable"`
	DefaultValue interface{} `json:"default_value"`
	IsPrimaryKey bool        `json:"is_primary_key"`
	IsForeignKey bool        `json:"is_foreign_key"`
	References   string      `json:"references,omitempty"`
}

// DataProfile represents analyzed data patterns
type DataProfile struct {
	SchemaID         string                 `json:"schema_id"`
	TableProfiles    []*TableProfile        `json:"table_profiles"`
	Relationships    []*DataRelationship    `json:"relationships"`
	DataPatterns     []*DataPattern         `json:"data_patterns"`
	QualityMetrics   *DataQualityMetrics    `json:"quality_metrics"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// TableProfile represents data patterns for a table
type TableProfile struct {
	TableName      string               `json:"table_name"`
	RowCount       int64                `json:"row_count"`
	ColumnProfiles []*ColumnProfile     `json:"column_profiles"`
	DataSample     []map[string]interface{} `json:"data_sample"`
}

// ColumnProfile represents data patterns for a column
type ColumnProfile struct {
	ColumnName   string      `json:"column_name"`
	DataType     string      `json:"data_type"`
	NullRate     float64     `json:"null_rate"`
	UniqueRate   float64     `json:"unique_rate"`
	MinValue     interface{} `json:"min_value"`
	MaxValue     interface{} `json:"max_value"`
	AvgLength    float64     `json:"avg_length"`
	Pattern      string      `json:"pattern"`
	Distribution map[string]int `json:"distribution"`
}

// TestDataSet represents generated test data
type TestDataSet struct {
	ID          string                         `json:"id"`
	SchemaID    string                         `json:"schema_id"`
	Volume      int                            `json:"volume"`
	Tables      map[string][]map[string]interface{} `json:"tables"`
	CreatedAt   time.Time                      `json:"created_at"`
	Metadata    map[string]interface{}         `json:"metadata"`
}

// Performance testing models

// BaselineConfig represents configuration for performance baseline
type BaselineConfig struct {
	Name         string                 `json:"name"`
	Target       string                 `json:"target"`
	Duration     time.Duration          `json:"duration"`
	Metrics      []string               `json:"metrics"`
	Thresholds   map[string]float64     `json:"thresholds"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PerformanceBaseline represents a performance baseline
type PerformanceBaseline struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	CreatedAt    time.Time              `json:"created_at"`
	Metrics      map[string]float64     `json:"metrics"`
	Thresholds   map[string]float64     `json:"thresholds"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// LoadTestConfig represents configuration for load testing
type LoadTestConfig struct {
	Name            string                 `json:"name"`
	Target          string                 `json:"target"`
	Pattern         LoadPattern            `json:"pattern"`
	MaxUsers        int                    `json:"max_users"`
	Duration        time.Duration          `json:"duration"`
	RampUpTime      time.Duration          `json:"ramp_up_time"`
	Scenarios       []*LoadTestScenario    `json:"scenarios"`
	Thresholds      map[string]float64     `json:"thresholds"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// LoadTestScenario represents a load test scenario
type LoadTestScenario struct {
	Name        string    `json:"name"`
	Weight      float64   `json:"weight"`
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string    `json:"body"`
	ThinkTime   time.Duration `json:"think_time"`
}

// LoadTestResults represents load test results
type LoadTestResults struct {
	ConfigID         string                 `json:"config_id"`
	ExecutionID      string                 `json:"execution_id"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	TotalRequests    int64                  `json:"total_requests"`
	SuccessfulRequests int64                `json:"successful_requests"`
	FailedRequests   int64                  `json:"failed_requests"`
	AverageResponseTime time.Duration       `json:"average_response_time"`
	P95ResponseTime  time.Duration          `json:"p95_response_time"`
	P99ResponseTime  time.Duration          `json:"p99_response_time"`
	Throughput       float64                `json:"throughput"`
	ErrorRate        float64                `json:"error_rate"`
	ResourceMetrics  *ResourceMetrics       `json:"resource_metrics"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Environment management models

// TestEnvironment represents a test environment
type TestEnvironment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        EnvironmentType        `json:"type"`
	Status      EnvironmentStatus      `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	Config      *EnvironmentConfig     `json:"config"`
	Resources   *EnvironmentResources  `json:"resources"`
	Endpoints   map[string]string      `json:"endpoints"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// EnvironmentResources represents environment resource allocation
type EnvironmentResources struct {
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	Storage   string `json:"storage"`
	Network   string `json:"network"`
	Containers int   `json:"containers"`
}

// Analysis models

// ResultClassification represents ML-based result classification
type ResultClassification struct {
	Category    string                 `json:"category"`
	Confidence  float64                `json:"confidence"`
	Tags        []string               `json:"tags"`
	Patterns    []string               `json:"patterns"`
	Anomalies   []string               `json:"anomalies"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TrendAnalysis represents long-term testing trends
type TrendAnalysis struct {
	TimeRange      TimeRange              `json:"time_range"`
	QualityTrend   string                 `json:"quality_trend"`
	PerformanceTrend string               `json:"performance_trend"`
	CoverageTrend  string                 `json:"coverage_trend"`
	Predictions    map[string]float64     `json:"predictions"`
	Recommendations []string              `json:"recommendations"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RiskAssessment represents production risk assessment
type RiskAssessment struct {
	OverallRisk     RiskLevel              `json:"overall_risk"`
	RiskFactors     []*RiskFactor          `json:"risk_factors"`
	Mitigations     []string               `json:"mitigations"`
	Confidence      float64                `json:"confidence"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	Name        string    `json:"name"`
	Level       RiskLevel `json:"level"`
	Probability float64   `json:"probability"`
	Impact      float64   `json:"impact"`
	Description string    `json:"description"`
}

// TestReport represents a comprehensive test report
type TestReport struct {
	ID               string                 `json:"id"`
	ExecutionID      string                 `json:"execution_id"`
	GeneratedAt      time.Time              `json:"generated_at"`
	Summary          *ReportSummary         `json:"summary"`
	DetailedResults  *TestResults           `json:"detailed_results"`
	Analysis         *TestAnalysis          `json:"analysis"`
	Visualizations   []*Visualization       `json:"visualizations"`
	Attachments      []string               `json:"attachments"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Supporting types and enums

type TestTargetType string
const (
	TestTargetTypeAPI      TestTargetType = "api"
	TestTargetTypeUI       TestTargetType = "ui"
	TestTargetTypeDatabase TestTargetType = "database"
	TestTargetTypeService  TestTargetType = "service"
	TestTargetTypeIntegration TestTargetType = "integration"
)

type TestCaseType string
const (
	TestCaseTypeUnit        TestCaseType = "unit"
	TestCaseTypeIntegration TestCaseType = "integration"
	TestCaseTypeE2E         TestCaseType = "e2e"
	TestCaseTypePerformance TestCaseType = "performance"
	TestCaseTypeSecurity    TestCaseType = "security"
)

type TestPriority string
const (
	TestPriorityLow    TestPriority = "low"
	TestPriorityMedium TestPriority = "medium"
	TestPriorityHigh   TestPriority = "high"
	TestPriorityCritical TestPriority = "critical"
)

type TestStatus string
const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusError   TestStatus = "error"
)

type RiskLevel string
const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type LoadPattern string
const (
	LoadPatternConstant   LoadPattern = "constant"
	LoadPatternRampUp     LoadPattern = "ramp_up"
	LoadPatternSpike      LoadPattern = "spike"
	LoadPatternStress     LoadPattern = "stress"
	LoadPatternAdaptive   LoadPattern = "adaptive"
)

type EnvironmentType string
const (
	EnvironmentTypeDevelopment EnvironmentType = "development"
	EnvironmentTypeStaging     EnvironmentType = "staging"
	EnvironmentTypeProduction  EnvironmentType = "production"
	EnvironmentTypeIsolated    EnvironmentType = "isolated"
)

type EnvironmentStatus string
const (
	EnvironmentStatusCreating   EnvironmentStatus = "creating"
	EnvironmentStatusReady      EnvironmentStatus = "ready"
	EnvironmentStatusInUse      EnvironmentStatus = "in_use"
	EnvironmentStatusDestroying EnvironmentStatus = "destroying"
	EnvironmentStatusFailed     EnvironmentStatus = "failed"
)

// Additional helper types

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type TestMetrics struct {
	ExecutionTime   time.Duration `json:"execution_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	CPUUsage        float64       `json:"cpu_usage"`
	NetworkTraffic  int64         `json:"network_traffic"`
	DatabaseQueries int           `json:"database_queries"`
}

type ResourceMetrics struct {
	Timestamp      time.Time              `json:"timestamp"`
	CPUUsage       float64                `json:"cpu_usage"`
	MemoryUsage    int64                  `json:"memory_usage"`
	DiskUsage      int64                  `json:"disk_usage"`
	NetworkIn      int64                  `json:"network_in"`
	NetworkOut     int64                  `json:"network_out"`
	DatabaseConnections int               `json:"database_connections"`
	CustomMetrics  map[string]float64     `json:"custom_metrics"`
}

type Bottleneck struct {
	Type        string    `json:"type"`
	Component   string    `json:"component"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	DetectedAt  time.Time `json:"detected_at"`
}

type TestingMetrics struct {
	TotalTestsExecuted    int64     `json:"total_tests_executed"`
	AverageExecutionTime  time.Duration `json:"average_execution_time"`
	SuccessRate           float64   `json:"success_rate"`
	CoveragePercentage    float64   `json:"coverage_percentage"`
	EnvironmentsActive    int       `json:"environments_active"`
	ResourceUtilization   float64   `json:"resource_utilization"`
	AIAccuracy            float64   `json:"ai_accuracy"`
	GeneratedTestsPerDay  int       `json:"generated_tests_per_day"`
}

type ReportSummary struct {
	TotalTests      int       `json:"total_tests"`
	PassedTests     int       `json:"passed_tests"`
	FailedTests     int       `json:"failed_tests"`
	SkippedTests    int       `json:"skipped_tests"`
	SuccessRate     float64   `json:"success_rate"`
	ExecutionTime   time.Duration `json:"execution_time"`
	CoveragePercent float64   `json:"coverage_percent"`
	QualityScore    float64   `json:"quality_score"`
	RiskLevel       RiskLevel `json:"risk_level"`
}

type Visualization struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        interface{}            `json:"data"`
	Config      map[string]interface{} `json:"config"`
}

// Configuration structures for each component

type AIConfig struct {
	ModelPath            string  `json:"model_path"`
	ConfidenceThreshold  float64 `json:"confidence_threshold"`
	MaxGeneratedTests    int     `json:"max_generated_tests"`
	LearningRate         float64 `json:"learning_rate"`
	EnableContinuousLearning bool `json:"enable_continuous_learning"`
}

type DataConfig struct {
	MaxDatasetSize       int     `json:"max_dataset_size"`
	AnonymizationLevel   string  `json:"anonymization_level"`
	PreservePIIPatterns  bool    `json:"preserve_pii_patterns"`
	EnableSyntheticData  bool    `json:"enable_synthetic_data"`
	DataQualityThreshold float64 `json:"data_quality_threshold"`
}

type PerformanceConfig struct {
	DefaultDuration      time.Duration `json:"default_duration"`
	MaxConcurrentUsers   int           `json:"max_concurrent_users"`
	ResourceMonitoringInterval time.Duration `json:"resource_monitoring_interval"`
	EnableBottleneckDetection bool      `json:"enable_bottleneck_detection"`
	PerformanceThresholds map[string]float64 `json:"performance_thresholds"`
}

type AnalysisConfig struct {
	MLModelPath          string  `json:"ml_model_path"`
	TrendAnalysisWindow  time.Duration `json:"trend_analysis_window"`
	RiskThreshold        float64 `json:"risk_threshold"`
	EnablePredictiveAnalysis bool `json:"enable_predictive_analysis"`
	GenerateVisualization bool   `json:"generate_visualizations"`
}

// Additional data structures

type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Path    string `json:"path"`
}

type RiskArea struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Severity    RiskLevel `json:"severity"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
}

type CoverageGap struct {
	Type        string `json:"type"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Priority    TestPriority `json:"priority"`
}

type CodePattern struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Occurrences int      `json:"occurrences"`
	Examples    []string `json:"examples"`
	TestStrategy string  `json:"test_strategy"`
}

type Relationship struct {
	Type       string `json:"type"`
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
}

type Index struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
	Unique  bool     `json:"unique"`
}

type Constraint struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Table  string `json:"table"`
	Column string `json:"column"`
	Rule   string `json:"rule"`
}

type DataRelationship struct {
	ParentTable string  `json:"parent_table"`
	ChildTable  string  `json:"child_table"`
	Cardinality string  `json:"cardinality"`
	Strength    float64 `json:"strength"`
}

type DataPattern struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Frequency   float64 `json:"frequency"`
	Examples    []string `json:"examples"`
}

type DataQualityMetrics struct {
	Completeness float64 `json:"completeness"`
	Accuracy     float64 `json:"accuracy"`
	Consistency  float64 `json:"consistency"`
	Timeliness   float64 `json:"timeliness"`
	Validity     float64 `json:"validity"`
	Uniqueness   float64 `json:"uniqueness"`
}