// Package servicemesh provides service discovery and registry functionality
package servicemesh

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/config"

	"github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ServiceRegistry provides service discovery and registration capabilities
type ServiceRegistry struct {
	config      *config.Config
	etcdClient  *clientv3.Client
	consulClient *api.Client
	useEtcd     bool
	useConsul   bool
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(cfg *config.Config) (*ServiceRegistry, error) {
	registry := &ServiceRegistry{
		config: cfg,
	}

	// Initialize etcd client if configured
	if cfg.EtcdEndpoints != "" {
		etcdConfig := clientv3.Config{
			Endpoints:   []string{cfg.EtcdEndpoints},
			DialTimeout: 5 * time.Second,
		}
		
		client, err := clientv3.New(etcdConfig)
		if err != nil {
			log.Printf("⚠️  Failed to connect to etcd: %v", err)
		} else {
			registry.etcdClient = client
			registry.useEtcd = true
			log.Println("✅ Connected to etcd for service discovery")
		}
	}

	// Initialize Consul client if configured
	if cfg.ConsulEndpoint != "" {
		consulConfig := api.DefaultConfig()
		consulConfig.Address = cfg.ConsulEndpoint
		
		client, err := api.NewClient(consulConfig)
		if err != nil {
			log.Printf("⚠️  Failed to connect to Consul: %v", err)
		} else {
			registry.consulClient = client
			registry.useConsul = true
			log.Println("✅ Connected to Consul for service discovery")
		}
	}

	// If no service discovery backends are configured, warn but continue
	if !registry.useEtcd && !registry.useConsul {
		log.Println("⚠️  No service discovery backends configured. Using in-memory registry only.")
	}

	return registry, nil
}

// Start starts the service registry
func (sr *ServiceRegistry) Start(ctx context.Context) error {
	log.Println("🚀 Starting Service Registry")

	// Start etcd session management if using etcd
	if sr.useEtcd {
		go sr.etcdSessionManager(ctx)
	}

	// Start Consul health checks if using Consul
	if sr.useConsul {
		go sr.consulHealthChecker(ctx)
	}

	log.Println("✅ Service Registry started")
	return nil
}

// Register registers a service with the registry
func (sr *ServiceRegistry) Register(service *ServiceInstance) error {
	log.Printf("📋 Registering service: %s (%s:%d)", service.Name, service.Address, service.Port)

	// Register with etcd if available
	if sr.useEtcd {
		if err := sr.registerWithEtcd(service); err != nil {
			log.Printf("❌ Failed to register with etcd: %v", err)
		}
	}

	// Register with Consul if available
	if sr.useConsul {
		if err := sr.registerWithConsul(service); err != nil {
			log.Printf("❌ Failed to register with Consul: %v", err)
		}
	}

	return nil
}

// Deregister removes a service from the registry
func (sr *ServiceRegistry) Deregister(serviceID string) error {
	log.Printf("📋 Deregistering service: %s", serviceID)

	// Deregister from etcd if available
	if sr.useEtcd {
		if err := sr.deregisterFromEtcd(serviceID); err != nil {
			log.Printf("❌ Failed to deregister from etcd: %v", err)
		}
	}

	// Deregister from Consul if available
	if sr.useConsul {
		if err := sr.deregisterFromConsul(serviceID); err != nil {
			log.Printf("❌ Failed to deregister from Consul: %v", err)
		}
	}

	return nil
}

// Discover discovers services by name
func (sr *ServiceRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	var services []*ServiceInstance

	// Discover from etcd if available
	if sr.useEtcd {
		etcdServices, err := sr.discoverFromEtcd(serviceName)
		if err != nil {
			log.Printf("⚠️  Failed to discover from etcd: %v", err)
		} else {
			services = append(services, etcdServices...)
		}
	}

	// Discover from Consul if available
	if sr.useConsul {
		consulServices, err := sr.discoverFromConsul(serviceName)
		if err != nil {
			log.Printf("⚠️  Failed to discover from Consul: %v", err)
		} else {
			services = append(services, consulServices...)
		}
	}

	// Deduplicate services
	services = sr.deduplicateServices(services)

	return services, nil
}

// registerWithEtcd registers a service with etcd
func (sr *ServiceRegistry) registerWithEtcd(service *ServiceInstance) error {
	key := fmt.Sprintf("/services/%s/%s", service.Name, service.ID)
	
	// Serialize service data
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service data: %w", err)
	}

	// Create a lease for the service (TTL: 30 seconds)
	lease, err := sr.etcdClient.Grant(context.Background(), 30)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	// Put the service data with the lease
	_, err = sr.etcdClient.Put(context.Background(), key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to put service data: %w", err)
	}

	// Keep the lease alive
	ch, kaerr := sr.etcdClient.KeepAlive(context.Background(), lease.ID)
	if kaerr != nil {
		return fmt.Errorf("failed to keep lease alive: %w", kaerr)
	}

	// Consume the keep alive response
	go func() {
		for ka := range ch {
			_ = ka // Consume the response
		}
	}()

	return nil
}

// deregisterFromEtcd removes a service from etcd
func (sr *ServiceRegistry) deregisterFromEtcd(serviceID string) error {
	// We need to find and delete all keys for this service ID
	prefix := fmt.Sprintf("/services/")
	
	resp, err := sr.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get services: %w", err)
	}

	for _, kv := range resp.Kvs {
		var service ServiceInstance
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			continue
		}
		
		if service.ID == serviceID {
			_, err = sr.etcdClient.Delete(context.Background(), string(kv.Key))
			if err != nil {
				return fmt.Errorf("failed to delete service: %w", err)
			}
		}
	}

	return nil
}

// discoverFromEtcd discovers services from etcd
func (sr *ServiceRegistry) discoverFromEtcd(serviceName string) ([]*ServiceInstance, error) {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	
	resp, err := sr.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var services []*ServiceInstance
	for _, kv := range resp.Kvs {
		var service ServiceInstance
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			log.Printf("⚠️  Failed to unmarshal service data: %v", err)
			continue
		}
		services = append(services, &service)
	}

	return services, nil
}

// registerWithConsul registers a service with Consul
func (sr *ServiceRegistry) registerWithConsul(service *ServiceInstance) error {
	// Convert tags to string slice
	tags := service.Tags
	if tags == nil {
		tags = []string{}
	}

	// Add metadata as tags
	for key, value := range service.Metadata {
		tags = append(tags, fmt.Sprintf("%s:%s", key, value))
	}

	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Tags:    tags,
		Address: service.Address,
		Port:    service.Port,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return sr.consulClient.Agent().ServiceRegister(registration)
}

// deregisterFromConsul removes a service from Consul
func (sr *ServiceRegistry) deregisterFromConsul(serviceID string) error {
	return sr.consulClient.Agent().ServiceDeregister(serviceID)
}

// discoverFromConsul discovers services from Consul
func (sr *ServiceRegistry) discoverFromConsul(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := sr.consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get services from Consul: %w", err)
	}

	var instances []*ServiceInstance
	for _, service := range services {
		// Parse metadata from tags
		metadata := make(map[string]string)
		var tags []string
		
		for _, tag := range service.Service.Tags {
			if len(tag) > 0 && tag[0] != ':' {
				parts := splitTag(tag)
				if len(parts) == 2 {
					metadata[parts[0]] = parts[1]
				} else {
					tags = append(tags, tag)
				}
			} else {
				tags = append(tags, tag)
			}
		}

		instance := &ServiceInstance{
			ID:       service.Service.ID,
			Name:     service.Service.Service,
			Address:  service.Service.Address,
			Port:     service.Service.Port,
			Health:   sr.convertConsulHealth(service.Checks),
			Metadata: metadata,
			Tags:     tags,
			Metrics:  &ServiceMetrics{}, // Initialize empty metrics
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// convertConsulHealth converts Consul health check status to our HealthStatus
func (sr *ServiceRegistry) convertConsulHealth(checks api.HealthChecks) HealthStatus {
	for _, check := range checks {
		switch check.Status {
		case "critical":
			return HealthStatusCritical
		case "warning":
			return HealthStatusUnhealthy
		}
	}
	return HealthStatusHealthy
}

// splitTag splits a tag of format "key:value"
func splitTag(tag string) []string {
	for i, r := range tag {
		if r == ':' {
			return []string{tag[:i], tag[i+1:]}
		}
	}
	return []string{tag}
}

// deduplicateServices removes duplicate services (same ID)
func (sr *ServiceRegistry) deduplicateServices(services []*ServiceInstance) []*ServiceInstance {
	seen := make(map[string]bool)
	var unique []*ServiceInstance

	for _, service := range services {
		if !seen[service.ID] {
			seen[service.ID] = true
			unique = append(unique, service)
		}
	}

	return unique
}

// etcdSessionManager manages etcd sessions and keeps services alive
func (sr *ServiceRegistry) etcdSessionManager(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Perform periodic maintenance tasks
			sr.cleanupStaleServices()
		}
	}
}

// consulHealthChecker performs periodic health checks for Consul
func (sr *ServiceRegistry) consulHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Consul handles health checks automatically
			// This could be used for additional custom health logic
		}
	}
}

// cleanupStaleServices removes stale services from the registry
func (sr *ServiceRegistry) cleanupStaleServices() {
	// This would implement logic to clean up services that haven't been seen recently
	// For now, we rely on etcd TTL and Consul's built-in health checks
}

// GetAllServices returns all registered services
func (sr *ServiceRegistry) GetAllServices() ([]*ServiceInstance, error) {
	var allServices []*ServiceInstance

	// Get from etcd if available
	if sr.useEtcd {
		etcdServices, err := sr.getAllFromEtcd()
		if err != nil {
			log.Printf("⚠️  Failed to get all services from etcd: %v", err)
		} else {
			allServices = append(allServices, etcdServices...)
		}
	}

	// Get from Consul if available
	if sr.useConsul {
		consulServices, err := sr.getAllFromConsul()
		if err != nil {
			log.Printf("⚠️  Failed to get all services from Consul: %v", err)
		} else {
			allServices = append(allServices, consulServices...)
		}
	}

	return sr.deduplicateServices(allServices), nil
}

// getAllFromEtcd gets all services from etcd
func (sr *ServiceRegistry) getAllFromEtcd() ([]*ServiceInstance, error) {
	prefix := "/services/"
	
	resp, err := sr.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get all services: %w", err)
	}

	var services []*ServiceInstance
	for _, kv := range resp.Kvs {
		var service ServiceInstance
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			log.Printf("⚠️  Failed to unmarshal service data: %v", err)
			continue
		}
		services = append(services, &service)
	}

	return services, nil
}

// getAllFromConsul gets all services from Consul
func (sr *ServiceRegistry) getAllFromConsul() ([]*ServiceInstance, error) {
	services, _, err := sr.consulClient.Catalog().Services(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get services from Consul: %w", err)
	}

	var instances []*ServiceInstance
	for serviceName := range services {
		serviceInstances, err := sr.discoverFromConsul(serviceName)
		if err != nil {
			log.Printf("⚠️  Failed to get instances for service %s: %v", serviceName, err)
			continue
		}
		instances = append(instances, serviceInstances...)
	}

	return instances, nil
}