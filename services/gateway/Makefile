# OpenPenPal API Gateway Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop deps lint fmt vet

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := api-gateway
DOCKER_IMAGE := openpenpal/api-gateway
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GO_VERSION := $(shell go version | awk '{print $$3}')

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## help: 显示帮助信息
help:
	@echo "OpenPenPal API Gateway 开发工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## deps: 安装依赖
deps:
	@echo "$(YELLOW)安装 Go 依赖...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)依赖安装完成$(NC)"

## fmt: 格式化代码
fmt:
	@echo "$(YELLOW)格式化代码...$(NC)"
	go fmt ./...
	@echo "$(GREEN)代码格式化完成$(NC)"

## lint: 代码检查
lint:
	@echo "$(YELLOW)运行代码检查...$(NC)"
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过高级检查$(NC)"; \
	fi
	@echo "$(GREEN)代码检查完成$(NC)"

## test: 运行测试
test:
	@echo "$(YELLOW)运行测试...$(NC)"
	go test -v ./...
	@echo "$(GREEN)测试完成$(NC)"

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "$(YELLOW)运行测试并生成覆盖率报告...$(NC)"
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告: coverage.html$(NC)"

## build: 构建应用
build: fmt lint
	@echo "$(YELLOW)构建应用...$(NC)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" \
		-o bin/$(APP_NAME) cmd/main.go
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)$(NC)"

## run: 运行应用
run:
	@echo "$(YELLOW)启动 API Gateway...$(NC)"
	go run cmd/main.go

## dev: 开发模式运行
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "$(YELLOW)启动热重载开发服务器...$(NC)"; \
		air; \
	else \
		echo "$(YELLOW)air 未安装，使用普通模式...$(NC)"; \
		$(MAKE) run; \
	fi

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "$(YELLOW)构建 Docker 镜像...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker 镜像构建完成$(NC)"

## docker-run: 运行 Docker 容器
docker-run:
	@echo "$(YELLOW)启动 Docker 服务...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Docker 服务已启动$(NC)"

## docker-stop: 停止 Docker 容器
docker-stop:
	@echo "$(YELLOW)停止 Docker 服务...$(NC)"
	docker-compose down
	@echo "$(GREEN)Docker 服务已停止$(NC)"

## docker-logs: 查看日志
docker-logs:
	docker-compose logs -f api-gateway

## monitoring: 启动监控服务
monitoring:
	@echo "$(YELLOW)启动监控服务...$(NC)"
	docker-compose --profile monitoring up -d
	@echo "$(GREEN)监控服务已启动$(NC)"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

## production: 生产环境部署
production:
	@echo "$(YELLOW)启动生产环境...$(NC)"
	docker-compose --profile production up -d
	@echo "$(GREEN)生产环境已启动$(NC)"

## clean: 清理构建文件
clean:
	@echo "$(YELLOW)清理构建文件...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f
	@echo "$(GREEN)清理完成$(NC)"

## health: 检查服务健康状态
health:
	@echo "$(YELLOW)检查服务健康状态...$(NC)"
	@curl -s http://localhost:8000/health | jq '.' || echo "Gateway 未运行"
	@curl -s http://localhost:8000/admin/gateway/services | jq '.' || echo "服务状态检查失败"

## metrics: 查看监控指标
metrics:
	@echo "$(YELLOW)获取监控指标...$(NC)"
	@curl -s http://localhost:8000/metrics | head -20

## reload: 重新加载配置
reload:
	@echo "$(YELLOW)重新加载配置...$(NC)"
	@curl -X POST http://localhost:8000/admin/gateway/reload \
		-H "Authorization: Bearer ${TOKEN}" | jq '.'

## status: 显示项目状态
status:
	@echo "$(BLUE)项目状态:$(NC)"
	@echo "应用: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go版本: $(GO_VERSION)"
	@echo ""
	@echo "$(BLUE)Docker 状态:$(NC)"
	@docker-compose ps 2>/dev/null || echo "Docker Compose 未运行"
	@echo ""
	@echo "$(BLUE)端口占用:$(NC)"
	@lsof -i :8000 2>/dev/null || echo "端口 8000 未被占用"

## install-tools: 安装开发工具
install-tools:
	@echo "$(YELLOW)安装开发工具...$(NC)"
	@if ! command -v air >/dev/null 2>&1; then \
		echo "安装 air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "$(GREEN)开发工具安装完成$(NC)"

## bench: 性能测试
bench:
	@echo "$(YELLOW)运行性能测试...$(NC)"
	@if command -v wrk >/dev/null 2>&1; then \
		wrk -t12 -c400 -d30s http://localhost:8000/health; \
	else \
		echo "$(RED)请安装 wrk 进行性能测试$(NC)"; \
	fi

## integration-test: 集成测试
integration-test:
	@echo "$(YELLOW)运行集成测试...$(NC)"
	@echo "测试健康检查..."
	@curl -f http://localhost:8000/health > /dev/null
	@echo "测试认证路由..."
	@curl -f http://localhost:8000/api/v1/auth/login -X POST > /dev/null || echo "认证测试失败（预期）"
	@echo "$(GREEN)集成测试完成$(NC)"

## quick: 快速开发流程
quick: deps fmt lint test run

## full: 完整开发流程  
full: clean deps install-tools fmt lint test-coverage build docker-build

## check-env: 检查环境
check-env:
	@echo "$(YELLOW)检查开发环境...$(NC)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)Go 未安装$(NC)" && exit 1)
	@command -v docker >/dev/null 2>&1 || (echo "$(RED)Docker 未安装$(NC)" && exit 1)
	@[ -f .env ] || (echo "$(YELLOW)复制环境配置...$(NC)" && cp .env.example .env)
	@echo "$(GREEN)环境检查通过$(NC)"