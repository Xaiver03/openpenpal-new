package performance

import (
	"context"
	"time"
)

// PerformanceTestingEngine is the main interface for performance testing capabilities
type PerformanceTestingEngine interface {
	// Core performance testing operations
	RunLoadTest(ctx context.Context, config *LoadTestConfig) (*LoadTestResult, error)
	RunStressTest(ctx context.Context, config *StressTestConfig) (*StressTestResult, error)
	RunSpikeTest(ctx context.Context, config *SpikeTestConfig) (*SpikeTestResult, error)
	RunVolumeTest(ctx context.Context, config *VolumeTestConfig) (*VolumeTestResult, error)
	RunEnduranceTest(ctx context.Context, config *EnduranceTestConfig) (*EnduranceTestResult, error)
	
	// Advanced testing capabilities
	RunAdaptiveTest(ctx context.Context, config *AdaptiveTestConfig) (*AdaptiveTestResult, error)
	RunChaosTest(ctx context.Context, config *ChaosTestConfig) (*ChaosTestResult, error)
	
	// Baseline and benchmarking
	EstablishBaseline(ctx context.Context, config *BaselineConfig) (*PerformanceBaseline, error)
	CompareWithBaseline(ctx context.Context, result *TestResult, baseline *PerformanceBaseline) (*ComparisonReport, error)
	
	// Monitoring and analysis
	GetRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error)
	AnalyzeBottlenecks(ctx context.Context, metrics *PerformanceMetrics) (*BottleneckAnalysis, error)
	PredictResourceUsage(ctx context.Context, testConfig *TestConfig) (*ResourcePrediction, error)
}

// LoadPatternRecognizer analyzes and identifies load patterns using AI
type LoadPatternRecognizer interface {
	AnalyzeTrafficPatterns(ctx context.Context, trafficData *TrafficData) (*PatternAnalysis, error)
	DetectSeasonality(ctx context.Context, historicalData *HistoricalMetrics) (*SeasonalityReport, error)
	PredictLoadPatterns(ctx context.Context, baselineData *BaselineData) (*LoadPrediction, error)
	GenerateRealisticLoadProfile(ctx context.Context, requirements *LoadRequirements) (*LoadProfile, error)
	OptimizeTestScenarios(ctx context.Context, scenarios []*TestScenario) (*OptimizedScenarios, error)
}

// PerformanceBaselineManager manages performance baselines and benchmarks
type PerformanceBaselineManager interface {
	CreateBaseline(ctx context.Context, config *BaselineConfig) (*PerformanceBaseline, error)
	UpdateBaseline(ctx context.Context, baselineID string, metrics *PerformanceMetrics) error
	GetBaseline(ctx context.Context, baselineID string) (*PerformanceBaseline, error)
	CompareBaselines(ctx context.Context, baseline1, baseline2 *PerformanceBaseline) (*BaselineComparison, error)
	ValidateBaseline(ctx context.Context, baseline *PerformanceBaseline) (*ValidationResult, error)
	
	// Baseline analytics
	AnalyzeBaselineTrends(ctx context.Context, baselineID string, timeRange *TimeRange) (*TrendAnalysis, error)
	DetectPerformanceRegression(ctx context.Context, current, baseline *PerformanceMetrics) (*RegressionReport, error)
	RecommendBaselineUpdates(ctx context.Context, baseline *PerformanceBaseline) (*UpdateRecommendations, error)
}

// BottleneckDetector identifies and analyzes performance bottlenecks
type BottleneckDetector interface {
	DetectBottlenecks(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error)
	AnalyzeRootCause(ctx context.Context, bottleneck *Bottleneck) (*RootCauseAnalysis, error)
	PrioritizeBottlenecks(ctx context.Context, bottlenecks []*Bottleneck) ([]*PrioritizedBottleneck, error)
	RecommendOptimizations(ctx context.Context, bottleneck *Bottleneck) ([]*OptimizationRecommendation, error)
	
	// Advanced analysis
	AnalyzeResourceContention(ctx context.Context, metrics *ResourceMetrics) (*ContentionAnalysis, error)
	DetectMemoryLeaks(ctx context.Context, memoryMetrics *MemoryMetrics) (*MemoryLeakReport, error)
	AnalyzeCPUHotspots(ctx context.Context, cpuMetrics *CPUMetrics) (*HotspotAnalysis, error)
	DetectDatabaseBottlenecks(ctx context.Context, dbMetrics *DatabaseMetrics) ([]*DatabaseBottleneck, error)
}

// ResourcePredictor predicts resource usage patterns
type ResourcePredictor interface {
	PredictCPUUsage(ctx context.Context, loadProfile *LoadProfile) (*CPUPrediction, error)
	PredictMemoryUsage(ctx context.Context, loadProfile *LoadProfile) (*MemoryPrediction, error)
	PredictNetworkUsage(ctx context.Context, loadProfile *LoadProfile) (*NetworkPrediction, error)
	PredictDiskUsage(ctx context.Context, loadProfile *LoadProfile) (*DiskPrediction, error)
	
	// Comprehensive predictions
	PredictSystemResources(ctx context.Context, testConfig *TestConfig) (*SystemResourcePrediction, error)
	PredictScalingRequirements(ctx context.Context, targetLoad *LoadTarget) (*ScalingPrediction, error)
	OptimizeResourceAllocation(ctx context.Context, prediction *ResourcePrediction) (*AllocationOptimization, error)
}

// MetricsCollector collects and aggregates performance metrics
type MetricsCollector interface {
	StartCollection(ctx context.Context, config *CollectionConfig) error
	StopCollection(ctx context.Context) (*CollectedMetrics, error)
	GetCurrentMetrics(ctx context.Context) (*RealTimeMetrics, error)
	
	// Metric aggregation
	AggregateMetrics(ctx context.Context, metrics []*MetricPoint, interval time.Duration) (*AggregatedMetrics, error)
	CalculatePercentiles(ctx context.Context, values []float64) (*PercentileMetrics, error)
	DetectAnomalies(ctx context.Context, metrics *MetricStream) ([]*Anomaly, error)
}

// PerformanceReporter generates performance test reports
type PerformanceReporter interface {
	GenerateLoadTestReport(ctx context.Context, result *LoadTestResult) (*PerformanceReport, error)
	GenerateComparisonReport(ctx context.Context, results []*TestResult) (*ComparisonReport, error)
	GenerateExecutiveSummary(ctx context.Context, results []*TestResult) (*ExecutiveSummary, error)
	GenerateTrendReport(ctx context.Context, historicalResults []*TestResult) (*TrendReport, error)
	
	// Export capabilities
	ExportToJSON(ctx context.Context, report *PerformanceReport) ([]byte, error)
	ExportToHTML(ctx context.Context, report *PerformanceReport) ([]byte, error)
	ExportToPDF(ctx context.Context, report *PerformanceReport) ([]byte, error)
	ExportToCSV(ctx context.Context, metrics *PerformanceMetrics) ([]byte, error)
}

// Data Models

// LoadTestConfig configures load testing parameters
type LoadTestConfig struct {
	TestName        string        `json:"test_name"`
	TargetURL       string        `json:"target_url"`
	VirtualUsers    int           `json:"virtual_users"`
	Duration        time.Duration `json:"duration"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	RampDownTime    time.Duration `json:"ramp_down_time"`
	RequestsPerSec  int           `json:"requests_per_sec"`
	ThinkTime       time.Duration `json:"think_time"`
	
	// Advanced configuration
	LoadPattern     LoadPattern     `json:"load_pattern"`
	RequestMix      []*RequestType  `json:"request_mix"`
	DataVariation   *DataVariation  `json:"data_variation"`
	NetworkProfile  *NetworkProfile `json:"network_profile"`
	
	// Thresholds and SLAs
	PerformanceThresholds *PerformanceThresholds `json:"performance_thresholds"`
	ResourceLimits        *ResourceLimits        `json:"resource_limits"`
}

// LoadPattern defines the pattern of load application
type LoadPattern string

const (
	LoadPatternConstant    LoadPattern = "constant"     // Steady load
	LoadPatternRamp        LoadPattern = "ramp"         // Gradual increase
	LoadPatternSpike       LoadPattern = "spike"        // Sudden increase
	LoadPatternStep        LoadPattern = "step"         // Step-wise increase
	LoadPatternWave        LoadPattern = "wave"         // Sine wave pattern
	LoadPatternRandom      LoadPattern = "random"       // Random variations
	LoadPatternRealistic   LoadPattern = "realistic"    // AI-generated realistic pattern
)

// RequestType defines different types of requests in the load mix
type RequestType struct {
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	Weight      float64           `json:"weight"`      // Percentage of total requests
	Complexity  int               `json:"complexity"`  // Computational complexity (1-10)
}

// DataVariation defines how test data varies during testing
type DataVariation struct {
	UserData      *UserDataConfig      `json:"user_data"`
	PayloadData   *PayloadDataConfig   `json:"payload_data"`
	SessionData   *SessionDataConfig   `json:"session_data"`
	VariationRate float64              `json:"variation_rate"` // How often data changes
}

// NetworkProfile simulates different network conditions
type NetworkProfile struct {
	Bandwidth    int           `json:"bandwidth_kbps"`
	Latency      time.Duration `json:"latency"`
	PacketLoss   float64       `json:"packet_loss_percent"`
	Jitter       time.Duration `json:"jitter"`
	NetworkType  NetworkType   `json:"network_type"`
}

// NetworkType defines different network conditions
type NetworkType string

const (
	NetworkTypeFiber     NetworkType = "fiber"      // High-speed fiber
	NetworkTypeBroadband NetworkType = "broadband"  // Standard broadband
	NetworkType4G        NetworkType = "4g"         // Mobile 4G
	NetworkType3G        NetworkType = "3g"         // Mobile 3G
	NetworkTypeWiFi      NetworkType = "wifi"       // WiFi connection
	NetworkTypeSatellite NetworkType = "satellite"  // Satellite connection
)

// PerformanceThresholds defines acceptable performance limits
type PerformanceThresholds struct {
	MaxResponseTime    time.Duration `json:"max_response_time"`
	MaxErrorRate       float64       `json:"max_error_rate_percent"`
	MinThroughput      float64       `json:"min_throughput_rps"`
	MaxCPUUsage        float64       `json:"max_cpu_usage_percent"`
	MaxMemoryUsage     float64       `json:"max_memory_usage_percent"`
	MaxDiskUsage       float64       `json:"max_disk_usage_percent"`
	MaxNetworkUsage    float64       `json:"max_network_usage_mbps"`
	
	// SLA definitions
	AvailabilityTarget float64       `json:"availability_target_percent"`
	P95ResponseTime    time.Duration `json:"p95_response_time"`
	P99ResponseTime    time.Duration `json:"p99_response_time"`
}

// ResourceLimits defines maximum resource usage during testing
type ResourceLimits struct {
	MaxCPUCores    int   `json:"max_cpu_cores"`
	MaxMemoryMB    int64 `json:"max_memory_mb"`
	MaxDiskSpaceGB int64 `json:"max_disk_space_gb"`
	MaxNetworkMbps int   `json:"max_network_mbps"`
	MaxFileHandles int   `json:"max_file_handles"`
	MaxConnections int   `json:"max_connections"`
}

// LoadTestResult contains the results of a load test
type LoadTestResult struct {
	TestID          string                 `json:"test_id"`
	TestName        string                 `json:"test_name"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Status          TestStatus             `json:"status"`
	
	// Core metrics
	TotalRequests   int64                  `json:"total_requests"`
	SuccessfulReqs  int64                  `json:"successful_requests"`
	FailedRequests  int64                  `json:"failed_requests"`
	ErrorRate       float64                `json:"error_rate_percent"`
	Throughput      float64                `json:"throughput_rps"`
	
	// Response time metrics
	ResponseTimes   *ResponseTimeMetrics   `json:"response_times"`
	
	// Resource utilization
	ResourceUsage   *ResourceUsageMetrics  `json:"resource_usage"`
	
	// Detailed analysis
	BottleneckAnalysis *BottleneckAnalysis  `json:"bottleneck_analysis"`
	ErrorAnalysis      *ErrorAnalysis       `json:"error_analysis"`
	PerformanceScore   float64              `json:"performance_score"`
	
	// Threshold violations
	ThresholdViolations []*ThresholdViolation `json:"threshold_violations"`
	
	// Raw data
	TimeSeries      []*MetricPoint         `json:"time_series"`
	RequestDetails  []*RequestDetail       `json:"request_details"`
}

// TestStatus represents the status of a performance test
type TestStatus string

const (
	TestStatusPending    TestStatus = "pending"
	TestStatusRunning    TestStatus = "running"
	TestStatusCompleted  TestStatus = "completed"
	TestStatusFailed     TestStatus = "failed"
	TestStatusCancelled  TestStatus = "cancelled"
	TestStatusTimeout    TestStatus = "timeout"
)

// ResponseTimeMetrics contains response time statistics
type ResponseTimeMetrics struct {
	Mean       time.Duration `json:"mean"`
	Median     time.Duration `json:"median"`
	Min        time.Duration `json:"min"`
	Max        time.Duration `json:"max"`
	StdDev     time.Duration `json:"std_dev"`
	
	// Percentiles
	P50        time.Duration `json:"p50"`
	P75        time.Duration `json:"p75"`
	P90        time.Duration `json:"p90"`
	P95        time.Duration `json:"p95"`
	P99        time.Duration `json:"p99"`
	P999       time.Duration `json:"p999"`
}

// ResourceUsageMetrics contains system resource utilization data
type ResourceUsageMetrics struct {
	CPU     *CPUMetrics     `json:"cpu"`
	Memory  *MemoryMetrics  `json:"memory"`
	Disk    *DiskMetrics    `json:"disk"`
	Network *NetworkMetrics `json:"network"`
}

// CPUMetrics contains CPU utilization information
type CPUMetrics struct {
	AverageUsage float64            `json:"average_usage_percent"`
	PeakUsage    float64            `json:"peak_usage_percent"`
	CoreUsage    map[string]float64 `json:"core_usage"`
	LoadAverage  *LoadAverage       `json:"load_average"`
	ContextSwitches int64           `json:"context_switches"`
	Interrupts   int64              `json:"interrupts"`
}

// LoadAverage contains system load average information
type LoadAverage struct {
	Load1Min  float64 `json:"load_1_min"`
	Load5Min  float64 `json:"load_5_min"`
	Load15Min float64 `json:"load_15_min"`
}

// MemoryMetrics contains memory utilization information
type MemoryMetrics struct {
	TotalMemory     int64   `json:"total_memory_bytes"`
	UsedMemory      int64   `json:"used_memory_bytes"`
	FreeMemory      int64   `json:"free_memory_bytes"`
	CachedMemory    int64   `json:"cached_memory_bytes"`
	BufferedMemory  int64   `json:"buffered_memory_bytes"`
	SwapUsed        int64   `json:"swap_used_bytes"`
	SwapTotal       int64   `json:"swap_total_bytes"`
	MemoryPressure  float64 `json:"memory_pressure"`
	
	// Application-specific metrics
	HeapSize        int64   `json:"heap_size_bytes"`
	HeapUsed        int64   `json:"heap_used_bytes"`
	GCCount         int64   `json:"gc_count"`
	GCTime          time.Duration `json:"gc_time"`
}

// DiskMetrics contains disk utilization information
type DiskMetrics struct {
	ReadOps         int64   `json:"read_ops"`
	WriteOps        int64   `json:"write_ops"`
	ReadBytes       int64   `json:"read_bytes"`
	WriteBytes      int64   `json:"write_bytes"`
	ReadLatency     time.Duration `json:"read_latency"`
	WriteLatency    time.Duration `json:"write_latency"`
	DiskUsage       float64 `json:"disk_usage_percent"`
	InodeUsage      float64 `json:"inode_usage_percent"`
	QueueDepth      int     `json:"queue_depth"`
}

// NetworkMetrics contains network utilization information
type NetworkMetrics struct {
	BytesReceived   int64   `json:"bytes_received"`
	BytesSent       int64   `json:"bytes_sent"`
	PacketsReceived int64   `json:"packets_received"`
	PacketsSent     int64   `json:"packets_sent"`
	PacketLoss      float64 `json:"packet_loss_percent"`
	Latency         time.Duration `json:"latency"`
	Bandwidth       int64   `json:"bandwidth_bps"`
	Connections     int     `json:"active_connections"`
	ErrorRate       float64 `json:"error_rate_percent"`
}

// BottleneckAnalysis contains bottleneck detection results
type BottleneckAnalysis struct {
	DetectedBottlenecks []*Bottleneck           `json:"detected_bottlenecks"`
	CriticalPath        []*PathSegment          `json:"critical_path"`
	ResourceContention  *ContentionAnalysis     `json:"resource_contention"`
	Recommendations     []*OptimizationRecommendation `json:"recommendations"`
	Severity            BottleneckSeverity      `json:"severity"`
}

// Bottleneck represents a detected performance bottleneck
type Bottleneck struct {
	ID            string             `json:"id"`
	Type          BottleneckType     `json:"type"`
	Component     string             `json:"component"`
	Description   string             `json:"description"`
	Severity      BottleneckSeverity `json:"severity"`
	Impact        float64            `json:"impact_percent"`
	Location      *BottleneckLocation `json:"location"`
	Metrics       map[string]float64 `json:"metrics"`
	RootCause     *RootCauseAnalysis `json:"root_cause"`
	DetectedAt    time.Time          `json:"detected_at"`
}

// BottleneckType defines different types of bottlenecks
type BottleneckType string

const (
	BottleneckTypeCPU      BottleneckType = "cpu"
	BottleneckTypeMemory   BottleneckType = "memory"
	BottleneckTypeDisk     BottleneckType = "disk"
	BottleneckTypeNetwork  BottleneckType = "network"
	BottleneckTypeDatabase BottleneckType = "database"
	BottleneckTypeAPI      BottleneckType = "api"
	BottleneckTypeCache    BottleneckType = "cache"
	BottleneckTypeLock     BottleneckType = "lock"
	BottleneckTypeQueue    BottleneckType = "queue"
)

// BottleneckSeverity defines the severity level of bottlenecks
type BottleneckSeverity string

const (
	SeverityLow      BottleneckSeverity = "low"
	SeverityMedium   BottleneckSeverity = "medium"
	SeverityHigh     BottleneckSeverity = "high"
	SeverityCritical BottleneckSeverity = "critical"
)

// BottleneckLocation provides location information for bottlenecks
type BottleneckLocation struct {
	Service   string `json:"service"`
	Component string `json:"component"`
	Function  string `json:"function"`
	Line      int    `json:"line"`
	File      string `json:"file"`
}

// RootCauseAnalysis contains root cause analysis results
type RootCauseAnalysis struct {
	PrimaryCause    string                 `json:"primary_cause"`
	ContributingFactors []string           `json:"contributing_factors"`
	Likelihood      float64                `json:"likelihood"`
	Evidence        []*Evidence            `json:"evidence"`
	Recommendations []*OptimizationRecommendation `json:"recommendations"`
}

// Evidence represents evidence supporting root cause analysis
type Evidence struct {
	Type        EvidenceType    `json:"type"`
	Description string          `json:"description"`
	Data        interface{}     `json:"data"`
	Confidence  float64         `json:"confidence"`
	Source      string          `json:"source"`
}

// EvidenceType defines types of evidence
type EvidenceType string

const (
	EvidenceTypeMetric      EvidenceType = "metric"
	EvidenceTypeLog         EvidenceType = "log"
	EvidenceTypeTrace       EvidenceType = "trace"
	EvidenceTypeProfile     EvidenceType = "profile"
	EvidenceTypeStatistical EvidenceType = "statistical"
)

// OptimizationRecommendation provides optimization suggestions
type OptimizationRecommendation struct {
	ID             string                    `json:"id"`
	Title          string                    `json:"title"`
	Description    string                    `json:"description"`
	Category       RecommendationCategory    `json:"category"`
	Priority       RecommendationPriority    `json:"priority"`
	Impact         float64                   `json:"expected_impact_percent"`
	Effort         RecommendationEffort      `json:"implementation_effort"`
	Confidence     float64                   `json:"confidence"`
	Dependencies   []string                  `json:"dependencies"`
	Steps          []*ImplementationStep     `json:"implementation_steps"`
	EstimatedCost  *CostEstimate            `json:"estimated_cost"`
	RiskAssessment *RiskAssessment          `json:"risk_assessment"`
}

// RecommendationCategory defines categories of recommendations
type RecommendationCategory string

const (
	CategoryInfrastructure RecommendationCategory = "infrastructure"
	CategoryApplication    RecommendationCategory = "application"
	CategoryDatabase       RecommendationCategory = "database"
	CategoryCaching        RecommendationCategory = "caching"
	CategoryNetwork        RecommendationCategory = "network"
	CategoryConfiguration  RecommendationCategory = "configuration"
	CategoryCode           RecommendationCategory = "code"
)

// RecommendationPriority defines priority levels
type RecommendationPriority string

const (
	PriorityLow      RecommendationPriority = "low"
	PriorityMedium   RecommendationPriority = "medium"
	PriorityHigh     RecommendationPriority = "high"
	PriorityCritical RecommendationPriority = "critical"
)

// RecommendationEffort defines implementation effort levels
type RecommendationEffort string

const (
	EffortLow    RecommendationEffort = "low"     // < 1 day
	EffortMedium RecommendationEffort = "medium"  // 1-5 days
	EffortHigh   RecommendationEffort = "high"    // 1-4 weeks
	EffortMajor  RecommendationEffort = "major"   // > 1 month
)

// ImplementationStep defines a step in implementing a recommendation
type ImplementationStep struct {
	StepNumber  int           `json:"step_number"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"estimated_duration"`
	Risk        string        `json:"risk_level"`
	Dependencies []string     `json:"dependencies"`
}

// CostEstimate provides cost estimation for recommendations
type CostEstimate struct {
	DevelopmentHours int     `json:"development_hours"`
	InfrastructureCost float64 `json:"infrastructure_cost"`
	OperationalCost  float64 `json:"operational_cost"`
	TotalCost        float64 `json:"total_cost"`
	Currency         string  `json:"currency"`
	TimeFrame        string  `json:"time_frame"`
}

// RiskAssessment evaluates risks of implementing recommendations
type RiskAssessment struct {
	OverallRisk     RiskLevel      `json:"overall_risk"`
	TechnicalRisk   RiskLevel      `json:"technical_risk"`
	BusinessRisk    RiskLevel      `json:"business_risk"`
	SecurityRisk    RiskLevel      `json:"security_risk"`
	Mitigations     []*Mitigation  `json:"mitigations"`
}

// RiskLevel defines risk levels
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// Mitigation defines a risk mitigation strategy
type Mitigation struct {
	Risk        string `json:"risk"`
	Strategy    string `json:"strategy"`
	Probability float64 `json:"probability_reduction"`
	Impact      float64 `json:"impact_reduction"`
}

// Additional configuration types for different test types

// StressTestConfig extends LoadTestConfig for stress testing
type StressTestConfig struct {
	*LoadTestConfig
	MaxUsers        int           `json:"max_users"`
	UserIncrement   int           `json:"user_increment"`
	IncrementInterval time.Duration `json:"increment_interval"`
	FailureThreshold float64       `json:"failure_threshold_percent"`
	RecoveryTime    time.Duration `json:"recovery_time"`
}

// SpikeTestConfig extends LoadTestConfig for spike testing
type SpikeTestConfig struct {
	*LoadTestConfig
	SpikeUsers      int           `json:"spike_users"`
	SpikeDuration   time.Duration `json:"spike_duration"`
	SpikeInterval   time.Duration `json:"spike_interval"`
	SpikeCount      int           `json:"spike_count"`
	BaselineUsers   int           `json:"baseline_users"`
}

// VolumeTestConfig configures volume testing with large datasets
type VolumeTestConfig struct {
	*LoadTestConfig
	DataVolume      int64         `json:"data_volume_bytes"`
	RecordCount     int64         `json:"record_count"`
	DatabaseSize    int64         `json:"database_size_bytes"`
	FileSize        int64         `json:"file_size_bytes"`
	TransactionSize int           `json:"transaction_size"`
}

// EnduranceTestConfig configures long-running endurance tests
type EnduranceTestConfig struct {
	*LoadTestConfig
	ExtendedDuration    time.Duration `json:"extended_duration"`
	MemoryLeakDetection bool          `json:"memory_leak_detection"`
	ResourceMonitoring  bool          `json:"resource_monitoring"`
	CheckpointInterval  time.Duration `json:"checkpoint_interval"`
	AutoRecovery        bool          `json:"auto_recovery"`
}

// Configuration types for user data, payload data, and session data
type UserDataConfig struct {
	UserCount       int      `json:"user_count"`
	DataSource      string   `json:"data_source"`
	RotationPolicy  string   `json:"rotation_policy"`
	UniqueUsers     bool     `json:"unique_users"`
}

type PayloadDataConfig struct {
	PayloadTypes    []string `json:"payload_types"`
	SizeVariation   string   `json:"size_variation"`
	ContentTypes    []string `json:"content_types"`
	CompressionEnabled bool  `json:"compression_enabled"`
}

type SessionDataConfig struct {
	SessionDuration time.Duration `json:"session_duration"`
	SessionTypes    []string      `json:"session_types"`
	CookiePolicy    string        `json:"cookie_policy"`
	StateManagement string        `json:"state_management"`
}

// Additional result types

type StressTestResult struct {
	*LoadTestResult
	MaxUsersReached     int     `json:"max_users_reached"`
	BreakingPoint       int     `json:"breaking_point_users"`
	RecoveryTime        time.Duration `json:"recovery_time"`
	DegradationPattern  string  `json:"degradation_pattern"`
}

type SpikeTestResult struct {
	*LoadTestResult
	SpikeResults       []*SpikeResult `json:"spike_results"`
	RecoveryMetrics    *RecoveryMetrics `json:"recovery_metrics"`
	SystemStability    float64        `json:"system_stability_score"`
}

type SpikeResult struct {
	SpikeNumber     int           `json:"spike_number"`
	StartTime       time.Time     `json:"start_time"`
	Duration        time.Duration `json:"duration"`
	PeakUsers       int           `json:"peak_users"`
	ImpactMetrics   *ImpactMetrics `json:"impact_metrics"`
}

type RecoveryMetrics struct {
	RecoveryTime    time.Duration `json:"recovery_time"`
	StabilityTime   time.Duration `json:"stability_time"`
	ErrorBurst      int           `json:"error_burst"`
	PerformanceDip  float64       `json:"performance_dip_percent"`
}

type VolumeTestResult struct {
	*LoadTestResult
	DataProcessed      int64          `json:"data_processed_bytes"`
	ProcessingRate     float64        `json:"processing_rate_bps"`
	StorageEfficiency  float64        `json:"storage_efficiency"`
	DataIntegrity      *IntegrityReport `json:"data_integrity"`
}

type EnduranceTestResult struct {
	*LoadTestResult
	MemoryLeakDetected bool             `json:"memory_leak_detected"`
	ResourceDrift      *ResourceDrift   `json:"resource_drift"`
	StabilityMetrics   *StabilityMetrics `json:"stability_metrics"`
	FailurePatterns    []*FailurePattern `json:"failure_patterns"`
}

// Supporting types

type ImpactMetrics struct {
	ResponseTimeDelta time.Duration `json:"response_time_delta"`
	ThroughputDelta   float64       `json:"throughput_delta"`
	ErrorRateDelta    float64       `json:"error_rate_delta"`
	ResourceImpact    *ResourceImpact `json:"resource_impact"`
}

type ResourceImpact struct {
	CPUImpact     float64 `json:"cpu_impact_percent"`
	MemoryImpact  float64 `json:"memory_impact_percent"`
	DiskImpact    float64 `json:"disk_impact_percent"`
	NetworkImpact float64 `json:"network_impact_percent"`
}

type IntegrityReport struct {
	CorruptedRecords   int64   `json:"corrupted_records"`
	MissingRecords     int64   `json:"missing_records"`
	DuplicateRecords   int64   `json:"duplicate_records"`
	IntegrityScore     float64 `json:"integrity_score"`
}

type ResourceDrift struct {
	CPUDrift      float64 `json:"cpu_drift_percent"`
	MemoryDrift   float64 `json:"memory_drift_percent"`
	DiskDrift     float64 `json:"disk_drift_percent"`
	NetworkDrift  float64 `json:"network_drift_percent"`
}

type StabilityMetrics struct {
	MeanTimeToFailure  time.Duration `json:"mean_time_to_failure"`
	MeanTimeToRecovery time.Duration `json:"mean_time_to_recovery"`
	AvailabilityScore  float64       `json:"availability_score"`
	ReliabilityScore   float64       `json:"reliability_score"`
}

type FailurePattern struct {
	Pattern     string    `json:"pattern"`
	Frequency   int       `json:"frequency"`
	Severity    string    `json:"severity"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Impact      string    `json:"impact"`
}

// Threshold violation tracking

type ThresholdViolation struct {
	ThresholdName   string        `json:"threshold_name"`
	ExpectedValue   interface{}   `json:"expected_value"`
	ActualValue     interface{}   `json:"actual_value"`
	ViolationType   ViolationType `json:"violation_type"`
	Severity        string        `json:"severity"`
	Duration        time.Duration `json:"duration"`
	FirstViolation  time.Time     `json:"first_violation"`
	LastViolation   time.Time     `json:"last_violation"`
	ViolationCount  int           `json:"violation_count"`
}

type ViolationType string

const (
	ViolationTypeExceeded ViolationType = "exceeded"
	ViolationTypeBelow    ViolationType = "below"
	ViolationTypeEqual    ViolationType = "equal"
	ViolationTypeNotEqual ViolationType = "not_equal"
)

// Request detail tracking

type RequestDetail struct {
	RequestID     string        `json:"request_id"`
	Timestamp     time.Time     `json:"timestamp"`
	Method        string        `json:"method"`
	URL           string        `json:"url"`
	ResponseTime  time.Duration `json:"response_time"`
	ResponseCode  int           `json:"response_code"`
	ResponseSize  int64         `json:"response_size"`
	Error         string        `json:"error,omitempty"`
	UserAgent     string        `json:"user_agent"`
	SessionID     string        `json:"session_id"`
}

// Real-time metrics and monitoring

type RealTimeMetrics struct {
	Timestamp           time.Time              `json:"timestamp"`
	ActiveUsers         int                    `json:"active_users"`
	CurrentThroughput   float64                `json:"current_throughput_rps"`
	AverageResponseTime time.Duration          `json:"average_response_time"`
	ErrorRate           float64                `json:"error_rate_percent"`
	ResourceUsage       *ResourceUsageMetrics  `json:"resource_usage"`
	ActiveBottlenecks   []*Bottleneck          `json:"active_bottlenecks"`
}

type MetricPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	MetricName string     `json:"metric_name"`
	Value     float64     `json:"value"`
	Unit      string      `json:"unit"`
	Tags      map[string]string `json:"tags"`
}

type CollectionConfig struct {
	Interval        time.Duration `json:"interval"`
	MetricTypes     []string      `json:"metric_types"`
	SamplingRate    float64       `json:"sampling_rate"`
	BufferSize      int           `json:"buffer_size"`
	EnableRealTime  bool          `json:"enable_real_time"`
}

// Time series and aggregation types

type CollectedMetrics struct {
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	MetricPoints []*MetricPoint `json:"metric_points"`
	Summary      *MetricSummary `json:"summary"`
}

type MetricSummary struct {
	TotalPoints  int                    `json:"total_points"`
	MetricTypes  []string               `json:"metric_types"`
	Aggregations map[string]*Aggregation `json:"aggregations"`
}

type Aggregation struct {
	Count    int64   `json:"count"`
	Sum      float64 `json:"sum"`
	Mean     float64 `json:"mean"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	StdDev   float64 `json:"std_dev"`
	Median   float64 `json:"median"`
}

type AggregatedMetrics struct {
	TimeRange    *TimeRange                `json:"time_range"`
	Interval     time.Duration             `json:"interval"`
	Aggregations map[string][]*Aggregation `json:"aggregations"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type PercentileMetrics struct {
	P50  float64 `json:"p50"`
	P75  float64 `json:"p75"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	P999 float64 `json:"p999"`
}

// Anomaly detection

type Anomaly struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	MetricName  string    `json:"metric_name"`
	Timestamp   time.Time `json:"timestamp"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Severity    string    `json:"severity"`
	Confidence  float64   `json:"confidence"`
	Description string    `json:"description"`
}

type MetricStream struct {
	MetricName string        `json:"metric_name"`
	Points     []*MetricPoint `json:"points"`
	Window     time.Duration `json:"window"`
}