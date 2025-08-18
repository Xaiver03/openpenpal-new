/**
 * 迁移协调器 - 整合统一迁移策略与共享包
 * 提供跨服务的迁移协调和状态管理
 */

package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MigrationCoordinator 迁移协调器
type MigrationCoordinator struct {
	services          map[string]*ServiceMigration
	sharedPackageDB   *gorm.DB
	backendDB         *gorm.DB
	migrationStrategy *MigrationStrategy
	mutex             sync.RWMutex
	status            CoordinatorStatus
}

// ServiceMigration 服务迁移信息
type ServiceMigration struct {
	ServiceName   string            `json:"service_name"`
	DatabaseType  string            `json:"database_type"`
	Status        MigrationStatus   `json:"status"`
	Progress      float64           `json:"progress"`
	LastUpdated   time.Time         `json:"last_updated"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	Dependencies  []string          `json:"dependencies"`
	MigrationLog  []MigrationStep   `json:"migration_log"`
}

// MigrationStatus 迁移状态
type MigrationStatus string

const (
	MigrationStatusPending    MigrationStatus = "pending"
	MigrationStatusRunning    MigrationStatus = "running"
	MigrationStatusCompleted  MigrationStatus = "completed"
	MigrationStatusFailed     MigrationStatus = "failed"
	MigrationStatusRolledBack MigrationStatus = "rolled_back"
)

// CoordinatorStatus 协调器状态
type CoordinatorStatus struct {
	OverallStatus     MigrationStatus `json:"overall_status"`
	CompletedServices int             `json:"completed_services"`
	TotalServices     int             `json:"total_services"`
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time,omitempty"`
	Duration          time.Duration   `json:"duration"`
	ProgressPercent   float64         `json:"progress_percent"`
}

// NewMigrationCoordinator 创建迁移协调器
func NewMigrationCoordinator(backendDB *gorm.DB, config *Config) *MigrationCoordinator {
	coordinator := &MigrationCoordinator{
		services:    make(map[string]*ServiceMigration),
		backendDB:   backendDB,
		mutex:       sync.RWMutex{},
		status: CoordinatorStatus{
			OverallStatus: MigrationStatusPending,
			StartTime:     time.Now(),
		},
	}

	// 创建统一迁移策略
	migrationOpts := &MigrationOptions{
		DryRun:              false,
		RollbackMode:        false,
		SkipSafeMigration:   false,
		SkipOptimizations:   false,
		ConcurrentIndexes:   true,
		Timeout:             30 * time.Minute,
		FailureStrategy:     "stop",
		BackupBeforeMigrate: true,
	}
	coordinator.migrationStrategy = NewMigrationStrategy(backendDB, config, migrationOpts)

	// 注册服务
	coordinator.registerServices()

	return coordinator
}

// registerServices 注册需要迁移的服务
func (mc *MigrationCoordinator) registerServices() {
	// 后端服务（主要数据库）
	mc.registerService("backend", []string{}, "Main backend database with all core models")

	// Courier服务
	mc.registerService("courier-service", []string{"backend"}, "Courier management service")

	// Write服务  
	mc.registerService("write-service", []string{"backend"}, "Letter writing service")

	// Gateway服务
	mc.registerService("gateway", []string{"backend"}, "API gateway service")

	// 共享包集成
	mc.registerService("shared-package", []string{"backend"}, "Shared database package integration")

	mc.status.TotalServices = len(mc.services)
}

// registerService 注册单个服务
func (mc *MigrationCoordinator) registerService(serviceName string, dependencies []string, description string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.services[serviceName] = &ServiceMigration{
		ServiceName:  serviceName,
		DatabaseType: "postgres",
		Status:       MigrationStatusPending,
		Progress:     0.0,
		LastUpdated:  time.Now(),
		Dependencies: dependencies,
		MigrationLog: make([]MigrationStep, 0),
	}

	log.Printf("📝 Registered service: %s (dependencies: %v)", serviceName, dependencies)
}

// ExecuteCoordinatedMigration 执行协调迁移
func (mc *MigrationCoordinator) ExecuteCoordinatedMigration() error {
	log.Println("🎯 Starting coordinated migration across all services")
	
	mc.mutex.Lock()
	mc.status.OverallStatus = MigrationStatusRunning
	mc.status.StartTime = time.Now()
	mc.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()

	// Phase 1: 执行依赖关系分析
	if err := mc.analyzeDependencies(); err != nil {
		return fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Phase 2: 执行预检查
	if err := mc.executePreChecks(ctx); err != nil {
		return fmt.Errorf("pre-checks failed: %w", err)
	}

	// Phase 3: 按依赖顺序执行迁移
	if err := mc.executeMigrationsByDependency(ctx); err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	// Phase 4: 后迁移验证
	if err := mc.executePostMigrationValidation(ctx); err != nil {
		return fmt.Errorf("post-migration validation failed: %w", err)
	}

	// Phase 5: 共享包集成
	if err := mc.integrateSharedPackage(ctx); err != nil {
		log.Printf("⚠️  Warning: Shared package integration failed: %v", err)
		// 不阻止整体迁移成功，但记录警告
	}

	mc.mutex.Lock()
	mc.status.OverallStatus = MigrationStatusCompleted
	mc.status.EndTime = time.Now()
	mc.status.Duration = mc.status.EndTime.Sub(mc.status.StartTime)
	mc.status.ProgressPercent = 100.0
	mc.mutex.Unlock()

	// 生成综合报告
	if err := mc.generateComprehensiveReport(); err != nil {
		log.Printf("⚠️  Warning: Failed to generate comprehensive report: %v", err)
	}

	log.Printf("🎉 Coordinated migration completed successfully in %v", mc.status.Duration)
	return nil
}

// analyzeDependencies 分析服务依赖关系
func (mc *MigrationCoordinator) analyzeDependencies() error {
	log.Println("🔍 Analyzing service dependencies...")

	// 验证依赖关系的有效性
	for serviceName, service := range mc.services {
		for _, dep := range service.Dependencies {
			if _, exists := mc.services[dep]; !exists {
				return fmt.Errorf("service %s has invalid dependency: %s", serviceName, dep)
			}
		}
	}

	// 检测循环依赖
	if err := mc.detectCircularDependencies(); err != nil {
		return fmt.Errorf("circular dependency detected: %w", err)
	}

	log.Println("✅ Dependency analysis completed")
	return nil
}

// detectCircularDependencies 检测循环依赖
func (mc *MigrationCoordinator) detectCircularDependencies() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for serviceName := range mc.services {
		if !visited[serviceName] {
			if mc.hasCycleDFS(serviceName, visited, recStack) {
				return fmt.Errorf("circular dependency involving service: %s", serviceName)
			}
		}
	}

	return nil
}

// hasCycleDFS DFS检测循环依赖
func (mc *MigrationCoordinator) hasCycleDFS(serviceName string, visited, recStack map[string]bool) bool {
	visited[serviceName] = true
	recStack[serviceName] = true

	service := mc.services[serviceName]
	for _, dep := range service.Dependencies {
		if !visited[dep] {
			if mc.hasCycleDFS(dep, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			return true
		}
	}

	recStack[serviceName] = false
	return false
}

// executePreChecks 执行预检查
func (mc *MigrationCoordinator) executePreChecks(ctx context.Context) error {
	log.Println("🔍 Executing pre-migration checks across all services...")

	// 检查数据库连接
	if err := mc.backendDB.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("backend database connection failed: %w", err)
	}

	// 检查PostgreSQL扩展
	extensions := []string{"pg_stat_statements", "btree_gist"}
	for _, ext := range extensions {
		var exists bool
		err := mc.backendDB.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = ?)", ext).Scan(&exists).Error
		if err != nil {
			log.Printf("⚠️  Warning: Could not check extension %s: %v", ext, err)
		} else if exists {
			log.Printf("✅ Extension %s is available", ext)
		} else {
			log.Printf("📝 Extension %s not installed (will be created if needed)", ext)
		}
	}

	// 检查磁盘空间
	var dbSize string
	if err := mc.backendDB.WithContext(ctx).Raw("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize).Error; err != nil {
		log.Printf("⚠️  Warning: Could not check database size: %v", err)
	} else {
		log.Printf("💾 Current database size: %s", dbSize)
	}

	log.Println("✅ Pre-checks completed")
	return nil
}

// executeMigrationsByDependency 按依赖关系执行迁移
func (mc *MigrationCoordinator) executeMigrationsByDependency(ctx context.Context) error {
	log.Println("🚀 Executing migrations in dependency order...")

	// 构建执行顺序
	executionOrder, err := mc.buildExecutionOrder()
	if err != nil {
		return fmt.Errorf("failed to build execution order: %w", err)
	}

	log.Printf("📋 Execution order: %v", executionOrder)

	// 按顺序执行迁移
	for i, serviceName := range executionOrder {
		log.Printf("📌 Migrating service %d/%d: %s", i+1, len(executionOrder), serviceName)

		if err := mc.migrateService(ctx, serviceName); err != nil {
			mc.updateServiceStatus(serviceName, MigrationStatusFailed, 0, err.Error())
			return fmt.Errorf("migration failed for service %s: %w", serviceName, err)
		}

		mc.updateServiceStatus(serviceName, MigrationStatusCompleted, 100, "")
		mc.updateOverallProgress()
	}

	return nil
}

// buildExecutionOrder 构建执行顺序（拓扑排序）
func (mc *MigrationCoordinator) buildExecutionOrder() ([]string, error) {
	var order []string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string) error
	dfs = func(serviceName string) error {
		if recStack[serviceName] {
			return fmt.Errorf("circular dependency detected at %s", serviceName)
		}
		if visited[serviceName] {
			return nil
		}

		visited[serviceName] = true
		recStack[serviceName] = true

		service := mc.services[serviceName]
		for _, dep := range service.Dependencies {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		recStack[serviceName] = false
		order = append(order, serviceName)
		return nil
	}

	for serviceName := range mc.services {
		if !visited[serviceName] {
			if err := dfs(serviceName); err != nil {
				return nil, err
			}
		}
	}

	return order, nil
}

// migrateService 迁移单个服务
func (mc *MigrationCoordinator) migrateService(ctx context.Context, serviceName string) error {
	log.Printf("🔧 Starting migration for service: %s", serviceName)

	mc.updateServiceStatus(serviceName, MigrationStatusRunning, 0, "")

	switch serviceName {
	case "backend":
		return mc.migrateBackendService(ctx)
	case "shared-package":
		return mc.migrateSharedPackage(ctx)
	default:
		// 其他服务的迁移逻辑
		return mc.migrateGenericService(ctx, serviceName)
	}
}

// migrateBackendService 迁移后端服务
func (mc *MigrationCoordinator) migrateBackendService(ctx context.Context) error {
	log.Println("🏗️  Migrating backend service using unified migration strategy...")

	// 使用统一迁移策略
	if err := mc.migrationStrategy.ExecuteUnifiedMigration(); err != nil {
		return fmt.Errorf("unified migration strategy failed: %w", err)
	}

	return nil
}

// migrateSharedPackage 迁移共享包
func (mc *MigrationCoordinator) migrateSharedPackage(ctx context.Context) error {
	log.Println("📦 Setting up shared package integration...")

	// 这里可以添加共享包的具体集成逻辑
	// 目前只做占位符实现
	log.Println("📝 Shared package integration placeholder")

	return nil
}

// migrateGenericService 通用服务迁移
func (mc *MigrationCoordinator) migrateGenericService(ctx context.Context, serviceName string) error {
	log.Printf("🔧 Migrating generic service: %s", serviceName)

	// 通用服务迁移逻辑
	// 大多数微服务只需要确保数据库连接配置正确
	log.Printf("✅ Service %s migration completed (configuration only)", serviceName)

	return nil
}

// executePostMigrationValidation 执行迁移后验证
func (mc *MigrationCoordinator) executePostMigrationValidation(ctx context.Context) error {
	log.Println("🔍 Executing post-migration validation...")

	// 验证关键表存在
	tables := []string{"users", "letters", "couriers", "courier_tasks", "letter_codes"}
	for _, table := range tables {
		var count int64
		err := mc.backendDB.WithContext(ctx).Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", table).Scan(&count).Error
		if err != nil {
			return fmt.Errorf("failed to validate table %s: %w", table, err)
		}
		if count == 0 {
			return fmt.Errorf("critical table %s not found", table)
		}
		log.Printf("✅ Table validated: %s", table)
	}

	// 验证索引
	var indexCount int64
	err := mc.backendDB.WithContext(ctx).Raw(`
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE schemaname = 'public' 
		AND indexname LIKE 'idx_%'
	`).Scan(&indexCount).Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not validate indexes: %v", err)
	} else {
		log.Printf("📊 Optimized indexes found: %d", indexCount)
	}

	log.Println("✅ Post-migration validation completed")
	return nil
}

// integrateSharedPackage 集成共享包
func (mc *MigrationCoordinator) integrateSharedPackage(ctx context.Context) error {
	log.Println("📦 Integrating shared package...")

	// TODO: 这里应该实现与共享包的实际集成
	// 目前共享包还在开发中，所以只做占位符
	log.Println("📝 Shared package integration is pending implementation")

	return nil
}

// updateServiceStatus 更新服务状态
func (mc *MigrationCoordinator) updateServiceStatus(serviceName string, status MigrationStatus, progress float64, errorMsg string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if service, exists := mc.services[serviceName]; exists {
		service.Status = status
		service.Progress = progress
		service.LastUpdated = time.Now()
		service.ErrorMessage = errorMsg
	}
}

// updateOverallProgress 更新整体进度
func (mc *MigrationCoordinator) updateOverallProgress() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	completed := 0
	for _, service := range mc.services {
		if service.Status == MigrationStatusCompleted {
			completed++
		}
	}

	mc.status.CompletedServices = completed
	mc.status.ProgressPercent = float64(completed) / float64(mc.status.TotalServices) * 100
}

// generateComprehensiveReport 生成综合报告
func (mc *MigrationCoordinator) generateComprehensiveReport() error {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	log.Printf(`
🎉 Comprehensive Migration Report
=====================================
Overall Status: %s
Completed Services: %d/%d
Total Duration: %v
Success Rate: %.2f%%

📊 Service Details:
`, mc.status.OverallStatus, mc.status.CompletedServices, mc.status.TotalServices, 
		mc.status.Duration, mc.status.ProgressPercent)

	for serviceName, service := range mc.services {
		statusIcon := "✅"
		if service.Status == MigrationStatusFailed {
			statusIcon = "❌"
		} else if service.Status != MigrationStatusCompleted {
			statusIcon = "⏳"
		}

		log.Printf("  %s %s: %s (%.1f%% complete)", statusIcon, serviceName, service.Status, service.Progress)
		if service.ErrorMessage != "" {
			log.Printf("    Error: %s", service.ErrorMessage)
		}
	}

	log.Println("\n🎯 Migration coordination completed successfully!")
	return nil
}

// GetStatus 获取协调器状态
func (mc *MigrationCoordinator) GetStatus() CoordinatorStatus {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return mc.status
}

// GetServiceStatus 获取特定服务状态
func (mc *MigrationCoordinator) GetServiceStatus(serviceName string) (*ServiceMigration, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	service, exists := mc.services[serviceName]
	return service, exists
}