package models

import (
	"strings"
	"time"
	
	"gorm.io/gorm"
)

// CommentStatus 评论状态
type CommentStatus string

const (
	CommentStatusActive   CommentStatus = "active"   // 正常显示
	CommentStatusPending  CommentStatus = "pending"  // 待审核
	CommentStatusRejected CommentStatus = "rejected" // 已拒绝
	CommentStatusHidden   CommentStatus = "hidden"   // 已隐藏
	CommentStatusDeleted  CommentStatus = "deleted"  // 已删除
)

// CommentType 评论目标类型
type CommentType string

const (
	CommentTypeLetter    CommentType = "letter"    // 信件评论
	CommentTypeMuseum    CommentType = "museum"    // 博物馆项目评论
	CommentTypeEnvelope  CommentType = "envelope"  // 信封评论
	CommentTypeShop      CommentType = "shop"      // 商店商品评论
	CommentTypeProfile   CommentType = "profile"   // 个人资料评论
)

// Comment 评论模型 - 支持多目标评论和嵌套回复
type Comment struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	// 多目标支持
	TargetID   string         `json:"target_id" gorm:"type:varchar(36);not null;index"`   // 目标对象ID（通用）
	TargetType CommentType    `json:"target_type" gorm:"type:varchar(20);not null;index"` // 目标类型
	// 向后兼容字段
	LetterID   string         `json:"letter_id" gorm:"type:varchar(36);index"` // 兼容信件评论
	UserID     string         `json:"user_id" gorm:"type:varchar(36);not null;index"`   // 评论用户ID
	ParentID   *string        `json:"parent_id" gorm:"type:varchar(36);index"`          // 父评论ID，支持嵌套回复
	RootID     *string        `json:"root_id" gorm:"type:varchar(36);index"`            // 根评论ID，用于多层嵌套
	Content    string         `json:"content" gorm:"type:text;not null"`                // 评论内容
	Status     CommentStatus  `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Level      int            `json:"level" gorm:"default:0"`       // 嵌套层级
	Path       string         `json:"path" gorm:"type:varchar(500)"` // 层级路径，如：root_id/parent_id/id
	LikeCount  int            `json:"like_count" gorm:"default:0"`  // 点赞数
	DislikeCount int          `json:"dislike_count" gorm:"default:0"` // 踩数
	ReplyCount int            `json:"reply_count" gorm:"default:0"` // 回复数
	ReportCount int           `json:"report_count" gorm:"default:0"` // 举报数
	IsTop      bool           `json:"is_top" gorm:"default:false"`  // 是否置顶
	IsAnonymous bool          `json:"is_anonymous" gorm:"default:false"` // 是否匿名
	IsEdited   bool           `json:"is_edited" gorm:"default:false"`    // 是否已编辑
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
	IsLike    bool           `json:"is_like" gorm:"default:true"`  // true=点赞, false=踩
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

// CommentCreateRequest 创建评论请求 - SOTA多目标支持
type CommentCreateRequest struct {
	// 新版多目标字段
	TargetID   string      `json:"target_id" binding:"required" validate:"uuid"`
	TargetType CommentType `json:"target_type" binding:"required" validate:"oneof=letter museum envelope shop"`
	// 向后兼容字段
	LetterID string  `json:"letter_id,omitempty" validate:"omitempty,uuid"`
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
	ID           string        `json:"id"`
	TargetID     string        `json:"target_id"`
	TargetType   CommentType   `json:"target_type"`
	LetterID     string        `json:"letter_id"`     // 兼容性字段
	UserID       string        `json:"user_id"`
	ParentID     *string       `json:"parent_id"`
	RootID       *string       `json:"root_id"`
	Content      string        `json:"content"`
	Status       CommentStatus `json:"status"`
	Level        int           `json:"level"`
	Path         string        `json:"path"`
	LikeCount    int           `json:"like_count"`
	DislikeCount int           `json:"dislike_count"`
	ReplyCount   int           `json:"reply_count"`
	ReportCount  int           `json:"report_count"`
	NetLikes     int           `json:"net_likes"`
	IsTop        bool          `json:"is_top"`
	IsAnonymous  bool          `json:"is_anonymous"`
	IsEdited     bool          `json:"is_edited"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	// 关联数据
	User    *UserBasicInfo    `json:"user,omitempty"`
	IsLiked bool              `json:"is_liked"`          // 当前用户是否点赞
	Replies []CommentResponse `json:"replies,omitempty"` // 回复列表
	
	// 权限字段
	CanEdit   bool `json:"can_edit"`
	CanDelete bool `json:"can_delete"`
	CanReport bool `json:"can_report"`
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
	LetterID        string `form:"letter_id" binding:"required"`
	Page            int    `form:"page,default=1" binding:"min=1"`
	Limit           int    `form:"limit,default=20" binding:"min=1,max=100"`
	SortBy          string `form:"sort_by,default=created_at" binding:"oneof=created_at like_count"`
	Order           string `form:"order,default=desc" binding:"oneof=asc desc"`
	ParentID        string `form:"parent_id,omitempty"`    // 获取特定评论的回复
	RootID          string `form:"root_id,omitempty"`      // 获取根评论的所有回复
	OnlyTopLevel    bool   `form:"only_top_level"`         // 仅获取顶级评论
	MaxLevel        int    `form:"max_level,omitempty"`    // 最大嵌套层级
	AuthorID        string `form:"author_id,omitempty"`    // 作者过滤
	IncludeReplies  bool   `form:"include_replies"`        // 包含回复
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

// IsSpamLikely 检测评论是否可能是垃圾信息
func (c *Comment) IsSpamLikely() bool {
	content := strings.ToLower(strings.TrimSpace(c.Content))
	
	// 空内容检测
	if len(content) == 0 {
		return true
	}
	
	// 重复字符检测 - 连续相同字符超过5个
	for i := 0; i < len(content)-4; i++ {
		if content[i] == content[i+1] && 
		   content[i] == content[i+2] && 
		   content[i] == content[i+3] && 
		   content[i] == content[i+4] {
			return true
		}
	}
	
	// 常见垃圾信息模式检测
	spamPatterns := []string{
		"点击链接",
		"立即购买",
		"免费获得",
		"http://",
		"https://",
		"www.",
		"加微信",
		"扫码",
		"联系电话",
		"qq:",
		"微信:",
		"代理",
		"赚钱",
		"投资",
		"理财",
		"贷款",
		"刷单",
		"兼职",
	}
	
	for _, pattern := range spamPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	
	// 纯数字或特殊字符检测
	if len(content) > 10 {
		digitCount := 0
		specialCount := 0
		for _, char := range content {
			if char >= '0' && char <= '9' {
				digitCount++
			} else if !((char >= 'a' && char <= 'z') || 
					   (char >= 'A' && char <= 'Z') || 
					   (char >= '\u4e00' && char <= '\u9fff')) { // 中文字符
				specialCount++
			}
		}
		
		// 如果数字或特殊字符占比超过60%，认为是垃圾信息
		if float64(digitCount+specialCount)/float64(len(content)) > 0.6 {
			return true
		}
	}
	
	return false
}

// GetDisplayAuthor 获取显示的作者信息（考虑匿名）
func (c *Comment) GetDisplayAuthor(userInfo *UserBasicInfo) *UserBasicInfo {
	if c.IsAnonymous {
		return &UserBasicInfo{
			ID:       "anonymous",
			Username: "匿名用户",
			Nickname: "匿名用户",
			Avatar:   "",
		}
	}
	return userInfo
}

// CanReport 判断用户是否可以举报评论
func (c *Comment) CanReport(userID string) bool {
	// 不能举报自己的评论
	if c.UserID == userID {
		return false
	}
	// 只能举报活跃状态的评论
	return c.Status == CommentStatusActive
}

// GetDisplayContent 获取显示的内容（可能经过处理）
func (c *Comment) GetDisplayContent() string {
	if c.Status == CommentStatusDeleted {
		return "[此评论已被删除]"
	}
	if c.Status == CommentStatusHidden {
		return "[此评论已被隐藏]"
	}
	return c.Content
}

// GetNetLikes 获取净点赞数（点赞数-踩数）
func (c *Comment) GetNetLikes() int {
	return c.LikeCount - c.DislikeCount
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


// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments     []CommentResponse `json:"comments"`
	Total        int64             `json:"total"`        // 总数（兼容service设置）
	TotalCount   int               `json:"total_count"`  // 总数（兼容前端）
	Page         int               `json:"page"`         // 当前页（兼容service设置）
	Limit        int               `json:"limit"`        // 每页限制
	TotalPages   int               `json:"total_pages"`
	CurrentPage  int               `json:"current_page"` // 当前页（兼容前端）
	HasNext      bool              `json:"has_next"`
	HasPrevious  bool              `json:"has_previous"`
	Stats        CommentStats      `json:"stats"`        // 统计信息
}

// CommentStats 评论统计
type CommentStats struct {
	TotalComments    int `json:"total_comments"`    // 总评论数
	TotalReplies     int `json:"total_replies"`     // 总回复数
	ActiveComments   int `json:"active_comments"`   // 活跃评论数
	PendingComments  int `json:"pending_comments"`  // 待审核评论数
	ReportedComments int `json:"reported_comments"` // 被举报评论数
	TotalLikes       int `json:"total_likes"`       // 总点赞数
	// 向后兼容字段
	TotalCount     int `json:"total_count"`
	LikedCount     int `json:"liked_count"`
	RepliedCount   int `json:"replied_count"`
	TopComments    int `json:"top_comments"`
	RecentComments int `json:"recent_comments"`
}

// CommentModerationRequest 评论审核请求
type CommentModerationRequest struct {
	Action       string `json:"action" binding:"required" validate:"oneof=approve reject hide"`
	Reason       string `json:"reason" binding:"max=500" validate:"max=500"`
	Notify       bool   `json:"notify"`                                 // 是否通知用户
	BanDuration  int    `json:"ban_duration,omitempty"`                 // 封禁时长（分钟）
	AutoModerate bool   `json:"auto_moderate"`                          // 是否自动审核相似内容
}
