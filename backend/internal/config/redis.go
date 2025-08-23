package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// SetupRedis 初始化Redis连接
func SetupRedis(cfg *Config) (*redis.Client, error) {
	// 默认Redis配置
	redisAddr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")
	redisDBStr := getEnvOrDefault("REDIS_DB", "0")
	
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB value: %s, using default 0", redisDBStr)
		redisDB = 0
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis connected successfully: %s", redisAddr)
	return client, nil
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}