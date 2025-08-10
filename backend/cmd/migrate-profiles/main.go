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

	log.Println("Running user profile migrations...")

	// 自动迁移新的表
	if err := db.AutoMigrate(
		&models.UserProfileExtended{},
		&models.UserStatsData{},
		&models.UserPrivacy{},
		&models.UserAchievement{},
	); err != nil {
		log.Fatal("Failed to migrate tables:", err)
	}

	log.Println("Creating default data for existing users...")

	// 获取所有用户
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	// 为每个用户创建默认档案数据
	for _, user := range users {
		// 创建扩展档案
		profile := models.UserProfileExtended{
			UserID:       user.ID,
			WritingLevel: 1,
			CourierLevel: getCourierLevelFromRole(user.Role),
		}
		db.FirstOrCreate(&profile, models.UserProfileExtended{UserID: user.ID})

		// 创建统计数据
		stats := models.UserStatsData{
			UserID: user.ID,
		}
		db.FirstOrCreate(&stats, models.UserStatsData{UserID: user.ID})

		// 创建隐私设置
		privacy := models.UserPrivacy{
			UserID:         user.ID,
			ShowEmail:      false,
			ShowOPCode:     true,
			ShowStats:      true,
			OPCodePrivacy:  models.OPCodePrivacyPartial,
			ProfileVisible: true,
		}
		db.FirstOrCreate(&privacy, models.UserPrivacy{UserID: user.ID})

		log.Printf("Created profile data for user: %s", user.Username)
	}

	// 为特定用户添加测试数据
	var alice models.User
	if err := db.Where("username = ?", "alice").First(&alice).Error; err == nil {
		// 更新alice的档案
		db.Model(&models.UserProfileExtended{}).Where("user_id = ?", alice.ID).Updates(map[string]interface{}{
			"bio":           "爱好写信的学生，希望通过文字传递温暖",
			"school":        "北京大学",
			"op_code":       "PK5F3D",
			"writing_level": 3,
		})

		// 更新alice的统计
		db.Model(&models.UserStatsData{}).Where("user_id = ?", alice.ID).Updates(map[string]interface{}{
			"letters_sent":         15,
			"letters_received":     12,
			"museum_contributions": 3,
			"total_points":         450,
			"writing_points":       320,
			"current_streak":       7,
		})

		// 添加成就
		achievements := []struct {
			Code        string
			Name        string
			Description string
			Icon        string
			Category    string
		}{
			{"first_letter", "初次来信", "发送第一封信", "✉️", "writing"},
			{"active_writer", "活跃写手", "发送10封信", "✍️", "writing"},
			{"museum_contributor", "博物馆贡献者", "贡献第一封信到博物馆", "🏛️", "museum"},
		}

		for _, ach := range achievements {
			achievement := models.UserAchievement{
				UserID:      alice.ID,
				Code:        ach.Code,
				Name:        ach.Name,
				Description: ach.Description,
				Icon:        ach.Icon,
				Category:    ach.Category,
			}
			db.FirstOrCreate(&achievement, models.UserAchievement{UserID: alice.ID, Code: ach.Code})
		}

		log.Println("Added test data for alice")
	}

	// 为admin用户添加测试数据
	var admin models.User
	if err := db.Where("username = ?", "admin").First(&admin).Error; err == nil {
		// 更新admin的档案
		db.Model(&models.UserProfileExtended{}).Where("user_id = ?", admin.ID).Updates(map[string]interface{}{
			"bio":           "OpenPenPal系统管理员，维护平台运行",
			"school":        "北京大学",
			"op_code":       "PK1L01",
			"writing_level": 5,
			"courier_level": 4,
		})

		// 更新admin的统计
		db.Model(&models.UserStatsData{}).Where("user_id = ?", admin.ID).Updates(map[string]interface{}{
			"letters_sent":         5,
			"letters_received":     8,
			"museum_contributions": 10,
			"total_points":         1000,
			"writing_points":       600,
			"courier_points":       400,
			"current_streak":       30,
		})

		// 添加成就
		achievements := []struct {
			Code        string
			Name        string
			Description string
			Icon        string
			Category    string
		}{
			{"system_admin", "系统管理员", "系统管理员权限", "👤", "system"},
			{"master_writer", "大师写手", "发送100封信", "🏆", "writing"},
			{"city_coordinator", "城市总代", "成为四级信使", "🌆", "courier"},
			{"museum_curator", "博物馆策展人", "贡献10封信到博物馆", "🎨", "museum"},
		}

		for _, ach := range achievements {
			achievement := models.UserAchievement{
				UserID:      admin.ID,
				Code:        ach.Code,
				Name:        ach.Name,
				Description: ach.Description,
				Icon:        ach.Icon,
				Category:    ach.Category,
			}
			db.FirstOrCreate(&achievement, models.UserAchievement{UserID: admin.ID, Code: ach.Code})
		}

		log.Println("Added test data for admin")
	}

	log.Println("User profile migration completed successfully!")
}

func getCourierLevelFromRole(role models.UserRole) int {
	switch role {
	case models.RoleCourierLevel1:
		return 1
	case models.RoleCourierLevel2:
		return 2
	case models.RoleCourierLevel3:
		return 3
	case models.RoleCourierLevel4:
		return 4
	default:
		return 0
	}
}