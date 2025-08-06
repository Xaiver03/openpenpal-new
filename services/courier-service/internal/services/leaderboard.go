package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// LeaderboardService 排行榜服务
type LeaderboardService struct {
	db        *gorm.DB
	wsManager *utils.WebSocketManager
}

// NewLeaderboardService 创建排行榜服务
func NewLeaderboardService(db *gorm.DB, wsManager *utils.WebSocketManager) *LeaderboardService {
	return &LeaderboardService{
		db:        db,
		wsManager: wsManager,
	}
}

// GetSchoolLeaderboard 获取学校排行榜
func (s *LeaderboardService) GetSchoolLeaderboard(req *models.CourierLeaderboardRequest) (*models.CourierLeaderboardResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.CourierRanking{}).
		Preload("Courier").
		Where("school_rank > 0")

	if req.ZoneCode != "" {
		query = query.Joins("JOIN couriers ON courier_rankings.courier_id = couriers.id").
			Where("couriers.zone_code LIKE ?", req.ZoneCode+"%")
	}

	var rankings []models.CourierRanking
	var total int64

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	err := query.Order("school_rank ASC").
		Limit(limit).
		Offset(offset).
		Find(&rankings).Error

	if err != nil {
		return nil, fmt.Errorf("获取学校排行榜失败: %w", err)
	}

	return &models.CourierLeaderboardResponse{
		Rankings: rankings,
		Total:    int(total),
		Page:     offset/limit + 1,
		Limit:    limit,
	}, nil
}

// GetZoneLeaderboard 获取片区排行榜
func (s *LeaderboardService) GetZoneLeaderboard(req *models.CourierLeaderboardRequest) (*models.CourierLeaderboardResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.CourierRanking{}).
		Preload("Courier").
		Where("zone_rank > 0")

	if req.ZoneCode != "" {
		query = query.Joins("JOIN couriers ON courier_rankings.courier_id = couriers.id").
			Where("couriers.zone_code = ?", req.ZoneCode)
	}

	var rankings []models.CourierRanking
	var total int64

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	err := query.Order("zone_rank ASC").
		Limit(limit).
		Offset(offset).
		Find(&rankings).Error

	if err != nil {
		return nil, fmt.Errorf("获取片区排行榜失败: %w", err)
	}

	return &models.CourierLeaderboardResponse{
		Rankings: rankings,
		Total:    int(total),
		Page:     offset/limit + 1,
		Limit:    limit,
	}, nil
}

// GetNationalLeaderboard 获取全国排行榜
func (s *LeaderboardService) GetNationalLeaderboard(req *models.CourierLeaderboardRequest) (*models.CourierLeaderboardResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var rankings []models.CourierRanking
	var total int64

	query := s.db.Model(&models.CourierRanking{}).
		Preload("Courier").
		Where("national_rank > 0")

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	err := query.Order("national_rank ASC").
		Limit(limit).
		Offset(offset).
		Find(&rankings).Error

	if err != nil {
		return nil, fmt.Errorf("获取全国排行榜失败: %w", err)
	}

	return &models.CourierLeaderboardResponse{
		Rankings: rankings,
		Total:    int(total),
		Page:     offset/limit + 1,
		Limit:    limit,
	}, nil
}

// GetPointsHistory 获取积分历史
func (s *LeaderboardService) GetPointsHistory(courierID string, limit int, offset int) ([]models.CourierPointsHistory, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	var history []models.CourierPointsHistory
	var total int64

	query := s.db.Where("courier_id = ?", courierID)

	// 获取总数
	query.Model(&models.CourierPointsHistory{}).Count(&total)

	// 获取分页数据
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, 0, fmt.Errorf("获取积分历史失败: %w", err)
	}

	return history, total, nil
}

// AddPoints 增加积分
func (s *LeaderboardService) AddPoints(courierID string, points int, pointsType string, description string, taskID *string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新信使积分
	if err := tx.Model(&models.Courier{}).
		Where("id = ?", courierID).
		Update("points", gorm.Expr("points + ?", points)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新积分失败: %w", err)
	}

	// 记录积分历史
	history := &models.CourierPointsHistory{
		CourierID:   courierID,
		TaskID:      taskID,
		Points:      points,
		Type:        pointsType,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录积分历史失败: %w", err)
	}

	tx.Commit()

	// 异步更新排行榜
	go s.updateRanking(courierID)

	// 发送积分变更通知
	s.notifyPointsChanged(courierID, points, pointsType, description)

	return nil
}

// UpdateRankings 更新所有排行榜
func (s *LeaderboardService) UpdateRankings() error {
	// 更新学校排行榜
	if err := s.updateSchoolRankings(); err != nil {
		return fmt.Errorf("更新学校排行榜失败: %w", err)
	}

	// 更新片区排行榜
	if err := s.updateZoneRankings(); err != nil {
		return fmt.Errorf("更新片区排行榜失败: %w", err)
	}

	// 更新全国排行榜
	if err := s.updateNationalRankings(); err != nil {
		return fmt.Errorf("更新全国排行榜失败: %w", err)
	}

	return nil
}

// GetCourierRank 获取信使排名信息
func (s *LeaderboardService) GetCourierRank(courierID string) (*models.CourierRanking, error) {
	var ranking models.CourierRanking
	err := s.db.Where("courier_id = ?", courierID).
		Preload("Courier").
		First(&ranking).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有排名记录，创建一个初始记录
			return s.createInitialRanking(courierID)
		}
		return nil, fmt.Errorf("获取排名信息失败: %w", err)
	}

	return &ranking, nil
}

// 私有方法

// createInitialRanking 创建初始排名记录
func (s *LeaderboardService) createInitialRanking(courierID string) (*models.CourierRanking, error) {
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err != nil {
		return nil, fmt.Errorf("信使不存在: %w", err)
	}

	ranking := &models.CourierRanking{
		CourierID:    courierID,
		Courier:      courier,
		SchoolRank:   0,
		ZoneRank:     0,
		NationalRank: 0,
		Points:       courier.Points,
		TotalTasks:   0,
		SuccessRate:  0,
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(ranking).Error; err != nil {
		return nil, fmt.Errorf("创建排名记录失败: %w", err)
	}

	return ranking, nil
}

// updateRanking 更新单个信使排名
func (s *LeaderboardService) updateRanking(courierID string) {
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err != nil {
		return
	}

	// 获取任务统计
	var totalTasks int64
	var completedTasks int64

	s.db.Model(&models.Task{}).Where("courier_id = ?", courierID).Count(&totalTasks)
	s.db.Model(&models.Task{}).Where("courier_id = ? AND status = 'completed'", courierID).Count(&completedTasks)

	successRate := 0.0
	if totalTasks > 0 {
		successRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	// 更新或创建排名记录
	ranking := &models.CourierRanking{
		CourierID:   courierID,
		Points:      courier.Points,
		TotalTasks:  int(totalTasks),
		SuccessRate: successRate,
		UpdatedAt:   time.Now(),
	}

	s.db.Where("courier_id = ?", courierID).
		Assign(ranking).
		FirstOrCreate(&ranking)
}

// updateSchoolRankings 更新学校排行榜
func (s *LeaderboardService) updateSchoolRankings() error {
	// 按学校分组更新排名
	query := `
		UPDATE courier_rankings 
		SET school_rank = ranked.rank
		FROM (
			SELECT 
				cr.id,
				ROW_NUMBER() OVER (
					PARTITION BY SUBSTRING(c.zone_code, 1, 6) 
					ORDER BY cr.points DESC, cr.success_rate DESC
				) as rank
			FROM courier_rankings cr
			JOIN couriers c ON cr.courier_id = c.id
			WHERE c.status = 'approved'
		) ranked
		WHERE courier_rankings.id = ranked.id`

	return s.db.Exec(query).Error
}

// updateZoneRankings 更新片区排行榜
func (s *LeaderboardService) updateZoneRankings() error {
	query := `
		UPDATE courier_rankings 
		SET zone_rank = ranked.rank
		FROM (
			SELECT 
				cr.id,
				ROW_NUMBER() OVER (
					PARTITION BY c.zone_code 
					ORDER BY cr.points DESC, cr.success_rate DESC
				) as rank
			FROM courier_rankings cr
			JOIN couriers c ON cr.courier_id = c.id
			WHERE c.status = 'approved'
		) ranked
		WHERE courier_rankings.id = ranked.id`

	return s.db.Exec(query).Error
}

// updateNationalRankings 更新全国排行榜
func (s *LeaderboardService) updateNationalRankings() error {
	query := `
		UPDATE courier_rankings 
		SET national_rank = ranked.rank
		FROM (
			SELECT 
				cr.id,
				ROW_NUMBER() OVER (
					ORDER BY cr.points DESC, cr.success_rate DESC
				) as rank
			FROM courier_rankings cr
			JOIN couriers c ON cr.courier_id = c.id
			WHERE c.status = 'approved'
		) ranked
		WHERE courier_rankings.id = ranked.id`

	return s.db.Exec(query).Error
}

// 通知方法

// notifyPointsChanged 通知积分变更
func (s *LeaderboardService) notifyPointsChanged(courierID string, points int, pointsType string, description string) {
	var courier models.Courier
	if s.db.First(&courier, courierID).Error != nil {
		return
	}

	event := utils.WebSocketEvent{
		Type: "POINTS_CHANGED",
		Data: map[string]interface{}{
			"courier_id":   courierID,
			"points":       points,
			"total_points": courier.Points,
			"type":         pointsType,
			"description":  description,
		},
		Timestamp: time.Now(),
	}

	s.wsManager.BroadcastToUser(courier.UserID, event)
}

// GetDB 获取数据库连接
func (s *LeaderboardService) GetDB() *gorm.DB {
	return s.db
}