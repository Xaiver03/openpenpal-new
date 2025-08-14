package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

// ScanEventType 扫描事件类型
type ScanEventType string

const (
	ScanEventTypeBind     ScanEventType = "bind"     // 绑定
	ScanEventTypePickup   ScanEventType = "pickup"   // 取件
	ScanEventTypeTransit  ScanEventType = "transit"  // 转运
	ScanEventTypeDelivery ScanEventType = "delivery" // 送达
	ScanEventTypeCancel   ScanEventType = "cancel"   // 取消
)

// ScanEvent 扫描事件模型 - PRD要求的完整扫描历史
type ScanEvent struct {
	ID            string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	BarcodeID     string            `json:"barcode_id" gorm:"type:varchar(36);not null;index"`
	LetterCodeID  string            `json:"letter_code_id" gorm:"type:varchar(36);not null;index"` // 关联LetterCode
	ScannedBy     string            `json:"scanned_by" gorm:"type:varchar(36);not null;index"`     // 扫描人员ID
	ScanType      ScanEventType     `json:"scan_type" gorm:"type:varchar(20);not null"`            // 扫描类型
	
	// 位置信息
	Location      string            `json:"location" gorm:"type:varchar(255)"`                     // 位置描述
	OPCode        string            `json:"op_code" gorm:"type:varchar(6);index"`                  // 扫描位置OP Code
	Latitude      *float64          `json:"latitude,omitempty"`                                    // 纬度
	Longitude     *float64          `json:"longitude,omitempty"`                                   // 经度
	
	// 状态信息
	OldStatus     BarcodeStatus     `json:"old_status" gorm:"type:varchar(20)"`                    // 扫描前状态
	NewStatus     BarcodeStatus     `json:"new_status" gorm:"type:varchar(20)"`                    // 扫描后状态
	
	// 扫描详情
	DeviceInfo    string            `json:"device_info,omitempty" gorm:"type:text"`                // 设备信息
	UserAgent     string            `json:"user_agent,omitempty" gorm:"type:text"`                 // 用户代理
	IPAddress     string            `json:"ip_address,omitempty" gorm:"type:varchar(45)"`          // IP地址
	Note          string            `json:"note,omitempty" gorm:"type:text"`                       // 备注
	
	// 元数据
	Metadata      map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`            // 扩展数据
	
	// 时间戳
	Timestamp     time.Time         `json:"timestamp" gorm:"not null;index"`                      // 扫描时间
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`
	
	// 关联关系
	LetterCode    *LetterCode       `json:"letter_code,omitempty" gorm:"foreignKey:LetterCodeID;references:ID;constraint:OnDelete:CASCADE;"`
	Scanner       *User             `json:"scanner,omitempty" gorm:"foreignKey:ScannedBy;references:ID;constraint:OnDelete:SET NULL;"`
}

// ScanEventSummary 扫描事件摘要
type ScanEventSummary struct {
	TotalScans      int64                    `json:"total_scans"`
	UniqueUsers     int64                    `json:"unique_users"`
	UniqueLocations int64                    `json:"unique_locations"`
	ByType          map[ScanEventType]int64  `json:"by_type"`
	ByStatus        map[BarcodeStatus]int64  `json:"by_status"`
	ByHour          map[int]int64            `json:"by_hour"`
	RecentEvents    []ScanEvent              `json:"recent_events"`
}

// ScanEventQuery 扫描事件查询参数
type ScanEventQuery struct {
	BarcodeID     string            `form:"barcode_id"`
	LetterCodeID  string            `form:"letter_code_id"`
	ScannedBy     string            `form:"scanned_by"`
	ScanType      ScanEventType     `form:"scan_type"`
	OPCode        string            `form:"op_code"`
	OldStatus     BarcodeStatus     `form:"old_status"`
	NewStatus     BarcodeStatus     `form:"new_status"`
	StartTime     *time.Time        `form:"start_time"`
	EndTime       *time.Time        `form:"end_time"`
	Page          int               `form:"page,default=1"`
	PageSize      int               `form:"page_size,default=20"`
	OrderBy       string            `form:"order_by,default=timestamp"`
	OrderDesc     bool              `form:"order_desc,default=true"`
}

// ScanEventCreateRequest 创建扫描事件请求
type ScanEventCreateRequest struct {
	BarcodeID     string                 `json:"barcode_id" binding:"required"`
	ScanType      ScanEventType          `json:"scan_type" binding:"required"`
	Location      string                 `json:"location,omitempty"`
	OPCode        string                 `json:"op_code,omitempty"`
	Latitude      *float64               `json:"latitude,omitempty"`
	Longitude     *float64               `json:"longitude,omitempty"`
	OldStatus     BarcodeStatus          `json:"old_status" binding:"required"`
	NewStatus     BarcodeStatus          `json:"new_status" binding:"required"`
	DeviceInfo    string                 `json:"device_info,omitempty"`
	Note          string                 `json:"note,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TableName 指定表名
func (ScanEvent) TableName() string {
	return "scan_events"
}

// BeforeCreate GORM钩子 - 设置默认值
func (se *ScanEvent) BeforeCreate(tx *gorm.DB) error {
	if se.ID == "" {
		se.ID = generateUUID()
	}
	if se.Timestamp.IsZero() {
		se.Timestamp = time.Now()
	}
	return nil
}

// GetScanTypeDisplayName 获取扫描类型显示名称
func (se *ScanEvent) GetScanTypeDisplayName() string {
	displayNames := map[ScanEventType]string{
		ScanEventTypeBind:     "绑定",
		ScanEventTypePickup:   "取件",
		ScanEventTypeTransit:  "转运",
		ScanEventTypeDelivery: "送达",
		ScanEventTypeCancel:   "取消",
	}
	
	if name, exists := displayNames[se.ScanType]; exists {
		return name
	}
	return string(se.ScanType)
}

// IsLocationValid 检查位置信息是否有效
func (se *ScanEvent) IsLocationValid() bool {
	return se.Location != "" || se.OPCode != "" || (se.Latitude != nil && se.Longitude != nil)
}

// GetLocationDescription 获取位置描述
func (se *ScanEvent) GetLocationDescription() string {
	if se.Location != "" {
		return se.Location
	}
	if se.OPCode != "" {
		return "OP Code: " + se.OPCode
	}
	if se.Latitude != nil && se.Longitude != nil {
		return fmt.Sprintf("坐标: %.6f, %.6f", *se.Latitude, *se.Longitude)
	}
	return "位置未知"
}

// ValidateScanType 验证扫描类型
func ValidateScanType(scanType ScanEventType) bool {
	validTypes := []ScanEventType{
		ScanEventTypeBind,
		ScanEventTypePickup,
		ScanEventTypeTransit,
		ScanEventTypeDelivery,
		ScanEventTypeCancel,
	}
	
	for _, validType := range validTypes {
		if scanType == validType {
			return true
		}
	}
	return false
}

// generateUUID 生成UUID（简化版）
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}