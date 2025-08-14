# OpenPenPal API 接口文档 v1.0

## 目录
- [概述](#概述)
- [认证与授权](#认证与授权)
- [通用规范](#通用规范)
- [主要服务API](#主要服务api)
- [微服务API](#微服务api)
- [错误处理](#错误处理)
- [部署信息](#部署信息)

---

## 概述

OpenPenPal 是一个基于微服务架构的信件传递平台，提供完整的信件创作、投递、追踪和展示功能。本文档描述了平台所有可用的API接口。

### 架构概述
- **主服务 (Go)**: 核心业务逻辑和数据管理
- **前端代理 (Next.js)**: 前端API路由和认证处理
- **微服务集群**: 专业化功能服务
- **网关服务**: 路由和负载均衡

### 版本信息
- **API版本**: v1.0
- **文档版本**: 1.0
- **最后更新**: 2025-01-25

---

## 认证与授权

### 认证方式
- **Primary**: JWT Bearer Token
- **CSRF**: CSRF Token (前端路由)
- **Session**: Redis Session (WebSocket)

### 角色体系
```
super_admin > platform_admin > school_admin > courier_coordinator > senior_courier > courier > user
```

### 权限系统
```javascript
permissions = [
  "write_letter", "deliver_letter", "manage_users", "manage_school", 
  "manage_couriers", "admin_panel", "system_config", "view_analytics",
  "moderate_content", "manage_barcodes", "manage_museum", "manage_ai",
  "view_all_letters", "export_data", "manage_permissions", "super_admin"
]
```

### 认证Header
```http
Authorization: Bearer <jwt-token>
X-CSRF-Token: <csrf-token>
Content-Type: application/json
```

---

## 通用规范

### 请求格式
- **URL编码**: UTF-8
- **JSON格式**: camelCase (前端), snake_case (后端)
- **分页参数**: `page`, `limit`, `sort`, `order`
- **时间格式**: ISO 8601 (RFC3339)

### 响应格式
```json
{
  "code": 0,
  "message": "Success",
  "data": {},
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  },
  "timestamp": "2025-01-25T10:00:00Z"
}
```

### HTTP状态码
- **200**: 成功
- **201**: 创建成功
- **400**: 请求错误
- **401**: 未认证
- **403**: 权限不足
- **404**: 资源不存在
- **409**: 资源冲突
- **422**: 验证失败
- **500**: 服务器错误

---

## 主要服务API

### 1. 认证服务

#### 1.1 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "student001",
  "email": "student@university.edu",
  "password": "SecurePass123",
  "nickname": "小明",
  "school_code": "UNIV001",
  "verification_code": "123456"
}

Response:
{
  "code": 0,
  "message": "Registration successful",
  "data": {
    "user": {
      "id": "uuid-string",
      "username": "student001",
      "email": "student@university.edu",
      "nickname": "小明",
      "school_code": "UNIV001",
      "role": "user",
      "is_active": true,
      "created_at": "2025-01-25T10:00:00Z"
    },
    "tokens": {
      "access_token": "jwt-access-token",
      "refresh_token": "jwt-refresh-token",
      "expires_in": 3600
    }
  }
}
```

#### 1.2 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "student001",
  "password": "SecurePass123"
}

Response:
{
  "code": 0,
  "message": "Login successful",
  "data": {
    "user": { /* user object */ },
    "tokens": { /* tokens object */ },
    "permissions": ["write_letter", "deliver_letter"]
  }
}
```

#### 1.3 获取用户信息
```http
GET /api/v1/users/me
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": {
    "id": "uuid-string",
    "username": "student001",
    "email": "student@university.edu",
    "nickname": "小明",
    "avatar": "avatar-url",
    "school_code": "UNIV001",
    "role": "user",
    "permissions": ["write_letter"],
    "stats": {
      "letters_sent": 5,
      "letters_received": 3,
      "points": 100
    },
    "last_login_at": "2025-01-25T09:00:00Z",
    "created_at": "2025-01-20T10:00:00Z"
  }
}
```

### 2. 信件管理

#### 2.1 创建信件
```http
POST /api/v1/letters
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "写给朋友的信",
  "content": "亲爱的朋友，最近过得怎么样...",
  "style": "classic",
  "recipient_type": "friend",
  "is_anonymous": false,
  "delivery_method": "courier",
  "scheduled_send": "2025-01-26T10:00:00Z"
}

Response:
{
  "code": 0,
  "message": "Letter created successfully",
  "data": {
    "id": "letter-uuid",
    "title": "写给朋友的信",
    "content": "亲爱的朋友，最近过得怎么样...",
    "style": "classic",
    "status": "draft",
    "code": null,
    "qr_code_url": null,
    "created_at": "2025-01-25T10:00:00Z",
    "updated_at": "2025-01-25T10:00:00Z"
  }
}
```

#### 2.2 生成信件代码
```http
POST /api/v1/letters/{id}/generate-code
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Code generated successfully",
  "data": {
    "letter_id": "letter-uuid",
    "code": "OP1234567890",
    "qr_code_url": "https://cdn.openpenpal.com/qr/OP1234567890.png",
    "tracking_url": "https://openpenpal.com/track/OP1234567890",
    "status": "generated",
    "generated_at": "2025-01-25T10:00:00Z"
  }
}
```

#### 2.3 获取信件列表
```http
GET /api/v1/letters?page=1&limit=20&status=sent&sort=created_at&order=desc
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "letter-uuid",
      "title": "写给朋友的信",
      "status": "delivered",
      "code": "OP1234567890",
      "created_at": "2025-01-25T10:00:00Z",
      "delivered_at": "2025-01-25T15:00:00Z",
      "read_at": null,
      "recipient_feedback": null
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

#### 2.4 通过代码读取信件
```http
GET /api/v1/letters/read/{code}

Response:
{
  "code": 0,
  "message": "Success",
  "data": {
    "id": "letter-uuid",
    "title": "写给朋友的信",
    "content": "亲爱的朋友，最近过得怎么样...",
    "style": "classic",
    "sender_info": {
      "nickname": "小明",
      "school": "某某大学",
      "is_anonymous": false
    },
    "delivery_info": {
      "sent_at": "2025-01-25T10:00:00Z",
      "delivered_at": "2025-01-25T15:00:00Z",
      "delivery_method": "courier"
    },
    "is_read": false,
    "can_reply": true,
    "photos": [
      {
        "id": "photo-uuid",
        "url": "https://cdn.openpenpal.com/photos/photo1.jpg",
        "thumbnail_url": "https://cdn.openpenpal.com/photos/thumb/photo1.jpg"
      }
    ]
  }
}
```

### 3. 信使系统

#### 3.1 申请成为信使
```http
POST /api/v1/courier/apply
Authorization: Bearer <token>
Content-Type: application/json

{
  "real_name": "张三",
  "student_id": "2021001001",
  "phone": "13800138000",
  "available_hours": ["09:00-12:00", "14:00-17:00"],
  "transport_method": "bicycle",
  "delivery_zones": ["zone_a", "zone_b"],
  "experience": "有丰富的配送经验",
  "has_printer": true
}

Response:
{
  "code": 0,
  "message": "Application submitted successfully",
  "data": {
    "application_id": "app-uuid",
    "status": "pending",
    "submitted_at": "2025-01-25T10:00:00Z",
    "estimated_review_time": "3-5 working days"
  }
}
```

#### 3.2 获取信使任务
```http
GET /api/v1/courier/tasks?status=available&zone=zone_a&limit=10
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "task-uuid",
      "letter_code": "OP1234567890",
      "pickup_location": {
        "name": "图书馆",
        "address": "某某大学图书馆一楼",
        "coordinates": [116.3974, 39.9093]
      },
      "delivery_location": {
        "name": "宿舍区",
        "address": "某某大学东区宿舍1号楼",
        "coordinates": [116.3984, 39.9103]
      },
      "priority": "normal",
      "estimated_time": 30,
      "reward_points": 10,
      "deadline": "2025-01-25T18:00:00Z",
      "special_instructions": "请在工作时间投递"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

#### 3.3 接受信使任务
```http
POST /api/v1/courier/tasks/{task_id}/accept
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Task accepted successfully",
  "data": {
    "task_id": "task-uuid",
    "status": "in_progress",
    "accepted_at": "2025-01-25T10:00:00Z",
    "estimated_completion": "2025-01-25T11:00:00Z",
    "tracking_code": "TRACK123456"
  }
}
```

### 4. 信件博物馆

#### 4.1 获取博物馆展品
```http
GET /api/v1/museum/entries?page=1&limit=20&categories=写给朋友&tags=温柔,治愈&sort_by=popularity&order=desc&search=青春
Authorization: Bearer <token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "entry-uuid",
      "letter_id": "letter-uuid",
      "display_title": "写给十年后的自己",
      "author_display_type": "anonymous",
      "author_display_name": "某某大学大三学生",
      "categories": ["写给自己", "人生感悟"],
      "tags": ["温柔", "治愈", "青春"],
      "status": "featured",
      "view_count": 1234,
      "like_count": 89,
      "bookmark_count": 45,
      "share_count": 23,
      "created_at": "2025-01-20T10:00:00Z",
      "featured_at": "2025-01-22T14:00:00Z",
      "ai_metadata": {
        "sentiment_score": 0.8,
        "emotion_tags": ["温柔", "治愈"],
        "confidence_score": 0.92
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 156,
    "total_pages": 8
  }
}
```

#### 4.2 提交信件到博物馆
```http
POST /api/v1/museum/submit
Authorization: Bearer <token>
Content-Type: application/json

{
  "letter_id": "letter-uuid",
  "display_title": "写给十年后的自己",
  "author_display_type": "anonymous",
  "author_display_name": "某某大学大三学生",
  "categories": ["写给自己", "人生感悟"],
  "tags": ["温柔", "治愈", "青春"]
}

Response:
{
  "code": 0,
  "message": "Letter submitted to museum successfully",
  "data": {
    "entry_id": "entry-uuid",
    "submission_id": "submission-uuid",
    "status": "pending",
    "estimated_review_time": "1-3 working days",
    "submitted_at": "2025-01-25T10:00:00Z"
  }
}
```

#### 4.3 博物馆互动
```http
POST /api/v1/museum/entries/{entry_id}/interact
Authorization: Bearer <token>
Content-Type: application/json

{
  "interaction_type": "like",
  "metadata": {
    "source": "gallery_view"
  }
}

Response:
{
  "code": 0,
  "message": "Interaction recorded successfully",
  "data": {
    "interaction_type": "like",
    "is_active": true,
    "total_count": 90,
    "created_at": "2025-01-25T10:00:00Z"
  }
}
```

### 5. 条码系统

#### 5.1 生成条码
```http
POST /api/v1/barcodes
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "delivery",
  "metadata": {
    "letter_id": "letter-uuid",
    "priority": "normal"
  }
}

Response:
{
  "code": 0,
  "message": "Barcode generated successfully",
  "data": {
    "id": "barcode-uuid",
    "code": "BC1234567890",
    "type": "delivery",
    "status": "unactivated",
    "qr_code_url": "https://cdn.openpenpal.com/barcodes/BC1234567890.png",
    "security_hash": "sha256-hash",
    "created_at": "2025-01-25T10:00:00Z",
    "expires_at": "2025-02-25T10:00:00Z"
  }
}
```

#### 5.2 扫描条码
```http
PATCH /api/v1/barcodes/{code}/status
Content-Type: application/json

{
  "status": "scanned",
  "location": {
    "latitude": 39.9093,
    "longitude": 116.3974,
    "address": "某某大学图书馆"
  },
  "scanner_info": {
    "user_id": "user-uuid",
    "device_type": "mobile"
  }
}

Response:
{
  "code": 0,
  "message": "Barcode status updated successfully",
  "data": {
    "code": "BC1234567890",
    "status": "in_transit",
    "location": {
      "latitude": 39.9093,
      "longitude": 116.3974,
      "address": "某某大学图书馆"
    },
    "scan_log": {
      "id": "scan-uuid",
      "scanned_at": "2025-01-25T10:00:00Z",
      "scanner_id": "user-uuid"
    },
    "next_action": "deliver_to_recipient"
  }
}
```

---

## 微服务API

### 1. 管理后台服务 (Java Spring Boot) - Port 8081

#### 1.1 仪表盘概览
```http
GET /api/dashboard/overview
Authorization: Bearer <admin-token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": {
    "statistics": {
      "total_users": 5000,
      "active_users_today": 1200,
      "total_letters": 15000,
      "letters_delivered_today": 45,
      "active_couriers": 89,
      "pending_moderations": 12
    },
    "trends": {
      "user_growth": "+12.5%",
      "letter_volume": "+8.3%",
      "delivery_efficiency": "94.2%"
    },
    "alerts": [
      {
        "type": "warning",
        "message": "High delivery volume in Zone A",
        "timestamp": "2025-01-25T10:00:00Z"
      }
    ]
  }
}
```

#### 1.2 用户管理
```http
GET /api/users?page=1&limit=50&role=courier&school=UNIV001&search=张三
Authorization: Bearer <admin-token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "user-uuid",
      "username": "courier001",
      "nickname": "张三",
      "email": "zhang@university.edu",
      "role": "courier",
      "school_code": "UNIV001",
      "is_active": true,
      "stats": {
        "letters_delivered": 156,
        "success_rate": 98.5,
        "rating": 4.8
      },
      "last_active": "2025-01-25T09:30:00Z",
      "created_at": "2024-12-01T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 1250,
    "total_pages": 25
  }
}
```

#### 1.3 内容审核
```http
GET /api/moderation/tasks?type=museum&status=pending&priority=high
Authorization: Bearer <admin-token>

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "moderation-uuid",
      "content_type": "museum_submission",
      "content_id": "entry-uuid",
      "priority": "high",
      "status": "pending",
      "flags": ["sensitive_content", "manual_review"],
      "content_preview": {
        "title": "写给前任的信",
        "excerpt": "我想对你说...",
        "author": "匿名用户"
      },
      "ai_analysis": {
        "sentiment_score": -0.2,
        "risk_level": "medium",
        "suggested_action": "manual_review"
      },
      "created_at": "2025-01-25T09:00:00Z",
      "due_date": "2025-01-26T09:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 15,
    "total_pages": 1
  }
}
```

#### 1.4 审核内容
```http
POST /api/moderation/tasks/{task_id}/approve
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "action": "approve",
  "reason": "Content meets community guidelines",
  "notes": "Approved for public display",
  "tags": ["approved", "featured"]
}

Response:
{
  "code": 0,
  "message": "Content approved successfully",
  "data": {
    "task_id": "moderation-uuid",
    "action": "approve",
    "moderator_id": "admin-uuid",
    "processed_at": "2025-01-25T10:00:00Z",
    "content_status": "approved"
  }
}
```

### 2. 写信服务 (Python FastAPI) - Port 8002

#### 2.1 创建草稿
```http
POST /api/drafts
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "给朋友的生日祝福",
  "content": "亲爱的朋友，生日快乐！",
  "template_id": "birthday_template",
  "auto_save": true
}

Response:
{
  "code": 0,
  "message": "Draft created successfully",
  "data": {
    "id": "draft-uuid",
    "title": "给朋友的生日祝福",
    "content": "亲爱的朋友，生日快乐！",
    "word_count": 12,
    "character_count": 24,
    "auto_save_enabled": true,
    "last_saved": "2025-01-25T10:00:00Z",
    "created_at": "2025-01-25T10:00:00Z"
  }
}
```

#### 2.2 获取写作模板
```http
GET /api/templates?category=birthday&style=formal&language=zh

Response:
{
  "code": 0,
  "message": "Success",
  "data": [
    {
      "id": "template-uuid",
      "name": "正式生日祝福模板",
      "category": "birthday",
      "style": "formal",
      "language": "zh",
      "content": "亲爱的{recipient_name}，在这个特别的日子里...",
      "variables": ["recipient_name", "age", "relationship"],
      "preview_url": "https://cdn.openpenpal.com/templates/preview.jpg",
      "usage_count": 1250,
      "rating": 4.7
    }
  ]
}
```

#### 2.3 AI写作助手
```http
POST /api/ai/writing/assist
Authorization: Bearer <token>
Content-Type: application/json

{
  "action": "complete",
  "content": "亲爱的朋友，最近过得怎么样？我想告诉你",
  "context": {
    "recipient": "friend",
    "mood": "cheerful",
    "length": "medium"
  }
}

Response:
{
  "code": 0,
  "message": "AI assistance generated successfully",
  "data": {
    "suggestions": [
      {
        "type": "completion",
        "text": "一个好消息，我找到了一份很棒的实习工作！",
        "confidence": 0.85
      },
      {
        "type": "improvement",
        "original": "最近过得怎么样？",
        "improved": "最近学习和生活还顺利吗？",
        "reason": "更自然的表达"
      }
    ],
    "writing_tips": [
      "可以添加一些具体的生活细节",
      "表达感情时可以更加真诚"
    ]
  }
}
```

### 3. 信使服务 (Go) - Port 8003

#### 3.1 信使注册
```http
POST /api/couriers/register
Authorization: Bearer <token>
Content-Type: application/json

{
  "profile": {
    "real_name": "李四",
    "student_id": "2021001002",
    "phone": "13800138001",
    "id_card": "110101199001011234"
  },
  "capabilities": {
    "transport_methods": ["bicycle", "walking"],
    "delivery_zones": ["zone_a", "zone_b", "zone_c"],
    "available_hours": [
      {"day": "monday", "start": "09:00", "end": "17:00"},
      {"day": "tuesday", "start": "09:00", "end": "17:00"}
    ],
    "max_daily_tasks": 20,
    "has_printer": true,
    "languages": ["zh", "en"]
  },
  "preferences": {
    "preferred_zones": ["zone_a"],
    "notification_methods": ["app", "sms"],
    "auto_accept_tasks": false
  }
}

Response:
{
  "code": 0,
  "message": "Courier registration successful",
  "data": {
    "courier_id": "courier-uuid",
    "status": "pending_verification",
    "level": "beginner",
    "verification_steps": [
      {"step": "identity_verification", "status": "pending"},
      {"step": "training_completion", "status": "pending"},
      {"step": "background_check", "status": "pending"}
    ],
    "estimated_activation": "3-5 working days"
  }
}
```

#### 3.2 任务路径优化
```http
POST /api/delivery/optimize-route
Authorization: Bearer <token>
Content-Type: application/json

{
  "tasks": [
    {"id": "task1", "location": [116.3974, 39.9093], "priority": "high"},
    {"id": "task2", "location": [116.3984, 39.9103], "priority": "normal"},
    {"id": "task3", "location": [116.3994, 39.9113], "priority": "low"}
  ],
  "start_location": [116.3964, 39.9083],
  "transport_method": "bicycle",
  "max_duration": 180
}

Response:
{
  "code": 0,
  "message": "Route optimized successfully",
  "data": {
    "optimized_route": [
      {
        "task_id": "task1",
        "order": 1,
        "estimated_arrival": "2025-01-25T10:15:00Z",
        "travel_time": 15
      },
      {
        "task_id": "task2",
        "order": 2,
        "estimated_arrival": "2025-01-25T10:35:00Z",
        "travel_time": 20
      },
      {
        "task_id": "task3",
        "order": 3,
        "estimated_arrival": "2025-01-25T10:50:00Z",
        "travel_time": 15
      }
    ],
    "total_distance": 5.2,
    "total_duration": 50,
    "efficiency_score": 0.92
  }
}
```

### 4. 网关服务 (Go) - Port 8080

#### 4.1 服务健康检查
```http
GET /api/services/health

Response:
{
  "code": 0,
  "message": "Services health check",
  "data": {
    "gateway": {
      "status": "healthy",
      "version": "1.0.0",
      "uptime": 86400,
      "memory_usage": "45.2MB"
    },
    "services": {
      "main-backend": {
        "status": "healthy",
        "response_time": "15ms",
        "last_check": "2025-01-25T10:00:00Z"
      },
      "admin-service": {
        "status": "healthy",
        "response_time": "22ms",
        "last_check": "2025-01-25T10:00:00Z"
      },
      "write-service": {
        "status": "degraded",
        "response_time": "150ms",
        "last_check": "2025-01-25T10:00:00Z",
        "issues": ["high_response_time"]
      }
    }
  }
}
```

#### 4.2 系统指标
```http
GET /api/metrics?time_range=1h&services=all

Response:
{
  "code": 0,
  "message": "System metrics",
  "data": {
    "time_range": "2025-01-25T09:00:00Z to 2025-01-25T10:00:00Z",
    "metrics": {
      "requests": {
        "total": 15420,
        "successful": 14876,
        "failed": 544,
        "success_rate": 96.47
      },
      "response_times": {
        "avg": 89,
        "p50": 65,
        "p95": 245,
        "p99": 456
      },
      "throughput": {
        "requests_per_second": 4.28,
        "bytes_per_second": "1.2MB"
      },
      "errors": {
        "4xx": 234,
        "5xx": 310,
        "timeouts": 12
      }
    },
    "services": {
      "main-backend": {
        "requests": 8950,
        "avg_response_time": 45,
        "error_rate": 2.1
      },
      "admin-service": {
        "requests": 1250,
        "avg_response_time": 78,
        "error_rate": 1.5
      }
    }
  }
}
```

### 5. OCR服务 (Python FastAPI) - Port 8004

#### 5.1 图像OCR处理
```http
POST /api/ocr/process
Authorization: Bearer <token>
Content-Type: multipart/form-data

Form Data:
- image: <image-file>
- language: zh-en
- enhance: true
- format: structured

Response:
{
  "code": 0,
  "message": "OCR processing completed",
  "data": {
    "task_id": "ocr-task-uuid",
    "status": "completed",
    "result": {
      "text": "亲爱的朋友，最近过得怎么样？",
      "confidence": 0.94,
      "language": "zh",
      "regions": [
        {
          "text": "亲爱的朋友",
          "bbox": [120, 50, 200, 80],
          "confidence": 0.96
        },
        {
          "text": "最近过得怎么样？",
          "bbox": [120, 90, 250, 120],
          "confidence": 0.92
        }
      ],
      "metadata": {
        "processing_time": 1.25,
        "image_size": [800, 600],
        "enhancement_applied": true
      }
    }
  }
}
```

#### 5.2 批量OCR处理
```http
POST /api/ocr/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "images": [
    {"id": "img1", "url": "https://example.com/image1.jpg"},
    {"id": "img2", "url": "https://example.com/image2.jpg"}
  ],
  "options": {
    "language": "zh-en",
    "enhance": true,
    "callback_url": "https://your-app.com/ocr-callback"
  }
}

Response:
{
  "code": 0,
  "message": "Batch OCR job submitted",
  "data": {
    "job_id": "batch-job-uuid",
    "status": "processing",
    "total_images": 2,
    "processed_images": 0,
    "estimated_completion": "2025-01-25T10:05:00Z"
  }
}
```

---

## 前端代理API (Next.js)

### 认证与用户管理
```http
POST /api/auth/login                    # 用户登录
POST /api/auth/register                 # 用户注册
GET  /api/auth/me                       # 获取当前用户
POST /api/auth/refresh                  # 刷新Token
GET  /api/users/me/permissions          # 获取用户权限
```

### 地址与学校服务
```http
GET /api/address/search                 # 地址搜索
GET /api/schools                        # 学校列表
GET /api/postcode/{code}               # 邮编查询
```

### 信使与博物馆
```http
GET /api/courier/me                     # 信使信息
GET /api/museum/entries                 # 博物馆展品
POST /api/museum/submit                 # 提交到博物馆
```

### 管理功能
```http
GET /api/admin/audit-logs              # 审计日志
GET /api/admin/permissions             # 权限管理
GET /api/admin/settings                # 系统设置
```

---

## 错误处理

### 错误响应格式
```json
{
  "code": 400,
  "message": "Validation failed",
  "data": null,
  "error": {
    "type": "validation_error",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "code": "INVALID_FORMAT"
      }
    ]
  },
  "timestamp": "2025-01-25T10:00:00Z"
}
```

### 常见错误码
- **1000-1999**: 认证错误
- **2000-2999**: 权限错误
- **3000-3999**: 数据验证错误
- **4000-4999**: 业务逻辑错误
- **5000-5999**: 系统错误
- **9000-9999**: 第三方服务错误

### 错误类型
- `authentication_error`: 认证失败
- `authorization_error`: 权限不足
- `validation_error`: 数据验证失败
- `not_found_error`: 资源不存在
- `conflict_error`: 资源冲突
- `rate_limit_error`: 请求频率限制
- `service_unavailable`: 服务不可用

---

## 部署信息

### 环境配置
- **开发环境**: http://localhost:8080
- **测试环境**: https://test-api.openpenpal.com
- **生产环境**: https://api.openpenpal.com

### 服务端口
- **网关服务**: 8080
- **主后端**: 8080 (通过网关)
- **管理后台**: 8081
- **写信服务**: 8002
- **信使服务**: 8003
- **OCR服务**: 8004

### 认证配置
- **JWT密钥**: 环境变量配置
- **Token有效期**: 访问令牌1小时，刷新令牌30天
- **CSRF保护**: 前端路由启用

### 速率限制
- **公共API**: 1000/小时
- **认证API**: 100/小时
- **管理API**: 10000/小时
- **文件上传**: 100MB/请求

### 监控与日志
- **健康检查**: `/health`, `/ping`
- **指标收集**: Prometheus格式
- **日志级别**: info, warn, error
- **追踪ID**: 每个请求唯一ID

---

## 更新日志

### v1.0.0 (2025-01-25)
- ✅ 完整的API文档首版发布
- ✅ 涵盖所有核心功能模块
- ✅ 包含认证、授权、信件、信使、博物馆等完整功能
- ✅ 微服务架构完整API说明
- ✅ 错误处理和部署信息

---

*本文档由 OpenPenPal 开发团队维护，如有疑问请联系技术支持。*