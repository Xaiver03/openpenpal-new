# OCR服务 API 文档

## 概述

OpenPenPal OCR服务提供强大的手写文字识别功能，支持多OCR引擎、图像预处理、缓存优化等特性。

**基础URL**: `http://localhost:8004`  
**认证方式**: JWT Bearer Token  
**数据格式**: JSON

## 目录

- [认证说明](#认证说明)
- [通用响应格式](#通用响应格式)
- [错误码说明](#错误码说明)
- [OCR识别接口](#ocr识别接口)
- [任务管理接口](#任务管理接口)
- [缓存管理接口](#缓存管理接口)
- [系统接口](#系统接口)

## 认证说明

所有API接口（除健康检查外）都需要JWT认证。

### 请求头格式
```http
Authorization: Bearer <jwt_token>
```

### JWT Payload示例
```json
{
  "user_id": "user_12345",
  "username": "xiaoming",
  "role": "user",
  "school_code": "BJFU",
  "iat": 1642684800,
  "exp": 1642771200
}
```

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    // 具体数据内容
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 错误响应
```json
{
  "code": 1,
  "msg": "参数错误",
  "data": null,
  "error": {
    "type": "validation_error",
    "details": "字段验证失败"
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## 错误码说明

| Code | 含义 | HTTP状态码 |
|------|------|------------|
| 0 | 成功 | 200/201 |
| 1 | 参数错误 | 400 |
| 2 | 无权限 | 403 |
| 3 | 资源不存在 | 404 |
| 4 | 业务逻辑错误 | 422 |
| 500 | 服务内部错误 | 500 |

## OCR识别接口

### 1. 图片OCR识别

**接口**: `POST /api/ocr/recognize`  
**描述**: 上传图片进行OCR文字识别

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | File | 是 | 图片文件 (jpg/png/jpeg，最大10MB) |
| language | String | 否 | 识别语言 (zh/en/auto，默认auto) |
| enhance | Boolean | 否 | 是否图像增强 (默认true) |
| confidence_threshold | Float | 否 | 置信度阈值 (0.0-1.0，默认0.7) |
| engine | String | 否 | OCR引擎 (paddle/tesseract/easyocr，默认paddle) |
| is_handwriting | Boolean | 否 | 是否手写模式 (默认false) |
| use_voting | Boolean | 否 | 是否使用多引擎投票 (默认false) |
| use_cache | Boolean | 否 | 是否使用缓存 (默认true) |

#### 请求示例

```bash
curl -X POST http://localhost:8004/api/ocr/recognize \
  -H "Authorization: Bearer <jwt_token>" \
  -F "image=@test.jpg" \
  -F "language=zh" \
  -F "enhance=true" \
  -F "is_handwriting=true"
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "识别成功",
  "data": {
    "task_id": "ocr_task_123456",
    "status": "completed",
    "results": {
      "text": "亲爱的朋友，\n最近过得怎么样？",
      "confidence": 0.85,
      "word_count": 12,
      "processing_time": 2.3,
      "language_detected": "zh",
      "from_cache": false,
      "blocks": [
        {
          "text": "亲爱的朋友，",
          "confidence": 0.92,
          "bbox": [45, 78, 180, 105],
          "line": 1
        }
      ]
    },
    "metadata": {
      "image_size": "1024x768",
      "image_format": "jpeg",
      "processing_method": "paddle_ocr",
      "enhancement_applied": true,
      "preprocessing_operations": ["denoise", "deskew", "handwriting_enhance"],
      "is_handwriting_mode": true,
      "image_hash": "abc123..."
    }
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 2. 批量OCR识别

**接口**: `POST /api/ocr/batch`  
**描述**: 批量上传图片进行OCR识别

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| images | File[] | 是 | 多个图片文件 (最多10个) |
| settings | String | 否 | JSON配置字符串 |

#### settings JSON格式
```json
{
  "language": "zh",
  "enhance": true,
  "engine": "paddle",
  "is_handwriting": false
}
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "批量任务已创建",
  "data": {
    "batch_id": "batch_789",
    "total_images": 5,
    "estimated_time": "30s",
    "status": "processing",
    "progress_url": "/api/ocr/tasks/batch/batch_789/progress"
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 3. 图像预处理和增强

**接口**: `POST /api/ocr/enhance`  
**描述**: 对图片进行预处理和增强

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image | File | 是 | 原始图片 |
| operations | String | 否 | 操作列表JSON (默认["denoise", "deskew", "contrast"]) |
| return_enhanced | Boolean | 否 | 是否返回增强后的图片 (默认false) |
| is_handwriting | Boolean | 否 | 是否手写模式 (默认false) |

#### 操作类型说明

| 操作 | 说明 |
|------|------|
| denoise | 降噪处理 |
| deskew | 倾斜矫正 |
| contrast | 对比度增强 |
| brightness | 亮度调整 |
| binarize | 二值化 |
| sharpen | 锐化 |
| handwriting_enhance | 手写文字专用增强 |

#### 响应示例

```json
{
  "code": 0,
  "msg": "图像增强完成",
  "data": {
    "enhanced_image_url": "/api/ocr/files/enhanced_123.jpg",
    "operations_applied": ["denoise", "deskew", "contrast"],
    "quality_score": 0.78,
    "enhancement_metrics": {
      "noise_reduction": 0.65,
      "contrast_improvement": 0.42,
      "psnr": 28.5
    },
    "original_size": [1024, 768],
    "is_handwriting_mode": false
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 4. 文本内容验证

**接口**: `POST /api/ocr/validate`  
**描述**: 验证OCR识别结果与原始文本的一致性

#### 请求参数

```json
{
  "original_text": "用户输入的原始文本",
  "ocr_text": "OCR识别的文本",
  "validation_rules": {
    "min_similarity": 0.8,
    "check_sensitive_words": true,
    "max_length": 5000
  }
}
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "验证完成",
  "data": {
    "is_valid": true,
    "similarity_score": 0.87,
    "issues": [],
    "suggestions": [
      {
        "type": "correction",
        "original": "问候",
        "suggested": "问候",
        "confidence": 0.95,
        "position": 45
      }
    ],
    "content_analysis": {
      "word_count": 156,
      "sentiment": "positive",
      "language": "zh",
      "contains_sensitive": false
    }
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 5. 获取可用OCR模型

**接口**: `GET /api/ocr/models`  
**描述**: 获取系统中可用的OCR引擎和模型信息

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "available_models": [
      {
        "name": "paddle",
        "available": true,
        "description": "百度PaddleOCR - 高精度中英文识别",
        "languages": ["zh", "en"],
        "accuracy": 0.95,
        "speed": "fast",
        "best_for": "printed_text"
      },
      {
        "name": "tesseract",
        "available": true,
        "description": "Tesseract OCR - 开源多语言识别",
        "languages": ["zh", "en", "ja"],
        "accuracy": 0.88,
        "speed": "medium",
        "best_for": "document_text"
      }
    ],
    "default_model": "paddle",
    "total_engines": 3,
    "available_engines": 2,
    "supports_gpu": false,
    "supports_voting": true
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## 任务管理接口

### 1. 获取任务状态

**接口**: `GET /api/ocr/tasks/{task_id}`  
**描述**: 查询OCR任务的执行状态

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "task_id": "ocr_task_123456",
    "status": "completed",
    "progress": 100,
    "created_at": "2025-07-20T11:58:30Z",
    "completed_at": "2025-07-20T12:00:00Z",
    "result": {
      "text": "识别结果文本内容...",
      "confidence": 0.85,
      "processing_time": 2.3
    },
    "error": null
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 2. 获取用户任务列表

**接口**: `GET /api/ocr/tasks/`  
**描述**: 获取当前用户的所有OCR任务

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | Integer | 否 | 页码 (默认1) |
| limit | Integer | 否 | 每页数量 (默认10，最大50) |
| status | String | 否 | 状态过滤 (processing/completed/failed) |

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "task_id": "ocr_task_123",
        "status": "completed",
        "created_at": "2025-07-20T12:00:00Z",
        "processing_time": 2.3,
        "confidence": 0.85
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "pages": 3,
      "has_next": true,
      "has_prev": false
    }
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 3. 删除任务

**接口**: `DELETE /api/ocr/tasks/{task_id}`  
**描述**: 删除指定的OCR任务

#### 响应示例

```json
{
  "code": 0,
  "msg": "任务删除成功",
  "data": {
    "task_id": "ocr_task_123456"
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 4. 获取批量任务进度

**接口**: `GET /api/ocr/tasks/batch/{batch_id}/progress`  
**描述**: 获取批量OCR任务的执行进度

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "batch_id": "batch_789",
    "total_images": 5,
    "completed_images": 3,
    "failed_images": 0,
    "progress_percentage": 60,
    "status": "processing",
    "estimated_time_remaining": "30s",
    "current_step": "正在处理第4张图片",
    "results": [
      {
        "image_name": "image1.jpg",
        "status": "completed",
        "text": "识别结果1",
        "confidence": 0.89
      }
    ]
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## 缓存管理接口

### 1. 获取缓存统计

**接口**: `GET /api/ocr/cache/stats`  
**描述**: 获取OCR服务的缓存使用统计信息

#### 响应示例

```json
{
  "code": 0,
  "msg": "缓存统计信息获取成功",
  "data": {
    "cache_type": "redis",
    "connected": true,
    "used_memory": "15.2M",
    "total_keys": 1250,
    "hits": 8540,
    "misses": 1230
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 2. 清理缓存

**接口**: `POST /api/ocr/cache/clear`  
**描述**: 清理OCR服务缓存 (需要管理员权限)

#### 请求参数

```json
{
  "pattern": "ocr_result:*"  // 可选，清理匹配的键
}
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "缓存清理成功",
  "data": {
    "pattern": "ocr_result:*",
    "message": "缓存清理完成"
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## 系统接口

### 1. 健康检查

**接口**: `GET /health`  
**描述**: 检查服务健康状态 (无需认证)

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "service": "ocr-service",
    "status": "healthy",
    "version": "1.0.0"
  },
  "timestamp": "2025-07-20T12:00:00Z"
}
```

### 2. Ping检查

**接口**: `GET /ping`  
**描述**: 简单的连通性检查 (无需认证)

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": "pong",
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## WebSocket事件推送

OCR服务支持通过WebSocket推送实时进度更新。

### 事件类型

| 事件类型 | 说明 |
|----------|------|
| OCR_PROGRESS_UPDATE | OCR识别进度更新 |
| OCR_TASK_COMPLETED | OCR任务完成 |
| OCR_BATCH_PROGRESS | 批量OCR进度更新 |
| IMAGE_ENHANCEMENT_PROGRESS | 图像增强进度 |
| SYSTEM_NOTIFICATION | 系统通知 |

### 事件格式

```json
{
  "type": "OCR_PROGRESS_UPDATE",
  "data": {
    "task_id": "ocr_task_123456",
    "progress": 75,
    "status": "processing",
    "current_step": "正在执行文字识别",
    "estimated_time_remaining": 10
  },
  "user_id": "user_12345",
  "timestamp": "2025-07-20T12:00:00Z"
}
```

## 使用示例

### Python示例

```python
import requests
import json

# 配置
BASE_URL = "http://localhost:8004"
JWT_TOKEN = "your-jwt-token"

headers = {
    "Authorization": f"Bearer {JWT_TOKEN}"
}

# 单张图片OCR识别
def ocr_recognize(image_path):
    url = f"{BASE_URL}/api/ocr/recognize"
    
    files = {"image": open(image_path, "rb")}
    data = {
        "language": "zh",
        "enhance": "true",
        "is_handwriting": "true"
    }
    
    response = requests.post(url, headers=headers, files=files, data=data)
    return response.json()

# 获取任务状态
def get_task_status(task_id):
    url = f"{BASE_URL}/api/ocr/tasks/{task_id}"
    response = requests.get(url, headers=headers)
    return response.json()

# 使用示例
result = ocr_recognize("test.jpg")
print(json.dumps(result, indent=2, ensure_ascii=False))
```

### JavaScript示例

```javascript
const BASE_URL = 'http://localhost:8004';
const JWT_TOKEN = 'your-jwt-token';

// OCR识别
async function ocrRecognize(imageFile) {
  const formData = new FormData();
  formData.append('image', imageFile);
  formData.append('language', 'zh');
  formData.append('enhance', 'true');
  
  const response = await fetch(`${BASE_URL}/api/ocr/recognize`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`
    },
    body: formData
  });
  
  return await response.json();
}

// 获取任务状态
async function getTaskStatus(taskId) {
  const response = await fetch(`${BASE_URL}/api/ocr/tasks/${taskId}`, {
    headers: {
      'Authorization': `Bearer ${JWT_TOKEN}`
    }
  });
  
  return await response.json();
}
```

## 最佳实践

### 1. 图片优化建议

- **格式**: 推荐使用JPG/PNG格式
- **分辨率**: 建议300-600 DPI
- **大小**: 控制在10MB以内
- **质量**: 确保文字清晰，对比度高

### 2. 参数选择建议

- **手写文字**: 设置 `is_handwriting=true`，使用 `tesseract` 或 `easyocr` 引擎
- **印刷文字**: 使用 `paddle` 引擎，性能最佳
- **多语言**: 使用 `tesseract` 引擎，支持语言最多
- **高精度**: 启用 `use_voting=true` 使用多引擎投票

### 3. 性能优化

- **缓存**: 保持 `use_cache=true` 避免重复识别
- **预处理**: 对模糊图片启用 `enhance=true`
- **批量处理**: 大量图片使用批量接口
- **异步处理**: 使用WebSocket监听进度

### 4. 错误处理

- **重试机制**: 对临时错误进行重试
- **降级处理**: 在高精度模式失败时降级到基础模式
- **超时处理**: 设置合理的请求超时时间
- **错误日志**: 记录详细的错误信息用于调试

---

如有问题，请查看系统日志或联系技术支持。