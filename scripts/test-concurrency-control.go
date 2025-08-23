package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 模拟用户积分记录
type UserCredit struct {
	ID        string
	UserID    string
	Total     int
	Available int
	Used      int
	Earned    int
	Level     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	// 初始化数据库连接
	dsn := "host=localhost user=postgres password=password dbname=openpenpal port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 初始化Redis连接
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 测试场景
	fmt.Println("🔬 OpenPenPal 并发控制测试")
	fmt.Println("=" * 50)
	
	// 场景1: 测试并发创建用户积分记录
	testConcurrentUserCreation(db, rdb)
	
	// 场景2: 测试并发积分操作
	testConcurrentCreditOperations(db, rdb)
	
	// 场景3: 测试频率限制
	testRateLimiting(db, rdb)
	
	// 场景4: 测试批量操作
	testBatchOperations(db, rdb)
}

// 场景1: 测试并发创建用户积分记录
func testConcurrentUserCreation(db *gorm.DB, rdb *redis.Client) {
	fmt.Println("\n📋 场景1: 并发创建用户积分记录")
	fmt.Println("-" * 40)
	
	userID := "test_user_concurrent_" + fmt.Sprintf("%d", time.Now().UnixNano())
	concurrency := 10
	
	var wg sync.WaitGroup
	errors := make([]error, concurrency)
	
	start := time.Now()
	
	// 模拟多个并发请求同时创建同一用户的积分记录
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// 模拟GetOrCreateUserCredit操作
			var credit UserCredit
			err := db.Where("user_id = ?", userID).First(&credit).Error
			if err == gorm.ErrRecordNotFound {
				// 尝试创建新记录
				credit = UserCredit{
					ID:        fmt.Sprintf("credit_%d_%d", index, time.Now().UnixNano()),
					UserID:    userID,
					Total:     0,
					Available: 0,
					Used:      0,
					Earned:    0,
					Level:     1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				err = db.Create(&credit).Error
			}
			errors[index] = err
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	// 统计结果
	successCount := 0
	duplicateCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		} else if isDuplicateError(err) {
			duplicateCount++
		}
	}
	
	fmt.Printf("并发请求数: %d\n", concurrency)
	fmt.Printf("成功创建数: %d\n", successCount)
	fmt.Printf("重复错误数: %d\n", duplicateCount)
	fmt.Printf("执行时间: %v\n", elapsed)
	
	// 验证最终只有一条记录
	var count int64
	db.Model(&UserCredit{}).Where("user_id = ?", userID).Count(&count)
	fmt.Printf("最终记录数: %d (应该为1)\n", count)
	
	if count != 1 {
		fmt.Printf("❌ 错误: 期望1条记录，实际%d条\n", count)
	} else {
		fmt.Printf("✅ 正确: 只创建了1条记录\n")
	}
}

// 场景2: 测试并发积分操作
func testConcurrentCreditOperations(db *gorm.DB, rdb *redis.Client) {
	fmt.Println("\n📋 场景2: 并发积分操作")
	fmt.Println("-" * 40)
	
	userID := "test_user_operations_" + fmt.Sprintf("%d", time.Now().UnixNano())
	
	// 先创建用户积分记录
	credit := UserCredit{
		ID:        fmt.Sprintf("credit_%d", time.Now().UnixNano()),
		UserID:    userID,
		Total:     1000,
		Available: 1000,
		Used:      0,
		Earned:    1000,
		Level:     1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&credit)
	
	concurrency := 20
	pointsPerOperation := 10
	
	var wg sync.WaitGroup
	start := time.Now()
	
	// 模拟多个并发扣减积分操作
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// 使用事务确保原子性
			db.Transaction(func(tx *gorm.DB) error {
				var c UserCredit
				if err := tx.Set("gorm:query_option", "FOR UPDATE").
					Where("user_id = ?", userID).First(&c).Error; err != nil {
					return err
				}
				
				if c.Available < pointsPerOperation {
					return fmt.Errorf("insufficient credits")
				}
				
				c.Available -= pointsPerOperation
				c.Used += pointsPerOperation
				
				return tx.Save(&c).Error
			})
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	// 验证最终结果
	var finalCredit UserCredit
	db.Where("user_id = ?", userID).First(&finalCredit)
	
	expectedAvailable := 1000 - (concurrency * pointsPerOperation)
	expectedUsed := concurrency * pointsPerOperation
	
	fmt.Printf("并发操作数: %d\n", concurrency)
	fmt.Printf("每次扣减: %d\n", pointsPerOperation)
	fmt.Printf("期望剩余: %d, 实际剩余: %d\n", expectedAvailable, finalCredit.Available)
	fmt.Printf("期望使用: %d, 实际使用: %d\n", expectedUsed, finalCredit.Used)
	fmt.Printf("执行时间: %v\n", elapsed)
	
	if finalCredit.Available == expectedAvailable && finalCredit.Used == expectedUsed {
		fmt.Printf("✅ 正确: 积分计算准确\n")
	} else {
		fmt.Printf("❌ 错误: 积分计算不一致\n")
	}
}

// 场景3: 测试频率限制
func testRateLimiting(db *gorm.DB, rdb *redis.Client) {
	fmt.Println("\n📋 场景3: 频率限制测试")
	fmt.Println("-" * 40)
	
	ctx := context.Background()
	userID := "test_user_rate_" + fmt.Sprintf("%d", time.Now().UnixNano())
	actionType := "test_action"
	
	// 清理可能存在的旧数据
	rdb.Del(ctx, fmt.Sprintf("rate_limit:%s:%s", userID, actionType))
	
	maxCount := 5
	windowSize := 10 * time.Second
	
	fmt.Printf("限制规则: %d次/%v\n", maxCount, windowSize)
	
	// 快速发送请求
	successCount := 0
	blockedCount := 0
	
	for i := 0; i < maxCount+3; i++ {
		// 模拟频率检查
		countKey := fmt.Sprintf("rate_limit:%s:%s", userID, actionType)
		
		pipe := rdb.Pipeline()
		now := time.Now()
		windowStart := now.Add(-windowSize)
		
		// 移除窗口外的记录
		pipe.ZRemRangeByScore(ctx, countKey, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))
		
		// 获取当前计数
		countCmd := pipe.ZCard(ctx, countKey)
		
		// 添加当前请求
		pipe.ZAdd(ctx, countKey, redis.Z{
			Score:  float64(now.UnixMilli()),
			Member: fmt.Sprintf("%d", now.UnixNano()),
		})
		
		// 设置过期时间
		pipe.Expire(ctx, countKey, windowSize+time.Minute)
		
		_, err := pipe.Exec(ctx)
		if err != nil {
			fmt.Printf("Redis错误: %v\n", err)
			continue
		}
		
		currentCount := countCmd.Val()
		if currentCount < int64(maxCount) {
			successCount++
			fmt.Printf("请求 %d: ✅ 允许 (当前计数: %d)\n", i+1, currentCount+1)
		} else {
			blockedCount++
			fmt.Printf("请求 %d: ❌ 拒绝 (达到限制: %d)\n", i+1, maxCount)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("\n结果: 成功 %d, 拒绝 %d\n", successCount, blockedCount)
	
	if successCount == maxCount && blockedCount == 3 {
		fmt.Printf("✅ 正确: 频率限制生效\n")
	} else {
		fmt.Printf("⚠️  警告: 频率限制可能未正确生效\n")
	}
}

// 场景4: 测试批量操作
func testBatchOperations(db *gorm.DB, rdb *redis.Client) {
	fmt.Println("\n📋 场景4: 批量操作测试")
	fmt.Println("-" * 40)
	
	// 准备测试数据
	totalItems := 1000
	batchSize := 100
	
	items := make([]interface{}, totalItems)
	for i := 0; i < totalItems; i++ {
		items[i] = UserCredit{
			ID:        fmt.Sprintf("batch_credit_%d", i),
			UserID:    fmt.Sprintf("batch_user_%d", i),
			Total:     100,
			Available: 100,
			Used:      0,
			Earned:    100,
			Level:     1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
	
	start := time.Now()
	processedCount := 0
	
	// 批量处理
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		
		batch := items[i:end]
		
		// 模拟批量插入
		db.Transaction(func(tx *gorm.DB) error {
			for _, item := range batch {
				if err := tx.Create(item).Error; err != nil {
					return err
				}
			}
			processedCount += len(batch)
			return nil
		})
		
		fmt.Printf("处理批次 %d/%d (已处理: %d)\n", (i/batchSize)+1, (totalItems+batchSize-1)/batchSize, processedCount)
		
		// 批次间延迟
		time.Sleep(10 * time.Millisecond)
	}
	
	elapsed := time.Since(start)
	
	// 验证结果
	var count int64
	db.Model(&UserCredit{}).Where("user_id LIKE 'batch_user_%'").Count(&count)
	
	fmt.Printf("\n总项目数: %d\n", totalItems)
	fmt.Printf("批次大小: %d\n", batchSize)
	fmt.Printf("实际插入: %d\n", count)
	fmt.Printf("执行时间: %v\n", elapsed)
	fmt.Printf("平均速度: %.2f items/秒\n", float64(totalItems)/elapsed.Seconds())
	
	if int(count) == totalItems {
		fmt.Printf("✅ 正确: 批量操作完成\n")
	} else {
		fmt.Printf("❌ 错误: 期望插入%d条，实际%d条\n", totalItems, count)
	}
	
	// 清理测试数据
	db.Where("user_id LIKE 'batch_user_%'").Delete(&UserCredit{})
}

// 判断是否是重复键错误
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "duplicate key") || contains(errStr, "UNIQUE constraint")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}