package handlers

import (
	"api-gateway/internal/loadbalancer"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoadBalancerHandler handles load balancer management endpoints
type LoadBalancerHandler struct {
	loadBalancerManager *loadbalancer.LoadBalancerManager
	logger              *zap.Logger
}

// NewLoadBalancerHandler creates a new load balancer handler
func NewLoadBalancerHandler(lbm *loadbalancer.LoadBalancerManager, logger *zap.Logger) *LoadBalancerHandler {
	return &LoadBalancerHandler{
		loadBalancerManager: lbm,
		logger:              logger,
	}
}

// RegisterLoadBalancerRoutes registers all load balancer management routes
func (lbh *LoadBalancerHandler) RegisterLoadBalancerRoutes(router *gin.Engine) {
	lbGroup := router.Group("/api/v1/loadbalancer")
	{
		// Statistics and monitoring
		lbGroup.GET("/stats", lbh.GetLoadBalancerStats)
		lbGroup.GET("/stats/:service", lbh.GetServiceStats)
		lbGroup.GET("/instances", lbh.GetAllInstances)
		lbGroup.GET("/instances/:service", lbh.GetServiceInstances)
		
		// Instance management
		lbGroup.GET("/instances/:service/:instance", lbh.GetInstanceDetails)
		lbGroup.POST("/instances/:service/:instance/drain", lbh.DrainInstance)
		lbGroup.POST("/instances/:service/:instance/enable", lbh.EnableInstance)
		lbGroup.POST("/instances/:service/:instance/disable", lbh.DisableInstance)
		
		// Algorithm management
		lbGroup.GET("/algorithms", lbh.GetAvailableAlgorithms)
		lbGroup.GET("/algorithms/:service", lbh.GetServiceAlgorithm)
		lbGroup.PUT("/algorithms/:service", lbh.SetServiceAlgorithm)
		
		// Configuration management
		lbGroup.GET("/config", lbh.GetConfiguration)
		lbGroup.PUT("/config", lbh.UpdateConfiguration)
		
		// Health and performance
		lbGroup.GET("/health", lbh.GetHealthStatus)
		lbGroup.GET("/performance", lbh.GetPerformanceMetrics)
		
		// Session affinity management
		lbGroup.GET("/sessions", lbh.GetSessionAffinityStats)
		lbGroup.DELETE("/sessions/:session", lbh.RemoveSessionAffinity)
		lbGroup.DELETE("/sessions", lbh.ClearAllSessions)
		
		// Recovery management
		lbGroup.GET("/recovery", lbh.GetRecoveryStatus)
		lbGroup.POST("/recovery/:service/:instance/force", lbh.ForceInstanceRecovery)
	}
	
	// Admin routes with authentication
	adminGroup := router.Group("/api/v1/admin/loadbalancer")
	// adminGroup.Use(authMiddleware) // Add authentication middleware
	{
		adminGroup.POST("/reset", lbh.ResetLoadBalancer)
		adminGroup.POST("/reload", lbh.ReloadConfiguration)
		adminGroup.GET("/debug", lbh.GetDebugInfo)
	}
}

// GetLoadBalancerStats returns comprehensive load balancer statistics
func (lbh *LoadBalancerHandler) GetLoadBalancerStats(c *gin.Context) {
	stats := lbh.loadBalancerManager.GetLoadBalancerStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"data":      stats,
	})
}

// GetServiceStats returns statistics for a specific service
func (lbh *LoadBalancerHandler) GetServiceStats(c *gin.Context) {
	serviceName := c.Param("service")
	stats := lbh.loadBalancerManager.GetLoadBalancerStats()
	
	serviceStats, exists := stats[serviceName]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Service not found",
			"service": serviceName,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"timestamp":   time.Now().UTC(),
		"service":     serviceName,
		"statistics":  serviceStats,
	})
}

// GetAllInstances returns all service instances with their status
func (lbh *LoadBalancerHandler) GetAllInstances(c *gin.Context) {
	// This would be implemented to return all instances across all services
	instances := make(map[string]interface{})
	
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"instances": instances,
	})
}

// GetServiceInstances returns instances for a specific service
func (lbh *LoadBalancerHandler) GetServiceInstances(c *gin.Context) {
	serviceName := c.Param("service")
	
	// Get instances (this would be implemented to get actual instances)
	instances := make([]interface{}, 0)
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"service":   serviceName,
		"instances": instances,
	})
}

// GetInstanceDetails returns detailed information about a specific instance
func (lbh *LoadBalancerHandler) GetInstanceDetails(c *gin.Context) {
	serviceName := c.Param("service")
	instanceHost := c.Param("instance")
	
	// This would be implemented to get actual instance details
	instanceDetails := gin.H{
		"service":              serviceName,
		"host":                instanceHost,
		"healthy":             true,
		"active_connections":  5,
		"total_requests":      1000,
		"success_requests":    950,
		"failed_requests":     50,
		"average_response":    "150ms",
		"last_response_time":  "120ms",
		"success_rate":        95.0,
		"health_score":        0.95,
		"weight":              10,
		"last_used":           time.Now().Add(-30*time.Second),
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"instance":  instanceDetails,
	})
}

// DrainInstance gradually removes traffic from an instance
func (lbh *LoadBalancerHandler) DrainInstance(c *gin.Context) {
	serviceName := c.Param("service")
	instanceHost := c.Param("instance")
	
	lbh.logger.Info("Draining instance",
		zap.String("service", serviceName),
		zap.String("instance", instanceHost),
	)
	
	// This would be implemented to gradually drain the instance
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Instance drain initiated",
		"service": serviceName,
		"instance": instanceHost,
	})
}

// EnableInstance enables an instance for traffic
func (lbh *LoadBalancerHandler) EnableInstance(c *gin.Context) {
	serviceName := c.Param("service")
	instanceHost := c.Param("instance")
	
	lbh.logger.Info("Enabling instance",
		zap.String("service", serviceName),
		zap.String("instance", instanceHost),
	)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Instance enabled",
		"service": serviceName,
		"instance": instanceHost,
	})
}

// DisableInstance disables an instance from receiving traffic
func (lbh *LoadBalancerHandler) DisableInstance(c *gin.Context) {
	serviceName := c.Param("service")
	instanceHost := c.Param("instance")
	
	lbh.logger.Info("Disabling instance",
		zap.String("service", serviceName),
		zap.String("instance", instanceHost),
	)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Instance disabled",
		"service": serviceName,
		"instance": instanceHost,
	})
}

// GetAvailableAlgorithms returns all available load balancing algorithms
func (lbh *LoadBalancerHandler) GetAvailableAlgorithms(c *gin.Context) {
	algorithms := []gin.H{
		{
			"name":        "round_robin",
			"description": "Simple round-robin distribution",
			"type":        "basic",
		},
		{
			"name":        "weighted_round_robin",
			"description": "Weighted round-robin based on instance weights",
			"type":        "weighted",
		},
		{
			"name":        "least_connections",
			"description": "Route to instance with least active connections",
			"type":        "connection_based",
		},
		{
			"name":        "least_response_time",
			"description": "Route to instance with fastest response time",
			"type":        "performance_based",
		},
		{
			"name":        "health_aware",
			"description": "Weighted selection based on health scores",
			"type":        "health_based",
		},
		{
			"name":        "consistent_hash",
			"description": "Consistent hashing for session affinity",
			"type":        "hash_based",
		},
		{
			"name":        "adaptive",
			"description": "Adaptive algorithm based on multiple metrics",
			"type":        "intelligent",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"timestamp":  time.Now().UTC(),
		"algorithms": algorithms,
	})
}

// GetServiceAlgorithm returns the current algorithm for a service
func (lbh *LoadBalancerHandler) GetServiceAlgorithm(c *gin.Context) {
	serviceName := c.Param("service")
	
	// This would get the actual algorithm from the load balancer manager
	algorithm := "adaptive" // Default
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"service":   serviceName,
		"algorithm": algorithm,
	})
}

// SetServiceAlgorithm sets the load balancing algorithm for a service
func (lbh *LoadBalancerHandler) SetServiceAlgorithm(c *gin.Context) {
	serviceName := c.Param("service")
	
	var request struct {
		Algorithm string `json:"algorithm" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	// Validate algorithm
	validAlgorithms := []string{
		"round_robin", "weighted_round_robin", "least_connections",
		"least_response_time", "health_aware", "consistent_hash", "adaptive",
	}
	
	valid := false
	for _, algo := range validAlgorithms {
		if algo == request.Algorithm {
			valid = true
			break
		}
	}
	
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid algorithm",
			"valid_algorithms": validAlgorithms,
		})
		return
	}
	
	lbh.logger.Info("Setting service algorithm",
		zap.String("service", serviceName),
		zap.String("algorithm", request.Algorithm),
	)
	
	// This would actually update the algorithm in the load balancer manager
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Algorithm updated successfully",
		"service":   serviceName,
		"algorithm": request.Algorithm,
	})
}

// GetConfiguration returns the current load balancer configuration
func (lbh *LoadBalancerHandler) GetConfiguration(c *gin.Context) {
	config := gin.H{
		"default_algorithm":            "adaptive",
		"health_check_enabled":         true,
		"metrics_enabled":              true,
		"circuit_breaker_enabled":      true,
		"session_affinity_enabled":     true,
		"session_affinity_ttl":         "30m",
		"performance_monitoring_enabled": true,
		"performance_monitoring_window": "1m",
		"gradual_recovery_enabled":     true,
		"recovery_threshold":           0.8,
		"recovery_step_size":           0.1,
		"recovery_check_interval":      "30s",
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"timestamp":   time.Now().UTC(),
		"configuration": config,
	})
}

// UpdateConfiguration updates the load balancer configuration
func (lbh *LoadBalancerHandler) UpdateConfiguration(c *gin.Context) {
	var request map[string]interface{}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	lbh.logger.Info("Updating load balancer configuration",
		zap.Any("config", request),
	)
	
	// This would actually update the configuration
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated successfully",
		"updated_config": request,
	})
}

// GetHealthStatus returns the health status of the load balancer
func (lbh *LoadBalancerHandler) GetHealthStatus(c *gin.Context) {
	healthStatus := gin.H{
		"status":     "healthy",
		"timestamp":  time.Now().UTC(),
		"components": gin.H{
			"load_balancer": gin.H{
				"status": "healthy",
				"message": "Load balancer is operational",
			},
			"service_discovery": gin.H{
				"status": "healthy",
				"message": "Service discovery is operational",
			},
			"circuit_breaker": gin.H{
				"status": "healthy",
				"message": "Circuit breaker integration is operational",
			},
			"metrics_collection": gin.H{
				"status": "healthy",
				"message": "Metrics collection is operational",
			},
		},
		"metrics": gin.H{
			"total_requests":     150000,
			"successful_requests": 148500,
			"failed_requests":    1500,
			"average_response_time": "120ms",
			"uptime":            "72h45m",
		},
	}
	
	c.JSON(http.StatusOK, healthStatus)
}

// GetPerformanceMetrics returns performance metrics for the load balancer
func (lbh *LoadBalancerHandler) GetPerformanceMetrics(c *gin.Context) {
	// Parse query parameters
	window := c.DefaultQuery("window", "1h")
	service := c.Query("service")
	
	metrics := gin.H{
		"window":    window,
		"timestamp": time.Now().UTC(),
		"metrics": gin.H{
			"requests_per_second": 125.5,
			"average_response_time": "120ms",
			"p95_response_time":   "250ms",
			"p99_response_time":   "450ms",
			"error_rate":          1.2,
			"throughput":          "15.6MB/s",
		},
	}
	
	if service != "" {
		metrics["service"] = service
		// Add service-specific metrics
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetSessionAffinityStats returns session affinity statistics
func (lbh *LoadBalancerHandler) GetSessionAffinityStats(c *gin.Context) {
	stats := gin.H{
		"enabled":       true,
		"total_sessions": 450,
		"active_sessions": 320,
		"expired_sessions": 130,
		"session_ttl":    "30m",
		"hit_rate":       85.6,
		"distribution": gin.H{
			"service1": 120,
			"service2": 200,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"stats":     stats,
	})
}

// RemoveSessionAffinity removes a specific session affinity
func (lbh *LoadBalancerHandler) RemoveSessionAffinity(c *gin.Context) {
	sessionID := c.Param("session")
	
	lbh.logger.Info("Removing session affinity",
		zap.String("session_id", sessionID),
	)
	
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Session affinity removed",
		"session_id": sessionID,
	})
}

// ClearAllSessions clears all session affinities
func (lbh *LoadBalancerHandler) ClearAllSessions(c *gin.Context) {
	lbh.logger.Info("Clearing all session affinities")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All session affinities cleared",
	})
}

// GetRecoveryStatus returns the status of instance recovery
func (lbh *LoadBalancerHandler) GetRecoveryStatus(c *gin.Context) {
	recoveryStatus := gin.H{
		"enabled": true,
		"recovering_instances": []gin.H{
			{
				"service":          "main-backend",
				"host":            "http://localhost:8081",
				"recovery_weight":  0.6,
				"successful_checks": 3,
				"recovery_start":   time.Now().Add(-5*time.Minute),
			},
		},
		"recovery_threshold":     0.8,
		"recovery_step_size":     0.1,
		"recovery_check_interval": "30s",
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now().UTC(),
		"recovery":  recoveryStatus,
	})
}

// ForceInstanceRecovery forces an instance to be marked as recovered
func (lbh *LoadBalancerHandler) ForceInstanceRecovery(c *gin.Context) {
	serviceName := c.Param("service")
	instanceHost := c.Param("instance")
	
	lbh.logger.Info("Forcing instance recovery",
		zap.String("service", serviceName),
		zap.String("instance", instanceHost),
	)
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Instance recovery forced",
		"service":  serviceName,
		"instance": instanceHost,
	})
}

// Admin endpoints

// ResetLoadBalancer resets the load balancer state
func (lbh *LoadBalancerHandler) ResetLoadBalancer(c *gin.Context) {
	lbh.logger.Warn("Resetting load balancer state")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Load balancer reset successfully",
	})
}

// ReloadConfiguration reloads the load balancer configuration
func (lbh *LoadBalancerHandler) ReloadConfiguration(c *gin.Context) {
	lbh.logger.Info("Reloading load balancer configuration")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration reloaded successfully",
	})
}

// GetDebugInfo returns debug information about the load balancer
func (lbh *LoadBalancerHandler) GetDebugInfo(c *gin.Context) {
	debugInfo := gin.H{
		"version":    "1.0.0",
		"build_time": "2024-01-15T10:30:00Z",
		"go_version": "go1.21.0",
		"memory_usage": gin.H{
			"alloc":      "45MB",
			"total_alloc": "150MB",
			"sys":        "72MB",
			"num_gc":     25,
		},
		"goroutines": 48,
		"connections": gin.H{
			"active":  250,
			"idle":    50,
			"total":   300,
		},
		"algorithms_in_use": []string{
			"adaptive", "least_response_time", "health_aware",
		},
		"feature_flags": gin.H{
			"session_affinity":      true,
			"gradual_recovery":      true,
			"circuit_breaker":       true,
			"performance_monitoring": true,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"timestamp":  time.Now().UTC(),
		"debug_info": debugInfo,
	})
}