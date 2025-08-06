package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductStatus 商品状态
type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"    // 草稿
	ProductStatusActive   ProductStatus = "active"   // 上架
	ProductStatusInactive ProductStatus = "inactive" // 下架
	ProductStatusDeleted  ProductStatus = "deleted"  // 已删除
)

// ProductType 商品类型
type ProductType string

const (
	ProductTypeEnvelope   ProductType = "envelope"   // 信封
	ProductTypeStationery ProductType = "stationery" // 信纸
	ProductTypeStamp      ProductType = "stamp"      // 邮票
	ProductTypePostcard   ProductType = "postcard"   // 明信片
	ProductTypeGift       ProductType = "gift"       // 礼品套装
	ProductTypeDigital    ProductType = "digital"    // 数字商品
)

// Product 商品模型
type Product struct {
	ID            uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	Name          string        `gorm:"type:varchar(200);not null" json:"name"`
	Description   string        `gorm:"type:text" json:"description"`
	Category      string        `gorm:"type:varchar(100)" json:"category"`
	ProductType   ProductType   `gorm:"type:varchar(50);not null" json:"product_type"`
	Price         float64       `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice float64       `gorm:"type:decimal(10,2)" json:"original_price"`
	Discount      int           `gorm:"type:int;default:0" json:"discount"` // 折扣百分比
	Stock         int           `gorm:"type:int;default:0" json:"stock"`
	Sold          int           `gorm:"type:int;default:0" json:"sold"`
	ImageURL      string        `gorm:"type:text" json:"image_url"`
	ThumbnailURL  string        `gorm:"type:text" json:"thumbnail_url"`
	Images        datatypes.JSON `gorm:"type:jsonb" json:"images"` // 商品图片列表
	Tags          datatypes.JSON `gorm:"type:jsonb" json:"tags"`   // 商品标签
	Specifications datatypes.JSON `gorm:"type:jsonb" json:"specifications"` // 商品规格
	Status        ProductStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	IsFeatured    bool          `gorm:"default:false" json:"is_featured"`
	Rating        float64       `gorm:"type:decimal(3,2);default:0" json:"rating"`
	ReviewCount   int           `gorm:"type:int;default:0" json:"review_count"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     *time.Time    `gorm:"index" json:"deleted_at,omitempty"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Cart 购物车模型
type Cart struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	UserID     string     `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items      []CartItem `gorm:"foreignKey:CartID" json:"items"`
	TotalItems int        `gorm:"type:int;default:0" json:"total_items"`
	TotalAmount float64   `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CartItem 购物车项目模型
type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null;index" json:"cart_id"`
	Cart      *Cart     `gorm:"foreignKey:CartID" json:"cart,omitempty"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"type:int;not null;default:1" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"` // 加入时的价格
	Subtotal  float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ci *CartItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // 待支付
	OrderStatusPaid       OrderStatus = "paid"       // 已支付
	OrderStatusProcessing OrderStatus = "processing" // 处理中
	OrderStatusShipped    OrderStatus = "shipped"    // 已发货
	OrderStatusDelivered  OrderStatus = "delivered"  // 已送达
	OrderStatusCompleted  OrderStatus = "completed"  // 已完成
	OrderStatusCancelled  OrderStatus = "cancelled"  // 已取消
	OrderStatusRefunded   OrderStatus = "refunded"   // 已退款
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"  // 待支付
	PaymentStatusPaid     PaymentStatus = "paid"     // 已支付
	PaymentStatusFailed   PaymentStatus = "failed"   // 支付失败
	PaymentStatusRefunded PaymentStatus = "refunded" // 已退款
)

// Order 订单模型
type Order struct {
	ID             uuid.UUID     `gorm:"type:uuid;primary_key" json:"id"`
	OrderNo        string        `gorm:"type:varchar(50);unique;not null" json:"order_no"`
	UserID         string        `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User           *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items          []OrderItem   `gorm:"foreignKey:OrderID" json:"items"`
	TotalItems     int           `gorm:"type:int;default:0" json:"total_items"`
	Subtotal       float64       `gorm:"type:decimal(10,2);default:0" json:"subtotal"`
	ShippingFee    float64       `gorm:"type:decimal(10,2);default:0" json:"shipping_fee"`
	TaxFee         float64       `gorm:"type:decimal(10,2);default:0" json:"tax_fee"`
	DiscountAmount float64       `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TotalAmount    float64       `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status         OrderStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaymentStatus  PaymentStatus `gorm:"type:varchar(20);default:'pending'" json:"payment_status"`
	PaymentMethod  string        `gorm:"type:varchar(50)" json:"payment_method"`
	PaymentID      string        `gorm:"type:varchar(100)" json:"payment_id"`
	ShippingAddress datatypes.JSON `gorm:"type:jsonb" json:"shipping_address"`
	TrackingNumber string        `gorm:"type:varchar(100)" json:"tracking_number"`
	Notes          string        `gorm:"type:text" json:"notes"`
	CouponCode     string        `gorm:"type:varchar(50)" json:"coupon_code"`
	PaidAt         *time.Time    `json:"paid_at,omitempty"`
	ShippedAt      *time.Time    `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time    `json:"delivered_at,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
	CancelledAt    *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.OrderNo == "" {
		o.OrderNo = generateOrderNo()
	}
	return nil
}

// OrderItem 订单项目模型
type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	Order     *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"type:int;not null;default:1" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"` // 购买时的价格
	Subtotal  float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}

// ProductReview 商品评价模型
type ProductReview struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrderID   uuid.UUID `gorm:"type:uuid;index" json:"order_id"`
	Order     *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Rating    int       `gorm:"type:int;not null" json:"rating"` // 1-5
	Comment   string    `gorm:"type:text" json:"comment"`
	Images    datatypes.JSON  `gorm:"type:jsonb" json:"images"` // 评价图片
	IsAnonymous bool    `gorm:"default:false" json:"is_anonymous"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pr *ProductReview) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

// ProductFavorite 商品收藏模型
type ProductFavorite struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (pf *ProductFavorite) BeforeCreate(tx *gorm.DB) error {
	if pf.ID == uuid.Nil {
		pf.ID = uuid.New()
	}
	return nil
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	// 格式: ORD + 年月日 + 6位随机数
	return "ORD" + time.Now().Format("20060102") + generateRandomString(6)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}