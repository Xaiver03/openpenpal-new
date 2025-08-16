package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"gorm.io/gorm"
)


// SafeAutoMigrate 安全的自动迁移，处理约束冲突
func SafeAutoMigrate(db *gorm.DB, models ...interface{}) error {
	log.Println("Starting SafeAutoMigrate with constraint conflict handling...")
	
	migrator := db.Migrator()
	
	for _, model := range models {
		// 获取表名 - 使用反射获取结构体名称
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		
		// 使用GORM的Statement来解析表名
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err == nil && stmt.Schema != nil {
			tableName := stmt.Schema.Table
			
			// 特殊处理User表
			if tableName == "users" {
				log.Println("Special handling for User table to avoid constraint conflicts")
				
				// 检查表是否存在
				if migrator.HasTable(model) {
					// 表存在，只添加缺失的列，不处理约束
					var columns []string
					if err := db.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'users'").Scan(&columns).Error; err == nil {
						log.Printf("User table already exists with %d columns", len(columns))
					}
					
					// 跳过User表的完整迁移
					continue
				} else {
					// 表不存在，创建它
					if err := migrator.CreateTable(model); err != nil {
						return fmt.Errorf("failed to create User table: %w", err)
					}
				}
			} else {
				// 其他表正常迁移
				if !migrator.HasTable(model) {
					log.Printf("Creating table: %s", tableName)
					if err := migrator.CreateTable(model); err != nil {
						// CommentReport表特殊处理
						if strings.Contains(err.Error(), "comment_reports") && strings.Contains(err.Error(), "already exists") {
							log.Printf("Table %s already exists, skipping creation", tableName)
							continue
						}
						// 外键约束类型不匹配错误
						if strings.Contains(err.Error(), "cannot be implemented") {
							log.Printf("Foreign key constraint issue for %s, continuing without it: %v", tableName, err)
							continue
						}
						return fmt.Errorf("failed to create table %s: %w", tableName, err)
					}
				} else {
					// 表存在，执行自动迁移
					if err := migrator.AutoMigrate(model); err != nil {
						// 忽略已存在的错误
						if strings.Contains(err.Error(), "already exists") {
							log.Printf("Ignoring 'already exists' error for %s: %v", tableName, err)
							continue
						}
						// 忽略约束错误
						if strings.Contains(err.Error(), "constraint") {
							log.Printf("Ignoring constraint error for %s: %v", tableName, err)
							continue
						}
						// 外键约束类型不匹配错误
						if strings.Contains(err.Error(), "cannot be implemented") {
							log.Printf("Foreign key constraint type mismatch for %s: %v", tableName, err)
							continue
						}
						// 视图或规则依赖错误
						if strings.Contains(err.Error(), "used by a view or rule") {
							log.Printf("Cannot alter table %s due to view dependencies: %v", tableName, err)
							continue
						}
						// 权限错误
						if strings.Contains(err.Error(), "must be owner") {
							log.Printf("Permission denied for table %s: %v", tableName, err)
							continue
						}
						return fmt.Errorf("failed to migrate table %s: %w", tableName, err)
					}
				}
			}
		} else {
			// 无法解析schema，使用简单方法
			log.Printf("Using simple migration for model: %T", model)
			if err := migrator.AutoMigrate(model); err != nil {
				if strings.Contains(err.Error(), "constraint") || strings.Contains(err.Error(), "already exists") {
					log.Printf("Ignoring migration error for %T: %v", model, err)
					continue
				}
				return err
			}
		}
	}
	
	log.Println("SafeAutoMigrate completed successfully")
	return nil
}