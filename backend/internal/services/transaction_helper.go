package services

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"openpenpal-backend/internal/config"
)

// TransactionHelper provides helper methods for database transactions
type TransactionHelper struct {
	db      *gorm.DB
	manager *config.TransactionManager
}

// NewTransactionHelper creates a new transaction helper
func NewTransactionHelper(db *gorm.DB) *TransactionHelper {
	transactionConfig := config.DefaultTransactionConfig()
	manager := config.NewTransactionManager(db, transactionConfig)
	
	return &TransactionHelper{
		db:      db,
		manager: manager,
	}
}

// NewTransactionHelperWithConfig creates a new transaction helper with custom configuration
func NewTransactionHelperWithConfig(db *gorm.DB, transactionConfig config.TransactionConfig) *TransactionHelper {
	manager := config.NewTransactionManager(db, transactionConfig)
	
	return &TransactionHelper{
		db:      db,
		manager: manager,
	}
}

// WithTransaction executes a function within a database transaction
func (h *TransactionHelper) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	// Check if we're already in a transaction
	tx, ok := ctx.Value("tx").(*gorm.DB)
	if ok && tx != nil {
		// Already in a transaction, just execute the function
		return fn(tx)
	}

	// Start a new transaction
	tx = h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Defer rollback in case of panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithNestedTransaction executes a function within a nested transaction (savepoint)
func (h *TransactionHelper) WithNestedTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	// Get existing transaction from context
	tx, ok := ctx.Value("tx").(*gorm.DB)
	if !ok || tx == nil {
		// No existing transaction, use regular transaction
		return h.WithTransaction(ctx, fn)
	}

	// Create a savepoint
	savepoint := fmt.Sprintf("sp_%d", ctx.Value("savepoint_count").(int))
	if err := tx.Exec("SAVEPOINT " + savepoint).Error; err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}

	// Execute the function
	if err := fn(tx); err != nil {
		// Rollback to savepoint
		tx.Exec("ROLLBACK TO SAVEPOINT " + savepoint)
		return err
	}

	// Release the savepoint
	if err := tx.Exec("RELEASE SAVEPOINT " + savepoint).Error; err != nil {
		return fmt.Errorf("failed to release savepoint: %w", err)
	}

	return nil
}

// ContextWithTransaction returns a new context with the transaction
func ContextWithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, "tx", tx)
}

// TransactionFromContext gets the transaction from context
func TransactionFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value("tx").(*gorm.DB)
	return tx, ok
}

// GetDB returns the database connection or transaction from context
func (h *TransactionHelper) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := TransactionFromContext(ctx); ok && tx != nil {
		return tx
	}
	return h.db.WithContext(ctx)
}

// WithCriticalTransaction executes a function with SERIALIZABLE isolation for critical operations
func (h *TransactionHelper) WithCriticalTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.CriticalTransaction(ctx, description, fn)
}

// WithHighConcurrencyTransaction executes a function with optimized settings for high concurrency
func (h *TransactionHelper) WithHighConcurrencyTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.HighConcurrencyTransaction(ctx, description, fn)
}

// WithReadOnlyTransaction executes a function with settings optimized for read operations
func (h *TransactionHelper) WithReadOnlyTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.ReadOnlyTransaction(ctx, description, fn)
}

// WithCreditTransferTransaction executes a credit transfer with appropriate isolation
func (h *TransactionHelper) WithCreditTransferTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.CreditTransferTransaction(ctx, description, fn)
}

// WithOrderCreationTransaction executes an order creation with inventory checks
func (h *TransactionHelper) WithOrderCreationTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.OrderCreationTransaction(ctx, description, fn)
}

// WithUserDataTransaction executes user data updates with appropriate isolation
func (h *TransactionHelper) WithUserDataTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.UserDataTransaction(ctx, description, fn)
}

// WithAnalyticsTransaction executes analytics queries with read consistency
func (h *TransactionHelper) WithAnalyticsTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.AnalyticsTransaction(ctx, description, fn)
}

// WithBulkDataTransaction executes bulk operations with optimized settings
func (h *TransactionHelper) WithBulkDataTransaction(ctx context.Context, description string, fn func(*gorm.DB) error) error {
	return h.manager.BulkDataTransaction(ctx, description, fn)
}

// GetTransactionStats returns transaction statistics
func (h *TransactionHelper) GetTransactionStats(ctx context.Context) (map[string]interface{}, error) {
	return h.manager.GetTransactionStats(ctx)
}

// OptimizeForHighConcurrency optimizes transaction settings for high concurrency scenarios
func (h *TransactionHelper) OptimizeForHighConcurrency() {
	h.manager.OptimizeForHighConcurrency()
}

// OptimizeForConsistency optimizes transaction settings for maximum consistency
func (h *TransactionHelper) OptimizeForConsistency() {
	h.manager.OptimizeForConsistency()
}
