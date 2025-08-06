package models

import (
	"time"
)

// ScanRecord 扫码记录模型 - 增强FSD条码系统支持
type ScanRecord struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TaskID    string    `gorm:"not null" json:"task_id"`
	CourierID string    `gorm:"not null" json:"courier_id"`
	LetterID  string    `gorm:"not null" json:"letter_id"`
	Action    string    `gorm:"not null" json:"action"`       // collected, in_transit, delivered, failed
	Location  string    `json:"location"`                     // 扫码地点
	Latitude  float64   `json:"latitude"`                     // 纬度
	Longitude float64   `json:"longitude"`                    // 经度
	Note      string    `json:"note"`                         // 备注说明
	PhotoURL  string    `json:"photo_url"`                    // 照片证明URL
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	
	// FSD条码系统增强字段
	BarcodeCode       string    `json:"barcode_code,omitempty" gorm:"type:varchar(100);index"`         // 条码编号
	RecipientOPCode   string    `json:"recipient_op_code,omitempty" gorm:"type:varchar(6);index"`       // 收件人OP Code  
	OperatorOPCode    string    `json:"operator_op_code,omitempty" gorm:"type:varchar(6);index"`        // 操作员OP Code位置
	ScannerLevel      int       `json:"scanner_level,omitempty" gorm:"type:int;default:1"`              // 扫码员级别
	ValidationResult  string    `json:"validation_result,omitempty" gorm:"type:varchar(20)"`            // 验证结果
	ValidationMessage string    `json:"validation_message,omitempty" gorm:"type:text"`                 // 验证消息
	BarcodeStatusOld  string    `json:"barcode_status_old,omitempty" gorm:"type:varchar(20)"`           // 扫码前条码状态
	BarcodeStatusNew  string    `json:"barcode_status_new,omitempty" gorm:"type:varchar(20)"`           // 扫码后条码状态
	DeviceInfo        string    `json:"device_info,omitempty" gorm:"type:json"`                        // 设备信息
	IPAddress         string    `json:"ip_address,omitempty" gorm:"type:varchar(45)"`                   // IP地址
	UserAgent         string    `json:"user_agent,omitempty" gorm:"type:text"`                         // 用户代理
}

// ScanRequest 扫码请求 - 增强FSD条码系统支持
type ScanRequest struct {
	Action    string  `json:"action" binding:"required,oneof=collected in_transit delivered failed"`
	Location  string  `json:"location"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Note      string  `json:"note"`
	PhotoURL  string  `json:"photo_url"`
	
	// FSD条码系统增强字段
	BarcodeCode     string `json:"barcode_code,omitempty"`     // 条码编号（兼容现有letter_id）
	RecipientOPCode string `json:"recipient_op_code,omitempty"` // 收件人OP Code
	OperatorOPCode  string `json:"operator_op_code,omitempty"`  // 操作员OP Code位置
	ScannerLevel    int    `json:"scanner_level,omitempty"`     // 扫码员级别（1-4）
	ValidationType  string `json:"validation_type,omitempty"`   // 验证类型：quick/full
}

// ScanResponse 扫码响应 - 增强FSD条码系统支持
type ScanResponse struct {
	LetterID  string    `json:"letter_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ScanTime  time.Time `json:"scan_time"`
	Location  string    `json:"location"`
	
	// FSD条码系统增强字段
	BarcodeCode       string            `json:"barcode_code,omitempty"`       // 条码编号
	BarcodeStatus     string            `json:"barcode_status,omitempty"`     // 条码状态
	RecipientOPCode   string            `json:"recipient_op_code,omitempty"`   // 收件人OP Code
	OperatorOPCode    string            `json:"operator_op_code,omitempty"`    // 操作员OP Code
	ValidationResult  string            `json:"validation_result,omitempty"`   // 验证结果：success/failed
	ValidationMessage string            `json:"validation_message,omitempty"`  // 验证消息
	NextAction        string            `json:"next_action,omitempty"`         // 下一步操作建议
	EstimatedDelivery *time.Time        `json:"estimated_delivery,omitempty"` // 预计送达时间
	Permissions       map[string]bool   `json:"permissions,omitempty"`         // 操作员权限
}

// 扫码动作常量
const (
	ScanActionCollected = "collected"   // 已收取
	ScanActionInTransit = "in_transit"  // 投递中
	ScanActionDelivered = "delivered"   // 已投递
	ScanActionFailed    = "failed"      // 投递失败
)

// ActionToStatus 将扫码动作转换为任务状态
var ActionToStatus = map[string]string{
	ScanActionCollected: TaskStatusCollected,
	ScanActionInTransit: TaskStatusInTransit,
	ScanActionDelivered: TaskStatusDelivered,
	ScanActionFailed:    TaskStatusFailed,
}

// GetTaskStatus 根据扫码动作获取对应的任务状态
func (sr *ScanRecord) GetTaskStatus() string {
	if status, exists := ActionToStatus[sr.Action]; exists {
		return status
	}
	return ""
}

// IsValid 检查扫码记录是否有效 - 增强FSD条码系统验证
func (sr *ScanRecord) IsValid() bool {
	_, exists := ActionToStatus[sr.Action]
	baseValid := exists && sr.TaskID != "" && sr.CourierID != "" && sr.LetterID != ""
	
	// FSD条码系统验证
	if sr.BarcodeCode != "" {
		// 如果有条码编号，验证格式（应该是8位格式：OP7X1F2K）
		if len(sr.BarcodeCode) < 8 {
			return false
		}
	}
	
	if sr.RecipientOPCode != "" {
		// 验证OP Code格式（应该是6位格式：AABBCC）
		if len(sr.RecipientOPCode) != 6 {
			return false
		}
	}
	
	return baseValid
}

// GetValidationStatus 获取验证状态
func (sr *ScanRecord) GetValidationStatus() string {
	if sr.ValidationResult == "" {
		return "pending"
	}
	return sr.ValidationResult
}

// IsSuccessfulValidation 检查验证是否成功
func (sr *ScanRecord) IsSuccessfulValidation() bool {
	return sr.ValidationResult == "success"
}

// HasOPCodeMatch 检查OP Code是否匹配
func (sr *ScanRecord) HasOPCodeMatch() bool {
	return sr.RecipientOPCode != "" && sr.OperatorOPCode != "" && 
		   sr.RecipientOPCode[:4] == sr.OperatorOPCode[:4] // 前4位匹配（学校+区域）
}