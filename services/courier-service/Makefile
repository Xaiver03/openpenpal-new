# Courier Service Test Makefile
# Usage: make test, make coverage, make test-models, etc.

.PHONY: test coverage test-models test-services test-integration test-verbose test-race help deps clean

# Default target
all: deps test

# Install test dependencies
deps:
	@echo "📦 Installing test dependencies..."
	go mod tidy
	go mod download

# Run all tests
test:
	@echo "🧪 Running all tests..."
	go run test_runner.go

# Run tests with verbose output
test-verbose:
	@echo "🔍 Running tests with verbose output..."
	GIN_MODE=test go test -v ./tests/...

# Run tests with race detection
test-race:
	@echo "🏃 Running tests with race detection..."
	GIN_MODE=test go test -race ./tests/...

# Run only model tests
test-models:
	@echo "📋 Running model tests..."
	GIN_MODE=test go test -v ./tests/models/...

# Run only service tests
test-services:
	@echo "⚙️ Running service tests..."
	GIN_MODE=test go test -v ./tests/services/...

# Run only integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	GIN_MODE=test go test -v ./tests/integration/...

# Generate coverage report
coverage:
	@echo "📊 Generating coverage report..."
	GIN_MODE=test go test -cover -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o=coverage.html
	@echo "✅ Coverage report saved to coverage.html"

# Generate coverage and open in browser (macOS)
coverage-open: coverage
	@echo "🌐 Opening coverage report in browser..."
	open coverage.html

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	GIN_MODE=test go test -bench=. ./tests/...

# Clean test artifacts
clean:
	@echo "🧹 Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	go clean -testcache

# Help target
help:
	@echo "🚀 Courier Service Test Commands"
	@echo "================================="
	@echo ""
	@echo "Basic Commands:"
	@echo "  make test            - Run all tests with custom runner"
	@echo "  make test-verbose    - Run tests with verbose output"
	@echo "  make coverage        - Generate coverage report"
	@echo "  make coverage-open   - Generate and open coverage in browser"
	@echo ""
	@echo "Specific Test Categories:"
	@echo "  make test-models     - Run only model tests"
	@echo "  make test-services   - Run only service tests"
	@echo "  make test-integration- Run only integration tests"
	@echo ""
	@echo "Advanced Testing:"
	@echo "  make test-race       - Run tests with race detection"
	@echo "  make bench           - Run benchmarks"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean           - Clean test artifacts"
	@echo "  make deps            - Install dependencies"
EOF < /dev/null