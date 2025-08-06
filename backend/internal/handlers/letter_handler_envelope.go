package handlers

import (
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
	"shared/pkg/response"
)

// BindEnvelope 为信件绑定信封
// POST /api/v1/letters/:id/bind-envelope
func (h *LetterHandler) BindEnvelope(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	var req struct {
		EnvelopeID string `json:"envelope_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证信件所有权
	letter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		resp.NotFound(c, "Letter not found or unauthorized")
		return
	}

	// 验证信件状态，只有已生成编号的信件才能绑定信封
	if letter.Status != models.StatusGenerated {
		resp.BadRequest(c, "Only letters with generated codes can be bound to envelopes")
		return
	}

	// 验证信封所有权和状态
	envelope, err := h.envelopeService.GetEnvelopeByID(req.EnvelopeID)
	if err != nil {
		resp.NotFound(c, "Envelope not found")
		return
	}

	if envelope.UsedBy != userID {
		resp.Error(c, 403, "Envelope not owned by user")
		return
	}

	if envelope.Status != models.EnvelopeStatusUnsent {
		resp.BadRequest(c, "Envelope is already used")
		return
	}

	// 检查信件是否已经绑定了信封
	if letter.EnvelopeID != nil && *letter.EnvelopeID != "" {
		resp.BadRequest(c, "Letter is already bound to an envelope")
		return
	}

	// 执行绑定
	err = h.envelopeService.BindEnvelopeToLetter(req.EnvelopeID, letterID, userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 更新信件的信封ID
	err = h.letterService.UpdateEnvelopeBinding(letterID, req.EnvelopeID)
	if err != nil {
		resp.InternalServerError(c, "Failed to update letter envelope binding")
		return
	}

	resp.OK(c, "Envelope bound to letter successfully")
}

// UnbindEnvelope 解除信件与信封的绑定
// DELETE /api/v1/letters/:id/bind-envelope
func (h *LetterHandler) UnbindEnvelope(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	// 验证信件所有权
	letter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		resp.NotFound(c, "Letter not found or unauthorized")
		return
	}

	// 检查信件是否绑定了信封
	if letter.EnvelopeID == nil || *letter.EnvelopeID == "" {
		resp.BadRequest(c, "Letter is not bound to any envelope")
		return
	}

	// 验证信封状态，只有未使用的信封才能解绑
	envelope, err := h.envelopeService.GetEnvelopeByID(*letter.EnvelopeID)
	if err != nil {
		resp.NotFound(c, "Bound envelope not found")
		return
	}

	if envelope.Status != models.EnvelopeStatusUnsent {
		resp.BadRequest(c, "Cannot unbind envelope that has been used")
		return
	}

	// 解除绑定 - 重置信封状态
	err = h.letterService.UpdateEnvelopeBinding(letterID, "")
	if err != nil {
		resp.InternalServerError(c, "Failed to unbind envelope from letter")
		return
	}

	resp.OK(c, "Envelope unbound from letter successfully")
}

// GetLetterEnvelope 获取信件绑定的信封信息
// GET /api/v1/letters/:id/envelope
func (h *LetterHandler) GetLetterEnvelope(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	// 验证信件所有权
	letter, err := h.letterService.GetLetterByID(letterID, userID)
	if err != nil {
		resp.NotFound(c, "Letter not found or unauthorized")
		return
	}

	// 检查是否绑定了信封
	if letter.EnvelopeID == nil || *letter.EnvelopeID == "" {
		resp.NotFound(c, "Letter is not bound to any envelope")
		return
	}

	// 获取信封详情
	envelope, err := h.envelopeService.GetEnvelopeByID(*letter.EnvelopeID)
	if err != nil {
		resp.NotFound(c, "Bound envelope not found")
		return
	}

	resp.Success(c, gin.H{
		"letter_id": letterID,
		"envelope":  envelope,
	})
}
