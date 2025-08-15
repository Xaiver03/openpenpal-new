package monitoring

import (
	"openpenpal-backend/internal/resilience"
	"sync"
	"time"
)

// CircuitBreakerMetricsIntegration provides automatic metrics collection for circuit breakers
type CircuitBreakerMetricsIntegration struct {
	collector *MetricsCollector
	manager   *resilience.CircuitBreakerManager
	ticker    *time.Ticker
	done      chan bool
	mutex     sync.RWMutex
}

// NewCircuitBreakerMetricsIntegration creates a new circuit breaker metrics integration
func NewCircuitBreakerMetricsIntegration(collector *MetricsCollector, manager *resilience.CircuitBreakerManager) *CircuitBreakerMetricsIntegration {
	return &CircuitBreakerMetricsIntegration{
		collector: collector,
		manager:   manager,
		done:      make(chan bool),
	}
}

// Start begins collecting circuit breaker metrics at regular intervals
func (cbmi *CircuitBreakerMetricsIntegration) Start(interval time.Duration) {
	cbmi.mutex.Lock()
	defer cbmi.mutex.Unlock()

	if cbmi.ticker != nil {
		return // Already started
	}

	cbmi.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-cbmi.ticker.C:
				cbmi.collectMetrics()
			case <-cbmi.done:
				return
			}
		}
	}()
}

// Stop stops collecting circuit breaker metrics
func (cbmi *CircuitBreakerMetricsIntegration) Stop() {
	cbmi.mutex.Lock()
	defer cbmi.mutex.Unlock()

	if cbmi.ticker != nil {
		cbmi.ticker.Stop()
		cbmi.ticker = nil
		close(cbmi.done)
		cbmi.done = make(chan bool)
	}
}

// collectMetrics collects and exports all circuit breaker metrics
func (cbmi *CircuitBreakerMetricsIntegration) collectMetrics() {
	breakers := cbmi.manager.GetAllBreakers()

	for name, cb := range breakers {
		// Get current state and counts
		state := cb.GetState()
		counts := cb.GetCounts()

		// Set state metric (0=closed, 1=open, 2=half-open)
		cbmi.collector.SetCircuitBreakerState(name, int(state))

		// Record request metrics
		cbmi.collector.IncrementCircuitBreakerRequests(name, "total")

		// Record failure metrics based on counts
		if counts.TotalFailures > 0 {
			// Calculate failure types
			recentFailures := counts.ConsecutiveFailures
			if recentFailures > 0 {
				cbmi.collector.IncrementCircuitBreakerFailures(name, "consecutive")
			}

			failureRate := counts.FailureRate()
			if failureRate > 50 {
				cbmi.collector.IncrementCircuitBreakerFailures(name, "high_failure_rate")
			}
		}

		// Log state changes for monitoring
		cbmi.logStateForMonitoring(name, state, counts)
	}
}

// logStateForMonitoring creates monitoring-friendly logs for circuit breaker states
func (cbmi *CircuitBreakerMetricsIntegration) logStateForMonitoring(name string, state resilience.CircuitBreakerState, counts resilience.Counts) {
	// Create structured monitoring data
	monitoringData := map[string]interface{}{
		"circuit_breaker_name":        name,
		"state":                      state.String(),
		"total_requests":             counts.Requests,
		"total_successes":            counts.TotalSuccesses,
		"total_failures":             counts.TotalFailures,
		"consecutive_successes":      counts.ConsecutiveSuccesses,
		"consecutive_failures":       counts.ConsecutiveFailures,
		"success_rate":              counts.SuccessRate(),
		"failure_rate":              counts.FailureRate(),
		"last_success_time":         counts.LastSuccessTime,
		"last_failure_time":         counts.LastFailureTime,
		"timestamp":                 time.Now().UTC(),
	}

	// In production, this would be sent to your logging/monitoring system
	// For now, we'll store it for potential alerting
	cbmi.checkForAlerts(name, state, counts, monitoringData)
}

// checkForAlerts checks if any alerts should be triggered based on circuit breaker metrics
func (cbmi *CircuitBreakerMetricsIntegration) checkForAlerts(name string, state resilience.CircuitBreakerState, counts resilience.Counts, data map[string]interface{}) {
	alerts := make([]string, 0)

	// Alert if circuit breaker is open
	if state == resilience.StateOpen {
		alerts = append(alerts, "circuit_breaker_open")
		cbmi.collector.IncrementCircuitBreakerFailures(name, "alert_circuit_open")
	}

	// Alert if failure rate is very high
	if counts.FailureRate() > 80 && counts.Requests > 10 {
		alerts = append(alerts, "high_failure_rate")
		cbmi.collector.IncrementCircuitBreakerFailures(name, "alert_high_failure_rate")
	}

	// Alert if too many consecutive failures
	if counts.ConsecutiveFailures > 10 {
		alerts = append(alerts, "excessive_consecutive_failures")
		cbmi.collector.IncrementCircuitBreakerFailures(name, "alert_consecutive_failures")
	}

	// Alert if circuit breaker has been half-open for too long
	if state == resilience.StateHalfOpen {
		// This would require tracking how long it's been half-open
		// For now, just increment a metric
		cbmi.collector.IncrementCircuitBreakerRequests(name, "half_open_duration")
	}

	// If there are alerts, they would be sent to your alerting system here
	if len(alerts) > 0 {
		cbmi.handleAlerts(name, alerts, data)
	}
}

// handleAlerts processes alerts that need to be sent
func (cbmi *CircuitBreakerMetricsIntegration) handleAlerts(serviceName string, alerts []string, data map[string]interface{}) {
	// In a production environment, this would:
	// 1. Send alerts to PagerDuty, Slack, etc.
	// 2. Create incidents in your incident management system
	// 3. Update status pages
	// 4. Trigger automatic remediation if configured

	// For now, we'll just log the alerts
	for _, alert := range alerts {
		// This would integrate with your alerting system
		// Example: Send to Slack, PagerDuty, etc.
		cbmi.processAlert(serviceName, alert, data)
	}
}

// processAlert processes a single alert
func (cbmi *CircuitBreakerMetricsIntegration) processAlert(serviceName, alertType string, data map[string]interface{}) {
	alertData := map[string]interface{}{
		"service":     serviceName,
		"alert_type":  alertType,
		"severity":    cbmi.getAlertSeverity(alertType),
		"timestamp":   time.Now().UTC(),
		"data":        data,
	}

	// In production, this would send to your alerting infrastructure
	// For example:
	// - Send webhook to Slack
	// - Create PagerDuty incident
	// - Send email notification
	// - Update monitoring dashboard

	// Record the alert in metrics for tracking
	cbmi.collector.IncrementCircuitBreakerFailures(serviceName, "alert_"+alertType)

	// Example integration points:
	cbmi.sendToMonitoringSystem(alertData)
}

// getAlertSeverity determines the severity of an alert
func (cbmi *CircuitBreakerMetricsIntegration) getAlertSeverity(alertType string) string {
	severityMap := map[string]string{
		"circuit_breaker_open":           "critical",
		"high_failure_rate":             "warning",
		"excessive_consecutive_failures": "warning",
		"half_open_duration":            "info",
	}

	if severity, exists := severityMap[alertType]; exists {
		return severity
	}
	return "info"
}

// sendToMonitoringSystem sends alert data to external monitoring systems
func (cbmi *CircuitBreakerMetricsIntegration) sendToMonitoringSystem(alertData map[string]interface{}) {
	// Integration examples:

	// 1. Datadog Integration
	// cbmi.sendToDatadog(alertData)

	// 2. New Relic Integration
	// cbmi.sendToNewRelic(alertData)

	// 3. Custom webhook integration
	// cbmi.sendWebhook(alertData)

	// 4. Internal monitoring system
	// cbmi.sendToInternalMonitoring(alertData)
}

// Helper methods for external integrations

// sendToDatadog would send metrics to Datadog
func (cbmi *CircuitBreakerMetricsIntegration) sendToDatadog(alertData map[string]interface{}) {
	// Example Datadog integration
	// This would use the Datadog API to send custom metrics and events
}

// sendToNewRelic would send metrics to New Relic
func (cbmi *CircuitBreakerMetricsIntegration) sendToNewRelic(alertData map[string]interface{}) {
	// Example New Relic integration
	// This would use the New Relic API to send custom events and metrics
}

// sendWebhook would send alert data to a custom webhook
func (cbmi *CircuitBreakerMetricsIntegration) sendWebhook(alertData map[string]interface{}) {
	// Example webhook integration
	// This would POST the alert data to a configured webhook URL
}

// GetIntegrationStatus returns the current status of the integration
func (cbmi *CircuitBreakerMetricsIntegration) GetIntegrationStatus() map[string]interface{} {
	cbmi.mutex.RLock()
	defer cbmi.mutex.RUnlock()

	isRunning := cbmi.ticker != nil

	breakers := cbmi.manager.GetAllBreakers()
	breakerStates := make(map[string]string)
	for name, cb := range breakers {
		breakerStates[name] = cb.GetState().String()
	}

	return map[string]interface{}{
		"running":           isRunning,
		"monitored_breakers": len(breakers),
		"breaker_states":    breakerStates,
		"last_collection":   time.Now().UTC(),
	}
}

// Default integration instance
var DefaultCircuitBreakerMetricsIntegration = NewCircuitBreakerMetricsIntegration(
	DefaultMetricsCollector,
	resilience.DefaultCircuitBreakerManager,
)