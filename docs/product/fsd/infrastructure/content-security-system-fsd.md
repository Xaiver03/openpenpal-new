# OpenPenPal 内容安全系统（XSS防护与敏感词过滤）FSD

## 一、系统概述

### 1.1 定位与目标

内容安全系统是OpenPenPal平台的核心安全防护组件，负责：
- **XSS攻击防护**：防止恶意脚本注入，保护用户安全
- **敏感词过滤**：维护平台内容健康，符合监管要求
- **内容风险评估**：多维度评估内容安全性，提供风险分数

### 1.2 什么是XSS防护？

XSS（Cross-Site Scripting，跨站脚本攻击）是一种常见的Web安全漏洞。攻击者通过在网页中注入恶意脚本代码，当其他用户浏览该页面时，嵌入的脚本会在用户的浏览器中执行，从而：
- 窃取用户的Cookie和会话信息
- 修改网页内容，进行钓鱼攻击
- 重定向用户到恶意网站
- 执行其他恶意操作

### 1.3 什么是内容过滤？

内容过滤是指对用户生成的内容进行检查和清理，包括：
- **敏感词检测**：识别并处理违规、不当词汇
- **HTML标签清理**：移除或转义危险的HTML标签
- **URL协议验证**：防止恶意链接注入
- **内容合规性检查**：确保内容符合平台规范

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────┐
│                  用户输入内容                      │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│           内容安全服务（ContentSecurityService）    │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐│
│  │ XSS检测引擎  │  │ HTML清理器   │  │敏感词过滤││
│  └─────────────┘  └─────────────┘  └──────────┘│
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐│
│  │ 风险评分器   │  │ 安全验证器   │  │事件记录器││
│  └─────────────┘  └─────────────┘  └──────────┘│
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│                  处理结果                         │
│  - 清理后的安全内容                               │
│  - 风险评分（0-100）                             │
│  - 违规类型和详情                                │
└─────────────────────────────────────────────────┘
```

### 2.2 核心组件

| 组件名称 | 职责 | 关键技术 |
|---------|------|---------|
| XSS检测引擎 | 识别恶意脚本模式 | 正则表达式、模式匹配 |
| HTML清理器 | 清理危险HTML标签 | bluemonday库 |
| 敏感词过滤 | 匹配和处理敏感词 | Aho-Corasick算法 |
| 风险评分器 | 综合评估内容风险 | 多因子加权算法 |
| 事件记录器 | 记录安全事件 | 数据库日志 |

---

## 三、功能设计

### 3.1 XSS防护功能

#### 3.1.1 检测模式

```go
// XSS攻击模式示例
xssPatterns := []string{
    `<script.*?>.*?</script>`,           // Script标签
    `javascript:`,                       // JavaScript协议
    `on\w+\s*=`,                        // 事件处理器
    `<iframe.*?>`,                      // iframe注入
    `<object.*?>`,                      // object标签
    `<embed.*?>`,                       // embed标签
    `vbscript:`,                        // VBScript协议
    `data:.*base64`,                    // base64数据URL
}
```

#### 3.1.2 清理策略

**标准模式（UGC内容）**：
- 保留安全的格式化标签：`<b>`, `<i>`, `<em>`, `<strong>`, `<br>`, `<p>`等
- 移除所有脚本相关标签
- 清理危险属性（onclick, onerror等）

**严格模式（高安全要求）**：
- 移除所有HTML标签
- 仅保留纯文本内容

### 3.2 敏感词管理

#### 3.2.1 敏感词分级

| 级别 | 说明 | 处理方式 |
|-----|------|---------|
| low | 低风险词汇 | 记录但不拦截 |
| medium | 中风险词汇 | 标记审核 |
| high | 高风险词汇 | 自动替换为*** |
| block | 屏蔽词汇 | 直接拒绝内容 |

#### 3.2.2 敏感词分类

- **spam**：垃圾信息
- **inappropriate**：不当内容
- **offensive**：冒犯性内容
- **political**：政治敏感
- **violence**：暴力内容
- **advertisement**：广告营销
- **other**：其他

### 3.3 权限控制

**可管理敏感词的角色**：
- 四级信使（courier_level4）：城市总代
- 平台管理员（platform_admin）
- 超级管理员（super_admin）

---

## 四、数据模型

### 4.1 敏感词模型（SensitiveWord）

```go
type SensitiveWord struct {
    ID        string          `json:"id"`
    Word      string          `json:"word"`      // 敏感词
    Category  string          `json:"category"`  // 分类
    Level     ModerationLevel `json:"level"`     // 级别
    IsActive  bool            `json:"is_active"` // 是否启用
    Reason    string          `json:"reason"`    // 添加原因
    CreatedBy string          `json:"created_by"`
    CreatedAt time.Time       `json:"created_at"`
    UpdatedAt time.Time       `json:"updated_at"`
}
```

### 4.2 安全事件模型（SecurityEvent）

```go
type SecurityEvent struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    EventType   string    `json:"event_type"`   // xss_attempt, sensitive_word等
    ContentType string    `json:"content_type"` // comment, letter等
    Content     string    `json:"content"`      // 触发内容
    RiskScore   int       `json:"risk_score"`   // 0-100
    Violations  int       `json:"violations"`   // 违规数量
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    Details     string    `json:"details"`      // JSON详情
    Handled     bool      `json:"handled"`
    HandledBy   *string   `json:"handled_by"`
    HandledAt   *time.Time `json:"handled_at"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 4.3 内容安全结果（ContentSecurityResult）

```go
type ContentSecurityResult struct {
    IsSafe           bool     `json:"is_safe"`
    RiskScore        int      `json:"risk_score"`        // 0-100
    CleanedContent   string   `json:"cleaned_content"`
    XSSDetected      bool     `json:"xss_detected"`
    SensitiveWords   []string `json:"sensitive_words"`
    Violations       []string `json:"violations"`
    RiskLevel        string   `json:"risk_level"`        // low/medium/high
    NeedManualReview bool     `json:"need_manual_review"`
}
```

---

## 五、API接口

### 5.1 管理员接口

#### 5.1.1 敏感词管理

```bash
# 获取敏感词列表
GET /api/admin/sensitive-words?page=1&limit=20&category=spam&is_active=true

# 添加敏感词
POST /api/admin/sensitive-words
{
    "word": "违规词",
    "category": "spam",
    "level": "high",
    "reason": "垃圾广告"
}

# 更新敏感词
PUT /api/admin/sensitive-words/:id
{
    "level": "block",
    "is_active": false
}

# 删除敏感词
DELETE /api/admin/sensitive-words/:id

# 获取统计信息
GET /api/admin/sensitive-words/stats

# 批量导入（CSV格式）
POST /api/admin/sensitive-words/batch-import
{
    "words": [
        {"word": "词1", "category": "spam", "level": "medium"},
        {"word": "词2", "category": "offensive", "level": "high"}
    ]
}

# 导出敏感词
GET /api/admin/sensitive-words/export

# 刷新内存词库
POST /api/admin/sensitive-words/refresh
```

#### 5.1.2 安全事件查询

```bash
# 获取安全事件列表
GET /api/admin/security-events?event_type=xss_attempt&limit=50

# 获取安全统计
GET /api/admin/security-stats
```

### 5.2 内部服务接口

```go
// 验证评论内容
func ValidateCommentContent(content string, userID string) (*ContentSecurityResult, error)

// 清理HTML内容
func CleanHTML(content string, strict bool) string

// 检测XSS攻击
func DetectXSS(content string) (bool, []string)

// 检查敏感词
func CheckSensitiveWords(content string) ([]string, string)
```

---

## 六、前端集成

### 6.1 敏感词管理界面

**组件路径**：`/frontend/src/components/admin/sensitive-words-management.tsx`

**主要功能**：
- 敏感词列表展示（分页、搜索、筛选）
- 添加/编辑/删除敏感词
- 批量导入导出
- 实时刷新词库
- 统计信息展示

### 6.2 权限控制

```tsx
// 权限检查
const hasPermission = user?.role === 'courier_level4' || 
                     user?.role === 'platform_admin' || 
                     user?.role === 'super_admin'

if (!hasPermission) {
    return <NoPermissionComponent />
}
```

---

## 七、安全最佳实践

### 7.1 防护策略

1. **纵深防御**：多层次的安全检查
   - 前端输入验证
   - 后端XSS检测
   - HTML内容清理
   - 敏感词过滤
   - 人工审核兜底

2. **最小权限原则**：
   - 仅必要角色可管理敏感词
   - 操作日志完整记录
   - 敏感操作需要二次确认

3. **性能优化**：
   - 敏感词使用内存缓存
   - 批量处理减少数据库访问
   - 异步处理大批量导入

### 7.2 监控与告警

- 高风险内容自动告警
- XSS攻击尝试实时通知
- 敏感词命中率统计
- 安全事件趋势分析

---

## 八、实施指南

### 8.1 部署要求

1. **依赖库**：
   - bluemonday：HTML清理库
   - Aho-Corasick：高效字符串匹配

2. **数据库**：
   - PostgreSQL 12+
   - 执行安全相关表迁移

3. **配置项**：
   ```yaml
   security:
     xss_detection: true
     sensitive_word_filter: true
     max_content_length: 10000
     risk_score_threshold: 70
   ```

### 8.2 测试建议

1. **XSS测试用例**：
   ```javascript
   // 应该被拦截的内容
   "<script>alert('xss')</script>"
   "<img src=x onerror=alert('xss')>"
   "<a href='javascript:alert(1)'>click</a>"
   ```

2. **敏感词测试**：
   - 单个敏感词检测
   - 变形词检测（如：敏感词 → 敏*感*词）
   - 组合词检测

3. **性能测试**：
   - 1000个敏感词的过滤性能
   - 并发内容验证压力测试

---

## 九、维护指南

### 9.1 日常维护

1. **敏感词库更新**：
   - 定期审查词库有效性
   - 根据实际情况调整级别
   - 清理过时或无效词汇

2. **安全事件处理**：
   - 每日查看高风险事件
   - 分析XSS攻击模式
   - 更新防护规则

3. **性能监控**：
   - 监控API响应时间
   - 检查内存使用情况
   - 优化慢查询

### 9.2 应急响应

**发现新型XSS攻击**：
1. 立即更新XSS检测规则
2. 扫描历史内容
3. 通知受影响用户
4. 更新安全补丁

**敏感词泄露事件**：
1. 快速添加到屏蔽词库
2. 全站内容扫描
3. 清理违规内容
4. 加强审核力度

---

## 十、总结

内容安全系统通过XSS防护和敏感词过滤双重机制，为OpenPenPal平台提供了全方位的内容安全保障。系统设计充分考虑了性能、可扩展性和易用性，特别是为四级信使和平台管理员提供了便捷的敏感词管理功能，确保平台内容的健康和安全。