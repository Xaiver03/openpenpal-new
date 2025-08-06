package models

import (
	"gorm.io/gorm"
	"time"
)

// LetterStatus 信件状态枚举
type LetterStatus string

const (
	StatusDraft     LetterStatus = "draft"
	StatusGenerated LetterStatus = "generated"
	StatusCollected LetterStatus = "collected"
	StatusInTransit LetterStatus = "in_transit"
	StatusDelivered LetterStatus = "delivered"
	StatusRead      LetterStatus = "read"
)

// LetterStyle 信件样式枚举
type LetterStyle string

const (
	StyleClassic LetterStyle = "classic"
	StyleModern  LetterStyle = "modern"
	StyleVintage LetterStyle = "vintage"
	StyleElegant LetterStyle = "elegant"
	StyleCasual  LetterStyle = "casual"
)

// LetterVisibility 信件可见性枚举
type LetterVisibility string

const (
	VisibilityPrivate LetterVisibility = "private"
	VisibilityPublic  LetterVisibility = "public"
	VisibilityFriends LetterVisibility = "friends"
)

// Letter 信件模型
type Letter struct {
	ID         string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     string            `json:"user_id" gorm:"type:varchar(36);not null;index"`
	AuthorID   string            `json:"author_id" gorm:"type:varchar(36);index;default:''"`
	Title      string            `json:"title" gorm:"type:varchar(255)"`
	Content    string            `json:"content" gorm:"type:text;not null"`
	Style      LetterStyle       `json:"style" gorm:"type:varchar(20);not null;default:'classic'"`
	Status     LetterStatus      `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	Visibility LetterVisibility  `json:"visibility" gorm:"type:varchar(20);not null;default:'private'"`
	LikeCount  int               `json:"like_count" gorm:"default:0"`
	
	// OP Code System - 核心地址标识
	RecipientOPCode string          `json:"recipient_op_code" gorm:"type:varchar(6);index"` // 收件人OP Code，如: PK5F3D
	SenderOPCode    string          `json:"sender_op_code" gorm:"type:varchar(6);index"`    // 发件人OP Code（可选）
	ShareCount int               `json:"share_count" gorm:"default:0"`
	ViewCount  int               `json:"view_count" gorm:"default:0"`
	ReplyTo    string            `json:"reply_to,omitempty" gorm:"type:varchar(36);index;constraint:OnDelete:SET NULL;"`
	EnvelopeID *string           `json:"envelope_id,omitempty" gorm:"type:varchar(36);index"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `json:"-" gorm:"index"`

	// 关联
	User       *User         `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	Author     *User         `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:ID;constraint:OnDelete:CASCADE;"`
	Code       *LetterCode   `json:"code,omitempty" gorm:"foreignKey:LetterID;constraint:OnDelete:CASCADE;"`
	StatusLogs []StatusLog   `json:"status_logs,omitempty" gorm:"foreignKey:LetterID;constraint:OnDelete:CASCADE;"`
	Photos     []LetterPhoto `json:"photos,omitempty" gorm:"foreignKey:LetterID;constraint:OnDelete:CASCADE;"`
	Envelope   *Envelope     `json:"envelope,omitempty" gorm:"foreignKey:EnvelopeID;references:ID"`
	Likes      []LetterLike  `json:"likes,omitempty" gorm:"foreignKey:LetterID;constraint:OnDelete:CASCADE;"`
	Shares     []LetterShare `json:"shares,omitempty" gorm:"foreignKey:LetterID;constraint:OnDelete:CASCADE;"`
}

// BarcodeStatus 条码状态枚举 - FSD规格
type BarcodeStatus string

const (
	BarcodeStatusUnactivated BarcodeStatus = "unactivated" // 未激活
	BarcodeStatusBound       BarcodeStatus = "bound"       // 已绑定
	BarcodeStatusInTransit   BarcodeStatus = "in_transit"  // 投递中
	BarcodeStatusDelivered   BarcodeStatus = "delivered"   // 已送达
	BarcodeStatusExpired     BarcodeStatus = "expired"     // 已过期
	BarcodeStatusCancelled   BarcodeStatus = "cancelled"   // 已取消
)

// LetterCode 信件编号 - 增强支持FSD条码系统规格
type LetterCode struct {
	ID         string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID   string     `json:"letter_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	Code       string     `json:"code" gorm:"type:varchar(50);not null;uniqueIndex"` // 兼容现有12位编码
	QRCodeURL  string     `json:"qr_code_url" gorm:"type:varchar(500)"`
	QRCodePath string     `json:"qr_code_path" gorm:"type:varchar(500)"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// FSD条码系统增强字段
	Status         BarcodeStatus `json:"status" gorm:"type:varchar(20);default:'unactivated';index"` // 条码状态
	RecipientCode  string        `json:"recipient_code,omitempty" gorm:"type:varchar(6);index"`      // 收件人OP Code
	EnvelopeID     string        `json:"envelope_id,omitempty" gorm:"type:varchar(36);index"`        // 关联信封ID
	BoundAt        *time.Time    `json:"bound_at,omitempty"`                                         // 绑定时间
	DeliveredAt    *time.Time    `json:"delivered_at,omitempty"`                                     // 送达时间
	LastScannedBy  string        `json:"last_scanned_by,omitempty" gorm:"type:varchar(36)"`          // 最后扫码人
	LastScannedAt  *time.Time    `json:"last_scanned_at,omitempty"`                                  // 最后扫码时间
	ScanCount      int           `json:"scan_count" gorm:"default:0"`                                // 扫码次数

	// 关联
	Letter   Letter    `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
	Envelope *Envelope `json:"envelope,omitempty" gorm:"foreignKey:EnvelopeID;references:ID;constraint:OnDelete:SET NULL;"`
}

// FSD条码系统方法

// IsValidTransition 检查状态转换是否有效
func (lc *LetterCode) IsValidTransition(newStatus BarcodeStatus) bool {
	validTransitions := map[BarcodeStatus][]BarcodeStatus{
		BarcodeStatusUnactivated: {BarcodeStatusBound, BarcodeStatusExpired, BarcodeStatusCancelled},
		BarcodeStatusBound:       {BarcodeStatusInTransit, BarcodeStatusCancelled},
		BarcodeStatusInTransit:   {BarcodeStatusDelivered, BarcodeStatusCancelled},
		BarcodeStatusDelivered:   {}, // 终态
		BarcodeStatusExpired:     {}, // 终态
		BarcodeStatusCancelled:   {}, // 终态
	}

	allowedStates, exists := validTransitions[lc.Status]
	if !exists {
		return false
	}

	for _, allowedState := range allowedStates {
		if allowedState == newStatus {
			return true
		}
	}
	return false
}

// IsActive 检查条码是否处于活跃状态
func (lc *LetterCode) IsActive() bool {
	return lc.Status != BarcodeStatusExpired && 
		   lc.Status != BarcodeStatusCancelled && 
		   lc.Status != BarcodeStatusDelivered
}

// CanBeBound 检查条码是否可以绑定
func (lc *LetterCode) CanBeBound() bool {
	return lc.Status == BarcodeStatusUnactivated && lc.IsActive()
}

// GetStatusDisplayName 获取状态显示名称
func (lc *LetterCode) GetStatusDisplayName() string {
	statusNames := map[BarcodeStatus]string{
		BarcodeStatusUnactivated: "未激活",
		BarcodeStatusBound:       "已绑定",
		BarcodeStatusInTransit:   "投递中",
		BarcodeStatusDelivered:   "已送达",
		BarcodeStatusExpired:     "已过期",
		BarcodeStatusCancelled:   "已取消",
	}
	
	if name, exists := statusNames[lc.Status]; exists {
		return name
	}
	return "未知状态"
}

// StatusLog 状态更新日志
type StatusLog struct {
	ID        string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID  string       `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	Status    LetterStatus `json:"status" gorm:"type:varchar(20);not null"`
	UpdatedBy string       `json:"updated_by" gorm:"type:varchar(36)"`
	Location  string       `json:"location,omitempty" gorm:"type:varchar(255)"`
	Note      string       `json:"note,omitempty" gorm:"type:text"`
	CreatedAt time.Time    `json:"created_at"`

	// 关联
	Letter Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
}

// LetterPhoto 信件照片
type LetterPhoto struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID  string    `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	ImageURL  string    `json:"image_url" gorm:"type:varchar(500);not null"`
	IsPublic  bool      `json:"is_public" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Letter Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
}

// CreateLetterRequest 创建信件请求
type CreateLetterRequest struct {
	Title   string      `json:"title"`
	Content string      `json:"content" binding:"required"`
	Style   LetterStyle `json:"style" binding:"required"`
	ReplyTo string      `json:"reply_to,omitempty"`
}

// UpdateLetterStatusRequest 更新信件状态请求
type UpdateLetterStatusRequest struct {
	Status   LetterStatus `json:"status" binding:"required"`
	Location string       `json:"location,omitempty"`
	Note     string       `json:"note,omitempty"`
}

// UpdateLetterRequest 更新信件内容请求
type UpdateLetterRequest struct {
	Title   string      `json:"title"`
	Content string      `json:"content" binding:"required"`
	Style   LetterStyle `json:"style" binding:"required"`
}

// LetterListParams 信件列表查询参数
type LetterListParams struct {
	Page      int          `form:"page,default=1"`
	Limit     int          `form:"limit,default=20"`
	Status    LetterStatus `form:"status"`
	Style     LetterStyle  `form:"style"`
	Search    string       `form:"search"`
	SortBy    string       `form:"sort_by,default=created_at"`
	SortOrder string       `form:"sort_order,default=desc"`
}

// LetterResponse 信件响应
type LetterResponse struct {
	*Letter
	QRCodeURL string `json:"qr_code_url,omitempty"`
	ReadURL   string `json:"read_url,omitempty"`
}

// LetterStats 信件统计
type LetterStats struct {
	TotalSent     int64 `json:"total_sent"`
	TotalReceived int64 `json:"total_received"`
	InTransit     int64 `json:"in_transit"`
	Delivered     int64 `json:"delivered"`
	Drafts        int64 `json:"drafts"`
}

// LetterThread 信件线程模型 - 管理对话线程
type LetterThread struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	OriginalLetter string    `json:"original_letter" gorm:"type:varchar(36);not null;index"`
	Participants   string    `json:"participants" gorm:"type:text"` // JSON array of user IDs
	ThreadTitle    string    `json:"thread_title" gorm:"type:varchar(200)"`
	LastReplyAt    time.Time `json:"last_reply_at" gorm:"index"`
	ReplyCount     int       `json:"reply_count" gorm:"default:0"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LetterReply 回信模型
type LetterReply struct {
	ID            string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ThreadID      string       `json:"thread_id" gorm:"type:varchar(36);not null;index"`
	ReplyToLetter string       `json:"reply_to_letter" gorm:"type:varchar(36);not null;index"`
	AuthorID      string       `json:"author_id" gorm:"type:varchar(36);not null;index"`
	Content       string       `json:"content" gorm:"type:text;not null"`
	Style         LetterStyle  `json:"style" gorm:"type:varchar(20);not null;default:'classic'"`
	Status        LetterStatus `json:"status" gorm:"type:varchar(20);not null;default:'sent'"`
	IsPublic      bool         `json:"is_public" gorm:"default:false"`
	DeliveryCode  string       `json:"delivery_code" gorm:"type:varchar(20);uniqueIndex"`
	QRCodePath    string       `json:"qr_code_path" gorm:"type:varchar(500)"`
	ReadAt        *time.Time   `json:"read_at"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// CreateReplyRequest 创建回信请求
type CreateReplyRequest struct {
	OriginalLetterCode string      `json:"original_letter_code" binding:"required"`
	Content            string      `json:"content" binding:"required"`
	Style              LetterStyle `json:"style" binding:"required"`
	IsPublic           bool        `json:"is_public"`
}

// ThreadResponse 线程响应
type ThreadResponse struct {
	ID             string                `json:"id"`
	OriginalLetter LetterResponse        `json:"original_letter"`
	Participants   []string              `json:"participants"`
	ThreadTitle    string                `json:"thread_title"`
	LastReplyAt    time.Time             `json:"last_reply_at"`
	ReplyCount     int                   `json:"reply_count"`
	IsActive       bool                  `json:"is_active"`
	Replies        []LetterReplyResponse `json:"replies,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// LetterReplyResponse 回信响应
type LetterReplyResponse struct {
	ID            string       `json:"id"`
	ThreadID      string       `json:"thread_id"`
	ReplyToLetter string       `json:"reply_to_letter"`
	AuthorID      string       `json:"author_id"`
	Content       string       `json:"content"`
	Style         LetterStyle  `json:"style"`
	Status        LetterStatus `json:"status"`
	IsPublic      bool         `json:"is_public"`
	DeliveryCode  string       `json:"delivery_code"`
	QRCodeURL     string       `json:"qr_code_url,omitempty"`
	ReadAt        *time.Time   `json:"read_at"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// LetterLike 信件点赞模型
type LetterLike struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID  string    `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Letter *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
}

// LetterShare 信件分享模型
type LetterShare struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID  string    `json:"letter_id" gorm:"type:varchar(36);not null;index"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Platform  string    `json:"platform" gorm:"type:varchar(50)"`
	ShareURL  string    `json:"share_url" gorm:"type:varchar(500)"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Letter *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
}

// LetterTemplate 信件模板模型
type LetterTemplate struct {
	ID              string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name            string      `json:"name" gorm:"type:varchar(255);not null"`
	Description     string      `json:"description" gorm:"type:text"`
	Content         string      `json:"content" gorm:"type:text;default:''"`
	ContentTemplate string      `json:"content_template" gorm:"type:text"`
	Style           LetterStyle `json:"style" gorm:"type:varchar(20);not null;default:'classic'"`
	StyleConfig     string      `json:"style_config" gorm:"type:text"` // JSON string for style configuration
	Category        string      `json:"category" gorm:"type:varchar(100)"`
	Tags            string      `json:"tags" gorm:"type:varchar(500)"` // JSON array of tags
	PreviewImage    string      `json:"preview_image" gorm:"type:varchar(500)"`
	IsPublic        bool        `json:"is_public" gorm:"default:true"`
	IsPremium       bool        `json:"is_premium" gorm:"default:false"`
	IsActive        bool        `json:"is_active" gorm:"default:true"`
	UsageCount      int         `json:"usage_count" gorm:"default:0"`
	Rating          float64     `json:"rating" gorm:"default:0"`
	CreatedBy       string      `json:"created_by" gorm:"type:varchar(36);index"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`

	// 关联
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnDelete:SET NULL;"`
}
