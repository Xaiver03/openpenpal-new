# OpenPenPal 统一构建系统
# 使用说明: make help

.PHONY: help install dev build test clean docker-up docker-down

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

# 项目变量
PROJECT_NAME := openpenpal
FRONTEND_DIR := frontend
BACKEND_DIR := backend
SERVICES_DIR := services

## 帮助信息
help: ## 显示帮助信息
	@echo ''
	@echo '使用方法:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<目标>${RESET}'
	@echo ''
	@echo '可用目标:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## 环境设置
install: ## 安装所有依赖
	@echo "📦 安装项目依赖..."
	@if [ -f ./startup/install-deps.sh ]; then \
		./startup/install-deps.sh; \
	else \
		cd $(FRONTEND_DIR) && npm install; \
		cd ../$(BACKEND_DIR) && go mod download; \
	fi
	@echo "✅ 依赖安装完成"

check-deps: ## 检查依赖
	@echo "🔍 检查系统依赖..."
	@command -v node >/dev/null 2>&1 || { echo "❌ 需要安装 Node.js"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ 需要安装 Go"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "❌ 需要安装 Docker"; exit 1; }
	@echo "✅ 依赖检查通过"

## 开发命令
dev: ## 启动开发环境（演示模式）
	@echo "🚀 启动开发环境..."
	@if [ -f ./startup/quick-start.sh ]; then \
		./startup/quick-start.sh demo --auto-open; \
	else \
		make dev-manual; \
	fi

dev-full: ## 启动完整开发环境
	@echo "🚀 启动完整开发环境..."
	./startup/quick-start.sh development --auto-open

dev-manual: ## 手动启动各服务
	@echo "🚀 手动启动服务..."
	@echo "启动前端..."
	cd $(FRONTEND_DIR) && npm run dev &
	@echo "启动后端..."
	cd $(BACKEND_DIR) && go run main.go &
	@echo "✅ 服务启动完成"

stop: ## 停止所有服务
	@echo "🛑 停止所有服务..."
	@if [ -f ./startup/stop-all.sh ]; then \
		./startup/stop-all.sh; \
	else \
		pkill -f "npm|node|go run" || true; \
	fi
	@echo "✅ 服务已停止"

restart: stop dev ## 重启所有服务

status: ## 检查服务状态
	@echo "📊 检查服务状态..."
	@if [ -f ./startup/check-status.sh ]; then \
		./startup/check-status.sh --detailed; \
	else \
		ps aux | grep -E "(npm|node|go)" | grep -v grep || echo "没有运行中的服务"; \
	fi

## 构建命令
build: ## 构建所有服务
	@echo "🔨 构建项目..."
	cd $(FRONTEND_DIR) && npm run build
	cd $(BACKEND_DIR) && go build -o bin/server main.go
	@echo "✅ 构建完成"

build-docker: ## 构建Docker镜像
	@echo "🐳 构建Docker镜像..."
	docker-compose build
	@echo "✅ Docker镜像构建完成"

## 测试命令
test: ## 运行所有测试
	@echo "🧪 运行测试..."
	@make test-unit
	@make test-integration
	@echo "✅ 所有测试通过"

test-unit: ## 运行单元测试
	@echo "🧪 运行单元测试..."
	cd $(FRONTEND_DIR) && npm test
	cd $(BACKEND_DIR) && go test ./...

test-integration: ## 运行集成测试
	@echo "🧪 运行集成测试..."
	@if [ -f ./test-kimi/run_tests.sh ]; then \
		./test-kimi/run_tests.sh; \
	fi

test-api: ## 测试API
	@echo "🧪 测试API..."
	@if [ -f ./scripts/test-apis.sh ]; then \
		./scripts/test-apis.sh; \
	fi

## Docker命令
docker-up: ## 启动Docker环境
	@echo "🐳 启动Docker环境..."
	docker-compose up -d
	@echo "✅ Docker环境已启动"

docker-down: ## 停止Docker环境
	@echo "🐳 停止Docker环境..."
	docker-compose down
	@echo "✅ Docker环境已停止"

docker-logs: ## 查看Docker日志
	docker-compose logs -f

## 代码质量
lint: ## 运行代码检查
	@echo "🔍 检查代码质量..."
	cd $(FRONTEND_DIR) && npm run lint
	cd $(BACKEND_DIR) && golangci-lint run
	@echo "✅ 代码检查完成"

format: ## 格式化代码
	@echo "✨ 格式化代码..."
	cd $(FRONTEND_DIR) && npm run format
	cd $(BACKEND_DIR) && go fmt ./...
	@echo "✅ 代码格式化完成"

## 清理命令
clean: ## 清理构建产物
	@echo "🧹 清理构建产物..."
	rm -rf $(FRONTEND_DIR)/.next
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(BACKEND_DIR)/vendor
	@echo "✅ 清理完成"

clean-all: clean ## 深度清理（包括依赖）
	@echo "🧹 深度清理..."
	./startup/force-cleanup.sh
	@echo "✅ 深度清理完成"

## 文档命令
docs: ## 生成文档
	@echo "📚 生成文档..."
	@echo "TODO: 实现文档生成"

docs-serve: ## 启动文档服务器
	@echo "📚 启动文档服务器..."
	cd docs && python -m http.server 8080

docs-check: ## 检查文档链接和一致性
	@echo "🔍 检查文档质量..."
	./scripts/check-doc-links.sh

docs-fix: ## 自动修复文档问题
	@echo "🔧 自动修复文档问题..."
	@echo "TODO: 实现自动修复脚本"

## 部署命令
deploy-dev: ## 部署到开发环境
	@echo "🚀 部署到开发环境..."
	@echo "TODO: 实现开发环境部署"

deploy-prod: ## 部署到生产环境
	@echo "🚀 部署到生产环境..."
	@echo "TODO: 实现生产环境部署"

## 实用工具
logs: ## 查看日志
	@echo "📋 查看日志..."
	tail -f logs/*.log

port-check: ## 检查端口占用
	@echo "🔍 检查端口占用..."
	@echo "端口 3000 (前端):"
	@lsof -i :3000 || echo "✅ 端口 3000 未被占用"
	@echo "\n端口 8000 (API网关):"
	@lsof -i :8000 || echo "✅ 端口 8000 未被占用"
	@echo "\n端口 8001-8004 (微服务):"
	@for port in 8001 8002 8003 8004; do \
		lsof -i :$$port || echo "✅ 端口 $$port 未被占用"; \
	done

init: check-deps install ## 初始化项目
	@echo "🎉 项目初始化完成！"
	@echo "运行 'make dev' 启动开发环境"