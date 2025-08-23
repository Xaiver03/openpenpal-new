package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ConcurrencyManager 管理数据库并发控制
type ConcurrencyManager struct {
	db          *gorm.DB
	redisClient *redis.Client
	userLocks   sync.Map // 用户级别的分布式锁缓存
	lockTimeout time.Duration
}

// NewConcurrencyManager 创建并发控制管理器
func NewConcurrencyManager(db *gorm.DB, redisClient *redis.Client) *ConcurrencyManager {
	return &ConcurrencyManager{
		db:          db,
		redisClient: redisClient,
		lockTimeout: 30 * time.Second, // 默认锁超时时间
	}
}

// UserOperationLock 用户操作级别的分布式锁
type UserOperationLock struct {
	manager   *ConcurrencyManager
	lockKey   string
	lockValue string
	acquired  bool
}

// AcquireUserLock 获取用户操作锁（防止同一用户的并发操作）
func (cm *ConcurrencyManager) AcquireUserLock(ctx context.Context, userID, operation string) (*UserOperationLock, error) {
	lockKey := fmt.Sprintf("user_lock:%s:%s", userID, operation)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	
	lock := &UserOperationLock{
		manager:   cm,
		lockKey:   lockKey,
		lockValue: lockValue,
		acquired:  false,
	}
	
	// 尝试获取Redis分布式锁
	result := cm.redisClient.SetNX(ctx, lockKey, lockValue, cm.lockTimeout)
	if result.Err() != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", result.Err())
	}
	
	if result.Val() {
		lock.acquired = true
		return lock, nil
	}
	
	// 锁已被占用
	return nil, fmt.Errorf("operation %s for user %s is already in progress", operation, userID)
}

// Release 释放锁
func (lock *UserOperationLock) Release(ctx context.Context) error {
	if !lock.acquired {
		return nil
	}
	
	// 使用Lua脚本确保只有锁的持有者可以释放锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	
	result := lock.manager.redisClient.Eval(ctx, script, []string{lock.lockKey}, lock.lockValue)
	if result.Err() != nil {
		return fmt.Errorf("failed to release lock: %w", result.Err())
	}
	
	lock.acquired = false
	return nil
}

// WithUserLock 在用户锁保护下执行操作
func (cm *ConcurrencyManager) WithUserLock(ctx context.Context, userID, operation string, fn func() error) error {
	lock, err := cm.AcquireUserLock(ctx, userID, operation)
	if err != nil {
		return err
	}
	
	defer func() {
		if releaseErr := lock.Release(ctx); releaseErr != nil {
			// 记录释放锁失败的错误，但不影响主操作结果
			fmt.Printf("Warning: failed to release lock for user %s operation %s: %v\n", userID, operation, releaseErr)
		}
	}()
	
	return fn()
}

// AtomicUserCreditOperation 原子性用户积分操作
func (cm *ConcurrencyManager) AtomicUserCreditOperation(ctx context.Context, userID string, operation func(*gorm.DB) error) error {
	return cm.WithUserLock(ctx, userID, "credit_operation", func() error {
		// 在事务中执行操作
		return cm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return operation(tx)
		})
	})
}

// GetOrCreateUserCreditSafe 线程安全的获取或创建用户积分记录
func (cm *ConcurrencyManager) GetOrCreateUserCreditSafe(ctx context.Context, userID string, userCreditModel interface{}) error {
	return cm.WithUserLock(ctx, userID, "get_or_create_credit", func() error {
		return cm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 尝试获取记录
			err := tx.Where("user_id = ?", userID).First(userCreditModel).Error
			if err == nil {
				return nil // 记录已存在
			}
			
			if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to query user credit: %w", err)
			}
			
			// 记录不存在，创建新记录
			return tx.Create(userCreditModel).Error
		})
	})
}

// RateLimitedOperation 带频率限制的操作
type RateLimitedOperation struct {
	UserID       string
	ActionType   string
	WindowSize   time.Duration // 时间窗口大小
	MaxCount     int           // 时间窗口内最大操作次数
	CountKey     string        // Redis计数器键
}

// CheckRateLimit 检查频率限制
func (cm *ConcurrencyManager) CheckRateLimit(ctx context.Context, op RateLimitedOperation) (bool, error) {
	if op.CountKey == "" {
		op.CountKey = fmt.Sprintf("rate_limit:%s:%s", op.UserID, op.ActionType)
	}
	
	// 使用Redis滑动窗口算法
	now := time.Now()
	windowStart := now.Add(-op.WindowSize)
	
	pipe := cm.redisClient.Pipeline()
	
	// 移除窗口外的记录
	pipe.ZRemRangeByScore(ctx, op.CountKey, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))
	
	// 获取当前窗口内的计数
	countCmd := pipe.ZCard(ctx, op.CountKey)
	
	// 添加当前操作的时间戳
	pipe.ZAdd(ctx, op.CountKey, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// 设置键的过期时间
	pipe.Expire(ctx, op.CountKey, op.WindowSize+time.Minute)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	currentCount := countCmd.Val()
	return currentCount < int64(op.MaxCount), nil
}

// BatchOperation 批量操作控制
type BatchOperation struct {
	BatchSize     int           // 批次大小
	DelayBetween  time.Duration // 批次间延迟
	MaxRetries    int           // 最大重试次数
	RetryDelay    time.Duration // 重试延迟
}

// ExecuteBatch 执行批量操作
func (cm *ConcurrencyManager) ExecuteBatch(ctx context.Context, items []interface{}, op BatchOperation, processor func([]interface{}) error) error {
	if op.BatchSize <= 0 {
		op.BatchSize = 100 // 默认批次大小
	}
	if op.MaxRetries <= 0 {
		op.MaxRetries = 3 // 默认重试次数
	}
	if op.RetryDelay <= 0 {
		op.RetryDelay = time.Second // 默认重试延迟
	}
	
	for i := 0; i < len(items); i += op.BatchSize {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		end := i + op.BatchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		
		// 带重试的批次处理
		var lastErr error
		for retry := 0; retry <= op.MaxRetries; retry++ {
			err := processor(batch)
			if err == nil {
				break // 成功处理
			}
			
			lastErr = err
			if retry < op.MaxRetries {
				// 等待后重试
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(op.RetryDelay * time.Duration(retry+1)): // 指数退避
				}
			}
		}
		
		if lastErr != nil {
			return fmt.Errorf("batch processing failed after %d retries: %w", op.MaxRetries, lastErr)
		}
		
		// 批次间延迟
		if op.DelayBetween > 0 && i+op.BatchSize < len(items) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(op.DelayBetween):
			}
		}
	}
	
	return nil
}

// ConcurrentWorkerPool 并发工作池
type ConcurrentWorkerPool struct {
	workerCount int
	jobQueue    chan func() error
	results     chan error
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewConcurrentWorkerPool 创建并发工作池
func (cm *ConcurrencyManager) NewConcurrentWorkerPool(ctx context.Context, workerCount int) *ConcurrentWorkerPool {
	ctx, cancel := context.WithCancel(ctx)
	
	pool := &ConcurrentWorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan func() error, workerCount*2), // 缓冲队列
		results:     make(chan error, workerCount*2),
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// 启动工作线程
	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
	
	return pool
}

// worker 工作线程
func (pool *ConcurrentWorkerPool) worker() {
	defer pool.wg.Done()
	
	for {
		select {
		case <-pool.ctx.Done():
			return
		case job, ok := <-pool.jobQueue:
			if !ok {
				return
			}
			
			err := job()
			pool.results <- err
		}
	}
}

// Submit 提交任务
func (pool *ConcurrentWorkerPool) Submit(job func() error) error {
	select {
	case <-pool.ctx.Done():
		return pool.ctx.Err()
	case pool.jobQueue <- job:
		return nil
	}
}

// Close 关闭工作池
func (pool *ConcurrentWorkerPool) Close() {
	close(pool.jobQueue)
	pool.cancel()
	pool.wg.Wait()
	close(pool.results)
}

// WaitForResults 等待所有结果
func (pool *ConcurrentWorkerPool) WaitForResults(expectedCount int) []error {
	var errors []error
	
	for i := 0; i < expectedCount; i++ {
		select {
		case <-pool.ctx.Done():
			errors = append(errors, pool.ctx.Err())
		case err := <-pool.results:
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	
	return errors
}

// OptimisticLockUpdate 乐观锁更新
func (cm *ConcurrencyManager) OptimisticLockUpdate(ctx context.Context, model interface{}, updateFn func() error, maxRetries int) error {
	if maxRetries <= 0 {
		maxRetries = 5 // 默认最大重试次数
	}
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := updateFn()
		if err == nil {
			return nil // 成功更新
		}
		
		// 检查是否是乐观锁冲突
		if isOptimisticLockError(err) {
			// 短暂等待后重试
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond * time.Duration(10*(attempt+1))): // 递增退避
			}
			continue
		}
		
		// 其他错误直接返回
		return err
	}
	
	return fmt.Errorf("optimistic lock update failed after %d attempts", maxRetries)
}

// isOptimisticLockError 检查是否是乐观锁错误
func isOptimisticLockError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// PostgreSQL乐观锁错误模式
	return gorm.ErrRecordNotFound == err || 
		   err == gorm.ErrInvalidTransaction ||
		   // 检查是否包含版本冲突相关的错误信息
		   (errStr != "" && (
			   fmt.Sprintf("%v", err) == "record not found" ||
			   fmt.Sprintf("%v", err) == "affected rows is 0"))
}

// GetStatistics 获取并发控制统计信息
func (cm *ConcurrencyManager) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 获取当前活跃的锁数量
	lockPattern := "user_lock:*"
	keys, err := cm.redisClient.Keys(ctx, lockPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get lock statistics: %w", err)
	}
	
	stats["active_locks"] = len(keys)
	
	// 获取频率限制统计
	rateLimitPattern := "rate_limit:*"
	rateLimitKeys, err := cm.redisClient.Keys(ctx, rateLimitPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit statistics: %w", err)
	}
	
	stats["rate_limited_users"] = len(rateLimitKeys)
	stats["lock_timeout"] = cm.lockTimeout.String()
	
	return stats, nil
}