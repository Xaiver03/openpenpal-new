package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held")
	ErrLockExpired     = errors.New("lock expired")
)

// DistributedLock implements a Redis-based distributed locking mechanism
type DistributedLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
	retries    int
	retryDelay time.Duration
}

// DistributedLockManager manages distributed locks across the system
type DistributedLockManager struct {
	client *redis.Client
	prefix string
}

// NewDistributedLockManager creates a new lock manager
func NewDistributedLockManager(client *redis.Client, prefix string) *DistributedLockManager {
	if prefix == "" {
		prefix = "openpenpal:lock:"
	}
	return &DistributedLockManager{
		client: client,
		prefix: prefix,
	}
}

// NewLock creates a new distributed lock instance
func (dlm *DistributedLockManager) NewLock(key string, expiration time.Duration) *DistributedLock {
	// Generate a unique value for this lock instance
	value := generateLockValue()
	
	return &DistributedLock{
		client:     dlm.client,
		key:        dlm.prefix + key,
		value:      value,
		expiration: expiration,
		retries:    3,
		retryDelay: 100 * time.Millisecond,
	}
}

// WithRetries sets the number of retries for acquiring the lock
func (dl *DistributedLock) WithRetries(retries int, delay time.Duration) *DistributedLock {
	dl.retries = retries
	dl.retryDelay = delay
	return dl
}

// Acquire attempts to acquire the lock
func (dl *DistributedLock) Acquire(ctx context.Context) error {
	for i := 0; i <= dl.retries; i++ {
		// Try to set the key only if it doesn't exist
		success, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		
		if success {
			log.Printf("[DistributedLock] Acquired lock for key: %s", dl.key)
			return nil
		}
		
		// If this wasn't the last attempt, wait before retrying
		if i < dl.retries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dl.retryDelay):
				continue
			}
		}
	}
	
	return ErrLockNotAcquired
}

// Release releases the lock if it's still held by this instance
func (dl *DistributedLock) Release(ctx context.Context) error {
	// Lua script to ensure we only delete if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	
	log.Printf("[DistributedLock] Released lock for key: %s", dl.key)
	return nil
}

// Extend extends the lock expiration if it's still held by this instance
func (dl *DistributedLock) Extend(ctx context.Context, duration time.Duration) error {
	// Lua script to ensure we only extend if we own the lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	
	milliseconds := duration.Milliseconds()
	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value, milliseconds).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}
	
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	
	log.Printf("[DistributedLock] Extended lock for key: %s by %v", dl.key, duration)
	return nil
}

// IsHeld checks if the lock is still held by this instance
func (dl *DistributedLock) IsHeld(ctx context.Context) bool {
	value, err := dl.client.Get(ctx, dl.key).Result()
	if err != nil {
		return false
	}
	return value == dl.value
}

// RunWithLock executes a function while holding a distributed lock
func (dlm *DistributedLockManager) RunWithLock(
	ctx context.Context,
	key string,
	expiration time.Duration,
	fn func() error,
) error {
	lock := dlm.NewLock(key, expiration)
	
	// Acquire the lock
	if err := lock.Acquire(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock for %s: %w", key, err)
	}
	
	// Ensure we release the lock when done
	defer func() {
		if err := lock.Release(ctx); err != nil {
			log.Printf("[DistributedLock] Failed to release lock for %s: %v", key, err)
		}
	}()
	
	// Execute the function
	return fn()
}

// RunWithLockExtension runs a function with automatic lock extension for long-running tasks
func (dlm *DistributedLockManager) RunWithLockExtension(
	ctx context.Context,
	key string,
	initialExpiration time.Duration,
	extensionInterval time.Duration,
	fn func(ctx context.Context) error,
) error {
	lock := dlm.NewLock(key, initialExpiration)
	
	// Acquire the lock
	if err := lock.Acquire(ctx); err != nil {
		return fmt.Errorf("failed to acquire lock for %s: %w", key, err)
	}
	
	// Create a context that we can cancel
	fnCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Start lock extension goroutine
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(extensionInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if err := lock.Extend(ctx, initialExpiration); err != nil {
					log.Printf("[DistributedLock] Failed to extend lock for %s: %v", key, err)
					cancel() // Cancel the function context
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	
	// Execute the function
	err := fn(fnCtx)
	
	// Stop the extension goroutine
	close(done)
	
	// Release the lock
	if releaseErr := lock.Release(ctx); releaseErr != nil {
		log.Printf("[DistributedLock] Failed to release lock for %s: %v", key, releaseErr)
	}
	
	return err
}

// generateLockValue generates a unique value for a lock instance
func generateLockValue() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// WaitForLock waits for a lock to become available
func (dl *DistributedLock) WaitForLock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		err := dl.Acquire(ctx)
		if err == nil {
			return nil
		}
		
		if err != ErrLockNotAcquired {
			return err
		}
		
		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			continue
		}
	}
	
	return fmt.Errorf("timeout waiting for lock after %v", timeout)
}