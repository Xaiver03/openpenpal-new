package performance

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// AIBottleneckDetector implements the BottleneckDetector interface with AI-driven analysis
type AIBottleneckDetector struct {
	config              *BottleneckDetectorConfig
	resourceAnalyzer    *ResourceAnalyzer
	patternDetector     *BottleneckPatternDetector
	rootCauseEngine     *RootCauseEngine
	optimizationEngine  *OptimizationEngine
	historicalData      *HistoricalBottleneckData
	mlModels           map[string]MLModel
	metrics            *DetectionMetrics
}

// BottleneckDetectorConfig configures the bottleneck detection system
type BottleneckDetectorConfig struct {
	DetectionThresholds    *DetectionThresholds    `json:"detection_thresholds"`
	AnalysisWindow         time.Duration           `json:"analysis_window"`
	SamplingInterval       time.Duration           `json:"sampling_interval"`
	EnableMLPrediction     bool                    `json:"enable_ml_prediction"`
	EnableRootCauseAnalysis bool                   `json:"enable_root_cause_analysis"`
	MinConfidenceLevel     float64                 `json:"min_confidence_level"`
	MaxBottlenecksPerRun   int                     `json:"max_bottlenecks_per_run"`
	HistoricalDataRetention time.Duration          `json:"historical_data_retention"`
	NotificationThresholds *NotificationThresholds `json:"notification_thresholds"`
}

// DetectionThresholds defines thresholds for bottleneck detection
type DetectionThresholds struct {
	CPUUsageHigh          float64 `json:"cpu_usage_high"`
	CPUUsageCritical      float64 `json:"cpu_usage_critical"`
	MemoryUsageHigh       float64 `json:"memory_usage_high"`
	MemoryUsageCritical   float64 `json:"memory_usage_critical"`
	DiskUsageHigh         float64 `json:"disk_usage_high"`
	DiskUsageCritical     float64 `json:"disk_usage_critical"`
	NetworkLatencyHigh    time.Duration `json:"network_latency_high"`
	NetworkLatencyCritical time.Duration `json:"network_latency_critical"`
	ResponseTimeHigh      time.Duration `json:"response_time_high"`
	ResponseTimeCritical  time.Duration `json:"response_time_critical"`
	ErrorRateHigh         float64 `json:"error_rate_high"`
	ErrorRateCritical     float64 `json:"error_rate_critical"`
	ThroughputLow         float64 `json:"throughput_low"`
	ThroughputCritical    float64 `json:"throughput_critical"`
}

// NotificationThresholds defines when to trigger notifications
type NotificationThresholds struct {
	CriticalSeverityImmediate bool          `json:"critical_severity_immediate"`
	HighSeverityDelay         time.Duration `json:"high_severity_delay"`
	MediumSeverityDelay       time.Duration `json:"medium_severity_delay"`
	LowSeverityDelay          time.Duration `json:"low_severity_delay"`
}

// NewAIBottleneckDetector creates a new AI-powered bottleneck detector
func NewAIBottleneckDetector(config *BottleneckDetectorConfig) *AIBottleneckDetector {
	if config == nil {
		config = getDefaultBottleneckDetectorConfig()
	}

	return &AIBottleneckDetector{
		config:              config,
		resourceAnalyzer:    NewResourceAnalyzer(),
		patternDetector:     NewBottleneckPatternDetector(),
		rootCauseEngine:     NewRootCauseEngine(),
		optimizationEngine:  NewOptimizationEngine(),
		historicalData:      NewHistoricalBottleneckData(config.HistoricalDataRetention),
		mlModels:           make(map[string]MLModel),
		metrics:            NewDetectionMetrics(),
	}
}

// DetectBottlenecks identifies performance bottlenecks in the system
func (d *AIBottleneckDetector) DetectBottlenecks(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error) {
	startTime := time.Now()
	d.metrics.IncrementDetectionRuns()

	// Step 1: Resource-based bottleneck detection
	resourceBottlenecks, err := d.detectResourceBottlenecks(ctx, metrics)
	if err != nil {
		d.metrics.IncrementDetectionErrors()
		return nil, fmt.Errorf("resource bottleneck detection failed: %w", err)
	}

	// Step 2: Pattern-based bottleneck detection
	patternBottlenecks, err := d.detectPatternBottlenecks(ctx, metrics)
	if err != nil {
		d.metrics.IncrementDetectionErrors()
		return nil, fmt.Errorf("pattern bottleneck detection failed: %w", err)
	}

	// Step 3: ML-based bottleneck prediction
	var mlBottlenecks []*Bottleneck
	if d.config.EnableMLPrediction {
		mlBottlenecks, err = d.detectMLBottlenecks(ctx, metrics)
		if err != nil {
			// Log warning but don't fail the entire detection
			d.metrics.IncrementMLPredictionErrors()
		}
	}

	// Step 4: Combine and deduplicate bottlenecks
	allBottlenecks := append(resourceBottlenecks, patternBottlenecks...)
	allBottlenecks = append(allBottlenecks, mlBottlenecks...)
	
	deduplicatedBottlenecks := d.deduplicateBottlenecks(allBottlenecks)

	// Step 5: Prioritize and filter bottlenecks
	prioritizedBottlenecks := d.prioritizeBottlenecks(deduplicatedBottlenecks)
	
	// Limit the number of bottlenecks returned
	if len(prioritizedBottlenecks) > d.config.MaxBottlenecksPerRun {
		prioritizedBottlenecks = prioritizedBottlenecks[:d.config.MaxBottlenecksPerRun]
	}

	// Step 6: Update historical data
	d.historicalData.AddDetectionRun(DetectionRun{
		Timestamp:  time.Now(),
		Metrics:    metrics,
		Bottlenecks: prioritizedBottlenecks,
		Duration:   time.Since(startTime),
	})

	d.metrics.RecordDetectionDuration(time.Since(startTime))
	d.metrics.RecordBottlenecksDetected(len(prioritizedBottlenecks))

	return prioritizedBottlenecks, nil
}

// detectResourceBottlenecks detects bottlenecks based on resource utilization
func (d *AIBottleneckDetector) detectResourceBottlenecks(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error) {
	var bottlenecks []*Bottleneck

	// CPU bottlenecks
	if cpuBottlenecks := d.analyzeCPUBottlenecks(metrics.ResourceUsage.CPU); len(cpuBottlenecks) > 0 {
		bottlenecks = append(bottlenecks, cpuBottlenecks...)
	}

	// Memory bottlenecks
	if memoryBottlenecks := d.analyzeMemoryBottlenecks(metrics.ResourceUsage.Memory); len(memoryBottlenecks) > 0 {
		bottlenecks = append(bottlenecks, memoryBottlenecks...)
	}

	// Disk bottlenecks
	if diskBottlenecks := d.analyzeDiskBottlenecks(metrics.ResourceUsage.Disk); len(diskBottlenecks) > 0 {
		bottlenecks = append(bottlenecks, diskBottlenecks...)
	}

	// Network bottlenecks
	if networkBottlenecks := d.analyzeNetworkBottlenecks(metrics.ResourceUsage.Network); len(networkBottlenecks) > 0 {
		bottlenecks = append(bottlenecks, networkBottlenecks...)
	}

	return bottlenecks, nil
}

// analyzeCPUBottlenecks analyzes CPU metrics for bottlenecks
func (d *AIBottleneckDetector) analyzeCPUBottlenecks(cpuMetrics *CPUMetrics) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// High CPU usage bottleneck
	if cpuMetrics.AverageUsage >= d.config.DetectionThresholds.CPUUsageHigh {
		severity := SeverityHigh
		if cpuMetrics.AverageUsage >= d.config.DetectionThresholds.CPUUsageCritical {
			severity = SeverityCritical
		}

		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("cpu", "high_usage"),
			Type:        BottleneckTypeCPU,
			Component:   "cpu",
			Description: fmt.Sprintf("High CPU usage detected: %.2f%%", cpuMetrics.AverageUsage),
			Severity:    severity,
			Impact:      calculateCPUImpact(cpuMetrics.AverageUsage),
			Location:    &BottleneckLocation{Component: "system", Function: "cpu"},
			Metrics: map[string]float64{
				"average_usage": cpuMetrics.AverageUsage,
				"peak_usage":    cpuMetrics.PeakUsage,
				"load_1min":     cpuMetrics.LoadAverage.Load1Min,
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// High context switches
	if cpuMetrics.ContextSwitches > 100000 { // Threshold for high context switches
		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("cpu", "context_switches"),
			Type:        BottleneckTypeCPU,
			Component:   "cpu",
			Description: fmt.Sprintf("High context switches detected: %d", cpuMetrics.ContextSwitches),
			Severity:    SeverityMedium,
			Impact:      calculateContextSwitchImpact(cpuMetrics.ContextSwitches),
			Location:    &BottleneckLocation{Component: "system", Function: "scheduler"},
			Metrics: map[string]float64{
				"context_switches": float64(cpuMetrics.ContextSwitches),
				"interrupts":       float64(cpuMetrics.Interrupts),
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// analyzeMemoryBottlenecks analyzes memory metrics for bottlenecks
func (d *AIBottleneckDetector) analyzeMemoryBottlenecks(memoryMetrics *MemoryMetrics) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// Calculate memory usage percentage
	memoryUsagePercent := float64(memoryMetrics.UsedMemory) / float64(memoryMetrics.TotalMemory) * 100

	// High memory usage bottleneck
	if memoryUsagePercent >= d.config.DetectionThresholds.MemoryUsageHigh {
		severity := SeverityHigh
		if memoryUsagePercent >= d.config.DetectionThresholds.MemoryUsageCritical {
			severity = SeverityCritical
		}

		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("memory", "high_usage"),
			Type:        BottleneckTypeMemory,
			Component:   "memory",
			Description: fmt.Sprintf("High memory usage detected: %.2f%%", memoryUsagePercent),
			Severity:    severity,
			Impact:      calculateMemoryImpact(memoryUsagePercent),
			Location:    &BottleneckLocation{Component: "system", Function: "memory"},
			Metrics: map[string]float64{
				"usage_percent":   memoryUsagePercent,
				"used_memory":     float64(memoryMetrics.UsedMemory),
				"memory_pressure": memoryMetrics.MemoryPressure,
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Memory pressure bottleneck
	if memoryMetrics.MemoryPressure >= 0.8 {
		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("memory", "pressure"),
			Type:        BottleneckTypeMemory,
			Component:   "memory",
			Description: fmt.Sprintf("High memory pressure detected: %.2f", memoryMetrics.MemoryPressure),
			Severity:    SeverityHigh,
			Impact:      memoryMetrics.MemoryPressure * 100,
			Location:    &BottleneckLocation{Component: "system", Function: "memory_management"},
			Metrics: map[string]float64{
				"memory_pressure": memoryMetrics.MemoryPressure,
				"swap_used":       float64(memoryMetrics.SwapUsed),
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Excessive garbage collection
	if memoryMetrics.GCTime > time.Second {
		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("memory", "gc_overhead"),
			Type:        BottleneckTypeMemory,
			Component:   "gc",
			Description: fmt.Sprintf("Excessive GC overhead detected: %v", memoryMetrics.GCTime),
			Severity:    SeverityMedium,
			Impact:      calculateGCImpact(memoryMetrics.GCTime),
			Location:    &BottleneckLocation{Component: "runtime", Function: "garbage_collector"},
			Metrics: map[string]float64{
				"gc_time_ms": float64(memoryMetrics.GCTime.Milliseconds()),
				"gc_count":   float64(memoryMetrics.GCCount),
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// analyzeDiskBottlenecks analyzes disk metrics for bottlenecks
func (d *AIBottleneckDetector) analyzeDiskBottlenecks(diskMetrics *DiskMetrics) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// High disk usage
	if diskMetrics.DiskUsage >= d.config.DetectionThresholds.DiskUsageHigh {
		severity := SeverityHigh
		if diskMetrics.DiskUsage >= d.config.DetectionThresholds.DiskUsageCritical {
			severity = SeverityCritical
		}

		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("disk", "high_usage"),
			Type:        BottleneckTypeDisk,
			Component:   "disk",
			Description: fmt.Sprintf("High disk usage detected: %.2f%%", diskMetrics.DiskUsage),
			Severity:    severity,
			Impact:      diskMetrics.DiskUsage,
			Location:    &BottleneckLocation{Component: "storage", Function: "disk"},
			Metrics: map[string]float64{
				"disk_usage": diskMetrics.DiskUsage,
				"queue_depth": float64(diskMetrics.QueueDepth),
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// High disk latency
	if diskMetrics.ReadLatency > time.Millisecond*100 || diskMetrics.WriteLatency > time.Millisecond*100 {
		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("disk", "high_latency"),
			Type:        BottleneckTypeDisk,
			Component:   "disk",
			Description: fmt.Sprintf("High disk latency detected: read=%v, write=%v", diskMetrics.ReadLatency, diskMetrics.WriteLatency),
			Severity:    SeverityMedium,
			Impact:      calculateDiskLatencyImpact(diskMetrics.ReadLatency, diskMetrics.WriteLatency),
			Location:    &BottleneckLocation{Component: "storage", Function: "io"},
			Metrics: map[string]float64{
				"read_latency_ms":  float64(diskMetrics.ReadLatency.Milliseconds()),
				"write_latency_ms": float64(diskMetrics.WriteLatency.Milliseconds()),
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// analyzeNetworkBottlenecks analyzes network metrics for bottlenecks
func (d *AIBottleneckDetector) analyzeNetworkBottlenecks(networkMetrics *NetworkMetrics) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// High network latency
	if networkMetrics.Latency >= d.config.DetectionThresholds.NetworkLatencyHigh {
		severity := SeverityHigh
		if networkMetrics.Latency >= d.config.DetectionThresholds.NetworkLatencyCritical {
			severity = SeverityCritical
		}

		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("network", "high_latency"),
			Type:        BottleneckTypeNetwork,
			Component:   "network",
			Description: fmt.Sprintf("High network latency detected: %v", networkMetrics.Latency),
			Severity:    severity,
			Impact:      calculateNetworkLatencyImpact(networkMetrics.Latency),
			Location:    &BottleneckLocation{Component: "network", Function: "routing"},
			Metrics: map[string]float64{
				"latency_ms":    float64(networkMetrics.Latency.Milliseconds()),
				"packet_loss":   networkMetrics.PacketLoss,
				"error_rate":    networkMetrics.ErrorRate,
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// High packet loss
	if networkMetrics.PacketLoss >= 1.0 { // 1% packet loss threshold
		bottleneck := &Bottleneck{
			ID:          generateBottleneckID("network", "packet_loss"),
			Type:        BottleneckTypeNetwork,
			Component:   "network",
			Description: fmt.Sprintf("High packet loss detected: %.2f%%", networkMetrics.PacketLoss),
			Severity:    SeverityHigh,
			Impact:      networkMetrics.PacketLoss * 10, // Amplify impact
			Location:    &BottleneckLocation{Component: "network", Function: "transmission"},
			Metrics: map[string]float64{
				"packet_loss": networkMetrics.PacketLoss,
				"error_rate":  networkMetrics.ErrorRate,
			},
			DetectedAt: time.Now(),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectPatternBottlenecks detects bottlenecks based on patterns
func (d *AIBottleneckDetector) detectPatternBottlenecks(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error) {
	return d.patternDetector.DetectPatterns(ctx, metrics)
}

// detectMLBottlenecks uses machine learning to predict bottlenecks
func (d *AIBottleneckDetector) detectMLBottlenecks(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error) {
	var bottlenecks []*Bottleneck

	for modelName, model := range d.mlModels {
		prediction, err := model.Predict(ctx, metrics)
		if err != nil {
			continue // Skip failed predictions
		}

		if prediction.Confidence >= d.config.MinConfidenceLevel {
			bottleneck := d.convertPredictionToBottleneck(modelName, prediction)
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks, nil
}

// AnalyzeRootCause performs root cause analysis for a bottleneck
func (d *AIBottleneckDetector) AnalyzeRootCause(ctx context.Context, bottleneck *Bottleneck) (*RootCauseAnalysis, error) {
	if !d.config.EnableRootCauseAnalysis {
		return nil, fmt.Errorf("root cause analysis is disabled")
	}

	return d.rootCauseEngine.AnalyzeRootCause(ctx, bottleneck, d.historicalData)
}

// PrioritizeBottlenecks sorts bottlenecks by priority
func (d *AIBottleneckDetector) PrioritizeBottlenecks(ctx context.Context, bottlenecks []*Bottleneck) ([]*PrioritizedBottleneck, error) {
	return d.prioritizeBottlenecks(bottlenecks), nil
}

// RecommendOptimizations provides optimization recommendations for a bottleneck
func (d *AIBottleneckDetector) RecommendOptimizations(ctx context.Context, bottleneck *Bottleneck) ([]*OptimizationRecommendation, error) {
	return d.optimizationEngine.GenerateRecommendations(ctx, bottleneck)
}

// AnalyzeResourceContention analyzes resource contention
func (d *AIBottleneckDetector) AnalyzeResourceContention(ctx context.Context, metrics *ResourceMetrics) (*ContentionAnalysis, error) {
	return d.resourceAnalyzer.AnalyzeContention(ctx, metrics)
}

// DetectMemoryLeaks detects memory leaks in the system
func (d *AIBottleneckDetector) DetectMemoryLeaks(ctx context.Context, memoryMetrics *MemoryMetrics) (*MemoryLeakReport, error) {
	return d.resourceAnalyzer.DetectMemoryLeaks(ctx, memoryMetrics)
}

// AnalyzeCPUHotspots analyzes CPU hotspots
func (d *AIBottleneckDetector) AnalyzeCPUHotspots(ctx context.Context, cpuMetrics *CPUMetrics) (*HotspotAnalysis, error) {
	return d.resourceAnalyzer.AnalyzeCPUHotspots(ctx, cpuMetrics)
}

// DetectDatabaseBottlenecks detects database-specific bottlenecks
func (d *AIBottleneckDetector) DetectDatabaseBottlenecks(ctx context.Context, dbMetrics *DatabaseMetrics) ([]*DatabaseBottleneck, error) {
	return d.resourceAnalyzer.DetectDatabaseBottlenecks(ctx, dbMetrics)
}

// Helper functions

func (d *AIBottleneckDetector) deduplicateBottlenecks(bottlenecks []*Bottleneck) []*Bottleneck {
	seen := make(map[string]bool)
	var unique []*Bottleneck

	for _, bottleneck := range bottlenecks {
		key := fmt.Sprintf("%s-%s", bottleneck.Type, bottleneck.Component)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, bottleneck)
		}
	}

	return unique
}

func (d *AIBottleneckDetector) prioritizeBottlenecks(bottlenecks []*Bottleneck) []*PrioritizedBottleneck {
	var prioritized []*PrioritizedBottleneck

	for _, bottleneck := range bottlenecks {
		priority := calculateBottleneckPriority(bottleneck)
		prioritized = append(prioritized, &PrioritizedBottleneck{
			Bottleneck: bottleneck,
			Priority:   priority,
			Score:      calculateBottleneckScore(bottleneck),
		})
	}

	// Sort by priority (higher first)
	sort.Slice(prioritized, func(i, j int) bool {
		return prioritized[i].Score > prioritized[j].Score
	})

	return prioritized
}

func (d *AIBottleneckDetector) convertPredictionToBottleneck(modelName string, prediction *MLPrediction) *Bottleneck {
	return &Bottleneck{
		ID:          generateBottleneckID("ml", modelName),
		Type:        BottleneckType(prediction.Type),
		Component:   prediction.Component,
		Description: fmt.Sprintf("ML-predicted bottleneck: %s (confidence: %.2f)", prediction.Description, prediction.Confidence),
		Severity:    determineSeverityFromPrediction(prediction),
		Impact:      prediction.ExpectedImpact,
		Location:    &BottleneckLocation{Component: prediction.Component},
		Metrics:     prediction.Metrics,
		DetectedAt:  time.Now(),
	}
}

// Utility functions

func generateBottleneckID(category, subtype string) string {
	return fmt.Sprintf("%s-%s-%d", category, subtype, time.Now().Unix())
}

func calculateCPUImpact(usage float64) float64 {
	if usage >= 95 {
		return 100
	} else if usage >= 80 {
		return usage * 1.2
	}
	return usage
}

func calculateMemoryImpact(usage float64) float64 {
	if usage >= 90 {
		return 100
	} else if usage >= 75 {
		return usage * 1.3
	}
	return usage
}

func calculateGCImpact(gcTime time.Duration) float64 {
	// GC time over 1 second is considered high impact
	seconds := gcTime.Seconds()
	if seconds >= 1 {
		return math.Min(100, seconds*50)
	}
	return seconds * 25
}

func calculateDiskLatencyImpact(readLatency, writeLatency time.Duration) float64 {
	avgLatency := (readLatency + writeLatency) / 2
	ms := avgLatency.Milliseconds()
	if ms >= 1000 {
		return 100
	} else if ms >= 500 {
		return 80
	} else if ms >= 100 {
		return 60
	}
	return float64(ms) / 10
}

func calculateNetworkLatencyImpact(latency time.Duration) float64 {
	ms := latency.Milliseconds()
	if ms >= 1000 {
		return 100
	} else if ms >= 500 {
		return 80
	} else if ms >= 200 {
		return 60
	}
	return float64(ms) / 10
}

func calculateContextSwitchImpact(contextSwitches int64) float64 {
	if contextSwitches >= 1000000 {
		return 100
	} else if contextSwitches >= 500000 {
		return 80
	} else if contextSwitches >= 100000 {
		return 60
	}
	return float64(contextSwitches) / 10000
}

func calculateBottleneckPriority(bottleneck *Bottleneck) int {
	switch bottleneck.Severity {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

func calculateBottleneckScore(bottleneck *Bottleneck) float64 {
	priorityWeight := float64(calculateBottleneckPriority(bottleneck)) * 25
	impactWeight := bottleneck.Impact * 0.5
	
	// Add type-specific weighting
	typeWeight := getBottleneckTypeWeight(bottleneck.Type)
	
	return priorityWeight + impactWeight + typeWeight
}

func getBottleneckTypeWeight(bottleneckType BottleneckType) float64 {
	switch bottleneckType {
	case BottleneckTypeCPU:
		return 20
	case BottleneckTypeMemory:
		return 18
	case BottleneckTypeDatabase:
		return 16
	case BottleneckTypeNetwork:
		return 14
	case BottleneckTypeDisk:
		return 12
	case BottleneckTypeAPI:
		return 10
	case BottleneckTypeCache:
		return 8
	case BottleneckTypeLock:
		return 6
	case BottleneckTypeQueue:
		return 4
	default:
		return 0
	}
}

func determineSeverityFromPrediction(prediction *MLPrediction) BottleneckSeverity {
	if prediction.Confidence >= 0.9 && prediction.ExpectedImpact >= 80 {
		return SeverityCritical
	} else if prediction.Confidence >= 0.8 && prediction.ExpectedImpact >= 60 {
		return SeverityHigh
	} else if prediction.Confidence >= 0.7 && prediction.ExpectedImpact >= 40 {
		return SeverityMedium
	}
	return SeverityLow
}

func getDefaultBottleneckDetectorConfig() *BottleneckDetectorConfig {
	return &BottleneckDetectorConfig{
		DetectionThresholds: &DetectionThresholds{
			CPUUsageHigh:           80.0,
			CPUUsageCritical:       95.0,
			MemoryUsageHigh:        85.0,
			MemoryUsageCritical:    95.0,
			DiskUsageHigh:          90.0,
			DiskUsageCritical:      98.0,
			NetworkLatencyHigh:     time.Millisecond * 500,
			NetworkLatencyCritical: time.Second * 2,
			ResponseTimeHigh:       time.Second * 2,
			ResponseTimeCritical:   time.Second * 5,
			ErrorRateHigh:          5.0,
			ErrorRateCritical:      10.0,
			ThroughputLow:          100.0,
			ThroughputCritical:     50.0,
		},
		AnalysisWindow:             time.Minute * 5,
		SamplingInterval:           time.Second * 30,
		EnableMLPrediction:         true,
		EnableRootCauseAnalysis:    true,
		MinConfidenceLevel:         0.7,
		MaxBottlenecksPerRun:       10,
		HistoricalDataRetention:    time.Hour * 24 * 7,
		NotificationThresholds: &NotificationThresholds{
			CriticalSeverityImmediate: true,
			HighSeverityDelay:         time.Minute * 2,
			MediumSeverityDelay:       time.Minute * 5,
			LowSeverityDelay:          time.Minute * 15,
		},
	}
}

// Supporting types for bottleneck detection

// PrioritizedBottleneck represents a bottleneck with priority information
type PrioritizedBottleneck struct {
	*Bottleneck
	Priority int     `json:"priority"`
	Score    float64 `json:"score"`
}

// DetectionRun represents a single bottleneck detection run
type DetectionRun struct {
	Timestamp   time.Time             `json:"timestamp"`
	Metrics     *PerformanceMetrics   `json:"metrics"`
	Bottlenecks []*Bottleneck         `json:"bottlenecks"`
	Duration    time.Duration         `json:"duration"`
}

// DetectionMetrics tracks bottleneck detection performance
type DetectionMetrics struct {
	TotalDetectionRuns      int64         `json:"total_detection_runs"`
	TotalBottlenecksDetected int64         `json:"total_bottlenecks_detected"`
	TotalDetectionErrors    int64         `json:"total_detection_errors"`
	MLPredictionErrors      int64         `json:"ml_prediction_errors"`
	AverageDetectionTime    time.Duration `json:"average_detection_time"`
	LastDetectionTime       time.Time     `json:"last_detection_time"`
}

// NewDetectionMetrics creates new detection metrics
func NewDetectionMetrics() *DetectionMetrics {
	return &DetectionMetrics{}
}

// IncrementDetectionRuns increments the detection run counter
func (m *DetectionMetrics) IncrementDetectionRuns() {
	m.TotalDetectionRuns++
	m.LastDetectionTime = time.Now()
}

// IncrementDetectionErrors increments the detection error counter
func (m *DetectionMetrics) IncrementDetectionErrors() {
	m.TotalDetectionErrors++
}

// IncrementMLPredictionErrors increments the ML prediction error counter
func (m *DetectionMetrics) IncrementMLPredictionErrors() {
	m.MLPredictionErrors++
}

// RecordDetectionDuration records the duration of a detection run
func (m *DetectionMetrics) RecordDetectionDuration(duration time.Duration) {
	// Simple moving average calculation
	if m.TotalDetectionRuns == 1 {
		m.AverageDetectionTime = duration
	} else {
		m.AverageDetectionTime = (m.AverageDetectionTime + duration) / 2
	}
}

// RecordBottlenecksDetected records the number of bottlenecks detected
func (m *DetectionMetrics) RecordBottlenecksDetected(count int) {
	m.TotalBottlenecksDetected += int64(count)
}

// MLPrediction represents a machine learning prediction
type MLPrediction struct {
	Type           string             `json:"type"`
	Component      string             `json:"component"`
	Description    string             `json:"description"`
	Confidence     float64            `json:"confidence"`
	ExpectedImpact float64            `json:"expected_impact"`
	Metrics        map[string]float64 `json:"metrics"`
}

// Additional types that need to be implemented

// HistoricalBottleneckData stores historical bottleneck detection data
type HistoricalBottleneckData struct {
	RetentionPeriod time.Duration    `json:"retention_period"`
	DetectionRuns   []*DetectionRun  `json:"detection_runs"`
}

// NewHistoricalBottleneckData creates new historical data storage
func NewHistoricalBottleneckData(retention time.Duration) *HistoricalBottleneckData {
	return &HistoricalBottleneckData{
		RetentionPeriod: retention,
		DetectionRuns:   make([]*DetectionRun, 0),
	}
}

// AddDetectionRun adds a new detection run to historical data
func (h *HistoricalBottleneckData) AddDetectionRun(run DetectionRun) {
	h.DetectionRuns = append(h.DetectionRuns, &run)
	h.cleanupOldRuns()
}

// cleanupOldRuns removes old detection runs based on retention period
func (h *HistoricalBottleneckData) cleanupOldRuns() {
	cutoff := time.Now().Add(-h.RetentionPeriod)
	var validRuns []*DetectionRun
	
	for _, run := range h.DetectionRuns {
		if run.Timestamp.After(cutoff) {
			validRuns = append(validRuns, run)
		}
	}
	
	h.DetectionRuns = validRuns
}

// Placeholder types for supporting components

type ResourceAnalyzer struct{}
type BottleneckPatternDetector struct{}
type RootCauseEngine struct{}
type OptimizationEngine struct{}
type ContentionAnalysis struct{}
type MemoryLeakReport struct{}
type HotspotAnalysis struct{}
type DatabaseBottleneck struct{}
type DatabaseMetrics struct{}
type ResourceMetrics struct{}

// Placeholder constructors
func NewResourceAnalyzer() *ResourceAnalyzer { return &ResourceAnalyzer{} }
func NewBottleneckPatternDetector() *BottleneckPatternDetector { return &BottleneckPatternDetector{} }
func NewRootCauseEngine() *RootCauseEngine { return &RootCauseEngine{} }
func NewOptimizationEngine() *OptimizationEngine { return &OptimizationEngine{} }

// Placeholder methods
func (r *ResourceAnalyzer) AnalyzeContention(ctx context.Context, metrics *ResourceMetrics) (*ContentionAnalysis, error) {
	return &ContentionAnalysis{}, nil
}

func (r *ResourceAnalyzer) DetectMemoryLeaks(ctx context.Context, memoryMetrics *MemoryMetrics) (*MemoryLeakReport, error) {
	return &MemoryLeakReport{}, nil
}

func (r *ResourceAnalyzer) AnalyzeCPUHotspots(ctx context.Context, cpuMetrics *CPUMetrics) (*HotspotAnalysis, error) {
	return &HotspotAnalysis{}, nil
}

func (r *ResourceAnalyzer) DetectDatabaseBottlenecks(ctx context.Context, dbMetrics *DatabaseMetrics) ([]*DatabaseBottleneck, error) {
	return []*DatabaseBottleneck{}, nil
}

func (p *BottleneckPatternDetector) DetectPatterns(ctx context.Context, metrics *PerformanceMetrics) ([]*Bottleneck, error) {
	return []*Bottleneck{}, nil
}

func (r *RootCauseEngine) AnalyzeRootCause(ctx context.Context, bottleneck *Bottleneck, historicalData *HistoricalBottleneckData) (*RootCauseAnalysis, error) {
	return &RootCauseAnalysis{
		PrimaryCause:        "Resource contention",
		ContributingFactors: []string{"High load", "Insufficient resources"},
		Likelihood:          0.8,
		Evidence:            []*Evidence{},
		Recommendations:     []*OptimizationRecommendation{},
	}, nil
}

func (o *OptimizationEngine) GenerateRecommendations(ctx context.Context, bottleneck *Bottleneck) ([]*OptimizationRecommendation, error) {
	return []*OptimizationRecommendation{}, nil
}

// PerformanceMetrics placeholder structure
type PerformanceMetrics struct {
	ResourceUsage *ResourceUsageMetrics `json:"resource_usage"`
	Timestamp     time.Time             `json:"timestamp"`
}