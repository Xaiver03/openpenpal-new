package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

type AppointmentService struct {
	db *gorm.DB
}

func NewAppointmentService(db *gorm.DB) *AppointmentService {
	return &AppointmentService{db: db}
}

// AppointmentRecord 任命记录
type AppointmentRecord struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	AppointerID string          `json:"appointer_id" gorm:"type:varchar(36);not null"`
	AppointeeID string          `json:"appointee_id" gorm:"type:varchar(36);not null"`
	FromRole    models.UserRole `json:"from_role" gorm:"type:varchar(20);not null"`
	ToRole      models.UserRole `json:"to_role" gorm:"type:varchar(20);not null"`
	Status      string          `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Reason      string          `json:"reason" gorm:"type:text"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CanAppoint 检查是否可以任命指定角色
func (s *AppointmentService) CanAppoint(appointer *models.User, targetRole models.UserRole) bool {
	appointerLevel := models.RoleHierarchy[appointer.Role]
	targetLevel := models.RoleHierarchy[targetRole]

	// 只能任命比自己低一级的角色
	return appointerLevel == targetLevel+1
}

// AppointUser 任命用户到指定角色
func (s *AppointmentService) AppointUser(appointerID, appointeeID string, newRole models.UserRole, reason string) error {
	var appointer, appointee models.User

	// 验证任命者
	if err := s.db.First(&appointer, "id = ?", appointerID).Error; err != nil {
		return fmt.Errorf("appointer not found: %w", err)
	}

	// 验证被任命者
	if err := s.db.First(&appointee, "id = ?", appointeeID).Error; err != nil {
		return fmt.Errorf("appointee not found: %w", err)
	}

	// 检查任命权限
	if !s.CanAppoint(&appointer, newRole) {
		return errors.New("insufficient appointment authority")
	}

	// 检查被任命者当前角色
	currentLevel := models.RoleHierarchy[appointee.Role]
	targetLevel := models.RoleHierarchy[newRole]

	if currentLevel >= targetLevel {
		return errors.New("cannot appoint to same or lower level")
	}

	// 创建任命记录
	record := AppointmentRecord{
		AppointerID: appointerID,
		AppointeeID: appointeeID,
		FromRole:    appointee.Role,
		ToRole:      newRole,
		Status:      "approved", // 简化流程，实际可添加审批
		Reason:      reason,
	}

	// 事务处理
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create appointment record: %w", err)
	}

	if err := tx.Model(&appointee).Update("role", newRole).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return tx.Commit().Error
}

// GetAppointmentHistory 获取任命历史
func (s *AppointmentService) GetAppointmentHistory(userID string) ([]AppointmentRecord, error) {
	var records []AppointmentRecord
	err := s.db.Where("appointee_id = ? OR appointer_id = ?", userID, userID).
		Order("created_at desc").
		Find(&records).Error
	return records, err
}

// GetAppointmentsByRole 获取特定角色的任命记录
func (s *AppointmentService) GetAppointmentsByRole(role models.UserRole) ([]AppointmentRecord, error) {
	var records []AppointmentRecord
	err := s.db.Where("to_role = ?", role).
		Order("created_at desc").
		Find(&records).Error
	return records, err
}
