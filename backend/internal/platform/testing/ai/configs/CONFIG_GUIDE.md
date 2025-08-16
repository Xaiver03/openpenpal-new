# AI Testing Module Configuration Guide

This guide explains how to configure the AI Testing Module using the provided configuration files.

## 📁 Configuration Files Overview

### 1. `analyzer-config.json`
Basic configuration for the Go code analyzer component.
- **Use case**: When you only need static code analysis
- **Key settings**: Complexity thresholds, file patterns, risk levels

### 2. `generator-config.json`
Configuration for the ML-driven test case generator.
- **Use case**: Core test generation with machine learning features
- **Key settings**: ML models, genetic algorithms, generation strategies

### 3. `enhanced-config.json`
Full-featured configuration with advanced AI capabilities.
- **Use case**: Production deployments requiring advanced features
- **Key settings**: Deep analysis, intelligent mocking, performance prediction

### 4. `adapter-config.json`
Configuration for integrating AI module with the core testing engine.
- **Use case**: Integration with existing testing infrastructure
- **Key settings**: Service integrations, performance limits, monitoring

### 5. `complete-config.yaml`
Comprehensive YAML configuration with all options and environment overrides.
- **Use case**: Complex deployments with multiple environments
- **Key settings**: Everything + environment-specific configurations

## 🚀 Quick Start

### Basic Setup
```go
// Load basic analyzer configuration
config := &ai.AnalyzerConfig{
    MaxComplexity:       15,
    EnableDeepAnalysis:  true,
    AnalyzeTestFiles:    false,
}

analyzer := ai.NewGoCodeAnalyzer(config)
```

### Enhanced AI Setup
```go
// Load enhanced configuration from file
enhancedConfig := &ai.EnhancedConfig{
    GeneratorConfig: &ai.GeneratorConfig{
        MaxTestCases:         50,
        MinCoverageTarget:    0.85,
        ComplexityThreshold:  15,
        EnablePatternLearning: true,
        EnableCoverageOpt:    true,
    },
    EnableAdvancedFeatures: true,
    AnalysisDepth:         "deep",
    LearningEnabled:       true,
}

generator := ai.NewEnhancedAIGenerator(enhancedConfig)
```

### Integration with Core Engine
```go
// Core engine integration
aiConfig := &core.EnhancedAIConfig{
    ModelPath:              "/path/to/models",
    ConfidenceThreshold:    0.8,
    MaxGeneratedTests:      100,
    EnableAdvancedFeatures: true,
    AnalysisDepth:          "deep",
    LearningEnabled:        true,
}

engine.aiGenerator = core.NewEnhancedAIGeneratorAdapter(aiConfig)
```

## ⚙️ Configuration Options Explained

### Analyzer Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_complexity` | int | 15 | Maximum cyclomatic complexity before flagging as high risk |
| `enable_deep_analysis` | bool | true | Enable detailed dependency and pattern analysis |
| `analyze_test_files` | bool | false | Whether to include test files in analysis |
| `ignore_patterns` | []string | `["_test.go"]` | File patterns to exclude from analysis |
| `focus_patterns` | []string | `["*.go"]` | File patterns to focus analysis on |

### Generator Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_test_cases` | int | 50 | Maximum number of test cases to generate |
| `min_coverage_target` | float64 | 0.85 | Minimum code coverage target (0.0-1.0) |
| `complexity_threshold` | int | 15 | Complexity threshold for generating comprehensive tests |
| `enable_pattern_learning` | bool | true | Enable ML pattern recognition |
| `enable_coverage_opt` | bool | true | Enable genetic algorithm coverage optimization |
| `test_data_size` | int | 1000 | Size of generated test data sets |
| `random_seed` | int | 42 | Random seed for reproducible test generation |

### Machine Learning Models

#### Complexity Model
- **Purpose**: Predicts test complexity based on function characteristics
- **Algorithm**: Neural network (multi-layer perceptron)
- **Features**: Parameters, dependencies, cyclomatic complexity
- **Output**: Complexity score (0.0-1.0)

#### Priority Model  
- **Purpose**: Classifies test priority based on function importance
- **Algorithm**: Decision tree classifier
- **Features**: Visibility, error handling, usage frequency
- **Output**: Priority level (Low/Medium/High/Critical)

#### Pattern Recognizer
- **Purpose**: Identifies code patterns for targeted test generation
- **Algorithm**: N-gram analysis with frequency scoring
- **Features**: Function names, return types, parameter patterns
- **Output**: Pattern matches with confidence scores

#### Coverage Optimizer
- **Purpose**: Optimizes test selection for maximum coverage
- **Algorithm**: Genetic algorithm with tournament selection
- **Fitness**: Coverage score + execution efficiency
- **Output**: Optimized test case subset

### Advanced Features

#### Deep Code Analysis
```yaml
deep_code_analysis:
  enabled: true
  analyze_dependencies: true    # Trace function dependencies
  trace_data_flow: true        # Follow data flow through functions
  identify_side_effects: true  # Detect functions with side effects
  detect_race_conditions: true # Find potential race conditions
```

#### Intelligent Mocking
```yaml
intelligent_mocking:
  enabled: true
  auto_generate_mocks: true     # Automatically create mocks for interfaces
  mock_external_apis: true      # Mock external API calls
  smart_stub_generation: true   # Generate intelligent stubs
  interface_based_mocking: true # Use interface-based mocking strategy
```

#### Performance Prediction
```yaml
performance_prediction:
  enabled: true
  estimate_execution_time: true # Predict test execution time
  memory_usage_analysis: true   # Analyze memory usage patterns
  resource_optimization: true   # Optimize resource usage
  bottleneck_detection: true    # Identify performance bottlenecks
```

### Quality Metrics

| Metric | Target | Description |
|--------|--------|-------------|
| `target_success_rate` | 0.95 | Target test success rate |
| `min_confidence_score` | 0.8 | Minimum AI confidence for generated tests |
| `max_false_positive_rate` | 0.05 | Maximum acceptable false positive rate |
| `coverage_improvement_threshold` | 0.1 | Minimum coverage improvement to justify new tests |
| `pattern_recognition_accuracy` | 0.85 | Target accuracy for pattern recognition |

## 🔧 Environment-Specific Configuration

### Development Environment
```yaml
environments:
  development:
    ai_testing:
      enhanced:
        log_level: "debug"      # Verbose logging for debugging
      generator:
        max_test_cases: 20      # Fewer tests for faster iteration
      monitoring:
        enable_metrics: true    # Enable detailed metrics
```

### Testing Environment
```yaml
environments:
  testing:
    ai_testing:
      generator:
        max_test_cases: 20      # Limited test cases for CI/CD
        test_data_size: 100     # Smaller datasets for speed
      performance:
        timeout_seconds: 120    # Shorter timeouts
```

### Production Environment
```yaml
environments:
  production:
    ai_testing:
      enhanced:
        log_level: "warn"       # Minimal logging
      performance:
        max_concurrent_analyses: 10  # Higher concurrency
        memory_limit_mb: 2048        # More memory allocation
      monitoring:
        export_prometheus_metrics: true  # Production monitoring
```

## 📊 Performance Tuning

### Memory Optimization
```yaml
performance:
  memory_limit_mb: 1024        # Adjust based on available memory
  batch_size: 20               # Process files in batches
  cache_ttl_seconds: 3600      # Cache results for 1 hour
```

### CPU Optimization
```yaml
performance:
  max_concurrent_analyses: 5   # Number of parallel analyses
  cpu_limit_percent: 80        # Max CPU usage percentage
```

### I/O Optimization
```yaml
analyzer:
  ignore_patterns:             # Exclude unnecessary files
    - "vendor/**"
    - "node_modules/**"
    - ".git/**"
```

## 🔍 Monitoring and Observability

### Metrics Configuration
```yaml
monitoring:
  enable_metrics: true               # Enable metrics collection
  export_prometheus_metrics: true    # Export to Prometheus
  metrics_port: 9090                # Metrics endpoint port
  health_check_interval: 30         # Health check frequency (seconds)
```

### Available Metrics
- `ai_tests_generated_total`: Total number of tests generated
- `ai_analysis_duration_seconds`: Time spent on code analysis
- `ai_generation_success_rate`: Success rate of test generation
- `ai_pattern_recognition_accuracy`: Pattern recognition accuracy
- `ai_coverage_improvement_percent`: Coverage improvement percentage

### Logging Configuration
```yaml
enhanced:
  log_level: "info"  # debug, info, warn, error
```

## 🔒 Security Configuration

### Data Security
```yaml
security:
  sanitize_test_data: true      # Remove sensitive data from test cases
  redact_sensitive_info: true   # Redact sensitive information in logs
  secure_model_storage: true    # Encrypt ML model files
  audit_ai_decisions: true      # Log AI decision making for audit
```

## 🧪 Testing the Configuration

### Validate Configuration
```bash
# Test analyzer configuration
go run cmd/testing-ai-demo/main.go --config-file=configs/analyzer-config.json --validate-only

# Test enhanced configuration  
go run cmd/testing-ai-demo/main.go --config-file=configs/enhanced-config.json --dry-run

# Test YAML configuration
go run cmd/testing-ai-demo/main.go --config-file=configs/complete-config.yaml --environment=development
```

### Configuration Validation Script
```bash
#!/bin/bash
# Validate all configuration files

echo "Validating AI module configurations..."

for config in configs/*.json configs/*.yaml; do
    echo "Validating $config..."
    go run cmd/validate-config/main.go --config="$config"
done

echo "Configuration validation complete!"
```

## 📚 Best Practices

### 1. Start Simple
Begin with basic configurations and gradually enable advanced features as needed.

### 2. Environment-Specific Tuning
Use different configurations for development, testing, and production environments.

### 3. Monitor Performance
Keep an eye on memory usage and CPU consumption, especially with deep analysis enabled.

### 4. Regular Model Updates
Update ML models regularly based on accumulated learning data.

### 5. Security First
Always enable security features in production environments.

## 🔧 Troubleshooting

### Common Issues

#### High Memory Usage
```yaml
# Solution: Reduce batch size and enable garbage collection
performance:
  batch_size: 10
  memory_limit_mb: 512
```

#### Slow Analysis
```yaml
# Solution: Reduce analysis depth or disable expensive features
enhanced:
  analysis_depth: "medium"
advanced_analysis:
  trace_data_flow: false
```

#### Low Test Quality
```yaml
# Solution: Increase confidence thresholds and enable learning
generator:
  enable_pattern_learning: true
quality_targets:
  min_confidence_score: 0.9
```

## 📖 Additional Resources

- [AI Module README](./README.md)
- [Core Testing Engine Documentation](../core/README.md)
- [Performance Benchmarking Guide](./BENCHMARKS.md)
- [ML Model Training Guide](./ML_TRAINING.md)

---

**Note**: This configuration guide covers the AI Testing Module (Phase 3.2) of the SOTA Testing Infrastructure. For complete system configuration, refer to the main testing engine documentation.