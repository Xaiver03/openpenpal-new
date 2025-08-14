package models

import (
	"gorm.io/gorm"
	"time"
)

// Envelope status constants
const (
	EnvelopeStatusUnsent    = "unsent"
	EnvelopeStatusUsed      = "used"
	EnvelopeStatusCancelled = "cancelled"
)

// Design status constants
const (
	DesignStatusPending  = "pending"
	DesignStatusApproved = "approved"
	DesignStatusRejected = "rejected"
)

// EnvelopeDesign 信封设计 - 增强OP Code支持
type EnvelopeDesign struct {
	ID           string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SchoolCode   string `json:"school_code" gorm:"type:varchar(20)"`           // 支持OP Code前2位学校代码
	Type         string `json:"type" gorm:"type:varchar(20);default:'school'"` // city, school
	Theme        string `json:"theme" gorm:"type:varchar(100)"`
	ImageURL     string `json:"image_url" gorm:"type:varchar(500)"`
	ThumbnailURL string `json:"thumbnail_url" gorm:"type:varchar(500)"`
	CreatorID    string `json:"creator_id" gorm:"type:varchar(36);not null;index"`
	CreatorName  string `json:"creator_name" gorm:"type:varchar(100)"`
	Description  string `json:"description" gorm:"type:text"`
	Status       string `json:"status" gorm:"type:varchar(20);default:'pending'"`
	VoteCount    int    `json:"vote_count" gorm:"default:0"`
	Period       string `json:"period" gorm:"type:varchar(50)"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`

	// OP Code系统增强字段
	SupportedOPCodePrefix string  `json:"supported_op_code_prefix,omitempty" gorm:"type:varchar(4);index"` // 支持的OP Code前缀(如:PK5F)
	Price                 float64 `json:"price" gorm:"type:decimal(10,2);default:3.00"`                    // 信封价格

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (EnvelopeDesign) TableName() string {
	return "envelope_designs"
}

// Envelope 信封实例 - 增强OP Code和条码集成
type Envelope struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DesignID  string         `json:"design_id" gorm:"type:varchar(36);not null"`
	Design    EnvelopeDesign `json:"design" gorm:"foreignKey:DesignID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36)"`                 // 购买者ID
	UsedBy    string         `json:"used_by" gorm:"type:varchar(36)"`                 // 使用者ID
	LetterID  string         `json:"letter_id" gorm:"type:varchar(36);index"`         // 关联信件ID
	BarcodeID string         `json:"barcode_id" gorm:"type:varchar(100);unique"`      // 条码ID(关联LetterCode)
	Status    string         `json:"status" gorm:"type:varchar(20);default:'unsent'"` // unsent/used/cancelled
	UsedAt    *time.Time     `json:"used_at"`

	// OP Code集成字段
	RecipientOPCode string     `json:"recipient_op_code,omitempty" gorm:"type:varchar(6);index"` // 收件人OP Code
	SenderOPCode    string     `json:"sender_op_code,omitempty" gorm:"type:varchar(6);index"`    // 发件人OP Code
	DeliveredAt     *time.Time `json:"delivered_at,omitempty"`                                   // 投递完成时间
	TrackingInfo    string     `json:"tracking_info,omitempty" gorm:"type:json"`                 // 追踪信息JSON

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Envelope) TableName() string {
	return "envelopes"
}

// EnvelopeVote 信封投票
type EnvelopeVote struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	DesignID  string    `json:"design_id" gorm:"type:varchar(36);not null"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null"`
	CreatedAt time.Time `json:"created_at"`
}

func (EnvelopeVote) TableName() string {
	return "envelope_votes"
}

// EnvelopeOrder 信封订单
type EnvelopeOrder struct {
	ID             string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID         string         `json:"user_id" gorm:"type:varchar(36);not null"`
	DesignID       string         `json:"design_id" gorm:"type:varchar(36);not null"`
	Design         EnvelopeDesign `json:"design" gorm:"foreignKey:DesignID"`
	Quantity       int            `json:"quantity" gorm:"not null"`
	TotalPrice     float64        `json:"total_price" gorm:"not null"`
	Status         string         `json:"status" gorm:"type:varchar(20);default:'pending'"`
	PaymentMethod  string         `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentID      string         `json:"payment_id" gorm:"type:varchar(100)"`
	DeliveryMethod string         `json:"delivery_method" gorm:"type:varchar(50)"`
	DeliveryInfo   string         `json:"delivery_info" gorm:"type:json"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (EnvelopeOrder) TableName() string {
	return "envelope_orders"
}

// Request DTOs

// CreateEnvelopeDesignRequest 创建信封设计请求
type CreateEnvelopeDesignRequest struct {
	SchoolCode  string `json:"school_code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=city school"`
	Theme       string `json:"theme" binding:"required"`
	ImageURL    string `json:"image_url" binding:"required"`
	Description string `json:"description"`
	Period      string `json:"period"`
}

// CreateEnvelopeOrderRequest 创建信封订单请求
type CreateEnvelopeOrderRequest struct {
	DesignID       string `json:"design_id" binding:"required"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	DeliveryMethod string `json:"delivery_method" binding:"required"`
	DeliveryInfo   string `json:"delivery_info"`
}

// EnvelopeDesignListRequest 信封设计列表请求
type EnvelopeDesignListRequest struct {
	SchoolCode string `form:"school_code"`
	Type       string `form:"type"`
	Status     string `form:"status"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

// FSD条码系统请求模型

// BindBarcodeRequest 绑定条码请求 - FSD 6.2
type BindBarcodeRequest struct {
	RecipientCode string `json:"recipient_code" binding:"required,len=6"` // OP Code编码
	LetterID      string `json:"letter_id" binding:"required"`            // 信件编号
	EnvelopeID    string `json:"envelope_id,omitempty"`                   // 可选，信封编号
}

// UpdateBarcodeStatusRequest 更新物流状态请求 - FSD 6.3
type UpdateBarcodeStatusRequest struct {
	Status     string `json:"status" binding:"required,oneof=picked in_transit delivered failed"` // 状态
	OperatorID string `json:"operator_id" binding:"required"`                                     // 操作信使ID
	Location   string `json:"location,omitempty"`                                                 // 位置信息
	OPCode     string `json:"op_code,omitempty"`                                                  // 位置OP Code
	Notes      string `json:"notes,omitempty"`                                                    // 备注
}

// BatchGenerateEnvelopeRequest 批量生成信封请求 - FSD接口
type BatchGenerateEnvelopeRequest struct {
	SchoolCode    string `json:"school_code" binding:"required,len=2"`      // 学校代码
	Quantity      int    `json:"quantity" binding:"required,min=1,max=500"` // 数量
	BarcodePrefix string `json:"barcode_prefix,omitempty"`                  // 条码前缀
	DistributorID string `json:"distributor_id" binding:"required"`         // 分发者ID
	DesignID      string `json:"design_id" binding:"required"`              // 信封设计ID
}

// BarcodeResponse 条码响应 - FSD规格
type BarcodeResponse struct {
	BarcodeID string `json:"barcode_id"`
	PDFURL    string `json:"pdf_url"`
	Status    string `json:"status"`
}

// EnvelopeWithBarcodeResponse 信封+条码响应
type EnvelopeWithBarcodeResponse struct {
	EnvelopeID      string `json:"envelope_id"`
	DesignID        string `json:"design_id"`
	BarcodeID       string `json:"barcode_id"`
	BarcodeCode     string `json:"barcode_code"`
	RecipientOPCode string `json:"recipient_op_code,omitempty"`
	Status          string `json:"status"`
	QRURL           string `json:"qr_url,omitempty"`
	PDFURL          string `json:"pdf_url,omitempty"`
}
