package migrations

import (
	"log"
	"openpenpal-backend/internal/config"
	
	"gorm.io/gorm"
)

// ApplyIndexOptimization 应用索引优化
func ApplyIndexOptimization(db *gorm.DB) error {
	log.Println("Applying PostgreSQL index optimization migration...")
	
	// 创建索引优化器
	optimizer := config.NewIndexOptimizer(db, false, true)
	
	// 执行优化
	if err := optimizer.OptimizeAll(); err != nil {
		return err
	}
	
	log.Println("Index optimization migration completed")
	return nil
}

// RollbackIndexOptimization 回滚索引优化
func RollbackIndexOptimization(db *gorm.DB) error {
	log.Println("Rolling back index optimization...")
	
	// 需要删除的索引列表
	indexesToDrop := []string{
		// Users
		"idx_users_school_role_active",
		"idx_users_created_at_desc",
		
		// Letters
		"idx_letters_user_status_created",
		"idx_letters_recipient_status",
		"idx_letters_deleted_at",
		"idx_letters_fulltext",
		
		// Letter Codes
		"idx_letter_codes_code_status",
		"idx_letter_codes_status_created",
		
		// Courier Tasks
		"idx_courier_tasks_courier_status",
		"idx_courier_tasks_pickup_delivery",
		"idx_courier_tasks_assigned_at",
		
		// Signal Codes
		"idx_signal_codes_prefix_lookup",
		"idx_signal_codes_school_area",
		
		// Museum
		"idx_museum_items_featured",
		"idx_museum_items_user_public",
		"idx_museum_items_fulltext",
		
		// Notifications
		"idx_notifications_user_unread",
		"idx_notifications_type_created",
		
		// Analytics
		"idx_analytics_time_type",
		"idx_user_analytics_period",
		
		// Credits
		"idx_credit_trans_user_time",
		"idx_credit_activities_active",
		
		// Comments
		"idx_comments_fulltext",
	}
	
	optimizer := config.NewIndexOptimizer(db, false, true)
	
	for _, indexName := range indexesToDrop {
		if err := optimizer.DropIndex(indexName, true); err != nil {
			log.Printf("Warning: Failed to drop index %s: %v", indexName, err)
			// 继续删除其他索引
		}
	}
	
	log.Println("Index optimization rollback completed")
	return nil
}

// IndexMigration 索引迁移结构
type IndexMigration struct {
	ID      uint   `gorm:"primaryKey"`
	Version string `gorm:"uniqueIndex"`
	Applied bool
}

// RegisterIndexMigration 注册索引迁移
func RegisterIndexMigration(db *gorm.DB) error {
	// 创建迁移记录表
	if err := db.AutoMigrate(&IndexMigration{}); err != nil {
		return err
	}
	
	// 检查是否已应用
	var migration IndexMigration
	result := db.Where("version = ?", "20250118_index_optimization").First(&migration)
	
	if result.Error == gorm.ErrRecordNotFound {
		// 应用迁移
		if err := ApplyIndexOptimization(db); err != nil {
			return err
		}
		
		// 记录迁移
		migration = IndexMigration{
			Version: "20250118_index_optimization",
			Applied: true,
		}
		return db.Create(&migration).Error
	}
	
	return nil
}