// Package servicemesh provides anomaly detection functionality
package servicemesh

import (
	"sync"
	"time"
)

// AnomalyDetector detects performance and behavioral anomalies
type AnomalyDetector struct {
	baselines   map[string]*Baseline
	mu          sync.RWMutex
	sensitivity float64
}

// Baseline represents the normal behavior baseline for a service
type Baseline struct {
	ServiceID            string                 `json:"service_id"`
	AverageResponseTime  float64                `json:"average_response_time"`
	AverageErrorRate     float64                `json:"average_error_rate"`
	RequestRateBaseline  float64                `json:"request_rate_baseline"`
	CPUBaseline          float64                `json:"cpu_baseline"`
	MemoryBaseline       float64                `json:"memory_baseline"`
	LastUpdated          time.Time              `json:"last_updated"`
	SampleCount          int                    `json:"sample_count"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// AnomalyResult represents the result of anomaly detection
type AnomalyResult struct {
	ServiceID     string                 `json:"service_id"`
	AnomalyType   string                 `json:"anomaly_type"`
	Severity      string                 `json:"severity"`
	Score         float64                `json:"score"`
	Description   string                 `json:"description"`
	Detected      bool                   `json:"detected"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata"`
	Baseline      *Baseline              `json:"baseline"`
	CurrentValue  float64                `json:"current_value"`
	ExpectedValue float64                `json:"expected_value"`
	Threshold     float64                `json:"threshold"`
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(sensitivity float64) *AnomalyDetector {
	return &AnomalyDetector{
		baselines:   make(map[string]*Baseline),
		sensitivity: sensitivity,
	}
}

// DetectAnomalies detects anomalies in service metrics
func (ad *AnomalyDetector) DetectAnomalies(serviceID string, metrics *ServiceMetrics) []*AnomalyResult {
	ad.mu.RLock()
	baseline, exists := ad.baselines[serviceID]
	ad.mu.RUnlock()

	if !exists {
		// Initialize baseline for new service
		ad.initializeBaseline(serviceID, metrics)
		return nil
	}

	var results []*AnomalyResult

	// Detect response time anomalies
	if rtAnomaly := ad.detectResponseTimeAnomaly(serviceID, metrics.AverageLatency, baseline); rtAnomaly != nil {
		results = append(results, rtAnomaly)
	}

	// Detect error rate anomalies
	if errAnomaly := ad.detectErrorRateAnomaly(serviceID, metrics.ErrorRate, baseline); errAnomaly != nil {
		results = append(results, errAnomaly)
	}

	// Detect CPU anomalies
	if cpuAnomaly := ad.detectCPUAnomaly(serviceID, metrics.CPUUsage, baseline); cpuAnomaly != nil {
		results = append(results, cpuAnomaly)
	}

	// Detect memory anomalies
	if memAnomaly := ad.detectMemoryAnomaly(serviceID, metrics.MemoryUsage, baseline); memAnomaly != nil {
		results = append(results, memAnomaly)
	}

	// Update baseline with new data
	ad.updateBaseline(serviceID, metrics)

	return results
}

// UpdateBaseline updates the baseline for a service
func (ad *AnomalyDetector) UpdateBaseline(serviceID string, metrics *ServiceMetrics) {
	ad.updateBaseline(serviceID, metrics)
}

// CalculateThreshold calculates the threshold for anomaly detection
func (ad *AnomalyDetector) CalculateThreshold(baseline float64, sensitivity float64) float64 {
	// Simple threshold calculation: baseline * (1 + sensitivity)
	return baseline * (1.0 + sensitivity)
}

// GetBaseline returns the baseline for a service
func (ad *AnomalyDetector) GetBaseline(serviceID string) *Baseline {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	
	if baseline, exists := ad.baselines[serviceID]; exists {
		return baseline
	}
	return nil
}

// initializeBaseline initializes the baseline for a new service
func (ad *AnomalyDetector) initializeBaseline(serviceID string, metrics *ServiceMetrics) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	baseline := &Baseline{
		ServiceID:           serviceID,
		AverageResponseTime: metrics.AverageLatency,
		AverageErrorRate:    metrics.ErrorRate,
		RequestRateBaseline: metrics.RequestsPerSecond,
		CPUBaseline:         metrics.CPUUsage,
		MemoryBaseline:      metrics.MemoryUsage,
		LastUpdated:         time.Now(),
		SampleCount:         1,
		Metadata:            make(map[string]interface{}),
	}

	ad.baselines[serviceID] = baseline
}

// updateBaseline updates the baseline with new metrics using exponential moving average
func (ad *AnomalyDetector) updateBaseline(serviceID string, metrics *ServiceMetrics) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	baseline, exists := ad.baselines[serviceID]
	if !exists {
		return
	}

	// Use exponential moving average with alpha = 0.1 for smoothing
	alpha := 0.1
	baseline.AverageResponseTime = (1-alpha)*baseline.AverageResponseTime + alpha*metrics.AverageLatency
	baseline.AverageErrorRate = (1-alpha)*baseline.AverageErrorRate + alpha*metrics.ErrorRate
	baseline.RequestRateBaseline = (1-alpha)*baseline.RequestRateBaseline + alpha*metrics.RequestsPerSecond
	baseline.CPUBaseline = (1-alpha)*baseline.CPUBaseline + alpha*metrics.CPUUsage
	baseline.MemoryBaseline = (1-alpha)*baseline.MemoryBaseline + alpha*metrics.MemoryUsage
	baseline.LastUpdated = time.Now()
	baseline.SampleCount++
}

// detectResponseTimeAnomaly detects response time anomalies
func (ad *AnomalyDetector) detectResponseTimeAnomaly(serviceID string, currentRT float64, baseline *Baseline) *AnomalyResult {
	threshold := ad.CalculateThreshold(baseline.AverageResponseTime, ad.sensitivity)
	
	if currentRT > threshold {
		severity := "medium"
		if currentRT > threshold*2 {
			severity = "high"
		}

		return &AnomalyResult{
			ServiceID:     serviceID,
			AnomalyType:   "response_time",
			Severity:      severity,
			Score:         currentRT / baseline.AverageResponseTime,
			Description:   "Response time significantly higher than baseline",
			Detected:      true,
			Timestamp:     time.Now(),
			Baseline:      baseline,
			CurrentValue:  currentRT,
			ExpectedValue: baseline.AverageResponseTime,
			Threshold:     threshold,
			Metadata: map[string]interface{}{
				"baseline_rt": baseline.AverageResponseTime,
				"current_rt":  currentRT,
				"deviation":   currentRT - baseline.AverageResponseTime,
			},
		}
	}

	return nil
}

// detectErrorRateAnomaly detects error rate anomalies
func (ad *AnomalyDetector) detectErrorRateAnomaly(serviceID string, currentER float64, baseline *Baseline) *AnomalyResult {
	threshold := ad.CalculateThreshold(baseline.AverageErrorRate, ad.sensitivity*2) // More sensitive for errors
	
	if currentER > threshold {
		severity := "high" // Error rate anomalies are always high severity
		if currentER > 0.1 { // 10% error rate is critical
			severity = "critical"
		}

		return &AnomalyResult{
			ServiceID:     serviceID,
			AnomalyType:   "error_rate",
			Severity:      severity,
			Score:         currentER / baseline.AverageErrorRate,
			Description:   "Error rate significantly higher than baseline",
			Detected:      true,
			Timestamp:     time.Now(),
			Baseline:      baseline,
			CurrentValue:  currentER,
			ExpectedValue: baseline.AverageErrorRate,
			Threshold:     threshold,
			Metadata: map[string]interface{}{
				"baseline_er": baseline.AverageErrorRate,
				"current_er":  currentER,
				"deviation":   currentER - baseline.AverageErrorRate,
			},
		}
	}

	return nil
}

// detectCPUAnomaly detects CPU usage anomalies
func (ad *AnomalyDetector) detectCPUAnomaly(serviceID string, currentCPU float64, baseline *Baseline) *AnomalyResult {
	threshold := ad.CalculateThreshold(baseline.CPUBaseline, ad.sensitivity)
	
	if currentCPU > threshold {
		severity := "medium"
		if currentCPU > 80.0 { // 80% CPU is high
			severity = "high"
		}

		return &AnomalyResult{
			ServiceID:     serviceID,
			AnomalyType:   "cpu_usage",
			Severity:      severity,
			Score:         currentCPU / baseline.CPUBaseline,
			Description:   "CPU usage significantly higher than baseline",
			Detected:      true,
			Timestamp:     time.Now(),
			Baseline:      baseline,
			CurrentValue:  currentCPU,
			ExpectedValue: baseline.CPUBaseline,
			Threshold:     threshold,
			Metadata: map[string]interface{}{
				"baseline_cpu": baseline.CPUBaseline,
				"current_cpu":  currentCPU,
				"deviation":    currentCPU - baseline.CPUBaseline,
			},
		}
	}

	return nil
}

// detectMemoryAnomaly detects memory usage anomalies
func (ad *AnomalyDetector) detectMemoryAnomaly(serviceID string, currentMem float64, baseline *Baseline) *AnomalyResult {
	threshold := ad.CalculateThreshold(baseline.MemoryBaseline, ad.sensitivity)
	
	if currentMem > threshold {
		severity := "medium"
		if currentMem > 85.0 { // 85% memory is high
			severity = "high"
		}

		return &AnomalyResult{
			ServiceID:     serviceID,
			AnomalyType:   "memory_usage",
			Severity:      severity,
			Score:         currentMem / baseline.MemoryBaseline,
			Description:   "Memory usage significantly higher than baseline",
			Detected:      true,
			Timestamp:     time.Now(),
			Baseline:      baseline,
			CurrentValue:  currentMem,
			ExpectedValue: baseline.MemoryBaseline,
			Threshold:     threshold,
			Metadata: map[string]interface{}{
				"baseline_mem": baseline.MemoryBaseline,
				"current_mem":  currentMem,
				"deviation":    currentMem - baseline.MemoryBaseline,
			},
		}
	}

	return nil
}

// GetAllBaselines returns all baselines
func (ad *AnomalyDetector) GetAllBaselines() map[string]*Baseline {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	result := make(map[string]*Baseline)
	for k, v := range ad.baselines {
		result[k] = v
	}
	return result
}

// SetSensitivity sets the sensitivity level for anomaly detection
func (ad *AnomalyDetector) SetSensitivity(sensitivity float64) {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.sensitivity = sensitivity
}

// GetSensitivity returns the current sensitivity level
func (ad *AnomalyDetector) GetSensitivity() float64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()
	return ad.sensitivity
}