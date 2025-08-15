# 积分系统开发者指南

> 版本: v1.0
> 更新时间: 2024-01-21

## 一、快速开始

### 1.1 添加新的积分奖励

要为新的用户行为添加积分奖励，需要：

1. **定义积分常量** (`backend/internal/services/credit_service.go`):
```go
const (
    PointsNewBehavior = 10  // 新行为的积分值
)
```

2. **创建奖励方法**:
```go
func (s *CreditService) RewardNewBehavior(userID string, description string) error {
    return s.createCreditTask(models.CreditTaskType("new_behavior"), userID, PointsNewBehavior, description, "")
}
```

3. **在业务逻辑中调用**:
```go
// 在相应的service中
if s.creditTaskSvc != nil {
    go func() {
        if err := s.creditTaskSvc.TriggerNewBehaviorReward(userID, referenceID); err != nil {
            log.Printf("Failed to trigger reward: %v", err)
        }
    }()
}
```

### 1.2 前端集成积分显示

在任何需要显示用户积分的组件中：

```tsx
import { useCreditInfo } from '@/stores/credit-store'
import { CreditLevelBadge } from '@/components/credit/credit-level-badge'
import { formatPoints } from '@/lib/api/credit'

export function MyComponent() {
  const { userCredit, loading } = useCreditInfo()
  
  if (loading) return <div>加载中...</div>
  
  return (
    <div>
      <p>当前积分: {formatPoints(userCredit?.total || 0)}</p>
      <CreditLevelBadge 
        level={userCredit?.level || 1}
        totalPoints={userCredit?.total || 0}
      />
    </div>
  )
}
```

## 二、核心概念

### 2.1 积分任务系统

积分不是直接添加的，而是通过任务系统处理：

```
用户行为 → 创建任务 → 后台处理 → 更新积分 → 通知用户
```

**任务状态流转**:
- `pending`: 待处理
- `executing`: 处理中
- `completed`: 已完成
- `failed`: 失败（可重试）

### 2.2 积分等级计算

等级根据总积分自动计算：

```go
// backend/internal/models/user_credit.go
var LEVEL_UP_POINTS = []int{0, 100, 300, 600, 1000, 1500, 2500, 4000, 6000, 10000}

func CalculateLevel(totalPoints int) int {
    for i := len(LEVEL_UP_POINTS) - 1; i >= 0; i-- {
        if totalPoints >= LEVEL_UP_POINTS[i] {
            return i + 1
        }
    }
    return 1
}
```

### 2.3 事务类型

系统支持两种事务类型：
- `earn`: 获得积分
- `spend`: 消费积分（待实现）

## 三、API接口

### 3.1 获取用户积分
```
GET /api/v1/credits/user/:userId
```

响应示例：
```json
{
  "user_id": "user123",
  "total": 1250,
  "balance": 1250,
  "level": 4,
  "level_progress": 25.0,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 3.2 获取积分历史
```
GET /api/v1/credits/history?page=1&limit=20&type=earn
```

### 3.3 获取积分任务
```
GET /api/v1/credits/tasks?status=completed&page=1&limit=20
```

### 3.4 获取排行榜
```
GET /api/v1/credits/leaderboard?time_range=month&limit=20
```

## 四、最佳实践

### 4.1 异步处理

积分奖励应该异步处理，避免阻塞主业务流程：

```go
// ✅ 正确做法
go func() {
    if err := creditService.RewardAction(userID); err != nil {
        log.Printf("Credit reward failed: %v", err)
    }
}()

// ❌ 错误做法
if err := creditService.RewardAction(userID); err != nil {
    return err  // 不要因为积分失败而中断业务
}
```

### 4.2 幂等性

确保重复的请求不会导致重复奖励：

```go
// 使用reference字段确保幂等性
func (s *CreditTaskService) CreateTask(taskType, userID, reference string) error {
    // 检查是否已存在相同reference的任务
    var existing models.CreditTask
    if err := s.db.Where("reference = ?", reference).First(&existing).Error; err == nil {
        return nil // 任务已存在，直接返回
    }
    // 创建新任务...
}
```

### 4.3 错误处理

积分系统的错误不应影响主业务：

```tsx
// 前端错误处理
const handleAction = async () => {
  try {
    await performMainAction()
    // 积分奖励在后台自动处理
  } catch (error) {
    // 只处理主业务错误
    showError('操作失败')
  }
}
```

## 五、调试技巧

### 5.1 查看任务队列

```sql
-- 查看待处理任务
SELECT * FROM credit_tasks WHERE status = 'pending' ORDER BY created_at DESC;

-- 查看失败任务
SELECT * FROM credit_tasks WHERE status = 'failed' AND attempts > 1;

-- 查看用户积分历史
SELECT * FROM credit_transactions WHERE user_id = 'USER_ID' ORDER BY created_at DESC;
```

### 5.2 手动触发任务处理

```go
// 在测试环境中手动处理任务
taskService := services.NewCreditTaskService(db)
taskService.ProcessPendingTasks()
```

### 5.3 前端调试

```tsx
// 在浏览器控制台查看积分状态
const creditStore = window.__CREDIT_STORE__
console.log(creditStore.getState())

// 手动刷新积分数据
creditStore.getState().refreshAll()
```

## 六、常见问题

### Q1: 积分没有及时更新？
A: 积分通过后台任务处理，可能有1-2秒延迟。可以通过WebSocket实现实时更新。

### Q2: 如何处理积分扣除？
A: 使用负数amount创建spend类型的transaction：
```go
s.AddTransaction(userID, "spend", -100, "兑换奖品")
```

### Q3: 如何实现每日限制？
A: 在创建任务时检查当日已完成任务数：
```go
var count int64
s.db.Model(&CreditTask{}).
    Where("user_id = ? AND task_type = ? AND DATE(created_at) = CURDATE()", userID, taskType).
    Count(&count)
if count >= dailyLimit {
    return ErrDailyLimitExceeded
}
```

## 七、扩展开发

### 7.1 添加新的任务类型

1. 在`credit_task.go`中添加新类型：
```go
const (
    TaskTypeNewFeature CreditTaskType = "new_feature"
)
```

2. 在`TASK_DESCRIPTIONS`中添加描述：
```tsx
export const TASK_DESCRIPTIONS: Record<string, string> = {
  'new_feature': '新功能奖励',
}
```

### 7.2 自定义积分规则

创建规则引擎接口：
```go
type CreditRule interface {
    CanApply(userID string, context map[string]interface{}) bool
    GetPoints() int
    GetDescription() string
}
```

### 7.3 积分活动系统

实现活动倍数：
```go
type CreditActivity struct {
    ID         string
    Multiplier float64
    StartTime  time.Time
    EndTime    time.Time
}

func (s *CreditService) GetActiveMultiplier() float64 {
    // 获取当前活动倍数
}
```

## 八、性能优化

### 8.1 缓存策略

```go
// 使用Redis缓存用户积分
func (s *CreditService) GetUserCreditCached(userID string) (*UserCredit, error) {
    key := fmt.Sprintf("credit:user:%s", userID)
    // 先从缓存获取
    if cached, err := s.redis.Get(key); err == nil {
        return parseUserCredit(cached), nil
    }
    // 从数据库获取并缓存
}
```

### 8.2 批量处理

```go
// 批量创建任务
func (s *CreditTaskService) CreateBatchTasks(tasks []CreditTask) error {
    return s.db.CreateInBatches(tasks, 100).Error
}
```

### 8.3 查询优化

```sql
-- 添加索引
CREATE INDEX idx_credit_tasks_status_created ON credit_tasks(status, created_at);
CREATE INDEX idx_credit_transactions_user_created ON credit_transactions(user_id, created_at DESC);
CREATE INDEX idx_user_credits_level_total ON user_credits(level, total DESC);
```

## 九、安全考虑

### 9.1 防作弊

- 使用reference字段防止重复奖励
- 记录IP和设备信息
- 异常行为检测（短时间大量积分）

### 9.2 权限控制

```go
// 只有管理员可以手动调整积分
func (h *CreditHandler) AdjustCredit(c *gin.Context) {
    if !isAdmin(c) {
        c.JSON(403, gin.H{"error": "权限不足"})
        return
    }
    // 处理积分调整...
}
```

### 9.3 审计日志

```go
type CreditAuditLog struct {
    ID         string
    OperatorID string
    Action     string
    TargetUser string
    Amount     int
    Reason     string
    CreatedAt  time.Time
}
```

## 十、监控告警

### 10.1 关键指标

- 任务队列长度
- 任务处理延迟
- 失败任务数量
- 异常积分波动

### 10.2 告警规则

```yaml
alerts:
  - name: credit_task_queue_high
    condition: credit_tasks_pending > 1000
    message: "积分任务队列过长"
    
  - name: credit_task_failure_rate
    condition: credit_tasks_failed_rate > 0.05
    message: "积分任务失败率过高"
```

## 联系方式

如有问题，请联系：
- 技术负责人：[联系方式]
- 文档维护：[联系方式]