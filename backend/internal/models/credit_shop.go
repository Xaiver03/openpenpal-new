package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreditShopProductStatus 积分商城商品状态
type CreditShopProductStatus string

const (
	CreditProductStatusDraft     CreditShopProductStatus = "draft"     // 草稿
	CreditProductStatusActive    CreditShopProductStatus = "active"    // 上架
	CreditProductStatusInactive  CreditShopProductStatus = "inactive"  // 下架
	CreditProductStatusSoldOut   CreditShopProductStatus = "sold_out"  // 已售罄
	CreditProductStatusDeleted   CreditShopProductStatus = "deleted"   // 已删除
)

// CreditShopProductType 积分商城商品类型
type CreditShopProductType string

const (
	CreditProductTypePhysical CreditShopProductType = "physical" // 实物商品
	CreditProductTypeVirtual  CreditShopProductType = "virtual"  // 虚拟商品
	CreditProductTypeService  CreditShopProductType = "service"  // 服务类商品
	CreditProductTypeVoucher  CreditShopProductType = "voucher"  // 优惠券
)

// CreditShopProduct 积分商城商品模型
type CreditShopProduct struct {
	ID             uuid.UUID               `gorm:"type:varchar(36);primary_key" json:"id"`
	Name           string                  `gorm:"type:varchar(200);not null" json:"name"`
	Description    string                  `gorm:"type:text" json:"description"`
	ShortDesc      string                  `gorm:"type:varchar(500)" json:"short_desc"`      // 简短描述
	Category       string                  `gorm:"type:varchar(100)" json:"category"`
	ProductType    CreditShopProductType   `gorm:"type:varchar(50);not null" json:"product_type"`
	CreditPrice    int                     `gorm:"not null" json:"credit_price"`             // 积分价格
	OriginalPrice  float64                 `gorm:"type:decimal(10,2)" json:"original_price"` // 原价（参考）
	Stock          int                     `gorm:"type:int;default:0" json:"stock"`          // 库存
	TotalStock     int                     `gorm:"type:int;default:0" json:"total_stock"`    // 总库存
	RedeemCount    int                     `gorm:"type:int;default:0" json:"redeem_count"`   // 兑换次数
	ImageURL       string                  `gorm:"type:text" json:"image_url"`
	Images         datatypes.JSON          `gorm:"type:jsonb" json:"images"`         // 商品图片列表
	Tags           datatypes.JSON          `gorm:"type:jsonb" json:"tags"`           // 商品标签
	Specifications datatypes.JSON          `gorm:"type:jsonb" json:"specifications"` // 商品规格
	Status         CreditShopProductStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	IsFeatured     bool                    `gorm:"default:false" json:"is_featured"`        // 是否推荐
	IsLimited      bool                    `gorm:"default:false" json:"is_limited"`         // 是否限量
	LimitPerUser   int                     `gorm:"type:int;default:0" json:"limit_per_user"` // 每用户限购数量
	Priority       int                     `gorm:"type:int;default:0" json:"priority"`      // 排序优先级
	ValidFrom      *time.Time              `json:"valid_from,omitempty"`                    // 有效期开始
	ValidTo        *time.Time              `json:"valid_to,omitempty"`                      // 有效期结束
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      *time.Time              `gorm:"index" json:"deleted_at,omitempty"`
}

func (p *CreditShopProduct) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsAvailable 检查商品是否可兑换
func (p *CreditShopProduct) IsAvailable() bool {
	now := time.Now()
	
	// 检查状态
	if p.Status != CreditProductStatusActive {
		return false
	}
	
	// 检查库存
	if p.Stock <= 0 {
		return false
	}
	
	// 检查有效期
	if p.ValidFrom != nil && now.Before(*p.ValidFrom) {
		return false
	}
	if p.ValidTo != nil && now.After(*p.ValidTo) {
		return false
	}
	
	return true
}

// CreditRedemptionStatus 积分兑换订单状态
type CreditRedemptionStatus string

const (
	RedemptionStatusPending     CreditRedemptionStatus = "pending"     // 待处理
	RedemptionStatusConfirmed   CreditRedemptionStatus = "confirmed"   // 已确认
	RedemptionStatusProcessing  CreditRedemptionStatus = "processing"  // 处理中
	RedemptionStatusShipped     CreditRedemptionStatus = "shipped"     // 已发货
	RedemptionStatusDelivered   CreditRedemptionStatus = "delivered"   // 已送达
	RedemptionStatusCompleted   CreditRedemptionStatus = "completed"   // 已完成
	RedemptionStatusCancelled   CreditRedemptionStatus = "cancelled"   // 已取消
	RedemptionStatusRefunded    CreditRedemptionStatus = "refunded"    // 已退款
)

// CreditRedemption 积分兑换订单模型
type CreditRedemption struct {
	ID              uuid.UUID              `gorm:"type:varchar(36);primary_key" json:"id"`
	RedemptionNo    string                 `gorm:"type:varchar(50);unique;not null" json:"redemption_no"`
	UserID          string                 `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User            *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProductID       uuid.UUID              `gorm:"type:varchar(36);not null;index" json:"product_id"`
	Product         *CreditShopProduct     `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity        int                    `gorm:"type:int;not null;default:1" json:"quantity"`
	CreditPrice     int                    `gorm:"not null" json:"credit_price"`         // 兑换时的积分价格
	TotalCredits    int                    `gorm:"not null" json:"total_credits"`        // 总积分消耗
	Status          CreditRedemptionStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	DeliveryInfo    datatypes.JSON         `gorm:"type:jsonb" json:"delivery_info"`      // 配送信息
	RedemptionCode  string                 `gorm:"type:varchar(100)" json:"redemption_code"` // 兑换码（虚拟商品）
	TrackingNumber  string                 `gorm:"type:varchar(100)" json:"tracking_number"` // 物流单号
	Notes           string                 `gorm:"type:text" json:"notes"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
	ShippedAt       *time.Time             `json:"shipped_at,omitempty"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func (r *CreditRedemption) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.RedemptionNo == "" {
		r.RedemptionNo = generateRedemptionNo()
	}
	return nil
}

// CreditCart 积分购物车模型
type CreditCart struct {
	ID           uuid.UUID        `gorm:"type:varchar(36);primary_key" json:"id"`
	UserID       string           `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User         *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items        []CreditCartItem `gorm:"foreignKey:CartID" json:"items"`
	TotalItems   int              `gorm:"type:int;default:0" json:"total_items"`
	TotalCredits int              `gorm:"type:int;default:0" json:"total_credits"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

func (c *CreditCart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CreditCartItem 积分购物车项目模型
type CreditCartItem struct {
	ID          uuid.UUID          `gorm:"type:varchar(36);primary_key" json:"id"`
	CartID      uuid.UUID          `gorm:"type:varchar(36);not null;index" json:"cart_id"`
	Cart        *CreditCart        `gorm:"foreignKey:CartID" json:"cart,omitempty"`
	ProductID   uuid.UUID          `gorm:"type:varchar(36);not null;index" json:"product_id"`
	Product     *CreditShopProduct `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity    int                `gorm:"type:int;not null;default:1" json:"quantity"`
	CreditPrice int                `gorm:"not null" json:"credit_price"` // 加入时的积分价格
	Subtotal    int                `gorm:"not null" json:"subtotal"`     // 小计积分
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (ci *CreditCartItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}

// CreditShopCategory 积分商城分类模型
type CreditShopCategory struct {
	ID          uuid.UUID `gorm:"type:varchar(36);primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IconURL     string    `gorm:"type:text" json:"icon_url"`
	ParentID    *uuid.UUID `gorm:"type:varchar(36);index" json:"parent_id,omitempty"`
	Parent      *CreditShopCategory `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []CreditShopCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	SortOrder   int       `gorm:"type:int;default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *CreditShopCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// UserRedemptionHistory 用户兑换历史统计
type UserRedemptionHistory struct {
	ID               uuid.UUID `gorm:"type:varchar(36);primary_key" json:"id"`
	UserID           string    `gorm:"type:varchar(36);not null;uniqueIndex" json:"user_id"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TotalRedemptions int       `gorm:"type:int;default:0" json:"total_redemptions"`     // 总兑换次数
	TotalCreditsUsed int       `gorm:"type:int;default:0" json:"total_credits_used"`   // 总消耗积分
	LastRedemptionAt *time.Time `json:"last_redemption_at,omitempty"`                 // 最后兑换时间
	FavoriteCategory string    `gorm:"type:varchar(100)" json:"favorite_category"`     // 最喜欢的分类
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (u *UserRedemptionHistory) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// CreditShopConfig 积分商城配置模型
type CreditShopConfig struct {
	ID                   uuid.UUID `gorm:"type:varchar(36);primary_key" json:"id"`
	Key                  string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Value                string    `gorm:"type:text;not null" json:"value"`
	Description          string    `gorm:"type:text" json:"description"`
	Category             string    `gorm:"type:varchar(50)" json:"category"`
	IsEditable           bool      `gorm:"default:true" json:"is_editable"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (c *CreditShopConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// generateRedemptionNo 生成兑换订单号
func generateRedemptionNo() string {
	// 格式: CRD + 年月日 + 8位随机数
	return "CRD" + time.Now().Format("20060102") + generateRandomString(8)
}

