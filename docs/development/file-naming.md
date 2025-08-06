# OpenPenPal 项目文件命名规范标准

## 📋 命名规范总则

### 1. 文件命名原则
- **清晰性**: 文件名应明确表示文件内容和用途
- **一致性**: 全项目采用统一的命名风格
- **可预测性**: 命名模式应易于理解和记忆
- **版本控制友好**: 避免特殊字符和空格

### 2. 命名风格指南

#### 2.1 代码文件
- **Go文件**: 小写字母+下划线，如 `user_service.go`
- **JavaScript/TypeScript**: 小写字母+连字符，如 `user-service.ts`
- **Python**: 小写字母+下划线，如 `user_service.py`
- **配置文件**: 使用点分隔，如 `config.prod.yml`

#### 2.2 文档文件
- **README文件**: 全大写 `README.md`
- **技术文档**: 小写连字符，如 `api-specification.md`
- **会议记录**: 日期前缀+主题，如 `2025-07-22-meeting-notes.md`

#### 2.3 测试文件
- **单元测试**: `[模块名].test.js` 或 `[模块名]_test.go`
- **集成测试**: `[功能]-integration.test.js`
- **E2E测试**: `[场景]-e2e.test.js`
- **测试报告**: `report-[类型]-[日期].md`

## 🎯 具体命名标准

### 目录结构命名
```
openpenpal/
├── frontend/                    # Next.js前端
├── backend/                     # Go后端API
├── services/                    # 微服务
│   ├── admin-service/          # 管理服务
│   ├── courier-service/        # 信使服务
│   ├── gateway/               # API网关
│   ├── ocr-service/           # OCR服务
│   └── write-service/         # 写信服务
├── test-kimi/                  # 测试套件
├── docs/                       # 项目文档
├── scripts/                    # 部署脚本
├── config/                     # 配置文件
├── logs/                       # 日志文件
└── archive/                    # 归档文件
```

### 文件扩展名规范

#### 代码文件扩展名
- **后端Go**: `.go`
- **前端React/Next.js**: `.tsx` (组件), `.ts` (工具), `.js` (脚本)
- **Python服务**: `.py`
- **配置文件**: `.yml` (YAML), `.json` (JSON), `.env` (环境变量)

#### 文档扩展名
- **Markdown文档**: `.md`
- **JSON配置**: `.json`
- **YAML配置**: `.yml` 或 `.yaml`
- **Shell脚本**: `.sh`
- **环境变量**: `.env` + 后缀

### 具体文件命名规范表

#### 前端文件命名
| 文件类型 | 命名示例 | 说明 |
|----------|----------|------|
| 页面组件 | `user-profile.tsx` | 功能+类型 |
| UI组件 | `button-component.tsx` | 组件+类型 |
| 工具函数 | `date-utils.ts` | 功能+工具 |
| 样式文件 | `user-profile.css` | 对应组件名 |
| API服务 | `user-api.ts` | 功能+api |

#### 后端文件命名
| 文件类型 | 命名示例 | 说明 |
|----------|----------|------|
| 处理器 | `user_handler.go` | 功能+handler |
| 服务层 | `user_service.go` | 功能+service |
| 模型层 | `user_model.go` | 功能+model |
| 中间件 | `auth_middleware.go` | 功能+middleware |
| 配置文件 | `config.dev.yml` | 功能+环境 |

#### 测试文件命名
| 测试类型 | 命名示例 | 说明 |
|----------|----------|------|
| 单元测试 | `user_service_test.go` | 模块+test |
| 集成测试 | `user_integration_test.js` | 功能+integration+test |
| E2E测试 | `user_journey_e2e.test.js` | 场景+e2e+test |
| 测试数据 | `test_users.json` | test+数据类型 |
| 测试报告 | `report-2025-07-22.md` | report+日期 |

#### 文档文件命名
| 文档类型 | 命名示例 | 说明 |
|----------|----------|------|
| API文档 | `api-authentication.md` | api+功能 |
| 部署指南 | `deploy-production.md` | deploy+环境 |
| 用户手册 | `user-guide-registration.md` | user-guide+功能 |
| 架构文档 | `architecture-overview.md` | architecture+主题 |
| 故障排除 | `troubleshooting-npm-issues.md` | troubleshooting+问题 |

## 🔄 版本控制命名

### Git分支命名
- **功能分支**: `feature/user-authentication`
- **修复分支**: `fix/login-validation-error`
- **重构分支**: `refactor/user-service-structure`
- **发布分支**: `release/v2.1.0`
- **热修复**: `hotfix/security-patch`

### 标签命名
- **版本标签**: `v2.1.0`
- **里程碑标签**: `milestone-alpha-1`
- **环境标签**: `env-production`

## 📋 命名检查清单

### 创建新文件时的检查项
- [ ] 文件名是否符合对应类型的命名规范？
- [ ] 是否使用了正确的分隔符（连字符vs下划线）？
- [ ] 是否有清晰的语义描述？
- [ ] 是否避免了缩写和模糊词汇？
- [ ] 是否检查了文件名的唯一性？

### 目录命名检查项
- [ ] 目录名是否使用小写字母？
- [ ] 是否使用了连字符分隔？
- [ ] 是否具有描述性和可读性？
- [ ] 是否符合项目整体结构？

## 🚨 禁止的命名方式

### ❌ 不允许的命名
- 使用空格: `User Service.js`
- 使用特殊字符: `user@service.js`
- 混合大小写混乱: `User-service_Test.js`
- 缩写不明确: `usr-srv.js`
- 数字开头: `2user_service.js`
- 版本号在文件名中: `user_service_v2.js`

### ✅ 推荐的命名示例
```
# 代码文件
user-authentication.service.ts
letter-management.controller.go
courier-validation.middleware.py

# 配置文件
database.config.prod.yml
redis.config.yml

# 测试文件
user-registration.e2e.test.ts
courier-levels.integration.test.js

# 文档文件
api-courier-management.md
deploy-production-guide.md
```

## 📊 命名规范实施检查表

### Phase 1: 现有文件重命名
- [x] 根目录松散文件已整理到对应目录
- [x] 文档文件已按规范命名和分类
- [x] 测试文件已集中到test-kimi目录
- [x] 图片文件已移动到docs/images目录

### Phase 2: 项目文件检查
- [ ] 检查所有代码文件是否符合命名规范
- [ ] 验证配置文件路径和命名
- [ ] 确认测试文件命名标准
- [ ] 检查文档文件命名规范

### Phase 3: 自动化工具
- [ ] 创建命名规范检查脚本
- [ ] 添加Git hooks进行命名检查
- [ ] 建立CI/CD命名验证流程

## 🎯 执行建议

### 立即行动项
1. 所有新创建的文件必须遵循本规范
2. 现有文件的重命名可以逐步进行
3. 关键路径文件优先考虑重命名
4. 建立文件命名规范培训机制

### 长期维护
- 定期审查文件命名规范执行情况
- 收集团队反馈并优化规范
- 更新文档和培训材料
- 监控命名规范的遵守率