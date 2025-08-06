# OpenPenPal 启动模式说明

## 启动模式总览

### 1. 🚀 开发模式 (development) - 默认
- **服务**: `go-backend` + `frontend`
- **端口**: 8080 (后端) + 3000 (前端)
- **特点**: 
  - 快速启动，适合日常开发
  - 热重载支持
  - 详细日志输出
  - 使用SQLite数据库（默认）
- **适用场景**: 前后端开发、功能调试

### 2. 🎯 简化模式 (simple)
- **服务**: `go-backend` + `frontend`
- **端口**: 8080 + 3000
- **特点**: 
  - 与开发模式相同，但更简洁的输出
  - 最小化配置
- **适用场景**: 快速体验、简单测试

### 3. 🎭 演示模式 (demo)
- **服务**: `go-backend` + `frontend`
- **端口**: 8080 + 3000
- **特点**: 
  - 自动打开浏览器
  - 预装演示数据
  - 优化的用户体验配置
- **适用场景**: 产品演示、新用户体验

### 4. 🔧 Mock模式 (mock)
- **服务**: `simple-mock` + `frontend`
- **端口**: 8000 (mock网关) + 3000
- **特点**: 
  - 使用Node.js Mock服务
  - 无需后端编译
  - 快速原型开发
- **适用场景**: 前端独立开发、API设计验证

### 5. 🏗️ 完整模式 (complete)
- **服务**: 所有微服务
  - `go-backend` (8080)
  - `real-gateway` (8000)
  - `real-write-service` (8001)
  - `real-courier-service` (8002)
  - `real-admin-service` (8003)
  - `real-ocr-service` (8004)
  - `frontend` (3000)
  - `admin-frontend` (3001)
- **特点**: 
  - 完整微服务架构
  - 需要PostgreSQL
  - 开发环境的完整体验
- **适用场景**: 集成测试、微服务开发

### 6. 🏭 生产模式 (production)
- **服务**: 与完整模式相同的所有服务
- **特点**: 
  - 生产级配置
  - 性能优化
  - 日志级别：info
  - 必须使用PostgreSQL
  - 包含基础设施服务（数据库、缓存）
- **适用场景**: 生产部署、性能测试

## 服务依赖关系

```
基础设施层:
├── PostgreSQL (5432) - 数据库
└── Redis (6379) - 缓存（可选）

服务层:
├── API网关 (8000) - 统一入口
├── Go主后端 (8080) - 核心服务
├── 写信服务 (8001) - Python/FastAPI
├── 信使服务 (8002) - Go微服务
├── 管理服务 (8003) - Java/Spring Boot
└── OCR服务 (8004) - Python/Flask

前端层:
├── 主前端 (3000) - Next.js
└── 管理后台 (3001) - Vue 3
```

## 快速启动命令

```bash
# 开发模式（默认）
./startup/quick-start.sh

# 演示模式（自动打开浏览器）
./startup/quick-start.sh demo --auto-open

# 完整微服务模式
./startup/quick-start.sh complete

# 生产模式（所有服务）
./startup/quick-start.sh production

# Mock模式（前端开发）
./startup/quick-start.sh mock

# 带详细日志
./startup/quick-start.sh development --verbose

# 预览模式（不实际启动）
./startup/quick-start.sh production --dry-run
```

## 环境变量配置

每个模式都会自动设置相应的环境变量：

| 模式 | NODE_ENV | DEBUG | LOG_LEVEL | DATABASE_TYPE |
|------|----------|-------|-----------|---------------|
| development | development | true | debug | sqlite |
| simple | development | true | info | sqlite |
| demo | demo | false | info | sqlite |
| mock | development | true | debug | mock |
| complete | development | true | debug | postgres |
| production | production | false | info | postgres |

## 选择建议

1. **新手入门**: 使用 `demo` 模式
2. **日常开发**: 使用 `development` 模式
3. **前端开发**: 使用 `mock` 模式
4. **集成测试**: 使用 `complete` 模式
5. **部署验证**: 使用 `production` 模式

## 常见问题

### 端口被占用
```bash
# 查看占用端口的进程
lsof -i :8080

# 停止所有服务
./startup/stop-all.sh --force
```

### PostgreSQL未启动
```bash
# 使用Docker启动
docker-compose up -d postgres

# 或者启动本地PostgreSQL
brew services start postgresql
```

### 服务编译失败
```bash
# Go服务
cd services/gateway && go mod tidy && go build -o bin/gateway cmd/main.go

# Python服务
cd services/write-service && python3 -m venv venv && source venv/bin/activate && pip install -r requirements.txt
```

### Docker未找到
```bash
# 设置Docker路径（macOS）
./startup/setup-docker-path.sh
```