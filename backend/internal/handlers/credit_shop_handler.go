package handlers

import (
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"openpenpal-backend/internal/pkg/response"
)

// CreditShopHandler 积分商城处理器
type CreditShopHandler struct {
	creditShopService *services.CreditShopService
	creditService     *services.CreditService
}

// NewCreditShopHandler 创建积分商城处理器实例
func NewCreditShopHandler(creditShopService *services.CreditShopService, creditService *services.CreditService) *CreditShopHandler {
	return &CreditShopHandler{
		creditShopService: creditShopService,
		creditService:     creditService,
	}
}

// ===================== 商品管理 API =====================

// GetCreditShopProducts 获取积分商城商品列表
func (h *CreditShopHandler) GetCreditShopProducts(c *gin.Context) {
	resp := response.NewGinResponse()
	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if category := c.Query("category"); category != "" {
		params["category"] = category
	}
	if productType := c.Query("product_type"); productType != "" {
		params["product_type"] = models.CreditShopProductType(productType)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		params["keyword"] = keyword
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		params["sort_by"] = sortBy
	}
	if inStockOnly := c.Query("in_stock_only"); inStockOnly == "true" {
		params["in_stock_only"] = true
	}
	if featuredOnly := c.Query("featured_only"); featuredOnly == "true" {
		params["featured_only"] = true
	}
	if minPrice, err := strconv.Atoi(c.Query("min_credit_price")); err == nil && minPrice >= 0 {
		params["min_credit_price"] = minPrice
	}
	if maxPrice, err := strconv.Atoi(c.Query("max_credit_price")); err == nil && maxPrice > 0 {
		params["max_credit_price"] = maxPrice
	}

	// 管理员可以查看所有状态的商品
	if user, exists := c.Get("user"); exists {
		if userModel, ok := user.(*models.User); ok {
			if userModel.Role == "admin" || userModel.Role == "super_admin" {
				if status := c.Query("status"); status != "" {
					params["status"] = models.CreditShopProductStatus(status)
				}
			}
		}
	}

	products, total, err := h.creditShopService.GetProducts(params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := 1
	limit := 20
	if p, ok := params["page"]; ok {
		page = p.(int)
	}
	if l, ok := params["limit"]; ok {
		limit = l.(int)
	}
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     products,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// GetCreditShopProduct 获取积分商城商品详情
func (h *CreditShopHandler) GetCreditShopProduct(c *gin.Context) {
	resp := response.NewGinResponse()

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	product, err := h.creditShopService.GetProductByID(productID)
	if err != nil {
		resp.NotFound(c, "商品不存在")
		return
	}

	resp.Success(c, product)
}

// CreateCreditShopProduct 创建积分商城商品（管理员）
func (h *CreditShopHandler) CreateCreditShopProduct(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建商品")
		return
	}

	var product models.CreditShopProduct
	if err := c.ShouldBindJSON(&product); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.CreateProduct(&product); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "商品创建成功", product)
}

// UpdateCreditShopProduct 更新积分商城商品（管理员）
func (h *CreditShopHandler) UpdateCreditShopProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	
	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新商品")
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.UpdateProduct(productID, updates); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品更新成功", nil)
}

// DeleteCreditShopProduct 删除积分商城商品（管理员）
func (h *CreditShopHandler) DeleteCreditShopProduct(c *gin.Context) {
	resp := response.NewGinResponse()
	
	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以删除商品")
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.creditShopService.DeleteProduct(productID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品删除成功", nil)
}

// ===================== 购物车管理 API =====================

// GetCreditCart 获取积分购物车
func (h *CreditShopHandler) GetCreditCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	cart, err := h.creditShopService.GetOrCreateCreditCart(user.ID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, cart)
}

// AddToCreditCart 添加商品到积分购物车
func (h *CreditShopHandler) AddToCreditCart(c *gin.Context) {
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

	cartItem, err := h.creditShopService.AddToCreditCart(user.ID, productID, req.Quantity)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, cartItem)
}

// UpdateCreditCartItem 更新积分购物车项目
func (h *CreditShopHandler) UpdateCreditCartItem(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的项目ID")
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.UpdateCreditCartItem(user.ID, itemID, req.Quantity); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "购物车更新成功", nil)
}

// RemoveFromCreditCart 从积分购物车移除商品
func (h *CreditShopHandler) RemoveFromCreditCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的项目ID")
		return
	}

	if err := h.creditShopService.RemoveFromCreditCart(user.ID, itemID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "商品已移除", nil)
}

// ClearCreditCart 清空积分购物车
func (h *CreditShopHandler) ClearCreditCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	if err := h.creditShopService.ClearCreditCart(user.ID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "购物车已清空", nil)
}

// ===================== 分类管理 API =====================

// GetCreditShopCategories 获取积分商城分类列表
func (h *CreditShopHandler) GetCreditShopCategories(c *gin.Context) {
	resp := response.NewGinResponse()

	includeInactive := false
	// 管理员可以查看已停用的分类
	if user, exists := c.Get("user"); exists {
		if userModel, ok := user.(*models.User); ok {
			if userModel.Role == "admin" || userModel.Role == "super_admin" {
				if c.Query("include_inactive") == "true" {
					includeInactive = true
				}
			}
		}
	}

	categories, err := h.creditShopService.GetCategories(includeInactive)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, categories)
}

// GetCreditShopCategory 获取积分商城分类详情
func (h *CreditShopHandler) GetCreditShopCategory(c *gin.Context) {
	resp := response.NewGinResponse()

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分类ID")
		return
	}

	category, err := h.creditShopService.GetCategoryByID(categoryID)
	if err != nil {
		resp.NotFound(c, "分类不存在")
		return
	}

	resp.Success(c, category)
}

// CreateCreditShopCategory 创建积分商城分类（管理员）
func (h *CreditShopHandler) CreateCreditShopCategory(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建分类")
		return
	}

	var category models.CreditShopCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.CreateCategory(&category); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "分类创建成功", category)
}

// UpdateCreditShopCategory 更新积分商城分类（管理员）
func (h *CreditShopHandler) UpdateCreditShopCategory(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新分类")
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分类ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.UpdateCategory(categoryID, updates); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "分类更新成功", nil)
}

// DeleteCreditShopCategory 删除积分商城分类（管理员）
func (h *CreditShopHandler) DeleteCreditShopCategory(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以删除分类")
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的分类ID")
		return
	}

	if err := h.creditShopService.DeleteCategory(categoryID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "分类删除成功", nil)
}

// ===================== 配置管理 API =====================

// GetCreditShopConfig 获取积分商城配置（管理员）
func (h *CreditShopHandler) GetCreditShopConfig(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看配置")
		return
	}

	// 获取指定的配置项，或所有配置
	keys := c.QueryArray("keys")
	
	configs, err := h.creditShopService.GetCreditShopConfig(keys...)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, configs)
}

// UpdateCreditShopConfig 更新积分商城配置（管理员）
func (h *CreditShopHandler) UpdateCreditShopConfig(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新配置")
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditShopService.UpdateCreditShopConfig(req.Key, req.Value); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "配置更新成功", nil)
}

// ===================== 统计信息 API =====================

// GetCreditShopStatistics 获取积分商城统计数据（管理员）
func (h *CreditShopHandler) GetCreditShopStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看统计数据")
		return
	}

	stats, err := h.creditShopService.GetCreditShopStatistics()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// ===================== 用户相关 API =====================

// GetUserCreditBalance 获取用户积分余额
func (h *CreditShopHandler) GetUserCreditBalance(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	// 获取用户积分信息
	userCredit, err := h.creditService.GetUserCredit(user.ID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"available": userCredit.Available,
		"total":     userCredit.Total,
		"used":      userCredit.Used,
		"level":     userCredit.Level,
	})
}

// ValidatePurchase 验证积分购买能力
func (h *CreditShopHandler) ValidatePurchase(c *gin.Context) {
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

	// 获取商品信息
	product, err := h.creditShopService.GetProductByID(productID)
	if err != nil {
		resp.NotFound(c, "商品不存在")
		return
	}

	// 获取用户积分
	userCredit, err := h.creditService.GetUserCredit(user.ID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	totalCredits := product.CreditPrice * req.Quantity
	canPurchase := true
	reason := ""

	// 检查各种限制条件
	if !product.IsAvailable() {
		canPurchase = false
		reason = "商品当前不可兑换"
	} else if product.Stock < req.Quantity {
		canPurchase = false
		reason = "库存不足"
	} else if userCredit.Available < totalCredits {
		canPurchase = false
		reason = "积分不足"
	} else if product.LimitPerUser > 0 {
		// 检查用户限购
		// 这里需要实现getUserProductRedemptionCount的公开版本
		// 暂时跳过详细检查
	}

	resp.Success(c, gin.H{
		"can_purchase":   canPurchase,
		"reason":         reason,
		"total_credits":  totalCredits,
		"user_balance":   userCredit.Available,
		"product_status": product.Status,
		"stock":          product.Stock,
	})
}

// ===================== 兑换订单管理 API =====================

// CreateCreditRedemption 创建积分兑换订单
func (h *CreditShopHandler) CreateCreditRedemption(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		ProductID    string                 `json:"product_id" binding:"required"`
		Quantity     int                    `json:"quantity" binding:"required,min=1"`
		DeliveryInfo map[string]interface{} `json:"delivery_info,omitempty"`
		Notes        string                 `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	redemptionData := map[string]interface{}{
		"product_id":    req.ProductID,
		"quantity":      req.Quantity,
		"delivery_info": req.DeliveryInfo,
		"notes":         req.Notes,
	}

	redemption, err := h.creditShopService.CreateCreditRedemption(user.ID, redemptionData)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "兑换订单创建成功", redemption)
}

// CreateCreditRedemptionFromCart 从购物车创建兑换订单
func (h *CreditShopHandler) CreateCreditRedemptionFromCart(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		DeliveryInfo map[string]interface{} `json:"delivery_info,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	redemptions, err := h.creditShopService.CreateCreditRedemptionFromCart(user.ID, req.DeliveryInfo)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "批量兑换订单创建成功", gin.H{
		"redemptions": redemptions,
		"count":       len(redemptions),
	})
}

// GetCreditRedemptions 获取用户兑换订单列表
func (h *CreditShopHandler) GetCreditRedemptions(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if status := c.Query("status"); status != "" {
		params["status"] = models.CreditRedemptionStatus(status)
	}
	if productType := c.Query("product_type"); productType != "" {
		params["product_type"] = models.CreditShopProductType(productType)
	}

	redemptions, total, err := h.creditShopService.GetCreditRedemptions(user.ID, params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := 1
	limit := 20
	if p, ok := params["page"]; ok {
		page = p.(int)
	}
	if l, ok := params["limit"]; ok {
		limit = l.(int)
	}
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     redemptions,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// GetCreditRedemption 获取兑换订单详情
func (h *CreditShopHandler) GetCreditRedemption(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	redemptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的兑换订单ID")
		return
	}

	redemption, err := h.creditShopService.GetCreditRedemptionByID(user.ID, redemptionID)
	if err != nil {
		resp.NotFound(c, "兑换订单不存在")
		return
	}

	resp.Success(c, redemption)
}

// CancelCreditRedemption 取消兑换订单
func (h *CreditShopHandler) CancelCreditRedemption(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	redemptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的兑换订单ID")
		return
	}

	if err := h.creditShopService.CancelCreditRedemption(user.ID, redemptionID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "兑换订单已取消", nil)
}

// ===================== 管理员兑换订单管理 =====================

// GetAllCreditRedemptions 获取所有兑换订单（管理员）
func (h *CreditShopHandler) GetAllCreditRedemptions(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看所有兑换订单")
		return
	}

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if status := c.Query("status"); status != "" {
		params["status"] = models.CreditRedemptionStatus(status)
	}
	if userID := c.Query("user_id"); userID != "" {
		params["user_id"] = userID
	}
	if productType := c.Query("product_type"); productType != "" {
		params["product_type"] = models.CreditShopProductType(productType)
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		params["sort_by"] = sortBy
	}

	redemptions, total, err := h.creditShopService.GetAllCreditRedemptions(params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := 1
	limit := 20
	if p, ok := params["page"]; ok {
		page = p.(int)
	}
	if l, ok := params["limit"]; ok {
		limit = l.(int)
	}
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     redemptions,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// UpdateCreditRedemptionStatus 更新兑换订单状态（管理员）
func (h *CreditShopHandler) UpdateCreditRedemptionStatus(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新兑换订单状态")
		return
	}

	redemptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的兑换订单ID")
		return
	}

	var req struct {
		Status    string `json:"status" binding:"required"`
		AdminNote string `json:"admin_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证状态是否有效
	status := models.CreditRedemptionStatus(req.Status)
	validStatuses := []models.CreditRedemptionStatus{
		models.RedemptionStatusPending,
		models.RedemptionStatusConfirmed,
		models.RedemptionStatusProcessing,
		models.RedemptionStatusShipped,
		models.RedemptionStatusDelivered,
		models.RedemptionStatusCompleted,
		models.RedemptionStatusCancelled,
		models.RedemptionStatusRefunded,
	}

	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		resp.BadRequest(c, "无效的兑换订单状态")
		return
	}

	if err := h.creditShopService.UpdateCreditRedemptionStatus(redemptionID, status, req.AdminNote); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "兑换订单状态更新成功", nil)
}