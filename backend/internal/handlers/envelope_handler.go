package handlers

import (
	"fmt"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shared/pkg/response"
)

type EnvelopeHandler struct {
	envelopeService *services.EnvelopeService
}

func NewEnvelopeHandler(envelopeService *services.EnvelopeService) *EnvelopeHandler {
	return &EnvelopeHandler{
		envelopeService: envelopeService,
	}
}

// GetMyEnvelopes 获取我的信封列表
// GET /api/v1/envelopes/my
func (h *EnvelopeHandler) GetMyEnvelopes(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	status := c.Query("status") // 可选过滤条件: unsent, bound, used

	envelopes, err := h.envelopeService.GetUserEnvelopes(userID)
	_ = status // TODO: 使用status过滤信封
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"envelopes": envelopes,
		"total":     len(envelopes),
	})
}

// GetEnvelopeDesigns 获取可用的信封设计列表
// GET /api/v1/envelopes/designs
func (h *EnvelopeHandler) GetEnvelopeDesigns(c *gin.Context) {
	resp := response.NewGinResponse()

	// 获取当前用户ID - FSD增强：用于OP Code区域过滤
	userID, _ := c.Get("user_id")

	// 获取查询参数
	filters := map[string]interface{}{
		"school_code": c.Query("school_code"),
		"type":        c.Query("type"),
		"status":      "approved", // 只显示已审核的设计
	}

	// 添加用户ID用于OP Code过滤
	if userID != nil {
		filters["user_id"] = userID.(string)
	}

	designs, err := h.envelopeService.GetDesigns(filters)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"designs": designs,
		"total":   len(designs),
	})
}

// CreateEnvelopeOrder 创建信封订单
// POST /api/v1/envelopes/orders
func (h *EnvelopeHandler) CreateEnvelopeOrder(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		DesignID string `json:"design_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	order, err := h.envelopeService.CreateEnvelopeOrder(userID, req.DesignID, req.Quantity)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "Envelope order created successfully", order)
}

// GetEnvelopeOrders 获取我的信封订单列表
// GET /api/v1/envelopes/orders
func (h *EnvelopeHandler) GetEnvelopeOrders(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	orders, err := h.envelopeService.GetUserEnvelopeOrders(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"orders": orders,
		"total":  len(orders),
	})
}

// ProcessEnvelopePayment 处理信封订单支付
// POST /api/v1/envelopes/orders/:id/pay
func (h *EnvelopeHandler) ProcessEnvelopePayment(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		resp.BadRequest(c, "Order ID is required")
		return
	}

	var req struct {
		PaymentMethod string `json:"payment_method" binding:"required"`
		PaymentID     string `json:"payment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证订单所有权
	order, err := h.envelopeService.GetEnvelopeOrder(orderID, userID)
	if err != nil {
		resp.NotFound(c, err.Error())
		return
	}

	if order.Status != "pending" {
		resp.BadRequest(c, "Order cannot be paid")
		return
	}

	// 更新支付状态
	paymentID := req.PaymentID
	if paymentID == "" {
		paymentID = fmt.Sprintf("PAY-%s", uuid.New().String()[0:8])
	}

	err = h.envelopeService.UpdateOrderPaymentStatus(orderID, paymentID, req.PaymentMethod)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 生成信封实例
	err = h.envelopeService.GenerateEnvelopesForOrder(orderID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Payment processed and envelopes generated successfully")
}
