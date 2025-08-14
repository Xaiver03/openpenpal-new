package main

import (
	"log"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// 查找alice用户
	var alice models.User
	if err := db.Where("username = ?", "alice").First(&alice).Error; err != nil {
		log.Fatal("Failed to find alice user:", err)
	}

	log.Printf("Found alice with ID: %s", alice.ID)

	// 更新alice的扩展档案
	profileUpdates := map[string]interface{}{
		"bio":           "爱好写信的学生，希望通过文字传递温暖",
		"school":        "北京大学",
		"op_code":       "PK5F3D",
		"writing_level": 3,
		"courier_level": 0,
	}

	if err := db.Model(&models.UserProfileExtended{}).Where("user_id = ?", alice.ID).Updates(profileUpdates).Error; err != nil {
		log.Printf("Failed to update profile: %v", err)
	} else {
		log.Println("Updated alice profile successfully")
	}

	// 更新alice的统计数据
	statsUpdates := map[string]interface{}{
		"letters_sent":         15,
		"letters_received":     12,
		"museum_contributions": 3,
		"total_points":         450,
		"writing_points":       320,
		"courier_points":       0,
		"current_streak":       7,
		"max_streak":           7,
	}

	if err := db.Model(&models.UserStatsData{}).Where("user_id = ?", alice.ID).Updates(statsUpdates).Error; err != nil {
		log.Printf("Failed to update stats: %v", err)
	} else {
		log.Println("Updated alice stats successfully")
	}

	// 更新alice的隐私设置
	privacyUpdates := map[string]interface{}{
		"show_email":      false,
		"show_op_code":    true,
		"show_stats":      true,
		"op_code_privacy": models.OPCodePrivacyPartial,
	}

	if err := db.Model(&models.UserPrivacy{}).Where("user_id = ?", alice.ID).Updates(privacyUpdates).Error; err != nil {
		log.Printf("Failed to update privacy: %v", err)
	} else {
		log.Println("Updated alice privacy settings successfully")
	}

	log.Println("Alice profile update completed!")
}
