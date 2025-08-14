package models

import (
	"gorm.io/gorm"
)

// BeforeSave hook for UserProfile to ensure JSON fields are not empty
func (u *UserProfile) BeforeSave(tx *gorm.DB) error {
	if u.Preferences == "" {
		u.Preferences = "{}"
	}
	return nil
}

// BeforeSave hook for Courier to ensure JSON fields are not empty
func (c *Courier) BeforeSave(tx *gorm.DB) error {
	if c.TimeSlots == "" {
		c.TimeSlots = "[]"
	}
	return nil
}

// BeforeSave hook for StorageFile to ensure JSON fields are not empty
func (s *StorageFile) BeforeSave(tx *gorm.DB) error {
	if s.Metadata == "" {
		s.Metadata = "{}"
	}
	return nil
}

// BeforeSave hook for StorageConfig to ensure JSON fields are not empty
func (s *StorageConfig) BeforeSave(tx *gorm.DB) error {
	if s.Config == "" {
		s.Config = "{}"
	}
	return nil
}

// BeforeSave hook for ScheduledTask to ensure JSON fields are not empty
func (s *ScheduledTask) BeforeSave(tx *gorm.DB) error {
	if s.Payload == "" {
		s.Payload = "{}"
	}
	return nil
}

// BeforeSave hook for TaskTemplate to ensure JSON fields are not empty
func (t *TaskTemplate) BeforeSave(tx *gorm.DB) error {
	if t.DefaultPayload == "" {
		t.DefaultPayload = "{}"
	}
	return nil
}

// BeforeSave hook for EnvelopeOrder to ensure JSON fields are not empty
func (e *EnvelopeOrder) BeforeSave(tx *gorm.DB) error {
	if e.DeliveryInfo == "" {
		e.DeliveryInfo = "{}"
	}
	return nil
}

// BeforeSave hook for LetterTemplate to ensure JSON fields are not empty
func (l *LetterTemplate) BeforeSave(tx *gorm.DB) error {
	if l.StyleConfig == "" {
		l.StyleConfig = `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`
	}
	return nil
}
