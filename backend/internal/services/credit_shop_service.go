package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditShopService 积分商城服务
type CreditShopService struct {
	db            *gorm.DB
	creditService *CreditService
	limiterService *CreditLimiterService
}

// NewCreditShopService 创建积分商城服务实例
func NewCreditShopService(db *gorm.DB, creditService *CreditService, limiterService *CreditLimiterService) *CreditShopService {
	return &CreditShopService{
		db:            db,
		creditService: creditService,
		limiterService: limiterService,
	}
}

// ===================== 商品管理 =====================

// CreateProduct 创建积分商城商品（管理员）
func (s *CreditShopService) CreateProduct(product *models.CreditShopProduct) error {
	// 设置默认值
	if product.Status == "" {
		product.Status = models.CreditProductStatusActive
	}
	
	// 验证商品数据
	if err := s.validateProduct(product); err != nil {
		return err
	}
	
	return s.db.Create(product).Error
}

// GetProductByID 获取积分商城商品详情
func (s *CreditShopService) GetProductByID(id uuid.UUID) (*models.CreditShopProduct, error) {
	var product models.CreditShopProduct
	err := s.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetProducts 获取积分商城商品列表
func (s *CreditShopService) GetProducts(params map[string]interface{}) ([]models.CreditShopProduct, int64, error) {
	var products []models.CreditShopProduct
	var total int64

	query := s.db.Model(&models.CreditShopProduct{}).Where("deleted_at IS NULL")

	// 应用过滤条件
	if status, ok := params["status"].(models.CreditShopProductStatus); ok && status != "" {
		query = query.Where("status = ?", status)
	} else {
		// 默认只显示上架商品
		query = query.Where("status = ?", models.CreditProductStatusActive)
	}
	
	if category, ok := params["category"].(string); ok && category != "" {
		query = query.Where("category = ?", category)
	}
	
	if productType, ok := params["product_type"].(models.CreditShopProductType); ok && productType != "" {
		query = query.Where("product_type = ?", productType)
	}
	
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ? OR short_desc LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	if inStockOnly, ok := params["in_stock_only"].(bool); ok && inStockOnly {
		query = query.Where("stock > 0")
	}
	
	if featuredOnly, ok := params["featured_only"].(bool); ok && featuredOnly {
		query = query.Where("is_featured = ?", true)
	}
	
	if minPrice, ok := params["min_credit_price"].(int); ok && minPrice > 0 {
		query = query.Where("credit_price >= ?", minPrice)
	}
	
	if maxPrice, ok := params["max_credit_price"].(int); ok && maxPrice > 0 {
		query = query.Where("credit_price <= ?", maxPrice)
	}

	// 检查有效期
	now := time.Now()
	query = query.Where("(valid_from IS NULL OR valid_from <= ?) AND (valid_to IS NULL OR valid_to >= ?)", now, now)

	// 计算总数
	query.Count(&total)

	// 排序
	sortBy := "priority DESC, created_at DESC"
	if sb, ok := params["sort_by"].(string); ok && sb != "" {
		switch sb {
		case "price_asc":
			sortBy = "credit_price ASC"
		case "price_desc":
			sortBy = "credit_price DESC"
		case "popularity":
			sortBy = "redeem_count DESC"
		case "newest":
			sortBy = "created_at DESC"
		case "oldest":
			sortBy = "created_at ASC"
		}
	}
	query = query.Order(sortBy)

	// 分页
	page := 1
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 20
	if l, ok := params["limit"].(int); ok && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Find(&products).Error
	return products, total, err
}

// UpdateProduct 更新积分商城商品（管理员）
func (s *CreditShopService) UpdateProduct(id uuid.UUID, updates map[string]interface{}) error {
	// 验证商品存在
	var product models.CreditShopProduct
	if err := s.db.First(&product, id).Error; err != nil {
		return err
	}

	// 验证更新数据
	if creditPrice, ok := updates["credit_price"]; ok {
		if price, ok := creditPrice.(int); !ok || price < 0 {
			return errors.New("invalid credit_price")
		}
	}

	if status, ok := updates["status"]; ok {
		if statusStr, ok := status.(string); ok {
			validStatuses := []models.CreditShopProductStatus{
				models.CreditProductStatusDraft,
				models.CreditProductStatusActive,
				models.CreditProductStatusInactive,
				models.CreditProductStatusSoldOut,
				models.CreditProductStatusDeleted,
			}
			isValid := false
			for _, s := range validStatuses {
				if models.CreditShopProductStatus(statusStr) == s {
					isValid = true
					break
				}
			}
			if !isValid {
				return errors.New("invalid status")
			}
		}
	}

	return s.db.Model(&product).Updates(updates).Error
}

// DeleteProduct 删除积分商城商品（软删除）
func (s *CreditShopService) DeleteProduct(id uuid.UUID) error {
	return s.db.Model(&models.CreditShopProduct{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.CreditProductStatusDeleted,
			"deleted_at": time.Now(),
		}).Error
}

// ===================== 购物车管理 =====================

// GetOrCreateCreditCart 获取或创建积分购物车
func (s *CreditShopService) GetOrCreateCreditCart(userID string) (*models.CreditCart, error) {
	var cart models.CreditCart
	err := s.db.Where("user_id = ?", userID).Preload("Items.Product").First(&cart).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新购物车
		cart = models.CreditCart{
			UserID: userID,
		}
		if err := s.db.Create(&cart).Error; err != nil {
			return nil, err
		}
		return &cart, nil
	}

	if err != nil {
		return nil, err
	}

	return &cart, nil
}

// AddToCreditCart 添加商品到积分购物车
func (s *CreditShopService) AddToCreditCart(userID string, productID uuid.UUID, quantity int) (*models.CreditCartItem, error) {
	// 验证数量
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	// 获取商品信息
	product, err := s.GetProductByID(productID)
	if err != nil {
		return nil, err
	}

	// 检查商品是否可用
	if !product.IsAvailable() {
		return nil, errors.New("product is not available")
	}

	// 检查库存
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// 获取或创建购物车
	cart, err := s.GetOrCreateCreditCart(userID)
	if err != nil {
		return nil, err
	}

	// 检查购物车商品数量限制
	config, _ := s.GetCreditShopConfig("max_cart_items")
	maxItems := 20
	if config != nil {
		if maxItemsInt, ok := config["max_cart_items"].(int); ok {
			maxItems = maxItemsInt
		}
	}

	if cart.TotalItems >= maxItems {
		return nil, fmt.Errorf("cart exceeds maximum %d items", maxItems)
	}

	// 检查商品是否已在购物车中
	var cartItem models.CreditCartItem
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&cartItem).Error

	if err == gorm.ErrRecordNotFound {
		// 检查每用户限购
		if product.LimitPerUser > 0 {
			userRedemptionCount, _ := s.getUserProductRedemptionCount(userID, productID)
			if userRedemptionCount+quantity > product.LimitPerUser {
				return nil, fmt.Errorf("exceeds per-user limit of %d", product.LimitPerUser)
			}
		}

		// 创建新的购物车项目
		cartItem = models.CreditCartItem{
			CartID:      cart.ID,
			ProductID:   productID,
			Quantity:    quantity,
			CreditPrice: product.CreditPrice,
			Subtotal:    product.CreditPrice * quantity,
		}
		if err := s.db.Create(&cartItem).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		// 更新已存在的购物车项目
		newQuantity := cartItem.Quantity + quantity
		
		// 再次检查库存和限购
		if product.Stock < newQuantity {
			return nil, errors.New("insufficient stock")
		}
		
		if product.LimitPerUser > 0 {
			userRedemptionCount, _ := s.getUserProductRedemptionCount(userID, productID)
			if userRedemptionCount+newQuantity > product.LimitPerUser {
				return nil, fmt.Errorf("exceeds per-user limit of %d", product.LimitPerUser)
			}
		}

		cartItem.Quantity = newQuantity
		cartItem.Subtotal = cartItem.CreditPrice * newQuantity
		if err := s.db.Save(&cartItem).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 更新购物车总计
	s.UpdateCreditCartTotals(cart.ID)

	// 重新加载购物车项目并关联商品信息
	s.db.Preload("Product").First(&cartItem, cartItem.ID)

	return &cartItem, nil
}

// UpdateCreditCartItem 更新积分购物车项目
func (s *CreditShopService) UpdateCreditCartItem(userID string, itemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	var cartItem models.CreditCartItem
	err := s.db.Joins("JOIN credit_carts ON credit_carts.id = credit_cart_items.cart_id").
		Where("credit_cart_items.id = ? AND credit_carts.user_id = ?", itemID, userID).
		First(&cartItem).Error

	if err != nil {
		return err
	}

	// 检查库存
	var product models.CreditShopProduct
	if err := s.db.First(&product, cartItem.ProductID).Error; err != nil {
		return err
	}

	if !product.IsAvailable() {
		return errors.New("product is no longer available")
	}

	if product.Stock < quantity {
		return errors.New("insufficient stock")
	}

	// 检查每用户限购
	if product.LimitPerUser > 0 {
		userRedemptionCount, _ := s.getUserProductRedemptionCount(userID, cartItem.ProductID)
		if userRedemptionCount+quantity > product.LimitPerUser {
			return fmt.Errorf("exceeds per-user limit of %d", product.LimitPerUser)
		}
	}

	// 更新数量和小计
	cartItem.Quantity = quantity
	cartItem.Subtotal = cartItem.CreditPrice * quantity

	if err := s.db.Save(&cartItem).Error; err != nil {
		return err
	}

	// 更新购物车总计
	s.UpdateCreditCartTotals(cartItem.CartID)

	return nil
}

// RemoveFromCreditCart 从积分购物车移除商品
func (s *CreditShopService) RemoveFromCreditCart(userID string, itemID uuid.UUID) error {
	var cartItem models.CreditCartItem
	err := s.db.Joins("JOIN credit_carts ON credit_carts.id = credit_cart_items.cart_id").
		Where("credit_cart_items.id = ? AND credit_carts.user_id = ?", itemID, userID).
		First(&cartItem).Error

	if err != nil {
		return err
	}

	cartID := cartItem.CartID

	if err := s.db.Delete(&cartItem).Error; err != nil {
		return err
	}

	// 更新购物车总计
	s.UpdateCreditCartTotals(cartID)

	return nil
}

// ClearCreditCart 清空积分购物车
func (s *CreditShopService) ClearCreditCart(userID string) error {
	cart, err := s.GetOrCreateCreditCart(userID)
	if err != nil {
		return err
	}

	// 删除所有购物车项目
	if err := s.db.Where("cart_id = ?", cart.ID).Delete(&models.CreditCartItem{}).Error; err != nil {
		return err
	}

	// 更新购物车总计
	cart.TotalItems = 0
	cart.TotalCredits = 0
	return s.db.Save(&cart).Error
}

// UpdateCreditCartTotals 更新积分购物车总计
func (s *CreditShopService) UpdateCreditCartTotals(cartID uuid.UUID) error {
	var cart models.CreditCart
	if err := s.db.First(&cart, cartID).Error; err != nil {
		return err
	}

	// 计算总计
	var totalItems int
	var totalCredits int

	rows, err := s.db.Model(&models.CreditCartItem{}).
		Where("cart_id = ?", cartID).
		Select("SUM(quantity) as total_items, SUM(subtotal) as total_credits").
		Rows()

	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&totalItems, &totalCredits)
	}

	// 更新购物车
	cart.TotalItems = totalItems
	cart.TotalCredits = totalCredits

	return s.db.Save(&cart).Error
}

// ===================== 分类管理 =====================

// CreateCategory 创建积分商城分类（管理员）
func (s *CreditShopService) CreateCategory(category *models.CreditShopCategory) error {
	return s.db.Create(category).Error
}

// GetCategories 获取积分商城分类列表
func (s *CreditShopService) GetCategories(includeInactive bool) ([]models.CreditShopCategory, error) {
	var categories []models.CreditShopCategory
	
	query := s.db.Model(&models.CreditShopCategory{})
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("sort_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

// GetCategoryByID 获取分类详情
func (s *CreditShopService) GetCategoryByID(id uuid.UUID) (*models.CreditShopCategory, error) {
	var category models.CreditShopCategory
	err := s.db.Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// UpdateCategory 更新分类（管理员）
func (s *CreditShopService) UpdateCategory(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.CreditShopCategory{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteCategory 删除分类（管理员）
func (s *CreditShopService) DeleteCategory(id uuid.UUID) error {
	// 检查是否有商品使用该分类
	var count int64
	s.db.Model(&models.CreditShopProduct{}).Where("category = ?", id.String()).Count(&count)
	if count > 0 {
		return errors.New("category is being used by products")
	}
	
	return s.db.Delete(&models.CreditShopCategory{}, id).Error
}

// ===================== 配置管理 =====================

// GetCreditShopConfig 获取积分商城配置
func (s *CreditShopService) GetCreditShopConfig(keys ...string) (map[string]interface{}, error) {
	var configs []models.CreditShopConfig
	
	query := s.db.Model(&models.CreditShopConfig{})
	if len(keys) > 0 {
		query = query.Where("key IN (?)", keys)
	}
	
	if err := query.Find(&configs).Error; err != nil {
		return nil, err
	}
	
	result := make(map[string]interface{})
	for _, config := range configs {
		// 尝试解析JSON，如果失败则作为字符串处理
		var value interface{}
		if err := json.Unmarshal([]byte(config.Value), &value); err != nil {
			result[config.Key] = config.Value
		} else {
			result[config.Key] = value
		}
	}
	
	return result, nil
}

// UpdateCreditShopConfig 更新积分商城配置（管理员）
func (s *CreditShopService) UpdateCreditShopConfig(key, value string) error {
	config := models.CreditShopConfig{
		Key:   key,
		Value: value,
	}
	
	return s.db.Save(&config).Error
}

// ===================== 统计信息 =====================

// GetCreditShopStatistics 获取积分商城统计数据（管理员）
func (s *CreditShopService) GetCreditShopStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 商品统计
	var productCount, activeProductCount int64
	s.db.Model(&models.CreditShopProduct{}).Where("deleted_at IS NULL").Count(&productCount)
	s.db.Model(&models.CreditShopProduct{}).Where("status = ? AND deleted_at IS NULL", models.CreditProductStatusActive).Count(&activeProductCount)
	stats["product_count"] = productCount
	stats["active_product_count"] = activeProductCount

	// 兑换统计
	var totalRedemptions int64
	var totalCreditsUsed int64
	s.db.Model(&models.CreditRedemption{}).Where("status != ?", models.RedemptionStatusCancelled).Count(&totalRedemptions)
	s.db.Model(&models.CreditRedemption{}).Where("status != ?", models.RedemptionStatusCancelled).Select("SUM(total_credits)").Row().Scan(&totalCreditsUsed)
	stats["total_redemptions"] = totalRedemptions
	stats["total_credits_used"] = totalCreditsUsed

	// 用户统计
	var activeUsers int64
	s.db.Model(&models.UserRedemptionHistory{}).Where("total_redemptions > 0").Count(&activeUsers)
	stats["active_users"] = activeUsers

	// 热门商品
	var popularProducts []struct {
		ProductID   uuid.UUID `json:"product_id"`
		Name        string    `json:"name"`
		RedeemCount int       `json:"redeem_count"`
		CreditPrice int       `json:"credit_price"`
	}
	s.db.Model(&models.CreditShopProduct{}).
		Select("id as product_id, name, redeem_count, credit_price").
		Where("status = ? AND deleted_at IS NULL", models.CreditProductStatusActive).
		Order("redeem_count desc").
		Limit(10).
		Find(&popularProducts)
	stats["popular_products"] = popularProducts

	return stats, nil
}

// ===================== 辅助方法 =====================

// validateProduct 验证商品数据
func (s *CreditShopService) validateProduct(product *models.CreditShopProduct) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	
	if product.CreditPrice < 0 {
		return errors.New("credit price cannot be negative")
	}
	
	if product.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	
	// 验证商品类型
	validTypes := []models.CreditShopProductType{
		models.CreditProductTypePhysical,
		models.CreditProductTypeVirtual,
		models.CreditProductTypeService,
		models.CreditProductTypeVoucher,
	}
	
	isValidType := false
	for _, t := range validTypes {
		if product.ProductType == t {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return errors.New("invalid product type")
	}
	
	// 验证有效期
	if product.ValidFrom != nil && product.ValidTo != nil {
		if product.ValidTo.Before(*product.ValidFrom) {
			return errors.New("valid_to must be after valid_from")
		}
	}
	
	return nil
}

// getUserProductRedemptionCount 获取用户对指定商品的兑换次数
func (s *CreditShopService) getUserProductRedemptionCount(userID string, productID uuid.UUID) (int, error) {
	var count int64
	err := s.db.Model(&models.CreditRedemption{}).
		Where("user_id = ? AND product_id = ? AND status NOT IN (?)", 
			userID, productID, []models.CreditRedemptionStatus{
				models.RedemptionStatusCancelled,
				models.RedemptionStatusRefunded,
			}).
		Count(&count).Error
	return int(count), err
}

// ===================== 兑换订单系统 =====================

// CreateCreditRedemption 创建积分兑换订单
func (s *CreditShopService) CreateCreditRedemption(userID string, redemptionData map[string]interface{}) (*models.CreditRedemption, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 解析兑换数据
	productID, err := uuid.Parse(redemptionData["product_id"].(string))
	if err != nil {
		tx.Rollback()
		return nil, errors.New("invalid product ID")
	}

	quantity, ok := redemptionData["quantity"].(int)
	if !ok || quantity <= 0 {
		tx.Rollback()
		return nil, errors.New("invalid quantity")
	}

	// 获取商品信息
	var product models.CreditShopProduct
	if err := tx.First(&product, productID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// 验证商品可用性
	if !product.IsAvailable() {
		tx.Rollback()
		return nil, errors.New("product is not available")
	}

	// 检查库存
	if product.Stock < quantity {
		tx.Rollback()
		return nil, errors.New("insufficient stock")
	}

	// 检查用户限购
	if product.LimitPerUser > 0 {
		userRedemptionCount, _ := s.getUserProductRedemptionCount(userID, productID)
		if userRedemptionCount+quantity > product.LimitPerUser {
			tx.Rollback()
			return nil, fmt.Errorf("exceeds per-user limit of %d", product.LimitPerUser)
		}
	}

	// 计算总积分
	totalCredits := product.CreditPrice * quantity

	// 获取用户积分信息
	userCredit, err := s.creditService.GetUserCredit(userID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get user credit: %w", err)
	}

	// 检查积分余额
	if userCredit.Available < totalCredits {
		tx.Rollback()
		return nil, errors.New("insufficient credits")
	}

	// 应用积分限制检查
	if s.limiterService != nil {
		limitStatus, err := s.limiterService.CheckLimit(userID, "credit_redemption", totalCredits)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to check limits: %w", err)
		}
		if !limitStatus.Allowed {
			tx.Rollback()
			return nil, fmt.Errorf("redemption blocked: %s", limitStatus.Reason)
		}
	}

	// 创建兑换订单
	redemption := models.CreditRedemption{
		UserID:       userID,
		ProductID:    productID,
		Quantity:     quantity,
		CreditPrice:  product.CreditPrice,
		TotalCredits: totalCredits,
		Status:       models.RedemptionStatusPending,
	}

	// 处理配送信息
	if deliveryInfo, ok := redemptionData["delivery_info"].(map[string]interface{}); ok {
		deliveryInfoBytes, _ := json.Marshal(deliveryInfo)
		redemption.DeliveryInfo = deliveryInfoBytes
	}

	// 处理备注
	if notes, ok := redemptionData["notes"].(string); ok {
		redemption.Notes = notes
	}

	// 保存兑换订单
	if err := tx.Create(&redemption).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	// 扣除积分
	if err := s.creditService.DeductCredits(userID, totalCredits, fmt.Sprintf("兑换商品: %s", product.Name), redemption.ID.String()); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to deduct credits: %w", err)
	}

	// 更新商品库存和兑换次数
	if err := tx.Model(&product).Updates(map[string]interface{}{
		"stock":        product.Stock - quantity,
		"redeem_count": product.RedeemCount + quantity,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// 生成兑换码（虚拟商品）
	if product.ProductType == models.CreditProductTypeVirtual {
		redemptionCode := s.generateRedemptionCode(product.Name, redemption.ID)
		redemption.RedemptionCode = redemptionCode
		
		// 检查是否自动确认虚拟商品
		config, _ := s.GetCreditShopConfig("auto_confirm_virtual")
		if config != nil {
			if autoConfirm, ok := config["auto_confirm_virtual"].(bool); ok && autoConfirm {
				redemption.Status = models.RedemptionStatusCompleted
				now := time.Now()
				redemption.ProcessedAt = &now
				redemption.CompletedAt = &now
			}
		}
		
		if err := tx.Save(&redemption).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update redemption: %w", err)
		}
	}

	// 更新用户兑换历史统计
	s.updateUserRedemptionHistory(tx, userID, totalCredits, product.Category)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 重新加载兑换订单并包含关联数据
	s.db.Preload("Product").Preload("User").First(&redemption, redemption.ID)

	return &redemption, nil
}

// CreateCreditRedemptionFromCart 从购物车创建兑换订单
func (s *CreditShopService) CreateCreditRedemptionFromCart(userID string, deliveryData map[string]interface{}) ([]*models.CreditRedemption, error) {
	// 获取用户购物车
	cart, err := s.GetOrCreateCreditCart(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	var redemptions []*models.CreditRedemption

	// 为每个购物车项目创建兑换订单
	for _, item := range cart.Items {
		redemptionData := map[string]interface{}{
			"product_id":    item.ProductID.String(),
			"quantity":      item.Quantity,
			"delivery_info": deliveryData,
		}

		redemption, err := s.CreateCreditRedemption(userID, redemptionData)
		if err != nil {
			// 如果有任何失败，需要回滚已创建的兑换
			for _, r := range redemptions {
				s.CancelCreditRedemption(userID, r.ID)
			}
			return nil, fmt.Errorf("failed to create redemption for product %s: %w", item.ProductID, err)
		}

		redemptions = append(redemptions, redemption)
	}

	// 清空购物车
	if err := s.ClearCreditCart(userID); err != nil {
		log.Printf("Warning: failed to clear cart after redemption: %v", err)
	}

	return redemptions, nil
}

// GetCreditRedemptions 获取用户兑换订单列表
func (s *CreditShopService) GetCreditRedemptions(userID string, params map[string]interface{}) ([]models.CreditRedemption, int64, error) {
	var redemptions []models.CreditRedemption
	var total int64

	query := s.db.Model(&models.CreditRedemption{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if status, ok := params["status"].(models.CreditRedemptionStatus); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if productType, ok := params["product_type"].(models.CreditShopProductType); ok && productType != "" {
		query = query.Joins("JOIN credit_shop_products ON credit_shop_products.id = credit_redemptions.product_id").
			Where("credit_shop_products.product_type = ?", productType)
	}

	// 计算总数
	query.Count(&total)

	// 排序
	query = query.Order("created_at desc")

	// 分页
	page := 1
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 20
	if l, ok := params["limit"].(int); ok && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Preload("Product").Find(&redemptions).Error
	return redemptions, total, err
}

// GetCreditRedemptionByID 获取兑换订单详情
func (s *CreditShopService) GetCreditRedemptionByID(userID string, redemptionID uuid.UUID) (*models.CreditRedemption, error) {
	var redemption models.CreditRedemption
	err := s.db.Where("id = ? AND user_id = ?", redemptionID, userID).
		Preload("Product").
		Preload("User").
		First(&redemption).Error

	if err != nil {
		return nil, err
	}

	return &redemption, nil
}

// UpdateCreditRedemptionStatus 更新兑换订单状态（管理员）
func (s *CreditShopService) UpdateCreditRedemptionStatus(redemptionID uuid.UUID, status models.CreditRedemptionStatus, adminNote string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var redemption models.CreditRedemption
	if err := tx.First(&redemption, redemptionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 验证状态转换的合法性
	if !s.isValidStatusTransition(redemption.Status, status) {
		tx.Rollback()
		return fmt.Errorf("invalid status transition from %s to %s", redemption.Status, status)
	}

	updates := map[string]interface{}{
		"status": status,
		"notes":  redemption.Notes + "\n[管理员]: " + adminNote,
	}

	// 根据状态更新相应的时间戳
	now := time.Now()
	switch status {
	case models.RedemptionStatusConfirmed:
		updates["processed_at"] = &now
	case models.RedemptionStatusShipped:
		updates["shipped_at"] = &now
	case models.RedemptionStatusDelivered:
		updates["delivered_at"] = &now
	case models.RedemptionStatusCompleted:
		updates["completed_at"] = &now
	case models.RedemptionStatusCancelled:
		updates["cancelled_at"] = &now
		// 取消时退还积分
		if err := s.refundCreditsForRedemption(tx, &redemption); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund credits: %w", err)
		}
	case models.RedemptionStatusRefunded:
		// 退款时退还积分
		if err := s.refundCreditsForRedemption(tx, &redemption); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund credits: %w", err)
		}
	}

	if err := tx.Model(&redemption).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CancelCreditRedemption 取消兑换订单（用户）
func (s *CreditShopService) CancelCreditRedemption(userID string, redemptionID uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var redemption models.CreditRedemption
	if err := tx.Where("id = ? AND user_id = ?", redemptionID, userID).First(&redemption).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 检查是否可以取消
	if redemption.Status != models.RedemptionStatusPending && redemption.Status != models.RedemptionStatusConfirmed {
		tx.Rollback()
		return errors.New("redemption cannot be cancelled at current status")
	}

	// 更新状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.RedemptionStatusCancelled,
		"cancelled_at": &now,
		"notes":        redemption.Notes + "\n[用户取消]",
	}

	if err := tx.Model(&redemption).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 退还积分
	if err := s.refundCreditsForRedemption(tx, &redemption); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund credits: %w", err)
	}

	// 恢复商品库存
	if err := s.restoreProductStock(tx, &redemption); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to restore stock: %w", err)
	}

	return tx.Commit().Error
}

// GetAllCreditRedemptions 获取所有兑换订单（管理员）
func (s *CreditShopService) GetAllCreditRedemptions(params map[string]interface{}) ([]models.CreditRedemption, int64, error) {
	var redemptions []models.CreditRedemption
	var total int64

	query := s.db.Model(&models.CreditRedemption{})

	// 应用过滤条件
	if status, ok := params["status"].(models.CreditRedemptionStatus); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if userID, ok := params["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if productType, ok := params["product_type"].(models.CreditShopProductType); ok && productType != "" {
		query = query.Joins("JOIN credit_shop_products ON credit_shop_products.id = credit_redemptions.product_id").
			Where("credit_shop_products.product_type = ?", productType)
	}

	// 计算总数
	query.Count(&total)

	// 排序
	sortBy := "created_at desc"
	if sb, ok := params["sort_by"].(string); ok && sb != "" {
		switch sb {
		case "amount_desc":
			sortBy = "total_credits desc"
		case "amount_asc":
			sortBy = "total_credits asc"
		case "status":
			sortBy = "status asc, created_at desc"
		}
	}
	query = query.Order(sortBy)

	// 分页
	page := 1
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 20
	if l, ok := params["limit"].(int); ok && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Preload("Product").Preload("User").Find(&redemptions).Error
	return redemptions, total, err
}

// ===================== 辅助方法 =====================

// generateRedemptionCode 生成兑换码
func (s *CreditShopService) generateRedemptionCode(productName string, redemptionID uuid.UUID) string {
	prefix := "RC"
	if len(productName) >= 2 {
		prefix = strings.ToUpper(productName[:2])
	}
	return fmt.Sprintf("%s-%s-%s", 
		prefix, 
		time.Now().Format("0102"), 
		redemptionID.String()[:8])
}

// isValidStatusTransition 检查状态转换是否合法
func (s *CreditShopService) isValidStatusTransition(from, to models.CreditRedemptionStatus) bool {
	validTransitions := map[models.CreditRedemptionStatus][]models.CreditRedemptionStatus{
		models.RedemptionStatusPending: {
			models.RedemptionStatusConfirmed,
			models.RedemptionStatusCancelled,
		},
		models.RedemptionStatusConfirmed: {
			models.RedemptionStatusProcessing,
			models.RedemptionStatusShipped,
			models.RedemptionStatusCompleted,
			models.RedemptionStatusCancelled,
		},
		models.RedemptionStatusProcessing: {
			models.RedemptionStatusShipped,
			models.RedemptionStatusDelivered,
			models.RedemptionStatusCompleted,
		},
		models.RedemptionStatusShipped: {
			models.RedemptionStatusDelivered,
			models.RedemptionStatusCompleted,
		},
		models.RedemptionStatusDelivered: {
			models.RedemptionStatusCompleted,
			models.RedemptionStatusRefunded,
		},
		models.RedemptionStatusCompleted: {
			models.RedemptionStatusRefunded,
		},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if to == allowed {
			return true
		}
	}

	return false
}

// refundCreditsForRedemption 为兑换退还积分
func (s *CreditShopService) refundCreditsForRedemption(tx *gorm.DB, redemption *models.CreditRedemption) error {
	// 加载商品信息
	var product models.CreditShopProduct
	if err := tx.First(&product, redemption.ProductID).Error; err != nil {
		return err
	}

	// 退还积分
	return s.creditService.AddCredits(
		redemption.UserID,
		redemption.TotalCredits,
		fmt.Sprintf("退款: %s", product.Name),
		redemption.ID.String(),
	)
}

// restoreProductStock 恢复商品库存
func (s *CreditShopService) restoreProductStock(tx *gorm.DB, redemption *models.CreditRedemption) error {
	return tx.Model(&models.CreditShopProduct{}).
		Where("id = ?", redemption.ProductID).
		UpdateColumn("stock", gorm.Expr("stock + ?", redemption.Quantity)).
		Error
}

// updateUserRedemptionHistory 更新用户兑换历史统计
func (s *CreditShopService) updateUserRedemptionHistory(tx *gorm.DB, userID string, creditsUsed int, category string) error {
	var history models.UserRedemptionHistory
	err := tx.Where("user_id = ?", userID).First(&history).Error

	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		history = models.UserRedemptionHistory{
			UserID:             userID,
			TotalRedemptions:   1,
			TotalCreditsUsed:   creditsUsed,
			LastRedemptionAt:   &now,
			FavoriteCategory:   category,
		}
		return tx.Create(&history).Error
	} else if err != nil {
		return err
	}

	// 更新现有记录
	updates := map[string]interface{}{
		"total_redemptions":   history.TotalRedemptions + 1,
		"total_credits_used":  history.TotalCreditsUsed + creditsUsed,
		"last_redemption_at":  &now,
		"favorite_category":   category, // 简化实现，直接使用最新分类
	}

	return tx.Model(&history).Updates(updates).Error
}