package services

import (
	"fmt"
	"openpenpal-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// AdminResetPassword 管理员重置用户密码
func (s *UserService) AdminResetPassword(userID string, newPassword string) error {
	// 生成新的密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新用户密码
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetLettersForAdmin 获取管理员信件列表
func (s *LetterService) GetLettersForAdmin(page, limit int, status, userID string) ([]models.Letter, int64, error) {
	var letters []models.Letter
	var total int64

	query := s.db.Model(&models.Letter{}).Preload("User")

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 用户过滤
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&letters).Error; err != nil {
		return nil, 0, err
	}

	return letters, total, nil
}

// AdminUpdateLetterStatus 管理员更新信件状态
func (s *LetterService) AdminUpdateLetterStatus(letterID, status, reason string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if reason != "" {
		updates["admin_note"] = reason
	}

	if err := s.db.Model(&models.Letter{}).Where("id = ?", letterID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update letter status: %w", err)
	}

	return nil
}
