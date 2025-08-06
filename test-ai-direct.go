package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env
	_ = godotenv.Load("backend/.env")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	fmt.Println("=== AI Configuration ===")
	fmt.Printf("Provider: %s\n", cfg.AIProvider)
	fmt.Printf("Moonshot Key: %s...%s\n", cfg.MoonshotAPIKey[:6], cfg.MoonshotAPIKey[len(cfg.MoonshotAPIKey)-4:])
	fmt.Println()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	err = db.AutoMigrate(
		&models.AIConfig{},
		&models.AIInspiration{},
		&models.AIUsageLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	// Create AI service
	aiService := services.NewAIService(db, cfg)

	// Test GetActiveProvider
	fmt.Println("=== Testing GetActiveProvider ===")
	aiConfig, err := aiService.GetActiveProvider()
	if err != nil {
		log.Printf("Error getting active provider: %v", err)
	} else {
		fmt.Printf("Active Provider: %s\n", aiConfig.Provider)
		fmt.Printf("Model: %s\n", aiConfig.Model)
		fmt.Printf("API Key: %s...%s\n", aiConfig.APIKey[:6], aiConfig.APIKey[len(aiConfig.APIKey)-4:])
		fmt.Printf("Endpoint: %s\n", aiConfig.APIEndpoint)
	}
	fmt.Println()

	// Test GetInspiration directly
	fmt.Println("=== Testing GetInspiration ===")
	req := &models.AIInspirationRequest{
		Theme: "日常生活",
		Count: 1,
	}

	ctx := context.Background()
	response, err := aiService.GetInspiration(ctx, req)
	if err != nil {
		log.Printf("Error getting inspiration: %v", err)
		os.Exit(1)
	}

	fmt.Println("Success! Got inspiration:")
	for _, insp := range response.Inspirations {
		fmt.Printf("- Theme: %s\n", insp.Theme)
		fmt.Printf("  Prompt: %s\n", insp.Prompt)
		fmt.Printf("  Style: %s\n", insp.Style)
		fmt.Printf("  Tags: %v\n", insp.Tags)
	}
}