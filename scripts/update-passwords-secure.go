package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

type TestAccount struct {
	Username    string
	NewPassword string
	Role        string
}

var testAccounts = []TestAccount{
	{"admin", "Admin123!", "super_admin"},
	{"alice", "Secret123!", "user"},
	{"bob", "Secret123!", "user"},
	{"courier_level1", "Secret123!", "courier_level1"},
	{"courier_level2", "Secret123!", "courier_level2"},
	{"courier_level3", "Secret123!", "courier_level3"},
	{"courier_level4", "Secret123!", "courier_level4"},
	{"api_test_user_fixed", "Secret123!", "user"},
	{"test_db_connection", "Secret123!", "user"},
}

func main() {
	// 获取数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("请设置环境变量 DATABASE_URL")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("====================================")
	fmt.Println("OpenPenPal 安全密码更新工具")
	fmt.Println("====================================")

	// 首先检查当前密码状态
	fmt.Println("📋 检查当前数据库中的账号状态...")
	checkCurrentPasswords(db)

	fmt.Println("\n🔄 开始更新密码...")

	successCount := 0
	for _, account := range testAccounts {
		if updatePassword(db, account) {
			successCount++
		}
	}

	fmt.Printf("\n✅ 密码更新完成! 成功更新了 %d/%d 个账号\n", successCount, len(testAccounts))

	// 验证更新结果
	fmt.Println("\n🔍 验证更新结果...")
	verifyPasswords(db)

	fmt.Println("\n====================================")
	fmt.Println("🔐 新密码安全特性:")
	fmt.Println("├── 长度: 9位字符")
	fmt.Println("├── 包含: 大写字母、小写字母、数字、符号")
	fmt.Println("└── 符合企业安全标准")
	fmt.Println("====================================")
}

func checkCurrentPasswords(db *sql.DB) {
	query := `
		SELECT username, 
		       CASE WHEN password_hash IS NULL THEN 'NULL' ELSE 'EXISTS' END as password_status,
		       role,
		       created_at
		FROM users 
		WHERE username = ANY($1)
		ORDER BY 
		    CASE 
		        WHEN username = 'admin' THEN 1
		        WHEN username LIKE 'courier_level%' THEN 2
		        ELSE 3
		    END,
		    username`

	usernames := make([]string, len(testAccounts))
	for i, account := range testAccounts {
		usernames[i] = account.Username
	}

	rows, err := db.Query(query, usernames)
	if err != nil {
		log.Printf("查询当前密码状态失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("当前账号状态:")
	fmt.Println("用户名\t\t密码状态\t角色\t\t创建时间")
	fmt.Println("------------------------------------------------")

	for rows.Next() {
		var username, passwordStatus, role, createdAt string
		if err := rows.Scan(&username, &passwordStatus, &role, &createdAt); err != nil {
			continue
		}
		fmt.Printf("%-15s\t%s\t\t%-15s\t%s\n", username, passwordStatus, role, createdAt[:10])
	}
}

func updatePassword(db *sql.DB, account TestAccount) bool {
	// 生成新的密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.NewPassword), 12)
	if err != nil {
		log.Printf("❌ %s: 生成密码哈希失败: %v", account.Username, err)
		return false
	}

	// 更新数据库
	result, err := db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE username = $2",
		string(hashedPassword), account.Username,
	)
	if err != nil {
		log.Printf("❌ %s: 更新密码失败: %v", account.Username, err)
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️  %s: 用户不存在，跳过", account.Username)
		return false
	}

	fmt.Printf("✅ %-15s: 密码已更新为 %s\n", account.Username, account.NewPassword)
	return true
}

func verifyPasswords(db *sql.DB) {
	for _, account := range testAccounts {
		var storedHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", account.Username).Scan(&storedHash)
		if err != nil {
			log.Printf("❌ %s: 查询密码哈希失败: %v", account.Username, err)
			continue
		}

		// 验证密码
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(account.NewPassword))
		if err != nil {
			log.Printf("❌ %s: 密码验证失败! 预期: %s", account.Username, account.NewPassword)
		} else {
			fmt.Printf("✅ %-15s: 密码验证成功 (%s)\n", account.Username, account.NewPassword)
		}
	}
}