package handlers

import (
	"encoding/json"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shared/pkg/response"
)

// ShopHandler 商店处理器
type ShopHandler struct {
	shopService *services.ShopService
	userService *services.UserService
}

// NewShopHandler 创建商店处理器实例
func NewShopHandler(shopService *services.ShopService, userService *services.UserService) *ShopHandler {
	return &ShopHandler{
		shopService: shopService,
		userService: userService,
	}
}

// Product Handlers

// GetProducts 获取商品列表
func (h *ShopHandler) GetProducts(c *gin.Context) {
	resp := response.NewGinResponse()
	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		params["limit"] = limit
	}
	if category := c.Query("category"); category != "" {
		params["category"] = category
	}
	if productType := c.Query("product_type"); productType != "" {
		params["product_type"] = models.ProductType(productType)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		params["keyword"] = keyword
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		params["sort_by"] = sortBy
	}
	if order := c.Query("order"); order != "" {
		params["order"] = order
	}
	if inStockOnly := c.Query("in_stock_only"); inStockOnly == "true" {
		params["in_stock_only"] = true
	}
	if featuredOnly := c.Query("featured_only"); featuredOnly == "true" {
		params["featured_only"] = true
	}
	if minPrice, err := strconv.ParseFloat(c.Query("min_price"), 64); err == nil {
		params["min_price"] = minPrice
	}
	if maxPrice, err := strconv.ParseFloat(c.Query("max_price"), 64); err == nil {
		params["max_price"] = maxPrice
	}

	products, total, err := h.shopService.GetProducts(params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := params["page"].(int)
	limit := params["limit"].(int)
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     products,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// GetProduct 获取商品详情
func (h *ShopHandler) GetProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	product, err := h.shopService.GetProductByID(productID)
	if err != nil {
		resp.NotFound(c, "商品不存在")
		return
	}

	resp.Success(c, product)
}

// CreateProduct 创建商品（管理员）
func (h *ShopHandler) CreateProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建商品")
		return
	}

	if err := h.shopService.CreateProduct(&product); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "商品创建成功", product)
}

// UpdateProduct 更新商品（管理员）
func (h *ShopHandler) UpdateProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新商品")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.shopService.UpdateProduct(productID, updates); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品更新成功", nil)
}

// DeleteProduct 删除商品（管理员）
func (h *ShopHandler) DeleteProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以删除商品")
		return
	}

	if err := h.shopService.DeleteProduct(productID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品删除成功", nil)
}

// Cart Handlers

// GetCart 获取购物车
func (h *ShopHandler) GetCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	cart, err := h.shopService.GetOrCreateCart(user.ID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, cart)
}

// AddToCart 添加商品到购物车
func (h *ShopHandler) AddToCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	cartItem, err := h.shopService.AddToCart(user.ID, productID, req.Quantity)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, cartItem)
}

// UpdateCartItem 更新购物车项目
func (h *ShopHandler) UpdateCartItem(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.shopService.UpdateCartItem(user.ID, itemID, req.Quantity); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "购物车更新成功", nil)
}

// RemoveFromCart 从购物车移除商品
func (h *ShopHandler) RemoveFromCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.shopService.RemoveFromCart(user.ID, itemID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品已移除", nil)
}

// ClearCart 清空购物车
func (h *ShopHandler) ClearCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	if err := h.shopService.ClearCart(user.ID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "购物车已清空", nil)
}

// Order Handlers

// CreateOrder 创建订单
func (h *ShopHandler) CreateOrder(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		PaymentMethod   string                 `json:"payment_method" binding:"required"`
		ShippingAddress map[string]interface{} `json:"shipping_address" binding:"required"`
		Notes           string                 `json:"notes"`
		CouponCode      string                 `json:"coupon_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	orderData := map[string]interface{}{
		"payment_method":   req.PaymentMethod,
		"shipping_address": req.ShippingAddress,
		"notes":            req.Notes,
		"coupon_code":      req.CouponCode,
	}

	order, err := h.shopService.CreateOrder(user.ID, orderData)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, order)
}

// GetOrders 获取订单列表
func (h *ShopHandler) GetOrders(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		params["limit"] = limit
	}
	if status := c.Query("status"); status != "" {
		params["status"] = models.OrderStatus(status)
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		params["payment_status"] = models.PaymentStatus(paymentStatus)
	}

	orders, total, err := h.shopService.GetOrders(user.ID, params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := params["page"].(int)
	limit := params["limit"].(int)

	resp.Success(c, gin.H{
		"items":     orders,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// GetOrder 获取订单详情
func (h *ShopHandler) GetOrder(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	order, err := h.shopService.GetOrderByID(user.ID, orderID)
	if err != nil {
		resp.NotFound(c, "订单不存在")
		return
	}

	resp.Success(c, order)
}

// UpdateOrderStatus 更新订单状态（管理员）
func (h *ShopHandler) UpdateOrderStatus(c *gin.Context) {
	resp := response.NewGinResponse()
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证状态是否有效
	status := models.OrderStatus(req.Status)
	validStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPaid,
		models.OrderStatusProcessing,
		models.OrderStatusShipped,
		models.OrderStatusDelivered,
		models.OrderStatusCompleted,
		models.OrderStatusCancelled,
		models.OrderStatusRefunded,
	}

	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		resp.BadRequest(c, "无效的订单状态")
		return
	}

	if err := h.shopService.UpdateOrderStatus(orderID, status); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "订单状态更新成功", nil)
}

// PayOrder 支付订单
func (h *ShopHandler) PayOrder(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	var req struct {
		PaymentID string `json:"payment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证订单属于当前用户
	order, err := h.shopService.GetOrderByID(user.ID, orderID)
	if err != nil {
		resp.NotFound(c, "订单不存在")
		return
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		resp.BadRequest(c, "订单已支付")
		return
	}

	if err := h.shopService.PayOrder(orderID, req.PaymentID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "订单支付成功", nil)
}

// Review Handlers

// GetProductReviews 获取商品评价
func (h *ShopHandler) GetProductReviews(c *gin.Context) {
	resp := response.NewGinResponse()
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		params["limit"] = limit
	}

	reviews, total, err := h.shopService.GetProductReviews(productID, params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := params["page"].(int)
	limit := params["limit"].(int)

	resp.Success(c, gin.H{
		"items":     reviews,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// CreateProductReview 创建商品评价
func (h *ShopHandler) CreateProductReview(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	var req struct {
		Rating      int      `json:"rating" binding:"required,min=1,max=5"`
		Comment     string   `json:"comment" binding:"required"`
		Images      []string `json:"images"`
		IsAnonymous bool     `json:"is_anonymous"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// Convert images to JSON
	imagesJSON, _ := json.Marshal(req.Images)

	review := &models.ProductReview{
		ProductID:   productID,
		UserID:      user.ID,
		Rating:      req.Rating,
		Comment:     req.Comment,
		Images:      imagesJSON,
		IsAnonymous: req.IsAnonymous,
	}

	if err := h.shopService.CreateReview(review); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, review)
}

// Favorite Handlers

// GetFavorites 获取收藏列表
func (h *ShopHandler) GetFavorites(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		params["limit"] = limit
	}

	favorites, total, err := h.shopService.GetFavorites(user.ID, params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := params["page"].(int)
	limit := params["limit"].(int)

	resp.Success(c, gin.H{
		"items":     favorites,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// AddToFavorites 添加收藏
func (h *ShopHandler) AddToFavorites(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.shopService.AddToFavorites(user.ID, productID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "收藏成功", nil)
}

// RemoveFromFavorites 取消收藏
func (h *ShopHandler) RemoveFromFavorites(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.shopService.RemoveFromFavorites(user.ID, productID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "已取消收藏", nil)
}

// GetShopStatistics 获取商店统计（管理员）
func (h *ShopHandler) GetShopStatistics(c *gin.Context) {
	resp := response.NewGinResponse()
	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看统计数据")
		return
	}

	stats, err := h.shopService.GetShopStatistics()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}