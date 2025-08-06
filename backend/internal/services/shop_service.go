package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopService 商店服务
type ShopService struct {
	db *gorm.DB
}

// NewShopService 创建商店服务实例
func NewShopService(db *gorm.DB) *ShopService {
	return &ShopService{db: db}
}

// Product Management

// CreateProduct 创建商品
func (s *ShopService) CreateProduct(product *models.Product) error {
	if product.OriginalPrice == 0 {
		product.OriginalPrice = product.Price
	}
	return s.db.Create(product).Error
}

// GetProductByID 获取商品详情
func (s *ShopService) GetProductByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := s.db.Where("id = ? AND status = ?", id, models.ProductStatusActive).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetProducts 获取商品列表
func (s *ShopService) GetProducts(params map[string]interface{}) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := s.db.Model(&models.Product{}).Where("status = ?", models.ProductStatusActive)

	// 应用过滤条件
	if category, ok := params["category"].(string); ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if productType, ok := params["product_type"].(models.ProductType); ok && productType != "" {
		query = query.Where("product_type = ?", productType)
	}
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if inStockOnly, ok := params["in_stock_only"].(bool); ok && inStockOnly {
		query = query.Where("stock > 0")
	}
	if featuredOnly, ok := params["featured_only"].(bool); ok && featuredOnly {
		query = query.Where("is_featured = ?", true)
	}
	if minPrice, ok := params["min_price"].(float64); ok && minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice, ok := params["max_price"].(float64); ok && maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}

	// 计算总数
	query.Count(&total)

	// 排序
	sortBy := "created_at"
	if sb, ok := params["sort_by"].(string); ok && sb != "" {
		sortBy = sb
	}
	order := "desc"
	if o, ok := params["order"].(string); ok && o != "" {
		order = o
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	// 分页
	page := 1
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 20
	if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Find(&products).Error
	return products, total, err
}

// UpdateProduct 更新商品
func (s *ShopService) UpdateProduct(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteProduct 删除商品（软删除）
func (s *ShopService) DeleteProduct(id uuid.UUID) error {
	return s.db.Model(&models.Product{}).Where("id = ?", id).Update("status", models.ProductStatusDeleted).Error
}

// Cart Management

// GetOrCreateCart 获取或创建购物车
func (s *ShopService) GetOrCreateCart(userID string) (*models.Cart, error) {
	var cart models.Cart
	err := s.db.Where("user_id = ?", userID).Preload("Items.Product").First(&cart).Error
	
	if err == gorm.ErrRecordNotFound {
		// 创建新购物车
		cart = models.Cart{
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

// AddToCart 添加商品到购物车
func (s *ShopService) AddToCart(userID string, productID uuid.UUID, quantity int) (*models.CartItem, error) {
	// 获取商品信息
	product, err := s.GetProductByID(productID)
	if err != nil {
		return nil, err
	}

	// 检查库存
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// 获取或创建购物车
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// 检查商品是否已在购物车中
	var cartItem models.CartItem
	err = s.db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&cartItem).Error
	
	if err == gorm.ErrRecordNotFound {
		// 创建新的购物车项目
		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: productID,
			Quantity:  quantity,
			Price:     product.Price,
			Subtotal:  product.Price * float64(quantity),
		}
		if err := s.db.Create(&cartItem).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		// 更新已存在的购物车项目
		cartItem.Quantity += quantity
		cartItem.Subtotal = cartItem.Price * float64(cartItem.Quantity)
		if err := s.db.Save(&cartItem).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 更新购物车总计
	s.UpdateCartTotals(cart.ID)

	// 重新加载购物车项目并关联商品信息
	s.db.Preload("Product").First(&cartItem, cartItem.ID)
	
	return &cartItem, nil
}

// UpdateCartItem 更新购物车项目
func (s *ShopService) UpdateCartItem(userID string, itemID uuid.UUID, quantity int) error {
	var cartItem models.CartItem
	err := s.db.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
		First(&cartItem).Error
	
	if err != nil {
		return err
	}

	// 检查库存
	var product models.Product
	if err := s.db.First(&product, cartItem.ProductID).Error; err != nil {
		return err
	}
	
	if product.Stock < quantity {
		return errors.New("insufficient stock")
	}

	// 更新数量和小计
	cartItem.Quantity = quantity
	cartItem.Subtotal = cartItem.Price * float64(quantity)
	
	if err := s.db.Save(&cartItem).Error; err != nil {
		return err
	}

	// 更新购物车总计
	s.UpdateCartTotals(cartItem.CartID)

	return nil
}

// RemoveFromCart 从购物车移除商品
func (s *ShopService) RemoveFromCart(userID string, itemID uuid.UUID) error {
	var cartItem models.CartItem
	err := s.db.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
		First(&cartItem).Error
	
	if err != nil {
		return err
	}

	cartID := cartItem.CartID
	
	if err := s.db.Delete(&cartItem).Error; err != nil {
		return err
	}

	// 更新购物车总计
	s.UpdateCartTotals(cartID)

	return nil
}

// ClearCart 清空购物车
func (s *ShopService) ClearCart(userID string) error {
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	// 删除所有购物车项目
	if err := s.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		return err
	}

	// 更新购物车总计
	cart.TotalItems = 0
	cart.TotalAmount = 0
	return s.db.Save(&cart).Error
}

// UpdateCartTotals 更新购物车总计
func (s *ShopService) UpdateCartTotals(cartID uuid.UUID) error {
	var cart models.Cart
	if err := s.db.First(&cart, cartID).Error; err != nil {
		return err
	}

	// 计算总计
	var totalItems int
	var totalAmount float64
	
	rows, err := s.db.Model(&models.CartItem{}).
		Where("cart_id = ?", cartID).
		Select("SUM(quantity) as total_items, SUM(subtotal) as total_amount").
		Rows()
	
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&totalItems, &totalAmount)
	}

	// 更新购物车
	cart.TotalItems = totalItems
	cart.TotalAmount = totalAmount
	
	return s.db.Save(&cart).Error
}

// Order Management

// CreateOrder 创建订单
func (s *ShopService) CreateOrder(userID string, orderData map[string]interface{}) (*models.Order, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取购物车
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(cart.Items) == 0 {
		tx.Rollback()
		return nil, errors.New("cart is empty")
	}

	// 创建订单
	order := models.Order{
		UserID:        userID,
		PaymentMethod: orderData["payment_method"].(string),
		Notes:         orderData["notes"].(string),
		Status:        models.OrderStatusPending,
		PaymentStatus: models.PaymentStatusPending,
	}

	// 处理 shipping address
	if shippingAddr, ok := orderData["shipping_address"].(map[string]interface{}); ok {
		addrBytes, _ := json.Marshal(shippingAddr)
		order.ShippingAddress = addrBytes
	}

	// 计算订单金额
	var subtotal float64
	var totalItems int

	// 创建订单项目并检查库存
	for _, cartItem := range cart.Items {
		// 重新检查库存
		var product models.Product
		if err := tx.First(&product, cartItem.ProductID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if product.Stock < cartItem.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock for product: %s", product.Name)
		}

		// 创建订单项目
		orderItem := models.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
			Subtotal:  cartItem.Subtotal,
		}

		order.Items = append(order.Items, orderItem)
		subtotal += cartItem.Subtotal
		totalItems += cartItem.Quantity

		// 更新商品库存和销量
		if err := tx.Model(&product).Updates(map[string]interface{}{
			"stock": product.Stock - cartItem.Quantity,
			"sold":  product.Sold + cartItem.Quantity,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 设置订单金额
	order.Subtotal = subtotal
	order.TotalItems = totalItems
	// 解析 shipping address JSON
	var shippingAddrMap map[string]interface{}
	if len(order.ShippingAddress) > 0 {
		json.Unmarshal(order.ShippingAddress, &shippingAddrMap)
	}
	order.ShippingFee = s.calculateShippingFee(shippingAddrMap)
	order.TaxFee = s.calculateTaxFee(subtotal)
	
	// 应用优惠券
	if couponCode, ok := orderData["coupon_code"].(string); ok && couponCode != "" {
		order.CouponCode = couponCode
		order.DiscountAmount = s.calculateCouponDiscount(couponCode, subtotal)
	}

	order.TotalAmount = subtotal + order.ShippingFee + order.TaxFee - order.DiscountAmount

	// 保存订单
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 清空购物车
	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 更新购物车总计
	cart.TotalItems = 0
	cart.TotalAmount = 0
	if err := tx.Save(&cart).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 重新加载订单并包含关联数据
	s.db.Preload("Items.Product").Preload("User").First(&order, order.ID)

	return &order, nil
}

// GetOrders 获取订单列表
func (s *ShopService) GetOrders(userID string, params map[string]interface{}) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{}).Where("user_id = ?", userID)

	// 应用过滤条件
	if status, ok := params["status"].(models.OrderStatus); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus, ok := params["payment_status"].(models.PaymentStatus); ok && paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
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
	if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Preload("Items.Product").Find(&orders).Error
	return orders, total, err
}

// GetOrderByID 获取订单详情
func (s *ShopService) GetOrderByID(userID string, orderID uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items.Product").
		Preload("User").
		First(&order).Error
	
	if err != nil {
		return nil, err
	}
	
	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *ShopService) UpdateOrderStatus(orderID uuid.UUID, status models.OrderStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态更新相应的时间戳
	now := time.Now()
	switch status {
	case models.OrderStatusPaid:
		updates["paid_at"] = &now
	case models.OrderStatusShipped:
		updates["shipped_at"] = &now
	case models.OrderStatusDelivered:
		updates["delivered_at"] = &now
	case models.OrderStatusCompleted:
		updates["completed_at"] = &now
	case models.OrderStatusCancelled:
		updates["cancelled_at"] = &now
	}

	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
}

// PayOrder 支付订单
func (s *ShopService) PayOrder(orderID uuid.UUID, paymentID string) error {
	updates := map[string]interface{}{
		"payment_status": models.PaymentStatusPaid,
		"payment_id":     paymentID,
		"status":         models.OrderStatusPaid,
		"paid_at":        time.Now(),
	}

	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
}

// Product Reviews

// CreateReview 创建商品评价
func (s *ShopService) CreateReview(review *models.ProductReview) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查用户是否购买过该商品
	var count int64
	tx.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.user_id = ? AND order_items.product_id = ? AND orders.status = ?", 
			review.UserID, review.ProductID, models.OrderStatusCompleted).
		Count(&count)
	
	if count == 0 {
		tx.Rollback()
		return errors.New("user has not purchased this product")
	}

	// 创建评价
	if err := tx.Create(review).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新商品评分
	var avgRating float64
	var reviewCount int64
	
	tx.Model(&models.ProductReview{}).
		Where("product_id = ?", review.ProductID).
		Select("AVG(rating) as avg_rating, COUNT(*) as review_count").
		Row().Scan(&avgRating, &reviewCount)

	// 更新商品
	if err := tx.Model(&models.Product{}).Where("id = ?", review.ProductID).Updates(map[string]interface{}{
		"rating":       avgRating,
		"review_count": reviewCount,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetProductReviews 获取商品评价列表
func (s *ShopService) GetProductReviews(productID uuid.UUID, params map[string]interface{}) ([]models.ProductReview, int64, error) {
	var reviews []models.ProductReview
	var total int64

	query := s.db.Model(&models.ProductReview{}).Where("product_id = ?", productID)

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
	if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Preload("User").Find(&reviews).Error
	return reviews, total, err
}

// Product Favorites

// AddToFavorites 添加到收藏
func (s *ShopService) AddToFavorites(userID string, productID uuid.UUID) error {
	// 检查是否已收藏
	var count int64
	s.db.Model(&models.ProductFavorite{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count)
	
	if count > 0 {
		return errors.New("product already in favorites")
	}

	favorite := models.ProductFavorite{
		UserID:    userID,
		ProductID: productID,
	}

	return s.db.Create(&favorite).Error
}

// RemoveFromFavorites 取消收藏
func (s *ShopService) RemoveFromFavorites(userID string, productID uuid.UUID) error {
	return s.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.ProductFavorite{}).Error
}

// GetFavorites 获取收藏列表
func (s *ShopService) GetFavorites(userID string, params map[string]interface{}) ([]models.ProductFavorite, int64, error) {
	var favorites []models.ProductFavorite
	var total int64

	query := s.db.Model(&models.ProductFavorite{}).Where("user_id = ?", userID)

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
	if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	err := query.Preload("Product").Find(&favorites).Error
	return favorites, total, err
}

// Helper methods

// calculateShippingFee 计算运费
func (s *ShopService) calculateShippingFee(address map[string]interface{}) float64 {
	// 简单实现：固定运费或根据地址计算
	return 0 // 免费配送
}

// calculateTaxFee 计算税费
func (s *ShopService) calculateTaxFee(amount float64) float64 {
	// 简单实现：按比例计算
	return 0 // 暂时不收税
}

// calculateCouponDiscount 计算优惠券折扣
func (s *ShopService) calculateCouponDiscount(code string, amount float64) float64 {
	// 简单实现：固定折扣
	switch code {
	case "SAVE10":
		return amount * 0.1
	case "SAVE20":
		return amount * 0.2
	default:
		return 0
	}
}

// GetShopStatistics 获取商店统计数据
func (s *ShopService) GetShopStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 商品统计
	var productCount int64
	s.db.Model(&models.Product{}).Where("status = ?", models.ProductStatusActive).Count(&productCount)
	stats["product_count"] = productCount

	// 订单统计
	var orderCount int64
	var totalRevenue float64
	s.db.Model(&models.Order{}).Where("status != ?", models.OrderStatusCancelled).Count(&orderCount)
	s.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPaid).Select("SUM(total_amount)").Row().Scan(&totalRevenue)
	stats["order_count"] = orderCount
	stats["total_revenue"] = totalRevenue

	// 用户统计
	var userCount int64
	s.db.Model(&models.User{}).Count(&userCount)
	stats["user_count"] = userCount

	// 热门商品
	var popularProducts []struct {
		ProductID uuid.UUID
		Name      string
		Sold      int
	}
	s.db.Model(&models.Product{}).
		Select("id as product_id, name, sold").
		Where("status = ?", models.ProductStatusActive).
		Order("sold desc").
		Limit(10).
		Find(&popularProducts)
	stats["popular_products"] = popularProducts

	return stats, nil
}