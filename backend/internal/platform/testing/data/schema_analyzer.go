// Package data implements the database schema analyzer for intelligent test data generation
package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"openpenpal-backend/internal/platform/testing/core"
)

// PostgreSQLSchemaAnalyzer analyzes PostgreSQL database schemas
type PostgreSQLSchemaAnalyzer struct {
	db     *sql.DB
	config *AnalyzerConfig
	cache  *AnalysisCache
	stats  *AnalyzerStats
}

// AnalyzerConfig configures schema analysis behavior
type AnalyzerConfig struct {
	MaxSampleRows        int           `json:"max_sample_rows"`
	AnalysisTimeout      time.Duration `json:"analysis_timeout"`
	EnableDeepAnalysis   bool          `json:"enable_deep_analysis"`
	CacheResults         bool          `json:"cache_results"`
	DetectPatterns       bool          `json:"detect_patterns"`
	AnalyzeDistribution  bool          `json:"analyze_distribution"`
	EstimateVolumes      bool          `json:"estimate_volumes"`
	ValidateConstraints  bool          `json:"validate_constraints"`
	SamplePercentage     float64       `json:"sample_percentage"`
}

// AnalysisCache caches analysis results for performance
type AnalysisCache struct {
	schemas      map[string]*SchemaProfile
	tables       map[string]*TableProfile
	lastUpdated  map[string]time.Time
	ttl          time.Duration
	hits         int64
	misses       int64
}

// AnalyzerStats tracks analyzer performance
type AnalyzerStats struct {
	TotalAnalyses       int64         `json:"total_analyses"`
	TotalTables         int64         `json:"total_tables"`
	TotalColumns        int64         `json:"total_columns"`
	AverageAnalysisTime time.Duration `json:"average_analysis_time"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
	ErrorCount          int64         `json:"error_count"`
	LastAnalysis        time.Time     `json:"last_analysis"`
}

// TableProfile contains detailed analysis of a table
type TableProfile struct {
	TableName        string                    `json:"table_name"`
	Schema           string                    `json:"schema"`
	RowCount         int64                     `json:"row_count"`
	DataSize         int64                     `json:"data_size_bytes"`
	Columns          []*ColumnProfile          `json:"columns"`
	Indexes          []*IndexProfile           `json:"indexes"`
	Constraints      []*ConstraintProfile      `json:"constraints"`
	Relationships    []*RelationshipProfile    `json:"relationships"`
	DataDistribution *TableDataDistribution    `json:"data_distribution"`
	Patterns         []*TablePattern           `json:"patterns"`
	ComplexityScore  float64                   `json:"complexity_score"`
	AnalyzedAt       time.Time                 `json:"analyzed_at"`
}

// ColumnProfile contains detailed analysis of a column
type ColumnProfile struct {
	ColumnName       string                 `json:"column_name"`
	DataType         string                 `json:"data_type"`
	Length           int                    `json:"length"`
	Precision        int                    `json:"precision"`
	Scale            int                    `json:"scale"`
	IsNullable       bool                   `json:"is_nullable"`
	DefaultValue     interface{}            `json:"default_value"`
	IsAutoIncrement  bool                   `json:"is_auto_increment"`
	IsUnique         bool                   `json:"is_unique"`
	IsPrimaryKey     bool                   `json:"is_primary_key"`
	IsForeignKey     bool                   `json:"is_foreign_key"`
	Distribution     *ColumnDistribution    `json:"distribution"`
	Patterns         []*ColumnPattern       `json:"patterns"`
	Constraints      []*ColumnConstraint    `json:"constraints"`
	SampleValues     []interface{}          `json:"sample_values"`
	Statistics       *ColumnStatistics      `json:"statistics"`
	QualityMetrics   *ColumnQualityMetrics  `json:"quality_metrics"`
	GenerationHints  *GenerationHints       `json:"generation_hints"`
}

// ColumnDistribution analyzes value distribution in a column
type ColumnDistribution struct {
	TotalValues     int64                  `json:"total_values"`
	UniqueValues    int64                  `json:"unique_values"`
	NullValues      int64                  `json:"null_values"`
	MinValue        interface{}            `json:"min_value"`
	MaxValue        interface{}            `json:"max_value"`
	AvgValue        interface{}            `json:"avg_value"`
	MedianValue     interface{}            `json:"median_value"`
	ModeValue       interface{}            `json:"mode_value"`
	StandardDev     float64                `json:"standard_deviation"`
	Variance        float64                `json:"variance"`
	Skewness        float64                `json:"skewness"`
	Kurtosis        float64                `json:"kurtosis"`
	DistributionType string                `json:"distribution_type"`
	ValueFrequency  map[string]int64       `json:"value_frequency"`
	Percentiles     map[string]interface{} `json:"percentiles"`
	Histogram       *HistogramData         `json:"histogram"`
}

// HistogramData represents histogram information
type HistogramData struct {
	Buckets    []string `json:"buckets"`
	Counts     []int64  `json:"counts"`
	BucketSize float64  `json:"bucket_size"`
	Range      float64  `json:"range"`
}

// ColumnStatistics contains statistical information about a column
type ColumnStatistics struct {
	Cardinality       float64 `json:"cardinality"`
	Entropy           float64 `json:"entropy"`
	NullPercentage    float64 `json:"null_percentage"`
	UniquePercentage  float64 `json:"unique_percentage"`
	DuplicatePercent  float64 `json:"duplicate_percentage"`
	DataTypeConsistency float64 `json:"data_type_consistency"`
	PatternConsistency  float64 `json:"pattern_consistency"`
	CompletenessScore   float64 `json:"completeness_score"`
	ValidityScore       float64 `json:"validity_score"`
}

// ColumnQualityMetrics measures column data quality
type ColumnQualityMetrics struct {
	OverallScore      float64 `json:"overall_score"`
	AccuracyScore     float64 `json:"accuracy_score"`
	CompletenessScore float64 `json:"completeness_score"`
	ConsistencyScore  float64 `json:"consistency_score"`
	ValidityScore     float64 `json:"validity_score"`
	FreshnessScore    float64 `json:"freshness_score"`
	IssueCount        int     `json:"issue_count"`
	Anomalies         []*ColumnAnomaly `json:"anomalies"`
}

// ColumnAnomaly represents an anomaly detected in a column
type ColumnAnomaly struct {
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Count       int64       `json:"count"`
	Percentage  float64     `json:"percentage"`
}

// GenerationHints provides hints for test data generation
type GenerationHints struct {
	Strategy         GenerationStrategy     `json:"strategy"`
	Patterns         []string               `json:"patterns"`
	ValueRanges      map[string]interface{} `json:"value_ranges"`
	Distributions    []string               `json:"distributions"`
	Templates        []string               `json:"templates"`
	Dependencies     []string               `json:"dependencies"`
	CustomRules      []*CustomRule          `json:"custom_rules"`
	QualityTargets   map[string]float64     `json:"quality_targets"`
}

// ColumnPattern represents a detected pattern in column data
type ColumnPattern struct {
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Confidence  float64 `json:"confidence"`
	Frequency   int64   `json:"frequency"`
	Examples    []string `json:"examples"`
	Description string  `json:"description"`
}

// ColumnConstraint represents a constraint on a column
type ColumnConstraint struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Definition  string                 `json:"definition"`
	Parameters  map[string]interface{} `json:"parameters"`
	IsEnforced  bool                   `json:"is_enforced"`
	Violations  int64                  `json:"violations"`
}

// IndexProfile contains analysis of database indexes
type IndexProfile struct {
	IndexName   string   `json:"index_name"`
	TableName   string   `json:"table_name"`
	Columns     []string `json:"columns"`
	IsUnique    bool     `json:"is_unique"`
	IsPrimary   bool     `json:"is_primary"`
	Type        string   `json:"type"`
	Size        int64    `json:"size_bytes"`
	Usage       *IndexUsage `json:"usage"`
	Selectivity float64  `json:"selectivity"`
}

// IndexUsage tracks index usage statistics
type IndexUsage struct {
	ScanCount     int64 `json:"scan_count"`
	TupleReads    int64 `json:"tuple_reads"`
	TupleFetches  int64 `json:"tuple_fetches"`
	LastUsed      time.Time `json:"last_used"`
	Efficiency    float64 `json:"efficiency"`
}

// ConstraintProfile contains analysis of database constraints
type ConstraintProfile struct {
	ConstraintName string   `json:"constraint_name"`
	Type          string   `json:"type"`
	TableName     string   `json:"table_name"`
	Columns       []string `json:"columns"`
	Definition    string   `json:"definition"`
	IsEnforced    bool     `json:"is_enforced"`
	Violations    int64    `json:"violations"`
	Impact        string   `json:"impact"`
}

// RelationshipProfile contains analysis of table relationships
type RelationshipProfile struct {
	Type              string  `json:"type"`
	FromTable         string  `json:"from_table"`
	ToTable           string  `json:"to_table"`
	FromColumns       []string `json:"from_columns"`
	ToColumns         []string `json:"to_columns"`
	Cardinality       string  `json:"cardinality"`
	Strength          float64 `json:"strength"`
	IsOptional        bool    `json:"is_optional"`
	CascadeDelete     bool    `json:"cascade_delete"`
	CascadeUpdate     bool    `json:"cascade_update"`
	ViolationCount    int64   `json:"violation_count"`
	DataConsistency   float64 `json:"data_consistency"`
}

// TableDataDistribution analyzes data distribution patterns
type TableDataDistribution struct {
	RowDistribution    *RowDistribution    `json:"row_distribution"`
	ColumnCorrelations []*ColumnCorrelation `json:"column_correlations"`
	DataSkew          *DataSkew           `json:"data_skew"`
	Hotspots          []*DataHotspot      `json:"hotspots"`
	Seasonality       *SeasonalityPattern `json:"seasonality"`
}

// RowDistribution analyzes how data is distributed across rows
type RowDistribution struct {
	TotalRows        int64   `json:"total_rows"`
	ActiveRows       int64   `json:"active_rows"`
	DeletedRows      int64   `json:"deleted_rows"`
	EmptyRows        int64   `json:"empty_rows"`
	DensityScore     float64 `json:"density_score"`
	FragmentationLevel float64 `json:"fragmentation_level"`
}

// ColumnCorrelation analyzes correlation between columns
type ColumnCorrelation struct {
	Column1     string  `json:"column1"`
	Column2     string  `json:"column2"`
	Correlation float64 `json:"correlation"`
	Type        string  `json:"type"`
	Strength    string  `json:"strength"`
	P_Value     float64 `json:"p_value"`
	Significant bool    `json:"significant"`
}

// DataSkew analyzes data skewness patterns
type DataSkew struct {
	OverallSkew     float64            `json:"overall_skew"`
	ColumnSkew      map[string]float64 `json:"column_skew"`
	HotValues       []*HotValue        `json:"hot_values"`
	ColdValues      []*ColdValue       `json:"cold_values"`
	SkewImpact      string             `json:"skew_impact"`
}

// HotValue represents frequently occurring values
type HotValue struct {
	Column     string      `json:"column"`
	Value      interface{} `json:"value"`
	Count      int64       `json:"count"`
	Percentage float64     `json:"percentage"`
	Impact     string      `json:"impact"`
}

// ColdValue represents rarely occurring values
type ColdValue struct {
	Column     string      `json:"column"`
	Value      interface{} `json:"value"`
	Count      int64       `json:"count"`
	Percentage float64     `json:"percentage"`
	Rarity     string      `json:"rarity"`
}

// DataHotspot represents areas of high data activity
type DataHotspot struct {
	Location    string  `json:"location"`
	Type        string  `json:"type"`
	Intensity   float64 `json:"intensity"`
	Size        int64   `json:"size"`
	Frequency   float64 `json:"frequency"`
	Description string  `json:"description"`
}

// SeasonalityPattern analyzes temporal patterns in data
type SeasonalityPattern struct {
	HasSeasonality bool                     `json:"has_seasonality"`
	Patterns       []*TemporalPattern       `json:"patterns"`
	Trends         []*TrendAnalysis         `json:"trends"`
	Cycles         []*CyclicPattern         `json:"cycles"`
	Anomalies      []*TemporalAnomaly       `json:"anomalies"`
}

// TemporalPattern represents a temporal pattern in data
type TemporalPattern struct {
	Type        string        `json:"type"`
	Period      time.Duration `json:"period"`
	Strength    float64       `json:"strength"`
	Confidence  float64       `json:"confidence"`
	Description string        `json:"description"`
}

// TrendAnalysis analyzes data trends over time
type TrendAnalysis struct {
	Direction   string  `json:"direction"`
	Slope       float64 `json:"slope"`
	Strength    float64 `json:"strength"`
	Confidence  float64 `json:"confidence"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// CyclicPattern represents cyclical patterns in data
type CyclicPattern struct {
	Type       string        `json:"type"`
	Frequency  float64       `json:"frequency"`
	Amplitude  float64       `json:"amplitude"`
	Phase      float64       `json:"phase"`
	Period     time.Duration `json:"period"`
	Confidence float64       `json:"confidence"`
}

// TemporalAnomaly represents temporal anomalies in data
type TemporalAnomaly struct {
	Time        time.Time `json:"time"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Value       interface{} `json:"value"`
	Expected    interface{} `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Description string    `json:"description"`
}

// TablePattern represents patterns detected in table structure/data
type TablePattern struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Frequency   int64                  `json:"frequency"`
	Parameters  map[string]interface{} `json:"parameters"`
	Examples    []string               `json:"examples"`
	Impact      string                 `json:"impact"`
}

// NewPostgreSQLSchemaAnalyzer creates a new PostgreSQL schema analyzer
func NewPostgreSQLSchemaAnalyzer(db *sql.DB, config *AnalyzerConfig) *PostgreSQLSchemaAnalyzer {
	if config == nil {
		config = &AnalyzerConfig{
			MaxSampleRows:       1000,
			AnalysisTimeout:     30 * time.Second,
			EnableDeepAnalysis:  true,
			CacheResults:        true,
			DetectPatterns:      true,
			AnalyzeDistribution: true,
			EstimateVolumes:     true,
			ValidateConstraints: true,
			SamplePercentage:    5.0,
		}
	}

	cache := &AnalysisCache{
		schemas:     make(map[string]*SchemaProfile),
		tables:      make(map[string]*TableProfile),
		lastUpdated: make(map[string]time.Time),
		ttl:         1 * time.Hour,
	}

	return &PostgreSQLSchemaAnalyzer{
		db:     db,
		config: config,
		cache:  cache,
		stats:  &AnalyzerStats{},
	}
}

// AnalyzeSchema analyzes a database schema comprehensively
func (psa *PostgreSQLSchemaAnalyzer) AnalyzeSchema(ctx context.Context, schema *DatabaseSchema) (*SchemaProfile, error) {
	startTime := time.Now()
	log.Printf("🔍 Starting comprehensive schema analysis for: %s", schema.Name)

	// Check cache first
	if psa.config.CacheResults {
		if cached := psa.getCachedSchema(schema.ID); cached != nil {
			log.Printf("📋 Using cached schema analysis for: %s", schema.Name)
			psa.cache.hits++
			return cached, nil
		}
		psa.cache.misses++
	}

	profile := &SchemaProfile{
		SchemaID:   schema.ID,
		AnalyzedAt: time.Now(),
	}

	// Analyze schema complexity
	complexity, err := psa.analyzeSchemaComplexity(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze schema complexity: %w", err)
	}
	profile.Complexity = complexity

	// Analyze relationships
	relationships, err := psa.analyzeSchemaRelationships(ctx, schema)
	if err != nil {
		log.Printf("⚠️  Failed to analyze relationships: %v", err)
		relationships = &RelationshipMap{
			Tables:       make(map[string]*TableRelationships),
			Dependencies: &DependencyGraph{},
			Cycles:       []*CyclicDependency{},
			Orphans:      []string{},
		}
	}
	profile.Relationships = relationships

	// Analyze constraints
	constraints, err := psa.analyzeSchemaConstraints(ctx, schema)
	if err != nil {
		log.Printf("⚠️  Failed to analyze constraints: %v", err)
		constraints = &ConstraintMap{}
	}
	profile.Constraints = constraints

	// Detect data patterns
	if psa.config.DetectPatterns {
		patterns, err := psa.detectDataPatterns(ctx, schema)
		if err != nil {
			log.Printf("⚠️  Failed to detect patterns: %v", err)
			patterns = &DataPatterns{}
		}
		profile.DataPatterns = patterns
	}

	// Estimate data sizes
	if psa.config.EstimateVolumes {
		estimates, err := psa.estimateDataSizes(ctx, schema)
		if err != nil {
			log.Printf("⚠️  Failed to estimate sizes: %v", err)
			estimates = &SizeEstimates{}
		}
		profile.SizeEstimates = estimates
	}

	// Generate recommendations
	recommendations := psa.generateRecommendations(profile)
	profile.Recommendations = recommendations

	// Cache result
	if psa.config.CacheResults {
		psa.cacheSchema(schema.ID, profile)
	}

	// Update statistics
	analysisTime := time.Since(startTime)
	psa.updateStats(analysisTime, len(schema.Tables))

	log.Printf("✅ Schema analysis completed in %v - %d tables analyzed", 
		analysisTime, len(schema.Tables))

	return profile, nil
}

// DetectRelationships detects relationships between tables
func (psa *PostgreSQLSchemaAnalyzer) DetectRelationships(ctx context.Context, tables []*Table) (*RelationshipMap, error) {
	log.Printf("🔗 Detecting relationships between %d tables", len(tables))

	relationshipMap := &RelationshipMap{
		Tables:       make(map[string]*TableRelationships),
		Dependencies: &DependencyGraph{},
		Cycles:       []*CyclicDependency{},
		Orphans:      []string{},
	}

	// Analyze each table for relationships
	for _, table := range tables {
		tableRels := &TableRelationships{
			TableName:      table.Name,
			Parents:        []*Relationship{},
			Children:       []*Relationship{},
			Siblings:       []*Relationship{},
			SelfReferences: []*Relationship{},
		}

		// Analyze foreign key relationships
		for _, fk := range table.ForeignKeys {
			relationship := &Relationship{
				Type:          "one-to-many",
				FromTable:     table.Name,
				ToTable:       fk.ReferencedTable,
				FromColumns:   fk.LocalColumns,
				ToColumns:     fk.ReferencedColumns,
				Strength:      0.8, // Default strength
				IsOptional:    psa.isForeignKeyOptional(table, fk),
				CascadeDelete: fk.OnDelete == "CASCADE",
				CascadeUpdate: fk.OnUpdate == "CASCADE",
			}

			// Analyze relationship strength
			strength, err := psa.analyzeRelationshipStrength(ctx, relationship)
			if err == nil {
				relationship.Strength = strength
			}

			// Check if it's a self-reference
			if fk.ReferencedTable == table.Name {
				tableRels.SelfReferences = append(tableRels.SelfReferences, relationship)
			} else {
				tableRels.Parents = append(tableRels.Parents, relationship)
			}
		}

		relationshipMap.Tables[table.Name] = tableRels
	}

	// Build dependency graph
	dependencyGraph, err := psa.buildDependencyGraph(tables)
	if err != nil {
		log.Printf("⚠️  Failed to build dependency graph: %v", err)
	} else {
		relationshipMap.Dependencies = dependencyGraph
	}

	// Detect circular dependencies
	cycles := psa.detectCircularDependencies(relationshipMap)
	relationshipMap.Cycles = cycles

	// Identify orphan tables
	orphans := psa.identifyOrphanTables(relationshipMap)
	relationshipMap.Orphans = orphans

	log.Printf("🔗 Relationship detection completed: %d relationships found", 
		psa.countTotalRelationships(relationshipMap))

	return relationshipMap, nil
}

// IdentifyConstraints identifies and analyzes table constraints
func (psa *PostgreSQLSchemaAnalyzer) IdentifyConstraints(ctx context.Context, table *Table) (*TableConstraints, error) {
	log.Printf("📏 Identifying constraints for table: %s", table.Name)

	constraints := &TableConstraints{
		TableName:         table.Name,
		MinRows:           0,
		MaxRows:           math.MaxInt32,
		RequiredColumns:   []string{},
		OptionalColumns:   []string{},
		DataDistribution:  &DistributionSpec{},
		CustomRules:       []*CustomRule{},
	}

	// Analyze primary key constraints
	if len(table.PrimaryKeys) > 0 {
		for _, pk := range table.PrimaryKeys {
			constraints.RequiredColumns = append(constraints.RequiredColumns, pk)
		}
	}

	// Analyze foreign key constraints
	for _, fk := range table.ForeignKeys {
		for _, col := range fk.LocalColumns {
			if !contains(constraints.RequiredColumns, col) {
				constraints.RequiredColumns = append(constraints.RequiredColumns, col)
			}
		}
	}

	// Analyze column constraints
	for _, column := range table.Columns {
		if !column.IsNullable {
			if !contains(constraints.RequiredColumns, column.Name) {
				constraints.RequiredColumns = append(constraints.RequiredColumns, column.Name)
			}
		} else {
			constraints.OptionalColumns = append(constraints.OptionalColumns, column.Name)
		}
	}

	// Analyze check constraints
	for _, constraint := range table.Constraints {
		if constraint.Type == "CHECK" {
			rule := psa.parseCheckConstraint(constraint)
			if rule != nil {
				constraints.CustomRules = append(constraints.CustomRules, rule)
			}
		}
	}

	// Estimate row constraints based on data analysis
	if psa.config.EstimateVolumes {
		minRows, maxRows, err := psa.estimateRowConstraints(ctx, table)
		if err == nil {
			constraints.MinRows = minRows
			constraints.MaxRows = maxRows
		}
	}

	log.Printf("📏 Identified %d required columns, %d optional columns, %d custom rules for table: %s",
		len(constraints.RequiredColumns), len(constraints.OptionalColumns), 
		len(constraints.CustomRules), table.Name)

	return constraints, nil
}

// AnalyzeDataDistribution analyzes data distribution patterns
func (psa *PostgreSQLSchemaAnalyzer) AnalyzeDataDistribution(ctx context.Context, sampleData *SampleData) (*DataDistribution, error) {
	log.Printf("📊 Analyzing data distribution for sample data")

	distribution := &DataDistribution{
		OverallPattern:    "unknown",
		ColumnDistributions: make(map[string]*ColumnDistribution),
		Correlations:      []*ColumnCorrelation{},
		Outliers:          []*Outlier{},
		Trends:            []*TrendAnalysis{},
	}

	// Analyze each column's distribution
	for columnName, values := range sampleData.Columns {
		colDist, err := psa.analyzeColumnDistribution(columnName, values)
		if err != nil {
			log.Printf("⚠️  Failed to analyze distribution for column %s: %v", columnName, err)
			continue
		}
		distribution.ColumnDistributions[columnName] = colDist
	}

	// Analyze correlations between numeric columns
	correlations := psa.analyzeColumnCorrelations(sampleData)
	distribution.Correlations = correlations

	// Detect outliers
	outliers := psa.detectOutliers(sampleData)
	distribution.Outliers = outliers

	// Analyze temporal trends if timestamp columns exist
	if psa.hasTemporalColumns(sampleData) {
		trends := psa.analyzeTemporalTrends(sampleData)
		distribution.Trends = trends
	}

	// Determine overall pattern
	overallPattern := psa.determineOverallPattern(distribution)
	distribution.OverallPattern = overallPattern

	log.Printf("📊 Data distribution analysis completed - pattern: %s", overallPattern)

	return distribution, nil
}

// GetSchemaComplexity calculates schema complexity metrics
func (psa *PostgreSQLSchemaAnalyzer) GetSchemaComplexity(schema *DatabaseSchema) *ComplexityMetrics {
	log.Printf("📐 Calculating schema complexity for: %s", schema.Name)

	metrics := &ComplexityMetrics{
		TableCount:        len(schema.Tables),
		ColumnCount:       0,
		RelationshipCount: 0,
		ConstraintCount:   0,
		IndexCount:        len(schema.Indexes),
	}

	// Count columns, relationships, and constraints
	for _, table := range schema.Tables {
		metrics.ColumnCount += len(table.Columns)
		metrics.RelationshipCount += len(table.ForeignKeys)
		metrics.ConstraintCount += len(table.Constraints)
		metrics.IndexCount += len(table.Indexes)
	}

	// Calculate normalization level (simplified)
	metrics.NormalizationLevel = psa.estimateNormalizationLevel(schema)

	// Calculate complexity score
	metrics.ComplexityScore = psa.calculateComplexityScore(metrics)

	// Estimate generation time
	metrics.EstimatedTime = psa.estimateGenerationTime(metrics)

	log.Printf("📐 Schema complexity: score=%.2f, tables=%d, columns=%d, relationships=%d",
		metrics.ComplexityScore, metrics.TableCount, metrics.ColumnCount, metrics.RelationshipCount)

	return metrics
}

// Private helper methods

func (psa *PostgreSQLSchemaAnalyzer) getCachedSchema(schemaID string) *SchemaProfile {
	if profile, exists := psa.cache.schemas[schemaID]; exists {
		if lastUpdated, ok := psa.cache.lastUpdated[schemaID]; ok {
			if time.Since(lastUpdated) < psa.cache.ttl {
				return profile
			}
		}
	}
	return nil
}

func (psa *PostgreSQLSchemaAnalyzer) cacheSchema(schemaID string, profile *SchemaProfile) {
	psa.cache.schemas[schemaID] = profile
	psa.cache.lastUpdated[schemaID] = time.Now()
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeSchemaComplexity(ctx context.Context, schema *DatabaseSchema) (*ComplexityMetrics, error) {
	return psa.GetSchemaComplexity(schema), nil
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeSchemaRelationships(ctx context.Context, schema *DatabaseSchema) (*RelationshipMap, error) {
	return psa.DetectRelationships(ctx, schema.Tables)
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeSchemaConstraints(ctx context.Context, schema *DatabaseSchema) (*ConstraintMap, error) {
	constraintMap := &ConstraintMap{
		Tables:          make(map[string]*TableConstraints),
		GlobalRules:     []*GlobalConstraintRule{},
		IntegrityRules:  []*IntegrityRule{},
		BusinessRules:   []*BusinessRule{},
	}

	for _, table := range schema.Tables {
		tableConstraints, err := psa.IdentifyConstraints(ctx, table)
		if err != nil {
			log.Printf("⚠️  Failed to analyze constraints for table %s: %v", table.Name, err)
			continue
		}
		constraintMap.Tables[table.Name] = tableConstraints
	}

	return constraintMap, nil
}

func (psa *PostgreSQLSchemaAnalyzer) detectDataPatterns(ctx context.Context, schema *DatabaseSchema) (*DataPatterns, error) {
	patterns := &DataPatterns{
		TablePatterns:  []*TablePattern{},
		ColumnPatterns: []*ColumnPattern{},
		DataPatterns:   []*DetectedPattern{},
		NamingPatterns: []*NamingPattern{},
	}

	// Detect table patterns
	for _, table := range schema.Tables {
		tablePatterns := psa.detectTablePatterns(table)
		patterns.TablePatterns = append(patterns.TablePatterns, tablePatterns...)

		// Detect column patterns
		for _, column := range table.Columns {
			columnPatterns := psa.detectColumnPatterns(column)
			patterns.ColumnPatterns = append(patterns.ColumnPatterns, columnPatterns...)
		}
	}

	// Detect naming patterns
	namingPatterns := psa.detectNamingPatterns(schema)
	patterns.NamingPatterns = namingPatterns

	return patterns, nil
}

func (psa *PostgreSQLSchemaAnalyzer) estimateDataSizes(ctx context.Context, schema *DatabaseSchema) (*SizeEstimates, error) {
	estimates := &SizeEstimates{
		TableSizes:     make(map[string]*TableSizeEstimate),
		TotalRows:      0,
		TotalDataSize:  0,
		GrowthRate:     0.1, // 10% default growth rate
		ProjectedSizes: make(map[string]*ProjectedSize),
	}

	for _, table := range schema.Tables {
		estimate := psa.estimateTableSize(table)
		estimates.TableSizes[table.Name] = estimate
		estimates.TotalRows += estimate.EstimatedRows
		estimates.TotalDataSize += estimate.EstimatedSize
	}

	return estimates, nil
}

func (psa *PostgreSQLSchemaAnalyzer) generateRecommendations(profile *SchemaProfile) *GenerationRecommendations {
	recommendations := &GenerationRecommendations{
		Strategy:           StrategyRealistic,
		QualityLevel:       QualityLevelStandard,
		PrivacyLevel:       PrivacyLevelMedium,
		VolumeRecommendations: []*VolumeRecommendation{},
		QualityRecommendations: []*QualityRecommendation{},
		PerformanceRecommendations: []*PerformanceRecommendation{},
		SecurityRecommendations: []*SecurityRecommendation{},
	}

	// Generate strategy recommendation based on complexity
	if profile.Complexity.ComplexityScore > 0.8 {
		recommendations.Strategy = StrategyML
		recommendations.QualityLevel = QualityLevelHigh
	} else if profile.Complexity.ComplexityScore > 0.6 {
		recommendations.Strategy = StrategyPattern
		recommendations.QualityLevel = QualityLevelStandard
	}

	// Generate volume recommendations
	for tableName, sizeEstimate := range profile.SizeEstimates.TableSizes {
		volRec := &VolumeRecommendation{
			TableName:        tableName,
			RecommendedRows:  int(sizeEstimate.EstimatedRows * 0.1), // 10% of estimated
			MinRows:          100,
			MaxRows:          int(sizeEstimate.EstimatedRows),
			ScalingFactor:    1.0,
			DistributionType: "proportional",
		}
		recommendations.VolumeRecommendations = append(recommendations.VolumeRecommendations, volRec)
	}

	return recommendations
}

func (psa *PostgreSQLSchemaAnalyzer) isForeignKeyOptional(table *Table, fk *ForeignKey) bool {
	for _, column := range table.Columns {
		for _, fkCol := range fk.LocalColumns {
			if column.Name == fkCol {
				return column.IsNullable
			}
		}
	}
	return false
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeRelationshipStrength(ctx context.Context, rel *Relationship) (float64, error) {
	// Simple strength calculation based on cardinality and constraints
	strength := 0.5 // Base strength

	if !rel.IsOptional {
		strength += 0.2
	}
	if rel.CascadeDelete {
		strength += 0.2
	}
	if rel.CascadeUpdate {
		strength += 0.1
	}

	return math.Min(strength, 1.0), nil
}

func (psa *PostgreSQLSchemaAnalyzer) buildDependencyGraph(tables []*Table) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes:  []*DependencyNode{},
		Edges:  []*DependencyEdge{},
		Levels: [][]string{},
	}

	// Create nodes for each table
	nodeMap := make(map[string]*DependencyNode)
	for _, table := range tables {
		node := &DependencyNode{
			TableName:    table.Name,
			Level:        0,
			Dependencies: len(table.ForeignKeys),
			Dependents:   0,
		}
		graph.Nodes = append(graph.Nodes, node)
		nodeMap[table.Name] = node
	}

	// Create edges for foreign key relationships
	for _, table := range tables {
		for _, fk := range table.ForeignKeys {
			if referencedNode, exists := nodeMap[fk.ReferencedTable]; exists {
				referencedNode.Dependents++
				
				edge := &DependencyEdge{
					FromTable: table.Name,
					ToTable:   fk.ReferencedTable,
					Weight:    1.0,
					Type:      "foreign_key",
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	// Calculate dependency levels
	levels := psa.calculateDependencyLevels(graph)
	graph.Levels = levels

	return graph, nil
}

func (psa *PostgreSQLSchemaAnalyzer) detectCircularDependencies(relationshipMap *RelationshipMap) []*CyclicDependency {
	cycles := []*CyclicDependency{}
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for tableName := range relationshipMap.Tables {
		if !visited[tableName] {
			cycle := psa.findCycleDFS(tableName, relationshipMap, visited, recursionStack, []string{})
			if cycle != nil {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

func (psa *PostgreSQLSchemaAnalyzer) identifyOrphanTables(relationshipMap *RelationshipMap) []string {
	orphans := []string{}

	for tableName, tableRels := range relationshipMap.Tables {
		hasRelationships := len(tableRels.Parents) > 0 || 
			len(tableRels.Children) > 0 || 
			len(tableRels.Siblings) > 0

		if !hasRelationships {
			orphans = append(orphans, tableName)
		}
	}

	return orphans
}

func (psa *PostgreSQLSchemaAnalyzer) countTotalRelationships(relationshipMap *RelationshipMap) int {
	count := 0
	for _, tableRels := range relationshipMap.Tables {
		count += len(tableRels.Parents) + len(tableRels.Children) + 
			len(tableRels.Siblings) + len(tableRels.SelfReferences)
	}
	return count / 2 // Each relationship is counted twice
}

func (psa *PostgreSQLSchemaAnalyzer) parseCheckConstraint(constraint *Constraint) *CustomRule {
	// Simple parsing of check constraints
	return &CustomRule{
		Name:        constraint.Name,
		Description: "Check constraint: " + constraint.Expression,
		Condition:   constraint.Expression,
		Action:      "validate",
		Parameters:  map[string]interface{}{"expression": constraint.Expression},
		Priority:    1,
	}
}

func (psa *PostgreSQLSchemaAnalyzer) estimateRowConstraints(ctx context.Context, table *Table) (int, int, error) {
	// Simple estimation based on table characteristics
	minRows := 0
	maxRows := 1000000 // Default maximum

	// Adjust based on table type patterns
	tableName := strings.ToLower(table.Name)
	
	if strings.Contains(tableName, "log") || strings.Contains(tableName, "audit") {
		maxRows = 10000000 // Logs can be very large
	} else if strings.Contains(tableName, "config") || strings.Contains(tableName, "setting") {
		maxRows = 1000 // Configuration tables are typically small
	} else if strings.Contains(tableName, "user") || strings.Contains(tableName, "account") {
		maxRows = 100000 // User tables moderate size
	}

	return minRows, maxRows, nil
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeColumnDistribution(columnName string, values []interface{}) (*ColumnDistribution, error) {
	distribution := &ColumnDistribution{
		TotalValues:    int64(len(values)),
		UniqueValues:   0,
		NullValues:     0,
		ValueFrequency: make(map[string]int64),
		Percentiles:    make(map[string]interface{}),
	}

	if len(values) == 0 {
		return distribution, nil
	}

	// Count unique values and nulls
	uniqueSet := make(map[interface{}]bool)
	for _, value := range values {
		if value == nil {
			distribution.NullValues++
		} else {
			uniqueSet[value] = true
			valueStr := fmt.Sprintf("%v", value)
			distribution.ValueFrequency[valueStr]++
		}
	}
	distribution.UniqueValues = int64(len(uniqueSet))

	// Calculate basic statistics
	if distribution.TotalValues > 0 {
		nonNullValues := distribution.TotalValues - distribution.NullValues
		if nonNullValues > 0 {
			// For numeric columns, calculate statistical measures
			if psa.isNumericColumn(values) {
				numericValues := psa.extractNumericValues(values)
				if len(numericValues) > 0 {
					sort.Float64s(numericValues)
					distribution.MinValue = numericValues[0]
					distribution.MaxValue = numericValues[len(numericValues)-1]
					distribution.AvgValue = psa.calculateMean(numericValues)
					distribution.MedianValue = psa.calculateMedian(numericValues)
					distribution.StandardDev = psa.calculateStandardDeviation(numericValues)
					distribution.Variance = distribution.StandardDev * distribution.StandardDev
					distribution.Skewness = psa.calculateSkewness(numericValues)
					distribution.Kurtosis = psa.calculateKurtosis(numericValues)
				}
			}
		}
	}

	// Determine distribution type
	distribution.DistributionType = psa.classifyDistribution(distribution)

	return distribution, nil
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeColumnCorrelations(sampleData *SampleData) []*ColumnCorrelation {
	correlations := []*ColumnCorrelation{}
	columnNames := make([]string, 0, len(sampleData.Columns))
	
	for colName := range sampleData.Columns {
		columnNames = append(columnNames, colName)
	}

	// Calculate correlations between numeric columns
	for i := 0; i < len(columnNames); i++ {
		for j := i + 1; j < len(columnNames); j++ {
			col1, col2 := columnNames[i], columnNames[j]
			values1, values2 := sampleData.Columns[col1], sampleData.Columns[col2]

			if psa.isNumericColumn(values1) && psa.isNumericColumn(values2) {
				correlation := psa.calculatePearsonCorrelation(values1, values2)
				if !math.IsNaN(correlation) && math.Abs(correlation) > 0.1 {
					corr := &ColumnCorrelation{
						Column1:     col1,
						Column2:     col2,
						Correlation: correlation,
						Type:        "pearson",
						Strength:    psa.classifyCorrelationStrength(correlation),
						Significant: math.Abs(correlation) > 0.3,
					}
					correlations = append(correlations, corr)
				}
			}
		}
	}

	return correlations
}

func (psa *PostgreSQLSchemaAnalyzer) detectOutliers(sampleData *SampleData) []*Outlier {
	outliers := []*Outlier{}

	for columnName, values := range sampleData.Columns {
		if psa.isNumericColumn(values) {
			numericValues := psa.extractNumericValues(values)
			if len(numericValues) > 0 {
				columnOutliers := psa.detectColumnOutliers(columnName, numericValues)
				outliers = append(outliers, columnOutliers...)
			}
		}
	}

	return outliers
}

func (psa *PostgreSQLSchemaAnalyzer) hasTemporalColumns(sampleData *SampleData) bool {
	for _, values := range sampleData.Columns {
		if len(values) > 0 {
			switch values[0].(type) {
			case time.Time, *time.Time:
				return true
			}
		}
	}
	return false
}

func (psa *PostgreSQLSchemaAnalyzer) analyzeTemporalTrends(sampleData *SampleData) []*TrendAnalysis {
	trends := []*TrendAnalysis{}
	// Implementation would analyze temporal patterns
	// This is a placeholder for temporal trend analysis
	return trends
}

func (psa *PostgreSQLSchemaAnalyzer) determineOverallPattern(distribution *DataDistribution) string {
	if len(distribution.ColumnDistributions) == 0 {
		return "empty"
	}

	uniformCount := 0
	normalCount := 0
	skewedCount := 0

	for _, colDist := range distribution.ColumnDistributions {
		switch colDist.DistributionType {
		case "uniform":
			uniformCount++
		case "normal":
			normalCount++
		case "skewed":
			skewedCount++
		}
	}

	total := len(distribution.ColumnDistributions)
	if float64(normalCount)/float64(total) > 0.6 {
		return "normal"
	} else if float64(uniformCount)/float64(total) > 0.6 {
		return "uniform"
	} else if float64(skewedCount)/float64(total) > 0.6 {
		return "skewed"
	}

	return "mixed"
}

func (psa *PostgreSQLSchemaAnalyzer) estimateNormalizationLevel(schema *DatabaseSchema) int {
	// Simplified normalization level estimation
	if len(schema.Tables) <= 3 {
		return 1 // Likely denormalized
	}

	totalFKs := 0
	for _, table := range schema.Tables {
		totalFKs += len(table.ForeignKeys)
	}

	if totalFKs == 0 {
		return 1 // No relationships
	}

	avgFKsPerTable := float64(totalFKs) / float64(len(schema.Tables))
	if avgFKsPerTable >= 3 {
		return 4 // Highly normalized
	} else if avgFKsPerTable >= 2 {
		return 3
	} else if avgFKsPerTable >= 1 {
		return 2
	}

	return 1
}

func (psa *PostgreSQLSchemaAnalyzer) calculateComplexityScore(metrics *ComplexityMetrics) float64 {
	// Weighted complexity calculation
	tableWeight := 0.2
	columnWeight := 0.3
	relationshipWeight := 0.3
	constraintWeight := 0.2

	// Normalize values (assuming reasonable maximums)
	tableScore := math.Min(float64(metrics.TableCount)/50.0, 1.0)
	columnScore := math.Min(float64(metrics.ColumnCount)/500.0, 1.0)
	relationshipScore := math.Min(float64(metrics.RelationshipCount)/100.0, 1.0)
	constraintScore := math.Min(float64(metrics.ConstraintCount)/200.0, 1.0)

	complexity := (tableScore * tableWeight) + 
		(columnScore * columnWeight) + 
		(relationshipScore * relationshipWeight) + 
		(constraintScore * constraintWeight)

	return math.Min(complexity, 1.0)
}

func (psa *PostgreSQLSchemaAnalyzer) estimateGenerationTime(metrics *ComplexityMetrics) time.Duration {
	// Base time per table
	baseTime := 100 * time.Millisecond
	
	// Additional time based on complexity
	complexityMultiplier := 1.0 + metrics.ComplexityScore
	
	// Time per relationship (more complex)
	relationshipTime := time.Duration(metrics.RelationshipCount) * 50 * time.Millisecond
	
	totalTime := time.Duration(float64(baseTime * time.Duration(metrics.TableCount)) * complexityMultiplier) + relationshipTime
	
	return totalTime
}

func (psa *PostgreSQLSchemaAnalyzer) calculateDependencyLevels(graph *DependencyGraph) [][]string {
	levels := [][]string{}
	processed := make(map[string]bool)
	
	for {
		currentLevel := []string{}
		
		// Find nodes with no unprocessed dependencies
		for _, node := range graph.Nodes {
			if processed[node.TableName] {
				continue
			}
			
			canProcess := true
			for _, edge := range graph.Edges {
				if edge.FromTable == node.TableName && !processed[edge.ToTable] {
					canProcess = false
					break
				}
			}
			
			if canProcess {
				currentLevel = append(currentLevel, node.TableName)
			}
		}
		
		if len(currentLevel) == 0 {
			break
		}
		
		// Mark nodes as processed
		for _, tableName := range currentLevel {
			processed[tableName] = true
		}
		
		levels = append(levels, currentLevel)
	}
	
	return levels
}

func (psa *PostgreSQLSchemaAnalyzer) findCycleDFS(tableName string, relationshipMap *RelationshipMap, visited, recursionStack map[string]bool, path []string) *CyclicDependency {
	visited[tableName] = true
	recursionStack[tableName] = true
	path = append(path, tableName)

	if tableRels, exists := relationshipMap.Tables[tableName]; exists {
		for _, parent := range tableRels.Parents {
			if !visited[parent.ToTable] {
				if cycle := psa.findCycleDFS(parent.ToTable, relationshipMap, visited, recursionStack, path); cycle != nil {
					return cycle
				}
			} else if recursionStack[parent.ToTable] {
				// Found a cycle
				cycleStart := -1
				for i, table := range path {
					if table == parent.ToTable {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycleTables := path[cycleStart:]
					return &CyclicDependency{
						Tables:     cycleTables,
						BreakPoint: cycleTables[0],
						Resolution: "defer_constraint",
					}
				}
			}
		}
	}

	recursionStack[tableName] = false
	return nil
}

func (psa *PostgreSQLSchemaAnalyzer) detectTablePatterns(table *Table) []*TablePattern {
	patterns := []*TablePattern{}
	tableName := strings.ToLower(table.Name)

	// Detect common table patterns
	if strings.Contains(tableName, "log") || strings.Contains(tableName, "audit") {
		patterns = append(patterns, &TablePattern{
			Type:        "audit_log",
			Name:        "Audit Log Pattern",
			Description: "Table appears to be used for logging or auditing",
			Confidence:  0.8,
			Frequency:   1,
		})
	}

	if strings.Contains(tableName, "junction") || strings.Contains(tableName, "mapping") || 
		len(table.ForeignKeys) >= 2 && len(table.Columns) <= 4 {
		patterns = append(patterns, &TablePattern{
			Type:        "junction_table",
			Name:        "Junction Table Pattern",
			Description: "Table appears to be a many-to-many junction table",
			Confidence:  0.7,
			Frequency:   1,
		})
	}

	if strings.Contains(tableName, "config") || strings.Contains(tableName, "setting") {
		patterns = append(patterns, &TablePattern{
			Type:        "configuration",
			Name:        "Configuration Pattern",
			Description: "Table appears to store configuration data",
			Confidence:  0.9,
			Frequency:   1,
		})
	}

	return patterns
}

func (psa *PostgreSQLSchemaAnalyzer) detectColumnPatterns(column *Column) []*ColumnPattern {
	patterns := []*ColumnPattern{}
	columnName := strings.ToLower(column.Name)

	// Detect common column patterns
	if strings.HasSuffix(columnName, "_id") || columnName == "id" {
		patterns = append(patterns, &ColumnPattern{
			Type:        "identifier",
			Pattern:     "ID_PATTERN",
			Confidence:  0.9,
			Frequency:   1,
			Description: "Column appears to be an identifier",
		})
	}

	if strings.Contains(columnName, "email") {
		patterns = append(patterns, &ColumnPattern{
			Type:        "email",
			Pattern:     "EMAIL_PATTERN",
			Confidence:  0.95,
			Frequency:   1,
			Description: "Column appears to store email addresses",
		})
	}

	if strings.Contains(columnName, "phone") || strings.Contains(columnName, "tel") {
		patterns = append(patterns, &ColumnPattern{
			Type:        "phone",
			Pattern:     "PHONE_PATTERN",
			Confidence:  0.85,
			Frequency:   1,
			Description: "Column appears to store phone numbers",
		})
	}

	if strings.Contains(columnName, "date") || strings.Contains(columnName, "time") ||
		column.Type == "timestamp" || column.Type == "date" {
		patterns = append(patterns, &ColumnPattern{
			Type:        "temporal",
			Pattern:     "TEMPORAL_PATTERN",
			Confidence:  0.9,
			Frequency:   1,
			Description: "Column appears to store temporal data",
		})
	}

	return patterns
}

func (psa *PostgreSQLSchemaAnalyzer) detectNamingPatterns(schema *DatabaseSchema) []*NamingPattern {
	patterns := []*NamingPattern{}

	// Analyze table naming patterns
	tableNames := make([]string, len(schema.Tables))
	for i, table := range schema.Tables {
		tableNames[i] = table.Name
	}

	if psa.usesSnakeCase(tableNames) {
		patterns = append(patterns, &NamingPattern{
			Type:        "table_naming",
			Pattern:     "snake_case",
			Confidence:  0.8,
			Coverage:    float64(len(tableNames)),
			Description: "Tables use snake_case naming convention",
		})
	}

	if psa.usesPluralNames(tableNames) {
		patterns = append(patterns, &NamingPattern{
			Type:        "table_naming",
			Pattern:     "plural_names",
			Confidence:  0.7,
			Coverage:    float64(len(tableNames)),
			Description: "Tables use plural naming convention",
		})
	}

	return patterns
}

func (psa *PostgreSQLSchemaAnalyzer) estimateTableSize(table *Table) *TableSizeEstimate {
	estimate := &TableSizeEstimate{
		TableName:     table.Name,
		EstimatedRows: 1000, // Default estimate
		EstimatedSize: 0,
		GrowthRate:    0.1,
	}

	// Estimate based on table characteristics
	tableName := strings.ToLower(table.Name)
	
	if strings.Contains(tableName, "log") || strings.Contains(tableName, "audit") {
		estimate.EstimatedRows = 100000
		estimate.GrowthRate = 0.5 // High growth for logs
	} else if strings.Contains(tableName, "config") || strings.Contains(tableName, "setting") {
		estimate.EstimatedRows = 100
		estimate.GrowthRate = 0.01 // Low growth for config
	} else if strings.Contains(tableName, "user") || strings.Contains(tableName, "account") {
		estimate.EstimatedRows = 10000
		estimate.GrowthRate = 0.2
	}

	// Estimate size based on columns
	avgRowSize := psa.estimateAverageRowSize(table)
	estimate.EstimatedSize = estimate.EstimatedRows * avgRowSize

	return estimate
}

func (psa *PostgreSQLSchemaAnalyzer) estimateAverageRowSize(table *Table) int64 {
	size := int64(0)
	
	for _, column := range table.Columns {
		switch strings.ToLower(column.Type) {
		case "integer", "int4":
			size += 4
		case "bigint", "int8":
			size += 8
		case "smallint", "int2":
			size += 2
		case "varchar", "text":
			if column.Length > 0 {
				size += int64(column.Length)
			} else {
				size += 50 // Default text size
			}
		case "char":
			size += int64(column.Length)
		case "boolean":
			size += 1
		case "timestamp", "timestamptz":
			size += 8
		case "date":
			size += 4
		case "uuid":
			size += 16
		case "json", "jsonb":
			size += 100 // Estimated JSON size
		default:
			size += 20 // Default size for unknown types
		}
	}
	
	return size
}

func (psa *PostgreSQLSchemaAnalyzer) updateStats(analysisTime time.Duration, tableCount int) {
	psa.stats.TotalAnalyses++
	psa.stats.TotalTables += int64(tableCount)
	
	if psa.stats.TotalAnalyses == 1 {
		psa.stats.AverageAnalysisTime = analysisTime
	} else {
		// Running average
		totalTime := time.Duration(float64(psa.stats.AverageAnalysisTime) * float64(psa.stats.TotalAnalyses-1))
		psa.stats.AverageAnalysisTime = (totalTime + analysisTime) / time.Duration(psa.stats.TotalAnalyses)
	}
	
	if psa.cache.hits+psa.cache.misses > 0 {
		psa.stats.CacheHitRate = float64(psa.cache.hits) / float64(psa.cache.hits+psa.cache.misses)
	}
	
	psa.stats.LastAnalysis = time.Now()
}

// Helper functions for data analysis

func (psa *PostgreSQLSchemaAnalyzer) isNumericColumn(values []interface{}) bool {
	if len(values) == 0 {
		return false
	}
	
	for _, value := range values {
		if value == nil {
			continue
		}
		switch value.(type) {
		case int, int32, int64, float32, float64:
			return true
		default:
			return false
		}
	}
	return false
}

func (psa *PostgreSQLSchemaAnalyzer) extractNumericValues(values []interface{}) []float64 {
	result := []float64{}
	
	for _, value := range values {
		if value == nil {
			continue
		}
		
		switch v := value.(type) {
		case int:
			result = append(result, float64(v))
		case int32:
			result = append(result, float64(v))
		case int64:
			result = append(result, float64(v))
		case float32:
			result = append(result, float64(v))
		case float64:
			result = append(result, v)
		}
	}
	
	return result
}

func (psa *PostgreSQLSchemaAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (psa *PostgreSQLSchemaAnalyzer) calculateMedian(sortedValues []float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return 0
	}
	
	if n%2 == 0 {
		return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	}
	return sortedValues[n/2]
}

func (psa *PostgreSQLSchemaAnalyzer) calculateStandardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	mean := psa.calculateMean(values)
	sumSquaredDiffs := 0.0
	
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}
	
	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

func (psa *PostgreSQLSchemaAnalyzer) calculateSkewness(values []float64) float64 {
	if len(values) < 3 {
		return 0
	}
	
	mean := psa.calculateMean(values)
	stdDev := psa.calculateStandardDeviation(values)
	
	if stdDev == 0 {
		return 0
	}
	
	sumCubedDeviations := 0.0
	for _, v := range values {
		deviation := (v - mean) / stdDev
		sumCubedDeviations += deviation * deviation * deviation
	}
	
	return sumCubedDeviations / float64(len(values))
}

func (psa *PostgreSQLSchemaAnalyzer) calculateKurtosis(values []float64) float64 {
	if len(values) < 4 {
		return 0
	}
	
	mean := psa.calculateMean(values)
	stdDev := psa.calculateStandardDeviation(values)
	
	if stdDev == 0 {
		return 0
	}
	
	sumFourthPowerDeviations := 0.0
	for _, v := range values {
		deviation := (v - mean) / stdDev
		sumFourthPowerDeviations += deviation * deviation * deviation * deviation
	}
	
	return (sumFourthPowerDeviations / float64(len(values))) - 3.0 // Excess kurtosis
}

func (psa *PostgreSQLSchemaAnalyzer) classifyDistribution(distribution *ColumnDistribution) string {
	if distribution.TotalValues == 0 {
		return "empty"
	}
	
	// Simple classification based on skewness and kurtosis
	if math.Abs(distribution.Skewness) < 0.5 && math.Abs(distribution.Kurtosis) < 0.5 {
		return "normal"
	} else if math.Abs(distribution.Skewness) > 1.0 {
		return "skewed"
	} else if distribution.UniqueValues == distribution.TotalValues {
		return "uniform"
	}
	
	return "unknown"
}

func (psa *PostgreSQLSchemaAnalyzer) calculatePearsonCorrelation(values1, values2 []interface{}) float64 {
	numeric1 := psa.extractNumericValues(values1)
	numeric2 := psa.extractNumericValues(values2)
	
	if len(numeric1) != len(numeric2) || len(numeric1) < 2 {
		return math.NaN()
	}
	
	mean1 := psa.calculateMean(numeric1)
	mean2 := psa.calculateMean(numeric2)
	
	var numerator, sum1Sq, sum2Sq float64
	
	for i := 0; i < len(numeric1); i++ {
		diff1 := numeric1[i] - mean1
		diff2 := numeric2[i] - mean2
		
		numerator += diff1 * diff2
		sum1Sq += diff1 * diff1
		sum2Sq += diff2 * diff2
	}
	
	denominator := math.Sqrt(sum1Sq * sum2Sq)
	if denominator == 0 {
		return math.NaN()
	}
	
	return numerator / denominator
}

func (psa *PostgreSQLSchemaAnalyzer) classifyCorrelationStrength(correlation float64) string {
	abs := math.Abs(correlation)
	if abs >= 0.8 {
		return "very_strong"
	} else if abs >= 0.6 {
		return "strong"
	} else if abs >= 0.4 {
		return "moderate"
	} else if abs >= 0.2 {
		return "weak"
	}
	return "very_weak"
}

func (psa *PostgreSQLSchemaAnalyzer) detectColumnOutliers(columnName string, values []float64) []*Outlier {
	outliers := []*Outlier{}
	
	if len(values) < 4 {
		return outliers
	}
	
	// Calculate quartiles
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	q1 := psa.calculatePercentile(sorted, 25)
	q3 := psa.calculatePercentile(sorted, 75)
	iqr := q3 - q1
	
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr
	
	for _, value := range values {
		if value < lowerBound || value > upperBound {
			outliers = append(outliers, &Outlier{
				Column:    columnName,
				Value:     value,
				Type:      "statistical",
				Severity:  psa.calculateOutlierSeverity(value, lowerBound, upperBound),
				ZScore:    psa.calculateZScore(value, values),
			})
		}
	}
	
	return outliers
}

func (psa *PostgreSQLSchemaAnalyzer) calculatePercentile(sortedValues []float64, percentile float64) float64 {
	n := len(sortedValues)
	if n == 0 {
		return 0
	}
	
	index := (percentile / 100.0) * float64(n-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sortedValues[lower]
	}
	
	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func (psa *PostgreSQLSchemaAnalyzer) calculateOutlierSeverity(value, lowerBound, upperBound float64) string {
	if value < lowerBound {
		distance := lowerBound - value
		if distance > 3*(upperBound-lowerBound) {
			return "extreme"
		} else if distance > 1.5*(upperBound-lowerBound) {
			return "moderate"
		}
		return "mild"
	} else if value > upperBound {
		distance := value - upperBound
		if distance > 3*(upperBound-lowerBound) {
			return "extreme"
		} else if distance > 1.5*(upperBound-lowerBound) {
			return "moderate"
		}
		return "mild"
	}
	return "none"
}

func (psa *PostgreSQLSchemaAnalyzer) calculateZScore(value float64, values []float64) float64 {
	mean := psa.calculateMean(values)
	stdDev := psa.calculateStandardDeviation(values)
	
	if stdDev == 0 {
		return 0
	}
	
	return (value - mean) / stdDev
}

func (psa *PostgreSQLSchemaAnalyzer) usesSnakeCase(names []string) bool {
	snakeCaseCount := 0
	for _, name := range names {
		if strings.Contains(name, "_") && strings.ToLower(name) == name {
			snakeCaseCount++
		}
	}
	return float64(snakeCaseCount)/float64(len(names)) > 0.7
}

func (psa *PostgreSQLSchemaAnalyzer) usesPluralNames(names []string) bool {
	pluralCount := 0
	for _, name := range names {
		if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
			pluralCount++
		}
	}
	return float64(pluralCount)/float64(len(names)) > 0.7
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Additional data structures for schema analysis

// SampleData represents a sample of data from tables
type SampleData struct {
	TableName string                            `json:"table_name"`
	Columns   map[string][]interface{}          `json:"columns"`
	RowCount  int                               `json:"row_count"`
	Metadata  map[string]interface{}            `json:"metadata"`
}

// DataDistribution represents overall data distribution patterns
type DataDistribution struct {
	OverallPattern      string                         `json:"overall_pattern"`
	ColumnDistributions map[string]*ColumnDistribution `json:"column_distributions"`
	Correlations        []*ColumnCorrelation           `json:"correlations"`
	Outliers            []*Outlier                     `json:"outliers"`
	Trends              []*TrendAnalysis               `json:"trends"`
}

// Outlier represents a statistical outlier in data
type Outlier struct {
	Column    string      `json:"column"`
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`
	Severity  string      `json:"severity"`
	ZScore    float64     `json:"z_score"`
}

// ConstraintMap organizes all constraints in a schema
type ConstraintMap struct {
	Tables         map[string]*TableConstraints `json:"tables"`
	GlobalRules    []*GlobalConstraintRule      `json:"global_rules"`
	IntegrityRules []*IntegrityRule             `json:"integrity_rules"`
	BusinessRules  []*BusinessRule              `json:"business_rules"`
}

// GlobalConstraintRule represents schema-wide constraint rules
type GlobalConstraintRule struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Tables      []string               `json:"tables"`
	Columns     []string               `json:"columns"`
	Parameters  map[string]interface{} `json:"parameters"`
	IsActive    bool                   `json:"is_active"`
}

// IntegrityRule represents referential integrity rules
type IntegrityRule struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	FromTable     string   `json:"from_table"`
	ToTable       string   `json:"to_table"`
	FromColumns   []string `json:"from_columns"`
	ToColumns     []string `json:"to_columns"`
	OnUpdate      string   `json:"on_update"`
	OnDelete      string   `json:"on_delete"`
	IsEnforced    bool     `json:"is_enforced"`
}

// BusinessRule represents business logic rules
type BusinessRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tables      []string               `json:"tables"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`
}

// DataPatterns contains detected patterns in the schema
type DataPatterns struct {
	TablePatterns  []*TablePattern   `json:"table_patterns"`
	ColumnPatterns []*ColumnPattern  `json:"column_patterns"`
	DataPatterns   []*DetectedPattern `json:"data_patterns"`
	NamingPatterns []*NamingPattern  `json:"naming_patterns"`
}

// DetectedPattern represents a detected data pattern
type DetectedPattern struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Frequency   int64                  `json:"frequency"`
	Tables      []string               `json:"tables"`
	Columns     []string               `json:"columns"`
	Parameters  map[string]interface{} `json:"parameters"`
	Examples    []string               `json:"examples"`
}

// NamingPattern represents naming convention patterns
type NamingPattern struct {
	Type        string  `json:"type"`
	Pattern     string  `json:"pattern"`
	Confidence  float64 `json:"confidence"`
	Coverage    float64 `json:"coverage"`
	Examples    []string `json:"examples"`
	Description string  `json:"description"`
}

// SizeEstimates contains size estimation data
type SizeEstimates struct {
	TableSizes     map[string]*TableSizeEstimate `json:"table_sizes"`
	TotalRows      int64                         `json:"total_rows"`
	TotalDataSize  int64                         `json:"total_data_size_bytes"`
	GrowthRate     float64                       `json:"growth_rate"`
	ProjectedSizes map[string]*ProjectedSize     `json:"projected_sizes"`
}

// TableSizeEstimate estimates the size of a table
type TableSizeEstimate struct {
	TableName     string  `json:"table_name"`
	EstimatedRows int64   `json:"estimated_rows"`
	EstimatedSize int64   `json:"estimated_size_bytes"`
	GrowthRate    float64 `json:"growth_rate"`
	Confidence    float64 `json:"confidence"`
}

// ProjectedSize represents projected future size
type ProjectedSize struct {
	TableName     string    `json:"table_name"`
	ProjectedDate time.Time `json:"projected_date"`
	ProjectedRows int64     `json:"projected_rows"`
	ProjectedSize int64     `json:"projected_size_bytes"`
	Methodology   string    `json:"methodology"`
}

// GenerationRecommendations provides recommendations for data generation
type GenerationRecommendations struct {
	Strategy                   GenerationStrategy         `json:"strategy"`
	QualityLevel               QualityLevel               `json:"quality_level"`
	PrivacyLevel               PrivacyLevel               `json:"privacy_level"`
	VolumeRecommendations      []*VolumeRecommendation    `json:"volume_recommendations"`
	QualityRecommendations     []*QualityRecommendation   `json:"quality_recommendations"`
	PerformanceRecommendations []*PerformanceRecommendation `json:"performance_recommendations"`
	SecurityRecommendations    []*SecurityRecommendation  `json:"security_recommendations"`
}

// VolumeRecommendation recommends data volumes
type VolumeRecommendation struct {
	TableName        string  `json:"table_name"`
	RecommendedRows  int     `json:"recommended_rows"`
	MinRows          int     `json:"min_rows"`
	MaxRows          int     `json:"max_rows"`
	ScalingFactor    float64 `json:"scaling_factor"`
	DistributionType string  `json:"distribution_type"`
	Rationale        string  `json:"rationale"`
}

// QualityRecommendation recommends quality settings
type QualityRecommendation struct {
	Area         string                 `json:"area"`
	Recommendation string               `json:"recommendation"`
	Priority     string                 `json:"priority"`
	Impact       string                 `json:"impact"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// PerformanceRecommendation recommends performance optimizations
type PerformanceRecommendation struct {
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Priority     string                 `json:"priority"`
	ExpectedGain string                 `json:"expected_gain"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// SecurityRecommendation recommends security measures
type SecurityRecommendation struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
}