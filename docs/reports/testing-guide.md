# OpenPenPal 系统测试指南

## 系统概述

OpenPenPal 是一个校园手写信件平台，结合四级信使体系、AI集成、博物馆功能和实时WebSocket通信。系统采用Go + Gin框架构建，支持SQLite和PostgreSQL数据库。

## 目录

1. [测试账号信息](#1-测试账号信息)
2. [API端点完整列表](#2-api端点完整列表)
3. [权限系统说明](#3-权限系统说明)
4. [数据库架构](#4-数据库架构)
5. [四级信使系统](#5-四级信使系统)
6. [实时WebSocket功能](#6-实时websocket功能)
7. [测试流程和用例](#7-测试流程和用例)
8. [错误处理和状态码](#8-错误处理和状态码)

---

## 1. 测试账号信息

### 🔑 管理员账号

| 用户名 | 密码 | 角色 | 邮箱 | 学校代码 | 说明 |
|--------|------|------|------|----------|------|
| `admin` | `admin123` | super_admin | admin@penpal.com | SYSTEM | 系统超级管理员 |
| `school_admin` | `secret` | school_admin | school_admin@example.com | BJDX01 | 学校管理员 |
| `platform_admin` | `secret` | platform_admin | platform_admin@example.com | SYSTEM | 平台管理员 |

### 🚴 四级信使系统账号

| 用户名 | 密码 | 角色 | 邮箱 | 级别 | 权限范围 |
|--------|------|------|------|------|----------|
| `courier_level1` | `secret` | courier_level1 | courier1@openpenpal.com | 1级 | 楼栋/班级投递 |
| `courier_level2` | `secret` | courier_level2 | courier2@openpenpal.com | 2级 | 片区/年级管理 |
| `courier_level3` | `secret` | courier_level3 | courier3@openpenpal.com | 3级 | 学校级协调 |
| `courier_level4` | `secret` | courier_level4 | courier4@openpenpal.com | 4级 | 城市级总代 |

### 🏢 兼容性角色账号（旧版本兼容）

| 用户名 | 密码 | 角色 | 邮箱 | 等效级别 |
|--------|------|------|------|----------|
| `courier_building` | `courier001` | courier | courier_building@penpal.com | 1级 |
| `courier_area` | `courier002` | senior_courier | courier_area@penpal.com | 2级 |
| `courier_school` | `courier003` | courier_coordinator | courier_school@penpal.com | 3级 |
| `courier_city` | `courier004` | platform_admin | courier_city@penpal.com | 4级 |

### 👥 普通用户账号

| 用户名 | 密码 | 角色 | 邮箱 | 学校代码 | 说明 |
|--------|------|------|------|----------|------|
| `alice` | `secret` | user | alice@example.com | BJDX01 | 普通用户A |
| `bob` | `secret` | user | bob@example.com | BJDX01 | 普通用户B |
| `courier1` | `secret` | courier | courier1@example.com | BJDX01 | 普通信使 |

### 🧪 测试数据

**预置信件：**
- `test-letter-1`: "给朋友的第一封信" (草稿状态)
- `test-letter-2`: "感谢信" (已生成编号状态)

---

## 2. API端点完整列表

### 🌐 公开端点（无需认证）

#### 健康检查
- `GET /health` - 系统健康检查
- `GET /ping` - 简单ping测试

#### 用户认证
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

#### 公开信件访问
- `GET /api/v1/letters/read/:code` - 通过编号读取信件
- `POST /api/v1/letters/read/:code/mark-read` - 标记信件为已读
- `GET /api/v1/letters/public` - 获取广场公开信件

#### 公开信使信息
- `GET /api/v1/courier/stats` - 信使统计信息

#### 公开博物馆访问
- `GET /api/v1/museum/entries` - 博物馆条目列表
- `GET /api/v1/museum/entries/:id` - 特定博物馆条目
- `GET /api/v1/museum/exhibitions` - 博物馆展览列表

### 🔐 受保护端点（需要认证）

#### 用户管理
- `GET /api/v1/users/me` - 获取当前用户信息
- `PUT /api/v1/users/me` - 更新用户信息
- `POST /api/v1/users/me/change-password` - 修改密码
- `GET /api/v1/users/me/stats` - 获取用户统计
- `DELETE /api/v1/users/me` - 停用账户

#### 📝 信件管理
- `POST /api/v1/letters/` - 创建草稿信件
- `GET /api/v1/letters/` - 获取用户信件列表
- `GET /api/v1/letters/stats` - 获取信件统计
- `GET /api/v1/letters/:id` - 获取特定信件
- `PUT /api/v1/letters/:id` - 更新信件
- `DELETE /api/v1/letters/:id` - 删除信件
- `POST /api/v1/letters/:id/generate-code` - 生成信件编号

#### 📮 信封绑定功能
- `POST /api/v1/letters/:id/bind-envelope` - 绑定信封到信件
- `DELETE /api/v1/letters/:id/bind-envelope` - 解绑信封
- `GET /api/v1/letters/:id/envelope` - 获取信件信封信息

#### 💌 SOTA回信系统（新功能）
- `GET /api/v1/letters/scan-reply/:code` - 扫码获取回信信息
- `POST /api/v1/letters/replies` - 创建回信
- `GET /api/v1/letters/threads` - 获取用户对话线程列表
- `GET /api/v1/letters/threads/:id` - 获取线程详情

#### 🚴 信使系统
- `POST /api/v1/courier/apply` - 申请成为信使
- `GET /api/v1/courier/status` - 获取信使状态
- `GET /api/v1/courier/profile` - 获取信使档案
- `POST /api/v1/courier/letters/:code/status` - 更新信件状态

#### 🏗️ 四级信使管理
- `POST /api/v1/courier/create` - 创建下级信使
- `GET /api/v1/courier/subordinates` - 获取下级信使列表
- `GET /api/v1/courier/me` - 获取当前信使信息
- `GET /api/v1/courier/candidates` - 获取信使候选人

#### 📊 各级信使管理统计
- `GET /api/v1/courier/management/level-1/stats` - 一级信使统计
- `GET /api/v1/courier/management/level-1/couriers` - 一级信使列表
- `GET /api/v1/courier/management/level-2/stats` - 二级信使统计
- `GET /api/v1/courier/management/level-2/couriers` - 二级信使列表
- `GET /api/v1/courier/management/level-3/stats` - 三级信使统计
- `GET /api/v1/courier/management/level-3/couriers` - 三级信使列表
- `GET /api/v1/courier/management/level-4/stats` - 四级信使统计
- `GET /api/v1/courier/management/level-4/couriers` - 四级信使列表

#### 📦 信封系统
- `GET /api/v1/envelopes/my` - 获取我的信封
- `GET /api/v1/envelopes/designs` - 获取信封设计
- `POST /api/v1/envelopes/orders` - 创建信封订单
- `GET /api/v1/envelopes/orders` - 获取信封订单列表
- `POST /api/v1/envelopes/orders/:id/pay` - 处理信封支付

#### 🏛️ 博物馆系统
- `POST /api/v1/museum/items` - 创建博物馆物品

#### 🤖 AI功能
- `POST /api/v1/ai/match` - AI笔友匹配
- `POST /api/v1/ai/reply` - AI回信生成
- `POST /api/v1/ai/inspiration` - 获取写作灵感
- `POST /api/v1/ai/curate` - AI策展
- `GET /api/v1/ai/personas` - 获取AI人设
- `GET /api/v1/ai/stats` - AI统计信息
- `GET /api/v1/ai/daily-inspiration` - 每日灵感

#### 📈 数据分析
- `GET /api/v1/analytics/dashboard` - 分析仪表盘
- `GET /api/v1/analytics/metrics` - 获取指标
- `POST /api/v1/analytics/metrics` - 记录指标
- `GET /api/v1/analytics/metrics/summary` - 指标摘要
- `GET /api/v1/analytics/users` - 用户分析
- `POST /api/v1/analytics/reports` - 生成报告
- `GET /api/v1/analytics/reports` - 获取报告
- `POST /api/v1/analytics/performance` - 记录性能数据

#### ⏰ 任务调度
- `POST /api/v1/scheduler/tasks` - 创建定时任务
- `GET /api/v1/scheduler/tasks` - 获取任务列表
- `GET /api/v1/scheduler/tasks/:id` - 获取特定任务
- `PUT /api/v1/scheduler/tasks/:id/status` - 更新任务状态
- `POST /api/v1/scheduler/tasks/:id/enable` - 启用任务
- `POST /api/v1/scheduler/tasks/:id/disable` - 禁用任务
- `POST /api/v1/scheduler/tasks/:id/execute` - 立即执行任务
- `DELETE /api/v1/scheduler/tasks/:id` - 删除任务
- `GET /api/v1/scheduler/tasks/:id/executions` - 获取任务执行记录
- `GET /api/v1/scheduler/stats` - 调度器统计
- `POST /api/v1/scheduler/tasks/defaults` - 创建默认任务

#### 🔔 通知系统
- `GET /api/v1/notifications/` - 获取用户通知
- `POST /api/v1/notifications/send` - 发送通知
- `POST /api/v1/notifications/:id/read` - 标记通知已读
- `POST /api/v1/notifications/read-all` - 全部标记已读
- `GET /api/v1/notifications/preferences` - 获取通知偏好
- `PUT /api/v1/notifications/preferences` - 更新通知偏好
- `POST /api/v1/notifications/test-email` - 测试邮件通知

#### 🌐 WebSocket实时通信
- `GET /api/v1/ws/connect` - 连接WebSocket
- `GET /api/v1/ws/connections` - 获取连接信息
- `GET /api/v1/ws/stats` - WebSocket统计
- `GET /api/v1/ws/rooms/:room/users` - 获取房间用户
- `POST /api/v1/ws/broadcast` - 广播消息
- `POST /api/v1/ws/direct` - 发送直接消息
- `GET /api/v1/ws/history` - 获取消息历史

### 👑 管理员端点（需要管理员权限）

#### 🎛️ 管理仪表盘
- `GET /api/v1/admin/dashboard/stats` - 仪表盘统计
- `GET /api/v1/admin/dashboard/activities` - 最近活动
- `GET /api/v1/admin/dashboard/analytics` - 分析数据
- `POST /api/v1/admin/seed-data` - 注入测试数据
- `GET /api/v1/admin/system/settings` - 系统设置

#### 👥 管理员用户管理
- `GET /api/v1/admin/users/` - 用户管理数据
- `GET /api/v1/admin/users/:id` - 获取特定用户（管理视图）
- `DELETE /api/v1/admin/users/:id` - 停用用户
- `POST /api/v1/admin/users/:id/reactivate` - 重新激活用户

#### 🚴 管理员信使管理
- `GET /api/v1/admin/courier/applications` - 待审核申请
- `POST /api/v1/admin/courier/:id/approve` - 批准信使申请
- `POST /api/v1/admin/courier/:id/reject` - 拒绝信使申请

#### 🏛️ 管理员博物馆管理
- `POST /api/v1/admin/museum/items/:id/approve` - 批准博物馆物品

#### 📊 管理员分析
- `GET /api/v1/admin/analytics/system` - 系统分析
- `GET /api/v1/admin/analytics/dashboard` - 管理分析仪表盘
- `GET /api/v1/admin/analytics/reports` - 管理报告

### 📁 静态文件服务
- `GET /uploads/*` - 静态文件服务（二维码、图片等）

---

## 3. 权限系统说明

### 🏆 用户角色层级

| 级别 | 角色 | 中文名称 | 权限范围 |
|------|------|----------|----------|
| 1 | `user` | 普通用户 | 基础用户权限 |
| 2 | `courier` / `courier_level1` | 一级信使 | 楼栋/班级投递 |
| 3 | `senior_courier` / `courier_level2` | 二级信使 | 片区/年级管理 |
| 4 | `courier_coordinator` / `courier_level3` | 三级信使 | 学校级协调 |
| 5 | `courier_level4` | 四级信使 | 城市级总代 |
| 6 | `school_admin` | 学校管理员 | 学校管理权限 |
| 7 | `platform_admin` | 平台管理员 | 平台管理权限 |
| 8 | `super_admin` | 超级管理员 | 系统管理权限 |

### 🔑 权限分类

#### 基础用户权限
- `write_letter` - 创建和编辑信件
- `read_letter` - 阅读信件
- `manage_profile` - 管理用户档案

#### 信使权限
- `deliver_letter` - 投递信件
- `scan_code` - 扫描二维码
- `view_tasks` - 查看分配任务

#### 协调员权限
- `manage_couriers` - 管理下级信使
- `assign_tasks` - 分配任务
- `view_reports` - 查看报告

#### 管理员权限
- `manage_users` - 管理平台用户
- `manage_school` - 管理学校设置
- `view_analytics` - 查看系统分析
- `manage_system` - 系统配置
- `manage_platform` - 平台级管理
- `manage_admins` - 管理其他管理员
- `system_config` - 系统配置访问

### 🛡️ JWT认证

- JWT令牌必须包含在受保护端点的请求中
- 令牌包含用户ID和过期信息
- 请求头格式：`Authorization: Bearer <token>`
- 令牌验证包括用户活跃状态检查

---

## 4. 数据库架构

### 📊 核心数据表

#### 👤 users（用户表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
username VARCHAR(50) UNIQUE      -- 唯一用户名
email VARCHAR(100) UNIQUE        -- 邮箱
password_hash VARCHAR(255)       -- bcrypt密码哈希
nickname VARCHAR(50)             -- 显示名称
avatar VARCHAR(500)              -- 头像URL
role VARCHAR(20)                 -- 用户角色
school_code VARCHAR(20)          -- 学校标识码
is_active BOOLEAN DEFAULT TRUE   -- 账户状态
last_login_at TIMESTAMP          -- 最后登录时间
created_at, updated_at, deleted_at -- 时间戳字段
```

#### 📝 letters（信件表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
user_id VARCHAR(36)              -- 外键关联用户
title VARCHAR(255)               -- 信件标题
content TEXT                     -- 信件内容
style VARCHAR(20)                -- 信件样式：classic/modern/vintage/elegant/casual
status VARCHAR(20)               -- 状态：draft/generated/collected/in_transit/delivered/read
reply_to VARCHAR(36)             -- 回复线程引用
envelope_id VARCHAR(36)          -- 关联信封
created_at, updated_at, deleted_at
```

#### 🔢 letter_codes（信件编号表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
letter_id VARCHAR(36)            -- 外键关联信件
code VARCHAR(50) UNIQUE          -- 唯一投递编号
qr_code_url VARCHAR(500)         -- 二维码URL
qr_code_path VARCHAR(500)        -- 二维码文件路径
expires_at TIMESTAMP             -- 编号过期时间
```

#### 🚴 couriers（信使表）
```sql
-- 主要字段
id PRIMARY KEY                   -- 主键
user_id                          -- 外键关联用户
name VARCHAR                     -- 信使姓名
contact VARCHAR                  -- 联系方式
school VARCHAR                   -- 学校名称
zone VARCHAR                     -- 覆盖区域
has_printer BOOLEAN              -- 是否有打印机
self_intro TEXT                  -- 自我介绍
can_mentor BOOLEAN               -- 是否可指导新人
weekly_hours INTEGER             -- 每周可用小时
max_daily_tasks INTEGER          -- 每日最大任务数
transport_method VARCHAR         -- 交通方式
time_slots TEXT                  -- 可用时间段（JSON）
status VARCHAR                   -- 申请状态：pending/approved/rejected
level INTEGER                    -- 信使级别（1-4）
task_count INTEGER               -- 完成任务数
points INTEGER                   -- 累积积分
```

#### 📦 envelopes（信封表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
design_id VARCHAR(36)            -- 外键关联信封设计
user_id VARCHAR(36)              -- 拥有者用户ID
used_by VARCHAR(36)              -- 使用者用户ID
letter_id VARCHAR(36)            -- 关联信件
barcode_id VARCHAR(100) UNIQUE   -- 唯一条形码
status VARCHAR(20)               -- 状态：unsent/used/cancelled
used_at TIMESTAMP                -- 使用时间
```

#### 🎨 envelope_designs（信封设计表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
school_code VARCHAR(20)          -- 学校标识
type VARCHAR(20)                 -- 设计类型：city/school
theme VARCHAR(100)               -- 设计主题
image_url VARCHAR(500)           -- 设计图片URL
thumbnail_url VARCHAR(500)       -- 缩略图URL
creator_id VARCHAR(36)           -- 创建者用户ID
creator_name VARCHAR(100)        -- 创建者姓名
description TEXT                 -- 设计描述
status VARCHAR(20)               -- 审核状态
vote_count INTEGER               -- 社区投票数
period VARCHAR(50)               -- 设计周期
is_active BOOLEAN                -- 活跃状态
```

#### 🏛️ museum_items（博物馆物品表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
source_type VARCHAR(20)          -- 物品类型：letter/photo/audio
source_id VARCHAR(36)            -- 源引用ID
title VARCHAR(200)               -- 物品标题
description TEXT                 -- 物品描述
tags TEXT                        -- 可搜索标签
status VARCHAR(20)               -- 审核状态：pending/approved/rejected
submitted_by VARCHAR(36)         -- 提交者用户ID
approved_by VARCHAR(36)          -- 审核者用户ID
approved_at TIMESTAMP            -- 审核时间
view_count INTEGER               -- 浏览量
like_count INTEGER               -- 点赞数
share_count INTEGER              -- 分享数
```

#### 🔔 notifications（通知表）
```sql
-- 主要字段
id VARCHAR(36) PRIMARY KEY       -- UUID主键
user_id VARCHAR(36)              -- 接收者用户ID
type VARCHAR(50)                 -- 通知类型
channel VARCHAR(20)              -- 投递渠道：websocket/email/sms/push
priority VARCHAR(20)             -- 优先级
title VARCHAR(200)               -- 通知标题
content TEXT                     -- 通知内容
data TEXT                        -- 附加数据（JSON）
status VARCHAR(20)               -- 投递状态
scheduled_at TIMESTAMP           -- 计划投递时间
sent_at TIMESTAMP                -- 实际投递时间
read_at TIMESTAMP                -- 阅读时间
```

#### 🤖 AI相关表
- `ai_matches` - AI笔友匹配记录
- `ai_replies` - AI生成的回复
- `ai_inspirations` - 写作灵感
- `ai_curations` - 内容策展记录
- `ai_configs` - AI提供商配置
- `ai_usage_logs` - AI服务使用跟踪

---

## 5. 四级信使系统

### 🏗️ 信使级别结构

#### 1级信使（楼栋/班级级）
- **职责范围**：楼栋内或班级内的基础信件投递
- **权限**：基础信使权限
- **管理者**：2级及以上信使

#### 2级信使（片区/年级级）
- **职责范围**：管理多个楼栋或年级级别
- **权限**：可创建和管理1级信使，查看报告
- **管理者**：3级及以上信使

#### 3级信使（学校级）
- **职责范围**：管理整个学校的信使网络
- **权限**：可创建和管理1-2级信使，分配任务
- **管理者**：4级信使

#### 4级信使（城市级）
- **职责范围**：协调跨学校投递
- **权限**：可创建和管理1-3级信使，学校管理权限
- **管理者**：平台管理员

### 📊 信使管理API

每个级别都有特定的管理端点用于查看统计信息和管理下级信使：

```bash
# 获取2级信使统计信息
GET /api/v1/courier/management/level-2/stats

# 获取3级信使列表
GET /api/v1/courier/management/level-3/couriers
```

---

## 6. 实时WebSocket功能

### 🌐 事件类型

#### 📝 信件事件
- `LETTER_STATUS_UPDATE` - 信件状态变更
- `LETTER_CREATED` - 新信件创建
- `LETTER_READ` - 信件被收件人阅读
- `LETTER_DELIVERED` - 信件投递完成

#### 🚴 信使事件
- `COURIER_LOCATION_UPDATE` - 信使位置更新
- `NEW_TASK_ASSIGNMENT` - 新任务分配给信使
- `TASK_STATUS_UPDATE` - 任务状态变更
- `COURIER_ONLINE/OFFLINE` - 信使在线/离线状态

#### 👤 用户事件
- `USER_ONLINE/OFFLINE` - 用户在线状态
- `NOTIFICATION` - 实时通知

#### 🛠️ 系统事件
- `SYSTEM_MESSAGE` - 系统公告
- `HEARTBEAT` - 连接健康检查
- `ERROR` - 错误通知

### 🏠 WebSocket房间

- `global` - 全局广播
- `system` - 系统消息
- `school:{code}` - 学校特定消息
- `couriers` - 信使专用消息
- `admins` - 管理员专用消息
- `user:{id}` - 个人消息
- `letter:{id}` - 信件跟踪更新

---

## 7. 测试流程和用例

### 🔐 基础认证测试

```bash
# 管理员登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 使用返回的token进行认证请求
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 📝 信件创建和管理测试

```bash
# 创建草稿信件
curl -X POST http://localhost:8080/api/v1/letters/ \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试信件",
    "content": "你好世界！",
    "style": "classic"
  }'

# 生成投递编号
curl -X POST http://localhost:8080/api/v1/letters/{id}/generate-code \
  -H "Authorization: Bearer USER_TOKEN"

# 获取用户信件列表
curl -X GET http://localhost:8080/api/v1/letters/?page=1&limit=10 \
  -H "Authorization: Bearer USER_TOKEN"
```

### 💌 SOTA回信系统测试

```bash
# 扫码获取回信信息
curl -X GET http://localhost:8080/api/v1/letters/scan-reply/LETTER_CODE \
  -H "Authorization: Bearer USER_TOKEN"

# 创建回信
curl -X POST http://localhost:8080/api/v1/letters/replies \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "original_letter_code": "LETTER_CODE",
    "content": "感谢你的来信！",
    "style": "modern",
    "is_public": false
  }'

# 获取用户对话线程
curl -X GET http://localhost:8080/api/v1/letters/threads?page=1&limit=10 \
  -H "Authorization: Bearer USER_TOKEN"
```

### 🚴 信使系统测试

```bash
# 申请成为信使
curl -X POST http://localhost:8080/api/v1/courier/apply \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试信使",
    "contact": "test@example.com",
    "school": "PKU001",
    "zone": "A栋",
    "hasPrinter": "yes",
    "canMentor": "yes",
    "weeklyHours": 10,
    "maxDailyTasks": 5,
    "transportMethod": "自行车",
    "timeSlots": ["9:00-12:00", "14:00-17:00"]
  }'

# 检查信使状态
curl -X GET http://localhost:8080/api/v1/courier/status \
  -H "Authorization: Bearer COURIER_TOKEN"

# 创建下级信使（需要2级及以上权限）
curl -X POST http://localhost:8080/api/v1/courier/create \
  -H "Authorization: Bearer COURIER_L2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_level": 1,
    "user_id": "target_user_id",
    "zone": "B栋"
  }'
```

### 📦 信封系统测试

```bash
# 获取信封设计
curl -X GET http://localhost:8080/api/v1/envelopes/designs \
  -H "Authorization: Bearer USER_TOKEN"

# 创建信封订单
curl -X POST http://localhost:8080/api/v1/envelopes/orders \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "design_id": "design-uuid",
    "quantity": 5
  }'

# 绑定信封到信件
curl -X POST http://localhost:8080/api/v1/letters/{letter_id}/bind-envelope \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "envelope_id": "envelope-uuid"
  }'
```

### 🏛️ 博物馆系统测试

```bash
# 获取博物馆条目
curl -X GET http://localhost:8080/api/v1/museum/entries

# 创建博物馆物品（需认证）
curl -X POST http://localhost:8080/api/v1/museum/items \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sourceType": "letter",
    "sourceId": "letter-uuid",
    "title": "美丽的信件",
    "description": "一份精彩的写作作品"
  }'
```

### 🤖 AI功能测试

```bash
# 获取AI写作灵感
curl -X POST http://localhost:8080/api/v1/ai/inspiration \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "theme": "友谊",
    "style": "休闲",
    "count": 3
  }'

# 请求AI笔友匹配
curl -X POST http://localhost:8080/api/v1/ai/match \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "letter_id": "letter-uuid",
    "max_matches": 3
  }'
```

### 🌐 WebSocket测试

```javascript
// 连接WebSocket并认证
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/connect', [], {
  headers: {
    'Authorization': 'Bearer YOUR_JWT_TOKEN'
  }
});

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('收到消息:', message);
};

// 发送消息
ws.send(JSON.stringify({
  type: 'join_room',
  room: 'school:PKU001',
  data: {}
}));
```

### 👑 管理员功能测试

```bash
# 获取仪表盘统计（仅管理员）
curl -X GET http://localhost:8080/api/v1/admin/dashboard/stats \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 批准信使申请
curl -X POST http://localhost:8080/api/v1/admin/courier/{id}/approve \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 获取用户管理数据
curl -X GET http://localhost:8080/api/v1/admin/users/ \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

---

## 8. 错误处理和状态码

### 📊 HTTP状态码说明

| 状态码 | 说明 | 常见场景 |
|--------|------|----------|
| 200 | 成功 | 正常请求成功 |
| 201 | 创建成功 | 创建新资源成功 |
| 400 | 请求错误 | 参数错误、数据验证失败 |
| 401 | 未授权 | 缺少或无效的认证token |
| 403 | 禁止访问 | 权限不足 |
| 404 | 未找到 | 资源不存在 |
| 409 | 冲突 | 资源冲突（如用户名已存在） |
| 422 | 无法处理的实体 | 数据格式正确但语义错误 |
| 500 | 内部服务器错误 | 服务器内部错误 |

### 🛠️ 错误响应格式

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "用户输入验证失败",
    "details": {
      "field": "email",
      "reason": "邮箱格式不正确"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 🔍 常见错误类型

#### 认证错误
- `INVALID_TOKEN` - 无效的JWT token
- `TOKEN_EXPIRED` - token已过期
- `INSUFFICIENT_PERMISSIONS` - 权限不足

#### 验证错误
- `VALIDATION_ERROR` - 数据验证失败
- `MISSING_REQUIRED_FIELD` - 缺少必填字段
- `INVALID_FORMAT` - 格式不正确

#### 业务逻辑错误
- `USER_NOT_FOUND` - 用户不存在
- `LETTER_NOT_FOUND` - 信件不存在
- `PERMISSION_DENIED` - 操作被拒绝
- `QUOTA_EXCEEDED` - 超出配额限制

---

## 🔧 配置和环境

### 环境变量配置

```bash
# 服务器配置
PORT=8080
HOST=0.0.0.0
ENVIRONMENT=development

# 数据库配置
DATABASE_TYPE=sqlite
DATABASE_URL=./openpenpal.db

# 安全配置
JWT_SECRET=your-secret-key-change-in-production
BCRYPT_COST=10

# 前端配置
FRONTEND_URL=http://localhost:3000

# AI服务配置
OPENAI_API_KEY=your-openai-api-key
CLAUDE_API_KEY=your-claude-api-key
AI_PROVIDER=openai

# 邮件配置（用于通知）
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@openpenpal.com
SMTP_PASSWORD=your-smtp-password
```

### 🔄 数据库支持

系统支持SQLite（开发环境）和PostgreSQL（生产环境），带有自动迁移和开发环境自动数据填充功能。

---

## 📝 测试检查清单

### ✅ 基础功能测试

- [ ] 用户注册和登录
- [ ] JWT token认证
- [ ] 用户档案管理
- [ ] 密码修改

### ✅ 信件系统测试

- [ ] 创建草稿信件
- [ ] 生成信件编号和二维码
- [ ] 信件状态更新流程
- [ ] 公开信件浏览
- [ ] 信件阅读和标记已读

### ✅ SOTA回信系统测试

- [ ] 扫码获取回信信息
- [ ] 创建回信
- [ ] 对话线程管理
- [ ] 线程详情查看

### ✅ 四级信使系统测试

- [ ] 信使申请流程
- [ ] 各级信使权限验证
- [ ] 下级信使创建
- [ ] 信使统计信息
- [ ] 任务分配和跟踪

### ✅ 信封系统测试

- [ ] 信封设计浏览
- [ ] 信封订单创建
- [ ] 信封支付处理
- [ ] 信封绑定到信件

### ✅ 博物馆系统测试

- [ ] 博物馆条目浏览
- [ ] 物品提交
- [ ] 管理员审核
- [ ] 展览管理

### ✅ AI功能测试

- [ ] 写作灵感生成
- [ ] 笔友匹配
- [ ] AI回信生成
- [ ] 内容策展

### ✅ 实时通信测试

- [ ] WebSocket连接
- [ ] 实时通知
- [ ] 房间管理
- [ ] 消息广播

### ✅ 管理员功能测试

- [ ] 仪表盘统计
- [ ] 用户管理
- [ ] 信使申请审核
- [ ] 系统配置

---

## 📞 支持和反馈

如需技术支持或有任何问题，请通过以下方式联系：

- **GitHub Issues**: 报告Bug和功能请求
- **文档更新**: 本文档将持续更新以反映系统变化
- **测试结果**: 请记录测试结果并报告发现的问题

---

**最后更新**: 2024年1月（基于当前数据库状态和系统架构）
**文档版本**: v1.0
**系统版本**: OpenPenPal Backend v1.0.0