# OpenPenPal Write Service

OpenPenPal 信件创建和管理服务 - 负责信件的创建、编号生成、状态管理和数据持久化。

## 🎯 服务概览

- **服务名称**: write-service
- **端口**: 8001
- **技术栈**: Python + FastAPI + PostgreSQL
- **责任**: 信件创建、编号生成、状态管理、WebSocket事件推送

## 🚀 快速开始

### 方式一：使用启动脚本（推荐）

```bash
# 克隆项目并进入目录
cd services/write-service

# 运行启动脚本
./start.sh
```

### 方式二：手动启动

```bash
# 1. 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 2. 安装依赖
pip install -r requirements.txt

# 3. 配置环境变量
cp .env.example .env
# 编辑 .env 文件

# 4. 启动服务
uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload
```

### 方式三：Docker 启动

```bash
# 开发环境
docker-compose -f docker-compose.dev.yml up

# 生产环境
docker-compose up -d
```

## 📡 API 接口

### 服务健康检查
```http
GET /health
```

### 信件管理

#### 1. 创建信件
```http
POST /api/letters
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "title": "给朋友的问候信",
  "content": "信件正文内容...",
  "receiver_hint": "北大宿舍楼，李同学", 
  "anonymous": false,
  "priority": "normal",
  "delivery_instructions": "请投递到宿舍管理员处"
}
```

#### 2. 获取信件详情
```http
GET /api/letters/{letter_id}
Authorization: Bearer <jwt_token>
```

#### 3. 更新信件状态
```http
PUT /api/letters/{letter_id}/status
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "status": "collected",
  "location": "北京大学宿舍楼下",
  "note": "已被信使收取"
}
```

#### 4. 获取用户信件列表
```http
GET /api/letters/user/{user_id}?status=all&page=1&limit=10
Authorization: Bearer <jwt_token>
```

#### 5. 通过编号读取信件（公开接口）
```http
GET /api/letters/read/{code}
```

## 📊 服务地址

- **API 服务**: http://localhost:8001
- **健康检查**: http://localhost:8001/health
- **API 文档**: http://localhost:8001/docs
- **ReDoc 文档**: http://localhost:8001/redoc

## 🗄️ 数据模型

### Letter 信件模型

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 信件编号 (OP + 10位随机字符) |
| title | string | 信件标题 |
| content | text | 信件内容 |
| sender_id | string | 发送者用户ID |
| sender_nickname | string | 发送者昵称 |
| receiver_hint | string | 接收者提示信息 |
| status | enum | 信件状态 |
| priority | enum | 优先级 |
| anonymous | boolean | 是否匿名 |
| delivery_instructions | text | 投递说明 |
| read_count | integer | 阅读次数 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |

### 状态流转

```
draft → generated → collected → in_transit → delivered/failed
```

- **draft**: 草稿
- **generated**: 已生成二维码
- **collected**: 已收取
- **in_transit**: 投递中
- **delivered**: 已投递
- **failed**: 投递失败

## 🔧 配置说明

### 环境变量 (.env)

```bash
# 数据库配置
DATABASE_URL=postgresql://user:password@localhost:5432/openpenpal

# JWT配置
JWT_SECRET=your-super-secret-jwt-key

# Redis配置 (可选)
REDIS_URL=redis://localhost:6379/0

# WebSocket配置
WEBSOCKET_URL=ws://localhost:8080/ws

# 前端地址
FRONTEND_URL=http://localhost:3000
```

## 🎮 WebSocket 事件

服务会向主WebSocket服务推送以下事件：

### 1. 信件创建事件
```json
{
  "type": "LETTER_CREATED",
  "data": {
    "letter_id": "OP1K2L3M4N5O",
    "action": "created",
    "timestamp": "2025-07-20T12:00:00Z"
  },
  "target_user": "user123"
}
```

### 2. 状态更新事件
```json
{
  "type": "LETTER_STATUS_UPDATE",
  "data": {
    "letter_id": "OP1K2L3M4N5O",
    "status": "collected",
    "timestamp": "2025-07-20T14:00:00Z"
  },
  "target_user": "user123"
}
```

### 3. 阅读事件
```json
{
  "type": "LETTER_READ",
  "data": {
    "letter_id": "OP1K2L3M4N5O",
    "action": "read",
    "read_count": 3,
    "timestamp": "2025-07-20T16:00:00Z"
  },
  "target_user": "user123"
}
```

## 🧪 测试

### API 测试示例

使用 curl 测试 API：

```bash
# 健康检查
curl http://localhost:8001/health

# 创建信件（需要JWT token）
curl -X POST http://localhost:8001/api/letters \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试信件",
    "content": "这是一封测试信件",
    "receiver_hint": "测试地址"
  }'

# 获取信件详情
curl -X GET http://localhost:8001/api/letters/OP1234567890 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 通过编号读取信件（无需认证）
curl http://localhost:8001/api/letters/read/OP1234567890
```

### 单元测试

```bash
# 安装测试依赖
pip install pytest pytest-asyncio httpx

# 运行测试
pytest tests/
```

## 🐳 Docker 部署

### 开发环境

```bash
# 启动所有服务（包括数据库）
docker-compose -f docker-compose.dev.yml up

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f write-service

# 停止服务
docker-compose -f docker-compose.dev.yml down
```

### 生产环境

```bash
# 启动生产环境
docker-compose up -d

# 扩容服务
docker-compose up -d --scale write-service=3
```

## 🔍 故障排除

### 常见问题

#### 1. 数据库连接失败
```bash
# 检查数据库配置
echo $DATABASE_URL

# 测试数据库连接
python3 -c "
from app.core.database import engine
with engine.connect() as conn:
    print('Database connected successfully')
"
```

#### 2. JWT 认证失败
- 检查 JWT_SECRET 配置
- 确认 token 格式正确
- 验证 token 未过期

#### 3. WebSocket 连接失败
- 检查 WEBSOCKET_URL 配置
- 确认主 WebSocket 服务运行正常
- 查看网络连接

#### 4. 端口被占用
```bash
# 查看端口占用
lsof -i :8001

# 杀死进程
kill -9 <PID>
```

## 📈 性能监控

### 健康检查指标

- **响应时间**: < 200ms
- **内存使用**: < 512MB
- **CPU 使用**: < 50%
- **数据库连接**: 正常

### 日志级别

- **INFO**: 正常操作日志
- **WARNING**: 可恢复的错误
- **ERROR**: 需要关注的错误
- **DEBUG**: 开发调试信息

## 🔄 版本更新

### v1.0.0 (当前版本)
- ✅ 完整的信件 CRUD 操作
- ✅ JWT 认证集成
- ✅ WebSocket 事件推送
- ✅ Docker 容器化部署
- ✅ 完整的 API 文档

### 未来计划
- [ ] 信件模板功能
- [ ] 批量操作接口
- [ ] 图片附件支持
- [ ] 信件加密功能

## 📞 技术支持

遇到问题？查看以下资源：

- 📖 [API 文档](http://localhost:8001/docs)
- 🔧 [项目 README](../../README.md)
- 🏠 [Agent 任务卡片](../../agent-tasks/AGENT-2-WRITE-SERVICE.md)

---

*OpenPenPal Write Service - 让每一封信都有温度* ✨