package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScanRecord - SOTA模型设计：复用现有模式，增强协同性
type ScanRecord struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	CourierID   string    `json:"courier_id" gorm:"type:varchar(36);not null;index"`
	LetterCode  string    `json:"letter_code" gorm:"type:varchar(20);not null;index"`
	ScanType    string    `json:"scan_type" gorm:"type:varchar(20);not null"` // pickup, delivery, transit
	Location    string    `json:"location" gorm:"type:varchar(255)"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timestamp   time.Time `json:"timestamp" gorm:"not null;index"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`

	// 关联关系 - 利用现有模型，避免重复
	Courier *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID;references:UserID"`
}

// TableName - 遵循现有命名约定
func (ScanRecord) TableName() string {
	return "scan_records"
}

// BeforeCreate - 复用UUID生成模式
func (s *ScanRecord) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}