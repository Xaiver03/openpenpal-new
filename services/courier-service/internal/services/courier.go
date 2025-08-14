package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CourierService 信使服务
type CourierService struct {
	db        *gorm.DB
	redis     *redis.Client
	wsManager *utils.WebSocketManager
}

// NewCourierService 创建信使服务实例
func NewCourierService(db *gorm.DB, redis *redis.Client, wsManager *utils.WebSocketManager) *CourierService {
	return &CourierService{
		db:        db,
		redis:     redis,
		wsManager: wsManager,
	}
}

// ApplyCourier 申请成为信使
func (s *CourierService) ApplyCourier(userID string, application *models.CourierApplication) (*models.Courier, error) {
	// 检查是否已经申请过
	var existingCourier models.Courier
	err := s.db.Where("user_id = ?", userID).First(&existingCourier).Error
	if err == nil {
		return nil, gorm.ErrRecordNotFound // 已存在申请
	}

	// 创建新的信使申请
	courier := &models.Courier{
		UserID:     userID,
		Zone:       application.Zone,
		Phone:      application.Phone,
		IDCard:     application.IDCard,
		Experience: application.Experience,
		Status:     models.CourierStatusPending,
		Rating:     5.0,
	}

	if err := s.db.Create(courier).Error; err != nil {
		return nil, err
	}

	// 发送通知给管理员
	s.wsManager.BroadcastToAdmins(utils.WebSocketEvent{
		Type: "NEW_COURIER_APPLICATION",
		Data: map[string]interface{}{
			"courier_id": courier.ID,
			"user_id":    courier.UserID,
			"zone":       courier.Zone,
		},
		Timestamp: time.Now(),
	})

	return courier, nil
}

// GetCourierByUserID 根据用户ID获取信使信息
func (s *CourierService) GetCourierByUserID(userID string) (*models.Courier, error) {
	var courier models.Courier
	err := s.db.Where("user_id = ?", userID).First(&courier).Error
	return &courier, err
}

// GetCourierStats 获取信使统计信息
func (s *CourierService) GetCourierStats(courierID string) (*models.CourierStats, error) {
	var stats models.CourierStats

	// 获取总任务数
	var totalTasks int64
	s.db.Model(&models.Task{}).Where("courier_id = ?", courierID).Count(&totalTasks)
	stats.TotalTasks = int(totalTasks)

	// 获取完成任务数
	var completedTasks int64
	s.db.Model(&models.Task{}).Where("courier_id = ? AND status = ?", courierID, models.TaskStatusDelivered).Count(&completedTasks)
	stats.CompletedTasks = int(completedTasks)

	// 计算成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// 获取信使评分
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err == nil {
		stats.AverageRating = courier.Rating
	}

	// 计算总收入
	var totalReward float64
	s.db.Model(&models.Task{}).Where("courier_id = ? AND status = ?", courierID, models.TaskStatusDelivered).
		Select("COALESCE(SUM(reward), 0)").Scan(&totalReward)
	stats.TotalEarnings = totalReward

	// 本月任务数
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	var thisMonthTasks int64
	s.db.Model(&models.Task{}).Where("courier_id = ? AND created_at >= ?", courierID, startOfMonth).Count(&thisMonthTasks)
	stats.ThisMonthTasks = int(thisMonthTasks)

	return &stats, nil
}

// ApproveCourier 审核通过信使申请
func (s *CourierService) ApproveCourier(courierID string, note string) error {
	now := time.Now()
	err := s.db.Model(&models.Courier{}).Where("id = ?", courierID).Updates(map[string]interface{}{
		"status":      models.CourierStatusApproved,
		"note":        note,
		"approved_at": &now,
	}).Error

	if err != nil {
		return err
	}

	// 获取信使信息并发送通知
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err == nil {
		s.wsManager.BroadcastToUser(courier.UserID, utils.WebSocketEvent{
			Type: "COURIER_APPLICATION_APPROVED",
			Data: map[string]interface{}{
				"courier_id": courier.ID,
				"status":     courier.Status,
				"note":       note,
			},
			Timestamp: time.Now(),
		})
	}

	return nil
}

// RejectCourier 拒绝信使申请
func (s *CourierService) RejectCourier(courierID string, note string) error {
	err := s.db.Model(&models.Courier{}).Where("id = ?", courierID).Updates(map[string]interface{}{
		"status": models.CourierStatusRejected,
		"note":   note,
	}).Error

	if err != nil {
		return err
	}

	// 获取信使信息并发送通知
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err == nil {
		s.wsManager.BroadcastToUser(courier.UserID, utils.WebSocketEvent{
			Type: "COURIER_APPLICATION_REJECTED",
			Data: map[string]interface{}{
				"courier_id": courier.ID,
				"status":     courier.Status,
				"note":       note,
			},
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateCourierRating 更新信使评分
func (s *CourierService) UpdateCourierRating(courierID string, rating float64) error {
	return s.db.Model(&models.Courier{}).Where("user_id = ?", courierID).Update("rating", rating).Error
}
