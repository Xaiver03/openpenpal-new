package services

import (
	"fmt"
	"openpenpal-backend/internal/models"
)

// GetLetterByID 获取信件详情
func (s *LetterService) GetLetterByID(letterID string, userID string) (*models.Letter, error) {
	var letter models.Letter
	if err := s.db.Preload("Code").Preload("Envelope").First(&letter, "id = ?", letterID).Error; err != nil {
		return nil, fmt.Errorf("letter not found")
	}

	// 验证权限：只有信件所有者或收件人可以查看
	if letter.UserID != userID && letter.ReplyTo != userID {
		return nil, fmt.Errorf("unauthorized to view this letter")
	}

	return &letter, nil
}

// UpdateLetter 更新信件
func (s *LetterService) UpdateLetter(letterID string, userID string, req *models.UpdateLetterRequest) error {
	var letter models.Letter
	if err := s.db.First(&letter, "id = ? AND user_id = ?", letterID, userID).Error; err != nil {
		return fmt.Errorf("letter not found or unauthorized")
	}

	// 只有草稿状态的信件可以编辑
	if letter.Status != models.StatusDraft {
		return fmt.Errorf("only draft letters can be edited")
	}

	// 更新信件内容
	updates := map[string]interface{}{
		"title":   req.Title,
		"content": req.Content,
		"style":   req.Style,
	}

	if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update letter: %w", err)
	}

	return nil
}

// DeleteLetter 删除信件
func (s *LetterService) DeleteLetter(letterID string, userID string) error {
	var letter models.Letter
	if err := s.db.First(&letter, "id = ? AND user_id = ?", letterID, userID).Error; err != nil {
		return fmt.Errorf("letter not found or unauthorized")
	}

	// 只有草稿状态的信件可以删除
	if letter.Status != models.StatusDraft {
		return fmt.Errorf("only draft letters can be deleted")
	}

	// 软删除
	if err := s.db.Delete(&letter).Error; err != nil {
		return fmt.Errorf("failed to delete letter: %w", err)
	}

	return nil
}

// UpdateEnvelopeBinding 更新信件的信封绑定
func (s *LetterService) UpdateEnvelopeBinding(letterID string, envelopeID string) error {
	var updates map[string]interface{}
	if envelopeID == "" {
		// 解绑
		updates = map[string]interface{}{
			"envelope_id": nil,
		}
	} else {
		// 绑定
		updates = map[string]interface{}{
			"envelope_id": envelopeID,
		}
	}

	if err := s.db.Model(&models.Letter{}).Where("id = ?", letterID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update envelope binding: %w", err)
	}

	return nil
}
