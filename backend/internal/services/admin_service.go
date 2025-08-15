package services

import (
	"fmt"
	"log"
	"math/rand"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	// 活跃信使数量（包含所有级别的信使）
	courierRoles := []string{"courier", "courier_level1", "courier_level2", "courier_level3", "courier_level4"}
	if err := s.db.Model(&models.User{}).Where("role IN ? AND is_active = ?", courierRoles, true).Count(&stats.ActiveCouriers).Error; err != nil {
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

		// 为每封信生成唯一编号
		letterCode := &models.LetterCode{
			ID:         uuid.New().String(),
			LetterID:   letter.ID,
			Code:       fmt.Sprintf("OPP%d%d", time.Now().Unix(), len(letters)*1000+rand.Intn(1000)),
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
			SchoolCode:   "BJDX01",
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
			SchoolCode:   "BJDX01",
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
			SchoolCode:   "BJDX01",
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

	// 创建分析数据种子 - 缺失的关键数据
	if err := s.createAnalyticsSeedData(tx, users); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create analytics seed data: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit seed data: %w", err)
	}

	return nil
}

// Analytics seed data constants - 修复硬编码问题
const (
	AnalyticsDaysHistory = 30
	PerformanceMetricsCount = 100
	MaxLettersSentPerDay = 5
	MaxLettersReceivedPerDay = 3
	MaxSessionDurationMinutes = 70
	MinSessionDurationMinutes = 10
	BaseEngagementScore = 25
	MaxEngagementBonus = 50
	CourierEngagementBonus = 20
)

// createAnalyticsSeedData 创建分析数据种子 - 解决分析表空数据问题
func (s *AdminService) createAnalyticsSeedData(tx *gorm.DB, users []models.User) error {
	log.Printf("Creating analytics seed data...")
	
	// 修复：初始化随机种子确保数据随机性
	rand.Seed(time.Now().UnixNano())
	
	// 修复：检查是否已存在分析数据
	var existingCount int64
	if err := tx.Model(&models.UserAnalytics{}).Count(&existingCount).Error; err != nil {
		return fmt.Errorf("failed to check existing analytics data: %w", err)
	}
	if existingCount > 0 {
		log.Printf("Analytics data already exists (%d records), skipping creation", existingCount)
		return nil
	}

	// 创建过去30天的用户分析数据
	now := time.Now()
	totalLettersSent := 0
	totalLettersReceived := 0
	
	for i := 0; i < AnalyticsDaysHistory; i++ {
		date := now.AddDate(0, 0, -i)
		dayLettersSent := 0
		dayLettersReceived := 0
		
		for _, user := range users {
			// 修复：更真实的数据生成逻辑
			lettersSent := rand.Intn(MaxLettersSentPerDay) + 1
			lettersReceived := rand.Intn(MaxLettersReceivedPerDay) + 1
			lettersRead := rand.Intn(lettersReceived + 1) // 读取数不能超过接收数
			
			dayLettersSent += lettersSent
			dayLettersReceived += lettersReceived
			
			// 修复：基于实际活动计算参与度分数
			engagementScore := float64(BaseEngagementScore)
			engagementScore += float64(lettersSent) * 5.0    // 发信件加分
			engagementScore += float64(lettersReceived) * 3.0 // 收信件加分
			engagementScore += float64(lettersRead) * 2.0     // 读信件加分
			
			userAnalytics := models.UserAnalytics{
				ID:              uuid.New().String(),
				UserID:          user.ID,
				Date:            date.Truncate(24 * time.Hour),
				LettersSent:     lettersSent,
				LettersReceived: lettersReceived,
				LettersRead:     lettersRead,
				LoginCount:      rand.Intn(2) + 1,
				SessionDuration: rand.Intn((MaxSessionDurationMinutes-MinSessionDurationMinutes)*60) + MinSessionDurationMinutes*60,
				CourierTasks:    0,
				MuseumVisits:    rand.Intn(2),
				EngagementScore: engagementScore,
				RetentionDays:   i + 1,
				CreatedAt:       date,
				UpdatedAt:       date,
			}
			
			// 为信使用户添加信使任务
			if user.Role == "courier" || user.Role == "admin" {
				courierTasks := rand.Intn(10) + 2
				userAnalytics.CourierTasks = courierTasks
				userAnalytics.EngagementScore += CourierEngagementBonus + float64(courierTasks)*2.0
			}
			
			if err := tx.Create(&userAnalytics).Error; err != nil {
				return fmt.Errorf("failed to create user analytics for %s: %w", user.Username, err)
			}
		}
		
		totalLettersSent += dayLettersSent
		totalLettersReceived += dayLettersReceived
	}

	// 修复：创建过去30天的系统分析数据，基于真实用户活动
	for i := 0; i < AnalyticsDaysHistory; i++ {
		date := now.AddDate(0, 0, -i)
		
		// 修复：基于当天实际用户分析数据计算系统指标
		var dayUserAnalytics []models.UserAnalytics
		if err := tx.Where("date = ?", date.Truncate(24*time.Hour)).Find(&dayUserAnalytics).Error; err != nil {
			log.Printf("Warning: Could not fetch user analytics for date %s: %v", date.Format("2006-01-02"), err)
		}
		
		// 基于实际用户活动计算系统指标
		activeUsers := len(dayUserAnalytics)
		if activeUsers == 0 {
			activeUsers = rand.Intn(len(users)) + 1 // 后备方案
		}
		
		totalLettersForDay := 0
		totalTasksForDay := 0
		for _, ua := range dayUserAnalytics {
			totalLettersForDay += ua.LettersSent
			totalTasksForDay += ua.CourierTasks
		}
		
		systemAnalytics := models.SystemAnalytics{
			ID:                    uuid.New().String(),
			Date:                  date.Truncate(24 * time.Hour),
			ActiveUsers:           activeUsers,
			NewUsers:              rand.Intn(3),                        // 0-2个新用户
			TotalUsers:            len(users) + (AnalyticsDaysHistory - i), // 历史递增
			LettersCreated:        totalLettersForDay,                  // 基于实际用户活动
			LettersDelivered:      int(float64(totalLettersForDay) * 0.85), // 85%送达率
			CourierTasksCompleted: totalTasksForDay,                    // 基于实际信使活动
			MuseumItemsAdded:      rand.Intn(2),                        // 0-1个博物馆物品
			AvgResponseTime:       float64(rand.Intn(200) + 100),       // 100-300ms
			ErrorRate:             float64(rand.Intn(5)) / 10.0,        // 0-0.5%
			ServerUptime:          float64(rand.Intn(5)+995) / 10.0,    // 99.5-99.9%
			CreatedAt:             date,
			UpdatedAt:             date,
		}
		
		if err := tx.Create(&systemAnalytics).Error; err != nil {
			return fmt.Errorf("failed to create system analytics: %w", err)
		}
	}

	// 修复：创建性能指标样本数据，更真实的API端点和方法组合
	type EndpointConfig struct {
		Path   string
		Method string
		AvgResponseTime float64
	}
	
	endpoints := []EndpointConfig{
		{"/api/v1/auth/login", "POST", 150.0},
		{"/api/v1/auth/me", "GET", 80.0},
		{"/api/v1/letters", "GET", 120.0},
		{"/api/v1/letters", "POST", 200.0},
		{"/api/v1/admin/dashboard/stats", "GET", 250.0},
		{"/api/v1/museum/entries", "GET", 180.0},
		{"/api/v1/courier/tasks", "GET", 160.0},
		{"/api/v1/notifications", "GET", 90.0},
	}
	
	// 修复：分批创建性能指标，避免事务过大
	batchSize := 20
	for batch := 0; batch < PerformanceMetricsCount/batchSize; batch++ {
		for i := 0; i < batchSize; i++ {
			endpointConfig := endpoints[rand.Intn(len(endpoints))]
			
			// 修复：基于端点类型生成更真实的响应时间
			baseResponseTime := endpointConfig.AvgResponseTime
			responseTime := baseResponseTime + float64(rand.Intn(100)) - 50.0 // ±50ms变化
			if responseTime < 10 {
				responseTime = 10 // 最小响应时间
			}
			
			statusCode := 200
			errorMultiplier := 1.0
			
			// 修复：更真实的错误率分布（5%错误率）
			if rand.Intn(20) == 0 {
				statusCode = []int{400, 401, 403, 404, 500}[rand.Intn(5)]
				errorMultiplier = 2.0 + rand.Float64() // 错误请求响应时间2-3倍
			}
			
			performanceMetric := models.PerformanceMetric{
				ID:           uuid.New().String(),
				Endpoint:     endpointConfig.Path,
				Method:       endpointConfig.Method,
				ResponseTime: responseTime * errorMultiplier,
				StatusCode:   statusCode,
				UserAgent:    "Mozilla/5.0 (OpenPenPal Client/1.0)",
				IPAddress:    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
				UserID:       &users[rand.Intn(len(users))].ID,
				Timestamp:    now.Add(-time.Duration(rand.Intn(86400*AnalyticsDaysHistory)) * time.Second),
				CreatedAt:    now,
			}
			
			if err := tx.Create(&performanceMetric).Error; err != nil {
				return fmt.Errorf("failed to create performance metric: %w", err)
			}
		}
	}

	log.Printf("✅ Analytics seed data created successfully:")
	log.Printf("   - User analytics: %d users × %d days = %d records", len(users), AnalyticsDaysHistory, len(users)*AnalyticsDaysHistory)
	log.Printf("   - System analytics: %d daily records", AnalyticsDaysHistory)
	log.Printf("   - Performance metrics: %d records", PerformanceMetricsCount)
	log.Printf("   - Total analytics records: %d", len(users)*AnalyticsDaysHistory + AnalyticsDaysHistory + PerformanceMetricsCount)
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

// ResetUserPassword 重置用户密码（管理员功能）
func (s *AdminService) ResetUserPassword(userID string, newPassword string) error {
	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 更新用户密码
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", string(hashedPassword))
	if result.Error != nil {
		return fmt.Errorf("重置密码失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// GetLetters 获取信件列表（管理员功能）
func (s *AdminService) GetLetters(page, limit int, filters map[string]interface{}) ([]models.Letter, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	var letters []models.Letter
	var total int64

	// 构建查询
	query := s.db.Model(&models.Letter{}).Preload("User")

	// 应用过滤条件
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if senderID, ok := filters["sender_id"].(string); ok && senderID != "" {
		query = query.Where("user_id = ?", senderID)
	}
	if schoolCode, ok := filters["school_code"].(string); ok && schoolCode != "" {
		query = query.Joins("JOIN users ON letters.user_id = users.id").Where("users.school_code = ?", schoolCode)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count letters: %w", err)
	}

	// 获取分页数据
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get letters: %w", err)
	}

	return letters, total, nil
}

// ModerateLetter 审核信件（管理员功能）
func (s *AdminService) ModerateLetter(letterID, action, reason, notes string) (*models.Letter, error) {
	var letter models.Letter

	// 查找信件
	if err := s.db.First(&letter, "id = ?", letterID).Error; err != nil {
		return nil, fmt.Errorf("信件不存在: %w", err)
	}

	// 根据审核动作更新信件状态
	updates := make(map[string]interface{})
	switch action {
	case "approve":
		updates["status"] = models.StatusApproved
	case "reject":
		updates["status"] = models.StatusRejected
	case "flag":
		updates["status"] = models.StatusFlagged
	case "archive":
		updates["status"] = models.StatusArchived
	default:
		return nil, fmt.Errorf("无效的审核动作: %s", action)
	}

	updates["updated_at"] = time.Now()

	// 这里可以添加审核记录到数据库
	// 创建审核记录
	moderationRecord := models.ModerationRecord{
		ID:           uuid.New().String(),
		ContentType:  "letter",
		ContentID:    letterID,
		Action:       action,
		Reason:       reason,
		Notes:        notes,
		Status:       "completed",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 开始事务
	tx := s.db.Begin()

	// 更新信件状态
	if err := tx.Model(&letter).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新信件状态失败: %w", err)
	}

	// 创建审核记录
	if err := tx.Create(&moderationRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建审核记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交审核失败: %w", err)
	}

	// 重新加载信件信息
	if err := s.db.Preload("User").First(&letter, "id = ?", letterID).Error; err != nil {
		return nil, fmt.Errorf("重新加载信件失败: %w", err)
	}

	return &letter, nil
}

// GetCouriers 获取信使列表（管理员功能）
func (s *AdminService) GetCouriers(page, limit int, filters map[string]interface{}) ([]models.Courier, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	var couriers []models.Courier
	var total int64

	// 构建查询
	query := s.db.Model(&models.Courier{})

	// 应用过滤条件
	if level, ok := filters["level"].(int); ok && level > 0 {
		query = query.Where("level = ?", level)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if schoolCode, ok := filters["school_code"].(string); ok && schoolCode != "" {
		query = query.Where("school = ? OR zone LIKE ?", schoolCode, schoolCode+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count couriers: %w", err)
	}

	// 获取分页数据
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&couriers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get couriers: %w", err)
	}

	return couriers, total, nil
}

// GetAppointableRoles 获取可任命的角色列表（管理员功能）
func (s *AdminService) GetAppointableRoles() ([]map[string]interface{}, error) {
	roles := []map[string]interface{}{
		{
			"id":          "user",
			"name":        "user",
			"displayName": "普通用户",
			"description": "平台的普通用户",
			"level":       1,
			"permissions": []string{"letter:create", "letter:read"},
		},
		{
			"id":          "courier_level1",
			"name":        "courier_level1",
			"displayName": "一级信使",
			"description": "楼栋级别的信使",
			"level":       2,
			"permissions": []string{"letter:create", "letter:read", "courier:deliver"},
		},
		{
			"id":          "courier_level2",
			"name":        "courier_level2",
			"displayName": "二级信使",
			"description": "片区级别的信使",
			"level":       3,
			"permissions": []string{"letter:create", "letter:read", "courier:deliver", "courier:manage_l1"},
		},
		{
			"id":          "courier_level3",
			"name":        "courier_level3",
			"displayName": "三级信使",
			"description": "学校级别的信使",
			"level":       4,
			"permissions": []string{"letter:create", "letter:read", "courier:deliver", "courier:manage_l1", "courier:manage_l2"},
		},
		{
			"id":          "courier_level4",
			"name":        "courier_level4",
			"displayName": "四级信使",
			"description": "城市级别的信使",
			"level":       5,
			"permissions": []string{"letter:create", "letter:read", "courier:deliver", "courier:manage_l1", "courier:manage_l2", "courier:manage_l3"},
		},
		{
			"id":          "admin",
			"name":        "admin",
			"displayName": "管理员",
			"description": "系统管理员",
			"level":       6,
			"permissions": []string{"admin:all"},
		},
	}

	return roles, nil
}

// AppointUser 任命用户角色（管理员功能）
func (s *AdminService) AppointUser(userID, newRole, reason string, effectiveAt *time.Time, metadata map[string]interface{}) (map[string]interface{}, error) {
	var user models.User

	// 查找用户
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 验证新角色
	validRoles := []string{"user", "courier_level1", "courier_level2", "courier_level3", "courier_level4", "admin"}
	isValidRole := false
	for _, role := range validRoles {
		if newRole == role {
			isValidRole = true
			break
		}
	}
	
	if !isValidRole {
		return nil, fmt.Errorf("无效的角色: %s", newRole)
	}

	// 创建任命记录
	appointmentRecord := map[string]interface{}{
		"id":               uuid.New().String(),
		"userId":           userID,
		"user_name":        user.Username,
		"user_email":       user.Email,
		"old_role":         string(user.Role),
		"new_role":         newRole,
		"reason":           reason,
		"appointed_by":     "admin", // 简化实现，实际应该从上下文获取
		"appointed_by_name": "系统管理员",
		"appointed_at":     time.Now().Format(time.RFC3339),
		"status":           "approved", // 简化实现，直接批准
		"metadata":         metadata,
	}

	if effectiveAt != nil {
		appointmentRecord["effective_at"] = effectiveAt.Format(time.RFC3339)
	}

	// 更新用户角色
	if err := s.db.Model(&user).Update("role", newRole).Error; err != nil {
		return nil, fmt.Errorf("更新用户角色失败: %w", err)
	}

	return appointmentRecord, nil
}

// GetAppointmentRecords 获取任命记录（管理员功能）
func (s *AdminService) GetAppointmentRecords(page, limit int, filters map[string]interface{}) ([]map[string]interface{}, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	// 简化实现：从用户表获取角色变更历史
	// 实际应该有专门的任命记录表
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	// 应用过滤条件
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("id = ?", userID)
	}
	if role, ok := filters["role"].(string); ok && role != "" {
		query = query.Where("role = ?", role)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count appointment records: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get appointment records: %w", err)
	}

	// 构建任命记录格式
	records := make([]map[string]interface{}, len(users))
	for i, user := range users {
		records[i] = map[string]interface{}{
			"id":               user.ID,
			"userId":           user.ID,
			"user_name":        user.Username,
			"user_email":       user.Email,
			"old_role":         "user", // 简化实现
			"new_role":         string(user.Role),
			"reason":           "系统记录",
			"appointed_by":     "admin",
			"appointed_by_name": "系统管理员",
			"appointed_at":     user.UpdatedAt.Format(time.RFC3339),
			"status":           "approved",
		}
	}

	return records, total, nil
}

// ReviewAppointment 审批任命申请（管理员功能）
func (s *AdminService) ReviewAppointment(appointmentID, status, notes string) (map[string]interface{}, error) {
	// 简化实现：查找用户并更新状态
	var user models.User
	if err := s.db.First(&user, "id = ?", appointmentID).Error; err != nil {
		return nil, fmt.Errorf("任命记录不存在: %w", err)
	}

	// 构建审批后的记录
	appointment := map[string]interface{}{
		"id":               appointmentID,
		"userId":           user.ID,
		"user_name":        user.Username,
		"user_email":       user.Email,
		"old_role":         "user",
		"new_role":         string(user.Role),
		"reason":           "管理员审批",
		"appointed_by":     "admin",
		"appointed_by_name": "系统管理员",
		"appointed_at":     time.Now().Format(time.RFC3339),
		"status":           status,
		"approval_notes":   notes,
	}

	return appointment, nil
}
