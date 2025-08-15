package models

import (
	"gorm.io/gorm"
	"time"
)

// CommentStatus 评论状态
type CommentStatus string

const (
	CommentStatusActive  CommentStatus = "active"  // 正常显示
	CommentStatusPending CommentStatus = "pending" // 待审核
	CommentStatusHidden  CommentStatus = "hidden"  // 已隐藏
	CommentStatusDeleted CommentStatus = "deleted" // 已删除
)

// Comment 评论模型 - 支持信件评论和嵌套回复
type Comment struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID   string         `json:"letter_id" gorm:"type:varchar(36);not null;index"` // 关联信件ID
	UserID     string         `json:"user_id" gorm:"type:varchar(36);not null;index"`   // 评论用户ID
	ParentID   *string        `json:"parent_id" gorm:"type:varchar(36);index"`          // 父评论ID，支持嵌套回复
	Content    string         `json:"content" gorm:"type:text;not null"`                // 评论内容
	Status     CommentStatus  `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	LikeCount  int            `json:"like_count" gorm:"default:0"`  // 点赞数
	ReplyCount int            `json:"reply_count" gorm:"default:0"` // 回复数
	IsTop      bool           `json:"is_top" gorm:"default:false"`  // 是否置顶
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Letter  *Letter       `json:"letter,omitempty" gorm:"foreignKey:LetterID;references:ID;constraint:OnDelete:CASCADE;"`
	User    *User         `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
	Parent  *Comment      `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE;"`
	Replies []Comment     `json:"replies,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Likes   []CommentLike `json:"likes,omitempty" gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE;"`
}

// CommentLike 评论点赞模型
type CommentLike struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CommentID string         `json:"comment_id" gorm:"type:varchar(36);not null;index"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Comment *Comment `json:"comment,omitempty" gorm:"foreignKey:CommentID;references:ID;constraint:OnDelete:CASCADE;"`
	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;"`
}

// 添加复合唯一索引，防止重复点赞
func (CommentLike) TableName() string {
	return "comment_likes"
}

// CommentCreateRequest 创建评论请求
type CommentCreateRequest struct {
	LetterID string  `json:"letter_id" binding:"required" validate:"uuid"`
	ParentID *string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Content  string  `json:"content" binding:"required,max=1000" validate:"min=1,max=1000"`
}

// CommentUpdateRequest 更新评论请求
type CommentUpdateRequest struct {
	Content string         `json:"content" binding:"required,max=1000" validate:"min=1,max=1000"`
	Status  *CommentStatus `json:"status,omitempty"`
}

// CommentResponse 评论响应 - 包含用户信息和统计
type CommentResponse struct {
	ID         string        `json:"id"`
	LetterID   string        `json:"letter_id"`
	UserID     string        `json:"user_id"`
	ParentID   *string       `json:"parent_id"`
	Content    string        `json:"content"`
	Status     CommentStatus `json:"status"`
	LikeCount  int           `json:"like_count"`
	ReplyCount int           `json:"reply_count"`
	IsTop      bool          `json:"is_top"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`

	// 关联数据
	User    *UserBasicInfo    `json:"user,omitempty"`
	IsLiked bool              `json:"is_liked"`          // 当前用户是否点赞
	Replies []CommentResponse `json:"replies,omitempty"` // 回复列表
}

// UserBasicInfo 评论中显示的用户基础信息
type UserBasicInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// CommentListQuery 评论列表查询参数
type CommentListQuery struct {
	LetterID     string `form:"letter_id" binding:"required"`
	Page         int    `form:"page,default=1" binding:"min=1"`
	Limit        int    `form:"limit,default=20" binding:"min=1,max=100"`
	SortBy       string `form:"sort_by,default=created_at" binding:"oneof=created_at like_count"`
	Order        string `form:"order,default=desc" binding:"oneof=asc desc"`
	ParentID     string `form:"parent_id,omitempty"` // 获取特定评论的回复
	OnlyTopLevel bool   `form:"only_top_level"`      // 仅获取顶级评论
}

// IsReply 判断是否为回复评论
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// CanEdit 判断用户是否可以编辑评论
func (c *Comment) CanEdit(userID string) bool {
	return c.UserID == userID && c.Status != CommentStatusDeleted
}

// CanDelete 判断用户是否可以删除评论
func (c *Comment) CanDelete(userID string, userRole UserRole) bool {
	// 评论作者可以删除
	if c.UserID == userID {
		return true
	}

	// 管理员可以删除任何评论
	if userRole == RolePlatformAdmin || userRole == RoleSuperAdmin {
		return true
	}

	return false
}

// GetDepth 获取评论嵌套深度
func (c *Comment) GetDepth() int {
	if c.ParentID == nil {
		return 0
	}
	// 这里可以实现递归查询深度，但为了性能考虑，限制最大深度为3
	return 1
}

// CommentReportStatus 举报状态
type CommentReportStatus string

const (
	CommentReportStatusPending  CommentReportStatus = "pending"  // 待处理
	CommentReportStatusHandled  CommentReportStatus = "handled"  // 已处理
	CommentReportStatusRejected CommentReportStatus = "rejected" // 已拒绝
)

// CommentReportReason 举报原因
type CommentReportReason string

const (
	CommentReportReasonSpam       CommentReportReason = "spam"       // 垃圾信息
	CommentReportReasonAbusive    CommentReportReason = "abusive"    // 恶意辱骂
	CommentReportReasonHateful    CommentReportReason = "hateful"    // 仇恨言论
	CommentReportReasonInappropriate CommentReportReason = "inappropriate" // 不当内容
	CommentReportReasonOther      CommentReportReason = "other"      // 其他
)

// CommentReport 评论举报模型
type CommentReport struct {
	ID          string              `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CommentID   string              `json:"comment_id" gorm:"type:varchar(36);not null;index"`
	ReporterID  string              `json:"reporter_id" gorm:"type:varchar(36);not null;index"`
	Reason      CommentReportReason `json:"reason" gorm:"type:varchar(50);not null"`
	Description string              `json:"description" gorm:"type:text"`
	Status      CommentReportStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	HandlerID   *string             `json:"handler_id" gorm:"type:varchar(36);index"`
	HandledAt   *time.Time          `json:"handled_at"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `json:"-" gorm:"index"`

	// 关联关系
	Comment  *Comment `json:"comment,omitempty" gorm:"foreignKey:CommentID;references:ID;constraint:OnDelete:CASCADE;"`
	Reporter *User    `json:"reporter,omitempty" gorm:"foreignKey:ReporterID;references:ID;constraint:OnDelete:CASCADE;"`
	Handler  *User    `json:"handler,omitempty" gorm:"foreignKey:HandlerID;references:ID;constraint:OnDelete:SET NULL;"`
}

// CommentReportRequest 举报评论请求
type CommentReportRequest struct {
	Reason      CommentReportReason `json:"reason" binding:"required" validate:"oneof=spam abusive hateful inappropriate other"`
	Description string              `json:"description" binding:"max=500" validate:"max=500"`
}

// CommentType 评论目标类型
type CommentType string

const (
	CommentTypeLetter  CommentType = "letter"  // 信件评论
	CommentTypeProfile CommentType = "profile" // 个人资料评论
	CommentTypeMuseum  CommentType = "museum"  // 博物馆展品评论
)

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments     []CommentResponse `json:"comments"`
	TotalCount   int               `json:"total_count"`
	TotalPages   int               `json:"total_pages"`
	CurrentPage  int               `json:"current_page"`
	HasNext      bool              `json:"has_next"`
	HasPrevious  bool              `json:"has_previous"`
}

// CommentStats 评论统计
type CommentStats struct {
	TotalCount     int `json:"total_count"`
	LikedCount     int `json:"liked_count"`
	RepliedCount   int `json:"replied_count"`
	TopComments    int `json:"top_comments"`
	RecentComments int `json:"recent_comments"`
}
