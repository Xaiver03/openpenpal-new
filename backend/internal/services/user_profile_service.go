package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

type UserProfileService struct {
	db *gorm.DB
}

func NewUserProfileService(db *gorm.DB) *UserProfileService {
	return &UserProfileService{db: db}
}

// UserProfileResponse 用户档案响应
type UserProfileResponse struct {
	ID           string                   `json:"id"`
	Username     string                   `json:"username"`
	Nickname     string                   `json:"nickname,omitempty"`
	Email        string                   `json:"email,omitempty"`
	Role         string                   `json:"role"`
	AvatarURL    string                   `json:"avatar_url,omitempty"`
	Bio          string                   `json:"bio,omitempty"`
	School       string                   `json:"school,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	OPCode       string                   `json:"op_code,omitempty"`
	WritingLevel int                      `json:"writing_level"`
	CourierLevel int                      `json:"courier_level"`
	Stats        *models.UserStatsData    `json:"stats,omitempty"`
	Privacy      *models.UserPrivacy      `json:"privacy,omitempty"`
	Achievements []models.UserAchievement `json:"achievements,omitempty"`
}

// GetUserProfile 获取用户档案
func (s *UserProfileService) GetUserProfile(username string, requestingUserID string) (*UserProfileResponse, error) {
	// 查找用户
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	// 检查隐私设置
	var privacy models.UserPrivacy
	s.db.Where("user_id = ?", user.ID).FirstOrCreate(&privacy, models.UserPrivacy{
		UserID:         user.ID,
		ShowEmail:      false,
		ShowOPCode:     true,
		ShowStats:      true,
		OPCodePrivacy:  models.OPCodePrivacyPartial,
		ProfileVisible: true,
	})

	// 如果不是本人且档案不可见，返回错误
	if user.ID != requestingUserID && !privacy.ProfileVisible {
		return nil, fmt.Errorf("该用户的资料未公开")
	}

	// 获取扩展档案
	var profile models.UserProfileExtended
	if err := s.db.Where("user_id = ?", user.ID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果不存在，创建默认档案
			profile = models.UserProfileExtended{
				UserID:       user.ID,
				WritingLevel: 1,
				CourierLevel: s.getCourierLevelFromRole(user.Role),
			}
			s.db.Create(&profile)
		}
	}

	// Debug logging
	fmt.Printf("DEBUG: Loaded profile for user %s: bio=%s, school=%s, op_code=%s, writing_level=%d\n",
		user.Username, profile.Bio, profile.School, profile.OPCode, profile.WritingLevel)

	// 获取统计数据
	var stats models.UserStatsData
	s.db.Where("user_id = ?", user.ID).FirstOrCreate(&stats, models.UserStatsData{
		UserID:         user.ID,
		LastActiveDate: time.Now(),
	})

	// 更新统计数据（实时计算）- 暂时注释掉，避免覆盖手动设置的数据
	// s.updateUserStats(&stats, user.ID)

	// 获取成就
	var achievements []models.UserAchievement
	s.db.Where("user_id = ?", user.ID).Find(&achievements)

	// 构建响应
	response := &UserProfileResponse{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Role:         string(user.Role),
		AvatarURL:    user.Avatar,
		Bio:          profile.Bio,
		School:       profile.School,
		CreatedAt:    user.CreatedAt,
		WritingLevel: profile.WritingLevel,
		CourierLevel: profile.CourierLevel,
	}

	// Debug logging response
	fmt.Printf("DEBUG Response: bio=%s, school=%s, writing_level=%d\n", response.Bio, response.School, response.WritingLevel)

	// 根据隐私设置和请求者身份决定显示哪些信息
	isOwner := user.ID == requestingUserID

	if isOwner || privacy.ShowEmail {
		response.Email = user.Email
	}

	if isOwner || privacy.ShowOPCode {
		response.OPCode = profile.GetFormattedOPCode(&privacy)
	}

	if isOwner || privacy.ShowStats {
		response.Stats = &stats
	}

	if isOwner {
		response.Privacy = &privacy
	}

	response.Achievements = achievements

	return response, nil
}

// GetUserLetters 获取用户信件列表
func (s *UserProfileService) GetUserLetters(username string, publicOnly bool, requestingUserID string) ([]models.Letter, error) {
	// 查找用户
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 构建查询
	query := s.db.Where("user_id = ? OR author_id = ?", user.ID, user.ID)

	if publicOnly || user.ID != requestingUserID {
		query = query.Where("visibility = ?", models.VisibilityPublic)
	}

	// 获取信件列表
	var letters []models.Letter
	if err := query.Order("created_at DESC").Limit(50).Find(&letters).Error; err != nil {
		return nil, err
	}

	return letters, nil
}

// UpdateUserProfile 更新用户档案
func (s *UserProfileService) UpdateUserProfile(userID string, updates map[string]interface{}) error {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户基本信息
	userUpdates := make(map[string]interface{})
	if nickname, ok := updates["nickname"]; ok {
		userUpdates["nickname"] = nickname
	}
	if avatar, ok := updates["avatar"]; ok {
		userUpdates["avatar"] = avatar
	}

	if len(userUpdates) > 0 {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新扩展档案
	profileUpdates := make(map[string]interface{})
	if bio, ok := updates["bio"]; ok {
		profileUpdates["bio"] = bio
	}
	if school, ok := updates["school"]; ok {
		profileUpdates["school"] = school
	}
	if opCode, ok := updates["op_code"]; ok {
		profileUpdates["op_code"] = opCode
	}

	if len(profileUpdates) > 0 {
		var profile models.UserProfileExtended
		if err := tx.Where("user_id = ?", userID).FirstOrCreate(&profile, models.UserProfileExtended{UserID: userID}).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Model(&profile).Updates(profileUpdates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateUserPrivacy 更新用户隐私设置
func (s *UserProfileService) UpdateUserPrivacy(userID string, privacy models.UserPrivacy) error {
	privacy.UserID = userID
	privacy.UpdatedAt = time.Now()

	return s.db.Save(&privacy).Error
}

// 辅助方法：从角色获取信使等级
func (s *UserProfileService) getCourierLevelFromRole(role models.UserRole) int {
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

// 辅助方法：更新用户统计数据
func (s *UserProfileService) updateUserStats(stats *models.UserStatsData, userID string) {
	// 统计发送的信件数
	var lettersSent int64
	s.db.Model(&models.Letter{}).Where("user_id = ?", userID).Count(&lettersSent)
	stats.LettersSent = int(lettersSent)

	// 统计收到的信件数（需要实现收件人字段）
	var lettersReceived int64
	s.db.Model(&models.Letter{}).Where("author_id = ? AND author_id != user_id", userID).Count(&lettersReceived)
	stats.LettersReceived = int(lettersReceived)

	// 统计博物馆贡献（需要实现博物馆模型）
	// stats.MuseumContributions = ...

	// 更新连续天数
	stats.UpdateStreak()

	// 保存更新
	s.db.Save(stats)
}

// GrantAchievement 授予成就
func (s *UserProfileService) GrantAchievement(userID string, achievementCode string) error {
	// 检查成就是否已存在
	var existing models.UserAchievement
	if err := s.db.Where("user_id = ? AND code = ?", userID, achievementCode).First(&existing).Error; err == nil {
		// 已经拥有该成就
		return nil
	}

	// 查找成就定义
	var achievementDef *models.AchievementDefinition
	for _, def := range models.Achievements {
		if def.Code == achievementCode {
			achievementDef = &def
			break
		}
	}

	if achievementDef == nil {
		return fmt.Errorf("成就不存在: %s", achievementCode)
	}

	// 创建成就记录
	achievement := models.UserAchievement{
		UserID:      userID,
		Code:        achievementDef.Code,
		Name:        achievementDef.Name,
		Description: achievementDef.Description,
		Icon:        achievementDef.Icon,
		Category:    achievementDef.Category,
		UnlockedAt:  time.Now(),
	}

	return s.db.Create(&achievement).Error
}

// CheckAndGrantAchievements 检查并授予成就
func (s *UserProfileService) CheckAndGrantAchievements(userID string) {
	var user models.User
	var stats models.UserStatsData
	var profile models.UserProfileExtended

	s.db.First(&user, "id = ?", userID)
	s.db.FirstOrCreate(&stats, models.UserStatsData{UserID: userID})
	s.db.FirstOrCreate(&profile, models.UserProfileExtended{UserID: userID})

	// 检查写信相关成就
	if stats.LettersSent >= 1 {
		s.GrantAchievement(userID, "first_letter")
	}
	if stats.LettersSent >= 10 {
		s.GrantAchievement(userID, "active_writer")
	}
	if stats.LettersSent >= 50 {
		s.GrantAchievement(userID, "prolific_writer")
	}
	if stats.LettersSent >= 100 {
		s.GrantAchievement(userID, "letter_master")
	}

	// 检查信使相关成就
	switch profile.CourierLevel {
	case 1:
		s.GrantAchievement(userID, "rookie_courier")
	case 2:
		s.GrantAchievement(userID, "area_coordinator")
	case 3:
		s.GrantAchievement(userID, "school_leader")
	case 4:
		s.GrantAchievement(userID, "city_coordinator")
	}

	// 检查连续登录成就
	if stats.CurrentStreak >= 7 {
		s.GrantAchievement(userID, "week_streak")
	}
	if stats.CurrentStreak >= 30 {
		s.GrantAchievement(userID, "month_streak")
	}

	// 检查注册时长成就
	if time.Since(user.CreatedAt) >= 365*24*time.Hour {
		s.GrantAchievement(userID, "year_member")
	}
}
