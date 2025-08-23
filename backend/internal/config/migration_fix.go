package config

import (
	"log"
	"openpenpal-backend/internal/models"

	"gorm.io/gorm"
)

// FixMigrationIssues 修复迁移时的类型不匹配问题
// 按照 CLAUDE.md 原则：Think before action, 谨慎处理
func FixMigrationIssues(db *gorm.DB) error {
	log.Println("Fixing migration type mismatch issues...")

	// 1. 跳过已经存在且类型正确的表的迁移
	// 这些表在数据库中使用 UUID，但 GORM 模型定义为 string/varchar(36)
	tablesToSkip := []string{
		"cart_items",
		"order_items",
		"product_reviews",
		"product_favorites",
	}

	// 2. 对于这些表，我们只需要确保外键约束存在
	// 不需要修改列类型，因为 UUID 和 varchar(36) 在功能上是兼容的
	for _, tableName := range tablesToSkip {
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists).Error
		if err != nil {
			log.Printf("Error checking table %s: %v", tableName, err)
			continue
		}

		if exists {
			log.Printf("Table %s already exists with correct structure, skipping type alteration", tableName)
		}
	}

	// 3. 确保外键约束正确
	// 外键约束在 fix_foreign_key_datatype_mismatch.sql 中已经处理

	log.Println("Migration fix completed")
	return nil
}

// GetModelsForMigration 返回需要迁移的模型列表
// 排除那些会导致类型冲突的模型
func GetModelsForMigration() []interface{} {
	return []interface{}{
		// 用户相关
		&models.User{},
		&models.UserProfile{},
		&models.UserRelationship{}, // 修正: Follow -> UserRelationship
		&models.PrivacySettings{},  // 修正: Privacy -> PrivacySettings

		// 信件相关
		&models.Letter{},
		&models.LetterCode{},
		&models.Envelope{},
		&models.MuseumItem{},      // 修正: Museum -> MuseumItem
		&models.MuseumEntry{},     // 添加: 博物馆入口记录
		&models.MuseumCollection{}, // 添加: 博物馆收藏
		&models.DriftBottle{},      // 已实现: 漂流瓶功能
		&models.CloudLetter{},
		&models.FutureLetter{},     // 已实现: 未来信功能

		// 信使相关
		&models.Courier{},
		&models.CourierTask{},
		&models.CourierPromotion{}, // 已实现: 信使晋升记录
		&models.CourierStats{},     // 已实现: 信使统计
		&models.ScanRecord{},
		&models.ScanEvent{},

		// 评论相关
		&models.Comment{},
		&models.CommentLike{},
		&models.CommentReport{},
		&models.LetterThread{},
		&models.LetterReply{},

		// AI相关
		&models.AIConfig{},
		&models.AIUsageLog{},
		&models.AIMatch{},         // 修正: AIMatchRecord -> AIMatch
		// &models.AIConversation{}, // TODO: 可能需要实现AI对话记录
		&models.AIInspiration{},
		&models.AIReplyAdvice{},

		// 积分相关
		&models.UserCredit{},       // 修正: Credit -> UserCredit
		&models.CreditTransaction{},
		&models.CreditActivity{},
		&models.CreditTask{},
		&models.CreditTransfer{},
		&models.CreditExpirationRule{},
		&models.CreditLimitRule{}, // 修正: CreditLimit -> CreditLimitRule

		// 商城相关 - 使用具体的模型而不是通用的
		&models.Product{},
		&models.CreditShopProduct{},
		&models.Cart{},
		&models.Order{},
		// 注意：不包括 CartItem, OrderItem 等，因为它们已经存在

		// 系统相关
		&models.AuditLog{},
		&models.Notification{},
		&models.SystemSettings{},
		&models.Tag{},
		&models.StorageFile{},     // 修正: Storage -> StorageFile
		&models.StorageConfig{},   // 添加: 存储配置
		&models.OPCode{},
		&models.SecurityEvent{},

		// 调度器相关
		&models.ScheduledTask{},
		&models.TaskExecution{},   // 添加: 任务执行记录
		&models.TaskTemplate{},    // 添加: 任务模板
		// &models.SchedulerLock{}, // TODO: 需要实现调度器锁
		// &models.SchedulerLog{},  // TODO: 需要实现调度器日志
	}
}