# OpenPenPal 开发规范与标准

## 目录
- [一、开发理念](#一开发理念)
- [二、Git版本管理规范](#二git版本管理规范)
- [三、SOTA原则实施细则](#三sota原则实施细则)
- [四、Ultrathink模式应用](#四ultrathink模式应用)
- [五、代码质量标准](#五代码质量标准)
- [六、项目协作规范](#六项目协作规范)

---

## 一、开发理念

### 1.1 核心价值观
- **人文关怀优先**：技术服务于人的情感需求，慢社交胜过快功能
- **质量胜于速度**：宁可延期也要保证系统稳定性和用户体验
- **持续改进**：每次迭代都是对前版本的全面优化
- **共同成长**：团队学习优于个人英雄主义

### 1.2 开发原则
1. **最小惊讶原则**：API设计符合直觉，命名见名知意
2. **渐进增强**：新功能不破坏现有功能
3. **故障隔离**：单个服务故障不影响整体系统
4. **数据完整性**：永远不丢失用户的信件和数据

---

## 二、Git版本管理规范

### 2.1 分支策略（Git Flow Plus）

```
main（主分支）
├── develop（开发分支）
│   ├── feature/task-A1-barcode-validation（功能分支）
│   ├── feature/task-B2-envelope-template（功能分支）
│   └── feature/task-E1-ai-matching（功能分支）
├── release/v1.1.0（发布分支）
├── hotfix/fix-auth-vulnerability（热修复分支）
└── sota/permissions-system-v2（SOTA优化分支）
```

### 2.2 分支命名规范

| 分支类型 | 命名格式 | 示例 | 说明 |
|----------|----------|------|------|
| 功能分支 | `feature/task-<任务ID>-<功能简述>` | `feature/task-A1-barcode-validation` | 对应开发计划中的任务 |
| 修复分支 | `bugfix/<issue-id>-<问题简述>` | `bugfix/123-qr-scan-crash` | 修复已知问题 |
| 热修复 | `hotfix/<严重程度>-<问题简述>` | `hotfix/critical-jwt-leak` | 紧急生产问题 |
| SOTA优化 | `sota/<模块名>-<版本>` | `sota/permissions-v2` | 系统性架构优化 |
| 发布分支 | `release/v<版本号>` | `release/v1.1.0` | 版本发布准备 |

### 2.3 提交消息规范（Conventional Commits + Chinese）

#### 格式
```
<类型>[可选作用域]: <简要描述>

[可选详细描述]

[可选Footer]
```

#### 提交类型
- `feat`: 新功能
- `fix`: 修复问题
- `docs`: 文档更新
- `style`: 格式调整（不影响代码功能）
- `refactor`: 重构代码
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具相关
- `sota`: SOTA优化改进

#### 示例
```bash
# 好的提交消息
feat(barcode): 实现条码唯一性校验机制

- 添加条码重复性检查算法
- 实现首次扫码绑定锁定功能  
- 支持条码状态转换：未使用→已绑定→已投递→已送达
- 添加条码审计日志记录

关联任务: Task-A1
测试: 单元测试覆盖率95%

# 不好的提交消息
fix: 修复bug
update: 更新代码
```

### 2.4 合并策略

#### 主分支保护规则
```bash
# main 分支保护设置
- 禁止直接推送
- 必须通过 Pull Request
- 至少2人代码审查通过
- 必须通过所有CI检查
- 必须更新到最新版本

# develop 分支保护设置  
- 禁止直接推送
- 至少1人代码审查通过
- 必须通过基础CI检查
```

#### Pull Request模板
```markdown
## 变更类型
- [ ] 🚀 新功能 (feature)
- [ ] 🐛 问题修复 (bugfix)  
- [ ] 📚 文档更新 (docs)
- [ ] 🎨 代码优化 (refactor)
- [ ] ⚡ 性能提升 (perf)
- [ ] 🧪 测试完善 (test)
- [ ] 🏗️ SOTA优化 (sota)

## 变更描述
简要描述本次变更的内容和原因

## 关联任务
- 任务ID: Task-XX
- PRD章节: [链接]

## 测试情况
- [ ] 单元测试通过
- [ ] 集成测试通过  
- [ ] 手动测试完成
- [ ] 性能测试通过（如适用）

## 部署说明
描述部署时需要注意的事项

## Screenshots（如适用）
添加截图说明UI变更

## Checklist
- [ ] 代码遵循团队规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 通过了所有检查
```

### 2.5 版本标签管理

#### 版本号规范（语义化版本）
```
v<主版本>.<次版本>.<修补版本>[-<预发布标识>]

示例：
v1.0.0        # 正式版本
v1.1.0-beta.1 # 测试版本
v1.0.1        # 修补版本
```

#### 标签创建流程
```bash
# 1. 确保在正确分支
git checkout main
git pull origin main

# 2. 创建带注释的标签
git tag -a v1.1.0 -m "Release v1.1.0: AI慢社交系统上线

主要功能:
- AI笔友匹配算法
- 延迟回信机制  
- 写作灵感推送
- 博物馆策展增强

🤖 Generated with SOTA & Ultrathink principles"

# 3. 推送标签
git push origin v1.1.0
```

---

## 三、SOTA原则实施细则

### 3.1 SOTA定义与应用层次

**SOTA (State-of-the-Art)**: 在OpenPenPal项目中，SOTA不仅指技术前沿，更强调在人文社交场景下的最佳实践。

#### 应用层次
1. **架构层SOTA**：微服务、事件驱动、容器化
2. **代码层SOTA**：设计模式、清洁代码、类型安全
3. **算法层SOTA**：AI匹配、性能优化、安全防护
4. **体验层SOTA**：交互设计、加载速度、错误处理
5. **运维层SOTA**：监控告警、自动部署、故障恢复

### 3.2 SOTA Review机制

#### 每月SOTA评估
```markdown
## SOTA月度评估报告

### 评估维度
1. **技术债务清理**: 是否按计划减少技术债务？
2. **性能指标**: 关键指标是否达到行业领先水平？
3. **用户体验**: 是否符合慢社交的产品理念？
4. **代码质量**: 测试覆盖率、复杂度、可维护性
5. **安全标准**: 是否采用最新安全实践？

### 改进计划
基于评估结果，制定下月SOTA改进任务
```

#### SOTA分支策略
```bash
# 创建SOTA优化分支
git checkout -b sota/auth-system-v2 develop

# SOTA分支特点：
# 1. 允许较大幅度重构
# 2. 必须向后兼容
# 3. 需要详细的性能对比报告
# 4. 必须有回滚预案
```

### 3.3 技术选型SOTA标准

#### 后端技术栈
- **Go**: 性能优异，并发友好，适合微服务
- **GORM**: 最新版本，支持泛型，代码生成
- **JWT**: RS256算法，支持刷新令牌
- **Redis**: 集群模式，持久化配置
- **PostgreSQL**: 最新稳定版，分区表，物化视图

#### 前端技术栈
- **Next.js 14**: App Router，RSC，边缘渲染
- **TypeScript**: 严格模式，类型安全
- **Tailwind CSS**: 原子化CSS，响应式设计
- **React Hook Form**: 性能优化表单处理
- **SWR/TanStack Query**: 数据获取和缓存

#### AI/ML技术栈
- **PyTorch**: 深度学习框架
- **Transformers**: 预训练模型
- **FastAPI**: 异步API框架
- **Celery**: 分布式任务队列
- **MLflow**: 模型版本管理

### 3.4 SOTA代码示例

#### Go服务SOTA模板
```go
// sota/permissions/service.go
package permissions

import (
    "context"
    "time"
    
    "github.com/opentelemetry/opentelemetry-go/trace"
    "go.uber.org/zap"
)

// Service SOTA权限服务实现
type Service struct {
    repo       Repository
    cache      Cache
    logger     *zap.Logger
    tracer     trace.Tracer
    metrics    Metrics
}

// CheckPermission SOTA权限检查 - 支持链路追踪和性能监控
func (s *Service) CheckPermission(ctx context.Context, userID string, permission string) (bool, error) {
    ctx, span := s.tracer.Start(ctx, "permissions.CheckPermission")
    defer span.End()
    
    timer := s.metrics.PermissionCheckDuration.Timer()
    defer timer.ObserveDuration()
    
    // 1. 缓存层检查
    if cached, hit := s.cache.Get(ctx, s.cacheKey(userID, permission)); hit {
        s.metrics.CacheHitRate.Inc()
        return cached.(bool), nil
    }
    
    // 2. 数据库查询
    hasPermission, err := s.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        s.logger.Error("permission check failed", 
            zap.String("userID", userID),
            zap.String("permission", permission),
            zap.Error(err))
        return false, err
    }
    
    // 3. 缓存结果（5分钟TTL）
    s.cache.Set(ctx, s.cacheKey(userID, permission), hasPermission, 5*time.Minute)
    
    return hasPermission, nil
}
```

---

## 四、Ultrathink模式应用

### 4.1 Ultrathink定义
**Ultrathink**: 深度思考模式，要求在每个决策点进行全方位、多维度的思考，考虑长期影响和用户真实需求。

### 4.2 Ultrathink应用场景

#### 技术决策前的Ultrathink
```markdown
## Ultrathink决策模板

### 问题描述
清晰描述需要解决的问题

### 多方案对比
| 方案 | 优势 | 劣势 | 成本 | 风险 | 维护性 |
|-----|------|------|------|------|--------|
| A   |      |      |      |      |        |
| B   |      |      |      |      |        |
| C   |      |      |      |      |        |

### 长期影响分析
- 对系统架构的影响
- 对团队技能要求的影响  
- 对用户体验的影响
- 对运维复杂度的影响

### 人文关怀考量
- 是否符合慢社交理念？
- 是否增加了用户的认知负担？
- 是否保护了用户隐私？
- 是否促进了真实的人际连接？

### 决策结论
基于以上分析，选择方案X，理由如下...
```

#### 代码Review中的Ultrathink
```markdown
## Ultrathink Code Review清单

### 功能层面
- [ ] 是否真正解决了用户问题？
- [ ] 是否符合PRD中的慢社交理念？
- [ ] 边界情况是否充分考虑？

### 架构层面  
- [ ] 是否符合系统整体架构？
- [ ] 是否引入了不必要的复杂性？
- [ ] 是否便于后续扩展？

### 性能层面
- [ ] 是否存在性能瓶颈？
- [ ] 资源使用是否合理？
- [ ] 并发安全是否保证？

### 安全层面
- [ ] 是否存在安全漏洞？
- [ ] 用户数据是否被妥善保护？
- [ ] 权限控制是否正确？

### 用户体验层面
- [ ] 错误处理是否友好？
- [ ] 响应时间是否可接受？
- [ ] 交互是否符合直觉？
```

### 4.3 Ultrathink会议机制

#### 每周Ultrathink深度讨论
- **时间**: 每周五下午2-4点
- **参与者**: 全体技术团队
- **流程**:
  1. 技术难题深度分析（30分钟）
  2. 用户反馈深度挖掘（30分钟）
  3. 架构演进方向讨论（30分钟）
  4. 下周重点任务规划（30分钟）

#### Ultrathink产出物
- 每周深度思考报告
- 技术决策记录（ADR）
- 用户体验改进建议
- 长期规划调整建议

---

## 五、代码质量标准

### 5.1 通用编码规范

#### 命名约定
```go
// ✅ 好的命名
type LetterService struct {
    repository LetterRepository
    validator  LetterValidator
    notifier   NotificationService
}

func (s *LetterService) CreateAnonymousLetter(ctx context.Context, req CreateLetterRequest) (*Letter, error) {
    // 实现
}

// ❌ 不好的命名
type LS struct {
    repo Repository
    val  Validator
}

func (s *LS) Create(req interface{}) (interface{}, error) {
    // 实现
}
```

#### 注释规范
```go
// CheckBarcodeUniqueness 检查条码唯一性
//
// 该函数实现PRD中条码系统的核心需求：
// 1. 条码全局唯一性验证
// 2. 防重复使用机制
// 3. 审计日志记录
//
// 参数:
//   - ctx: 请求上下文，用于链路追踪和超时控制
//   - barcode: 待检查的条码字符串，格式为8位数字+字母组合
//
// 返回值:
//   - bool: true表示条码可用，false表示已被使用
//   - error: 检查过程中的错误，nil表示检查成功
//
// 示例:
//   available, err := CheckBarcodeUniqueness(ctx, "AB123456")
//   if err != nil {
//       return fmt.Errorf("条码检查失败: %w", err)
//   }
func CheckBarcodeUniqueness(ctx context.Context, barcode string) (bool, error) {
    // 实现
}
```

### 5.2 测试标准

#### 测试覆盖率要求
- **单元测试**: ≥ 80%
- **集成测试**: ≥ 60%  
- **核心业务逻辑**: ≥ 95%

#### 测试命名规范
```go
func TestLetterService_CreateAnonymousLetter_Success(t *testing.T) {
    // 测试正常情况下创建匿名信件
}

func TestLetterService_CreateAnonymousLetter_DuplicateBarcode_ReturnsError(t *testing.T) {
    // 测试条码重复时返回错误
}

func TestBarcodeValidator_Validate_InvalidFormat_ReturnsError(t *testing.T) {
    // 测试无效条码格式返回错误
}
```

### 5.3 性能要求

#### 响应时间标准
| 操作类型 | 目标响应时间 | 最大可接受时间 |
|----------|--------------|----------------|
| 用户登录 | < 500ms | < 1s |
| 写信页面加载 | < 2s | < 3s |
| 条码扫描 | < 500ms | < 1s |
| AI匹配 | < 5s | < 10s |
| 信件发送 | < 1s | < 2s |

#### 性能测试要求
```go
func BenchmarkLetterService_CreateLetter(b *testing.B) {
    service := setupLetterService()
    req := CreateLetterRequest{
        Content: "测试信件内容",
        IsAnonymous: true,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CreateLetter(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## 六、项目协作规范

### 6.1 沟通协作

#### 异步沟通优先
- 技术讨论优先使用书面形式（Issue、Wiki、文档）
- 代码Review必须书面反馈
- 重要决策必须有书面记录

#### 会议效率
- 会议必须有明确议程和时间限制
- 每个会议必须有行动项和负责人
- 会议纪要24小时内发布

### 6.2 知识管理

#### 文档维护
- 每个模块必须有README.md
- API变更必须更新文档
- 重要决策记录在ADR中
- 故障分析报告归档保存

#### 知识分享
- 每月技术分享会
- 代码Review中的学习分享
- 新技术调研报告共享

### 6.3 质量保证

#### 代码审查标准
- 每个PR至少2人审查
- 核心模块至少1名高级工程师审查
- 安全相关代码必须安全专家审查

#### 发布检查清单
```markdown
## 发布前检查清单

### 代码质量
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 静态代码分析通过
- [ ] 性能测试通过

### 文档更新
- [ ] API文档已更新
- [ ] 用户文档已更新
- [ ] 部署文档已更新

### 安全检查
- [ ] 依赖漏洞扫描通过
- [ ] 代码安全扫描通过
- [ ] 权限控制验证通过

### 部署准备
- [ ] 数据库迁移脚本准备
- [ ] 配置文件更新
- [ ] 回滚方案确认
- [ ] 监控告警配置
```

---

## 七、持续改进

### 7.1 定期评估
- **每周**: 代码质量报告
- **每月**: SOTA评估和技术债务清理
- **每季**: 架构演进规划
- **每年**: 技术栈全面评估

### 7.2 学习成长
- 鼓励团队成员参加技术会议
- 内部技术分享激励机制
- 定期Code Review经验总结
- 建立技术成长路径

### 7.3 工具支持
- 自动化代码检查工具
- 性能监控和告警系统
- 自动化测试和部署流水线
- 知识库和文档管理系统

---

**这份开发规范是活文档，随着项目发展和团队成长不断完善。每个开发者都有责任维护和改进这些标准，确保OpenPenPal项目始终保持技术领先和人文关怀的完美结合。**