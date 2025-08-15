# Advanced Load Balancing Integration Guide

## Overview

This document provides a comprehensive guide for integrating the advanced load balancing system with the OpenPenPal API Gateway. The implementation includes sophisticated algorithms, circuit breaker integration, metrics collection, and performance monitoring.

## Architecture Components

### 1. Load Balancing Algorithms

#### Available Algorithms:
- **Round Robin**: Simple round-robin distribution
- **Weighted Round Robin**: Distribution based on instance weights
- **Least Connections**: Routes to instance with fewest active connections
- **Least Response Time**: Routes to fastest responding instance
- **Health-Aware**: Weighted selection based on comprehensive health scores
- **Consistent Hash**: Session affinity using consistent hashing
- **Adaptive**: Intelligent algorithm adapting to multiple performance metrics

#### Algorithm Selection:
```go
// Service-specific algorithm configuration
serviceAlgorithms := map[string]string{
    "main-backend":     "adaptive",
    "write-service":    "least_response_time", 
    "courier-service":  "health_aware",
    "admin-service":    "weighted_round_robin",
    "ocr-service":      "least_connections",
}
```

### 2. Circuit Breaker Integration

The load balancer integrates with the OpenPenPal backend circuit breaker system:

```go
// Circuit breaker integration
cbIntegration := loadbalancer.NewOpenPenPalCircuitBreakerIntegration(
    "http://localhost:8080", // OpenPenPal backend URL
    logger,
)
lbManager.SetCircuitBreakerIntegration(cbIntegration)
```

**Features:**
- Real-time circuit breaker status monitoring
- Automatic failure/success event reporting
- Service availability checks before routing
- Integration with OpenPenPal monitoring endpoints

### 3. Metrics Collection

Comprehensive metrics integration with OpenPenPal backend:

```go
// Metrics integration
metricsCollector := loadbalancer.NewOpenPenPalMetricsCollector(
    "http://localhost:8080", // OpenPenPal backend URL
    logger,
)
lbManager.SetMetricsCollector(metricsCollector)
```

**Collected Metrics:**
- Request count and success rate
- Response time distributions
- Instance performance scores
- Active connection counts
- Load balancing algorithm effectiveness
- Session affinity statistics

### 4. Session Affinity

Advanced session management for stateful applications:

```go
config := &loadbalancer.LoadBalancerConfig{
    SessionAffinityEnabled: true,
    SessionAffinityTTL:     30 * time.Minute,
}
```

**Features:**
- JWT token-based session identification
- Cookie-based session tracking
- Automatic session cleanup
- Cross-service session consistency

### 5. Gradual Recovery

Intelligent recovery system for unhealthy instances:

```go
config := &loadbalancer.LoadBalancerConfig{
    GradualRecoveryEnabled: true,
    RecoveryThreshold:      0.8,
    RecoveryStepSize:       0.1,
    RecoveryCheckInterval:  30 * time.Second,
}
```

**Process:**
1. Unhealthy instances are gradually reintroduced
2. Traffic weight starts at 10% and increases incrementally
3. Continuous health monitoring during recovery
4. Automatic fallback if recovery fails

## Implementation Setup

### 1. Gateway Main Setup

```go
// main.go
func main() {
    // Initialize service discovery
    serviceDiscovery := discovery.NewServiceDiscovery(config)
    
    // Create load balancer configuration
    lbConfig := loadbalancer.CreateComprehensiveConfig()
    
    // Initialize load balancer manager
    lbManager := loadbalancer.NewLoadBalancerManager(lbConfig, serviceDiscovery, logger)
    
    // Set up integrations
    setupIntegrations(lbManager)
    
    // Create enhanced proxy manager
    proxyConfig := proxy.DefaultEnhancedProxyConfig()
    proxyManager := proxy.NewEnhancedProxyManager(
        serviceDiscovery,
        lbManager,
        logger,
        proxyConfig,
    )
    
    // Start load balancer
    lbManager.Start()
    
    // Setup routes
    setupRoutes(router, proxyManager, lbManager)
}

func setupIntegrations(lbManager *loadbalancer.LoadBalancerManager) {
    // Circuit breaker integration
    cbIntegration := loadbalancer.NewOpenPenPalCircuitBreakerIntegration(
        "http://localhost:8080", logger)
    lbManager.SetCircuitBreakerIntegration(cbIntegration)
    
    // Metrics integration
    metricsCollector := loadbalancer.NewOpenPenPalMetricsCollector(
        "http://localhost:8080", logger)
    lbManager.SetMetricsCollector(metricsCollector)
}
```

### 2. Route Configuration

```go
func setupRoutes(router *gin.Engine, proxyManager *proxy.EnhancedProxyManager, lbManager *loadbalancer.LoadBalancerManager) {
    // API routes with enhanced proxy
    api := router.Group("/api/v1")
    {
        // Main backend routes
        api.Any("/auth/*path", proxyManager.EnhancedProxyHandler("main-backend"))
        api.Any("/users/*path", proxyManager.EnhancedProxyHandler("main-backend"))
        
        // Write service routes
        api.Any("/letters/write/*path", proxyManager.EnhancedProxyHandler("write-service"))
        
        // Courier service routes
        api.Any("/courier/*path", proxyManager.EnhancedProxyHandler("courier-service"))
        
        // Admin service routes
        api.Any("/admin/*path", proxyManager.EnhancedProxyHandler("admin-service"))
        
        // OCR service routes
        api.Any("/ocr/*path", proxyManager.EnhancedProxyHandler("ocr-service"))
    }
    
    // Load balancer management endpoints
    lbHandler := handlers.NewLoadBalancerHandler(lbManager, logger)
    lbHandler.RegisterLoadBalancerRoutes(router)
}
```

### 3. Environment Configuration

```bash
# Gateway Configuration
PORT=8000
ENVIRONMENT=production
LOG_LEVEL=info

# Load Balancer Configuration
LB_DEFAULT_ALGORITHM=adaptive
LB_HEALTH_CHECK_ENABLED=true
LB_METRICS_ENABLED=true
LB_CIRCUIT_BREAKER_ENABLED=true
LB_SESSION_AFFINITY_ENABLED=true
LB_SESSION_AFFINITY_TTL=30m

# Service Discovery Configuration
MAIN_BACKEND_HOSTS=http://localhost:8080,http://localhost:8081
WRITE_SERVICE_HOSTS=http://localhost:8001,http://localhost:8011
COURIER_SERVICE_HOSTS=http://localhost:8002,http://localhost:8012
ADMIN_SERVICE_HOSTS=http://localhost:8003
OCR_SERVICE_HOSTS=http://localhost:8004

# Service Weights
MAIN_BACKEND_WEIGHT=10
WRITE_SERVICE_WEIGHT=10
COURIER_SERVICE_WEIGHT=8
ADMIN_SERVICE_WEIGHT=5
OCR_SERVICE_WEIGHT=6

# Circuit Breaker Integration
CB_BACKEND_URL=http://localhost:8080
CB_CHECK_INTERVAL=10s

# Metrics Integration
METRICS_BACKEND_URL=http://localhost:8080
METRICS_COLLECTION_INTERVAL=30s
```

## API Endpoints

### Load Balancer Management

```bash
# Get overall statistics
GET /api/v1/loadbalancer/stats

# Get service-specific statistics
GET /api/v1/loadbalancer/stats/{service}

# Get all instances
GET /api/v1/loadbalancer/instances

# Get service instances
GET /api/v1/loadbalancer/instances/{service}

# Get instance details
GET /api/v1/loadbalancer/instances/{service}/{instance}

# Drain an instance
POST /api/v1/loadbalancer/instances/{service}/{instance}/drain

# Enable/disable instances
POST /api/v1/loadbalancer/instances/{service}/{instance}/enable
POST /api/v1/loadbalancer/instances/{service}/{instance}/disable

# Algorithm management
GET /api/v1/loadbalancer/algorithms
GET /api/v1/loadbalancer/algorithms/{service}
PUT /api/v1/loadbalancer/algorithms/{service}

# Configuration management
GET /api/v1/loadbalancer/config
PUT /api/v1/loadbalancer/config

# Health and performance
GET /api/v1/loadbalancer/health
GET /api/v1/loadbalancer/performance

# Session affinity
GET /api/v1/loadbalancer/sessions
DELETE /api/v1/loadbalancer/sessions/{session}
DELETE /api/v1/loadbalancer/sessions

# Recovery management
GET /api/v1/loadbalancer/recovery
POST /api/v1/loadbalancer/recovery/{service}/{instance}/force
```

### Example API Calls

```bash
# Change algorithm for a service
curl -X PUT http://localhost:8000/api/v1/loadbalancer/algorithms/main-backend \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "least_response_time"}'

# Get performance metrics
curl "http://localhost:8000/api/v1/loadbalancer/performance?window=1h&service=main-backend"

# Drain an instance
curl -X POST http://localhost:8000/api/v1/loadbalancer/instances/main-backend/http://localhost:8081/drain

# Get health status
curl http://localhost:8000/api/v1/loadbalancer/health
```

## Monitoring and Observability

### 1. Prometheus Metrics

The load balancer exports comprehensive metrics to the OpenPenPal monitoring system:

```prometheus
# Request metrics
gateway_load_balancer_request_total{service, algorithm, instance, success}
gateway_load_balancer_response_time_seconds{service, algorithm, instance}

# Instance metrics
gateway_instance_active_connections{service, instance}
gateway_instance_total_requests{service, instance}
gateway_instance_success_rate{service, instance}
gateway_instance_average_response_time{service, instance}
gateway_instance_health_score{service, instance}
```

### 2. Health Checks

Built-in health monitoring:

```bash
# Gateway health
GET /health

# Load balancer health
GET /api/v1/loadbalancer/health

# Service instance health
GET /api/v1/loadbalancer/instances/{service}/{instance}
```

### 3. Logging

Structured logging with detailed request information:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Enhanced proxy request completed",
  "service": "main-backend",
  "instance": "http://localhost:8080",
  "method": "POST",
  "path": "/api/v1/letters",
  "status": 200,
  "duration": "120ms",
  "client_ip": "192.168.1.100",
  "success": true,
  "instance_score": 0.95,
  "instance_connections": 5,
  "instance_success_rate": 98.5,
  "trace_id": "gw-1705315800-abc123"
}
```

## Performance Features

### 1. Intelligent Routing

- **Performance-based**: Routes based on response time and throughput
- **Health-aware**: Considers instance health scores
- **Load-aware**: Avoids overloaded instances
- **Adaptive**: Learns from historical performance

### 2. Advanced Retry Logic

```go
config := &proxy.EnhancedProxyConfig{
    RetryEnabled:     true,
    MaxRetries:       2,
    RetryDelay:       100 * time.Millisecond,
    RetryBackoff:     2.0,
}
```

### 3. Request Timeout Management

```go
config := &proxy.EnhancedProxyConfig{
    RequestTimeout:   30 * time.Second,
    ResponseTimeout:  30 * time.Second,
}
```

## Best Practices

### 1. Algorithm Selection

- **High-traffic services**: Use `adaptive` or `least_response_time`
- **Stateful services**: Use `consistent_hash` for session affinity
- **Resource-intensive services**: Use `least_connections`
- **Simple services**: Use `weighted_round_robin`

### 2. Configuration Tuning

```go
// For high-performance services
config := &loadbalancer.LoadBalancerConfig{
    DefaultAlgorithm:    "adaptive",
    PerformanceMonitoringWindow: 30 * time.Second,
    RecoveryCheckInterval: 15 * time.Second,
}

// For stable services
config := &loadbalancer.LoadBalancerConfig{
    DefaultAlgorithm:    "weighted_round_robin",
    PerformanceMonitoringWindow: 5 * time.Minute,
    RecoveryCheckInterval: 60 * time.Second,
}
```

### 3. Monitoring Setup

1. **Enable comprehensive metrics collection**
2. **Set up alerting for high error rates**
3. **Monitor instance health scores**
4. **Track session affinity effectiveness**
5. **Monitor circuit breaker state changes**

## Troubleshooting

### Common Issues

1. **Uneven load distribution**
   - Check instance weights
   - Verify algorithm selection
   - Monitor health scores

2. **High response times**
   - Check instance performance metrics
   - Consider algorithm change
   - Monitor circuit breaker status

3. **Session affinity issues**
   - Verify session ID extraction
   - Check TTL configuration
   - Monitor session statistics

### Debug Endpoints

```bash
# Get debug information
GET /api/v1/admin/loadbalancer/debug

# Reset load balancer state
POST /api/v1/admin/loadbalancer/reset

# Reload configuration
POST /api/v1/admin/loadbalancer/reload
```

## Integration with OpenPenPal Backend

The advanced load balancer seamlessly integrates with the OpenPenPal monitoring and circuit breaker systems:

1. **Circuit Breaker Events**: Automatic reporting to backend circuit breaker system
2. **Metrics Collection**: Integration with Prometheus metrics collection
3. **Health Monitoring**: Coordination with backend health check systems
4. **Configuration Management**: Dynamic configuration updates via API

This provides a unified monitoring and management experience across the entire OpenPenPal platform.