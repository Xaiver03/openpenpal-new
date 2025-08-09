# 📚 OpenPenPal 文档中心

<div align="center">
  
## 🎯 快速导航

| [🚀 快速开始](#-快速开始) | [📖 产品文档](#-产品文档) | [🏗️ 技术文档](#️-技术文档) | [🔧 开发指南](#-开发指南) | [📦 部署运维](#-部署运维) |
|:---:|:---:|:---:|:---:|:---:|

</div>

---

> 📅 **最后更新**: 2025-08-07  
> 🏷️ **版本**: v1.0  
> 📝 **状态**: 生产就绪

## 🚀 快速开始

### 新手必读
1. **[5分钟快速上手](docs/getting-started/5min-guide.md)** - 最快速度运行项目
2. **[测试账号说明](docs/getting-started/test-accounts.md)** - 所有测试账号和权限
3. **[项目介绍](README.md)** - 了解OpenPenPal是什么

### 一键启动
```bash
# 推荐：演示模式（包含测试数据）
./startup/quick-start.sh demo --auto-open

# 开发模式
./startup/quick-start.sh development

# 完整模式（所有微服务）
./startup/quick-start.sh complete
```

## 📖 产品文档

### 产品需求文档（PRD）
- **[主PRD - OpenPenPal产品需求文档](docs/product/openpenpal-product-requirements-document-v1.md)**
- **子系统PRD**：
  - [📝 写信系统](docs/product/sub-module-prd/写信系统%20PRD.md)
  - [🏃 信使系统](docs/product/sub-module-prd/信使系统%20PRD.md)
  - [🏛️ 博物馆系统](docs/product/sub-module-prd/信件博物馆子系统%20PRD.md)
  - [🤖 AI子系统](docs/product/sub-module-prd/AI子系统%20PRD（AI%20Assistant%20System）.md)
  - [📍 编码系统](docs/product/sub-module-prd/编码系统%20PRD（OP%20Code%20System）.md)
  - [📊 条码系统](docs/product/sub-module-prd/条码系统%20PRD.md)

### 功能规格说明书（FSD）

#### 基础设施FSD
- [👤 用户系统](docs/specifications/infrastructure-fsd/OpenPenPal%20用户系统%20FSD（User%20System%20Functional%20Specification%20Document）.md)
- [🔐 权限系统](docs/specifications/infrastructure-fsd/OpenPenPal%20权限系统%20FSD（Auth%20&%20Role%20System）.md)
- [📬 通知系统](docs/specifications/infrastructure-fsd/OpenPenPal%20通知系统%20FSD（Notification%20System）.md)
- [📈 数据统计系统](docs/specifications/infrastructure-fsd/OpenPenPal%20数据统计系统%20FSD（Data%20Analytics%20System）.md)
- [🛡️ 内容审核系统](docs/specifications/infrastructure-fsd/OpenPenPal%20内容审核系统%20FSD（Moderation%20System）.md)
- [💰 积分系统](docs/specifications/infrastructure-fsd/OpenPenPal%20积分与激励系统%20FSD（Credit%20&%20Incentive%20System）.md)

#### 子系统FSD
- [✍️ 写信系统](docs/specifications/subsystem-fsd/写信系统%20FSD（功能规格说明书）.md)
- [📮 信使系统](docs/specifications/subsystem-fsd/信使系统%20FSD（Courier%20System%20Functional%20Specification%20Document）.md)
- [✉️ 信封系统](docs/specifications/subsystem-fsd/信封系统%20FSD（Envelope%20System%20Functional%20Specification%20Document）.md)
- [🏛️ 博物馆系统](docs/specifications/subsystem-fsd/OpenPenPal%20信件博物馆子系统%20FSD（Letter%20Museum%20System）.md)
- [🤖 AI子系统](docs/specifications/subsystem-fsd/AI子系统%20FSD（AI%20Subsystem%20Functional%20Specification%20Document）.md)
- [📍 编码系统](docs/specifications/subsystem-fsd/编码系统%20FSD（OP%20Code%20System%20Functional%20Specification%20Document）.md)
- [📊 条码系统](docs/specifications/subsystem-fsd/条码系统%20FSD（Barcode%20System%20Functional%20Specification%20Document）.md)

## 🏗️ 技术文档

### 系统架构
- [📐 系统架构总览](docs/architecture/README.md)
- [🌟 SOTA改进总结](docs/architecture/SOTA_IMPROVEMENTS_SUMMARY.md)
- [📊 API覆盖率分析](docs/architecture/API-COVERAGE-ANALYSIS.md)
- [🚀 中间件优化](docs/architecture/MIDDLEWARE-OPTIMIZATION-SUMMARY.md)

### API文档
- [📡 统一API规范](docs/api/unified-specification.md)
- [🔌 API集成状态](docs/API_INTEGRATION_STATUS.md)

### 数据库
- [🐘 PostgreSQL快速配置](docs/deployment/POSTGRESQL_QUICKSTART.md)
- [🔄 GORM迁移指南](docs/deployment/GORM_POSTGRESQL_MIGRATION.md)

## 🔧 开发指南

### 开发规范
- [💻 编码标准](docs/development/coding-standards.md)
- [🧩 组件管理](docs/development/component-management.md)
- [📝 文件命名规范](docs/development/file-naming.md)
- [🤝 贡献指南](CONTRIBUTING.md)

### 开发工具
- [🤖 Claude AI助手指南](CLAUDE.md) - AI辅助开发
- [🛠️ 启动脚本指南](docs/guides/startup-scripts-guide.md)
- [👥 多Agent开发](docs/guides/multi-agent-guide.md)

### 测试
- [🧪 测试指南](docs/testing/README.md)
- [✅ 权限测试](startup/tests/test-permissions.sh)
- [📊 API测试](scripts/test-apis.sh)

## 📦 部署运维

### 部署指南
- [🚀 部署总览](docs/deployment/README.md)
- [🔧 启动模式说明](docs/deployment/STARTUP-MODES.md)
- [🐳 Docker部署](docs/deployment/JAVA-DOCKER-SETUP.md)
- [⚡ 生产环境部署](docs/deployment/PRODUCTION-SERVICES.md)

### 运维工具
- [📜 脚本使用指南](docs/operations/scripts-usage.md)
- [🔍 健康检查](startup/check-status.sh)
- [🛑 服务管理](startup/stop-all.sh)

## 📊 项目报告

### 最新报告
- [✅ 系统健康检查](docs/reports/SYSTEM_HEALTH_CHECK_REPORT.md)
- [🔒 安全审计报告](docs/reports/SECURITY_AUDIT_REPORT.md)
- [📈 项目分析报告](docs/reports/PROJECT_ANALYSIS_REPORT.md)
- [🏗️ 微服务状态](docs/reports/MICROSERVICES-STATUS-REPORT.md)

### 代码审查
- [📋 完整代码审查报告](code-review-reports/COMPLETE_CODE_REVIEW_REPORT.md)
- [🔐 安全审计报告](code-review-reports/SECURITY_AUDIT_REPORT_2025.md)
- [⚡ 性能分析报告](code-review-reports/PERFORMANCE_ANALYSIS_REPORT.md)

## 🗂️ 文档索引

### 按类型分类
- **产品文档**: [PRD](docs/product/) | [FSD](docs/specifications/)
- **技术文档**: [架构](docs/architecture/) | [API](docs/api/) | [数据库](docs/deployment/)
- **开发文档**: [规范](docs/development/) | [指南](docs/guides/) | [测试](docs/testing/)
- **运维文档**: [部署](docs/deployment/) | [脚本](docs/operations/)

### 按角色分类
- **产品经理**: [PRD](docs/product/) | [FSD](docs/specifications/)
- **开发人员**: [快速开始](#-快速开始) | [开发指南](#-开发指南) | [API文档](docs/api/)
- **运维人员**: [部署指南](#-部署运维) | [脚本工具](docs/operations/)
- **测试人员**: [测试文档](docs/testing/) | [测试账号](docs/getting-started/test-accounts.md)

## 🔍 快速查找

### 常用命令
```bash
# 查看服务状态
./startup/check-status.sh

# 查看日志
tail -f logs/*.log

# 运行测试
./scripts/test-apis.sh

# 停止所有服务
./startup/stop-all.sh
```

### 常见问题
1. **如何快速启动？** → [5分钟指南](docs/getting-started/5min-guide.md)
2. **测试账号是什么？** → [测试账号说明](docs/getting-started/test-accounts.md)
3. **如何部署到生产？** → [生产部署指南](docs/deployment/PRODUCTION-SERVICES.md)
4. **API文档在哪？** → [API规范](docs/api/unified-specification.md)

## 📝 文档维护

- **归档文档**: [临时文档](docs/archive/temp-docs/) | [历史文档](docs/archive/)
- **文档规范**: [文档指南](docs/project/documentation-guide.md)
- **更新日志**: [CHANGELOG](CHANGELOG.md)

---

<div align="center">

📧 **联系方式** | 🐛 **问题反馈** | 🤝 **贡献代码**

[返回顶部](#-openpenpal-文档中心)

</div>