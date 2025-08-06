package utils

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化Redis客户端
func InitRedis(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}

	client := redis.NewClient(opt)

	// 测试连接
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Redis connected successfully")
	return client
}