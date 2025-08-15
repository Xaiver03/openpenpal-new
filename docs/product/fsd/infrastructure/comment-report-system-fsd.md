# OpenPenPal 评论举报系统 FSD（Comment Report System）

**功能规格说明书**

## 1. 系统概述

### 1.1 系统目标
评论举报系统是OpenPenPal平台的内容安全保障核心组件，旨在为用户提供便捷的举报机制，并为管理员提供高效的举报处理工具，确保平台内容的健康和安全。

### 1.2 业务背景
- **社区内容管理**: 维护良好的社区氛围，及时处理不当内容
- **用户体验保障**: 保护用户免受垃圾信息、骚扰等不良行为影响
- **法规合规**: 符合相关法律法规对网络内容管理的要求
- **平台声誉**: 维护OpenPenPal作为高质量校园交流平台的品牌形象

### 1.3 系统范围
- **用户举报功能**: 支持对评论内容的多类型举报
- **管理员审核**: 完整的举报处理工作流程
- **自动化辅助**: 智能内容检测和风险评估
- **统计分析**: 举报数据统计和趋势分析
- **通知反馈**: 举报处理结果通知机制

## 2. 功能需求规格

### 2.1 核心功能模块

#### 2.1.1 用户举报功能
**FR-UR-001: 举报入口**
- 用户可在任何评论的操作菜单中找到"举报"选项
- 举报按钮在用户hover时显示，不会对自己的评论显示举报选项
- 支持快速举报和详细举报两种模式

**FR-UR-002: 举报分类**
系统支持以下举报类型：
```
1. 垃圾信息 (spam)
   - 描述：垃圾邮件、广告或重复内容
   - 处理优先级：中等
   
2. 不当内容 (inappropriate)
   - 描述：不符合社区准则的内容
   - 处理优先级：高
   
3. 冒犯性内容 (offensive)
   - 描述：仇恨言论、骚扰或恶意攻击
   - 处理优先级：高
   
4. 虚假信息 (false_info)
   - 描述：误导性或不准确的信息
   - 处理优先级：中等
   
5. 其他 (other)
   - 描述：其他违规行为
   - 处理优先级：低
```

**FR-UR-003: 举报表单**
- 必填字段：举报原因（单选）
- 可选字段：详细描述（最多500字符）
- 表单验证：实时验证用户输入
- 用户友好提示：说明举报流程和注意事项

**FR-UR-004: 重复举报防护**
- 同一用户对同一评论只能举报一次
- 系统自动检测并阻止重复举报
- 对重复举报尝试给出友好提示

#### 2.1.2 管理员审核系统
**FR-MA-001: 举报列表管理**
- 支持按状态筛选：待处理、已解决、已驳回
- 支持按举报类型筛选
- 支持按时间范围筛选
- 分页显示，每页20条记录

**FR-MA-002: 举报详情查看**
```
举报详情包含：
- 举报基本信息：ID、时间、状态
- 举报原因和详细描述
- 被举报评论的完整内容
- 评论作者信息（脱敏处理）
- 举报者信息（脱敏处理）
- 相关上下文（评论所在的讨论串）
```

**FR-MA-003: 举报处理操作**
管理员可执行以下操作：
```
1. 解决举报 (resolved)
   - 隐藏被举报评论
   - 删除被举报评论
   - 对评论作者发出警告
   - 记录处理原因

2. 驳回举报 (dismissed)
   - 保留评论内容
   - 标记举报为无效
   - 记录驳回原因
   
3. 批量处理
   - 支持同时处理多个相似举报
   - 批量操作确认机制
```

**FR-MA-004: 处理历史记录**
- 记录所有处理操作的详细信息
- 包含处理人员、处理时间、处理方式、处理原因
- 支持处理历史的查询和导出

#### 2.1.3 自动化检测辅助
**FR-AD-001: 智能风险评估**
系统自动评估举报内容的风险等级：
```
风险评估算法：
1. 关键词匹配
   - 维护敏感词库
   - 支持正则表达式匹配
   - 动态更新词库

2. 行为模式分析
   - 用户历史举报记录
   - 评论发布频率异常
   - 账户注册时间和活跃度

3. 内容特征分析
   - 重复字符检测
   - 链接和联系方式检测
   - 语言风格异常分析
```

**FR-AD-002: 自动预处理**
- 高风险内容自动标记为"待审核"状态
- 明显违规内容自动隐藏，等待人工确认
- 疑似垃圾内容批量检测和标记

#### 2.1.4 统计分析功能
**FR-SA-001: 实时统计面板**
```
管理员控制台显示：
- 总举报数量
- 今日新增举报数
- 待处理举报数量
- 各类型举报分布
- 处理效率统计
```

**FR-SA-002: 趋势分析报告**
- 按时间维度的举报趋势图
- 按类型维度的举报分布图
- 重点用户举报行为分析
- 社区内容健康度指标

### 2.2 数据模型规格

#### 2.2.1 核心数据模型
```sql
-- 评论举报表
CREATE TABLE comment_reports (
    id VARCHAR(36) PRIMARY KEY,
    comment_id VARCHAR(36) NOT NULL,
    reporter_id VARCHAR(36) NOT NULL,
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    handled_at TIMESTAMP NULL,
    handled_by VARCHAR(36) NULL,
    handler_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- 索引
    INDEX idx_comment_id (comment_id),
    INDEX idx_reporter_id (reporter_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    
    -- 外键约束
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (handled_by) REFERENCES users(id) ON DELETE SET NULL,
    
    -- 唯一约束（防止重复举报）
    UNIQUE KEY uk_comment_reporter (comment_id, reporter_id)
);
```

#### 2.2.2 状态流转模型
```
举报状态流转：
pending -> resolved (解决)
pending -> dismissed (驳回)

状态说明：
- pending: 待处理（初始状态）
- resolved: 已解决（管理员确认举报有效并处理）
- dismissed: 已驳回（管理员确认举报无效）
```

### 2.3 API接口规格

#### 2.3.1 用户举报接口
```http
POST /api/v2/comments/{commentId}/report
Content-Type: application/json
Authorization: Bearer {token}

Request Body:
{
  "reason": "spam|inappropriate|offensive|false_info|other",
  "description": "详细描述（可选，最多500字符）"
}

Response (200 OK):
{
  "success": true,
  "message": "举报已提交成功",
  "report_id": "report-uuid"
}

Response (400 Bad Request):
{
  "success": false,
  "error": "ALREADY_REPORTED",
  "message": "您已经举报过此评论"
}
```

#### 2.3.2 管理员审核接口
```http
GET /api/v2/admin/reports
Authorization: Bearer {admin-token}
Query Parameters:
- status: pending|resolved|dismissed|all (默认: pending)
- reason: 举报类型筛选
- page: 页码 (默认: 1)
- limit: 每页数量 (默认: 20)
- sort: created_at|status (默认: created_at)
- order: asc|desc (默认: desc)

Response (200 OK):
{
  "success": true,
  "data": {
    "reports": [
      {
        "id": "report-uuid",
        "comment_id": "comment-uuid",
        "reason": "spam",
        "description": "这是垃圾广告",
        "status": "pending",
        "created_at": "2025-08-15T10:00:00Z",
        "comment": {
          "id": "comment-uuid",
          "content": "评论内容...",
          "user": {
            "id": "user-uuid",
            "username": "user123",
            "nickname": "用户昵称"
          }
        },
        "reporter": {
          "id": "reporter-uuid",
          "username": "reporter1",
          "nickname": "举报者昵称"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 156,
      "pages": 8
    },
    "stats": {
      "total": 156,
      "pending": 23,
      "resolved": 108,
      "dismissed": 25
    }
  }
}
```

```http
PUT /api/v2/admin/reports/{reportId}
Authorization: Bearer {admin-token}
Content-Type: application/json

Request Body:
{
  "status": "resolved|dismissed",
  "handler_note": "处理说明（可选）",
  "comment_action": "hide|delete|none" // 对评论的处理动作
}

Response (200 OK):
{
  "success": true,
  "message": "举报处理完成",
  "report": {
    "id": "report-uuid",
    "status": "resolved",
    "handled_at": "2025-08-15T11:00:00Z",
    "handled_by": "admin-uuid",
    "handler_note": "内容确实违规，已隐藏"
  }
}
```

## 3. 系统集成规格

### 3.1 与评论系统集成
- **数据关联**: 举报记录通过comment_id与评论系统关联
- **状态同步**: 评论被删除时，相关举报自动标记为"已解决"
- **权限控制**: 继承评论系统的权限模型

### 3.2 与用户系统集成
- **身份验证**: 使用统一的JWT身份认证
- **权限验证**: 管理员权限验证通过用户角色系统
- **用户行为记录**: 举报行为记录到用户活动日志

### 3.3 与通知系统集成
- **举报确认通知**: 用户提交举报后收到确认通知
- **处理结果通知**: 举报处理完成后通知相关用户
- **管理员提醒**: 新举报自动通知在线管理员

### 3.4 与审核系统集成
- **内容审核**: 被举报内容自动进入审核队列
- **规则引擎**: 与平台审核规则引擎联动
- **黑名单管理**: 严重违规用户自动加入黑名单

## 4. 性能与可用性需求

### 4.1 性能要求
- **响应时间**: 举报提交响应时间 < 500ms
- **查询性能**: 管理员列表查询响应时间 < 1s
- **并发处理**: 支持1000+并发举报提交
- **数据库优化**: 重要查询字段建立索引

### 4.2 可用性要求
- **系统可用性**: 99.9%
- **故障恢复**: 自动故障检测和恢复
- **数据备份**: 每日自动备份举报数据
- **容灾机制**: 支持异地容灾部署

### 4.3 扩展性要求
- **水平扩展**: 支持微服务架构下的水平扩展
- **存储扩展**: 支持分表分库策略
- **功能扩展**: 预留接口支持新的举报类型
- **集成扩展**: 支持第三方内容审核服务集成

## 5. 安全性要求

### 5.1 数据安全
- **敏感信息保护**: 举报者身份信息脱敏存储
- **访问控制**: 严格的角色权限控制
- **数据加密**: 敏感字段数据库加密存储
- **审计日志**: 完整的操作审计日志

### 5.2 业务安全
- **防刷机制**: 防止恶意批量举报
- **验证码保护**: 敏感操作增加验证码验证
- **IP限制**: 对异常IP进行访问限制
- **内容过滤**: 举报描述内容安全过滤

### 5.3 系统安全
- **SQL注入防护**: 参数化查询防止SQL注入
- **XSS防护**: 输出内容XSS过滤
- **CSRF防护**: CSRF Token验证
- **接口限流**: API接口访问频率限制

## 6. 监控与告警

### 6.1 业务监控
- **举报数量监控**: 实时监控举报提交量
- **处理效率监控**: 监控举报处理时效
- **异常行为监控**: 监控异常举报模式
- **用户行为监控**: 监控用户举报行为异常

### 6.2 技术监控
- **系统性能监控**: CPU、内存、数据库性能
- **接口响应监控**: API响应时间和成功率
- **错误日志监控**: 系统错误和异常监控
- **依赖服务监控**: 外部依赖服务可用性

### 6.3 告警策略
```
告警等级：
1. 紧急告警（立即处理）
   - 系统服务不可用
   - 大量举报处理失败
   - 数据安全异常

2. 重要告警（1小时内处理）
   - 举报处理积压超过阈值
   - API响应时间超过2秒
   - 异常举报行为检测

3. 一般告警（24小时内处理）
   - 举报数量异常增长
   - 处理效率下降
   - 非关键功能异常
```

## 7. 实施计划

### 7.1 开发阶段划分
```
Phase 1: 核心功能开发 (已完成)
- ✅ 基础数据模型设计
- ✅ 用户举报API实现
- ✅ 管理员审核API实现
- ✅ 前端举报组件开发
- ✅ 管理员审核界面开发

Phase 2: 增强功能开发 (进行中)
- 🔄 自动化检测算法优化
- 🔄 统计分析功能完善
- 🔄 通知系统集成
- 🔄 批量处理功能

Phase 3: 优化与监控 (计划中)
- 📋 性能优化
- 📋 监控告警系统
- 📋 安全加固
- 📋 用户体验优化
```

### 7.2 测试策略
- **单元测试**: 核心业务逻辑单元测试覆盖率 > 90%
- **集成测试**: API接口集成测试
- **压力测试**: 高并发场景压力测试
- **安全测试**: 安全漏洞扫描和渗透测试
- **用户体验测试**: 真实用户场景测试

### 7.3 部署策略
- **灰度发布**: 逐步推广，降低风险
- **数据迁移**: 平滑的数据迁移策略
- **回滚机制**: 快速回滚机制
- **监控部署**: 部署过程全程监控

## 8. 运维与维护

### 8.1 日常运维
- **数据备份**: 每日自动备份，每周全量备份
- **性能监控**: 实时性能指标监控
- **日志管理**: 日志轮转和归档策略
- **容量规划**: 定期评估系统容量需求

### 8.2 应急响应
- **故障处理流程**: 标准化故障处理流程
- **应急预案**: 各类故障场景应急预案
- **联系机制**: 7x24小时应急联系机制
- **恢复验证**: 故障恢复后的验证流程

### 8.3 持续改进
- **用户反馈收集**: 定期收集用户反馈
- **系统优化**: 基于监控数据的持续优化
- **功能迭代**: 基于业务需求的功能迭代
- **技术升级**: 定期技术栈升级和安全更新

---

## 附录

### A. 术语定义
- **举报**: 用户对违规内容的投诉行为
- **审核**: 管理员对举报内容的处理过程
- **脱敏**: 对敏感信息进行匿名化处理
- **风险评估**: 基于算法的内容风险等级判定

### B. 参考文档
- [OpenPenPal 评论系统 FSD](comment-system-fsd.md)
- [OpenPenPal 用户系统 FSD](user-system-fsd.md)
- [OpenPenPal 通知系统 FSD](notification-system-fsd.md)
- [OpenPenPal 内容审核系统 FSD](moderation-system-fsd.md)

### C. 变更历史
| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 1.0 | 2025-08-15 | 初始版本，完整的举报系统功能规格 | AI Assistant |

---

**文档状态**: ✅ 已完成  
**最后更新**: 2025-08-15  
**下次评审**: 2025-09-15