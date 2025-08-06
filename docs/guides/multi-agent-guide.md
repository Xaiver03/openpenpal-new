# 多Agent上下文获取指南

> **作者**: Agent #1 (队长)  
> **目的**: 确保所有Agent都能获取必要的开发上下文

## 🎯 核心原则

不同Agent获取上下文的关键是**文档驱动**和**规范先行**。每个Agent都应该能够通过阅读项目文档快速理解整体架构和自己的职责。

## 📋 Agent上下文获取策略

### 1. **必读文档清单** (所有Agent都需要)

```bash
# 核心文档
- /MULTI_AGENT_COORDINATION.md      # 协同开发框架
- /docs/api/UNIFIED_API_SPECIFICATION.md  # API规范
- /AGENT_CONTEXT_MANAGEMENT.md      # 共享配置信息
- /agent-tasks/AGENT-{N}-*.md       # 自己的任务卡片
```

### 2. **上下文传递方式**

#### 方式一：通过任务卡片 (推荐)
每个Agent的任务卡片包含完整的开发要求：
```yaml
# agent-tasks/AGENT-2-WRITE-SERVICE.md
task_name: 写信服务模块
dependencies:
  - JWT认证 (从网关获取)
  - WebSocket事件推送
  - PostgreSQL数据库
interfaces:
  - POST /api/letters
  - GET /api/letters/{id}
```

#### 方式二：通过共享配置文件
```bash
# 读取共享配置
cat AGENT_CONTEXT_MANAGEMENT.md

# 获取数据库连接信息
# 获取其他服务的端口和API前缀
# 了解认证机制和WebSocket事件
```

#### 方式三：通过API文档
```bash
# 查看统一API规范
cat docs/api/UNIFIED_API_SPECIFICATION.md

# 了解响应格式、状态码、认证方式
```

### 3. **实际操作示例**

#### Agent #2 开始开发时的上下文获取
```bash
# 1. 阅读自己的任务卡片
任务：开发写信服务
技术栈：Python FastAPI
端口：8001
API前缀：/api/letters

# 2. 了解认证集成
从 AGENT_CONTEXT_MANAGEMENT.md 获取：
- JWT认证方式
- Token验证endpoint

# 3. 了解数据库schema
从已有代码获取：
- backend/internal/models/letter.go
- 字段定义和关系

# 4. 了解WebSocket事件
需要推送的事件：
- LETTER_STATUS_UPDATE
- 格式和推送方式
```

#### Agent #3 开始开发时的上下文获取
```bash
# 1. 阅读任务卡片
任务：信使任务调度系统
依赖：写信服务的API
需要调用：GET /api/letters/{id}

# 2. 跨服务通信
从 Agent #2 的服务获取信件信息
使用统一的响应格式处理结果

# 3. 事件推送
推送 COURIER_TASK_ASSIGNMENT 事件
订阅 LETTER_STATUS_UPDATE 事件
```

### 4. **开发时的上下文同步**

#### 使用Git分支管理
```bash
# 每个Agent在自己的分支开发
git checkout -b feature/agent-2-write-service
git checkout -b feature/agent-3-courier-service

# 定期合并到develop分支
git checkout develop
git merge feature/agent-2-write-service
```

#### 更新共享文档
当Agent完成某个功能时，需要更新：
1. API文档 (如果有新接口)
2. 共享配置 (如果有新的服务配置)
3. WebSocket事件文档 (如果有新事件)

### 5. **跨Agent通信模板**

#### HTTP调用模板
```python
# Agent #3 调用 Agent #2 的服务
import requests

class LetterServiceClient:
    def __init__(self):
        self.base_url = "http://localhost:8001"
        self.headers = {"Authorization": f"Bearer {token}"}
    
    def get_letter(self, letter_id: str):
        response = requests.get(
            f"{self.base_url}/api/letters/{letter_id}",
            headers=self.headers
        )
        # 使用统一响应格式
        data = response.json()
        if data["code"] == 0:
            return data["data"]
        raise Exception(data["msg"])
```

#### WebSocket事件订阅模板
```go
// Agent #3 订阅信件状态更新
func SubscribeLetterUpdates() {
    ws.Subscribe("LETTER_STATUS_UPDATE", func(event Event) {
        // 处理信件状态更新
        letterID := event.Data["letter_id"]
        status := event.Data["status"]
        // 更新本地任务状态
    })
}
```

### 6. **上下文检查清单**

每个Agent开发前确认：
- [ ] 已阅读 MULTI_AGENT_COORDINATION.md
- [ ] 已阅读 UNIFIED_API_SPECIFICATION.md
- [ ] 已理解自己的任务卡片
- [ ] 已了解依赖的其他服务
- [ ] 已知道需要推送/订阅的事件
- [ ] 已了解数据库schema
- [ ] 已设置正确的端口和API前缀

## 🔄 持续更新机制

### 文档更新责任
- **Agent #1 (队长)**: 维护总体架构文档
- **各Agent**: 更新自己模块的API文档
- **所有Agent**: 及时同步共享配置变更

### 每日同步会议 (建议)
```
1. 各Agent汇报进度
2. 讨论接口变更
3. 解决集成问题
4. 更新共享文档
```

## 🛠️ 工具支持

### 1. API Mock工具
当依赖的服务还未完成时，可以使用Mock：
```bash
# 启动Mock服务器
npx json-server --watch mock-data.json --port 8001
```

### 2. 文档生成
```bash
# 从代码生成API文档
# Python: 使用 FastAPI 自动生成
# Go: 使用 swag init
# Java: 使用 SpringDoc
```

### 3. 集成测试
```bash
# 运行跨服务集成测试
./scripts/multi-agent-dev.sh test
```

## 📚 最佳实践

1. **先读文档，再写代码**
2. **遵循统一规范，不要自创标准**
3. **及时更新文档，保持同步**
4. **使用标准化的错误处理**
5. **记录所有的API变更**
6. **保持向后兼容**

---

通过这套上下文管理机制，每个Agent都能：
- 快速了解项目全貌
- 明确自己的开发边界
- 知道如何与其他服务集成
- 保持开发的一致性