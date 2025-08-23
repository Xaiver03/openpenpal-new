package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CartItem 购物车项目
type CartItem struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	ProductID   string    `gorm:"type:varchar(36);not null;index" json:"product_id"`
	Quantity    int       `gorm:"not null;default:1" json:"quantity"`
	PaymentType string    `gorm:"type:varchar(20);not null" json:"payment_type"` // cash, credit
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	User    User               `gorm:"foreignKey:UserID" json:"-"`
	Product CreditShopProduct  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName 指定表名
func (CartItem) TableName() string {
	return "cart_items"
}

// BeforeCreate 创建前生成ID
func (c *CartItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}