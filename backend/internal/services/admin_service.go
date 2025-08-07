package services

import (
	"fmt"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminService 管理后台服务
type AdminService struct {
	db     *gorm.DB
	config *config.Config
}

// NewAdminService 创建管理后台服务实例
func NewAdminService(db *gorm.DB, config *config.Config) *AdminService {
	return &AdminService{
		db:     db,
		config: config,
	}
}

// GetDashboardStats 获取管理后台统计数据
func (s *AdminService) GetDashboardStats() (*models.AdminDashboardStats, error) {
	stats := &models.AdminDashboardStats{}

	// 用户统计
	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 今日新用户
	today := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&stats.NewUsersToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count new users today: %w", err)
	}

	// 信件统计
	if err := s.db.Model(&models.Letter{}).Count(&stats.TotalLetters).Error; err != nil {
		return nil, fmt.Errorf("failed to count letters: %w", err)
	}

	// 今日新信件
	if err := s.db.Model(&models.Letter{}).Where("created_at >= ?", today).Count(&stats.LettersToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count letters today: %w", err)
	}

	// 活跃信使数量
	if err := s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", "courier", true).Count(&stats.ActiveCouriers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active couriers: %w", err)
	}

	// 信件状态分布
	statusCounts := make(map[string]int64)
	rows, err := s.db.Model(&models.Letter{}).Select("status, count(*) as count").Group("status").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get letter status distribution: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		statusCounts[status] = count
	}
	stats.LetterStatusDistribution = statusCounts

	// 博物馆展品统计
	if err := s.db.Model(&models.MuseumItem{}).Count(&stats.MuseumExhibits).Error; err != nil {
		return nil, fmt.Errorf("failed to count museum exhibits: %w", err)
	}

	// 信封订单统计
	if err := s.db.Model(&models.EnvelopeOrder{}).Count(&stats.EnvelopeOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count envelope orders: %w", err)
	}

	// 通知统计
	if err := s.db.Model(&models.Notification{}).Count(&stats.TotalNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// 系统健康状态
	stats.SystemHealth = &models.SystemHealth{
		DatabaseStatus: "healthy",
		ServiceStatus:  "running",
		LastUpdated:    time.Now(),
	}

	return stats, nil
}

// GetRecentActivities 获取最近活动
func (s *AdminService) GetRecentActivities(limit int) ([]models.AdminActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var activities []models.AdminActivity

	// 获取最近的信件活动
	var letters []models.Letter
	if err := s.db.Preload("User").Order("created_at DESC").Limit(limit / 2).Find(&letters).Error; err == nil {
		for _, letter := range letters {
			activities = append(activities, models.AdminActivity{
				ID:          uuid.New().String(),
				Type:        "letter_created",
				Description: fmt.Sprintf("用户 %s 创建了新信件: %s", letter.User.Username, letter.Title),
				UserID:      letter.UserID,
				CreatedAt:   letter.CreatedAt,
			})
		}
	}

	// 获取最近的用户注册
	var users []models.User
	if err := s.db.Order("created_at DESC").Limit(limit / 2).Find(&users).Error; err == nil {
		for _, user := range users {
			activities = append(activities, models.AdminActivity{
				ID:          uuid.New().String(),
				Type:        "user_registered",
				Description: fmt.Sprintf("新用户注册: %s", user.Username),
				UserID:      user.ID,
				CreatedAt:   user.CreatedAt,
			})
		}
	}

	// 按时间排序
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].CreatedAt.Before(activities[j].CreatedAt) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// InjectSeedData 注入种子数据
func (s *AdminService) InjectSeedData() error {
	tx := s.db.Begin()

	// 检查是否已经有种子数据
	var userCount int64
	if err := tx.Model(&models.User{}).Count(&userCount).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check existing users: %w", err)
	}

	if userCount > 5 {
		tx.Rollback()
		return fmt.Errorf("seed data already exists")
	}

	// 创建测试用户
	users := []models.User{
		{
			ID:           uuid.New().String(),
			Username:     "admin",
			Email:        "admin@openpenpal.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Nickname:     "管理员",
			Avatar:       "/avatars/admin.png",
			Role:         "admin",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, -1, 0),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Username:     "courier_alice",
			Email:        "alice@courier.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Nickname:     "爱丽丝",
			Avatar:       "/avatars/alice.png",
			Role:         "courier",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -15),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Username:     "writer_bob",
			Email:        "bob@writer.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Nickname:     "鲍勃",
			Avatar:       "/avatars/bob.png",
			Role:         "user",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -10),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Username:     "reader_carol",
			Email:        "carol@reader.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Nickname:     "卡罗尔",
			Avatar:       "/avatars/carol.png",
			Role:         "user",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -7),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Username:     "student_david",
			Email:        "david@student.com",
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Nickname:     "大卫",
			Avatar:       "/avatars/david.png",
			Role:         "user",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -3),
			UpdatedAt:    time.Now(),
		},
	}

	for _, user := range users {
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}

	// 创建测试信件
	letters := []models.Letter{
		{
			ID:        uuid.New().String(),
			UserID:    users[2].ID, // writer_bob
			Title:     "给未来的自己",
			Content:   "亲爱的未来的我：\n\n当你读到这封信的时候，不知道你是否还记得写下这些文字时的心情。今天是一个平凡的日子，但我想记录下此刻的感受...\n\n希望你过得很好。\n\n此致\n敬礼\n\n过去的你",
			Style:     "formal",
			Status:    models.StatusGenerated,
			ReplyTo:   users[3].ID, // reader_carol
			CreatedAt: time.Now().AddDate(0, 0, -5),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			UserID:    users[3].ID, // reader_carol
			Title:     "感谢信",
			Content:   "亲爱的朋友：\n\n感谢你在我最困难的时候陪伴我，虽然我们可能从未见过面，但你的信件给了我温暖和力量。\n\n这个世界因为有你而更美好。\n\n愿你每天都开心快乐！\n\n你的朋友",
			Style:     "casual",
			Status:    models.StatusDelivered,
			ReplyTo:   users[2].ID, // writer_bob
			CreatedAt: time.Now().AddDate(0, 0, -3),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			UserID:    users[4].ID, // student_david
			Title:     "关于梦想",
			Content:   "每个人都有梦想，但不是每个人都敢于追逐梦想。\n\n我想说，无论你的梦想是什么，都要勇敢地去追求它。即使路上会有挫折，即使别人不理解，但只要你坚持，总有一天会实现的。\n\n加油，为了梦想而奋斗的人！",
			Style:     "inspirational",
			Status:    models.StatusInTransit,
			CreatedAt: time.Now().AddDate(0, 0, -1),
			UpdatedAt: time.Now(),
		},
	}

	for _, letter := range letters {
		if err := tx.Create(&letter).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create letter: %w", err)
		}

		// 为每封信生成编号
		letterCode := &models.LetterCode{
			ID:         uuid.New().String(),
			LetterID:   letter.ID,
			Code:       fmt.Sprintf("OPP%d", time.Now().Unix()+int64(len(letters))),
			QRCodeURL:  fmt.Sprintf("/qr/%s.png", letter.ID),
			QRCodePath: fmt.Sprintf("/uploads/qr/%s.png", letter.ID),
		}
		if err := tx.Create(letterCode).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create letter code: %w", err)
		}
	}

	// 创建信封设计
	envelopeDesigns := []models.EnvelopeDesign{
		{
			ID:           uuid.New().String(),
			SchoolCode:   "BJDX",
			Type:         "school",
			Theme:        "经典白色",
			ImageURL:     "/envelopes/classic-white.png",
			ThumbnailURL: "/envelopes/classic-white-thumb.png",
			CreatorID:    users[0].ID, // admin created
			CreatorName:  users[0].Nickname,
			Description:  "简洁优雅的白色信封，适合正式信件",
			Status:       "approved",
			VoteCount:    25,
			Period:       "2024春季",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, -1, 0),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			SchoolCode:   "BJDX",
			Type:         "school",
			Theme:        "温馨粉色",
			ImageURL:     "/envelopes/warm-pink.png",
			ThumbnailURL: "/envelopes/warm-pink-thumb.png",
			CreatorID:    users[1].ID, // alice created
			CreatorName:  users[1].Nickname,
			Description:  "温柔的粉色信封，传递温暖的情感",
			Status:       "approved",
			VoteCount:    18,
			Period:       "2024春季",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -20),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			SchoolCode:   "BJDX",
			Type:         "school",
			Theme:        "复古牛皮",
			ImageURL:     "/envelopes/vintage-kraft.png",
			ThumbnailURL: "/envelopes/vintage-kraft-thumb.png",
			CreatorID:    users[2].ID, // bob created
			CreatorName:  users[2].Nickname,
			Description:  "复古风格的牛皮纸信封，充满怀旧气息",
			Status:       "approved",
			VoteCount:    32,
			Period:       "2024春季",
			IsActive:     true,
			CreatedAt:    time.Now().AddDate(0, 0, -15),
			UpdatedAt:    time.Now(),
		},
	}

	for _, design := range envelopeDesigns {
		if err := tx.Create(&design).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create envelope design: %w", err)
		}
	}

	// 创建一些信封订单
	envelopeOrders := []models.EnvelopeOrder{
		{
			ID:             uuid.New().String(),
			UserID:         users[2].ID, // writer_bob
			DesignID:       envelopeDesigns[0].ID,
			Quantity:       5,
			TotalPrice:     14.95, // 5 * 2.99
			Status:         "completed",
			PaymentMethod:  "wechat",
			PaymentID:      "wx_pay_123456",
			DeliveryMethod: "pickup",
			DeliveryInfo:   `{"pickup_location": "学生活动中心", "pickup_time": "09:00-17:00"}`,
			CreatedAt:      time.Now().AddDate(0, 0, -8),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New().String(),
			UserID:         users[3].ID, // reader_carol
			DesignID:       envelopeDesigns[1].ID,
			Quantity:       3,
			TotalPrice:     11.97, // 3 * 3.99
			Status:         "pending",
			PaymentMethod:  "alipay",
			DeliveryMethod: "delivery",
			DeliveryInfo:   `{"address": "学生宿舍A栋201", "phone": "13800138000"}`,
			CreatedAt:      time.Now().AddDate(0, 0, -2),
			UpdatedAt:      time.Now(),
		},
	}

	for _, order := range envelopeOrders {
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create envelope order: %w", err)
		}
	}

	// 创建博物馆展品
	exhibits := []models.MuseumItem{
		{
			ID:          uuid.New().String(),
			SourceType:  models.SourceTypeLetter,
			SourceID:    letters[0].ID, // Reference first letter
			Title:       "战争年代的家书",
			Description: "一封来自抗战时期的家书，展现了那个年代人们的坚韧与思念",
			Tags:        "历史文物,战争,家书,1940年代",
			Status:      models.MuseumItemApproved,
			SubmittedBy: users[0].ID, // admin submitted
			ViewCount:   1250,
			LikeCount:   89,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			SourceType:  models.SourceTypeLetter,
			SourceID:    letters[1].ID, // Reference second letter
			Title:       "情书的艺术",
			Description: "20世纪50年代的爱情书信，展现了那个时代纯真的爱情",
			Tags:        "爱情文学,情书,1950年代",
			Status:      models.MuseumItemApproved,
			SubmittedBy: users[0].ID, // admin submitted
			ViewCount:   2150,
			LikeCount:   156,
			CreatedAt:   time.Now().AddDate(0, 0, -45),
			UpdatedAt:   time.Now(),
		},
	}

	for _, exhibit := range exhibits {
		if err := tx.Create(&exhibit).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create museum exhibit: %w", err)
		}
	}

	// 创建通知
	notifications := []models.Notification{
		{
			ID:        uuid.New().String(),
			UserID:    users[2].ID,
			Type:      models.NotificationLetter,
			Channel:   models.ChannelWebSocket,
			Priority:  models.PriorityNormal,
			Title:     "信件已送达",
			Content:   "您的信件已成功送达收件人",
			Status:    models.NotificationSent,
			CreatedAt: time.Now().AddDate(0, 0, -2),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			UserID:    users[3].ID,
			Type:      models.NotificationLetter,
			Channel:   models.ChannelEmail,
			Priority:  models.PriorityHigh,
			Title:     "您有新的信件",
			Content:   "您收到了一封新的手写信件，请及时查看",
			Status:    models.NotificationSent,
			CreatedAt: time.Now().AddDate(0, 0, -1),
			UpdatedAt: time.Now(),
		},
	}

	for _, notification := range notifications {
		if err := tx.Create(&notification).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit seed data: %w", err)
	}

	return nil
}

// GetUserManagement 获取用户管理数据
func (s *AdminService) GetUserManagement(page, limit int) (*models.UserManagementResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	// 获取总数
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取用户列表
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return &models.UserManagementResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// GetSystemSettings 获取系统设置
func (s *AdminService) GetSystemSettings() (*models.AdminSystemSettings, error) {
	settings := &models.AdminSystemSettings{
		SiteName:             "OpenPenPal",
		SiteDescription:      "手写信的温暖传递平台",
		RegistrationOpen:     true,
		MaintenanceMode:      false,
		MaxLettersPerDay:     10,
		MaxEnvelopesPerOrder: 20,
		EmailEnabled:         true,
		SMSEnabled:           false,
		LastUpdated:          time.Now(),
	}

	return settings, nil
}

// UpdateUser 更新用户信息（管理员功能）
func (s *AdminService) UpdateUser(userID string, req *models.AdminUpdateUserRequest) (*models.User, error) {
	var user models.User
	
	// 查找用户
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var count int64
		if err := s.db.Model(&models.User{}).Where("email = ? AND id != ?", req.Email, userID).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if count > 0 {
			return nil, fmt.Errorf("邮箱已被使用")
		}
		updates["email"] = req.Email
	}
	
	if req.Role != "" {
		// 验证角色是否有效
		validRoles := []string{"user", "courier", "courier_level1", "courier_level2", "courier_level3", "courier_level4", "school_admin", "admin", "super_admin"}
		isValidRole := false
		for _, validRole := range validRoles {
			if req.Role == validRole {
				isValidRole = true
				break
			}
		}
		if !isValidRole {
			return nil, fmt.Errorf("无效的角色: %s", req.Role)
		}
		updates["role"] = req.Role
	}
	
	if req.SchoolCode != "" {
		updates["school_code"] = req.SchoolCode
	}
	
	// 更新激活状态
	updates["is_active"] = req.IsActive
	updates["updated_at"] = time.Now()

	// 执行更新
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 重新加载用户信息
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("重新加载用户失败: %w", err)
	}

	return &user, nil
}
