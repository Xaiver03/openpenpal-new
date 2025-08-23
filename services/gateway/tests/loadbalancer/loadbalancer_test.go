package loadbalancer_test

import (
	"api-gateway/internal/config"
	"api-gateway/internal/loadbalancer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type LoadBalancerTestSuite struct {
	suite.Suite
	cfg     *config.Config
	logger  *zap.Logger
	manager *loadbalancer.Manager
}

func TestLoadBalancerTestSuite(t *testing.T) {
	suite.Run(t, new(LoadBalancerTestSuite))
}

func (suite *LoadBalancerTestSuite) SetupTest() {
	suite.cfg = &config.Config{
		Services: map[string]*config.ServiceConfig{
			"test-service": {
				Name: "test-service",
				Hosts: []string{
					"localhost:8001",
					"localhost:8002",
					"localhost:8003",
				},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"weighted-service": {
				Name: "weighted-service",
				Hosts: []string{
					"server1:8001", // Will have weight 100
					"server2:8002", // Will have weight 200
					"server3:8003", // Will have weight 50
				},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
		},
	}
	
	suite.logger = zap.NewNop()
	suite.manager = loadbalancer.NewManager(suite.cfg, suite.logger)
}

func (suite *LoadBalancerTestSuite) TestRoundRobinAlgorithm() {
	suite.Run("Basic round robin selection", func() {
		service := suite.cfg.Services["test-service"]
		algorithm := loadbalancer.NewRoundRobinAlgorithm()
		
		// Create instances
		instances := make([]*loadbalancer.Instance, len(service.Hosts))
		for i, host := range service.Hosts {
			instances[i] = &loadbalancer.Instance{
				ID:       host,
				Host:     host,
				Weight:   100,
				Healthy:  true,
				LastUsed: time.Now(),
			}
		}

		selectedHosts := make(map[string]int)
		
		// Make multiple selections
		for i := 0; i < 15; i++ {
			selected := algorithm.Select(instances)
			assert.NotNil(suite.T(), selected)
			selectedHosts[selected.Host]++
		}

		// Should distribute evenly across all instances
		assert.Equal(suite.T(), len(service.Hosts), len(selectedHosts))
		for _, count := range selectedHosts {
			assert.Equal(suite.T(), 5, count) // 15 / 3 = 5 each
		}
	})

	suite.Run("Round robin with unhealthy instances", func() {
		service := suite.cfg.Services["test-service"]
		algorithm := loadbalancer.NewRoundRobinAlgorithm()
		
		instances := make([]*loadbalancer.Instance, len(service.Hosts))
		for i, host := range service.Hosts {
			instances[i] = &loadbalancer.Instance{
				ID:      host,
				Host:    host,
				Weight:  100,
				Healthy: i != 1, // Mark second instance as unhealthy
			}
		}

		selectedHosts := make(map[string]int)
		
		for i := 0; i < 10; i++ {
			selected := algorithm.Select(instances)
			assert.NotNil(suite.T(), selected)
			selectedHosts[selected.Host]++
		}

		// Should only select healthy instances
		assert.Equal(suite.T(), 2, len(selectedHosts))
		assert.NotContains(suite.T(), selectedHosts, "localhost:8002")
		assert.Contains(suite.T(), selectedHosts, "localhost:8001")
		assert.Contains(suite.T(), selectedHosts, "localhost:8003")
	})
}

func (suite *LoadBalancerTestSuite) TestWeightedRoundRobinAlgorithm() {
	suite.Run("Weighted round robin selection", func() {
		algorithm := loadbalancer.NewWeightedRoundRobinAlgorithm()
		
		instances := []*loadbalancer.Instance{
			{
				ID:      "server1",
				Host:    "server1:8001",
				Weight:  100,
				Healthy: true,
			},
			{
				ID:      "server2",
				Host:    "server2:8002",
				Weight:  200, // Double weight
				Healthy: true,
			},
			{
				ID:      "server3",
				Host:    "server3:8003",
				Weight:  50, // Half weight
				Healthy: true,
			},
		}

		selectedHosts := make(map[string]int)
		
		// Make many selections to see weight distribution
		for i := 0; i < 350; i++ { // 100+200+50 = 350 for even distribution
			selected := algorithm.Select(instances)
			assert.NotNil(suite.T(), selected)
			selectedHosts[selected.Host]++
		}

		// Check weight-proportional distribution
		assert.Equal(suite.T(), 3, len(selectedHosts))
		
		// Server2 should get approximately twice as many requests as Server1
		server1Count := selectedHosts["server1:8001"]
		server2Count := selectedHosts["server2:8002"]
		server3Count := selectedHosts["server3:8003"]
		
		assert.Greater(suite.T(), server2Count, server1Count)
		assert.Greater(suite.T(), server1Count, server3Count)
		
		// Approximate ratio check (allowing for some variance)
		ratio2to1 := float64(server2Count) / float64(server1Count)
		ratio1to3 := float64(server1Count) / float64(server3Count)
		
		assert.InDelta(suite.T(), 2.0, ratio2to1, 0.3) // Should be close to 2:1
		assert.InDelta(suite.T(), 2.0, ratio1to3, 0.3) // Should be close to 2:1
	})
}

func (suite *LoadBalancerTestSuite) TestLeastConnectionsAlgorithm() {
	suite.Run("Least connections selection", func() {
		algorithm := loadbalancer.NewLeastConnectionsAlgorithm()
		
		instances := []*loadbalancer.Instance{
			{
				ID:               "server1",
				Host:             "server1:8001",
				Weight:           100,
				Healthy:          true,
				ActiveConnections: 5,
			},
			{
				ID:               "server2",
				Host:             "server2:8002",
				Weight:           100,
				Healthy:          true,
				ActiveConnections: 2, // Fewer connections
			},
			{
				ID:               "server3",
				Host:             "server3:8003",
				Weight:           100,
				Healthy:          true,
				ActiveConnections: 8,
			},
		}

		// Should select server with least connections
		selected := algorithm.Select(instances)
		assert.NotNil(suite.T(), selected)
		assert.Equal(suite.T(), "server2:8002", selected.Host)
		
		// Update connections and test again
		instances[1].ActiveConnections = 10 // Now server2 has most connections
		
		selected = algorithm.Select(instances)
		assert.NotNil(suite.T(), selected)
		assert.Equal(suite.T(), "server1:8001", selected.Host) // Should select server1 now
	})
}

func (suite *LoadBalancerTestSuite) TestAdaptiveAlgorithm() {
	suite.Run("Adaptive algorithm considers response time", func() {
		algorithm := loadbalancer.NewAdaptiveAlgorithm()
		
		instances := []*loadbalancer.Instance{
			{
				ID:               "fast-server",
				Host:             "fast-server:8001",
				Weight:           100,
				Healthy:          true,
				ActiveConnections: 3,
				ResponseTime:     50 * time.Millisecond,
			},
			{
				ID:               "slow-server",
				Host:             "slow-server:8002",
				Weight:           100,
				Healthy:          true,
				ActiveConnections: 2,
				ResponseTime:     500 * time.Millisecond, // Much slower
			},
		}

		selectedHosts := make(map[string]int)
		
		// Make multiple selections
		for i := 0; i < 20; i++ {
			selected := algorithm.Select(instances)
			assert.NotNil(suite.T(), selected)
			selectedHosts[selected.Host]++
		}

		// Fast server should be selected more often despite having more connections
		fastCount := selectedHosts["fast-server:8001"]
		slowCount := selectedHosts["slow-server:8002"]
		
		assert.Greater(suite.T(), fastCount, slowCount)
	})
}

func (suite *LoadBalancerTestSuite) TestLoadBalancerManager() {
	suite.Run("Load balancer manager initialization", func() {
		// Manager should be properly initialized
		assert.NotNil(suite.T(), suite.manager)
		
		// Should have load balancers for configured services
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		nonExistentLB := suite.manager.GetLoadBalancer("non-existent-service")
		assert.Nil(suite.T(), nonExistentLB)
	})

	suite.Run("Select instance from service", func() {
		// Get load balancer for test service
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		// Select an instance
		instance := lb.SelectInstance()
		assert.NotNil(suite.T(), instance)
		assert.Contains(suite.T(), []string{"localhost:8001", "localhost:8002", "localhost:8003"}, instance.Host)
		assert.True(suite.T(), instance.Healthy)
	})

	suite.Run("Update instance health", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		// Mark an instance as unhealthy
		lb.MarkInstanceUnhealthy("localhost:8002")
		
		// Select multiple instances - should not include unhealthy one
		selectedHosts := make(map[string]bool)
		for i := 0; i < 10; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host] = true
			}
		}
		
		assert.NotContains(suite.T(), selectedHosts, "localhost:8002")
		assert.Contains(suite.T(), selectedHosts, "localhost:8001")
		assert.Contains(suite.T(), selectedHosts, "localhost:8003")
		
		// Mark it as healthy again
		lb.MarkInstanceHealthy("localhost:8002")
		
		// Should now include all instances
		selectedHosts = make(map[string]bool)
		for i := 0; i < 15; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host] = true
			}
		}
		
		assert.Contains(suite.T(), selectedHosts, "localhost:8001")
		assert.Contains(suite.T(), selectedHosts, "localhost:8002")
		assert.Contains(suite.T(), selectedHosts, "localhost:8003")
	})
}

func (suite *LoadBalancerTestSuite) TestSessionAffinity() {
	suite.Run("Session affinity maintains stickiness", func() {
		algorithm := loadbalancer.NewSessionAffinityAlgorithm("ip-hash")
		
		instances := []*loadbalancer.Instance{
			{ID: "server1", Host: "server1:8001", Healthy: true},
			{ID: "server2", Host: "server2:8002", Healthy: true},
			{ID: "server3", Host: "server3:8003", Healthy: true},
		}

		// Same session key should always get same instance
		sessionKey := "user123"
		
		var selectedInstance *loadbalancer.Instance
		for i := 0; i < 10; i++ {
			selected := algorithm.SelectWithSession(instances, sessionKey)
			assert.NotNil(suite.T(), selected)
			
			if selectedInstance == nil {
				selectedInstance = selected
			} else {
				assert.Equal(suite.T(), selectedInstance.Host, selected.Host)
			}
		}
		
		// Different session key might get different instance
		differentSession := algorithm.SelectWithSession(instances, "user456")
		assert.NotNil(suite.T(), differentSession)
	})
}

func (suite *LoadBalancerTestSuite) TestInstanceMetrics() {
	suite.Run("Instance metrics tracking", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		instance := lb.SelectInstance()
		assert.NotNil(suite.T(), instance)
		
		// Record a request
		lb.RecordRequest(instance.ID, 100*time.Millisecond, true)
		
		// Get metrics
		metrics := lb.GetInstanceMetrics(instance.ID)
		assert.NotNil(suite.T(), metrics)
		assert.Equal(suite.T(), uint64(1), metrics.RequestCount)
		assert.Equal(suite.T(), uint64(1), metrics.SuccessCount)
		assert.Equal(suite.T(), uint64(0), metrics.ErrorCount)
		assert.Equal(suite.T(), 100*time.Millisecond, metrics.AvgResponseTime)
		
		// Record a failed request
		lb.RecordRequest(instance.ID, 500*time.Millisecond, false)
		
		metrics = lb.GetInstanceMetrics(instance.ID)
		assert.Equal(suite.T(), uint64(2), metrics.RequestCount)
		assert.Equal(suite.T(), uint64(1), metrics.SuccessCount)
		assert.Equal(suite.T(), uint64(1), metrics.ErrorCount)
		assert.Greater(suite.T(), metrics.AvgResponseTime, 100*time.Millisecond)
	})
}

func (suite *LoadBalancerTestSuite) TestCircuitBreakerIntegration() {
	suite.Run("Circuit breaker integration", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		instance := lb.SelectInstance()
		assert.NotNil(suite.T(), instance)
		
		// Record multiple failed requests to trigger circuit breaker
		for i := 0; i < 10; i++ {
			lb.RecordRequest(instance.ID, 1*time.Second, false)
		}
		
		// Instance should be marked as unhealthy
		metrics := lb.GetInstanceMetrics(instance.ID)
		assert.Greater(suite.T(), metrics.ErrorCount, uint64(5))
		
		// Circuit breaker should eventually mark instance as unhealthy
		time.Sleep(100 * time.Millisecond) // Allow circuit breaker to process
		
		// Check if instance is excluded from selection
		selectedHosts := make(map[string]int)
		for i := 0; i < 20; i++ {
			selected := lb.SelectInstance()
			if selected != nil {
				selectedHosts[selected.Host]++
			}
		}
		
		// Failing instance should get fewer or no requests
		if count, exists := selectedHosts[instance.Host]; exists {
			assert.LessOrEqual(suite.T(), count, 5) // Should get limited requests
		}
	})
}

func (suite *LoadBalancerTestSuite) TestDynamicConfiguration() {
	suite.Run("Dynamic instance addition and removal", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		// Add a new instance dynamically
		newInstance := &loadbalancer.Instance{
			ID:      "dynamic-server",
			Host:    "dynamic-server:8004",
			Weight:  100,
			Healthy: true,
		}
		
		lb.AddInstance(newInstance)
		
		// Should now be able to select the new instance
		selectedHosts := make(map[string]bool)
		for i := 0; i < 20; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host] = true
			}
		}
		
		assert.Contains(suite.T(), selectedHosts, "dynamic-server:8004")
		
		// Remove the instance
		lb.RemoveInstance("dynamic-server")
		
		// Should no longer select the removed instance
		selectedHosts = make(map[string]bool)
		for i := 0; i < 20; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host] = true
			}
		}
		
		assert.NotContains(suite.T(), selectedHosts, "dynamic-server:8004")
	})
}

func (suite *LoadBalancerTestSuite) TestWeightedSelection() {
	suite.Run("Weight-based selection distribution", func() {
		algorithm := loadbalancer.NewWeightedRoundRobinAlgorithm()
		
		instances := []*loadbalancer.Instance{
			{ID: "light", Host: "light:8001", Weight: 1, Healthy: true},
			{ID: "medium", Host: "medium:8002", Weight: 3, Healthy: true},
			{ID: "heavy", Host: "heavy:8003", Weight: 6, Healthy: true},
		}

		selectedHosts := make(map[string]int)
		totalSelections := 1000
		
		for i := 0; i < totalSelections; i++ {
			selected := algorithm.Select(instances)
			assert.NotNil(suite.T(), selected)
			selectedHosts[selected.Host]++
		}
		
		// Calculate weight ratios
		totalWeight := 1 + 3 + 6 // 10
		expectedLight := float64(totalSelections) * 1 / 10   // ~100
		expectedMedium := float64(totalSelections) * 3 / 10  // ~300
		expectedHeavy := float64(totalSelections) * 6 / 10   // ~600
		
		// Allow 10% variance
		tolerance := 0.1
		assert.InDelta(suite.T(), expectedLight, selectedHosts["light:8001"], expectedLight*tolerance)
		assert.InDelta(suite.T(), expectedMedium, selectedHosts["medium:8002"], expectedMedium*tolerance)
		assert.InDelta(suite.T(), expectedHeavy, selectedHosts["heavy:8003"], expectedHeavy*tolerance)
	})
}

func (suite *LoadBalancerTestSuite) TestAlgorithmSwitching() {
	suite.Run("Switch between different algorithms", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		// Start with round robin
		lb.SetAlgorithm("round-robin")
		
		selectedHosts := make(map[string]int)
		for i := 0; i < 15; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host]++
			}
		}
		
		// Should distribute evenly
		for _, count := range selectedHosts {
			assert.Equal(suite.T(), 5, count)
		}
		
		// Switch to weighted round robin
		lb.SetAlgorithm("weighted-round-robin")
		
		// Reset counters
		selectedHosts = make(map[string]int)
		for i := 0; i < 30; i++ {
			instance := lb.SelectInstance()
			if instance != nil {
				selectedHosts[instance.Host]++
			}
		}
		
		// Distribution should still be relatively even for equal weights
		assert.Equal(suite.T(), 3, len(selectedHosts))
	})
}

func (suite *LoadBalancerTestSuite) TestConcurrentAccess() {
	suite.Run("Concurrent load balancer access", func() {
		lb := suite.manager.GetLoadBalancer("test-service")
		assert.NotNil(suite.T(), lb)
		
		// Run concurrent selections
		results := make(chan string, 100)
		
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					instance := lb.SelectInstance()
					if instance != nil {
						results <- instance.Host
					}
				}
			}()
		}
		
		// Collect results
		selectedHosts := make(map[string]int)
		for i := 0; i < 100; i++ {
			host := <-results
			selectedHosts[host]++
		}
		
		// Should have selected all available hosts
		assert.Equal(suite.T(), 3, len(selectedHosts))
		assert.Equal(suite.T(), 100, selectedHosts["localhost:8001"]+selectedHosts["localhost:8002"]+selectedHosts["localhost:8003"])
	})
}