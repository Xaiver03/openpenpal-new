// Package pool provides intelligent database connection pool management
package pool

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// SmartConnectionPool implements an intelligent connection pool with AI-driven optimization
type SmartConnectionPool struct {
	config *core.ConnectionPoolConfig
	
	// Pool configuration
	minSize     int
	maxSize     int
	currentSize int
	
	// Connection management
	connections  chan *PooledConnection
	activeConns  map[string]*PooledConnection
	metrics      *PoolMetrics
	healthChecker *HealthChecker
	loadPredictor *LoadPredictor
	
	// Database configuration
	dbConfig     *core.DatabaseConfig
	driverName   string
	
	// State management
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// PooledConnection represents a connection in the pool
type PooledConnection struct {
	ID           string
	DB           *sql.DB
	CreatedAt    time.Time
	LastUsed     time.Time
	InUse        bool
	FailureCount int
	Healthy      bool
	Metadata     map[string]interface{}
	mu           sync.Mutex
}

// PoolMetrics tracks pool performance metrics
type PoolMetrics struct {
	mu                    sync.RWMutex
	ActiveConnections     int
	IdleConnections       int
	TotalConnections      int
	WaitingRequests       int
	ConnectionsCreated    int64
	ConnectionsClosed     int64
	FailedConnections     int64
	TotalRequests         int64
	AverageWaitTime       float64
	PeakConnections       int
	LastOptimization      time.Time
	PerformanceScore      float64
}

// NewSmartConnectionPool creates a new smart connection pool
func NewSmartConnectionPool(config *core.ConnectionPoolConfig, dbConfig *core.DatabaseConfig) (*SmartConnectionPool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &SmartConnectionPool{
		config:        config,
		minSize:       config.DefaultMinSize,
		maxSize:       config.DefaultMaxSize,
		connections:   make(chan *PooledConnection, config.DefaultMaxSize),
		activeConns:   make(map[string]*PooledConnection),
		dbConfig:      dbConfig,
		driverName:    getDriverName(dbConfig.Type),
		ctx:           ctx,
		cancel:        cancel,
		metrics:       NewPoolMetrics(),
		healthChecker: NewHealthChecker(config),
		loadPredictor: NewLoadPredictor(),
	}
	
	// Initialize with minimum connections
	if err := pool.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}
	
	// Start background processes
	go pool.healthCheckLoop()
	go pool.optimizationLoop()
	go pool.metricsCollectionLoop()
	
	return pool, nil
}

// GetConnection retrieves a connection from the pool
func (p *SmartConnectionPool) GetConnection(ctx context.Context) (*PooledConnection, error) {
	startTime := time.Now()
	p.metrics.recordRequest()
	
	// Try to get a connection from the pool
	select {
	case conn := <-p.connections:
		// Update connection usage
		conn.mu.Lock()
		conn.InUse = true
		conn.LastUsed = time.Now()
		conn.mu.Unlock()
		
		// Add to active connections
		p.mu.Lock()
		p.activeConns[conn.ID] = conn
		p.mu.Unlock()
		
		p.metrics.recordSuccessfulConnection(time.Since(startTime))
		return conn, nil
		
	case <-time.After(p.config.ConnectionTimeout):
		p.metrics.recordFailedConnection()
		return nil, fmt.Errorf("connection timeout after %v", p.config.ConnectionTimeout)
		
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ReleaseConnection returns a connection to the pool
func (p *SmartConnectionPool) ReleaseConnection(conn *PooledConnection) error {
	if conn == nil {
		return fmt.Errorf("cannot release nil connection")
	}
	
	conn.mu.Lock()
	conn.InUse = false
	conn.LastUsed = time.Now()
	conn.mu.Unlock()
	
	// Remove from active connections
	p.mu.Lock()
	delete(p.activeConns, conn.ID)
	p.mu.Unlock()
	
	// Check connection health before returning to pool
	if !p.healthChecker.CheckConnection(conn) {
		return p.closeConnection(conn)
	}
	
	// Return to pool
	select {
	case p.connections <- conn:
		return nil
	default:
		// Pool is full, close the connection
		return p.closeConnection(conn)
	}
}

// GetStats returns current pool statistics
func (p *SmartConnectionPool) GetStats() *core.PoolStats {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	
	return &core.PoolStats{
		ActiveConnections:   p.metrics.ActiveConnections,
		IdleConnections:     p.metrics.IdleConnections,
		TotalConnections:    p.metrics.TotalConnections,
		WaitingRequests:     p.metrics.WaitingRequests,
		AverageWaitTime:     p.metrics.AverageWaitTime,
		ConnectionsCreated:  p.metrics.ConnectionsCreated,
		ConnectionsClosed:   p.metrics.ConnectionsClosed,
		FailedConnections:   p.metrics.FailedConnections,
	}
}

// ResizePool dynamically resizes the connection pool
func (p *SmartConnectionPool) ResizePool(minSize, maxSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if minSize < 0 || maxSize < minSize {
		return fmt.Errorf("invalid pool size: min=%d, max=%d", minSize, maxSize)
	}
	
	oldMinSize := p.minSize
	oldMaxSize := p.maxSize
	
	p.minSize = minSize
	p.maxSize = maxSize
	
	log.Printf("📊 Resizing connection pool: %d-%d -> %d-%d", 
		oldMinSize, oldMaxSize, minSize, maxSize)
	
	// Adjust current connections
	return p.adjustPoolSize()
}

// HealthCheck performs health check on all connections
func (p *SmartConnectionPool) HealthCheck(ctx context.Context) (map[string]*core.ConnectionHealth, error) {
	p.mu.RLock()
	connections := make([]*PooledConnection, 0, len(p.activeConns))
	for _, conn := range p.activeConns {
		connections = append(connections, conn)
	}
	p.mu.RUnlock()
	
	// Add idle connections
	idleConns := p.getIdleConnections()
	connections = append(connections, idleConns...)
	
	health := make(map[string]*core.ConnectionHealth)
	
	for _, conn := range connections {
		connHealth := p.healthChecker.GetConnectionHealth(conn)
		health[conn.ID] = connHealth
	}
	
	return health, nil
}

// Private methods

func (p *SmartConnectionPool) initialize() error {
	log.Printf("📊 Initializing connection pool with %d-%d connections", p.minSize, p.maxSize)
	
	// Create minimum number of connections
	for i := 0; i < p.minSize; i++ {
		conn, err := p.createConnection()
		if err != nil {
			return fmt.Errorf("failed to create initial connection %d: %w", i, err)
		}
		
		p.connections <- conn
	}
	
	p.currentSize = p.minSize
	p.running = true
	
	log.Printf("✅ Connection pool initialized with %d connections", p.minSize)
	return nil
}

func (p *SmartConnectionPool) createConnection() (*PooledConnection, error) {
	db, err := sql.Open(p.driverName, p.dbConfig.ConnectionString)
	if err != nil {
		p.metrics.recordFailedConnection()
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	
	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		p.metrics.recordFailedConnection()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Configure connection
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(p.config.MaxLifetime)
	db.SetConnMaxIdleTime(p.config.IdleTimeout)
	
	conn := &PooledConnection{
		ID:        generateConnectionID(),
		DB:        db,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		InUse:     false,
		Healthy:   true,
		Metadata:  make(map[string]interface{}),
	}
	
	p.metrics.recordConnectionCreated()
	return conn, nil
}

func (p *SmartConnectionPool) closeConnection(conn *PooledConnection) error {
	if conn == nil || conn.DB == nil {
		return nil
	}
	
	err := conn.DB.Close()
	p.metrics.recordConnectionClosed()
	p.currentSize--
	
	log.Printf("🔒 Closed connection %s", conn.ID)
	return err
}

func (p *SmartConnectionPool) healthCheckLoop() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

func (p *SmartConnectionPool) optimizationLoop() {
	if !p.config.EnableAdaptiveSizing {
		return
	}
	
	ticker := time.NewTicker(p.config.AdaptiveCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.optimizePoolSize()
		}
	}
}

func (p *SmartConnectionPool) metricsCollectionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
		}
	}
}

func (p *SmartConnectionPool) performHealthCheck() {
	log.Println("🏥 Performing connection health check")
	
	// Check idle connections
	idleConns := p.getIdleConnections()
	for _, conn := range idleConns {
		if !p.healthChecker.CheckConnection(conn) {
			p.closeConnection(conn)
		}
	}
	
	// Check active connections
	p.mu.RLock()
	activeConns := make([]*PooledConnection, 0, len(p.activeConns))
	for _, conn := range p.activeConns {
		activeConns = append(activeConns, conn)
	}
	p.mu.RUnlock()
	
	for _, conn := range activeConns {
		if !conn.InUse && !p.healthChecker.CheckConnection(conn) {
			p.closeConnection(conn)
		}
	}
}

func (p *SmartConnectionPool) optimizePoolSize() {
	log.Println("🧠 Optimizing connection pool size")
	
	// Get current metrics
	stats := p.GetStats()
	
	// Predict optimal pool size using AI
	optimalSize := p.loadPredictor.PredictOptimalPoolSize(stats)
	
	// Apply optimization
	if optimalSize > p.maxSize {
		optimalSize = p.maxSize
	}
	if optimalSize < p.minSize {
		optimalSize = p.minSize
	}
	
	if optimalSize != p.currentSize {
		log.Printf("📊 Pool size optimization: %d -> %d", p.currentSize, optimalSize)
		p.adjustPoolSizeToTarget(optimalSize)
	}
}

func (p *SmartConnectionPool) adjustPoolSize() error {
	// Add connections if needed
	if p.currentSize < p.minSize {
		needed := p.minSize - p.currentSize
		for i := 0; i < needed; i++ {
			conn, err := p.createConnection()
			if err != nil {
				return err
			}
			p.connections <- conn
			p.currentSize++
		}
	}
	
	// Remove excess connections if needed
	if p.currentSize > p.maxSize {
		excess := p.currentSize - p.maxSize
		for i := 0; i < excess; i++ {
			select {
			case conn := <-p.connections:
				p.closeConnection(conn)
			default:
				// No idle connections to remove
				break
			}
		}
	}
	
	return nil
}

func (p *SmartConnectionPool) adjustPoolSizeToTarget(target int) {
	if target > p.currentSize {
		// Add connections
		needed := target - p.currentSize
		for i := 0; i < needed && p.currentSize < p.maxSize; i++ {
			conn, err := p.createConnection()
			if err != nil {
				log.Printf("⚠️  Failed to create connection: %v", err)
				break
			}
			p.connections <- conn
			p.currentSize++
		}
	} else if target < p.currentSize {
		// Remove connections
		excess := p.currentSize - target
		for i := 0; i < excess; i++ {
			select {
			case conn := <-p.connections:
				p.closeConnection(conn)
			default:
				return
			}
		}
	}
}

func (p *SmartConnectionPool) getIdleConnections() []*PooledConnection {
	var idleConns []*PooledConnection
	
	// This is a simplified version - in practice, you'd need a way to peek into the channel
	// or maintain a separate tracking mechanism for idle connections
	
	return idleConns
}

func (p *SmartConnectionPool) updateMetrics() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()
	
	p.metrics.ActiveConnections = len(p.activeConns)
	p.metrics.IdleConnections = len(p.connections)
	p.metrics.TotalConnections = p.currentSize
	
	// Calculate performance score
	p.metrics.PerformanceScore = p.calculatePerformanceScore()
}

func (p *SmartConnectionPool) calculatePerformanceScore() float64 {
	// Calculate performance score based on various metrics
	// This is a simplified version
	
	if p.metrics.TotalRequests == 0 {
		return 1.0
	}
	
	successRate := float64(p.metrics.TotalRequests-p.metrics.FailedConnections) / float64(p.metrics.TotalRequests)
	utilizationRate := float64(p.metrics.ActiveConnections) / float64(p.maxSize)
	
	// Higher is better for success rate, optimal utilization is around 70%
	score := successRate * 0.7
	if utilizationRate <= 0.7 {
		score += utilizationRate * 0.3
	} else {
		score += (1.4 - utilizationRate) * 0.3
	}
	
	return score
}

// Helper functions

func getDriverName(dbType core.DatabaseType) string {
	switch dbType {
	case core.DatabaseTypePostgreSQL:
		return "postgres"
	case core.DatabaseTypeMySQL:
		return "mysql"
	case core.DatabaseTypeSQLite:
		return "sqlite3"
	default:
		return "postgres"
	}
}

func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// NewPoolMetrics creates a new pool metrics instance
func NewPoolMetrics() *PoolMetrics {
	return &PoolMetrics{
		LastOptimization: time.Now(),
		PerformanceScore: 1.0,
	}
}

func (pm *PoolMetrics) recordRequest() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.TotalRequests++
}

func (pm *PoolMetrics) recordSuccessfulConnection(waitTime time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Update average wait time
	totalWaitTime := pm.AverageWaitTime * float64(pm.TotalRequests-1)
	pm.AverageWaitTime = (totalWaitTime + waitTime.Seconds()) / float64(pm.TotalRequests)
}

func (pm *PoolMetrics) recordFailedConnection() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.FailedConnections++
}

func (pm *PoolMetrics) recordConnectionCreated() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ConnectionsCreated++
}

func (pm *PoolMetrics) recordConnectionClosed() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ConnectionsClosed++
}