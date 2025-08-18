package config

import (
	"runtime"
	"testing"
	"time"
)

func TestPoolPresets(t *testing.T) {
	tests := []struct {
		name   string
		preset PoolPreset
		want   struct {
			maxOpen         int
			maxIdle         int
			minIdle         int
			maxLifetime     time.Duration
			maxIdleTime     time.Duration
			environment     string
		}
	}{
		{
			name:   "Development preset",
			preset: PoolPresetDevelopment,
			want: struct {
				maxOpen     int
				maxIdle     int
				minIdle     int
				maxLifetime time.Duration
				maxIdleTime time.Duration
				environment string
			}{
				maxOpen:     10,
				maxIdle:     5,
				minIdle:     1,
				maxLifetime: time.Hour,
				maxIdleTime: 15 * time.Minute,
				environment: "development",
			},
		},
		{
			name:   "Production preset",
			preset: PoolPresetProduction,
			want: struct {
				maxOpen     int
				maxIdle     int
				minIdle     int
				maxLifetime time.Duration
				maxIdleTime time.Duration
				environment string
			}{
				maxOpen:     100,
				maxIdle:     30,
				minIdle:     10,
				maxLifetime: 15 * time.Minute,
				maxIdleTime: 5 * time.Minute,
				environment: "production",
			},
		},
		{
			name:   "High traffic preset",
			preset: PoolPresetHighTraffic,
			want: struct {
				maxOpen     int
				maxIdle     int
				minIdle     int
				maxLifetime time.Duration
				maxIdleTime time.Duration
				environment string
			}{
				maxOpen:     200,
				maxIdle:     50,
				minIdle:     20,
				maxLifetime: 10 * time.Minute,
				maxIdleTime: 3 * time.Minute,
				environment: "high_traffic",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPoolPreset(tt.preset)
			
			if got.MaxOpenConns != tt.want.maxOpen {
				t.Errorf("MaxOpenConns = %v, want %v", got.MaxOpenConns, tt.want.maxOpen)
			}
			if got.MaxIdleConns != tt.want.maxIdle {
				t.Errorf("MaxIdleConns = %v, want %v", got.MaxIdleConns, tt.want.maxIdle)
			}
			if got.MinIdleConns != tt.want.minIdle {
				t.Errorf("MinIdleConns = %v, want %v", got.MinIdleConns, tt.want.minIdle)
			}
			if got.ConnMaxLifetime != tt.want.maxLifetime {
				t.Errorf("ConnMaxLifetime = %v, want %v", got.ConnMaxLifetime, tt.want.maxLifetime)
			}
			if got.ConnMaxIdleTime != tt.want.maxIdleTime {
				t.Errorf("ConnMaxIdleTime = %v, want %v", got.ConnMaxIdleTime, tt.want.maxIdleTime)
			}
			if got.Environment != tt.want.environment {
				t.Errorf("Environment = %v, want %v", got.Environment, tt.want.environment)
			}
		})
	}
}

func TestCalculateOptimalPoolSize(t *testing.T) {
	config := CalculateOptimalPoolSize()
	
	// Test basic constraints
	if config.MaxOpenConns < 10 {
		t.Errorf("MaxOpenConns too low: %v", config.MaxOpenConns)
	}
	if config.MaxOpenConns > 200 {
		t.Errorf("MaxOpenConns too high: %v", config.MaxOpenConns)
	}
	
	// Test proportions
	expectedMaxIdle := int(float64(config.MaxOpenConns) * 0.3)
	if config.MaxIdleConns != expectedMaxIdle && config.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns incorrect proportion: got %v, expected ~%v", 
			config.MaxIdleConns, expectedMaxIdle)
	}
	
	// Test CPU-based calculation
	numCPU := runtime.NumCPU()
	expectedMax := numCPU * 4
	if expectedMax < 10 {
		expectedMax = 10
	}
	if expectedMax > 200 {
		expectedMax = 200
	}
	if config.MaxOpenConns != expectedMax {
		t.Errorf("MaxOpenConns not based on CPU count: got %v, expected %v", 
			config.MaxOpenConns, expectedMax)
	}
}

func TestPoolConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *PoolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &PoolConfig{
				MaxOpenConns:    100,
				MaxIdleConns:    30,
				MinIdleConns:    10,
				ConnMaxLifetime: 15 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnectionTimeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid MaxOpenConns",
			config: &PoolConfig{
				MaxOpenConns:    0,
				MaxIdleConns:    30,
				MinIdleConns:    10,
				ConnMaxLifetime: 15 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnectionTimeout: 10 * time.Second,
			},
			wantErr: true,
			errMsg:  "MaxOpenConns must be greater than 0",
		},
		{
			name: "MaxIdleConns > MaxOpenConns",
			config: &PoolConfig{
				MaxOpenConns:    50,
				MaxIdleConns:    100,
				MinIdleConns:    10,
				ConnMaxLifetime: 15 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnectionTimeout: 10 * time.Second,
			},
			wantErr: true,
			errMsg:  "MaxIdleConns cannot be greater than MaxOpenConns",
		},
		{
			name: "MinIdleConns > MaxIdleConns",
			config: &PoolConfig{
				MaxOpenConns:    100,
				MaxIdleConns:    30,
				MinIdleConns:    50,
				ConnMaxLifetime: 15 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnectionTimeout: 10 * time.Second,
			},
			wantErr: true,
			errMsg:  "MinIdleConns cannot be greater than MaxIdleConns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestPoolOptimization(t *testing.T) {
	config := &PoolConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    30,
		MinIdleConns:    10,
		ConnMaxLifetime: 15 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	metrics := &PoolMetrics{
		OpenConnections:   90,
		IdleConnections:   5,
		InUseConnections:  85,
		WaitCount:         250,
		ConnectionChurn:   0.8,
	}

	// Test optimization
	config.Optimize(metrics)

	// Should increase MaxOpenConns due to high wait count
	if config.MaxOpenConns <= 100 {
		t.Errorf("MaxOpenConns should increase, got %v", config.MaxOpenConns)
	}

	// Should increase ConnMaxLifetime due to high churn
	if config.ConnMaxLifetime <= 15*time.Minute {
		t.Errorf("ConnMaxLifetime should increase, got %v", config.ConnMaxLifetime)
	}

	// MaxIdleConns should not exceed MaxOpenConns
	if config.MaxIdleConns > config.MaxOpenConns {
		t.Errorf("MaxIdleConns (%v) exceeds MaxOpenConns (%v)", 
			config.MaxIdleConns, config.MaxOpenConns)
	}
}

func TestGetRecommendedConfig(t *testing.T) {
	scenarios := []string{"web_api", "microservice", "analytics", "realtime", "unknown"}
	
	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			config := GetRecommendedConfig(scenario)
			
			// Validate the returned config
			if err := config.Validate(); err != nil {
				t.Errorf("Invalid config for scenario %s: %v", scenario, err)
			}
			
			// Check reasonable values
			if config.MaxOpenConns < 1 || config.MaxOpenConns > 500 {
				t.Errorf("Unreasonable MaxOpenConns for %s: %v", scenario, config.MaxOpenConns)
			}
		})
	}
}

func TestAnalyzePoolPerformance(t *testing.T) {
	config := &PoolConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    30,
		MinIdleConns:    10,
		ConnMaxLifetime: 15 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	// Test high usage scenario
	highUsageMetrics := &PoolMetrics{
		OpenConnections:   95,
		IdleConnections:   5,
		InUseConnections:  90,
		WaitCount:         100,
		WaitDuration:      200 * time.Millisecond,
		MaxLifetimeClosed: 150,
	}

	recommendations := AnalyzePoolPerformance(config, highUsageMetrics)
	
	// Should have recommendations for high usage
	if len(recommendations) == 0 {
		t.Error("Expected recommendations for high usage scenario")
	}
	
	// Check for specific recommendations
	hasHighUsageRec := false
	hasWaitTimeRec := false
	for _, rec := range recommendations {
		if rec.MetricName == "Connection Usage" {
			hasHighUsageRec = true
		}
		if rec.MetricName == "Wait Duration" {
			hasWaitTimeRec = true
		}
	}
	
	if !hasHighUsageRec {
		t.Error("Missing high usage recommendation")
	}
	if !hasWaitTimeRec {
		t.Error("Missing wait time recommendation")
	}

	// Test healthy scenario
	healthyMetrics := &PoolMetrics{
		OpenConnections:   50,
		IdleConnections:   20,
		InUseConnections:  30,
		WaitCount:         5,
		WaitDuration:      10 * time.Millisecond,
		MaxLifetimeClosed: 10,
	}

	healthyRecs := AnalyzePoolPerformance(config, healthyMetrics)
	
	// Should have fewer or no recommendations
	if len(healthyRecs) > len(recommendations) {
		t.Error("Healthy scenario should have fewer recommendations")
	}
}