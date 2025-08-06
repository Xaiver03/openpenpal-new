# OpenPenPal 阅读分析API文档

## 📊 概述

OpenPenPal阅读分析API提供了强大的信件阅读数据分析和统计功能，帮助用户了解信件的阅读情况、用户行为模式和趋势分析。

## 🚀 功能特性

### 核心分析功能
- ✅ **阅读统计分析** - 总阅读量、独立读者、阅读时长、完成率等
- ✅ **趋势分析** - 时间序列数据、增长率、高峰时段等
- ✅ **用户行为分析** - 个人阅读偏好、设备使用、时间分布等
- ✅ **热门内容排行** - 热门信件、活跃用户排行榜
- ✅ **实时统计** - 在线阅读者、实时事件流
- ✅ **信件对比分析** - 多封信件的对比分析和洞察
- ✅ **综合仪表板** - 多维度数据综合展示
- ✅ **数据导出** - 支持JSON/CSV/Excel格式导出

### 技术亮点
- 🎯 **智能缓存** - Redis缓存优化，5分钟TTL
- 📊 **多时间维度** - 小时/天/周/月/季度/年/自定义
- 🔐 **安全认证** - JWT认证，用户权限控制
- ⚡ **高性能** - 优化的SQL查询，异步处理
- 📈 **实时更新** - WebSocket事件驱动更新

## 📡 API端点列表

### 基础统计
- `GET /api/analytics/reading-stats` - 获取阅读统计数据
- `GET /api/analytics/trends` - 获取趋势分析数据
- `GET /api/analytics/popular` - 获取热门内容排行
- `GET /api/analytics/realtime` - 获取实时统计数据

### 详细分析
- `GET /api/analytics/letter/{letter_id}/analytics` - 获取单封信件详细分析
- `GET /api/analytics/user/{user_id}/behavior` - 获取用户行为分析
- `POST /api/analytics/compare` - 进行信件对比分析

### 综合功能
- `GET /api/analytics/dashboard` - 获取分析仪表板数据
- `POST /api/analytics/export` - 导出分析数据
- `GET /api/analytics/health` - 分析服务健康检查

## 🔧 详细API说明

### 1. 阅读统计分析

#### 请求
```http
GET /api/analytics/reading-stats?time_range=week&letter_id=OP123&user_id=user123
Authorization: Bearer {jwt_token}
```

#### 参数
- `time_range`: 时间范围 (hour/day/week/month/quarter/year/custom)
- `start_date`: 开始时间 (time_range为custom时必填)
- `end_date`: 结束时间 (time_range为custom时必填)
- `letter_id`: 特定信件ID (可选)
- `user_id`: 特定用户ID (可选)

#### 响应
```json
{
  "code": 0,
  "msg": "获取阅读统计成功",
  "data": {
    "total_reads": 1256,
    "unique_readers": 892,
    "avg_read_duration": 125.6,
    "complete_read_rate": 0.845,
    "device_distribution": {
      "mobile": 680,
      "desktop": 456,
      "tablet": 120
    },
    "browser_distribution": {
      "chrome": 567,
      "safari": 345,
      "firefox": 234,
      "edge": 110
    },
    "location_distribution": {
      "北京": 234,
      "上海": 198,
      "广州": 167
    },
    "hourly_distribution": {
      "9": 45,
      "10": 67,
      "11": 89,
      "14": 123,
      "20": 98
    }
  },
  "timestamp": "2025-07-21T12:00:00Z"
}
```

### 2. 信件详细分析

#### 请求
```http
GET /api/analytics/letter/OP1K2L3M4N5O/analytics
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取信件分析成功",
  "data": {
    "letter_id": "OP1K2L3M4N5O",
    "letter_title": "给朋友的问候信",
    "total_reads": 156,
    "unique_readers": 89,
    "first_read_at": "2025-07-20T14:30:00Z",
    "last_read_at": "2025-07-21T11:45:00Z",
    "avg_read_duration": 145.6,
    "max_read_duration": 450,
    "complete_reads": 132,
    "device_stats": {
      "mobile": 89,
      "desktop": 45,
      "tablet": 22
    },
    "browser_stats": {
      "chrome": 67,
      "safari": 45,
      "firefox": 34,
      "edge": 10
    },
    "time_distribution": [
      {"hour": 0, "count": 2, "label": "00:00"},
      {"hour": 1, "count": 1, "label": "01:00"},
      {"hour": 9, "count": 15, "label": "09:00"},
      {"hour": 14, "count": 23, "label": "14:00"},
      {"hour": 20, "count": 18, "label": "20:00"}
    ]
  }
}
```

### 3. 用户行为分析

#### 请求
```http
GET /api/analytics/user/user123/behavior?time_range=month
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取用户行为分析成功",
  "data": {
    "user_id": "user123",
    "total_letters_sent": 25,
    "total_reads_received": 456,
    "avg_reads_per_letter": 18.24,
    "most_read_letter": {
      "letter_id": "OP1K2L3M4N5O",
      "title": "最受欢迎的信件",
      "read_count": 89
    },
    "reading_time_stats": {
      "avg_duration": 156.7,
      "max_duration": 450,
      "min_duration": 15,
      "total_reading_time": 3425.6
    },
    "reader_demographics": {
      "unique_readers": 234,
      "device_preferences": {
        "mobile": 145,
        "desktop": 67,
        "tablet": 22
      },
      "browser_preferences": {
        "chrome": 123,
        "safari": 67,
        "firefox": 34,
        "edge": 10
      }
    }
  }
}
```

### 4. 趋势分析

#### 请求
```http
GET /api/analytics/trends?time_range=month&start_date=2025-06-21&end_date=2025-07-21
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取趋势分析成功",
  "data": {
    "time_series": [
      {
        "time": "2025-06-21",
        "count": 45,
        "timestamp": "2025-06-21T00:00:00Z"
      },
      {
        "time": "2025-06-22",
        "count": 67,
        "timestamp": "2025-06-22T00:00:00Z"
      }
    ],
    "growth_rate": 15.6,
    "peak_hours": [14, 20, 21],
    "peak_days": ["Wednesday", "Thursday", "Friday"],
    "seasonal_patterns": {
      "hourly": {
        "9": 45,
        "14": 123,
        "20": 98
      },
      "daily": {
        "Monday": 156,
        "Tuesday": 134,
        "Wednesday": 198
      },
      "monthly": {
        "6": 1245,
        "7": 1456
      }
    }
  }
}
```

### 5. 热门内容排行

#### 请求
```http
GET /api/analytics/popular?limit=10&time_range=week
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取热门内容成功",
  "data": {
    "top_letters": [
      {
        "letter_id": "OP1K2L3M4N5O",
        "title": "给朋友的问候信",
        "sender_id": "user123",
        "read_count": 156,
        "unique_readers": 89,
        "avg_duration": 145.6,
        "created_at": "2025-07-20T14:30:00Z"
      }
    ],
    "top_users": [
      {
        "user_id": "user123",
        "letters_count": 15,
        "total_reads": 456,
        "avg_reads_per_letter": 30.4
      }
    ],
    "trending_topics": []
  }
}
```

### 6. 实时统计

#### 请求
```http
GET /api/analytics/realtime
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取实时统计成功",
  "data": {
    "current_online_readers": 45,
    "reads_last_hour": 67,
    "reads_today": 234,
    "active_letters": [
      {
        "letter_id": "OP1K2L3M4N5O",
        "title": "热门信件",
        "recent_reads": 12
      }
    ],
    "live_events": [
      {
        "event_type": "letter_read",
        "letter_id": "OP1K2L3M4N5O",
        "letter_title": "信件标题",
        "read_at": "2025-07-21T12:00:00Z",
        "duration": 120,
        "complete": true
      }
    ]
  }
}
```

### 7. 信件对比分析

#### 请求
```http
POST /api/analytics/compare
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "letter_ids": ["OP1K2L3M4N5O", "OP2K2L3M4N5P", "OP3K2L3M4N5Q"],
  "metrics": ["reads", "duration", "completion_rate"]
}
```

#### 响应
```json
{
  "code": 0,
  "msg": "对比分析完成",
  "data": {
    "comparison_data": {
      "OP1K2L3M4N5O": {
        "total_reads": 156,
        "avg_read_duration": 145.6,
        "complete_reads": 132
      },
      "OP2K2L3M4N5P": {
        "total_reads": 89,
        "avg_read_duration": 167.3,
        "complete_reads": 78
      }
    },
    "insights": [
      "信件 OP1K2L3M4N5O 获得了最多的阅读次数：156 次",
      "信件 OP2K2L3M4N5P 有最长的平均阅读时长：167.3 秒"
    ],
    "recommendations": [
      "考虑分析高阅读量信件的内容特点，应用到其他信件中",
      "关注读者的阅读时长，适当调整内容长度和结构"
    ]
  }
}
```

### 8. 综合仪表板

#### 请求
```http
GET /api/analytics/dashboard?time_range=week
Authorization: Bearer {jwt_token}
```

#### 响应
```json
{
  "code": 0,
  "msg": "获取仪表板数据成功",
  "data": {
    "overview": {
      "total_reads": 1256,
      "unique_readers": 892,
      "avg_read_duration": 125.6,
      "complete_read_rate": 0.845
    },
    "trends": {
      "time_series": [...],
      "growth_rate": 15.6,
      "peak_hours": [14, 20, 21],
      "peak_days": ["Wednesday", "Thursday", "Friday"]
    },
    "popular_content": {
      "top_letters": [...],
      "top_users": [...]
    },
    "realtime": {
      "reads_last_hour": 67,
      "reads_today": 234
    },
    "device_distribution": {
      "mobile": 680,
      "desktop": 456,
      "tablet": 120
    },
    "time_distribution": {
      "9": 45,
      "14": 123,
      "20": 98
    }
  }
}
```

### 9. 数据导出

#### 请求
```http
POST /api/analytics/export
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "data_type": "reading_stats",
  "format": "json",
  "include_raw_data": false,
  "time_range": "month",
  "start_date": "2025-06-21T00:00:00Z",
  "end_date": "2025-07-21T23:59:59Z"
}
```

#### 响应
```json
{
  "code": 0,
  "msg": "数据导出成功",
  "data": {
    "export_format": "json",
    "data": {
      // 导出的数据内容
    },
    "generated_at": "2025-07-21T12:00:00Z"
  }
}
```

## 🔐 认证和权限

所有API端点都需要JWT认证：
```http
Authorization: Bearer {jwt_token}
```

## ⚡ 性能优化

### 缓存策略
- Redis缓存：5分钟TTL
- 缓存键格式：`{function}:{params_hash}`
- 自动缓存清理和更新

### 查询优化
- 数据库索引优化
- 分页查询支持
- 异步处理大数据量

## 🧪 测试和开发

### 测试脚本
```bash
# 运行API测试
python test_analytics_api.py

# 检查服务状态
curl http://localhost:8001/api/analytics/health
```

### 开发环境
```bash
# 启动服务
cd /path/to/write-service
source venv/bin/activate
uvicorn app.main:app --reload --port 8001
```

## 📈 数据模型

### ReadLog表结构
```sql
CREATE TABLE read_logs (
    id SERIAL PRIMARY KEY,
    letter_id VARCHAR(20) NOT NULL,
    reader_ip VARCHAR(45),
    reader_user_agent TEXT,
    reader_location VARCHAR(200),
    read_duration INTEGER,
    is_complete_read BOOLEAN DEFAULT TRUE,
    referer VARCHAR(500),
    device_info TEXT,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_letter_id (letter_id),
    INDEX idx_read_at (read_at),
    INDEX idx_reader_ip (reader_ip)
);
```

## 🔍 监控和告警

### 健康检查
- `/api/analytics/health` - 服务健康状态
- 数据库连接检查
- 缓存服务检查

### 性能指标
- API响应时间监控
- 缓存命中率统计
- 数据库查询性能

## 🚀 未来扩展计划

### 计划功能
- 📊 更多图表类型支持
- 🤖 AI驱动的内容分析
- 🌍 地理位置热力图
- 📧 自动化报告推送
- 📱 移动端专用API

### 集成计划
- 📈 Grafana仪表板集成
- 📊 BI工具数据对接
- 🔔 WebSocket实时推送
- 📤 邮件报告功能

---

## 📞 技术支持

- 📚 API文档: http://localhost:8001/docs
- 🔧 ReDoc文档: http://localhost:8001/redoc
- 🐛 问题报告: GitHub Issues
- 💬 技术讨论: 项目Wiki

---

*最后更新: 2025-07-21*
*版本: v1.0.0*