# 批量操作API文档

## 概述

批量操作API提供了对OpenPenPal写信服务中各种资源的批量管理功能，支持高效的批量处理操作，包括删除、更新、导出、归档等。

## 特性

- 🚀 **高性能批量处理** - 支持最多1000个目标的批量操作
- 🔄 **实时进度跟踪** - WebSocket实时进度通知
- 🧪 **试运行验证** - dry_run模式验证操作合法性
- 🛡️ **权限控制** - 细粒度权限验证
- 📊 **详细日志** - 完整的操作结果记录
- 🔍 **灵活过滤** - 支持多种导出格式和字段过滤
- ⚡ **异步处理** - 后台异步执行，提高系统响应性

## 支持的操作类型

| 操作类型 | 描述 | 支持的目标 |
|----------|------|------------|
| `delete` | 批量删除（支持软删除） | letters, products, drafts, plaza_posts |
| `update` | 批量更新字段 | letters, products, drafts |
| `status_update` | 批量状态更新 | letters, products, plaza_posts |
| `export` | 批量导出数据 | 所有目标类型 |
| `archive` | 批量归档 | letters, products, plaza_posts |
| `restore` | 批量恢复 | letters, products, plaza_posts |
| `bulk_create` | 批量创建 | letters, products, drafts |

## 支持的目标类型

- `letters` - 信件
- `products` - 商品
- `orders` - 订单
- `users` - 用户
- `drafts` - 草稿
- `plaza_posts` - 广场帖子
- `museum_items` - 博物馆物品

## API端点

### 基础操作

#### 1. 执行批量操作
```http
POST /api/batch/execute
```

**请求体:**
```json
{
  "operation": "delete",
  "target_type": "letters",
  "target_ids": ["LETTER001", "LETTER002"],
  "operation_data": {
    "soft_delete": true,
    "delete_reason": "批量清理"
  },
  "dry_run": false
}
```

**响应:**
```json
{
  "code": 0,
  "msg": "批量操作执行成功",
  "data": {
    "operation_id": "uuid-string",
    "operation": "delete",
    "target_type": "letters",
    "total_count": 2,
    "success_count": 2,
    "failure_count": 0,
    "results": [
      {
        "target_id": "LETTER001",
        "success": true,
        "message": "Deleted successfully"
      }
    ],
    "started_at": "2024-01-20T10:00:00Z",
    "completed_at": "2024-01-20T10:00:05Z",
    "duration_ms": 5000,
    "dry_run": false
  }
}
```

#### 2. 验证批量操作（试运行）
```http
POST /api/batch/validate
```

**请求体:**
```json
{
  "operation": "delete",
  "target_type": "letters",
  "target_ids": ["LETTER001", "LETTER002"],
  "operation_data": {
    "soft_delete": true
  }
}
```

### 专用操作端点

#### 3. 批量删除信件
```http
POST /api/batch/letters/delete
```

**请求体:**
```json
{
  "target_ids": ["LETTER001", "LETTER002"],
  "soft_delete": true,
  "delete_reason": "批量清理"
}
```

#### 4. 批量更新信件状态
```http
POST /api/batch/letters/status
```

**请求体:**
```json
{
  "target_ids": ["LETTER001", "LETTER002"],
  "new_status": "generated",
  "reason": "批量生成编码",
  "force": false
}
```

#### 5. 批量删除商品
```http
POST /api/batch/products/delete
```

#### 6. 批量更新商品状态
```http
POST /api/batch/products/status
```

#### 7. 批量归档
```http
POST /api/batch/archive?target_type=letters
```

**请求体:**
```json
{
  "target_ids": ["LETTER001", "LETTER002"],
  "archive_reason": "定期归档",
  "archive_location": "archive/2024"
}
```

#### 8. 批量恢复
```http
POST /api/batch/restore?target_type=letters
```

**请求体:**
```json
{
  "target_ids": ["LETTER001", "LETTER002"]
}
```

#### 9. 批量导出
```http
POST /api/batch/export?target_type=letters
```

**请求体:**
```json
{
  "target_ids": ["LETTER001", "LETTER002"],
  "export_format": "json",
  "include_fields": ["id", "title", "status"],
  "exclude_fields": ["content"]
}
```

#### 10. 批量创建
```http
POST /api/batch/bulk-create?target_type=letters
```

**请求体:**
```json
{
  "items": [
    {
      "title": "新信件1",
      "content": "信件内容1",
      "anonymous": false
    },
    {
      "title": "新信件2", 
      "content": "信件内容2",
      "anonymous": true
    }
  ],
  "skip_validation": false,
  "continue_on_error": true
}
```

### 作业管理

#### 11. 获取作业状态
```http
GET /api/batch/jobs/{job_id}
```

**响应:**
```json
{
  "code": 0,
  "msg": "获取作业状态成功",
  "data": {
    "job_id": "uuid-string",
    "status": "completed",
    "progress": 100.0,
    "current_item": null,
    "estimated_remaining": null,
    "created_at": "2024-01-20T10:00:00Z",
    "updated_at": "2024-01-20T10:00:05Z",
    "error_message": null
  }
}
```

#### 12. 取消作业
```http
POST /api/batch/jobs/{job_id}/cancel
```

#### 13. 下载导出文件
```http
GET /api/batch/export/{export_id}
```

### 服务管理

#### 14. 服务健康检查
```http
GET /api/batch/health
```

**响应:**
```json
{
  "code": 0,
  "msg": "批量操作服务运行正常",
  "data": {
    "status": "healthy",
    "active_jobs": 0,
    "supported_operations": ["delete", "update", "export", "..."],
    "supported_targets": ["letters", "products", "..."],
    "timestamp": "2024-01-20T10:00:00Z"
  }
}
```

### 管理员功能

#### 15. 获取所有作业状态（管理员）
```http
GET /api/batch/admin/jobs
```

#### 16. 清理已完成作业（管理员）
```http
POST /api/batch/admin/cleanup?older_than_hours=24
```

## 权限控制

### 用户权限
- 只能对自己创建的资源进行批量操作
- 不能执行管理员专用功能

### 管理员权限
- 可以对所有资源执行批量操作
- 可以强制删除和归档
- 可以查看和管理所有作业
- 可以执行系统清理

### 权限检查机制
```python
# 自动权限检查（基于资源所有者）
if hasattr(target, 'sender_id') or hasattr(target, 'user_id'):
    owner_id = getattr(target, 'sender_id', None) or getattr(target, 'user_id', None)
    if owner_id != current_user:
        # 权限拒绝
        
# 管理员权限检查
if request.operation in [BatchOperationEnum.DELETE, BatchOperationEnum.ARCHIVE]:
    await check_admin_permission(current_user)
```

## 错误处理

### 常见错误码

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数无效 | 检查请求格式和参数 |
| 401 | 未认证 | 提供有效的JWT token |
| 403 | 权限不足 | 确认用户权限或联系管理员 |
| 404 | 资源不存在 | 检查目标ID是否正确 |
| 429 | 请求过于频繁 | 等待后重试 |
| 500 | 服务器内部错误 | 联系技术支持 |

### 操作级别错误

每个批量操作的单个项目可能成功或失败，详细结果在`results`数组中：

```json
{
  "results": [
    {
      "target_id": "LETTER001",
      "success": true,
      "message": "操作成功"
    },
    {
      "target_id": "LETTER002",
      "success": false,
      "message": "权限不足",
      "error_code": "PERMISSION_DENIED"
    }
  ]
}
```

## WebSocket通知

批量操作支持实时进度通知，客户端可以监听WebSocket事件：

```javascript
// 监听批量操作进度
websocket.on('batch_operation_progress', (data) => {
  console.log(`操作进度: ${data.progress}%`);
  console.log(`状态: ${data.status}`);
  console.log(`成功/失败: ${data.success_count}/${data.failure_count}`);
});
```

## 使用示例

### Python客户端示例

```python
import aiohttp
import asyncio

async def batch_delete_letters():
    headers = {
        "Authorization": "Bearer your-jwt-token",
        "Content-Type": "application/json"
    }
    
    payload = {
        "target_ids": ["LETTER001", "LETTER002"],
        "soft_delete": True,
        "delete_reason": "批量清理"
    }
    
    async with aiohttp.ClientSession(headers=headers) as session:
        async with session.post(
            "http://localhost:8001/api/batch/letters/delete",
            json=payload
        ) as response:
            result = await response.json()
            print(f"批量删除结果: {result}")
            return result

# 运行示例
asyncio.run(batch_delete_letters())
```

### JavaScript客户端示例

```javascript
const batchUpdateStatus = async () => {
  const response = await fetch('/api/batch/letters/status', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer your-jwt-token',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      target_ids: ['LETTER001', 'LETTER002'],
      new_status: 'generated',
      reason: '批量生成编码'
    })
  });
  
  const result = await response.json();
  console.log('批量状态更新结果:', result);
  return result;
};
```

## 最佳实践

### 1. 批量大小控制
- 建议单次批量操作不超过100个目标
- 大量数据可分批处理

### 2. 错误处理
- 始终检查操作结果中的`success_count`和`failure_count`
- 处理单个项目的失败情况
- 记录失败的目标ID用于重试

### 3. 权限管理
- 在批量操作前验证权限
- 使用试运行模式预检查操作合法性

### 4. 性能优化
- 使用WebSocket监听进度，避免轮询
- 合理设置超时时间
- 考虑在低峰期执行大批量操作

### 5. 数据安全
- 导出敏感数据时使用`exclude_fields`排除敏感字段
- 软删除优于硬删除
- 归档重要数据而非删除

## 监控和日志

### 操作审计
所有批量操作都会记录详细日志：
- 操作用户和时间
- 操作类型和目标
- 执行结果和耗时
- 失败原因和错误码

### 性能指标
- 操作成功率
- 平均执行时间
- 并发作业数量
- 系统资源使用情况

## 故障排除

### 1. 操作超时
- 检查目标数量是否过多
- 验证网络连接稳定性
- 确认服务器资源充足

### 2. 权限错误
- 验证JWT token有效性
- 检查用户角色和权限
- 确认目标资源所有权

### 3. 操作失败
- 查看详细错误信息
- 检查目标资源状态
- 验证操作数据格式

### 4. 性能问题
- 减少批量大小
- 避免高峰期操作
- 使用更高效的过滤条件

## 版本兼容性

当前API版本：v1.0.0

主要版本变更将通过以下方式通知：
- API响应头中的版本信息
- 文档更新和迁移指南
- 向后兼容性支持期

## 技术架构

### 核心组件
- **BatchOperationService** - 批量操作服务核心
- **BatchOperationRequest/Response** - 请求/响应模型
- **权限验证** - 基于角色的访问控制
- **WebSocket通知** - 实时进度更新
- **缓存管理** - 导出结果缓存

### 依赖服务
- PostgreSQL - 数据持久化
- Redis - 缓存和作业状态
- WebSocket - 实时通知
- JWT认证 - 用户身份验证

---

*本文档持续更新，如有问题请联系开发团队。*