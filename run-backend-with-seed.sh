#!/bin/bash

# Script to run backend with seed data in development mode

echo "🚀 Starting OpenPenPal backend with seed data..."

# Change to backend directory
cd backend

# Set environment to development to trigger seed data
export ENVIRONMENT=development
export DATABASE_TYPE=postgres
export DATABASE_NAME=openpenpal
export DATABASE_URL="postgres://$(whoami):password@localhost:5432/openpenpal"
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=$(whoami)
export DB_PASSWORD=password
export DB_SSLMODE=disable

# Other necessary environment variables
export JWT_SECRET=dev-secret-key-do-not-use-in-production
export BCRYPT_COST=10
export FRONTEND_URL=http://localhost:3000
export PORT=8080
export HOST=0.0.0.0

echo "📊 Database configuration:"
echo "  - Type: PostgreSQL"
echo "  - Database: openpenpal"
echo "  - User: $DB_USER"
echo "  - Host: $DB_HOST:$DB_PORT"
echo "  - Environment: $ENVIRONMENT (seed data will be loaded)"

# Run the backend
echo ""
echo "🏃 Starting backend service..."
go run main.go