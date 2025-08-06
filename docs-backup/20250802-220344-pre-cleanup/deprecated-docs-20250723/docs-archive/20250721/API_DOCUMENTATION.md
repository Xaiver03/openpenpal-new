# API Gateway 性能监控API文档

## 🎯 新增API接口

### 1. 性能指标上报API

**接口地址**: `POST /api/v1/metrics/performance`

**描述**: 前端提交性能指标数据，用于监控Core Web Vitals和其他性能数据

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "session_id": "user-session-123",
  "page_url": "/dashboard",
  "lcp": 1200.5,
  "fid": 45.2,
  "cls": 0.08,
  "ttfb": 350.1,
  "load_time": 2100.3,
  "dom_ready": 1800.7,
  "first_paint": 950.2,
  "js_heap_size": 15728640,
  "connection_type": "4g",
  "download_speed": 12.5,
  "device_type": "desktop",
  "screen_size": "1920x1080"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "session_id": "user-session-123",
    "timestamp": "2025-07-21T09:30:00Z"
  },
  "timestamp": "2025-07-21T09:30:00Z"
}
```

### 2. 获取性能仪表板数据

**接口地址**: `GET /api/v1/metrics/dashboard`

**描述**: 获取聚合的性能指标数据用于仪表板展示

**需要认证**: JWT Token

**查询参数**:
- `time_range`: 时间范围 (1h/24h/7d/30d)，默认24h

**响应示例**:
```json
{
  "code": 0,
  "msg": "success", 
  "data": {
    "time_range": "24h",
    "avg_lcp": 1856.3,
    "avg_fid": 67.8,
    "avg_cls": 0.12,
    "avg_ttfb": 445.6,
    "performance_score": 78,
    "top_slow_pages": [
      {
        "url": "/heavy-page",
        "avg_load_time": 4500.2,
        "avg_lcp": 3200.1,
        "visit_count": 245
      }
    ],
    "top_fast_pages": [
      {
        "url": "/light-page", 
        "avg_load_time": 850.3,
        "avg_lcp": 650.2,
        "visit_count": 1024
      }
    ],
    "device_breakdown": {
      "desktop": 1250,
      "mobile": 890,
      "tablet": 160
    },
    "error_count": 23,
    "critical_alerts": 5,
    "trend_data": [
      {
        "timestamp": "2025-07-21T08:00:00Z",
        "lcp": 1800.5,
        "fid": 65.2,
        "cls": 0.11,
        "ttfb": 420.3
      }
    ],
    "last_updated": "2025-07-21T09:30:00Z"
  },
  "timestamp": "2025-07-21T09:30:00Z"
}
```

### 3. 获取性能告警列表

**接口地址**: `GET /api/v1/metrics/alerts`

**描述**: 获取活跃的性能告警信息

**需要认证**: JWT Token

**查询参数**:
- `limit`: 返回记录数限制，默认50，最大1000

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "alerts": [
      {
        "id": 1,
        "metric_type": "lcp",
        "threshold": 2500.0,
        "value": 3200.5,
        "page_url": "/slow-page",
        "session_id": "session-456",
        "user_id": "user123",
        "severity": "high",
        "status": "active",
        "message": "LCP exceeded threshold: 3200.50 > 2500.00",
        "created_at": "2025-07-21T09:15:00Z"
      }
    ],
    "total": 15
  },
  "timestamp": "2025-07-21T09:30:00Z"
}
```

### 4. 创建性能告警

**接口地址**: `POST /api/v1/metrics/alerts`

**描述**: 手动创建性能告警

**需要认证**: JWT Token

**请求体**:
```json
{
  "metric_type": "fid",
  "threshold": 100.0,
  "value": 150.5,
  "page_url": "/interactive-page",
  "session_id": "session-789",
  "severity": "medium",
  "message": "FID response time too slow"
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "metric_type": "fid",
    "severity": "medium",
    "timestamp": "2025-07-21T09:30:00Z"
  },
  "timestamp": "2025-07-21T09:30:00Z"
}
```

### 5. 获取服务健康状态

**接口地址**: `GET /api/v1/health/status`

**描述**: 获取系统和各微服务的健康状态

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "gateway": {
      "status": "healthy",
      "uptime": "24h 15m 30s",
      "version": "1.0.0",
      "cpu_usage": "15%",
      "memory": "256MB/1GB"
    },
    "services": {
      "main-backend": {
        "status": "healthy",
        "response_time": "45ms"
      },
      "write-service": {
        "status": "healthy", 
        "response_time": "38ms"
      },
      "courier-service": {
        "status": "healthy",
        "response_time": "52ms"
      }
    },
    "database": {
      "status": "healthy",
      "connection_pool": "8/10",
      "query_time": "15ms"
    },
    "redis": {
      "status": "healthy",
      "memory_usage": "45%",
      "connections": "12/100"
    },
    "timestamp": "2025-07-21T09:30:00Z"
  },
  "timestamp": "2025-07-21T09:30:00Z"
}
```

## 🔒 认证说明

- `/api/v1/metrics/performance` - 无需认证（前端直接上报）
- 其他API需要在请求头中携带JWT Token：
  ```
  Authorization: Bearer <your-jwt-token>
  ```

## 📊 Core Web Vitals 指标说明

- **LCP (Largest Contentful Paint)**: 最大内容绘制时间，单位毫秒
  - 优秀: ≤ 2500ms
  - 需要改进: 2500ms - 4000ms  
  - 差: > 4000ms

- **FID (First Input Delay)**: 首次输入延迟，单位毫秒
  - 优秀: ≤ 100ms
  - 需要改进: 100ms - 300ms
  - 差: > 300ms

- **CLS (Cumulative Layout Shift)**: 累积布局偏移，无单位
  - 优秀: ≤ 0.1
  - 需要改进: 0.1 - 0.25
  - 差: > 0.25

- **TTFB (Time to First Byte)**: 首字节时间，单位毫秒
  - 优秀: ≤ 800ms
  - 需要改进: 800ms - 1800ms
  - 差: > 1800ms

## 🚨 错误码说明

- `400`: 请求参数错误
- `401`: 未认证或Token无效
- `403`: 权限不足
- `429`: 请求频率过高，触发限流
- `500`: 服务器内部错误
- `503`: 服务不可用

## 📈 与Agent #1前端集成指南

前端需要在适当的时机调用性能监控API：

```javascript
// 1. 页面加载完成后上报性能数据
window.addEventListener('load', () => {
  const perfData = {
    session_id: generateSessionId(),
    page_url: window.location.pathname,
    lcp: getLCP(),
    fid: getFID(), 
    cls: getCLS(),
    ttfb: getTTFB(),
    // ... 其他指标
  };
  
  fetch('/api/v1/metrics/performance', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(perfData)
  });
});

// 2. 仪表板页面获取聚合数据
const getDashboardData = async (timeRange = '24h') => {
  const response = await fetch(`/api/v1/metrics/dashboard?time_range=${timeRange}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return response.json();
};
```

## 🔄 与Agent #4管理后台集成

管理后台可以通过这些API获取性能监控数据，用于管理面板展示和告警管理。

**此API系统已完全实现并可投入使用！** ✅