package models

import (
	"time"
	"gorm.io/gorm"
)

type DriftBottleStatus string

const (
	DriftBottleStatusFloating  DriftBottleStatus = "floating"   // 漂流中
	DriftBottleStatusCollected DriftBottleStatus = "collected"  // 已捞取
	DriftBottleStatusExpired   DriftBottleStatus = "expired"    // 已过期
)

type DriftBottle struct {
	ID            string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID      string            `json:"letter_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	SenderID      string            `json:"sender_id" gorm:"type:varchar(36);not null;index"`
	CollectorID   string            `json:"collector_id,omitempty" gorm:"type:varchar(36);index"`
	Status        DriftBottleStatus `json:"status" gorm:"type:varchar(20);not null;default:'floating'"`
	Theme         string            `json:"theme" gorm:"type:varchar(50)"` // 主题标签
	Region        string            `json:"region" gorm:"type:varchar(50)"` // 漂流区域
	CollectedAt   *time.Time        `json:"collected_at,omitempty"`
	ExpiresAt     time.Time         `json:"expires_at" gorm:"not null"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`
	
	// 关联
	Letter    *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID"`
	Sender    *User   `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Collector *User   `json:"collector,omitempty" gorm:"foreignKey:CollectorID"`
}

func (DriftBottle) TableName() string {
	return "drift_bottles"
}