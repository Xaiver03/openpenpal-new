package main

import (
	"fmt"
	"openpenpal-backend/internal/middleware"
)

func main() {
	// 测试原始的snakeToCamelCase函数
	testCases := []string{
		"school_code",
		"is_active", 
		"created_at",
		"user_id",
		"like_count",
	}
	
	fmt.Println("=== 原始转换测试 ===")
	for _, test := range testCases {
		fmt.Printf("%s -> %s\n", test, middleware.SnakeToCamelCase(test))
	}
	
	fmt.Println("\n=== 智能中间件将保留以下字段为snake_case ===")
	whitelistFields := []string{
		"school_code", "is_active", "created_at", "updated_at",
		"user_id", "like_count", "share_count",
	}
	
	for _, field := range whitelistFields {
		fmt.Printf("- %s (保持不变)\n", field)
	}
	
	fmt.Println("\n=== 其他字段将转换为camelCase ===")
	otherFields := []string{
		"some_other_field",
		"test_field_name",
		"random_attribute",
	}
	
	for _, field := range otherFields {
		fmt.Printf("%s -> %s\n", field, middleware.SnakeToCamelCase(field))
	}
}

// 提供给中间件使用的函数
func SnakeToCamelCase(s string) string {
	return middleware.SnakeToCamelCase(s)
}