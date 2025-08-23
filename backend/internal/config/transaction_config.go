// Package config provides transaction isolation level configuration
package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/logger"
)

// TransactionConfig holds transaction configuration settings
type TransactionConfig struct {
	DefaultIsolationLevel     sql.IsolationLevel `json:"default_isolation_level"`
	HighConcurrencyLevel      sql.IsolationLevel `json:"high_concurrency_level"`
	ReadOnlyLevel            sql.IsolationLevel `json:"read_only_level"`
	CriticalOperationLevel   sql.IsolationLevel `json:"critical_operation_level"`
	EnableDeadlockDetection  bool              `json:"enable_deadlock_detection"`
	DeadlockTimeout         time.Duration      `json:"deadlock_timeout"`
	MaxRetries              int               `json:"max_retries"`
	RetryDelay              time.Duration     `json:"retry_delay"`
}

// DefaultTransactionConfig returns default transaction configuration optimized for PostgreSQL
func DefaultTransactionConfig() TransactionConfig {
	return TransactionConfig{
		DefaultIsolationLevel:    sql.LevelReadCommitted,   // PostgreSQL default
		HighConcurrencyLevel:     sql.LevelReadCommitted,   // Best for high throughput
		ReadOnlyLevel:           sql.LevelRepeatableRead,   // Consistent reads
		CriticalOperationLevel:  sql.LevelSerializable,    // Maximum consistency
		EnableDeadlockDetection: true,
		DeadlockTimeout:         30 * time.Second,
		MaxRetries:              3,
		RetryDelay:              100 * time.Millisecond,
	}
}

// TransactionManager provides enhanced transaction management with configurable isolation levels
type TransactionManager struct {
	db     *gorm.DB
	config TransactionConfig
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB, config TransactionConfig) *TransactionManager {
	return &TransactionManager{
		db:     db,
		config: config,
	}
}

// TransactionContext holds transaction context information
type TransactionContext struct {
	IsolationLevel sql.IsolationLevel
	ReadOnly      bool
	MaxRetries    int
	RetryDelay    time.Duration
	Description   string
}

// CriticalTransaction executes a transaction with SERIALIZABLE isolation for critical operations
func (tm *TransactionManager) CriticalTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	txCtx := TransactionContext{
		IsolationLevel: tm.config.CriticalOperationLevel,
		ReadOnly:      false,
		MaxRetries:    tm.config.MaxRetries,
		RetryDelay:    tm.config.RetryDelay,
		Description:   description,
	}
	
	return tm.ExecuteTransactionWithContext(ctx, txCtx, fn)
}

// HighConcurrencyTransaction executes a transaction optimized for high concurrency
func (tm *TransactionManager) HighConcurrencyTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	txCtx := TransactionContext{
		IsolationLevel: tm.config.HighConcurrencyLevel,
		ReadOnly:      false,
		MaxRetries:    tm.config.MaxRetries,
		RetryDelay:    tm.config.RetryDelay,
		Description:   description,
	}
	
	return tm.ExecuteTransactionWithContext(ctx, txCtx, fn)
}

// ReadOnlyTransaction executes a transaction optimized for read operations
func (tm *TransactionManager) ReadOnlyTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	txCtx := TransactionContext{
		IsolationLevel: tm.config.ReadOnlyLevel,
		ReadOnly:      true,
		MaxRetries:    1, // No retries needed for read-only
		RetryDelay:    0,
		Description:   description,
	}
	
	return tm.ExecuteTransactionWithContext(ctx, txCtx, fn)
}

// ExecuteTransactionWithContext executes a transaction with specific context settings
func (tm *TransactionManager) ExecuteTransactionWithContext(ctx context.Context, txCtx TransactionContext, fn func(*gorm.DB) error) error {
	var lastErr error
	
	for attempt := 0; attempt <= txCtx.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(txCtx.RetryDelay):
			}
			
			logger.Info("Retrying transaction",
				"attempt", attempt,
				"description", txCtx.Description,
				"lastError", lastErr,
			)
		}
		
		err := tm.executeTransaction(ctx, txCtx, fn)
		if err == nil {
			if attempt > 0 {
				logger.Info("Transaction succeeded after retry",
					"attempt", attempt,
					"description", txCtx.Description,
				)
			}
			return nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !tm.isRetryableError(err) {
			logger.Error("Non-retryable transaction error",
				err,
				"description", txCtx.Description,
				"attempt", attempt,
			)
			return err
		}
		
		logger.Warn("Retryable transaction error",
			"error", err,
			"description", txCtx.Description,
			"attempt", attempt,
		)
	}
	
	return fmt.Errorf("transaction failed after %d attempts: %w", txCtx.MaxRetries+1, lastErr)
}

// executeTransaction executes a single transaction attempt
func (tm *TransactionManager) executeTransaction(ctx context.Context, txCtx TransactionContext, fn func(*gorm.DB) error) error {
	startTime := time.Now()
	
	// Begin transaction with specific isolation level
	tx := tm.db.Begin(&sql.TxOptions{
		Isolation: txCtx.IsolationLevel,
		ReadOnly:  txCtx.ReadOnly,
	})
	
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	// Set up transaction timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, tm.config.DeadlockTimeout)
	defer cancel()
	
	// Execute with timeout context
	tx = tx.WithContext(timeoutCtx)
	
	success := false
	defer func() {
		duration := time.Since(startTime)
		
		if success {
			if err := tx.Commit().Error; err != nil {
				logger.Error("Transaction commit failed", err,
					"description", txCtx.Description,
					"duration", duration,
					"isolationLevel", isolationLevelToString(txCtx.IsolationLevel),
				)
			} else {
				logger.Debug("Transaction committed successfully",
					"description", txCtx.Description,
					"duration", duration,
					"isolationLevel", isolationLevelToString(txCtx.IsolationLevel),
				)
			}
		} else {
			if err := tx.Rollback().Error; err != nil {
				logger.Error("Transaction rollback failed", err,
					"description", txCtx.Description,
					"duration", duration,
				)
			} else {
				logger.Debug("Transaction rolled back",
					"description", txCtx.Description,
					"duration", duration,
				)
			}
		}
	}()
	
	// Execute the transaction function
	if err := fn(tx); err != nil {
		return err
	}
	
	success = true
	return nil
}

// isRetryableError determines if an error is retryable
func (tm *TransactionManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// PostgreSQL-specific retryable errors
	retryableErrors := []string{
		"could not serialize access",           // Serialization failure
		"deadlock detected",                   // Deadlock
		"lock timeout",                        // Lock timeout
		"connection reset",                    // Connection issues
		"connection refused",                  // Connection issues
		"server closed the connection",       // Connection issues
		"context deadline exceeded",          // Timeout
	}
	
	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}
	
	return false
}

// CreditTransferTransaction executes a credit transfer with appropriate isolation level
func (tm *TransactionManager) CreditTransferTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	// Credit transfers require high consistency to prevent double spending
	return tm.CriticalTransaction(ctx, fmt.Sprintf("Credit Transfer: %s", description), fn)
}

// OrderCreationTransaction executes an order creation with inventory checks
func (tm *TransactionManager) OrderCreationTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	// Orders need consistency for inventory management
	return tm.CriticalTransaction(ctx, fmt.Sprintf("Order Creation: %s", description), fn)
}

// UserDataTransaction executes user data updates with appropriate isolation
func (tm *TransactionManager) UserDataTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	// User data updates can use default isolation level
	return tm.HighConcurrencyTransaction(ctx, fmt.Sprintf("User Data: %s", description), fn)
}

// AnalyticsTransaction executes analytics queries with read consistency
func (tm *TransactionManager) AnalyticsTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	// Analytics queries benefit from consistent reads
	return tm.ReadOnlyTransaction(ctx, fmt.Sprintf("Analytics: %s", description), fn)
}

// BulkDataTransaction executes bulk operations with optimized settings
func (tm *TransactionManager) BulkDataTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	// Bulk operations use high concurrency settings
	txCtx := TransactionContext{
		IsolationLevel: sql.LevelReadCommitted,
		ReadOnly:      false,
		MaxRetries:    1, // Don't retry bulk operations
		RetryDelay:    0,
		Description:   fmt.Sprintf("Bulk Operation: %s", description),
	}
	
	return tm.ExecuteTransactionWithContext(ctx, txCtx, fn)
}

// GetTransactionStats returns transaction statistics
func (tm *TransactionManager) GetTransactionStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		ActiveTransactions int64 `json:"active_transactions"`
		CommittedTxns     int64 `json:"committed_txns"`
		RolledBackTxns    int64 `json:"rolled_back_txns"`
		DeadlockCount     int64 `json:"deadlock_count"`
	}
	
	// Query PostgreSQL statistics
	rows, err := tm.db.Raw(`
		SELECT 
			(SELECT count(*) FROM pg_stat_activity WHERE state = 'active') as active_transactions,
			(SELECT xact_commit FROM pg_stat_database WHERE datname = current_database()) as committed_txns,
			(SELECT xact_rollback FROM pg_stat_database WHERE datname = current_database()) as rolled_back_txns,
			(SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()) as deadlock_count
	`).Rows()
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	if rows.Next() {
		if err := rows.Scan(&stats.ActiveTransactions, &stats.CommittedTxns, &stats.RolledBackTxns, &stats.DeadlockCount); err != nil {
			return nil, err
		}
	}
	
	result := map[string]interface{}{
		"active_transactions": stats.ActiveTransactions,
		"committed_txns":     stats.CommittedTxns,
		"rolled_back_txns":   stats.RolledBackTxns,
		"deadlock_count":     stats.DeadlockCount,
		"config": map[string]interface{}{
			"default_isolation":     isolationLevelToString(tm.config.DefaultIsolationLevel),
			"high_concurrency":      isolationLevelToString(tm.config.HighConcurrencyLevel),
			"read_only":            isolationLevelToString(tm.config.ReadOnlyLevel),
			"critical_operation":   isolationLevelToString(tm.config.CriticalOperationLevel),
			"deadlock_timeout":     tm.config.DeadlockTimeout,
			"max_retries":          tm.config.MaxRetries,
		},
	}
	
	return result, nil
}

// OptimizeForHighConcurrency configures transaction manager for high concurrency scenarios
func (tm *TransactionManager) OptimizeForHighConcurrency() {
	tm.config.DefaultIsolationLevel = sql.LevelReadCommitted
	tm.config.HighConcurrencyLevel = sql.LevelReadCommitted
	tm.config.DeadlockTimeout = 15 * time.Second
	tm.config.MaxRetries = 5
	tm.config.RetryDelay = 50 * time.Millisecond
	
	logger.Info("Transaction manager optimized for high concurrency")
}

// OptimizeForConsistency configures transaction manager for maximum consistency
func (tm *TransactionManager) OptimizeForConsistency() {
	tm.config.DefaultIsolationLevel = sql.LevelRepeatableRead
	tm.config.HighConcurrencyLevel = sql.LevelRepeatableRead
	tm.config.DeadlockTimeout = 60 * time.Second
	tm.config.MaxRetries = 3
	tm.config.RetryDelay = 200 * time.Millisecond
	
	logger.Info("Transaction manager optimized for consistency")
}

// Helper functions

func isolationLevelToString(level sql.IsolationLevel) string {
	switch level {
	case sql.LevelDefault:
		return "DEFAULT"
	case sql.LevelReadUncommitted:
		return "READ_UNCOMMITTED"
	case sql.LevelReadCommitted:
		return "READ_COMMITTED"
	case sql.LevelWriteCommitted:
		return "WRITE_COMMITTED"
	case sql.LevelRepeatableRead:
		return "REPEATABLE_READ"
	case sql.LevelSnapshot:
		return "SNAPSHOT"
	case sql.LevelSerializable:
		return "SERIALIZABLE"
	case sql.LevelLinearizable:
		return "LINEARIZABLE"
	default:
		return "UNKNOWN"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || 
		    (len(s) > 0 && (s == substr || 
		     (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		      func() bool {
		      	for i := 0; i <= len(s)-len(substr); i++ {
		      		if s[i:i+len(substr)] == substr {
		      			return true
		      		}
		      	}
		      	return false
		      }())))))
}