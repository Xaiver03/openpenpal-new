# OpenPenPal API Gateway

> 统一API网关 - OpenPenPal微服务架构的统一入口和流量管理中心

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Gin](https://img.shields.io/badge/Gin-1.9+-green.svg)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-ready-blue.svg)](https://docker.com/)
[![Prometheus](https://img.shields.io/badge/Prometheus-monitoring-orange.svg)](https://prometheus.io/)

## 🎯 系统概述

API Gateway 是 OpenPenPal 微服务架构的统一入口，负责：

- 🌐 **统一路由** - 所有外部请求的唯一入口
- 🔐 **认证授权** - JWT认证和权限控制  
- ⚖️ **负载均衡** - 多实例服务的智能分发
- 🛡️ **安全防护** - 限流、熔断、CORS等安全机制
- 📊 **监控日志** - 完整的请求追踪和性能监控
- 🔄 **服务发现** - 动态发现和健康检查

## 🏗️ 系统架构

### 微服务路由表

| 路由前缀 | 目标服务 | 端口 | 认证要求 | 限流(req/min) |
|---------|---------|------|----------|---------------|
| `/api/v1/auth/*` | main-backend | 8080 | ❌ | 60 |
| `/api/v1/users/*` | main-backend | 8080 | ✅ | 120 |
| `/api/v1/letters/*` | write-service | 8001 | ✅ | 100 |
| `/api/v1/courier/*` | courier-service | 8002 | ✅ | 80 |
| `/api/v1/ocr/*` | ocr-service | 8004 | ✅ | 20 |
| `/admin/*` | admin-service | 8003 | 👨‍💼 | 30 |

### 架构流程图

```
                     ┌─────────────────┐
                     │   外部客户端     │
                     │ (Web/Mobile)    │
                     └────────┬────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │  API Gateway    │
                     │   (Port 8000)   │
                     │                 │
                     │ • 认证授权      │
                     │ • 限流熔断      │
                     │ • 负载均衡      │
                     │ • 监控日志      │
                     └────────┬────────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
            ▼                 ▼                 ▼
   ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
   │ main-backend  │ │ write-service │ │courier-service│
   │   (Port 8080) │ │  (Port 8001)  │ │ (Port 8002)   │
   └───────────────┘ └───────────────┘ └───────────────┘
            │                 │                 │
            └─────────────────┼─────────────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │   PostgreSQL    │
                     │     Redis       │
                     └─────────────────┘
```

## 🚀 快速开始

### 环境要求
- **Go**: 1.21+
- **Docker**: 20+
- **Redis**: 7+

### 本地开发

1. **克隆项目**
```bash
cd services/gateway
```

2. **配置环境**
```bash
cp .env.example .env
# 编辑 .env 文件设置服务地址
```

3. **安装依赖**
```bash
make deps
```

4. **启动Redis**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

5. **运行网关**
```bash
make run
# 或热重载开发模式
make dev
```

### Docker部署

```bash
# 构建并启动
make docker-build
make docker-run

# 查看服务状态
make status

# 查看日志
make docker-logs
```

### 生产环境

```bash
# 启动生产环境（包含Nginx）
make production

# 启动监控服务
make monitoring
```

## 📡 API接口

### 健康检查和监控

```bash
# 网关健康检查
GET /health

# 监控指标 (Prometheus格式)
GET /metrics  

# 版本信息
GET /version

# 网关信息
GET /info
```

### 认证相关 (无需认证)

```bash
# 用户注册
POST /api/v1/auth/register

# 用户登录
POST /api/v1/auth/login

# 刷新Token
POST /api/v1/auth/refresh
```

### 信件服务 (需要认证)

```bash
# 创建信件
POST /api/v1/letters

# 获取信件列表
GET /api/v1/letters

# 获取信件详情
GET /api/v1/letters/{id}

# 生成信件二维码
POST /api/v1/letters/{id}/generate-code
```

### 信使服务 (需要认证)

```bash
# 申请成为信使
POST /api/v1/courier/apply

# 获取可用任务 (需要信使权限)
GET /api/v1/courier/tasks

# 接受任务 (需要信使权限)
PUT /api/v1/courier/tasks/{id}/accept

# 扫码更新状态 (需要信使权限)
POST /api/v1/courier/scan/{code}
```

### 管理接口 (需要管理员权限)

```bash
# 网关状态
GET /admin/gateway/status

# 服务状态
GET /admin/gateway/services

# 重新加载配置
POST /admin/gateway/reload

# 服务健康检查
GET /admin/health
GET /admin/health/{service}
```

## 🔐 认证与授权

### JWT认证流程

1. 客户端通过 `/api/v1/auth/login` 获取JWT Token
2. 后续请求在Header中携带: `Authorization: Bearer <token>`
3. 网关验证Token并提取用户信息
4. 将用户信息通过Header传递给后端服务

### 权限等级

- **公开接口**: 无需认证 (如登录、注册)
- **用户接口**: 需要有效JWT Token
- **信使接口**: 需要信使角色权限
- **管理接口**: 需要管理员权限

### 用户信息传递

网关验证JWT后，会在请求头中添加用户信息：

```http
X-User-ID: user123
X-Username: alice
X-User-Role: courier
X-Trace-ID: gw-1234567890-abcd1234
```

## ⚖️ 负载均衡

### 服务发现机制

- **健康检查**: 每30秒检查服务实例健康状态
- **权重分配**: 基于配置权重进行负载分发
- **故障转移**: 自动剔除不健康实例
- **服务恢复**: 健康实例自动重新加入

### 负载策略

```go
// 加权随机算法
func selectInstanceByWeight(instances []*ServiceInstance) *ServiceInstance {
    // 按权重随机选择健康实例
    // 权重越高，被选中概率越大
}
```

## 🛡️ 安全机制

### 限流策略

| 服务类型 | 限制 | 说明 |
|---------|------|------|
| 认证接口 | 60/min | 防止暴力破解 |
| 用户接口 | 120/min | 正常使用频率 |
| 信件接口 | 100/min | 写信操作限制 |
| 信使接口 | 80/min | 信使操作限制 |
| OCR接口 | 20/min | 资源密集型操作 |
| 管理接口 | 30/min | 管理操作限制 |

### 安全头设置

```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000
Content-Security-Policy: default-src 'self'
```

### CORS配置

```go
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Authorization
```

## 📊 监控与日志

### Prometheus指标

```bash
# HTTP请求总数
gateway_http_requests_total{method, path, status}

# 请求耗时分布
gateway_http_request_duration_seconds{method, path, status}

# 服务健康状态
gateway_service_health{service, instance}

# 代理请求计数
gateway_proxy_requests_total{service, status}

# 限流触发次数
gateway_rate_limit_triggered_total{client_type}
```

### 日志格式

```json
{
  "timestamp": "2025-07-20T12:00:00Z",
  "level": "info",
  "message": "Gateway request",
  "method": "POST",
  "path": "/api/v1/letters",
  "status": 201,
  "duration": "45ms",
  "client_ip": "192.168.1.100",
  "user_id": "user123",
  "trace_id": "gw-1234567890-abcd1234"
}
```

### 监控面板

启动监控服务后可访问：

- **Prometheus**: http://localhost:9090 - 指标采集
- **Grafana**: http://localhost:3000 - 可视化面板 (admin/admin)

## 🔧 配置管理

### 环境变量

```bash
# 基础配置
PORT=8000                    # 网关端口
ENVIRONMENT=development      # 环境类型
LOG_LEVEL=info              # 日志级别
JWT_SECRET=your-secret      # JWT密钥

# 服务地址
MAIN_BACKEND_HOSTS=http://localhost:8080
WRITE_SERVICE_HOSTS=http://localhost:8001
COURIER_SERVICE_HOSTS=http://localhost:8002

# 限流配置  
RATE_LIMIT_ENABLED=true     # 启用限流
RATE_LIMIT_DEFAULT=100      # 默认限制
RATE_LIMIT_BURST=10         # 突发允许

# 超时配置
PROXY_TIMEOUT=30            # 代理超时(秒)
CONNECT_TIMEOUT=5           # 连接超时(秒)
```

### 服务配置

```go
// 支持多实例负载均衡
WRITE_SERVICE_HOSTS=http://write1:8001,http://write2:8001,http://write3:8001

// 服务权重配置
WRITE_SERVICE_WEIGHT=10
COURIER_SERVICE_WEIGHT=5
```

## 🐳 Docker部署

### 服务编排

```yaml
services:
  api-gateway:
    build: .
    ports:
      - "8000:8000"  # API端口
      - "9000:9000"  # 监控端口
    environment:
      - MAIN_BACKEND_HOSTS=http://main-backend:8080
      - WRITE_SERVICE_HOSTS=http://write-service:8001
    depends_on:
      - redis
    networks:
      - openpenpal-network
```

### 部署命令

```bash
# 开发环境
docker-compose up -d

# 生产环境（含Nginx）
docker-compose --profile production up -d

# 监控环境（含Prometheus + Grafana）
docker-compose --profile monitoring up -d
```

## 🔄 开发工具

### Makefile命令

```bash
# 开发相关
make deps          # 安装依赖
make fmt           # 格式化代码  
make lint          # 代码检查
make test          # 运行测试
make dev           # 热重载开发

# 构建部署
make build         # 构建二进制
make docker-build  # 构建镜像
make docker-run    # 启动容器

# 监控运维
make health        # 健康检查
make metrics       # 查看指标
make monitoring    # 启动监控
make status        # 查看状态
```

### 开发流程

```bash
# 快速开发流程
make quick

# 完整开发流程  
make full

# 性能测试
make bench
```

## 🧪 测试

### 健康检查测试

```bash
curl http://localhost:8000/health
```

### 认证流程测试

```bash
# 1. 用户登录
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}' \
  | jq -r '.data.token')

# 2. 访问受保护接口
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/api/v1/letters
```

### 限流测试

```bash
# 快速发送多个请求测试限流
for i in {1..120}; do
  curl -s http://localhost:8000/api/v1/auth/login > /dev/null &
done
```

## 📈 性能优化

### 优化建议

1. **连接池优化**
```go
Transport: &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     30 * time.Second,
}
```

2. **缓存策略**
- Redis缓存JWT验证结果
- 服务发现结果缓存
- 限流计数器缓存

3. **监控告警**
- 响应时间 > 1s 告警
- 错误率 > 5% 告警  
- 服务不可用告警

### 压力测试

```bash
# 使用wrk进行压力测试
wrk -t12 -c400 -d30s http://localhost:8000/health

# 结果示例：
# Requests/sec: 45000
# Transfer/sec: 12.3MB
```

## 🚨 故障排除

### 常见问题

**1. 服务无法访问**
```bash
# 检查网关状态
make health

# 检查服务发现
curl http://localhost:8000/admin/gateway/services
```

**2. 认证失败**
```bash
# 检查JWT密钥配置
echo $JWT_SECRET

# 验证Token格式
curl -H "Authorization: Bearer invalid-token" \
  http://localhost:8000/api/v1/letters
```

**3. 限流触发**
```bash
# 检查限流配置
curl http://localhost:8000/admin/gateway/status

# 查看限流指标
curl http://localhost:8000/metrics | grep rate_limit
```

### 日志分析

```bash
# 查看网关日志
make docker-logs

# 过滤错误日志
docker logs openpenpal-api-gateway 2>&1 | grep ERROR

# 实时监控
docker logs -f openpenpal-api-gateway
```

## 🔮 后续规划

### v1.1 增强功能
- [ ] 熔断器机制
- [ ] 请求缓存
- [ ] API版本管理
- [ ] GraphQL支持

### v1.2 高级功能
- [ ] 服务网格集成
- [ ] 分布式追踪
- [ ] 智能路由
- [ ] A/B测试支持

### v1.3 企业功能
- [ ] 多租户支持
- [ ] API计费
- [ ] 高级安全策略
- [ ] 自动扩缩容

## 📞 技术支持

### 相关文档
- [OpenPenPal 项目总览](../../README.md)
- [信使服务文档](../courier-service/README.md)
- [写信服务文档](../write-service/README.md)

### 监控面板
- **网关监控**: http://localhost:8000/admin/gateway/status
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

---

*统一入口，安全可靠，性能卓越* 🚀