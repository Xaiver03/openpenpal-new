// Package data provides intelligent test data generation capabilities
package data

import (
	"context"
	"time"

	"openpenpal-backend/internal/platform/testing/core"
)

// SmartDataGenerator defines the interface for AI-driven test data generation
type SmartDataGenerator interface {
	// Schema Analysis
	AnalyzeSchema(ctx context.Context, schema *DatabaseSchema) (*SchemaProfile, error)
	
	// Data Generation
	GenerateTestData(ctx context.Context, profile *SchemaProfile, volume int) (*TestDataSet, error)
	GenerateSyntheticData(ctx context.Context, constraints *DataConstraints) (*SyntheticDataSet, error)
	GenerateRelationalData(ctx context.Context, relationships *RelationshipMap) (*RelationalDataSet, error)
	
	// Privacy and Security
	GeneratePrivacyCompliantData(ctx context.Context, schema *DatabaseSchema, privacyRules *PrivacyRules) (*TestDataSet, error)
	AnonymizeExistingData(ctx context.Context, data *TestDataSet, anonymizationRules *AnonymizationRules) (*TestDataSet, error)
	
	// Advanced Generation
	GenerateRealisticData(ctx context.Context, profile *SchemaProfile, patterns *DataPatterns) (*TestDataSet, error)
	GenerateBoundaryData(ctx context.Context, constraints *DataConstraints) (*BoundaryDataSet, error)
	GenerateEdgeCaseData(ctx context.Context, schema *DatabaseSchema) (*EdgeCaseDataSet, error)
	
	// Performance Testing Data
	GeneratePerformanceData(ctx context.Context, scenario *PerformanceScenario) (*PerformanceDataSet, error)
	GenerateLoadTestData(ctx context.Context, loadProfile *LoadProfile) (*LoadDataSet, error)
	
	// Quality and Validation
	ValidateGeneratedData(ctx context.Context, data *TestDataSet, rules *ValidationRules) (*ValidationResults, error)
	OptimizeDataDistribution(ctx context.Context, data *TestDataSet) (*TestDataSet, error)
	
	// Learning and Adaptation
	LearnFromDataUsage(ctx context.Context, usageMetrics *DataUsageMetrics) error
	UpdateGenerationStrategies(ctx context.Context, feedback *GenerationFeedback) error
	
	// Statistics and Monitoring
	GetGenerationStats() *DataGenerationStats
	GetDataQualityMetrics() *DataQualityMetrics
}

// DatabaseSchemaAnalyzer analyzes database schemas for intelligent data generation
type DatabaseSchemaAnalyzer interface {
	AnalyzeSchema(ctx context.Context, schema *DatabaseSchema) (*SchemaProfile, error)
	DetectRelationships(ctx context.Context, tables []*Table) (*RelationshipMap, error)
	IdentifyConstraints(ctx context.Context, table *Table) (*TableConstraints, error)
	AnalyzeDataDistribution(ctx context.Context, sampleData *SampleData) (*DataDistribution, error)
	GetSchemaComplexity(schema *DatabaseSchema) *ComplexityMetrics
}

// SyntheticDataEngine generates realistic synthetic data using ML algorithms
type SyntheticDataEngine interface {
	TrainGenerationModel(ctx context.Context, trainingData *TrainingDataSet) (*GenerationModel, error)
	GenerateFromModel(ctx context.Context, model *GenerationModel, constraints *DataConstraints) (*SyntheticDataSet, error)
	RefineModel(ctx context.Context, model *GenerationModel, feedback *ModelFeedback) (*GenerationModel, error)
	ValidateModelAccuracy(ctx context.Context, model *GenerationModel, testData *TestDataSet) (*ModelValidation, error)
}

// RelationshipPreserver maintains referential integrity during data generation
type RelationshipPreserver interface {
	AnalyzeRelationships(ctx context.Context, schema *DatabaseSchema) (*RelationshipGraph, error)
	PreserveConstraints(ctx context.Context, data *TestDataSet, relationships *RelationshipGraph) (*TestDataSet, error)
	ValidateIntegrity(ctx context.Context, data *TestDataSet, constraints *IntegrityConstraints) (*IntegrityValidation, error)
	RepairConstraintViolations(ctx context.Context, data *TestDataSet, violations *ConstraintViolations) (*TestDataSet, error)
}

// PrivacyEngine handles privacy-compliant data generation and anonymization
type PrivacyEngine interface {
	AnalyzePrivacyRequirements(ctx context.Context, schema *DatabaseSchema) (*PrivacyRequirements, error)
	GenerateCompliantData(ctx context.Context, requirements *PrivacyRequirements) (*PrivateDataSet, error)
	AnonymizeData(ctx context.Context, data *TestDataSet, techniques *AnonymizationTechniques) (*AnonymizedDataSet, error)
	ValidatePrivacyCompliance(ctx context.Context, data *TestDataSet, rules *PrivacyRules) (*PrivacyValidation, error)
}

// DataQualityValidator ensures generated data meets quality standards
type DataQualityValidator interface {
	ValidateDataQuality(ctx context.Context, data *TestDataSet, standards *QualityStandards) (*QualityValidation, error)
	AnalyzeDataDistribution(ctx context.Context, data *TestDataSet) (*DistributionAnalysis, error)
	DetectAnomalies(ctx context.Context, data *TestDataSet) (*AnomalyReport, error)
	ScoreDataRealism(ctx context.Context, data *TestDataSet, referenceData *ReferenceDataSet) (*RealismScore, error)
}

// Core Data Models

// DatabaseSchema represents the structure of a database
type DatabaseSchema struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Tables      []*Table               `json:"tables"`
	Views       []*View                `json:"views"`
	Indexes     []*Index               `json:"indexes"`
	Procedures  []*StoredProcedure     `json:"procedures"`
	Triggers    []*Trigger             `json:"triggers"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Table represents a database table
type Table struct {
	Name         string                 `json:"name"`
	Schema       string                 `json:"schema"`
	Columns      []*Column              `json:"columns"`
	PrimaryKeys  []string               `json:"primary_keys"`
	ForeignKeys  []*ForeignKey          `json:"foreign_keys"`
	Indexes      []*Index               `json:"indexes"`
	Constraints  []*Constraint          `json:"constraints"`
	RowCount     int64                  `json:"row_count"`
	DataSize     int64                  `json:"data_size"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Column represents a table column
type Column struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Length       int                    `json:"length"`
	Precision    int                    `json:"precision"`
	Scale        int                    `json:"scale"`
	IsNullable   bool                   `json:"is_nullable"`
	DefaultValue interface{}            `json:"default_value"`
	IsAutoIncrement bool                `json:"is_auto_increment"`
	IsUnique     bool                   `json:"is_unique"`
	Comment      string                 `json:"comment"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Name            string   `json:"name"`
	LocalColumns    []string `json:"local_columns"`
	ReferencedTable string   `json:"referenced_table"`
	ReferencedColumns []string `json:"referenced_columns"`
	OnUpdate        string   `json:"on_update"`
	OnDelete        string   `json:"on_delete"`
}

// Constraint represents a table constraint
type Constraint struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"` // CHECK, UNIQUE, NOT NULL, etc.
	Columns    []string    `json:"columns"`
	Expression string      `json:"expression"`
	IsEnabled  bool        `json:"is_enabled"`
}

// Index represents a database index
type Index struct {
	Name     string   `json:"name"`
	Table    string   `json:"table"`
	Columns  []string `json:"columns"`
	IsUnique bool     `json:"is_unique"`
	Type     string   `json:"type"`
}

// View represents a database view
type View struct {
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	Definition string `json:"definition"`
	Columns    []*Column `json:"columns"`
}

// StoredProcedure represents a stored procedure
type StoredProcedure struct {
	Name       string       `json:"name"`
	Schema     string       `json:"schema"`
	Parameters []*Parameter `json:"parameters"`
	ReturnType string       `json:"return_type"`
	Body       string       `json:"body"`
}

// Parameter represents a procedure parameter
type Parameter struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Direction string `json:"direction"` // IN, OUT, INOUT
}

// Trigger represents a database trigger
type Trigger struct {
	Name      string `json:"name"`
	Table     string `json:"table"`
	Event     string `json:"event"` // INSERT, UPDATE, DELETE
	Timing    string `json:"timing"` // BEFORE, AFTER
	Body      string `json:"body"`
	IsEnabled bool   `json:"is_enabled"`
}

// Schema Analysis Results

// SchemaProfile contains the analysis results of a database schema
type SchemaProfile struct {
	SchemaID        string                 `json:"schema_id"`
	Complexity      *ComplexityMetrics     `json:"complexity"`
	Relationships   *RelationshipMap       `json:"relationships"`
	Constraints     *ConstraintMap         `json:"constraints"`
	DataPatterns    *DataPatterns          `json:"data_patterns"`
	SizeEstimates   *SizeEstimates         `json:"size_estimates"`
	Recommendations *GenerationRecommendations `json:"recommendations"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
}

// ComplexityMetrics measures schema complexity
type ComplexityMetrics struct {
	TableCount         int     `json:"table_count"`
	ColumnCount        int     `json:"column_count"`
	RelationshipCount  int     `json:"relationship_count"`
	ConstraintCount    int     `json:"constraint_count"`
	IndexCount         int     `json:"index_count"`
	NormalizationLevel int     `json:"normalization_level"`
	ComplexityScore    float64 `json:"complexity_score"`
	EstimatedTime      time.Duration `json:"estimated_generation_time"`
}

// RelationshipMap represents relationships between tables
type RelationshipMap struct {
	Tables        map[string]*TableRelationships `json:"tables"`
	Dependencies  *DependencyGraph               `json:"dependencies"`
	Cycles        []*CyclicDependency            `json:"cycles"`
	Orphans       []string                       `json:"orphan_tables"`
}

// TableRelationships contains relationships for a specific table
type TableRelationships struct {
	TableName    string             `json:"table_name"`
	Parents      []*Relationship    `json:"parents"`
	Children     []*Relationship    `json:"children"`
	Siblings     []*Relationship    `json:"siblings"`
	SelfReferences []*Relationship  `json:"self_references"`
}

// Relationship represents a relationship between tables
type Relationship struct {
	Type            string  `json:"type"` // one-to-one, one-to-many, many-to-many
	FromTable       string  `json:"from_table"`
	ToTable         string  `json:"to_table"`
	FromColumns     []string `json:"from_columns"`
	ToColumns       []string `json:"to_columns"`
	Strength        float64 `json:"strength"` // 0.0-1.0
	IsOptional      bool    `json:"is_optional"`
	CascadeDelete   bool    `json:"cascade_delete"`
	CascadeUpdate   bool    `json:"cascade_update"`
}

// DependencyGraph represents table dependency order
type DependencyGraph struct {
	Nodes []*DependencyNode `json:"nodes"`
	Edges []*DependencyEdge `json:"edges"`
	Levels [][]string       `json:"levels"` // Generation order by level
}

// DependencyNode represents a table in the dependency graph
type DependencyNode struct {
	TableName    string `json:"table_name"`
	Level        int    `json:"level"`
	Dependencies int    `json:"dependency_count"`
	Dependents   int    `json:"dependent_count"`
}

// DependencyEdge represents a dependency between tables
type DependencyEdge struct {
	FromTable string  `json:"from_table"`
	ToTable   string  `json:"to_table"`
	Weight    float64 `json:"weight"`
	Type      string  `json:"type"`
}

// CyclicDependency represents a circular dependency
type CyclicDependency struct {
	Tables       []string `json:"tables"`
	BreakPoint   string   `json:"break_point"`
	Resolution   string   `json:"resolution"`
}

// Data Generation Models

// TestDataSet represents a generated set of test data
type TestDataSet struct {
	ID            string                    `json:"id"`
	SchemaID      string                    `json:"schema_id"`
	Volume        int                       `json:"volume"`
	Tables        map[string]*TableData     `json:"tables"`
	Relationships *PreservedRelationships   `json:"relationships"`
	Metadata      map[string]interface{}    `json:"metadata"`
	Quality       *DataQualityMetrics       `json:"quality"`
	GeneratedAt   time.Time                 `json:"generated_at"`
	GeneratedBy   string                    `json:"generated_by"`
}

// TableData represents generated data for a table
type TableData struct {
	TableName   string                   `json:"table_name"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"row_count"`
	Columns     []*ColumnData            `json:"columns"`
	Statistics  *TableStatistics         `json:"statistics"`
}

// ColumnData represents generated data for a column
type ColumnData struct {
	ColumnName   string                 `json:"column_name"`
	DataType     string                 `json:"data_type"`
	Values       []interface{}          `json:"values"`
	Distribution *ValueDistribution     `json:"distribution"`
	Constraints  *ColumnConstraints     `json:"constraints"`
}

// ValueDistribution represents the distribution of values in a column
type ValueDistribution struct {
	UniqueValues    int                    `json:"unique_values"`
	NullValues      int                    `json:"null_values"`
	MinValue        interface{}            `json:"min_value"`
	MaxValue        interface{}            `json:"max_value"`
	AvgValue        interface{}            `json:"avg_value"`
	MedianValue     interface{}            `json:"median_value"`
	ModeValue       interface{}            `json:"mode_value"`
	StandardDev     float64                `json:"standard_deviation"`
	Skewness        float64                `json:"skewness"`
	Kurtosis        float64                `json:"kurtosis"`
	Distribution    string                 `json:"distribution_type"`
	Histogram       map[string]int         `json:"histogram"`
}

// ColumnConstraints represents constraints for a column
type ColumnConstraints struct {
	MinLength    int         `json:"min_length"`
	MaxLength    int         `json:"max_length"`
	MinValue     interface{} `json:"min_value"`
	MaxValue     interface{} `json:"max_value"`
	Pattern      string      `json:"pattern"`
	AllowedValues []interface{} `json:"allowed_values"`
	IsRequired   bool        `json:"is_required"`
	IsUnique     bool        `json:"is_unique"`
}

// TableStatistics contains statistics for generated table data
type TableStatistics struct {
	RowCount         int                    `json:"row_count"`
	NullPercentage   float64                `json:"null_percentage"`
	DuplicateRows    int                    `json:"duplicate_rows"`
	DataSize         int64                  `json:"data_size_bytes"`
	GenerationTime   time.Duration          `json:"generation_time"`
	QualityScore     float64                `json:"quality_score"`
	Anomalies        []*DataAnomaly         `json:"anomalies"`
}

// DataAnomaly represents an anomaly in generated data
type DataAnomaly struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Column      string      `json:"column"`
	Value       interface{} `json:"value"`
	Expected    interface{} `json:"expected"`
}

// PreservedRelationships tracks maintained relationships in generated data
type PreservedRelationships struct {
	ForeignKeys      []*PreservedForeignKey `json:"foreign_keys"`
	Constraints      []*PreservedConstraint `json:"constraints"`
	IntegrityChecks  []*IntegrityCheck      `json:"integrity_checks"`
	ViolationCount   int                    `json:"violation_count"`
}

// PreservedForeignKey represents a maintained foreign key relationship
type PreservedForeignKey struct {
	FromTable     string `json:"from_table"`
	ToTable       string `json:"to_table"`
	LocalColumns  []string `json:"local_columns"`
	RemoteColumns []string `json:"remote_columns"`
	ValidReferences int    `json:"valid_references"`
	InvalidReferences int  `json:"invalid_references"`
}

// PreservedConstraint represents a maintained constraint
type PreservedConstraint struct {
	Type         string `json:"type"`
	Table        string `json:"table"`
	Columns      []string `json:"columns"`
	ValidRows    int    `json:"valid_rows"`
	InvalidRows  int    `json:"invalid_rows"`
}

// IntegrityCheck represents an integrity validation
type IntegrityCheck struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CheckedAt   time.Time `json:"checked_at"`
}

// Generation Configuration and Constraints

// DataConstraints defines constraints for data generation
type DataConstraints struct {
	TableConstraints   map[string]*TableConstraints `json:"table_constraints"`
	ColumnConstraints  map[string]*ColumnConstraints `json:"column_constraints"`
	VolumeConstraints  *VolumeConstraints            `json:"volume_constraints"`
	QualityConstraints *QualityConstraints           `json:"quality_constraints"`
	PrivacyConstraints *PrivacyConstraints           `json:"privacy_constraints"`
}

// TableConstraints defines constraints for a specific table
type TableConstraints struct {
	TableName         string                 `json:"table_name"`
	MinRows           int                    `json:"min_rows"`
	MaxRows           int                    `json:"max_rows"`
	RequiredColumns   []string               `json:"required_columns"`
	OptionalColumns   []string               `json:"optional_columns"`
	DataDistribution  *DistributionSpec      `json:"data_distribution"`
	CustomRules       []*CustomRule          `json:"custom_rules"`
}

// VolumeConstraints defines volume-related constraints
type VolumeConstraints struct {
	TotalRows         int     `json:"total_rows"`
	MaxTableSize      int     `json:"max_table_size"`
	MaxDataSize       int64   `json:"max_data_size_bytes"`
	DistributionRatio float64 `json:"distribution_ratio"`
	ScalingFactor     float64 `json:"scaling_factor"`
}

// QualityConstraints defines quality requirements
type QualityConstraints struct {
	MinQualityScore     float64 `json:"min_quality_score"`
	MaxNullPercentage   float64 `json:"max_null_percentage"`
	MaxDuplicatePercent float64 `json:"max_duplicate_percentage"`
	RequireRealism      bool    `json:"require_realism"`
	ValidateConstraints bool    `json:"validate_constraints"`
	EnforceRelationships bool   `json:"enforce_relationships"`
}

// PrivacyConstraints defines privacy requirements
type PrivacyConstraints struct {
	AnonymizePII       bool     `json:"anonymize_pii"`
	PseudonymizeFields []string `json:"pseudonymize_fields"`
	RedactFields       []string `json:"redact_fields"`
	EncryptFields      []string `json:"encrypt_fields"`
	ComplianceLevel    string   `json:"compliance_level"` // GDPR, HIPAA, CCPA
	RetentionPolicy    string   `json:"retention_policy"`
}

// DistributionSpec defines how data should be distributed
type DistributionSpec struct {
	Type       string                 `json:"type"` // uniform, normal, exponential, custom
	Parameters map[string]interface{} `json:"parameters"`
	Weights    map[string]float64     `json:"weights"`
}

// CustomRule defines custom generation rules
type CustomRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
}

// Statistics and Monitoring

// DataGenerationStats provides statistics about data generation
type DataGenerationStats struct {
	TotalDataSets       int64         `json:"total_datasets_generated"`
	TotalRows           int64         `json:"total_rows_generated"`
	TotalDataSize       int64         `json:"total_data_size_bytes"`
	AverageGenTime      time.Duration `json:"average_generation_time"`
	SuccessRate         float64       `json:"success_rate"`
	QualityScore        float64       `json:"average_quality_score"`
	PrivacyCompliance   float64       `json:"privacy_compliance_rate"`
	ConstraintViolations int64        `json:"constraint_violations"`
	LastGeneration      time.Time     `json:"last_generation"`
}

// DataQualityMetrics provides quality metrics for generated data
type DataQualityMetrics struct {
	OverallScore        float64                `json:"overall_score"`
	CompletenessScore   float64                `json:"completeness_score"`
	AccuracyScore       float64                `json:"accuracy_score"`
	ConsistencyScore    float64                `json:"consistency_score"`
	ValidityScore       float64                `json:"validity_score"`
	RealismScore        float64                `json:"realism_score"`
	TableScores         map[string]float64     `json:"table_scores"`
	ColumnScores        map[string]float64     `json:"column_scores"`
	Issues              []*QualityIssue        `json:"issues"`
	Recommendations     []*QualityRecommendation `json:"recommendations"`
}

// QualityIssue represents a data quality issue
type QualityIssue struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Table       string    `json:"table"`
	Column      string    `json:"column"`
	Count       int       `json:"count"`
	Examples    []string  `json:"examples"`
	DetectedAt  time.Time `json:"detected_at"`
}

// QualityRecommendation represents a recommendation for improving data quality
type QualityRecommendation struct {
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Impact      string                 `json:"expected_impact"`
}

// Advanced Features

// SyntheticDataSet represents AI-generated synthetic data
type SyntheticDataSet struct {
	ID              string                 `json:"id"`
	ModelID         string                 `json:"model_id"`
	GenerationMode  string                 `json:"generation_mode"`
	Data            *TestDataSet           `json:"data"`
	Authenticity    *AuthenticityMetrics   `json:"authenticity"`
	PrivacyLevel    string                 `json:"privacy_level"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AuthenticityMetrics measures how realistic synthetic data is
type AuthenticityMetrics struct {
	RealismScore      float64 `json:"realism_score"`
	DistributionMatch float64 `json:"distribution_match"`
	PatternMatch      float64 `json:"pattern_match"`
	CorrelationMatch  float64 `json:"correlation_match"`
	StatisticalMatch  float64 `json:"statistical_match"`
}

// RelationalDataSet represents data with complex relationships
type RelationalDataSet struct {
	ID              string                 `json:"id"`
	Tables          map[string]*TableData  `json:"tables"`
	Relationships   *RelationshipGraph     `json:"relationships"`
	IntegrityScore  float64                `json:"integrity_score"`
	Violations      []*IntegrityViolation  `json:"violations"`
}

// RelationshipGraph represents the relationship structure
type RelationshipGraph struct {
	Nodes []*TableNode      `json:"nodes"`
	Edges []*RelationshipEdge `json:"edges"`
	Paths []*RelationshipPath `json:"paths"`
}

// TableNode represents a table in the relationship graph
type TableNode struct {
	TableName   string  `json:"table_name"`
	RowCount    int     `json:"row_count"`
	Centrality  float64 `json:"centrality"`
	Importance  float64 `json:"importance"`
}

// RelationshipEdge represents a relationship between tables
type RelationshipEdge struct {
	FromTable    string  `json:"from_table"`
	ToTable      string  `json:"to_table"`
	Type         string  `json:"type"`
	Strength     float64 `json:"strength"`
	Cardinality  string  `json:"cardinality"`
}

// RelationshipPath represents a path through relationships
type RelationshipPath struct {
	Tables    []string `json:"tables"`
	Length    int      `json:"length"`
	Strength  float64  `json:"strength"`
	IsCyclic  bool     `json:"is_cyclic"`
}

// IntegrityViolation represents a relationship integrity violation
type IntegrityViolation struct {
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Table        string      `json:"table"`
	Column       string      `json:"column"`
	Value        interface{} `json:"value"`
	ExpectedValue interface{} `json:"expected_value"`
	Severity     string      `json:"severity"`
}

// Privacy and Anonymization

// PrivateDataSet represents privacy-compliant data
type PrivateDataSet struct {
	ID                string                 `json:"id"`
	Data              *TestDataSet           `json:"data"`
	PrivacyTechniques []*PrivacyTechnique    `json:"privacy_techniques"`
	ComplianceLevel   string                 `json:"compliance_level"`
	PrivacyScore      float64                `json:"privacy_score"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// PrivacyTechnique represents an applied privacy technique
type PrivacyTechnique struct {
	Type        string                 `json:"type"` // anonymization, pseudonymization, encryption
	Fields      []string               `json:"fields"`
	Parameters  map[string]interface{} `json:"parameters"`
	Strength    float64                `json:"strength"`
	Reversible  bool                   `json:"reversible"`
}

// AnonymizedDataSet represents anonymized data
type AnonymizedDataSet struct {
	ID              string                   `json:"id"`
	OriginalDataID  string                   `json:"original_data_id"`
	Data            *TestDataSet             `json:"data"`
	Techniques      []*AnonymizationTechnique `json:"techniques"`
	PrivacyLevel    string                   `json:"privacy_level"`
	UtilityScore    float64                  `json:"utility_score"`
	PrivacyScore    float64                  `json:"privacy_score"`
}

// AnonymizationTechnique represents an anonymization technique
type AnonymizationTechnique struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Fields      []string               `json:"fields"`
	Strength    float64                `json:"strength"`
	Information float64                `json:"information_loss"`
}

// Performance Testing Data

// PerformanceDataSet represents data for performance testing
type PerformanceDataSet struct {
	ID              string                 `json:"id"`
	ScenarioID      string                 `json:"scenario_id"`
	Data            *TestDataSet           `json:"data"`
	LoadProfile     *LoadProfile           `json:"load_profile"`
	Characteristics *PerformanceCharacteristics `json:"characteristics"`
}

// LoadProfile defines load testing parameters
type LoadProfile struct {
	ConcurrentUsers int           `json:"concurrent_users"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	TestDuration    time.Duration `json:"test_duration"`
	ThinkTime       time.Duration `json:"think_time"`
	DataVariation   float64       `json:"data_variation"`
	AccessPatterns  []*AccessPattern `json:"access_patterns"`
}

// AccessPattern defines how data is accessed during testing
type AccessPattern struct {
	Name        string  `json:"name"`
	Percentage  float64 `json:"percentage"`
	Operations  []string `json:"operations"`
	Tables      []string `json:"tables"`
	Complexity  string  `json:"complexity"`
}

// PerformanceCharacteristics describes performance aspects of the data
type PerformanceCharacteristics struct {
	DataSize        int64   `json:"data_size_bytes"`
	IndexUsage      float64 `json:"index_usage"`
	QueryComplexity float64 `json:"query_complexity"`
	JoinFactor      float64 `json:"join_factor"`
	ReadWriteRatio  float64 `json:"read_write_ratio"`
}

// LoadDataSet represents data specifically for load testing
type LoadDataSet struct {
	ID           string             `json:"id"`
	Users        []*VirtualUser     `json:"virtual_users"`
	Scenarios    []*LoadScenario    `json:"scenarios"`
	DataPools    map[string]*DataPool `json:"data_pools"`
	Distribution *LoadDistribution  `json:"distribution"`
}

// VirtualUser represents a virtual user in load testing
type VirtualUser struct {
	ID       string                 `json:"id"`
	Profile  *UserProfile           `json:"profile"`
	Data     map[string]interface{} `json:"data"`
	Behavior *UserBehavior          `json:"behavior"`
}

// UserProfile defines characteristics of a virtual user
type UserProfile struct {
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Experience   string    `json:"experience"`
	Preferences  map[string]interface{} `json:"preferences"`
	Constraints  map[string]interface{} `json:"constraints"`
}

// UserBehavior defines how a virtual user behaves
type UserBehavior struct {
	ThinkTime    time.Duration `json:"think_time"`
	ErrorRate    float64       `json:"error_rate"`
	Patterns     []string      `json:"patterns"`
	Preferences  []string      `json:"preferences"`
}

// LoadScenario represents a load testing scenario
type LoadScenario struct {
	Name        string        `json:"name"`
	Weight      float64       `json:"weight"`
	Steps       []*ScenarioStep `json:"steps"`
	Duration    time.Duration `json:"duration"`
	DataRequirements map[string]interface{} `json:"data_requirements"`
}

// ScenarioStep represents a step in a load scenario
type ScenarioStep struct {
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Data        map[string]interface{} `json:"data"`
	Validation  []string               `json:"validation"`
	ThinkTime   time.Duration          `json:"think_time"`
}

// DataPool represents a pool of test data
type DataPool struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Size     int                    `json:"size"`
	Data     []map[string]interface{} `json:"data"`
	Usage    string                 `json:"usage"` // shared, exclusive, cyclic
	Metadata map[string]interface{} `json:"metadata"`
}

// LoadDistribution defines how load is distributed
type LoadDistribution struct {
	Pattern     string    `json:"pattern"` // constant, ramp-up, spike, wave
	Parameters  map[string]interface{} `json:"parameters"`
	Timeline    []*LoadPoint `json:"timeline"`
}

// LoadPoint represents a point in load distribution
type LoadPoint struct {
	Time  time.Duration `json:"time"`
	Load  float64       `json:"load"`
	Users int           `json:"users"`
}

// Edge Case and Boundary Data

// BoundaryDataSet represents boundary value test data
type BoundaryDataSet struct {
	ID          string                    `json:"id"`
	Tables      map[string]*BoundaryTable `json:"tables"`
	TestCases   []*BoundaryTestCase       `json:"test_cases"`
	Coverage    *BoundaryCoverage         `json:"coverage"`
}

// BoundaryTable represents boundary data for a table
type BoundaryTable struct {
	TableName string                   `json:"table_name"`
	Columns   map[string]*BoundaryColumn `json:"columns"`
	Scenarios []*BoundaryScenario      `json:"scenarios"`
}

// BoundaryColumn represents boundary values for a column
type BoundaryColumn struct {
	ColumnName    string        `json:"column_name"`
	DataType      string        `json:"data_type"`
	MinValue      interface{}   `json:"min_value"`
	MaxValue      interface{}   `json:"max_value"`
	BoundaryValues []interface{} `json:"boundary_values"`
	EdgeCases     []interface{} `json:"edge_cases"`
}

// BoundaryScenario represents a boundary testing scenario
type BoundaryScenario struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // min, max, empty, overflow
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Expected    string                 `json:"expected_result"`
}

// BoundaryTestCase represents a boundary test case
type BoundaryTestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Input       map[string]interface{} `json:"input"`
	Expected    map[string]interface{} `json:"expected"`
	Description string                 `json:"description"`
}

// BoundaryCoverage tracks boundary test coverage
type BoundaryCoverage struct {
	TotalBoundaries  int     `json:"total_boundaries"`
	CoveredBoundaries int    `json:"covered_boundaries"`
	CoveragePercent  float64 `json:"coverage_percent"`
	MissingBoundaries []string `json:"missing_boundaries"`
}

// EdgeCaseDataSet represents edge case test data
type EdgeCaseDataSet struct {
	ID        string            `json:"id"`
	EdgeCases []*EdgeCase       `json:"edge_cases"`
	Coverage  *EdgeCaseCoverage `json:"coverage"`
	Scenarios []*EdgeScenario   `json:"scenarios"`
}

// EdgeCase represents an edge case
type EdgeCase struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Probability float64                `json:"probability"`
	Impact      string                 `json:"impact"`
}

// EdgeCaseCoverage tracks edge case coverage
type EdgeCaseCoverage struct {
	TotalEdgeCases    int     `json:"total_edge_cases"`
	CoveredEdgeCases  int     `json:"covered_edge_cases"`
	CoveragePercent   float64 `json:"coverage_percent"`
	HighImpactCases   int     `json:"high_impact_cases"`
	MediumImpactCases int     `json:"medium_impact_cases"`
	LowImpactCases    int     `json:"low_impact_cases"`
}

// EdgeScenario represents an edge case testing scenario
type EdgeScenario struct {
	Name        string     `json:"name"`
	EdgeCases   []*EdgeCase `json:"edge_cases"`
	Description string     `json:"description"`
	Setup       map[string]interface{} `json:"setup"`
	Teardown    map[string]interface{} `json:"teardown"`
}

// Learning and Feedback

// DataUsageMetrics tracks how generated data is used
type DataUsageMetrics struct {
	DataSetID       string                 `json:"dataset_id"`
	UsageFrequency  int                    `json:"usage_frequency"`
	TestSuccess     float64                `json:"test_success_rate"`
	Issues          []*UsageIssue          `json:"issues"`
	Performance     *UsagePerformance      `json:"performance"`
	UserFeedback    []*UserFeedback        `json:"user_feedback"`
	RecordedAt      time.Time              `json:"recorded_at"`
}

// UsageIssue represents an issue with generated data usage
type UsageIssue struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Frequency   int       `json:"frequency"`
	Severity    string    `json:"severity"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// UsagePerformance tracks performance of data usage
type UsagePerformance struct {
	LoadTime      time.Duration `json:"load_time"`
	ProcessTime   time.Duration `json:"process_time"`
	MemoryUsage   int64         `json:"memory_usage"`
	CPUUsage      float64       `json:"cpu_usage"`
	IOOperations  int64         `json:"io_operations"`
}

// UserFeedback represents feedback from data users
type UserFeedback struct {
	UserID      string                 `json:"user_id"`
	Rating      float64                `json:"rating"`
	Comments    string                 `json:"comments"`
	Issues      []string               `json:"issues"`
	Suggestions []string               `json:"suggestions"`
	Context     map[string]interface{} `json:"context"`
	SubmittedAt time.Time              `json:"submitted_at"`
}

// GenerationFeedback provides feedback for improving generation
type GenerationFeedback struct {
	DataSetID     string                 `json:"dataset_id"`
	QualityRating float64                `json:"quality_rating"`
	UsabilityRating float64              `json:"usability_rating"`
	Issues        []*FeedbackIssue       `json:"issues"`
	Improvements  []*FeedbackImprovement `json:"improvements"`
	Metrics       map[string]float64     `json:"metrics"`
	Timestamp     time.Time              `json:"timestamp"`
}

// FeedbackIssue represents an issue reported in feedback
type FeedbackIssue struct {
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Frequency   float64                `json:"frequency"`
	Context     map[string]interface{} `json:"context"`
}

// FeedbackImprovement represents a suggested improvement
type FeedbackImprovement struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Impact      string                 `json:"expected_impact"`
	Effort      string                 `json:"effort_required"`
	Details     map[string]interface{} `json:"details"`
}

// Data Types and Enums

// DataType represents supported data types
type DataType string

const (
	DataTypeString    DataType = "string"
	DataTypeInteger   DataType = "integer"
	DataTypeFloat     DataType = "float"
	DataTypeBoolean   DataType = "boolean"
	DataTypeDate      DataType = "date"
	DataTypeTime      DataType = "time"
	DataTypeDateTime  DataType = "datetime"
	DataTypeTimestamp DataType = "timestamp"
	DataTypeUUID      DataType = "uuid"
	DataTypeJSON      DataType = "json"
	DataTypeBinary    DataType = "binary"
	DataTypeText      DataType = "text"
	DataTypeEmail     DataType = "email"
	DataTypeURL       DataType = "url"
	DataTypePhone     DataType = "phone"
	DataTypeIPAddress DataType = "ip_address"
	DataTypeCurrency  DataType = "currency"
)

// PrivacyLevel represents privacy protection levels
type PrivacyLevel string

const (
	PrivacyLevelNone   PrivacyLevel = "none"
	PrivacyLevelLow    PrivacyLevel = "low"
	PrivacyLevelMedium PrivacyLevel = "medium"
	PrivacyLevelHigh   PrivacyLevel = "high"
	PrivacyLevelMaximum PrivacyLevel = "maximum"
)

// GenerationStrategy represents data generation strategies
type GenerationStrategy string

const (
	StrategyRandom     GenerationStrategy = "random"
	StrategyRealistic  GenerationStrategy = "realistic"
	StrategyPattern    GenerationStrategy = "pattern_based"
	StrategyML         GenerationStrategy = "ml_generated"
	StrategyTemplate   GenerationStrategy = "template_based"
	StrategyRule       GenerationStrategy = "rule_based"
	StrategySynthetic  GenerationStrategy = "synthetic"
	StrategyAnonymized GenerationStrategy = "anonymized"
)

// QualityLevel represents data quality levels
type QualityLevel string

const (
	QualityLevelBasic      QualityLevel = "basic"
	QualityLevelStandard   QualityLevel = "standard"
	QualityLevelHigh       QualityLevel = "high"
	QualityLevelPremium    QualityLevel = "premium"
	QualityLevelEnterprise QualityLevel = "enterprise"
)