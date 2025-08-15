# OpenPenPal 内容审核系统实现总结

## 📋 实现概览

基于FSD规范，已完成内容审核系统的Mock实现，为后续接入AI模型API做好准备。系统采用模块化设计，支持多种内容类型的审核和管理。

---

## 🏗️ 系统架构

### 核心组件
1. **数据模型层** (`models/moderation.go`)
2. **服务层** (`services/moderation_service.go`)
3. **集成层** (`services/moderation_integration.go`)
4. **HTTP处理层** (`handlers/moderation_handler.go`)
5. **数据库迁移** (`config/moderation_migrations.sql`)

### 审核流程
```
内容提交 → 自动预审 → 风险评估 → 状态判定 → 人工复审(可选) → 最终结果
```

---

## 📊 数据库设计

### 主要数据表

| 表名 | 说明 | 关键字段 |
|------|------|----------|
| `moderation_records` | 审核记录主表 | `content_type`, `status`, `risk_level` |
| `moderation_rules` | 审核规则配置 | `rule_type`, `action`, `priority` |
| `moderation_stats` | 审核统计数据 | `date`, `content_type`, 统计字段 |
| `moderation_logs` | 审核操作日志 | `moderation_id`, `action`, `performed_by` |

### 支持的内容类型
- **letter**: 用户信件（匿名信/公开信）
- **ai_reply**: AI生成回信
- **image**: 图片内容（手写信照片/信封设计）
- **envelope**: 信封设计稿
- **user_profile**: 用户资料

---

## 🔧 API接口设计

### 用户端接口
```
POST /api/v1/moderation/submit                    # 提交内容审核
GET  /api/v1/moderation/status/:id                # 查询审核状态
POST /api/v1/moderation/batch-status              # 批量查询状态
```

### 管理员接口
```
GET  /api/v1/admin/moderation/pending             # 获取待审核队列
POST /api/v1/admin/moderation/:id/approve         # 审核通过
POST /api/v1/admin/moderation/:id/reject          # 审核拒绝
GET  /api/v1/admin/moderation/stats               # 审核统计
GET  /api/v1/admin/moderation/rules               # 获取审核规则
POST /api/v1/admin/moderation/rules               # 创建审核规则
PUT  /api/v1/admin/moderation/rules/:id           # 更新审核规则
DELETE /api/v1/admin/moderation/rules/:id         # 删除审核规则
```

---

## 🤖 Mock实现特性

### 自动审核逻辑
- **85%自动通过率**：模拟真实审核场景
- **关键词检测**：高/中风险关键词识别
- **内容长度检查**：异常长度内容标记
- **风险等级评估**：low/medium/high三级分类

### 审核规则示例
```json
{
  "high_risk_keywords": ["暴力", "色情", "政治", "恐怖", "犯罪"],
  "medium_risk_keywords": ["愤怒", "不满", "抱怨", "争吵"],
  "auto_approval_score": 0.3,
  "auto_rejection_score": 0.8
}
```

### 状态流转
```
pending → approved ✅
       → rejected ❌  
       → reviewing → approved/rejected
```

---

## 🔗 业务集成点

### 信件创建流程
1. 用户创建信件草稿
2. 自动提交内容审核
3. 审核通过后允许发送
4. 公开信件优先级更高

### AI回信生成
1. AI生成回信内容
2. 自动提交审核（标记AI生成）
3. 通过后发送给用户
4. 拒绝后重新生成

### 图片上传
1. 用户上传手写信照片/信封设计
2. 提取文字内容（OCR）
3. 图片+文字双重审核
4. 通过后允许展示

---

## 📈 运营功能

### 审核队列管理
- 按内容类型分类展示
- 按优先级排序处理
- 快速审核操作界面

### 统计分析
- 日/周/月审核量统计
- 通过率趋势分析
- 不同内容类型表现
- 审核员工作量统计

### 规则配置
- 动态添加/修改审核规则
- 关键词库管理
- 审核策略版本控制
- A/B测试支持

---

## 🚀 后续AI接入方案

### 接入点设计
当前Mock实现为AI接入预留了接口：

```go
// 在 analyzeContent 方法中替换为真实AI调用
func (m *MockModerationService) analyzeContent(content string, contentType string) (riskLevel string, autoScore float64, tags []string) {
    // TODO: 替换为真实AI API调用
    // 1. 文本分析API：腾讯云/阿里云内容安全
    // 2. 图像识别API：图片内容检测
    // 3. 自定义模型：微调后的内容审核模型
}
```

### 推荐AI服务
1. **文本审核**：阿里云内容安全、腾讯云天御
2. **图像审核**：百度AI、华为云等
3. **自定义模型**：基于用户微调的专用模型

---

## ✅ 实现完成度

- [✅] 数据库表结构设计
- [✅] Mock审核服务实现
- [✅] HTTP接口完整实现
- [✅] 业务流程集成接口
- [✅] 管理员操作界面API
- [✅] 统计分析功能
- [✅] 审核规则管理
- [✅] 操作日志记录

### 待接入AI功能
- [ ] 真实文本内容分析API
- [ ] 图像内容识别API
- [ ] 用户自定义模型集成
- [ ] 审核策略智能优化

---

## 🛡️ 内容安全系统（XSS防护和敏感词过滤）

### 系统架构

基于业界最佳实践，实现了多层次的内容安全防护体系：

1. **XSS防护层**：使用bluemonday库进行HTML清理
2. **敏感词过滤层**：支持多级别的敏感词管理
3. **内容验证层**：综合风险评分和安全验证

### 核心组件

#### ContentSecurityService (`services/content_security_service.go`)

主要功能：
- **HTML清理**：标准模式和严格模式的HTML内容清理
- **XSS检测**：检测JavaScript注入、危险HTML标签、危险URL协议等
- **敏感词管理**：支持增删改查和实时刷新
- **内容验证**：综合安全评分（0-100）

```go
// 核心方法
ValidateCommentContent(content string, userID string) (*ContentSecurityResult, error)
AddSensitiveWord(word SensitiveWord) error
RefreshSensitiveWords() error
```

#### 安全模型 (`models/moderation.go`)

新增安全相关模型：
- **SecurityEvent**：安全事件记录（XSS攻击、违规内容等）
- **SecurityStats**：安全统计信息

### XSS防护实现

#### 防护策略
1. **HTML清理器**：
   - 标准模式：允许安全的HTML标签（b, i, em, strong等）
   - 严格模式：移除所有HTML标签

2. **XSS模式检测**：
   - JavaScript注入：`<script>`, `javascript:`, `onerror=`等
   - 危险HTML标签：`<iframe>`, `<object>`, `<embed>`等
   - 危险URL协议：`javascript:`, `vbscript:`, `data:`等

3. **风险评分系统**：
   - 0-30分：低风险
   - 31-70分：中风险
   - 71-100分：高风险（自动拒绝）

### 敏感词管理系统

#### 功能特性
1. **权限控制**：仅四级信使、平台管理员和超级管理员可访问
2. **分级管理**：低风险、中风险、高风险、屏蔽四个级别
3. **分类支持**：垃圾信息、不当内容、冒犯性内容等7个分类
4. **批量操作**：支持批量导入导出（CSV格式）
5. **实时刷新**：支持热更新敏感词库到内存

#### API端点

管理员端点（需要权限）：
```
GET    /api/admin/sensitive-words              # 获取敏感词列表
POST   /api/admin/sensitive-words              # 添加敏感词
PUT    /api/admin/sensitive-words/:id          # 更新敏感词
DELETE /api/admin/sensitive-words/:id          # 删除敏感词
GET    /api/admin/sensitive-words/stats        # 获取统计信息
POST   /api/admin/sensitive-words/batch-import # 批量导入
GET    /api/admin/sensitive-words/export       # 导出敏感词
POST   /api/admin/sensitive-words/refresh      # 刷新词库
```

### 前端管理界面

#### SensitiveWordsManagement组件 (`components/admin/sensitive-words-management.tsx`)

功能特点：
- **权限验证**：自动检查用户角色权限
- **实时搜索**：支持关键词、分类、级别筛选
- **批量操作**：一键导入导出
- **统计面板**：显示总词数、活跃词数、分类统计等
- **操作日志**：记录所有敏感词操作

### 集成点

#### 评论系统集成
```go
// 在创建评论前进行安全检查
securityResult, err := s.securitySvc.ValidateCommentContent(req.Content, userID)
if securityResult.XSSDetected || !securityResult.IsSafe {
    return nil, fmt.Errorf("comment content contains security violations")
}
```

#### 数据库迁移
新增SecurityEvent表用于记录安全事件：
- XSS攻击尝试
- 敏感词触发
- 高风险内容拦截

### 安全最佳实践

1. **多层防护**：XSS检测 → HTML清理 → 敏感词过滤 → 人工审核
2. **实时更新**：敏感词库支持热更新，无需重启服务
3. **日志记录**：所有安全事件都会记录，便于分析和改进
4. **性能优化**：使用内存缓存和高效的字符串匹配算法

### 使用示例

#### 添加敏感词
```bash
curl -X POST /api/admin/sensitive-words \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "word": "违规词汇",
    "category": "spam",
    "level": "high"
  }'
```

#### 刷新词库
```bash
curl -X POST /api/admin/sensitive-words/refresh \
  -H "Authorization: Bearer $TOKEN"
```

### 实现完成度

- [✅] XSS防护引擎（bluemonday）
- [✅] 敏感词过滤系统
- [✅] 内容安全验证API
- [✅] 权限控制（四级信使+管理员）
- [✅] 前端管理界面
- [✅] 批量导入导出
- [✅] 实时刷新机制
- [✅] 安全事件记录

### 待完善功能
- [ ] 输入速率限制
- [ ] CSP（内容安全策略）头配置
- [ ] AI驱动的智能识别
- [ ] 多语言敏感词支持

### 待接入AI功能
- [ ] 真实文本内容分析API
- [ ] 图像内容识别API
- [ ] 用户自定义模型集成
- [ ] 审核策略智能优化

---

## 🔧 使用指南

### 启动服务
系统已集成到主应用，启动后自动加载Mock审核服务。

### 测试审核流程
```bash
# 提交内容审核
curl -X POST /api/v1/moderation/submit \
  -H "Content-Type: application/json" \
  -d '{
    "content_type": "letter",
    "source_id": "letter_123",
    "content": "这是一封测试信件",
    "priority": "medium"
  }'

# 查询审核状态
curl -X GET /api/v1/moderation/status/{moderation_id}
```

### 管理员操作
```bash
# 获取待审核队列
curl -X GET /api/v1/admin/moderation/pending?content_type=letter&limit=20

# 审核通过
curl -X POST /api/v1/admin/moderation/{id}/approve \
  -H "Content-Type: application/json" \
  -d '{"notes": "内容符合规范"}'
```

---

## 📋 总结

内容审核系统已按照FSD要求完成Mock实现，具备完整的审核流程、管理界面和统计功能。系统设计充分考虑了后续AI接入的需求，提供了标准化的接口和灵活的配置机制。

当用户准备接入微调的AI模型时，只需替换`analyzeContent`方法中的Mock逻辑为真实API调用，即可实现生产级的智能内容审核功能。