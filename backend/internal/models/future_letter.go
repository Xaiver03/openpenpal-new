package models

import (
	"time"
	"gorm.io/gorm"
)

type FutureLetterStatus string

const (
	FutureLetterStatusScheduled FutureLetterStatus = "scheduled" // 已计划
	FutureLetterStatusSent      FutureLetterStatus = "sent"      // 已发送
	FutureLetterStatusCancelled FutureLetterStatus = "cancelled" // 已取消
)

type FutureLetter struct {
	ID               string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID         string             `json:"letter_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	SenderID         string             `json:"sender_id" gorm:"type:varchar(36);not null;index"`
	RecipientID      string             `json:"recipient_id,omitempty" gorm:"type:varchar(36);index"`
	RecipientOPCode  string             `json:"recipient_op_code,omitempty" gorm:"type:varchar(6);index"`
	Status           FutureLetterStatus `json:"status" gorm:"type:varchar(20);not null;default:'scheduled'"`
	ScheduledDate    time.Time          `json:"scheduled_date" gorm:"not null;index"`
	DeliveryMethod   string             `json:"delivery_method" gorm:"type:varchar(20);default:'system'"` // system/courier
	ReminderEnabled  bool               `json:"reminder_enabled" gorm:"default:true"`
	ReminderDays     int                `json:"reminder_days" gorm:"default:7"` // 提前提醒天数
	LastReminderSent *time.Time         `json:"last_reminder_sent,omitempty"`
	SentAt           *time.Time         `json:"sent_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	DeletedAt        gorm.DeletedAt     `json:"-" gorm:"index"`
	
	// 关联
	Letter    *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID"`
	Sender    *User   `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Recipient *User   `json:"recipient,omitempty" gorm:"foreignKey:RecipientID"`
}

func (FutureLetter) TableName() string {
	return "future_letters"
}