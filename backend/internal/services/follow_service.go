package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

type FollowService struct {
	db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{db: db}
}

// FollowUser 关注用户
func (s *FollowService) FollowUser(followerID, followingID string, notificationEnabled bool) (*models.FollowActionResponse, error) {
	if followerID == followingID {
		return &models.FollowActionResponse{
			Success: false,
			Message: "Cannot follow yourself",
		}, fmt.Errorf("cannot follow yourself")
	}

	// 检查目标用户是否存在
	var targetUser models.User
	if err := s.db.First(&targetUser, "id = ?", followingID).Error; err != nil {
		return &models.FollowActionResponse{
			Success: false,
			Message: "User not found",
		}, err
	}

	// 检查是否已经关注
	var existingRelation models.UserRelationship
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existingRelation).Error

	if err == nil {
		// 已存在关注关系
		if existingRelation.Status == models.FollowStatusActive {
			return &models.FollowActionResponse{
				Success:     false,
				IsFollowing: true,
				Message:     "Already following this user",
			}, nil
		}

		// 重新激活关注关系
		existingRelation.Status = models.FollowStatusActive
		existingRelation.NotificationEnabled = notificationEnabled
		existingRelation.UpdatedAt = time.Now()

		if err := s.db.Save(&existingRelation).Error; err != nil {
			return nil, err
		}
	} else if err == gorm.ErrRecordNotFound {
		// 创建新的关注关系
		newRelation := models.UserRelationship{
			FollowerID:          followerID,
			FollowingID:         followingID,
			Status:              models.FollowStatusActive,
			NotificationEnabled: notificationEnabled,
		}

		if err := s.db.Create(&newRelation).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 更新统计数据
	if err := s.updateFollowStats(followerID, followingID); err != nil {
		// 统计更新失败不影响关注操作
		fmt.Printf("Warning: Failed to update follow stats: %v\n", err)
	}

	// 创建活动记录
	activity := models.FollowActivity{
		ActorID:  followerID,
		TargetID: followingID,
		Type:     "new_follower",
	}
	s.db.Create(&activity)

	// 获取最新的关注统计
	followerStats, _ := s.getUserFollowStats(followingID)
	followerCount := 0
	if followerStats != nil {
		followerCount = followerStats.FollowersCount
	}

	return &models.FollowActionResponse{
		Success:       true,
		IsFollowing:   true,
		FollowerCount: followerCount,
		FollowedAt:    time.Now().Format(time.RFC3339),
		Message:       "Successfully followed user",
	}, nil
}

// UnfollowUser 取消关注用户
func (s *FollowService) UnfollowUser(followerID, followingID string) (*models.FollowActionResponse, error) {
	// 查找关注关系
	var relation models.UserRelationship
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&relation).Error

	if err == gorm.ErrRecordNotFound {
		return &models.FollowActionResponse{
			Success:     false,
			IsFollowing: false,
			Message:     "Not following this user",
		}, nil
	}

	if err != nil {
		return nil, err
	}

	// 删除关注关系
	if err := s.db.Delete(&relation).Error; err != nil {
		return nil, err
	}

	// 更新统计数据
	if err := s.updateFollowStats(followerID, followingID); err != nil {
		fmt.Printf("Warning: Failed to update follow stats: %v\n", err)
	}

	// 获取最新的关注统计
	followerStats, _ := s.getUserFollowStats(followingID)
	followerCount := 0
	if followerStats != nil {
		followerCount = followerStats.FollowersCount
	}

	return &models.FollowActionResponse{
		Success:       true,
		IsFollowing:   false,
		FollowerCount: followerCount,
		Message:       "Successfully unfollowed user",
	}, nil
}

// GetFollowers 获取用户的粉丝列表
func (s *FollowService) GetFollowers(userID string, req *models.FollowListRequest) (*models.FollowListResponse, error) {
	offset := (req.Page - 1) * req.Limit

	query := s.db.Table("user_relationships").
		Select("users.*").
		Joins("JOIN users ON users.id = user_relationships.follower_id").
		Where("user_relationships.following_id = ? AND user_relationships.status = ?", userID, models.FollowStatusActive)

	// 搜索过滤
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(users.username) LIKE ? OR LOWER(users.nickname) LIKE ?", searchTerm, searchTerm)
	}

	// 学校过滤
	if req.SchoolFilter != "" {
		query = query.Where("users.school_code = ?", req.SchoolFilter)
	}

	// 排序
	orderField := "user_relationships.created_at"
	switch req.SortBy {
	case "nickname":
		orderField = "users.nickname"
	case "username":
		orderField = "users.username"
	case "created_at":
		orderField = "user_relationships.created_at"
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	// 获取总数
	var total int64
	countQuery := *query
	countQuery.Count(&total)

	// 获取数据
	var users []models.User
	if err := query.Order(fmt.Sprintf("%s %s", orderField, order)).
		Offset(offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为FollowUser格式
	followUsers := make([]models.FollowUser, 0, len(users))
	for _, user := range users {
		stats, _ := s.getUserFollowStats(user.ID)
		profile, _ := s.getUserProfile(user.ID)

		followUser := user.ToFollowUser(stats, profile)
		followUsers = append(followUsers, *followUser)
	}

	response := &models.FollowListResponse{
		Users: followUsers,
	}

	response.Pagination.Page = req.Page
	response.Pagination.Limit = req.Limit
	response.Pagination.Total = int(total)
	response.Pagination.Pages = int(math.Ceil(float64(total) / float64(req.Limit)))

	return response, nil
}

// GetFollowing 获取用户的关注列表
func (s *FollowService) GetFollowing(userID string, req *models.FollowListRequest) (*models.FollowListResponse, error) {
	offset := (req.Page - 1) * req.Limit

	query := s.db.Table("user_relationships").
		Select("users.*").
		Joins("JOIN users ON users.id = user_relationships.following_id").
		Where("user_relationships.follower_id = ? AND user_relationships.status = ?", userID, models.FollowStatusActive)

	// 搜索过滤
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(users.username) LIKE ? OR LOWER(users.nickname) LIKE ?", searchTerm, searchTerm)
	}

	// 学校过滤
	if req.SchoolFilter != "" {
		query = query.Where("users.school_code = ?", req.SchoolFilter)
	}

	// 排序
	orderField := "user_relationships.created_at"
	switch req.SortBy {
	case "nickname":
		orderField = "users.nickname"
	case "username":
		orderField = "users.username"
	case "created_at":
		orderField = "user_relationships.created_at"
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	// 获取总数
	var total int64
	countQuery := *query
	countQuery.Count(&total)

	// 获取数据
	var users []models.User
	if err := query.Order(fmt.Sprintf("%s %s", orderField, order)).
		Offset(offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为FollowUser格式
	followUsers := make([]models.FollowUser, 0, len(users))
	for _, user := range users {
		stats, _ := s.getUserFollowStats(user.ID)
		profile, _ := s.getUserProfile(user.ID)

		followUser := user.ToFollowUser(stats, profile)
		// 标记为正在关注
		followUser.IsFollowing = true
		followUsers = append(followUsers, *followUser)
	}

	response := &models.FollowListResponse{
		Users: followUsers,
	}

	response.Pagination.Page = req.Page
	response.Pagination.Limit = req.Limit
	response.Pagination.Total = int(total)
	response.Pagination.Pages = int(math.Ceil(float64(total) / float64(req.Limit)))

	return response, nil
}

// SearchUsers 搜索用户
func (s *FollowService) SearchUsers(req *models.UserSearchRequest, currentUserID string) (*models.UserSearchResponse, error) {
	query := s.db.Model(&models.User{}).Where("is_active = ?", true)

	// 排除当前用户
	if currentUserID != "" {
		query = query.Where("id != ?", currentUserID)
	}

	// 搜索条件
	if req.Query != "" {
		searchTerm := "%" + strings.ToLower(req.Query) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?", searchTerm, searchTerm)
	}

	// 学校过滤
	if req.SchoolCode != "" {
		query = query.Where("school_code = ?", req.SchoolCode)
	}

	// 角色过滤
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// 排序
	orderField := "created_at"
	switch req.SortBy {
	case "followers":
		// 需要join统计表排序
		orderField = "created_at" // 暂时使用创建时间排序
	case "activity":
		orderField = "last_login_at"
	case "joined":
		orderField = "created_at"
	case "relevance":
		orderField = "username" // 按字母顺序
	}

	order := "DESC"
	if req.Order == "asc" {
		order = "ASC"
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取数据
	var users []models.User
	if err := query.Order(fmt.Sprintf("%s %s", orderField, order)).
		Offset(req.Offset).Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为FollowUser格式并检查关注状态
	followUsers := make([]models.FollowUser, 0, len(users))
	for _, user := range users {
		stats, _ := s.getUserFollowStats(user.ID)
		profile, _ := s.getUserProfile(user.ID)

		followUser := user.ToFollowUser(stats, profile)

		// 检查当前用户是否关注了这个用户
		if currentUserID != "" {
			isFollowing, _ := s.IsFollowing(currentUserID, user.ID)
			followUser.IsFollowing = isFollowing
		}

		followUsers = append(followUsers, *followUser)
	}

	return &models.UserSearchResponse{
		Users:          followUsers,
		Total:          int(total),
		Query:          req.Query,
		FiltersApplied: req,
	}, nil
}

// GetUserSuggestions 获取用户推荐
func (s *FollowService) GetUserSuggestions(userID string, req *models.UserSuggestionsRequest) (*models.UserSuggestionsResponse, error) {
	// 简单的推荐逻辑：推荐同学校的活跃用户
	var currentUser models.User
	if err := s.db.First(&currentUser, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	query := s.db.Model(&models.User{}).
		Where("is_active = ? AND id != ?", true, userID)

	// 优先推荐同学校的用户
	if currentUser.SchoolCode != "" {
		query = query.Where("school_code = ?", currentUser.SchoolCode)
	}

	// 排除已关注的用户
	if req.ExcludeFollowed {
		followingIDs, _ := s.getFollowingIDs(userID)
		if len(followingIDs) > 0 {
			query = query.Where("id NOT IN ?", followingIDs)
		}
	}

	// 按最近活跃排序
	var users []models.User
	if err := query.Order("last_login_at DESC").
		Limit(req.Limit).Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为推荐格式
	suggestions := make([]models.FollowSuggestionItem, 0, len(users))
	for _, user := range users {
		stats, _ := s.getUserFollowStats(user.ID)
		profile, _ := s.getUserProfile(user.ID)

		followUser := user.ToFollowUser(stats, profile)

		// 确定推荐理由
		reason := "活跃用户"
		if user.SchoolCode == currentUser.SchoolCode {
			reason = "同校推荐"
		}

		suggestion := models.FollowSuggestionItem{
			User:            *followUser,
			Reason:          reason,
			ConfidenceScore: 0.8, // 固定置信度
		}

		suggestions = append(suggestions, suggestion)
	}

	return &models.UserSuggestionsResponse{
		Suggestions:        suggestions,
		AlgorithmUsed:      "school_based",
		RefreshAvailableAt: time.Now().Add(time.Hour).Format(time.RFC3339),
	}, nil
}

// IsFollowing 检查是否关注用户
func (s *FollowService) IsFollowing(followerID, followingID string) (bool, error) {
	var relation models.UserRelationship
	err := s.db.Where("follower_id = ? AND following_id = ? AND status = ?",
		followerID, followingID, models.FollowStatusActive).First(&relation).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// GetFollowStatus 获取关注状态
func (s *FollowService) GetFollowStatus(currentUserID, targetUserID string) (map[string]interface{}, error) {
	isFollowing, err := s.IsFollowing(currentUserID, targetUserID)
	if err != nil {
		return nil, err
	}

	isFollower, err := s.IsFollowing(targetUserID, currentUserID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"is_following": isFollowing,
		"is_follower":  isFollower,
		"is_mutual":    isFollowing && isFollower,
	}, nil
}

// Helper functions

// updateFollowStats 更新关注统计
func (s *FollowService) updateFollowStats(followerID, followingID string) error {
	// 更新关注者的关注数
	var followingCount int64
	s.db.Model(&models.UserRelationship{}).
		Where("follower_id = ? AND status = ?", followerID, models.FollowStatusActive).
		Count(&followingCount)

	// 更新被关注者的粉丝数
	var followersCount int64
	s.db.Model(&models.UserRelationship{}).
		Where("following_id = ? AND status = ?", followingID, models.FollowStatusActive).
		Count(&followersCount)

	// 更新或创建统计记录
	followingStats := models.FollowStats{UserID: followerID}
	s.db.Model(&followingStats).Where("user_id = ?", followerID).
		Updates(map[string]interface{}{
			"following_count": followingCount,
			"updated_at":      time.Now(),
		})

	followersStats := models.FollowStats{UserID: followingID}
	s.db.Model(&followersStats).Where("user_id = ?", followingID).
		Updates(map[string]interface{}{
			"followers_count": followersCount,
			"updated_at":      time.Now(),
		})

	return nil
}

// getUserFollowStats 获取用户关注统计
func (s *FollowService) getUserFollowStats(userID string) (*models.FollowStats, error) {
	var stats models.FollowStats
	err := s.db.Where("user_id = ?", userID).First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 如果统计记录不存在，创建一个
		stats = models.FollowStats{UserID: userID}
		s.db.Create(&stats)
		return &stats, nil
	}

	return &stats, err
}

// getUserProfile 获取用户档案
func (s *FollowService) getUserProfile(userID string) (*models.UserProfileExtended, error) {
	var profile models.UserProfileExtended
	err := s.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 档案不存在不是错误
	}
	return &profile, err
}

// getFollowingIDs 获取用户关注的ID列表
func (s *FollowService) getFollowingIDs(userID string) ([]string, error) {
	var relationships []models.UserRelationship
	if err := s.db.Where("follower_id = ? AND status = ?", userID, models.FollowStatusActive).
		Find(&relationships).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(relationships))
	for i, rel := range relationships {
		ids[i] = rel.FollowingID
	}
	return ids, nil
}

// GetUserStats 获取用户关注统计
func (s *FollowService) GetUserStats(userID string) (*models.UserStats, error) {
	var stats models.UserFollowStats
	
	// 查找用户统计信息
	if err := s.db.Where("user_id = ?", userID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有统计记录，创建一个新的
			stats = models.UserFollowStats{
				UserID:         userID,
				FollowingCount: 0,
				FollowersCount: 0,
			}
			if err := s.db.Create(&stats).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	
	// 返回统计数据
	return &models.UserStats{
		UserID:         userID,
		FollowingCount: int64(stats.FollowingCount),
		FollowersCount: int64(stats.FollowersCount),
		MutualCount:    int64(stats.MutualCount),
		LastActive:     stats.LastActive,
	}, nil
}
