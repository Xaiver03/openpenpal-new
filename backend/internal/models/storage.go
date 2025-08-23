package models

import (
	"gorm.io/gorm"
	"time"
)

// StorageProvider 存储提供商类型
type StorageProvider string

const (
	StorageProviderLocal      StorageProvider = "local"       // 本地存储
	StorageProviderAliOSS     StorageProvider = "ali_oss"     // 阿里云OSS
	StorageProviderTencentCOS StorageProvider = "tencent_cos" // 腾讯云COS
	StorageProviderAwsS3      StorageProvider = "aws_s3"      // AWS S3
	StorageProviderQiniu      StorageProvider = "qiniu"       // 七牛云
	StorageProviderUpyun      StorageProvider = "upyun"       // 又拍云
)

// FileCategory 文件分类
type FileCategory string

const (
	FileCategoryImage            FileCategory = "image"            // 图片文件
	FileCategoryDocument         FileCategory = "document"         // 文档文件
	FileCategoryQRCode           FileCategory = "qrcode"           // 二维码
	FileCategoryAvatar           FileCategory = "avatar"           // 头像
	FileCategoryEnvelope         FileCategory = "envelope"         // 信封图片
	FileCategoryAttachment       FileCategory = "attachment"       // 附件
	FileCategoryThumbnail        FileCategory = "thumbnail"        // 缩略图
	FileCategoryBackup           FileCategory = "backup"           // 备份文件
	FileCategoryMuseum           FileCategory = "museum"           // 博物馆图片
	FileCategoryHandwrittenLetter FileCategory = "handwritten_letter" // 手写信件图片
)

// FileStatus 文件状态
type FileStatus string

const (
	FileStatusUploading FileStatus = "uploading" // 上传中
	FileStatusActive    FileStatus = "active"    // 活跃状态
	FileStatusArchived  FileStatus = "archived"  // 已归档
	FileStatusDeleted   FileStatus = "deleted"   // 已删除
	FileStatusCorrupted FileStatus = "corrupted" // 文件损坏
)

// StorageFile 存储文件模型
type StorageFile struct {
	ID           string       `json:"id" gorm:"primaryKey"`
	FileName     string       `json:"file_name" gorm:"size:255;not null"`
	OriginalName string       `json:"original_name" gorm:"size:255;not null"`
	FileSize     int64        `json:"file_size" gorm:"not null"`
	MimeType     string       `json:"mime_type" gorm:"size:100"`
	Extension    string       `json:"extension" gorm:"size:20"`
	Category     FileCategory `json:"category" gorm:"size:50;not null"`

	// 存储信息
	Provider   StorageProvider `json:"provider" gorm:"size:50;not null"`
	BucketName string          `json:"bucket_name" gorm:"size:100"`
	ObjectKey  string          `json:"object_key" gorm:"size:500;not null"` // 存储对象键/路径
	LocalPath  string          `json:"local_path" gorm:"size:500"`          // 本地路径（本地存储使用）

	// URL信息
	PublicURL    string `json:"public_url" gorm:"size:1000"`    // 公共访问URL
	PrivateURL   string `json:"private_url" gorm:"size:1000"`   // 私有访问URL（带签名）
	ThumbnailURL string `json:"thumbnail_url" gorm:"size:1000"` // 缩略图URL

	// 元数据
	Metadata string     `json:"metadata" gorm:"type:json"` // 额外元数据（JSON格式）
	Status   FileStatus `json:"status" gorm:"size:20;default:'active'"`

	// 关联信息
	UploadedBy  string `json:"uploaded_by" gorm:"size:50;index"` // 上传者ID
	RelatedType string `json:"related_type" gorm:"size:50"`      // 关联类型（letter, user, envelope等）
	RelatedID   string `json:"related_id" gorm:"size:50;index"`  // 关联对象ID

	// 安全信息
	HashMD5    string `json:"hash_md5" gorm:"size:32"`    // 文件MD5哈希
	HashSHA256 string `json:"hash_sha256" gorm:"size:64"` // 文件SHA256哈希

	// 访问控制
	IsPublic    bool       `json:"is_public" gorm:"default:false"` // 是否公开访问
	AccessToken string     `json:"access_token" gorm:"size:100"`   // 访问令牌
	ExpiresAt   *time.Time `json:"expires_at"`                     // 过期时间

	// 审计字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// StorageConfig 存储配置模型
type StorageConfig struct {
	ID          string          `json:"id" gorm:"primaryKey"`
	Provider    StorageProvider `json:"provider" gorm:"size:50;not null;unique"`
	DisplayName string          `json:"display_name" gorm:"size:100;not null"`

	// 配置信息
	Config    string `json:"config" gorm:"type:json;not null"` // 配置信息（JSON格式）
	IsEnabled bool   `json:"is_enabled" gorm:"default:false"`  // 是否启用
	IsDefault bool   `json:"is_default" gorm:"default:false"`  // 是否为默认存储
	Priority  int    `json:"priority" gorm:"default:1"`        // 优先级

	// 容量限制
	MaxFileSize  int64 `json:"max_file_size" gorm:"default:104857600"` // 最大文件大小（默认100MB）
	MaxTotalSize int64 `json:"max_total_size"`                         // 最大总容量
	CurrentSize  int64 `json:"current_size" gorm:"default:0"`          // 当前使用容量

	// 支持的文件类型
	AllowedTypes string `json:"allowed_types" gorm:"type:text"` // 允许的文件类型（JSON数组）

	// 审计字段
	CreatedBy string         `json:"created_by" gorm:"size:50"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// StorageOperation 存储操作记录
type StorageOperation struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	FileID    string      `json:"file_id" gorm:"size:50;not null;index"`
	File      StorageFile `json:"file" gorm:"foreignKey:FileID"`
	Operation string      `json:"operation" gorm:"size:50;not null"` // upload, download, delete, move, copy
	UserID    string      `json:"user_id" gorm:"size:50;index"`
	IPAddress string      `json:"ip_address" gorm:"size:45"`
	UserAgent string      `json:"user_agent" gorm:"size:500"`

	// 操作结果
	Status           string `json:"status" gorm:"size:20"`      // success, failed, pending
	BytesTransferred int64  `json:"bytes_transferred"`          // 传输字节数
	Duration         int    `json:"duration"`                   // 操作耗时（毫秒）
	ErrorMsg         string `json:"error_msg" gorm:"type:text"` // 错误信息

	// 审计字段
	CreatedAt time.Time `json:"created_at"`
}

// 请求和响应模型

// UploadRequest 文件上传请求
type UploadRequest struct {
	Category    FileCategory `json:"category" binding:"required"`
	RelatedType string       `json:"related_type"`
	RelatedID   string       `json:"related_id"`
	IsPublic    bool         `json:"is_public"`
	ExpiresIn   int          `json:"expires_in"` // 过期时间（秒）
}

// UploadResponse 文件上传响应
type UploadResponse struct {
	FileID       string `json:"file_id"`
	FileName     string `json:"file_name"`
	PublicURL    string `json:"public_url"`
	PrivateURL   string `json:"private_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
}

// FileQuery 文件查询参数
type FileQuery struct {
	Category    FileCategory    `json:"category" form:"category"`
	Provider    StorageProvider `json:"provider" form:"provider"`
	Status      FileStatus      `json:"status" form:"status"`
	RelatedType string          `json:"related_type" form:"related_type"`
	RelatedID   string          `json:"related_id" form:"related_id"`
	UploadedBy  string          `json:"uploaded_by" form:"uploaded_by"`
	StartDate   *time.Time      `json:"start_date" form:"start_date"`
	EndDate     *time.Time      `json:"end_date" form:"end_date"`
	Page        int             `json:"page" form:"page" binding:"min=1"`
	PageSize    int             `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	SortBy      string          `json:"sort_by" form:"sort_by"`
	SortOrder   string          `json:"sort_order" form:"sort_order"`
}

// StorageStats 存储统计信息
type StorageStats struct {
	TotalFiles      int64            `json:"total_files"`
	TotalSize       int64            `json:"total_size"`
	FilesByCategory map[string]int64 `json:"files_by_category"`
	FilesByProvider map[string]int64 `json:"files_by_provider"`
	FilesByStatus   map[string]int64 `json:"files_by_status"`
	RecentUploads   int64            `json:"recent_uploads"` // 最近24小时上传数
	StorageUsage    map[string]int64 `json:"storage_usage"`  // 各存储提供商使用量
	LastUpdate      time.Time        `json:"last_update"`
}

// TableName 指定表名
func (StorageFile) TableName() string {
	return "storage_files"
}

func (StorageConfig) TableName() string {
	return "storage_configs"
}

func (StorageOperation) TableName() string {
	return "storage_operations"
}
