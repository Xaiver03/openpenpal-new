package config

import (
	"log"
	"openpenpal-backend/internal/models"

	"gorm.io/gorm"
)

// MigrateExtendedModels 迁移扩展模型
func MigrateExtendedModels(db *gorm.DB) error {
	log.Println("Starting extended model migration...")
	// 博物馆扩展模型
	log.Println("Migrating museum extended models...")
	err := db.AutoMigrate(
		&models.MuseumTag{},
		&models.MuseumInteraction{},
		&models.MuseumReaction{},
		&models.MuseumSubmission{},
		&models.MuseumExhibitionEntry{},
	)
	if err != nil {
		log.Printf("Museum extended models migration error: %v", err)
		return err
	}
	log.Println("Museum extended models migrated successfully")

	// 信件扩展模型
	log.Println("Migrating letter extended models...")
	
	// 先迁移表结构
	err = db.AutoMigrate(&models.LetterTemplate{})
	if err != nil {
		log.Printf("Letter template migration error: %v", err)
		return err
	}
	
	// 然后更新现有的 letter_templates 表的 content 字段
	log.Println("Updating existing letter templates with content from content_template...")
	if err := db.Exec("UPDATE letter_templates SET content = COALESCE(content_template, name) WHERE content IS NULL OR content = ''").Error; err != nil {
		log.Printf("Warning: Could not update existing letter templates: %v", err)
	}
	
	err = db.AutoMigrate(
		&models.LetterLike{},
		&models.LetterShare{},
	)
	if err != nil {
		log.Printf("Letter extended models migration error: %v", err)
		return err
	}
	log.Println("Letter extended models migrated successfully")

	// 创建默认模板
	log.Println("Creating default templates...")
	createDefaultTemplates(db)
	log.Println("Extended migration completed successfully")

	return nil
}

// createDefaultTemplates 创建默认信件模板
func createDefaultTemplates(db *gorm.DB) {
	templates := []models.LetterTemplate{
		{
			ID:          "template_greeting",
			Name:        "温馨问候",
			Description: "适合日常问候和关怀",
			Category:    "greeting",
			ContentTemplate: `亲爱的朋友：

好久不见，近来可好？

[在这里写下你的问候和关怀...]

祝好！
[你的名字]`,
			StyleConfig: `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`,
			IsPremium:  false,
			IsActive:   true,
			UsageCount: 0,
			Rating:     4.5,
		},
		{
			ID:          "template_thanks",
			Name:        "感谢信",
			Description: "表达感激之情",
			Category:    "thanks",
			ContentTemplate: `亲爱的[收信人]：

我想对你说声谢谢。

[在这里写下你要感谢的具体事情...]

你的善意让我深受感动，谢谢你！

此致
敬礼

[你的名字]`,
			StyleConfig: `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`,
			IsPremium:  false,
			IsActive:   true,
			UsageCount: 0,
			Rating:     4.8,
		},
		{
			ID:          "template_apology",
			Name:        "道歉信",
			Description: "真诚地表达歉意",
			Category:    "apology",
			ContentTemplate: `亲爱的[收信人]：

关于[事件]，我想向你道歉。

[详细说明情况和你的歉意...]

希望你能原谅我，我会努力改正。

真诚的
[你的名字]`,
			StyleConfig: `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`,
			IsPremium:  false,
			IsActive:   true,
			UsageCount: 0,
			Rating:     4.6,
		},
		{
			ID:          "template_invitation",
			Name:        "邀请函",
			Description: "邀请参加活动",
			Category:    "invitation",
			ContentTemplate: `亲爱的[收信人]：

诚挚邀请你参加[活动名称]。

时间：[日期时间]
地点：[地点]
活动内容：[简要说明]

期待你的到来！

[你的名字]`,
			StyleConfig: `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`,
			IsPremium:  false,
			IsActive:   true,
			UsageCount: 0,
			Rating:     4.7,
		},
		{
			ID:          "template_love",
			Name:        "情书",
			Description: "表达爱意的浪漫信件",
			Category:    "love",
			ContentTemplate: `亲爱的[名字]：

有些话在心里藏了很久，今天想通过这封信告诉你。

[表达你的感情...]

无论结果如何，能够认识你是我的幸运。

爱你的
[你的名字]`,
			StyleConfig: `{"fontFamily":"serif","fontSize":"16px","color":"#d63384"}`,
			IsPremium:  true,
			IsActive:   true,
			UsageCount: 0,
			Rating:     4.9,
		},
	}

	// 获取系统管理员用户ID
	var adminUser models.User
	if err := db.Where("role = ? OR username = ?", "super_admin", "admin").First(&adminUser).Error; err != nil {
		log.Printf("Warning: Cannot find admin user for templates: %v", err)
		return
	}

	for _, template := range templates {
		// 设置创建者为管理员
		template.CreatedBy = adminUser.ID
		template.Content = template.ContentTemplate // 设置content字段
		
		// 检查是否已存在
		var existing models.LetterTemplate
		if err := db.Where("id = ?", template.ID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&template).Error; err != nil {
				log.Printf("Warning: Could not create template %s: %v", template.ID, err)
			}
		}
	}
}