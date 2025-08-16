// Package migration provides Git-like database migration management with zero-downtime capabilities
package migration

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// ZeroDowntimeMigrationManager provides Git-like database migration management
type ZeroDowntimeMigrationManager struct {
	config           *core.MigrationConfig
	db               *sql.DB
	lockManager      *MigrationLockManager
	versionControl   *VersionControlManager
	riskAssessment   *RiskAssessmentEngine
	executionPlanner *ExecutionPlanner
	
	// State management
	mu              sync.RWMutex
	migrationHistory map[string]*core.MigrationHistory
	currentVersion   string
}

// MigrationLockManager handles distributed locking for migrations
type MigrationLockManager struct {
	db          *sql.DB
	lockTimeout time.Duration
	lockID      string
}

// VersionControlManager provides Git-like version control for migrations
type VersionControlManager struct {
	migrationsPath string
	branches       map[string]*MigrationBranch
	currentBranch  string
}

// RiskAssessmentEngine analyzes migration risks
type RiskAssessmentEngine struct {
	rules map[string]RiskRule
}

// ExecutionPlanner creates optimal execution plans for migrations
type ExecutionPlanner struct {
	config *core.MigrationConfig
}

// NewZeroDowntimeMigrationManager creates a new migration manager
func NewZeroDowntimeMigrationManager(config *core.MigrationConfig, db *sql.DB) (*ZeroDowntimeMigrationManager, error) {
	manager := &ZeroDowntimeMigrationManager{
		config:           config,
		db:              db,
		lockManager:     NewMigrationLockManager(db, config.LockTimeout),
		versionControl:  NewVersionControlManager(config.MigrationsPath),
		riskAssessment:  NewRiskAssessmentEngine(),
		executionPlanner: NewExecutionPlanner(config),
		migrationHistory: make(map[string]*core.MigrationHistory),
	}
	
	// Initialize migration tracking table
	if err := manager.initializeMigrationTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize migration table: %w", err)
	}
	
	// Load migration history
	if err := manager.loadMigrationHistory(); err != nil {
		return nil, fmt.Errorf("failed to load migration history: %w", err)
	}
	
	return manager, nil
}

// ApplyMigration applies a migration with zero-downtime strategy
func (m *ZeroDowntimeMigrationManager) ApplyMigration(ctx context.Context, migration *core.Migration) error {
	log.Printf("🔄 Starting migration %s: %s", migration.Version, migration.Name)
	
	// Acquire migration lock
	if err := m.lockManager.AcquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer m.lockManager.ReleaseLock()
	
	// Validate migration
	if err := m.validateMigration(migration); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}
	
	// Assess risks
	riskLevel, risks := m.riskAssessment.AssessRisks(migration)
	log.Printf("📊 Migration risk level: %s", riskLevel)
	for _, risk := range risks {
		log.Printf("  ⚠️  %s", risk)
	}
	
	// Create execution plan
	plan, err := m.executionPlanner.CreatePlan(migration)
	if err != nil {
		return fmt.Errorf("failed to create execution plan: %w", err)
	}
	
	log.Printf("📋 Execution plan created with %d steps", len(plan.Steps))
	
	// Execute migration
	return m.executeMigration(ctx, migration, plan)
}

// RollbackMigration rollbacks a migration to a specific version
func (m *ZeroDowntimeMigrationManager) RollbackMigration(ctx context.Context, version string) error {
	log.Printf("🔙 Starting rollback to version %s", version)
	
	// Acquire migration lock
	if err := m.lockManager.AcquireLock(ctx); err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer m.lockManager.ReleaseLock()
	
	// Find migrations to rollback
	toRollback := m.findMigrationsToRollback(version)
	
	// Execute rollbacks in reverse order
	for i := len(toRollback) - 1; i >= 0; i-- {
		migration := toRollback[i]
		log.Printf("🔄 Rolling back migration %s", migration.Version)
		
		if err := m.executeRollback(ctx, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}
	
	log.Printf("✅ Rollback to version %s completed", version)
	return nil
}

// GetMigrationHistory returns the migration history
func (m *ZeroDowntimeMigrationManager) GetMigrationHistory(ctx context.Context) ([]*core.MigrationHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var history []*core.MigrationHistory
	for _, h := range m.migrationHistory {
		history = append(history, h)
	}
	
	// Sort by applied date
	sort.Slice(history, func(i, j int) bool {
		return history[i].AppliedAt.Before(history[j].AppliedAt)
	})
	
	return history, nil
}

// PlanMigration creates a zero-downtime migration plan
func (m *ZeroDowntimeMigrationManager) PlanMigration(ctx context.Context, migration *core.Migration) (*core.MigrationPlan, error) {
	log.Printf("📋 Planning migration %s: %s", migration.Version, migration.Name)
	
	// Assess risks
	riskLevel, risks := m.riskAssessment.AssessRisks(migration)
	
	// Create execution plan
	plan, err := m.executionPlanner.CreatePlan(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}
	
	// Calculate estimated time
	estimatedTime := m.calculateEstimatedTime(plan)
	
	migrationPlan := &core.MigrationPlan{
		Steps:            plan.Steps,
		EstimatedTime:    estimatedTime,
		RollbackPlan:     m.generateRollbackPlan(migration),
		RiskAssessment:   fmt.Sprintf("Risk Level: %s\nRisks: %s", riskLevel, strings.Join(risks, ", ")),
		RequiresDowntime: m.requiresDowntime(migration),
	}
	
	return migrationPlan, nil
}

// CreateBranch creates a new migration branch (Git-like)
func (m *ZeroDowntimeMigrationManager) CreateBranch(branchName string) error {
	return m.versionControl.CreateBranch(branchName)
}

// SwitchBranch switches to a different migration branch
func (m *ZeroDowntimeMigrationManager) SwitchBranch(branchName string) error {
	return m.versionControl.SwitchBranch(branchName)
}

// MergeBranch merges a migration branch
func (m *ZeroDowntimeMigrationManager) MergeBranch(sourceBranch, targetBranch string) error {
	return m.versionControl.MergeBranch(sourceBranch, targetBranch)
}

// Private methods

func (m *ZeroDowntimeMigrationManager) initializeMigrationTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		duration_ms BIGINT NOT NULL,
		success BOOLEAN NOT NULL,
		error_message TEXT,
		checksum VARCHAR(32) NOT NULL,
		created_by VARCHAR(100) DEFAULT 'system'
	)`
	
	_, err := m.db.Exec(createTableSQL)
	return err
}

func (m *ZeroDowntimeMigrationManager) loadMigrationHistory() error {
	rows, err := m.db.Query(`
		SELECT version, name, applied_at, duration_ms, success, error_message 
		FROM schema_migrations 
		ORDER BY applied_at DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for rows.Next() {
		var h core.MigrationHistory
		var durationMs int64
		var errorMessage sql.NullString
		
		err := rows.Scan(&h.Version, &h.Name, &h.AppliedAt, &durationMs, &h.Success, &errorMessage)
		if err != nil {
			return err
		}
		
		h.Duration = time.Duration(durationMs) * time.Millisecond
		if errorMessage.Valid {
			h.ErrorMessage = errorMessage.String
		}
		
		m.migrationHistory[h.Version] = &h
		
		// Track current version (latest successful migration)
		if h.Success && (m.currentVersion == "" || h.AppliedAt.After(m.migrationHistory[m.currentVersion].AppliedAt)) {
			m.currentVersion = h.Version
		}
	}
	
	return rows.Err()
}

func (m *ZeroDowntimeMigrationManager) validateMigration(migration *core.Migration) error {
	// Check if migration already applied
	if _, exists := m.migrationHistory[migration.Version]; exists {
		return fmt.Errorf("migration %s already applied", migration.Version)
	}
	
	// Validate checksum
	expectedChecksum := m.calculateChecksum(migration.UpScript)
	if migration.Checksum != expectedChecksum {
		return fmt.Errorf("migration checksum mismatch: expected %s, got %s", expectedChecksum, migration.Checksum)
	}
	
	// Validate SQL syntax (basic check)
	if err := m.validateSQL(migration.UpScript); err != nil {
		return fmt.Errorf("invalid SQL in up script: %w", err)
	}
	
	if err := m.validateSQL(migration.DownScript); err != nil {
		return fmt.Errorf("invalid SQL in down script: %w", err)
	}
	
	return nil
}

func (m *ZeroDowntimeMigrationManager) executeMigration(ctx context.Context, migration *core.Migration, plan *ExecutionPlan) error {
	startTime := time.Now()
	
	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	success := false
	defer func() {
		if success {
			tx.Commit()
			m.recordMigrationHistory(migration, time.Since(startTime), true, "")
		} else {
			tx.Rollback()
			m.recordMigrationHistory(migration, time.Since(startTime), false, "Migration failed")
		}
	}()
	
	// Execute migration steps
	for i, step := range plan.Steps {
		log.Printf("🔄 Executing step %d/%d: %s", i+1, len(plan.Steps), step.Name)
		
		stepStart := time.Now()
		
		if err := m.executeStep(ctx, tx, step); err != nil {
			return fmt.Errorf("failed to execute step %d (%s): %w", i+1, step.Name, err)
		}
		
		stepDuration := time.Since(stepStart)
		log.Printf("✅ Step %d completed in %v", i+1, stepDuration)
	}
	
	success = true
	log.Printf("✅ Migration %s completed successfully in %v", migration.Version, time.Since(startTime))
	
	return nil
}

func (m *ZeroDowntimeMigrationManager) executeRollback(ctx context.Context, migration *core.Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Execute down script
	if _, err := tx.ExecContext(ctx, migration.DownScript); err != nil {
		return fmt.Errorf("failed to execute rollback script: %w", err)
	}
	
	// Remove from migration history
	if _, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}
	
	// Update in-memory state
	m.mu.Lock()
	delete(m.migrationHistory, migration.Version)
	m.mu.Unlock()
	
	return nil
}

func (m *ZeroDowntimeMigrationManager) executeStep(ctx context.Context, tx *sql.Tx, step *core.MigrationStep) error {
	if step.SQL == "" {
		return nil
	}
	
	_, err := tx.ExecContext(ctx, step.SQL)
	return err
}

func (m *ZeroDowntimeMigrationManager) recordMigrationHistory(migration *core.Migration, duration time.Duration, success bool, errorMsg string) {
	durationMs := duration.Milliseconds()
	
	// Record in database
	_, err := m.db.Exec(`
		INSERT INTO schema_migrations (version, name, duration_ms, success, error_message, checksum)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, migration.Version, migration.Name, durationMs, success, errorMsg, migration.Checksum)
	
	if err != nil {
		log.Printf("⚠️  Failed to record migration history: %v", err)
		return
	}
	
	// Update in-memory state
	m.mu.Lock()
	defer m.mu.Unlock()
	
	history := &core.MigrationHistory{
		Version:      migration.Version,
		Name:         migration.Name,
		AppliedAt:    time.Now(),
		Duration:     duration,
		Success:      success,
		ErrorMessage: errorMsg,
	}
	
	m.migrationHistory[migration.Version] = history
	
	if success {
		m.currentVersion = migration.Version
	}
}

func (m *ZeroDowntimeMigrationManager) findMigrationsToRollback(targetVersion string) []*core.Migration {
	// This is a simplified implementation
	// In practice, you'd load migrations from files and determine which to rollback
	var toRollback []*core.Migration
	
	// Find all migrations applied after the target version
	for version, history := range m.migrationHistory {
		if history.Success && version > targetVersion {
			// Load migration from file
			migration, err := m.loadMigrationFromFile(version)
			if err != nil {
				log.Printf("⚠️  Failed to load migration %s for rollback: %v", version, err)
				continue
			}
			toRollback = append(toRollback, migration)
		}
	}
	
	// Sort by version (descending for rollback)
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].Version > toRollback[j].Version
	})
	
	return toRollback
}

func (m *ZeroDowntimeMigrationManager) loadMigrationFromFile(version string) (*core.Migration, error) {
	// Load migration file
	migrationFile := filepath.Join(m.config.MigrationsPath, fmt.Sprintf("%s.sql", version))
	
	content, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		return nil, err
	}
	
	// Parse migration file (simplified)
	parts := strings.Split(string(content), "-- DOWN")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid migration file format")
	}
	
	upScript := strings.TrimSpace(strings.TrimPrefix(parts[0], "-- UP"))
	downScript := strings.TrimSpace(parts[1])
	
	return &core.Migration{
		Version:    version,
		Name:       fmt.Sprintf("Migration %s", version),
		UpScript:   upScript,
		DownScript: downScript,
		Checksum:   m.calculateChecksum(upScript),
		CreatedAt:  time.Now(),
	}, nil
}

func (m *ZeroDowntimeMigrationManager) calculateChecksum(content string) string {
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (m *ZeroDowntimeMigrationManager) validateSQL(sql string) error {
	// Basic SQL validation
	if strings.TrimSpace(sql) == "" {
		return fmt.Errorf("empty SQL script")
	}
	
	// Check for dangerous operations in production
	if m.config.ValidationMode == "strict" {
		dangerous := []string{"DROP DATABASE", "TRUNCATE", "DELETE FROM"}
		sqlUpper := strings.ToUpper(sql)
		
		for _, danger := range dangerous {
			if strings.Contains(sqlUpper, danger) {
				return fmt.Errorf("dangerous operation detected: %s", danger)
			}
		}
	}
	
	return nil
}

func (m *ZeroDowntimeMigrationManager) calculateEstimatedTime(plan *ExecutionPlan) time.Duration {
	total := time.Duration(0)
	for _, step := range plan.Steps {
		total += step.Duration
	}
	return total
}

func (m *ZeroDowntimeMigrationManager) generateRollbackPlan(migration *core.Migration) string {
	return fmt.Sprintf("Execute down script: %s", migration.DownScript)
}

func (m *ZeroDowntimeMigrationManager) requiresDowntime(migration *core.Migration) bool {
	// Analyze migration to determine if downtime is required
	sql := strings.ToUpper(migration.UpScript)
	
	// Operations that typically require downtime
	downtimeOperations := []string{
		"ALTER TABLE", "DROP TABLE", "ADD CONSTRAINT", "DROP CONSTRAINT",
	}
	
	for _, op := range downtimeOperations {
		if strings.Contains(sql, op) {
			return true
		}
	}
	
	return false
}

// Data structures and helper types

type MigrationBranch struct {
	Name        string
	Migrations  []string
	BaseVersion string
	CreatedAt   time.Time
}

type RiskRule struct {
	Pattern     string
	RiskLevel   string
	Description string
}

type ExecutionPlan struct {
	Steps []*core.MigrationStep
}

// Constructor functions

func NewMigrationLockManager(db *sql.DB, timeout time.Duration) *MigrationLockManager {
	return &MigrationLockManager{
		db:          db,
		lockTimeout: timeout,
		lockID:      fmt.Sprintf("migration_lock_%d", time.Now().UnixNano()),
	}
}

func NewVersionControlManager(migrationsPath string) *VersionControlManager {
	return &VersionControlManager{
		migrationsPath: migrationsPath,
		branches:      make(map[string]*MigrationBranch),
		currentBranch: "main",
	}
}

func NewRiskAssessmentEngine() *RiskAssessmentEngine {
	rules := map[string]RiskRule{
		"drop_table": {
			Pattern:     "DROP TABLE",
			RiskLevel:   "HIGH",
			Description: "Dropping tables can cause data loss",
		},
		"alter_table": {
			Pattern:     "ALTER TABLE",
			RiskLevel:   "MEDIUM",
			Description: "Table alterations may require table locks",
		},
		"add_index": {
			Pattern:     "CREATE INDEX",
			RiskLevel:   "LOW",
			Description: "Index creation is generally safe",
		},
	}
	
	return &RiskAssessmentEngine{rules: rules}
}

func NewExecutionPlanner(config *core.MigrationConfig) *ExecutionPlanner {
	return &ExecutionPlanner{config: config}
}

// MigrationLockManager methods

func (mlm *MigrationLockManager) AcquireLock(ctx context.Context) error {
	// Simple lock implementation using database
	query := `
		INSERT INTO migration_locks (lock_id, acquired_at, expires_at)
		VALUES ($1, NOW(), NOW() + INTERVAL '%d seconds')
		ON CONFLICT (lock_id) DO NOTHING
	`
	
	result, err := mlm.db.ExecContext(ctx, fmt.Sprintf(query, int(mlm.lockTimeout.Seconds())), mlm.lockID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("failed to acquire migration lock - another migration in progress")
	}
	
	return nil
}

func (mlm *MigrationLockManager) ReleaseLock() error {
	_, err := mlm.db.Exec("DELETE FROM migration_locks WHERE lock_id = $1", mlm.lockID)
	return err
}

// RiskAssessmentEngine methods

func (rae *RiskAssessmentEngine) AssessRisks(migration *core.Migration) (string, []string) {
	var risks []string
	maxRiskLevel := "LOW"
	
	sql := strings.ToUpper(migration.UpScript)
	
	for _, rule := range rae.rules {
		if strings.Contains(sql, rule.Pattern) {
			risks = append(risks, rule.Description)
			
			if rule.RiskLevel == "HIGH" {
				maxRiskLevel = "HIGH"
			} else if rule.RiskLevel == "MEDIUM" && maxRiskLevel != "HIGH" {
				maxRiskLevel = "MEDIUM"
			}
		}
	}
	
	if len(risks) == 0 {
		risks = append(risks, "No specific risks identified")
	}
	
	return maxRiskLevel, risks
}

// ExecutionPlanner methods

func (ep *ExecutionPlanner) CreatePlan(migration *core.Migration) (*ExecutionPlan, error) {
	// Split migration into logical steps
	statements := ep.parseStatements(migration.UpScript)
	
	var steps []*core.MigrationStep
	
	for i, stmt := range statements {
		step := &core.MigrationStep{
			Order:      i + 1,
			Name:       fmt.Sprintf("Execute statement %d", i+1),
			SQL:        stmt,
			Duration:   ep.estimateStepDuration(stmt),
			Reversible: ep.isReversible(stmt),
		}
		steps = append(steps, step)
	}
	
	return &ExecutionPlan{Steps: steps}, nil
}

func (ep *ExecutionPlanner) parseStatements(sql string) []string {
	// Simple statement splitting - in practice, use a proper SQL parser
	statements := strings.Split(sql, ";")
	
	var result []string
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}
	
	return result
}

func (ep *ExecutionPlanner) estimateStepDuration(sql string) time.Duration {
	// Simple duration estimation based on statement type
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
	
	if strings.HasPrefix(sqlUpper, "CREATE INDEX") {
		return 30 * time.Second
	} else if strings.HasPrefix(sqlUpper, "ALTER TABLE") {
		return 10 * time.Second
	} else {
		return 1 * time.Second
	}
}

func (ep *ExecutionPlanner) isReversible(sql string) bool {
	// Determine if a statement is reversible
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
	
	nonReversible := []string{"DROP", "DELETE", "TRUNCATE"}
	
	for _, nr := range nonReversible {
		if strings.HasPrefix(sqlUpper, nr) {
			return false
		}
	}
	
	return true
}

// VersionControlManager methods

func (vcm *VersionControlManager) CreateBranch(branchName string) error {
	if _, exists := vcm.branches[branchName]; exists {
		return fmt.Errorf("branch %s already exists", branchName)
	}
	
	vcm.branches[branchName] = &MigrationBranch{
		Name:        branchName,
		Migrations:  make([]string, 0),
		BaseVersion: vcm.getCurrentVersion(),
		CreatedAt:   time.Now(),
	}
	
	return nil
}

func (vcm *VersionControlManager) SwitchBranch(branchName string) error {
	if _, exists := vcm.branches[branchName]; !exists {
		return fmt.Errorf("branch %s does not exist", branchName)
	}
	
	vcm.currentBranch = branchName
	return nil
}

func (vcm *VersionControlManager) MergeBranch(sourceBranch, targetBranch string) error {
	source, exists := vcm.branches[sourceBranch]
	if !exists {
		return fmt.Errorf("source branch %s does not exist", sourceBranch)
	}
	
	target, exists := vcm.branches[targetBranch]
	if !exists {
		return fmt.Errorf("target branch %s does not exist", targetBranch)
	}
	
	// Simple merge - append migrations from source to target
	target.Migrations = append(target.Migrations, source.Migrations...)
	
	return nil
}

func (vcm *VersionControlManager) getCurrentVersion() string {
	// Get current version from migration files
	files, err := ioutil.ReadDir(vcm.migrationsPath)
	if err != nil {
		return "000000"
	}
	
	var versions []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			version := strings.TrimSuffix(file.Name(), ".sql")
			versions = append(versions, version)
		}
	}
	
	if len(versions) == 0 {
		return "000000"
	}
	
	sort.Strings(versions)
	return versions[len(versions)-1]
}