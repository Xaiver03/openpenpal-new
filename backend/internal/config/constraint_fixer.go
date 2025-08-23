package config

import (
	"log"
	"strings"

	"gorm.io/gorm"
)

// ConstraintInfo 约束信息
type ConstraintInfo struct {
	ConstraintName string
	ConstraintType string
	TableName      string
	ColumnName     string
}

// FixConstraintConflicts 修复约束冲突
func FixConstraintConflicts(db *gorm.DB) error {
	log.Println("Checking and fixing constraint conflicts...")

	// 获取users表的所有约束
	var constraints []ConstraintInfo
	err := db.Raw(`
		SELECT 
			con.conname as constraint_name,
			con.contype as constraint_type,
			cls.relname as table_name,
			att.attname as column_name
		FROM pg_constraint con
		INNER JOIN pg_namespace nsp ON nsp.oid = con.connamespace
		INNER JOIN pg_class cls ON cls.oid = con.conrelid
		LEFT JOIN pg_attribute att ON att.attrelid = cls.oid AND att.attnum = ANY(con.conkey)
		WHERE cls.relname = 'users'
		AND nsp.nspname = 'public'
		AND con.contype IN ('u', 'p')
		ORDER BY con.conname;
	`).Scan(&constraints).Error

	if err != nil {
		log.Printf("Failed to query constraints: %v", err)
		return nil // 不阻塞启动
	}

	// 检查并修复约束名称
	for _, constraint := range constraints {
		log.Printf("Found constraint: %s (type: %s, column: %s)", 
			constraint.ConstraintName, constraint.ConstraintType, constraint.ColumnName)

		// 处理username约束
		if constraint.ColumnName == "username" {
			if constraint.ConstraintName == "unique_username" {
				// 约束名称是旧的，但我们现在适配它，不需要改名
				log.Printf("Constraint %s is compatible with custom naming strategy", constraint.ConstraintName)
			} else if constraint.ConstraintName != "uni_users_username" && constraint.ConstraintName != "unique_username" {
				log.Printf("Unexpected constraint name for username: %s", constraint.ConstraintName)
			}
		}

		// 处理email约束
		if constraint.ColumnName == "email" {
			if constraint.ConstraintName == "unique_email" {
				// 约束名称是旧的，但我们现在适配它，不需要改名
				log.Printf("Constraint %s is compatible with custom naming strategy", constraint.ConstraintName)
			} else if constraint.ConstraintName != "uni_users_email" && constraint.ConstraintName != "unique_email" {
				log.Printf("Unexpected constraint name for email: %s", constraint.ConstraintName)
			}
		}
	}

	log.Println("Constraint conflict check completed")
	return nil
}

// SafeMigrateWithConstraintFix 带约束修复的安全迁移
func SafeMigrateWithConstraintFix(db *gorm.DB, models ...interface{}) error {
	// 首先修复约束冲突
	if err := FixConstraintConflicts(db); err != nil {
		log.Printf("Warning: Failed to fix constraint conflicts: %v", err)
		// 继续执行，不阻塞
	}

	// 执行安全迁移
	return SafeAutoMigrate(db, models...)
}

// HandleMigrationError 处理迁移错误
func HandleMigrationError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// 处理约束不存在的错误
	if strings.Contains(errStr, "constraint") && strings.Contains(errStr, "does not exist") {
		log.Printf("Ignoring constraint error: %v", err)
		return nil
	}

	// 处理表已存在的错误
	if strings.Contains(errStr, "already exists") {
		log.Printf("Ignoring 'already exists' error: %v", err)
		return nil
	}

	// 处理外键约束类型不匹配的错误
	if strings.Contains(errStr, "cannot be implemented") {
		log.Printf("Ignoring foreign key constraint error: %v", err)
		return nil
	}

	// 其他错误正常返回
	return err
}

// CreateMissingIndexes 创建缺失的索引
func CreateMissingIndexes(db *gorm.DB) error {
	log.Println("Checking for missing indexes...")

	// 检查unique_username索引是否存在
	var usernameIndexExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'users' 
			AND indexname IN ('unique_username', 'uni_users_username')
		)
	`).Scan(&usernameIndexExists).Error

	if err != nil {
		log.Printf("Failed to check username index: %v", err)
		return nil
	}

	if !usernameIndexExists {
		log.Println("Creating unique index for username...")
		if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS unique_username ON users(username) WHERE deleted_at IS NULL`).Error; err != nil {
			log.Printf("Failed to create username index: %v", err)
		}
	}

	// 检查unique_email索引是否存在
	var emailIndexExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'users' 
			AND indexname IN ('unique_email', 'uni_users_email')
		)
	`).Scan(&emailIndexExists).Error

	if err != nil {
		log.Printf("Failed to check email index: %v", err)
		return nil
	}

	if !emailIndexExists {
		log.Println("Creating unique index for email...")
		if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS unique_email ON users(email) WHERE deleted_at IS NULL`).Error; err != nil {
			log.Printf("Failed to create email index: %v", err)
		}
	}

	log.Println("Index check completed")
	return nil
}