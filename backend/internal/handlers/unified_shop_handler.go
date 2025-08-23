package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UnifiedShopHandler 统一商城处理器
type UnifiedShopHandler struct {
	db               *gorm.DB
	shopService      *services.ShopService
	creditService    *services.CreditService
	creditShopService *services.CreditShopService
}

// NewUnifiedShopHandler 创建统一商城处理器
func NewUnifiedShopHandler(
	db *gorm.DB,
	shopService *services.ShopService,
	creditService *services.CreditService,
	creditShopService *services.CreditShopService,
) *UnifiedShopHandler {
	return &UnifiedShopHandler{
		db:               db,
		shopService:      shopService,
		creditService:    creditService,
		creditShopService: creditShopService,
	}
}

// UnifiedProduct 统一商品结构
type UnifiedProduct struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`           // 现金价格
	CreditPrice  int      `json:"credit_price"`    // 积分价格
	Stock        int      `json:"stock"`
	Category     string   `json:"category"`
	Type         string   `json:"type"`            // regular, credit, both
	Features     []string `json:"features"`
	ImageURL     string   `json:"image_url"`
	Rating       float64  `json:"rating"`
	SalesCount   int      `json:"sales_count"`
	IsPopular    bool     `json:"is_popular"`
	Discount     int      `json:"discount"`
}

// GetProducts 获取商品列表（支持现金和积分商品）
func (h *UnifiedShopHandler) GetProducts(c *gin.Context) {
	// 获取查询参数
	productType := c.Query("type")      // all, cash, credit
	category := c.Query("category")
	sort := c.Query("sort")             // popular, price_asc, price_desc, newest
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var products []UnifiedProduct
	query := h.db.Model(&models.CreditShopProduct{})

	// 基础查询
	baseQuery := query

	// 类型过滤
	switch productType {
	case "cash":
		baseQuery = baseQuery.Where("type IN ?", []string{"regular", "both"})
	case "credit":
		baseQuery = baseQuery.Where("type IN ?", []string{"credit", "both"})
	}

	// 分类过滤
	if category != "" && category != "all" {
		baseQuery = baseQuery.Where("category = ?", category)
	}

	// 搜索
	if search != "" {
		baseQuery = baseQuery.Where("name LIKE ? OR description LIKE ?", 
			"%"+search+"%", "%"+search+"%")
	}

	// 只显示有库存的商品
	baseQuery = baseQuery.Where("stock > 0").Where("is_active = ?", true)

	// 排序
	switch sort {
	case "price_asc":
		baseQuery = baseQuery.Order("price ASC")
	case "price_desc":
		baseQuery = baseQuery.Order("price DESC")
	case "newest":
		baseQuery = baseQuery.Order("created_at DESC")
	default: // popular
		baseQuery = baseQuery.Order("is_popular DESC, sales_count DESC")
	}

	// 分页
	offset := (page - 1) * limit
	
	// 获取现金商品
	var cashProducts []models.Product
	if productType != "credit" {
		h.db.Table("products").
			Where("deleted_at IS NULL").
			Scopes(func(db *gorm.DB) *gorm.DB { return baseQuery }).
			Limit(limit).
			Offset(offset).
			Find(&cashProducts)
	}

	// 获取积分商品
	var creditProducts []models.CreditShopProduct
	if productType != "cash" {
		h.db.Table("credit_shop_products").
			Where("deleted_at IS NULL").
			Scopes(func(db *gorm.DB) *gorm.DB { return baseQuery }).
			Limit(limit).
			Offset(offset).
			Find(&creditProducts)
	}

	// 合并结果
	for _, p := range cashProducts {
		product := UnifiedProduct{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			Category:    p.Category,
			Type:        "regular",
			Features:    []string{}, // Product没有Features字段
			ImageURL:    p.ImageURL,
			Rating:      p.Rating,
			SalesCount:  p.Sold, // 使用Sold字段代替SalesCount
			IsPopular:   p.IsFeatured, // 使用IsFeatured代替IsPopular
			Discount:    p.Discount,
		}
		products = append(products, product)
	}

	for _, p := range creditProducts {
		// 检查是否已存在（支持双重支付的商品）
		exists := false
		for i, existing := range products {
			if existing.Name == p.Name && existing.Category == p.Category {
				products[i].Type = "both"
				products[i].CreditPrice = p.CreditPrice
				exists = true
				break
			}
		}
		
		if !exists {
			product := UnifiedProduct{
				ID:          p.ID.String(),
				Name:        p.Name,
				Description: p.Description,
				CreditPrice: p.CreditPrice,
				Stock:       p.Stock,
				Category:    p.Category,
				Type:        "credit",
				Features:    []string{}, // CreditShopProduct没有Features字段
				ImageURL:    p.ImageURL,
				IsPopular:   p.IsFeatured, // 使用IsFeatured代替IsPopular
			}
			products = append(products, product)
		}
	}

	// 获取总数
	var total int64
	baseQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    products,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCart 获取购物车
func (h *UnifiedShopHandler) GetCart(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 获取购物车项目
	var cartItems []models.CartItem
	h.db.Where("user_id = ?", user.ID).
		Preload("Product").
		Find(&cartItems)

	// 转换为统一格式
	var items []gin.H
	for _, item := range cartItems {
		unifiedItem := gin.H{
			"id":           item.ID,
			"product_id":   item.ProductID,
			"quantity":     item.Quantity,
			"payment_type": item.PaymentType,
			"product": gin.H{
				"id":           item.Product.ID,
				"name":         item.Product.Name,
				"description":  item.Product.Description,
				"price":        0.0, // CreditShopProduct没有Price字段
				"credit_price": item.Product.CreditPrice,
				"stock":        item.Product.Stock,
				"image_url":    item.Product.ImageURL,
			},
		}
		items = append(items, unifiedItem)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"items": items,
		},
	})
}

// AddToCart 添加到购物车
func (h *UnifiedShopHandler) AddToCart(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		ProductID   string `json:"product_id" binding:"required"`
		Quantity    int    `json:"quantity" binding:"required,min=1"`
		PaymentType string `json:"payment_type" binding:"required,oneof=cash credit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 检查商品是否存在
	var productStock int
	
	if req.PaymentType == "cash" {
		var product models.Product
		if err := h.db.First(&product, "id = ?", req.ProductID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"code":    4004,
				"message": "商品不存在",
			})
			return
		}
		productStock = product.Stock
	} else {
		var creditProduct models.CreditShopProduct
		if err := h.db.First(&creditProduct, "id = ?", req.ProductID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"code":    4004,
				"message": "积分商品不存在",
			})
			return
		}
		productStock = creditProduct.Stock
	}

	// 检查库存
	if productStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4002,
			"message": "库存不足",
		})
		return
	}

	// 检查是否已在购物车中
	var existingItem models.CartItem
	err := h.db.Where("user_id = ? AND product_id = ? AND payment_type = ?", 
		user.ID, req.ProductID, req.PaymentType).First(&existingItem).Error
	
	if err == nil {
		// 更新数量
		existingItem.Quantity += req.Quantity
		h.db.Save(&existingItem)
	} else {
		// 创建新项目
		cartItem := models.CartItem{
			UserID:      user.ID,
			ProductID:   req.ProductID,
			Quantity:    req.Quantity,
			PaymentType: req.PaymentType,
		}
		h.db.Create(&cartItem)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "添加成功",
	})
}

// UpdateCartItem 更新购物车项目
func (h *UnifiedShopHandler) UpdateCartItem(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	itemID := c.Param("id")
	
	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 更新购物车项目
	result := h.db.Model(&models.CartItem{}).
		Where("id = ? AND user_id = ?", itemID, user.ID).
		Update("quantity", req.Quantity)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "购物车项目不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "更新成功",
	})
}

// RemoveCartItem 删除购物车项目
func (h *UnifiedShopHandler) RemoveCartItem(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	itemID := c.Param("id")

	result := h.db.Where("id = ? AND user_id = ?", itemID, user.ID).
		Delete(&models.CartItem{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "购物车项目不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "删除成功",
	})
}

// ClearCart 清空购物车
func (h *UnifiedShopHandler) ClearCart(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	h.db.Where("user_id = ?", user.ID).Delete(&models.CartItem{})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "购物车已清空",
	})
}