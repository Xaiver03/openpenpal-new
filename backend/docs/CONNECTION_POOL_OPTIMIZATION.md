# Connection Pool Optimization Guide

## Overview

This guide provides comprehensive instructions for optimizing PostgreSQL connection pool settings in the OpenPenPal backend system. Proper connection pool configuration is crucial for application performance and stability.

## Connection Pool Parameters

### Core Parameters

| Parameter | Description | Default | Range |
|-----------|-------------|---------|-------|
| `MaxOpenConns` | Maximum number of open connections | 100 | 1-500 |
| `MaxIdleConns` | Maximum number of idle connections | 30 | 0-MaxOpenConns |
| `ConnMaxLifetime` | Maximum connection lifetime | 15min | 1min-2hours |
| `ConnMaxIdleTime` | Maximum idle time before closing | 5min | 1min-1hour |

### Advanced Parameters (PoolConfig)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `MinIdleConns` | Minimum idle connections to maintain | 10 |
| `HealthCheckInterval` | Connection health check interval | 1min |
| `ConnectionTimeout` | Timeout for establishing connection | 15s |
| `RetryAttempts` | Number of retry attempts | 5 |
| `RetryDelay` | Delay between retries | 500ms |

## Environment-Based Presets

### Development Environment
```go
MaxOpenConns:    10
MaxIdleConns:    5
MinIdleConns:    1
ConnMaxLifetime: 1 hour
ConnMaxIdleTime: 15 minutes
```
- Low resource usage
- Suitable for single developer
- Quick connection recycling

### Testing Environment
```go
MaxOpenConns:    5
MaxIdleConns:    2
MinIdleConns:    0
ConnMaxLifetime: 30 minutes
ConnMaxIdleTime: 5 minutes
```
- Minimal connections
- Fast cleanup
- Isolated testing

### Staging Environment
```go
MaxOpenConns:    50
MaxIdleConns:    20
MinIdleConns:    5
ConnMaxLifetime: 30 minutes
ConnMaxIdleTime: 10 minutes
```
- Moderate capacity
- Production-like behavior
- Performance testing ready

### Production Environment
```go
MaxOpenConns:    100
MaxIdleConns:    30
MinIdleConns:    10
ConnMaxLifetime: 15 minutes
ConnMaxIdleTime: 5 minutes
```
- High capacity
- Optimal performance
- Connection freshness

### High Traffic Mode
```go
MaxOpenConns:    200
MaxIdleConns:    50
MinIdleConns:    20
ConnMaxLifetime: 10 minutes
ConnMaxIdleTime: 3 minutes
```
- Maximum capacity
- Rapid recycling
- Peak load handling

### Low Latency Mode
```go
MaxOpenConns:    150
MaxIdleConns:    100
MinIdleConns:    50
ConnMaxLifetime: 5 minutes
ConnMaxIdleTime: 2 minutes
```
- High idle pool
- Instant connections
- Minimal wait time

## Optimal Configuration Calculation

The system can automatically calculate optimal settings based on CPU cores:

```go
// Automatic calculation formula
MaxOpenConns = CPU_CORES * 4
MaxIdleConns = MaxOpenConns * 0.3
MinIdleConns = MaxIdleConns * 0.3
```

Example for 8-core system:
- MaxOpenConns: 32
- MaxIdleConns: 10
- MinIdleConns: 3

## Use Case Recommendations

### Web API Service
- **Characteristics**: Short queries, high concurrency, quick responses
- **Recommended**: 
  ```
  MaxOpenConns: 50
  MaxIdleConns: 20
  ConnMaxLifetime: 30 minutes
  ```

### Microservice Architecture
- **Characteristics**: Very high concurrency, distributed load
- **Recommended**:
  ```
  MaxOpenConns: 100
  MaxIdleConns: 30
  ConnMaxLifetime: 15 minutes
  ```

### Analytics System
- **Characteristics**: Long queries, low concurrency, complex operations
- **Recommended**:
  ```
  MaxOpenConns: 20
  MaxIdleConns: 10
  ConnMaxLifetime: 2 hours
  ```

### Real-time System
- **Characteristics**: Ultra-low latency, high throughput
- **Recommended**:
  ```
  MaxOpenConns: 200
  MaxIdleConns: 100
  ConnMaxLifetime: 5 minutes
  ```

## Monitoring and Tuning

### Key Metrics to Monitor

1. **Connection Usage Rate**
   ```
   Usage = ActiveConnections / MaxOpenConns
   ```
   - Healthy: < 70%
   - Warning: 70-85%
   - Critical: > 85%

2. **Idle Connection Ratio**
   ```
   IdleRatio = IdleConnections / MaxIdleConns
   ```
   - Optimal: 30-70%
   - Too low: < 20% (increase MaxIdleConns)
   - Too high: > 80% (decrease MaxIdleConns)

3. **Connection Wait Time**
   - Optimal: < 10ms
   - Acceptable: 10-100ms
   - Poor: > 100ms

4. **Connection Churn Rate**
   ```
   ChurnRate = (ClosedConnections / TotalConnections) per minute
   ```
   - Low: < 10%
   - Medium: 10-30%
   - High: > 30% (increase lifetime)

### Monitoring Tools

1. **Real-time Monitoring**
   ```bash
   ./scripts/monitor-pool.sh monitor
   ```

2. **Historical Analysis**
   ```bash
   ./scripts/monitor-pool.sh analyze
   ```

3. **Performance Testing**
   ```bash
   ./scripts/monitor-pool.sh test
   ```

4. **Export Metrics**
   ```bash
   ./scripts/monitor-pool.sh export metrics.csv
   ```

### PostgreSQL Queries

Monitor active connections:
```sql
SELECT 
    COUNT(*) FILTER (WHERE state = 'active') as active,
    COUNT(*) FILTER (WHERE state = 'idle') as idle,
    COUNT(*) as total
FROM pg_stat_activity
WHERE datname = 'openpenpal';
```

Check connection age:
```sql
SELECT 
    pid,
    usename,
    application_name,
    backend_start,
    age(now(), backend_start) as connection_age
FROM pg_stat_activity
WHERE datname = 'openpenpal'
ORDER BY backend_start;
```

## Common Issues and Solutions

### 1. Connection Pool Exhaustion
**Symptoms**: "too many connections" errors
**Solutions**:
- Increase MaxOpenConns
- Reduce ConnMaxLifetime
- Optimize slow queries
- Implement connection retry logic

### 2. Connection Leaks
**Symptoms**: Gradually increasing connection count
**Solutions**:
- Ensure proper connection closing
- Set appropriate ConnMaxIdleTime
- Monitor idle in transaction connections
- Review transaction management

### 3. High Connection Churn
**Symptoms**: Frequent connection creation/destruction
**Solutions**:
- Increase ConnMaxLifetime
- Increase MinIdleConns
- Review connection usage patterns
- Consider connection pooler (PgBouncer)

### 4. Slow Connection Acquisition
**Symptoms**: High wait times for connections
**Solutions**:
- Increase MaxIdleConns
- Increase MinIdleConns
- Reduce ConnectionTimeout
- Pre-warm connection pool

## Best Practices

1. **Start Conservative**
   - Begin with lower values
   - Gradually increase based on monitoring
   - Avoid over-provisioning

2. **Monitor Continuously**
   - Set up automated monitoring
   - Track trends over time
   - Alert on anomalies

3. **Test Changes**
   - Test configuration changes in staging
   - Use load testing tools
   - Measure impact on performance

4. **Consider Database Limits**
   - Check PostgreSQL max_connections
   - Account for other applications
   - Leave headroom for maintenance

5. **Implement Circuit Breakers**
   - Fail fast on connection issues
   - Prevent cascade failures
   - Implement graceful degradation

## Configuration Examples

### Environment Variables
```bash
# Basic configuration
export HIGH_TRAFFIC_MODE=false

# Custom pool settings (if not using presets)
export DB_MAX_OPEN_CONNS=100
export DB_MAX_IDLE_CONNS=30
export DB_CONN_MAX_LIFETIME=15m
export DB_CONN_MAX_IDLE_TIME=5m
```

### Code Configuration
```go
// Using presets
poolConfig := config.GetPoolPreset(config.PoolPresetProduction)

// Using scenario-based config
poolConfig := config.GetRecommendedConfig("web_api")

// Custom configuration
poolConfig := &config.PoolConfig{
    MaxOpenConns:    75,
    MaxIdleConns:    25,
    MinIdleConns:    10,
    ConnMaxLifetime: 20 * time.Minute,
    ConnMaxIdleTime: 7 * time.Minute,
}
```

## Performance Impact

### Connection Pool Size vs Performance

| Pool Size | Throughput | Latency | Resource Usage |
|-----------|------------|---------|----------------|
| Too Small | Low | High | Low |
| Optimal | High | Low | Moderate |
| Too Large | Moderate | Low | High |

### Lifetime Settings Impact

| Lifetime | Pros | Cons |
|----------|------|------|
| Short (5-10min) | Fresh connections, less drift | Higher churn, more overhead |
| Medium (15-30min) | Balanced | Moderate everything |
| Long (1-2hr) | Low overhead, stable | Stale connections, drift risk |

## Troubleshooting Commands

```bash
# Check current connections
psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'openpenpal';"

# Show connection states
psql -c "SELECT state, count(*) FROM pg_stat_activity WHERE datname = 'openpenpal' GROUP BY state;"

# Find long-running queries
psql -c "SELECT pid, now() - query_start as duration, query FROM pg_stat_activity WHERE state != 'idle' AND query_start < now() - interval '1 minute';"

# Kill idle connections
psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND query_start < now() - interval '10 minutes';"
```

## Conclusion

Proper connection pool configuration is critical for application performance. Use the provided presets as starting points, monitor continuously, and adjust based on actual usage patterns. Remember that optimal settings vary based on workload characteristics and infrastructure capabilities.