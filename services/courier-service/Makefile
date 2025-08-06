# OpenPenPal 信使服务 Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop deps lint fmt vet

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := courier-service
DOCKER_IMAGE := openpenpal/courier-service
VERSION := $(shell git describe --tags --always --dirty)
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
	@echo "OpenPenPal 信使服务开发工具"
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
	@echo "$(YELLOW)格式化 Go 代码...$(NC)"
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
	@echo "$(YELLOW)运行单元测试...$(NC)"
	go test -v ./...
	@echo "$(GREEN)测试完成$(NC)"

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "$(YELLOW)运行测试并生成覆盖率报告...$(NC)"
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告生成完成: coverage.html$(NC)"

## build: 构建应用
build: fmt lint
	@echo "$(YELLOW)构建应用...$(NC)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)" \
		-o bin/$(APP_NAME) cmd/main.go
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)$(NC)"

## run: 运行应用
run:
	@echo "$(YELLOW)启动应用...$(NC)"
	go run cmd/main.go

## dev: 开发模式运行（热重载）
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "$(YELLOW)启动热重载开发服务器...$(NC)"; \
		air; \
	else \
		echo "$(YELLOW)air 未安装，使用普通模式运行...$(NC)"; \
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

## docker-logs: 查看 Docker 日志
docker-logs:
	docker-compose logs -f $(APP_NAME)

## deploy: 部署服务
deploy:
	@echo "$(YELLOW)部署服务...$(NC)"
	./deploy.sh start
	@echo "$(GREEN)服务部署完成$(NC)"

## migrate: 运行数据库迁移
migrate:
	@echo "$(YELLOW)运行数据库迁移...$(NC)"
	go run cmd/main.go -migrate
	@echo "$(GREEN)数据库迁移完成$(NC)"

## seed: 生成测试数据
seed:
	@echo "$(YELLOW)生成测试数据...$(NC)"
	go run cmd/main.go -seed
	@echo "$(GREEN)测试数据生成完成$(NC)"

## clean: 清理构建文件
clean:
	@echo "$(YELLOW)清理构建文件...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f
	@echo "$(GREEN)清理完成$(NC)"

## install-tools: 安装开发工具
install-tools:
	@echo "$(YELLOW)安装开发工具...$(NC)"
	# 安装 air (热重载)
	@if ! command -v air >/dev/null 2>&1; then \
		echo "安装 air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	# 安装 golangci-lint (代码检查)
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "$(GREEN)开发工具安装完成$(NC)"

## status: 显示项目状态
status:
	@echo "$(BLUE)项目状态信息:$(NC)"
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go 版本: $(GO_VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo ""
	@echo "$(BLUE)Git 状态:$(NC)"
	@git status --short || echo "不是 Git 仓库"
	@echo ""
	@echo "$(BLUE)Docker 状态:$(NC)"
	@docker-compose ps 2>/dev/null || echo "Docker Compose 未运行"

## bench: 运行性能测试
bench:
	@echo "$(YELLOW)运行性能测试...$(NC)"
	go test -bench=. -benchmem ./...
	@echo "$(GREEN)性能测试完成$(NC)"

## doc: 生成文档
doc:
	@echo "$(YELLOW)生成 API 文档...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/main.go; \
		echo "$(GREEN)API 文档生成完成$(NC)"; \
	else \
		echo "$(RED)swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi

## release: 创建发布版本
release: clean fmt lint test build docker-build
	@echo "$(GREEN)发布版本 $(VERSION) 准备完成$(NC)"

## quick: 快速开发流程
quick: deps fmt lint test run

## full: 完整开发流程
full: clean deps install-tools fmt lint test-coverage build docker-build

# 检查必要的环境
check-env:
	@echo "$(YELLOW)检查开发环境...$(NC)"
	@command -v go >/dev/null 2>&1 || (echo "$(RED)Go 未安装$(NC)" && exit 1)
	@command -v docker >/dev/null 2>&1 || (echo "$(RED)Docker 未安装$(NC)" && exit 1)
	@command -v docker-compose >/dev/null 2>&1 || (echo "$(RED)Docker Compose 未安装$(NC)" && exit 1)
	@[ -f .env ] || (echo "$(YELLOW).env 文件不存在，从 .env.example 复制$(NC)" && cp .env.example .env)
	@echo "$(GREEN)环境检查通过$(NC)"