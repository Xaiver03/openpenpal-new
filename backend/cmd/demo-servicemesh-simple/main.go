// Package main demonstrates a simplified service mesh implementation without external dependencies
package main

import (
	"context"
	"log"
	"time"

	"openpenpal-backend/internal/config"
)

// Simplified ServiceMeshController for demo purposes
type ServiceMeshController struct {
	config   *config.Config
	services map[string]*ServiceInstance
}

type ServiceInstance struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Health  string `json:"health"`
}

type LoadBalanceRequest struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	UserID  string `json:"user_id"`
	Timeout time.Duration `json:"timeout"`
}

// NewServiceMeshController creates a simplified service mesh controller
func NewServiceMeshController(cfg *config.Config) (*ServiceMeshController, error) {
	return &ServiceMeshController{
		config:   cfg,
		services: make(map[string]*ServiceInstance),
	}, nil
}

// Start starts the service mesh controller
func (smc *ServiceMeshController) Start(ctx context.Context) error {
	log.Println("✅ Service Mesh Controller started successfully")
	return nil
}

// RegisterService registers a service with the mesh
func (smc *ServiceMeshController) RegisterService(service *ServiceInstance) error {
	smc.services[service.ID] = service
	return nil
}

// LoadBalance selects a service instance for load balancing
func (smc *ServiceMeshController) LoadBalance(serviceName string, request *LoadBalanceRequest) (*ServiceInstance, error) {
	for _, service := range smc.services {
		if service.Name == serviceName {
			return service, nil
		}
	}
	return nil, nil
}

// DiscoverServices discovers services by name
func (smc *ServiceMeshController) DiscoverServices(serviceName string) ([]*ServiceInstance, error) {
	var instances []*ServiceInstance
	for _, service := range smc.services {
		if service.Name == serviceName {
			instances = append(instances, service)
		}
	}
	return instances, nil
}

// GetHealthyServices returns all healthy services
func (smc *ServiceMeshController) GetHealthyServices() []*ServiceInstance {
	var healthy []*ServiceInstance
	for _, service := range smc.services {
		if service.Health == "healthy" {
			healthy = append(healthy, service)
		}
	}
	return healthy
}

func main() {
	log.Println("🚀 Starting Simplified Service Mesh Demo")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Service Mesh Controller
	meshController, err := NewServiceMeshController(cfg)
	if err != nil {
		log.Fatalf("Failed to create service mesh controller: %v", err)
	}

	// Start the service mesh
	if err := meshController.Start(ctx); err != nil {
		log.Fatalf("Failed to start service mesh: %v", err)
	}

	// Demo 1: Service Registration
	log.Println("\n=== Demo 1: Service Registration ===")
	demoServiceRegistration(meshController)

	// Demo 2: Load Balancing
	log.Println("\n=== Demo 2: Load Balancing ===")
	demoLoadBalancing(meshController)

	// Demo 3: Service Discovery
	log.Println("\n=== Demo 3: Service Discovery ===")
	demoServiceDiscovery(meshController)

	log.Println("\n✅ Simplified Service Mesh Demo Completed Successfully!")
	log.Println("🎯 Core features demonstrated:")
	log.Println("   - Service Registration")
	log.Println("   - Load Balancing")
	log.Println("   - Service Discovery")
	log.Println("   - Health Management")
}

func demoServiceRegistration(controller *ServiceMeshController) {
	log.Println("📋 Registering services with the mesh...")

	services := []*ServiceInstance{
		{
			ID:      "backend-1",
			Name:    "openpenpal-backend",
			Address: "localhost",
			Port:    8080,
			Health:  "healthy",
		},
		{
			ID:      "frontend-1", 
			Name:    "openpenpal-frontend",
			Address: "localhost",
			Port:    3000,
			Health:  "healthy",
		},
		{
			ID:      "courier-service-1",
			Name:    "courier-service",
			Address: "localhost",
			Port:    8002,
			Health:  "healthy",
		},
	}

	for _, service := range services {
		if err := controller.RegisterService(service); err != nil {
			log.Printf("❌ Failed to register service %s: %v", service.Name, err)
		} else {
			log.Printf("✅ Registered service: %s (%s:%d)", service.Name, service.Address, service.Port)
		}
	}

	log.Printf("📊 Mesh Status: %d total services", len(controller.services))
}

func demoLoadBalancing(controller *ServiceMeshController) {
	log.Println("⚖️  Demonstrating load balancing...")

	requests := []*LoadBalanceRequest{
		{
			Method:  "GET",
			Path:    "/api/letters",
			UserID:  "user-123",
			Timeout: 5 * time.Second,
		},
		{
			Method:  "POST",
			Path:    "/api/comments",
			UserID:  "user-789",
			Timeout: 3 * time.Second,
		},
	}

	for i, req := range requests {
		instance, err := controller.LoadBalance("openpenpal-backend", req)
		if err != nil {
			log.Printf("❌ Load balancing failed for request %d: %v", i+1, err)
		} else if instance != nil {
			log.Printf("✅ Request %d routed to: %s (%s:%d)", 
				i+1, instance.Name, instance.Address, instance.Port)
		}
	}
}

func demoServiceDiscovery(controller *ServiceMeshController) {
	log.Println("🔍 Demonstrating service discovery...")

	serviceNames := []string{"openpenpal-backend", "courier-service", "openpenpal-frontend"}

	for _, serviceName := range serviceNames {
		instances, err := controller.DiscoverServices(serviceName)
		if err != nil {
			log.Printf("❌ Failed to discover service %s: %v", serviceName, err)
			continue
		}

		log.Printf("🔍 Discovered %d instances of service: %s", len(instances), serviceName)
		for _, instance := range instances {
			log.Printf("   📍 %s (%s:%d) - Health: %s", 
				instance.ID, instance.Address, instance.Port, instance.Health)
		}
	}

	healthyServices := controller.GetHealthyServices()
	log.Printf("💚 Total healthy services: %d", len(healthyServices))
}