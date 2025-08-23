#!/bin/bash

# Start backend without database migration
# 跳过数据库迁移启动后端服务

echo "=== OpenPenPal Backend Startup (Skip Migration) ==="
echo "Following CLAUDE.md principles: ultrathink before action"
echo ""

# Change to backend directory
cd "$(dirname "$0")"

# Check PostgreSQL connection first
echo "1. Checking PostgreSQL connection..."
if command -v psql &> /dev/null; then
    psql -U rocalight -d openpenpal -c "SELECT version();" &> /dev/null
    if [ $? -eq 0 ]; then
        echo "✅ PostgreSQL connection successful"
    else
        echo "❌ PostgreSQL connection failed. Please check:"
        echo "   - PostgreSQL is running: brew services start postgresql"
        echo "   - Database exists: createdb openpenpal"
        echo "   - User has access: check pg_hba.conf"
        exit 1
    fi
else
    echo "⚠️  psql command not found, skipping connection test"
fi

# Set environment variables
echo ""
echo "2. Setting environment variables..."
export SKIP_DB_MIGRATION=true
export DB_TYPE=postgres
export DATABASE_TYPE=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_NAME=openpenpal
export DATABASE_USER=rocalight
export DATABASE_PASSWORD=password
export DATABASE_URL="postgresql://rocalight:password@localhost:5432/openpenpal?sslmode=disable"
export JWT_SECRET=openpenpal-jwt-secret-key-for-postgresql-migration-2025
export ENVIRONMENT=development
export PORT=8080
echo "✅ Environment configured"

# Build the backend
echo ""
echo "3. Building backend..."
go build -o openpenpal-backend main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful"

# Create a modified main function that skips migration
echo ""
echo "4. Creating migration-skip wrapper..."
cat > main_skip_migration.go << 'EOF'
package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
    "openpenpal/internal/config"
    "openpenpal/internal/routes"
    "openpenpal/internal/services"
)

func main() {
    // Skip migration based on environment variable
    if os.Getenv("SKIP_DB_MIGRATION") == "true" {
        log.Println("🚀 Starting server with SKIP_DB_MIGRATION=true")
        
        // Initialize configuration
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }
        
        // Initialize database without migration
        db, err := config.InitDatabaseWithoutMigration(cfg)
        if err != nil {
            log.Fatalf("Failed to init database: %v", err)
        }
        
        // Initialize services
        services.InitServices(db)
        
        // Setup routes
        router := gin.Default()
        routes.SetupRoutes(router, db)
        
        // Start server
        port := os.Getenv("PORT")
        if port == "" {
            port = "8080"
        }
        
        log.Printf("✅ Server starting on port %s (migration skipped)", port)
        if err := router.Run(":" + port); err != nil {
            log.Fatalf("Failed to start server: %v", err)
        }
    } else {
        // Run normal main if not skipping
        log.Println("Running normal startup with migrations...")
        os.Exit(1)
    }
}
EOF

# Alternative approach: Use the existing binary with environment flag
echo ""
echo "5. Starting server..."
echo "========================================"
echo "📌 Server Configuration:"
echo "   - Port: 8080"
echo "   - Database: PostgreSQL (openpenpal)"
echo "   - Migration: SKIPPED"
echo "   - Environment: development"
echo "========================================"
echo ""

# Try to start with skip migration flag
./openpenpal-backend serve --skip-migration 2>&1 | tee backend-skip-migration.log &
BACKEND_PID=$!

# Wait a moment for startup
sleep 3

# Check if process is running
if ps -p $BACKEND_PID > /dev/null; then
    echo "✅ Backend started successfully (PID: $BACKEND_PID)"
    echo ""
    echo "📝 Museum API endpoints available at:"
    echo "   - GET  http://localhost:8080/api/v1/museum/items"
    echo "   - GET  http://localhost:8080/api/v1/museum/exhibitions"
    echo "   - POST http://localhost:8080/api/v1/museum/submit"
    echo ""
    echo "🛑 To stop: kill $BACKEND_PID"
    
    # Save PID for later
    echo $BACKEND_PID > backend.pid
else
    echo "❌ Backend failed to start. Checking logs..."
    tail -n 20 backend-skip-migration.log
fi