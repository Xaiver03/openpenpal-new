package services

import "gorm.io/gorm"

// CourierTaskService 信使任务服务
type CourierTaskService struct {
	db              *gorm.DB
	notificationSvc *NotificationService
}

// NewCourierTaskService 创建信使任务服务实例
func NewCourierTaskService(db *gorm.DB) *CourierTaskService {
	return &CourierTaskService{db: db}
}

// SetNotificationService 设置通知服务（避免循环依赖）
func (s *CourierTaskService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// CreateDeliveryTask 创建配送任务（用于回信等）
func (s *CourierTaskService) CreateDeliveryTask(replyID, deliveryCode string) error {
	// 创建简单的配送任务记录
	// 这里是简化实现，实际可能需要更复杂的任务分配逻辑

	// 可以通过通知服务发送任务创建通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyUser("system", "delivery_task_created", map[string]interface{}{
			"reply_id":      replyID,
			"delivery_code": deliveryCode,
		})
	}

	return nil
}
