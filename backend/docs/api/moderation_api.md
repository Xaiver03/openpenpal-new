# 内容审核系统 API 文档

## 概述

内容审核系统提供自动和人工审核功能，支持敏感词过滤、规则匹配和AI审核。系统可以审核信件内容、用户资料、照片、博物馆内容和信封设计等多种内容类型。

## API基础信息

- **基础路径**: `/api/v1/moderation`（普通用户）, `/api/v1/admin/moderation`（管理员）
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`

## API端点

### 1. 内容审核

触发对内容的自动审核。

**端点**: `POST /api/v1/moderation/check`

**请求体**:
```json
{
  "content_type": "letter",  // 必填：letter/profile/photo/museum/envelope
  "content_id": "uuid-string",  // 必填：内容ID
  "content": "需要审核的文本内容",
  "image_urls": ["https://example.com/image1.jpg"],  // 可选：图片URL数组
  "user_id": "user-uuid"  // 可选：内容创建者ID
}
```

**响应示例**:
```json
{
  "id": "moderation-uuid",
  "status": "approved",  // pending/approved/rejected/review
  "level": "low",  // low/medium/high/block
  "score": 0.1,  // 风险分数 0-1
  "reasons": ["敏感词: xxx"],
  "categories": ["政治敏感", "广告营销"],
  "need_review": false
}
```

### 2. 人工审核（管理员）

管理员对内容进行人工审核。

**端点**: `POST /api/v1/admin/moderation/review`

**请求体**:
```json
{
  "record_id": "moderation-uuid",  // 必填：审核记录ID
  "status": "approved",  // 必填：approved/rejected
  "review_note": "审核通过，内容健康"  // 可选：审核备注
}
```

**响应示例**:
```json
{
  "message": "Content reviewed successfully",
  "status": "approved"
}
```

### 3. 获取待审核队列（管理员）

获取需要人工审核的内容队列。

**端点**: `GET /api/v1/admin/moderation/queue`

**查询参数**:
- `limit`: 限制数量（默认20，最大100）

**响应示例**:
```json
{
  "queue": [
    {
      "id": "queue-uuid",
      "record_id": "moderation-uuid",
      "record": {
        "id": "moderation-uuid",
        "content_type": "letter",
        "content_id": "letter-uuid",
        "content": "信件内容...",
        "status": "review",
        "level": "medium",
        "score": 0.6,
        "reasons": ["包含敏感词"],
        "created_at": "2024-01-20T10:00:00Z"
      },
      "priority": 75,
      "status": "pending",
      "created_at": "2024-01-20T10:00:00Z"
    }
  ],
  "total": 5
}
```

### 4. 敏感词管理

#### 获取敏感词列表（管理员）

**端点**: `GET /api/v1/admin/moderation/sensitive-words`

**查询参数**:
- `category`: 分类（政治/色情/暴力/广告等）
- `level`: 等级（low/medium/high/block）

**响应示例**:
```json
{
  "words": [
    {
      "id": "word-uuid",
      "word": "敏感词",
      "category": "政治",
      "level": "high",
      "is_active": true,
      "created_at": "2024-01-20T10:00:00Z"
    }
  ],
  "total": 100
}
```

#### 添加敏感词（管理员）

**端点**: `POST /api/v1/admin/moderation/sensitive-words`

**请求体**:
```json
{
  "word": "新敏感词",  // 必填
  "category": "广告",  // 可选
  "level": "medium"  // 必填：low/medium/high/block
}
```

#### 更新敏感词（管理员）

**端点**: `PUT /api/v1/admin/moderation/sensitive-words/:id`

**请求体**:
```json
{
  "word": "更新的敏感词",
  "category": "广告",
  "level": "high"
}
```

#### 删除敏感词（管理员）

**端点**: `DELETE /api/v1/admin/moderation/sensitive-words/:id`

### 5. 审核规则管理

#### 获取审核规则列表（管理员）

**端点**: `GET /api/v1/admin/moderation/rules`

**查询参数**:
- `content_type`: 内容类型（letter/profile/photo/museum/envelope）

**响应示例**:
```json
{
  "rules": [
    {
      "id": "rule-uuid",
      "name": "禁止联系方式",
      "description": "检测并拦截包含手机号、微信号等联系方式的内容",
      "content_type": "letter",
      "rule_type": "regex",  // keyword/regex/ai
      "pattern": "\\d{11}|微信|QQ",
      "action": "review",  // block/review/pass
      "priority": 100,
      "is_active": true,
      "created_at": "2024-01-20T10:00:00Z"
    }
  ],
  "total": 10
}
```

#### 添加审核规则（管理员）

**端点**: `POST /api/v1/admin/moderation/rules`

**请求体**:
```json
{
  "name": "规则名称",  // 必填
  "description": "规则描述",
  "content_type": "letter",  // 必填
  "rule_type": "regex",  // 必填：keyword/regex/ai
  "pattern": "正则表达式或关键词",  // 必填
  "action": "review",  // 必填：block/review/pass
  "priority": 100  // 优先级，数字越大越优先
}
```

#### 更新审核规则（管理员）

**端点**: `PUT /api/v1/admin/moderation/rules/:id`

#### 删除审核规则（管理员）

**端点**: `DELETE /api/v1/admin/moderation/rules/:id`

### 6. 获取审核统计（管理员）

获取指定时间范围内的审核统计数据。

**端点**: `GET /api/v1/admin/moderation/stats`

**查询参数**:
- `start_date`: 开始日期（YYYY-MM-DD）
- `end_date`: 结束日期（YYYY-MM-DD）

**响应示例**:
```json
{
  "stats": [
    {
      "date": "2024-01-20",
      "content_type": "letter",
      "total_count": 1000,
      "approved_count": 950,
      "rejected_count": 30,
      "review_count": 20,
      "auto_moderate_count": 980,
      "avg_process_time": 2.5
    }
  ]
}
```

## 审核流程说明

1. **自动审核流程**:
   - 敏感词检查 → 规则匹配 → AI审核（如配置）→ 综合判断
   - 根据风险等级自动决定：通过、拒绝或人工审核

2. **人工审核流程**:
   - 系统将需要人工审核的内容加入队列
   - 管理员从队列获取待审内容
   - 管理员做出审核决定并填写备注

3. **风险等级**:
   - `low`: 低风险，通常自动通过
   - `medium`: 中风险，可能需要人工审核
   - `high`: 高风险，通常需要人工审核
   - `block`: 直接拒绝，包含严重违规内容

## 内容类型说明

- `letter`: 信件内容
- `profile`: 用户资料（昵称、签名等）
- `photo`: 照片图片
- `museum`: 博物馆展品内容
- `envelope`: 信封设计

## 错误码

| 错误码 | 说明 |
|-------|------|
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 注意事项

1. 所有内容在发布前都应该经过审核
2. 审核记录一旦创建不可删除，只能更新状态
3. 敏感词和规则的修改会立即生效
4. AI审核需要配置相应的API密钥
5. 人工审核应该在24小时内完成