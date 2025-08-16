package main

import (
	"fmt"
	"time"

	"openpenpal-backend/internal/utils"
)

func main() {
	fmt.Println("🧠 智能日志聚合系统演示")
	fmt.Println("===========================")

	// 创建智能日志管理器
	config := &utils.SmartLoggerConfig{
		TimeWindow:              2 * time.Minute,  // 短时间窗口便于演示
		MaxAggregation:          1000,
		VerboseThreshold:        3,                // 3次后进入静默模式
		CircuitBreakerThreshold: 10,               // 10次后断路器开启
		SamplingRate:            5,                // 每5次采样一次
		CleanupInterval:         1 * time.Minute,
	}
	
	smartLogger := utils.NewSmartLogger(config)

	fmt.Println("\n🟢 演示场景1: 正常错误记录")
	fmt.Println("================================")
	
	// 正常错误，应该完整记录
	for i := 1; i <= 2; i++ {
		smartLogger.LogError(fmt.Sprintf("Database connection failed: %d", i), map[string]interface{}{
			"attempt": i,
			"timeout": "5s",
		})
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n🟡 演示场景2: 重复错误聚合")
	fmt.Println("================================")
	
	// 重复错误，应该触发聚合机制
	for i := 1; i <= 8; i++ {
		smartLogger.LogError("Task 4fa8f991-3886-41f4-8984-d14677e870aa failed: letter not found", map[string]interface{}{
			"task_id": "4fa8f991-3886-41f4-8984-d14677e870aa",
			"user_id": "test-admin",
			"attempt": i,
		})
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\n🔴 演示场景3: 高频错误断路器")
	fmt.Println("================================")
	
	// 继续相同错误，应该触发断路器
	for i := 9; i <= 15; i++ {
		smartLogger.LogError("Task 4fa8f991-3886-41f4-8984-d14677e870aa failed: letter not found", map[string]interface{}{
			"task_id": "4fa8f991-3886-41f4-8984-d14677e870aa",
			"user_id": "test-admin",
			"attempt": i,
		})
		time.Sleep(30 * time.Millisecond)
	}

	fmt.Println("\n🟦 演示场景4: 不同错误模式")
	fmt.Println("================================")
	
	// 不同的错误模式，应该分别处理
	errorPatterns := []struct {
		message string
		context map[string]interface{}
	}{
		{
			message: "Redis connection timeout",
			context: map[string]interface{}{"service": "redis", "timeout": "2s"},
		},
		{
			message: "PostgreSQL query timeout: SELECT * FROM users WHERE id = 123",
			context: map[string]interface{}{"query_type": "select", "table": "users"},
		},
		{
			message: "PostgreSQL query timeout: SELECT * FROM users WHERE id = 456", 
			context: map[string]interface{}{"query_type": "select", "table": "users"},
		},
		{
			message: "API rate limit exceeded for endpoint /api/letters",
			context: map[string]interface{}{"endpoint": "/api/letters", "limit": "100/hour"},
		},
	}

	for i, pattern := range errorPatterns {
		for j := 0; j < 3; j++ {
			smartLogger.LogError(pattern.message, pattern.context)
			time.Sleep(20 * time.Millisecond)
		}
		
		if i < len(errorPatterns)-1 {
			fmt.Printf("   --- 错误类型 %d 完成 ---\n", i+1)
		}
	}

	fmt.Println("\n🔍 演示场景5: 混合日志类型")
	fmt.Println("================================")
	
	// 混合不同级别的日志
	smartLogger.LogInfo("Service started successfully")
	smartLogger.LogWarning("Memory usage is high: 85%", map[string]interface{}{
		"memory_percent": 85,
		"threshold": 80,
	})
	
	// 重复警告
	for i := 0; i < 8; i++ {
		smartLogger.LogWarning("Memory usage is high: 85%", map[string]interface{}{
			"memory_percent": 85,
			"check_number": i + 1,
		})
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\n📊 日志统计报告")
	fmt.Println("===============")
	
	// 获取统计信息
	stats := smartLogger.GetStats()
	fmt.Printf("总错误数: %d\n", stats.TotalErrors)
	fmt.Printf("聚合错误数: %d\n", stats.AggregatedErrors)
	fmt.Printf("静默错误数: %d\n", stats.SilencedErrors)
	fmt.Printf("断路器阻断: %d\n", stats.CircuitedErrors)
	fmt.Printf("日志减少率: %.1f%%\n", stats.LogReduction)

	fmt.Println("\n📋 错误摘要")
	fmt.Println("============")
	
	// 打印错误摘要
	smartLogger.PrintSummary()

	fmt.Println("\n🎯 对比分析")
	fmt.Println("============")
	
	fmt.Printf("🔥 传统日志系统: 如果不使用智能聚合，上述演示会产生 %d 条日志记录\n", stats.TotalErrors)
	fmt.Printf("✨ 智能日志系统: 实际产生的有效日志记录大幅减少\n")
	fmt.Printf("💾 空间节省: %.1f%% 的日志被聚合或阻断\n", stats.LogReduction)
	
	if stats.LogReduction > 70 {
		fmt.Println("🎉 优秀！智能日志系统显著减少了日志膨胀")
	} else if stats.LogReduction > 40 {
		fmt.Println("👍 良好！智能日志系统有效控制了日志量")
	} else {
		fmt.Println("⚠️  需要调整：可以进一步优化聚合策略")
	}

	fmt.Println("\n🔧 实际应用建议")
	fmt.Println("================")
	fmt.Println("1. 集成到现有服务：替换标准 log.Printf")
	fmt.Println("2. 配置适当阈值：根据服务特点调整参数")
	fmt.Println("3. 监控统计数据：定期检查日志减少率")
	fmt.Println("4. 设置告警规则：断路器开启时发送通知")
	fmt.Println("5. 定期清理归档：配合日志轮转机制")

	fmt.Println("\n✅ 智能日志聚合系统演示完成！")
	
	// 演示如何在生产环境中使用
	fmt.Println("\n🚀 生产环境集成示例")
	fmt.Println("====================")
	
	productionExample()
}

func productionExample() {
	// 生产环境配置示例
	prodConfig := &utils.SmartLoggerConfig{
		TimeWindow:              10 * time.Minute, // 生产环境时间窗口更长
		MaxAggregation:          10000,            // 更大的聚合容量
		VerboseThreshold:        10,               // 更高的详细阈值
		CircuitBreakerThreshold: 100,              // 更高的断路器阈值
		SamplingRate:            50,               // 更低的采样频率
		CleanupInterval:         1 * time.Hour,    // 更长的清理间隔
	}
	
	prodLogger := utils.NewSmartLogger(prodConfig)
	
	fmt.Println("生产环境智能日志配置:")
	fmt.Printf("- 时间窗口: %s\n", prodConfig.TimeWindow)
	fmt.Printf("- 详细阈值: %d次后进入静默模式\n", prodConfig.VerboseThreshold)
	fmt.Printf("- 断路器阈值: %d次后停止记录\n", prodConfig.CircuitBreakerThreshold)
	fmt.Printf("- 采样率: 每%d次记录一次\n", prodConfig.SamplingRate)
	
	// 模拟生产环境使用
	prodLogger.LogInfo("Production smart logger initialized")
	
	// 模拟一个典型的生产错误
	for i := 0; i < 5; i++ {
		prodLogger.LogError("External API timeout", map[string]interface{}{
			"api_endpoint": "https://external-service.com/api/data",
			"timeout":      "30s",
			"retry_count":  i,
		})
	}
	
	fmt.Println("\n生产环境示例完成 - 错误被智能聚合处理")
}