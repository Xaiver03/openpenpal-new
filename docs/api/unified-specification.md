# OpenPenPal 统一API规范 v2.0

> **Version**: 2.0  
> **Last Updated**: 2025-07-22  
> **Status**: 生产就绪 (97% 完成度)  
> **目标**: 基于实际实现情况更新的完整API规范文档

## 🎯 规范原则

1. **RESTful设计**: 遵循REST架构风格
2. **统一响应格式**: 所有API使用相同的响应结构
3. **标准HTTP状态码**: 合理使用HTTP状态码
4. **JWT认证**: 统一的身份验证机制
5. **版本控制**: API版本管理策略
6. **实时通信**: WebSocket事件推送

## 📡 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 具体数据内容
  },
  "timestamp": "2025-07-22T12:00:00Z"
}
```

### 错误响应
```json
{
  "code": 1,
  "msg": "参数错误",
  "data": null,
  "error": {
    "details": "字段 'title' 不能为空",
    "field": "title",
    "type": "validation_error"
  },
  "timestamp": "2025-07-22T12:00:00Z"
}
```

### 分页响应
```json
{
  "code": 0,
  "msg": "success", 
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "pages": 10,
      "has_next": true,
      "has_prev": false
    }
  },
  "timestamp": "2025-07-22T12:00:00Z"
}
```

## 🔢 标准状态码约定

| **Code** | **含义** | **HTTP Status** | **使用场景** |
|----------|----------|------------------|-------------|
| 0 | 成功 | 200/201 | 操作成功完成 |
| 1 | 参数错误 | 400 | 请求参数不合法 |
| 2 | 无权限 | 403 | 权限不足 |
| 3 | 数据不存在 | 404 | 资源未找到 |
| 4 | 业务逻辑错误 | 422 | 业务规则冲突 |
| 5 | 频率限制 | 429 | 请求过于频繁 |
| 500 | 服务内部异常 | 500 | 服务器内部错误 |

## 🌐 服务架构与端口分配

### 服务拓扑
```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (Port 8000)                  │
├─────────────────────────────────────────────────────────────┤
│  认证/限流/负载均衡/监控                                     │
└─────────────┬───────────────────────────────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
    ▼         ▼         ▼
┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐
│前端   │ │写信   │ │信使   │ │管理   │ │OCR    │ │认证   │
│3000   │ │8001   │ │8002   │ │8003   │ │8004   │ │8080   │
└───────┘ └───────┘ └───────┘ └───────┘ └───────┘ └───────┘
     │         │         │         │         │         │
     └─────────┼─────────┼─────────┼─────────┼─────────┘
               │         │         │         │
            ┌──▼─────────▼─────────▼─────────▼──┐
            │        PostgreSQL (5432)        │
            │        Redis (6379)              │
            └─────────────────────────────────┘
```

### URL路径规范
```
/api/auth/*          - 认证服务 (Port 8080)
/api/letters/*       - 写信服务 (Port 8001)  
/api/courier/*       - 信使服务 (Port 8002)
/api/admin/*         - 管理后台 (Port 8003)
/api/ocr/*          - OCR服务 (Port 8004)
/api/signal-codes/* - 信号编码 (Port 8002)
```

## 🔐 认证规范

### JWT Token格式
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Token Payload结构
```json
{
  "user_id": "user_12345",
  "username": "xiaoming",
  "role": "user", 
  "school_code": "BJFU",
  "permissions": ["read_letters", "write_letters"],
  "iat": 1642684800,
  "exp": 1642771200
}
```

### 权限级别
- **user**: 普通用户
- **courier**: 信使 (1-4级)
- **admin**: 管理员
- **super_admin**: 超级管理员

## 📝 API接口详细规范

### 1. 认证服务 (Port 8080)

#### 用户认证
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "xiaoming",
  "password": "password123"
}

Response:
{
  "code": 0,
  "msg": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "user_12345",
      "username": "xiaoming",
      "role": "user",
      "school_code": "BJFU"
    }
  }
}
```

#### 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "xiaoming",
  "password": "password123",
  "email": "xiaoming@example.com",
  "school_code": "BJFU",
  "student_id": "20210001"
}
```

#### Token刷新
```http
POST /api/auth/refresh
Authorization: Bearer <refresh_token>
```

### 2. 写信服务 (Port 8001)

#### 信件管理
```http
# 创建信件
POST /api/letters
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "title": "给朋友的问候信",
  "content": "信件内容...",
  "receiver_hint": "北大宿舍楼，李同学",
  "delivery_method": "courier",
  "priority": "normal"
}

Response:
{
  "code": 0,
  "msg": "信件创建成功",
  "data": {
    "letter_id": "OP9691NL7ZBOWC",
    "qr_code_url": "https://example.com/qr/OP9691NL7ZBOWC.png",
    "status": "generated",
    "estimated_delivery": "2025-07-23T15:00:00Z"
  }
}
```

```http
# 获取信件列表
GET /api/letters?status=delivered&page=1&limit=10
Authorization: Bearer <jwt_token>

# 获取信件详情
GET /api/letters/OP9691NL7ZBOWC
Authorization: Bearer <jwt_token>

# 更新信件状态
PUT /api/letters/OP9691NL7ZBOWC/status
Content-Type: application/json

{
  "status": "collected",
  "location": "北京大学宿舍楼下",
  "note": "已被信使收取",
  "photo_url": "https://example.com/photo.jpg"
}
```

#### 博物馆功能
```http
# 获取博物馆信件
GET /api/letters/museum?category=love&page=1&limit=20

# 提交到博物馆
POST /api/letters/OP9691NL7ZBOWC/submit-to-museum
Content-Type: application/json

{
  "category": "friendship",
  "tags": ["校园", "友谊"],
  "is_anonymous": false
}
```

#### 广场功能
```http
# 获取广场信件
GET /api/letters/plaza?sort=latest&page=1&limit=20

# 点赞信件
POST /api/letters/OP9691NL7ZBOWC/like

# 评论信件
POST /api/letters/OP9691NL7ZBOWC/comments
Content-Type: application/json

{
  "content": "很棒的信件!",
  "is_anonymous": false
}
```

### 3. 信使服务 (Port 8002)

#### 信使申请与管理
```http
# 申请成为信使
POST /api/courier/apply
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "zone": "北京大学",
  "phone": "138****5678",
  "id_card": "110101********1234",
  "experience": "有快递配送经验"
}

# 获取信使信息
GET /api/courier/info
Authorization: Bearer <jwt_token>

# 管理员审核信使
PUT /api/courier/admin/approve/{courier_id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "approved": true,
  "level": 1,
  "zone_assignment": "area_01",
  "note": "审核通过"
}
```

#### 任务管理
```http
# 获取可用任务
GET /api/courier/tasks?zone=北京大学&status=available&limit=10
Authorization: Bearer <courier_token>

Response:
{
  "code": 0,
  "msg": "success",
  "data": {
    "tasks": [
      {
        "task_id": "T001",
        "letter_id": "OP9691NL7ZBOWC",
        "pickup_location": "北大宿舍楼32栋",
        "delivery_location": "清华大学图书馆",
        "priority": "urgent",
        "estimated_distance": "15km",
        "reward": 8.00,
        "created_at": "2025-07-22T10:00:00Z"
      }
    ],
    "total": 5
  }
}

# 接受任务
PUT /api/courier/tasks/T001/accept
Content-Type: application/json

{
  "estimated_time": "2小时",
  "note": "预计下午完成投递"
}

# 扫码更新状态
POST /api/courier/scan/OP9691NL7ZBOWC
Content-Type: application/json

{
  "action": "collected",
  "location": "北京大学宿舍楼下信箱",
  "note": "已从发件人处收取",
  "photo_url": "https://example.com/photo.jpg"
}
```

#### 层级管理 (4级信使系统)
```http
# 获取下级信使列表
GET /api/courier/subordinates
Authorization: Bearer <courier_token>

# 分配下级任务
POST /api/courier/assign-task
Content-Type: application/json

{
  "task_id": "T001",
  "subordinate_id": "courier_456",
  "priority": "high",
  "deadline": "2025-07-22T18:00:00Z"
}

# 上报异常
POST /api/courier/report-exception
Content-Type: application/json

{
  "task_id": "T001",
  "exception_type": "delivery_failed",
  "description": "收件人不在宿舍",
  "suggested_action": "重新投递"
}
```

#### 积分与排行榜
```http
# 获取个人积分
GET /api/courier/points
Authorization: Bearer <courier_token>

Response:
{
  "code": 0,
  "data": {
    "current_points": 1250,
    "level": "bronze",
    "level_progress": 0.65,
    "next_level_points": 2000,
    "badges": ["新手", "百单达人"],
    "monthly_ranking": 15
  }
}

# 获取排行榜
GET /api/courier/leaderboard?type=school&period=monthly
GET /api/courier/leaderboard?type=national&period=all_time

# 积分兑换
POST /api/courier/points/exchange
Content-Type: application/json

{
  "item_id": "coupon_001",
  "points_cost": 500,
  "quantity": 1
}
```

#### 信使等级权限
```http
# 获取权限列表
GET /api/courier/level/permissions
Authorization: Bearer <courier_token>

# 申请等级升级
POST /api/courier/level/upgrade/request
Content-Type: application/json

{
  "target_level": 2,
  "reason": "任务完成率90%以上，积分达标",
  "supporting_documents": ["performance_report.pdf"]
}
```

### 4. 信号编码服务 (Port 8002)

#### 编码管理
```http
# 生成编码批次
POST /api/signal-codes/batch
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "school_id": "school_001",
  "area_id": "area_01", 
  "code_type": "letter",
  "start_code": "001001",
  "end_code": "001100",
  "batch_no": "BATCH_20250722_001"
}

# 申请信号编码
POST /api/signal-codes/request
Authorization: Bearer <courier_token>
Content-Type: application/json

{
  "school_id": "school_001",
  "area_id": "area_01",
  "code_type": "letter",
  "quantity": 10,
  "reason": "新任务分配需要"
}

# 分配信号编码
POST /api/signal-codes/assign
Content-Type: application/json

{
  "code": "001001",
  "user_id": "courier_123",
  "target_id": "letter_456",
  "target_type": "letter",
  "reason": "分配给信使使用"
}
```

#### 编码查询统计
```http
# 搜索编码
GET /api/signal-codes/search?school_id=school_001&is_used=false&page=1

# 获取统计信息
GET /api/signal-codes/stats/school_001

Response:
{
  "code": 0,
  "data": {
    "school_id": "school_001",
    "school_name": "北京大学",
    "total_codes": 1000,
    "used_codes": 650,
    "available_codes": 350,
    "usage_rate": 65.0,
    "by_type": {
      "letter": 800,
      "zone": 150,
      "building": 50
    }
  }
}

# 获取使用日志
GET /api/signal-codes/001001/logs?limit=10
```

### 5. 管理后台服务 (Port 8003)

#### 用户管理
```http
# 获取用户列表
GET /api/admin/users?role=all&page=1&limit=20
Authorization: Bearer <admin_token>

# 更新用户信息
PUT /api/admin/users/{user_id}
Content-Type: application/json

{
  "role": "courier",
  "status": "active",
  "permissions": ["courier_level_1"]
}

# 重置用户密码
POST /api/admin/users/{user_id}/reset-password
```

#### 信使管理
```http
# 获取信使列表
GET /api/admin/couriers?status=pending&level=1

# 审核信使申请
PUT /api/admin/couriers/{courier_id}/review
Content-Type: application/json

{
  "action": "approve",
  "level": 1,
  "zone_assignment": "area_01",
  "note": "符合要求，予以通过"
}
```

#### 系统统计
```http
# 获取系统统计
GET /api/admin/statistics?period=monthly

Response:
{
  "code": 0,
  "data": {
    "total_users": 15000,
    "active_couriers": 150,
    "letters_delivered": 8500,
    "success_rate": 0.94,
    "monthly_growth": 0.15
  }
}
```

### 6. OCR服务 (Port 8004)

#### 图像识别
```http
# 单个图像识别
POST /api/ocr/recognize
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data

image: <image_file>
letter_id: "OP9691NL7ZBOWC"

Response:
{
  "code": 0,
  "data": {
    "text": "识别出的文字内容",
    "confidence": 0.95,
    "language": "zh-CN",
    "processing_time": 1.2
  }
}

# 批量识别
POST /api/ocr/batch
Content-Type: multipart/form-data

images: <multiple_files>
```

#### 任务管理
```http
# 获取识别任务状态
GET /api/ocr/tasks/{task_id}

# 获取任务列表
GET /api/ocr/tasks?status=completed&page=1
```

## 🔔 WebSocket事件规范

### 连接方式
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = () => {
  // 发送认证信息
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'Bearer ' + jwt_token
  }));
};
```

### 事件格式
```json
{
  "type": "LETTER_STATUS_UPDATE",
  "data": {
    "letter_id": "OP9691NL7ZBOWC",
    "old_status": "generated",
    "new_status": "collected",
    "location": "北京大学宿舍楼下",
    "courier_id": "courier_123",
    "timestamp": "2025-07-22T14:00:00Z"
  },
  "user_id": "user_12345",
  "timestamp": "2025-07-22T14:00:00Z"
}
```

### 标准事件类型
```typescript
type WebSocketEventType = 
  | 'LETTER_STATUS_UPDATE'       // 信件状态更新
  | 'COURIER_TASK_ASSIGNMENT'    // 信使任务分配
  | 'COURIER_LOCATION_UPDATE'    // 信使位置更新
  | 'NEW_MESSAGE'                // 新消息通知
  | 'SYSTEM_NOTIFICATION'        // 系统通知
  | 'POINTS_UPDATED'             // 积分更新
  | 'LEVEL_UPGRADE'              // 等级提升
  | 'EXCEPTION_REPORTED'         // 异常上报
```

## 🗄️ 数据模型规范

### 核心数据模型

#### Letter (信件)
```json
{
  "id": "OP9691NL7ZBOWC",
  "title": "给朋友的问候信",
  "content": "加密后的信件内容",
  "sender_id": "user_12345",
  "receiver_hint": "北大宿舍楼，李同学",
  "status": "collected",
  "priority": "normal",
  "qr_code_url": "https://example.com/qr/OP9691NL7ZBOWC.png",
  "delivery_method": "courier",
  "estimated_delivery": "2025-07-23T15:00:00Z",
  "created_at": "2025-07-22T10:00:00Z",
  "updated_at": "2025-07-22T14:00:00Z"
}
```

#### Courier (信使)
```json
{
  "id": "courier_123",
  "user_id": "user_456", 
  "level": 1,
  "parent_id": "courier_789",
  "zone_code": "school_001_area_01",
  "zone_type": "building",
  "status": "active",
  "rating": 4.8,
  "total_tasks": 156,
  "completed_tasks": 142,
  "success_rate": 0.91,
  "points": 1250,
  "badges": ["新手", "百单达人"],
  "created_at": "2025-06-01T00:00:00Z",
  "last_active": "2025-07-22T13:30:00Z"
}
```

#### Task (任务)
```json
{
  "id": "T001",
  "letter_id": "OP9691NL7ZBOWC",
  "courier_id": "courier_123",
  "pickup_location": "北大宿舍楼32栋",
  "delivery_location": "清华大学图书馆", 
  "status": "in_progress",
  "priority": "urgent",
  "reward": 8.00,
  "estimated_distance": "15km",
  "estimated_time": "2小时",
  "accepted_at": "2025-07-22T12:30:00Z",
  "deadline": "2025-07-22T18:00:00Z",
  "created_at": "2025-07-22T10:00:00Z"
}
```

### 状态枚举

#### 信件状态
```json
{
  "letter_status": [
    "draft",        // 草稿
    "generated",    // 已生成二维码
    "collected",    // 已收取
    "in_transit",   // 投递中
    "delivered",    // 已投递
    "failed"        // 投递失败
  ]
}
```

#### 信使状态
```json
{
  "courier_status": [
    "pending",      // 申请中
    "approved",     // 已批准
    "active",       // 活跃
    "suspended",    // 暂停
    "banned"        // 禁用
  ]
}
```

#### 任务状态
```json
{
  "task_status": [
    "available",    // 可接取
    "accepted",     // 已接取
    "in_progress",  // 进行中
    "completed",    // 已完成
    "failed",       // 失败
    "cancelled"     // 已取消
  ]
}
```

## 🚨 错误处理规范

### 错误响应结构
```json
{
  "code": 1,
  "msg": "参数验证失败",
  "data": null,
  "error": {
    "type": "validation_error",
    "details": "字段验证失败",
    "fields": [
      {
        "field": "title",
        "message": "标题不能为空",
        "value": ""
      }
    ],
    "trace_id": "req_12345"
  },
  "timestamp": "2025-07-22T12:00:00Z"
}
```

### 错误类型约定
```
validation_error    - 参数验证错误
business_error      - 业务逻辑错误
permission_error    - 权限错误
not_found_error     - 资源不存在
rate_limit_error    - 频率限制错误
internal_error      - 服务内部错误
timeout_error       - 请求超时
dependency_error    - 依赖服务错误
```

## 📊 性能与质量要求

### 性能指标
- **响应时间**: < 200ms (P95)
- **可用性**: > 99.9%
- **并发处理**: > 1000 QPS
- **错误率**: < 0.1%

### 质量标准
- **代码覆盖率**: > 80%
- **API文档覆盖**: 100%
- **接口测试覆盖**: 100%
- **监控指标覆盖**: 100%

## 🔧 开发工具要求

### OpenAPI文档
- 每个服务必须提供OpenAPI 3.0规范文档
- 包含完整的请求/响应示例
- 错误响应的详细说明
- 业务场景描述

### API测试
- 单元测试覆盖率 > 80%
- 集成测试覆盖核心业务流程
- 性能测试确保响应时间要求
- 自动化测试流水线

### 监控日志
- 统一日志格式 (JSON)
- 包含trace_id用于链路追踪
- 关键业务操作必须记录审计日志
- 实时监控和告警

## 🔄 部署与运维

### 容器化部署
```yaml
# docker-compose.yml 示例
version: '3.8'
services:
  gateway:
    image: openpenpal/gateway:latest
    ports: ["8000:8000"]
    
  write-service:
    image: openpenpal/write-service:latest
    ports: ["8001:8001"]
    
  courier-service:
    image: openpenpal/courier-service:latest
    ports: ["8002:8002"]
    
  admin-service:
    image: openpenpal/admin-service:latest
    ports: ["8003:8003"]
    
  ocr-service:
    image: openpenpal/ocr-service:latest
    ports: ["8004:8004"]
```

### 服务发现
```yaml
服务注册:
  - 服务启动时自动注册到网关
  - 健康检查端点: /health
  - 服务元数据: version, status, capabilities

负载均衡:
  - 基于权重的轮询算法
  - 故障自动摘除
  - 服务实例动态扩缩容
```

### 监控体系
```yaml
指标监控:
  - Prometheus采集业务指标
  - Grafana可视化面板
  - 关键指标实时告警

链路追踪:
  - 分布式链路追踪
  - 请求流程可视化
  - 性能瓶颈分析

日志聚合:
  - 结构化日志收集
  - 日志检索和分析
  - 异常日志告警
```

---

## ✅ API规范检查清单

### 开发前确认
- [ ] 熟悉统一响应格式规范
- [ ] 了解JWT认证集成方式
- [ ] 掌握WebSocket事件推送协议
- [ ] 理解错误处理和状态码约定
- [ ] 准备OpenAPI文档生成
- [ ] 设置单元测试框架

### 开发完成验收
- [ ] API符合RESTful设计原则
- [ ] 响应格式完全统一
- [ ] JWT认证集成成功
- [ ] WebSocket事件正确推送
- [ ] 错误处理完善
- [ ] OpenAPI文档完整
- [ ] 测试覆盖率达标
- [ ] 性能指标符合要求

---

## 🎯 下一步改进方向

### 短期优化 (1-2周)
1. **完善API文档自动生成**
2. **加强接口测试覆盖**
3. **优化错误处理统一性**
4. **完善监控指标体系**

### 中期发展 (1-2月)
1. **API版本管理策略**
2. **GraphQL接口支持**
3. **服务治理平台建设**
4. **API网关功能增强**

### 长期规划 (3-6月)
1. **开放API平台建设**
2. **第三方开发者支持**
3. **API生态系统构建**
4. **国际化标准适配**

---

**记住**: 统一的API规范是微服务架构成功的关键基础！