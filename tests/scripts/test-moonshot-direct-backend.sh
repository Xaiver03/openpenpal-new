#!/bin/bash

# Test Moonshot API directly through backend service
# This bypasses the handler layer to test the service directly

echo "Testing Moonshot API through backend service..."
echo "========================================"

# First, let's create a simple Go test program
cat > test-moonshot-service.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "openpenpal-backend/internal/config"
    "openpenpal-backend/internal/models"
    "openpenpal-backend/internal/services"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    fmt.Printf("AI Provider: %s\n", cfg.AIProvider)
    fmt.Printf("Moonshot API Key: %s...%s\n", cfg.MoonshotAPIKey[:6], cfg.MoonshotAPIKey[len(cfg.MoonshotAPIKey)-4:])
    
    // Connect to database
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DatabaseName, cfg.DBSSLMode)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Create AI service
    aiService := services.NewAIService(db, cfg)
    
    // Test direct AI call
    fmt.Println("\nTesting GetInspiration directly (bypassing limits)...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    req := &models.AIInspirationRequest{
        Theme: "科技与未来",
        Count: 1,
    }
    
    response, err := aiService.GetInspiration(ctx, req)
    if err != nil {
        log.Printf("ERROR: %v", err)
    } else {
        fmt.Printf("SUCCESS! Got %d inspirations\n", len(response.Inspirations))
        for i, insp := range response.Inspirations {
            fmt.Printf("\nInspiration %d:\n", i+1)
            fmt.Printf("Theme: %s\n", insp.Theme)
            fmt.Printf("Prompt: %s\n", insp.Prompt)
            fmt.Printf("Style: %s\n", insp.Style)
            fmt.Printf("Tags: %v\n", insp.Tags)
        }
    }
}
EOF

# Run the test
echo "Compiling and running test..."
cd backend
go run ../test-moonshot-service.go

# Clean up
cd ..
rm test-moonshot-service.go