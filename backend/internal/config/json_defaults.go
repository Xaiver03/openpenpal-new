package config

import (
	"gorm.io/gorm"
	"log"
	"openpenpal-backend/internal/models"
)

// SetJSONDefaults sets default values for JSON columns to prevent PostgreSQL errors
func SetJSONDefaults(db *gorm.DB) error {
	log.Println("Setting JSON column defaults...")

	// Set default for Courier TimeSlots
	if err := db.Model(&models.Courier{}).Where("time_slots IS NULL").Update("time_slots", "[]").Error; err != nil {
		log.Printf("Failed to update Courier time_slots: %v", err)
	}

	// Set default for UserProfile Preferences
	if err := db.Model(&models.UserProfile{}).Where("preferences IS NULL").Update("preferences", "{}").Error; err != nil {
		log.Printf("Failed to update UserProfile preferences: %v", err)
	}

	// Set default for StorageFile Metadata
	if err := db.Model(&models.StorageFile{}).Where("metadata IS NULL").Update("metadata", "{}").Error; err != nil {
		log.Printf("Failed to update StorageFile metadata: %v", err)
	}

	// Set default for StorageConfig Config
	if err := db.Model(&models.StorageConfig{}).Where("config IS NULL").Update("config", "{}").Error; err != nil {
		log.Printf("Failed to update StorageConfig config: %v", err)
	}

	// Set default for ScheduledTask Payload
	if err := db.Model(&models.ScheduledTask{}).Where("payload IS NULL").Update("payload", "{}").Error; err != nil {
		log.Printf("Failed to update ScheduledTask payload: %v", err)
	}

	// Set default for TaskTemplate DefaultPayload
	if err := db.Model(&models.TaskTemplate{}).Where("default_payload IS NULL").Update("default_payload", "{}").Error; err != nil {
		log.Printf("Failed to update TaskTemplate default_payload: %v", err)
	}

	// Set default for EnvelopeOrder DeliveryInfo
	if err := db.Model(&models.EnvelopeOrder{}).Where("delivery_info IS NULL").Update("delivery_info", "{}").Error; err != nil {
		log.Printf("Failed to update EnvelopeOrder delivery_info: %v", err)
	}

	// Set default for LetterTemplate StyleConfig
	if err := db.Model(&models.LetterTemplate{}).Where("style_config IS NULL").Update("style_config", `{"fontFamily":"serif","fontSize":"16px","color":"#333"}`).Error; err != nil {
		log.Printf("Failed to update LetterTemplate style_config: %v", err)
	}

	log.Println("JSON column defaults set successfully")
	return nil
}