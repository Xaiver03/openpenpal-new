# OpenPenPal 信使任务调度系统

> Agent #3 开发的信使任务调度系统 - 负责信使管理、任务分配、扫码更新和地理位置匹配

## 🎯 系统概述

信使任务调度系统是 OpenPenPal 的核心模块之一，负责：

- 🚀 **信使申请与管理** - 信使注册、审核、状态管理
- 📋 **智能任务分配** - 基于地理位置和信使评分的自动分配算法
- 📱 **扫码状态更新** - 信使扫码更新投递状态
- 📊 **实时数据同步** - Redis队列和WebSocket实时通知
- 🌍 **地理位置服务** - 距离计算、区域匹配、路径优化

## 🏗️ 技术架构

### 核心技术栈
- **后端**: Go 1.21 + Gin Framework
- **数据库**: PostgreSQL 15 + GORM
- **缓存**: Redis 7 + 任务队列
- **通信**: WebSocket + gRPC
- **部署**: Docker + Docker Compose

### 系统架构图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   WebSocket     │
│   (Next.js)     │◄──►│   (Nginx)       │◄──►│   Manager       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Courier       │    │   Task          │    │   Assignment    │
│   Service       │◄──►│   Service       │◄──►│   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   Redis Queue   │    │   Location      │
│   Database      │    │   + PubSub      │    │   Service       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📡 API 接口

### 信使相关接口
```bash
# 申请成为信使
POST /api/courier/apply
Content-Type: application/json
Authorization: Bearer <token>

{
  "zone": "北京大学",
  "phone": "138****5678",
  "id_card": "110101********1234",
  "experience": "有快递配送经验"
}

# 获取信使信息
GET /api/courier/info
Authorization: Bearer <token>

# 获取信使统计
GET /api/courier/stats/{courier_id}
Authorization: Bearer <token>
```

### 任务相关接口
```bash
# 获取可用任务
GET /api/courier/tasks?zone=北京大学&status=available&limit=10
Authorization: Bearer <token>

# 接受任务
PUT /api/courier/tasks/{task_id}/accept
Content-Type: application/json
Authorization: Bearer <token>

{
  "estimated_time": "2小时",
  "note": "预计下午完成投递"
}

# 获取任务详情
GET /api/courier/tasks/{task_id}
Authorization: Bearer <token>
```

### 扫码相关接口
```bash
# 扫码更新状态
POST /api/courier/scan/{letter_code}
Content-Type: application/json
Authorization: Bearer <token>

{
  "action": "collected",
  "location": "北京大学宿舍楼下信箱",
  "latitude": 39.9912,
  "longitude": 116.3064,
  "note": "已从发件人处收取",
  "photo_url": "https://example.com/photo.jpg"
}

# 获取扫码历史
GET /api/courier/scan/{letter_code}/history
Authorization: Bearer <token>
```

## 🔧 本地开发

### 环境要求
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### 快速开始

1. **克隆代码**
```bash
cd services/courier-service
```

2. **配置环境**
```bash
cp .env.example .env
# 编辑 .env 文件配置数据库等信息
```

3. **安装依赖**
```bash
go mod download
```

4. **启动数据库服务**
```bash
docker-compose up -d postgres redis
```

5. **运行服务**
```bash
go run cmd/main.go
```

服务将在 `http://localhost:8002` 启动

### 开发工具

```bash
# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 运行测试
go test ./...

# 生成 API 文档
swag init -g cmd/main.go
```

## 🐳 Docker 部署

### 使用部署脚本（推荐）

```bash
# 构建并启动所有服务
./deploy.sh start

# 查看服务状态
./deploy.sh status

# 查看日志
./deploy.sh logs courier-service

# 停止服务
./deploy.sh stop
```

### 手动 Docker 命令

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f courier-service

# 停止服务
docker-compose down
```

### 生产环境部署

```bash
# 使用生产配置
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 检查服务健康状态
curl http://localhost:8002/health
```

## 📊 核心功能

### 1. 智能任务分配算法

```go
// 基于多因素评分的任务分配
type CourierScore struct {
    Courier      models.Courier
    Score        float64    // 综合评分
    Distance     float64    // 距离评分
    CurrentTasks int        // 当前任务数
}

// 评分计算公式
totalScore = ratingScore(20%) + distanceScore(50%) + workloadScore(30%)
```

### 2. 地理位置匹配

```go
// Haversine 公式计算距离
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371 // 地球半径 (km)
    // ... 距离计算逻辑
}

// 查找附近信使
func FindNearbyTasks(courierLat, courierLng, radiusKm float64) []Task
```

### 3. Redis 任务队列

```bash
# 队列优先级
tasks:express   # 特急任务
tasks:urgent    # 紧急任务  
tasks:normal    # 普通任务
tasks:assignment # 自动分配队列
notifications   # 通知队列
```

### 4. 实时 WebSocket 通知

```json
{
  "type": "COURIER_TASK_UPDATE",
  "data": {
    "task_id": "T20231120001",
    "status": "delivered",
    "courier_id": "courier1"
  },
  "timestamp": "2023-11-20T12:00:00Z"
}
```

## 📈 监控与运维

### 健康检查

```bash
# 服务健康状态
curl http://localhost:8002/health

# 队列状态监控
curl http://localhost:8002/admin/queue/stats
```

### 日志管理

```bash
# 查看应用日志
docker-compose logs courier-service

# 查看数据库日志
docker-compose logs postgres

# 查看 Redis 日志
docker-compose logs redis
```

### 性能监控

```bash
# 资源使用情况
docker stats

# 数据库连接数
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Redis 内存使用
docker-compose exec redis redis-cli info memory
```

## 🔒 安全配置

### JWT 认证
- 所有 API 接口都需要有效的 JWT token
- Token 包含用户ID、角色等信息
- 支持 token 过期和刷新机制

### 数据安全
- 数据库连接使用 SSL 加密
- 敏感信息（身份证号、手机号）进行脱敏处理
- Redis 连接可配置密码认证

### 权限控制
```go
// 角色权限映射
const (
    RoleUser      = "user"       // 普通用户
    RoleCourier   = "courier"    // 信使
    RoleAdmin     = "admin"      // 管理员
    RoleSuperAdmin = "super_admin" // 超级管理员
)
```

## 🧪 测试

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/services

# 生成测试覆盖率报告
go test -cover ./...
```

### 集成测试
```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行集成测试
go test -tags=integration ./tests/...
```

### API 测试
```bash
# 使用 curl 测试
curl -X POST http://localhost:8002/api/courier/apply \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"zone":"北京大学","phone":"13800138000","id_card":"110101199001011234"}'
```

## 📚 开发文档

### 项目结构
```
courier-service/
├── cmd/                    # 应用入口
├── internal/
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── models/           # 数据模型
│   ├── services/         # 业务逻辑
│   └── utils/            # 工具函数
├── docker-compose.yml    # Docker 编排
├── Dockerfile           # Docker 构建
├── deploy.sh           # 部署脚本
└── README.md          # 项目文档
```

### 扩展指南

#### 添加新的 API 接口
1. 在 `internal/models/` 定义数据模型
2. 在 `internal/services/` 实现业务逻辑
3. 在 `internal/handlers/` 添加HTTP处理器
4. 在 `cmd/main.go` 注册路由

#### 添加新的队列消费者
1. 在 `internal/services/queue.go` 添加队列类型
2. 实现消费者逻辑
3. 在 `cmd/main.go` 启动消费者协程

## 🚨 故障排除

### 常见问题

**1. 数据库连接失败**
```bash
# 检查数据库状态
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres

# 重启数据库
docker-compose restart postgres
```

**2. Redis 连接失败**
```bash
# 检查 Redis 状态
docker-compose ps redis

# 测试 Redis 连接
docker-compose exec redis redis-cli ping
```

**3. 服务启动失败**
```bash
# 查看详细日志
docker-compose logs courier-service

# 检查端口占用
lsof -i :8002

# 重新构建镜像
docker-compose build --no-cache courier-service
```

**4. 任务分配不工作**
```bash
# 检查队列消费者状态
curl http://localhost:8002/admin/queue/stats

# 查看队列长度
docker-compose exec redis redis-cli llen tasks:normal
```

## 📞 技术支持

### 开发团队
- **Agent #3**: 信使任务调度系统架构师
- **技术栈**: Go + PostgreSQL + Redis + Docker

### 相关文档
- [OpenPenPal 项目总览](../../README.md)
- [多Agent协同开发指南](../../MULTI_AGENT_COORDINATION.md)
- [API规范文档](../../docs/api/UNIFIED_API_SPECIFICATION.md)

### 问题反馈
如遇到问题，请查看：
1. 本文档的故障排除部分
2. 项目 Issues 页面
3. 联系开发团队

---

*信使系统，连接校园每一个角落* 🚀