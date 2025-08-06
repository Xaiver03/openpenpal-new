# OpenPenPal 文档中心

> **版本**: 2.1.0  
> **更新时间**: 2025-07-24  
> **维护者**: 文档优化系统

> 🔗 **主入口**: [项目统一入口文档](../UNIFIED_ENTRY.md) - 推荐从这里开始

## 📚 文档导航

### 🏗️ 架构与设计
- [多Agent协同开发框架](./architecture/multi-agent-coordination.md) - 多Agent开发模式和分工
- [统一API规范](./architecture/api-specification.md) - RESTful API设计规范
- [共享上下文配置](./architecture/shared-context.md) - 服务配置和环境信息
- [系统架构设计](./architecture/system-design.md) - 整体技术架构

### 🚀 快速开始
- [项目快速开始](./guides/quick-start.md) - 5分钟上手指南
- [开发环境搭建](./guides/dev-environment.md) - 完整环境配置
- [多Agent开发指南](./guides/multi-agent-guide.md) - Agent协同开发流程

### 📖 开发文档
- [开发规范](./development/coding-standards.md) - 代码规范和最佳实践
- [开发计划](./development/roadmap.md) - 功能规划和里程碑
- [WebSocket实时通信](./development/websocket.md) - 实时通信架构
- [认证与权限](./development/authentication.md) - JWT认证和RBAC权限

### 📋 Agent任务管理
- [Agent任务总览](../agent-tasks/README.md) - 所有Agent的任务分配
- [Agent #1 - 团队协调+前端](../agent-tasks/AGENT-1-TEAM-LEAD.md) - 技术架构师和前端负责人
- [Agent #2 - 写信服务](../agent-tasks/AGENT-2-WRITE-SERVICE.md) - Python FastAPI服务
- [Agent #3 - 信使服务](../agent-tasks/AGENT-3-COURIER-SERVICE.md) - Go Gin任务调度系统
- [Agent #4 - 管理后台](../agent-tasks/AGENT-4-ADMIN-SERVICE.md) *(待创建)*
- [Agent #5 - OCR服务](../agent-tasks/AGENT-5-OCR-SERVICE.md) *(待创建)*

### 🛠️ 运维指南
- [部署指南](./operations/deployment.md) - 生产环境部署
- [脚本使用说明](./operations/scripts-usage.md) - 各种脚本的用法
- [监控与日志](./operations/monitoring.md) - 系统监控配置

### 🐛 问题解决
- [常见问题FAQ](./troubleshooting/faq.md) - 常见问题汇总
- [NPM权限问题](./troubleshooting/npm-permissions.md) - NPM相关问题解决
- [调试技巧](./troubleshooting/debugging.md) - 开发调试指南

### 📊 项目管理
- [版本发布记录](./project/changelog.md) - 版本更新历史
- [贡献指南](./project/contributing.md) - 如何参与项目
- [文档维护规范](./project/documentation-guide.md) - 文档编写标准

## 🔍 快速查找

### 按角色查找
- **前端开发者**: [前端开发指南](./development/frontend-guide.md)
- **后端开发者**: [后端开发指南](./development/backend-guide.md)
- **运维人员**: [运维手册](./operations/ops-manual.md)
- **产品经理**: [产品文档](./product/product-docs.md)

### 按技术栈查找
- **Next.js**: [前端架构](./tech-stack/nextjs.md)
- **Go + Gin**: [Go服务开发](./tech-stack/golang.md)
- **Python + FastAPI**: [Python服务开发](./tech-stack/python.md)
- **PostgreSQL**: [数据库设计](./tech-stack/database.md)

## 📝 文档更新记录

| 日期 | 更新内容 | 维护者 |
|------|---------|--------|
| 2025-07-24 | 实施文档优化计划，创建统一入口 | 文档优化系统 |
| 2025-07-24 | 清理重复文档，建立归档体系 | 文档优化系统 |
| 2025-07-24 | 制定文档格式标准规范 | 文档优化系统 |
| 2025-07-20 | 创建统一文档系统 | Agent #1 |
| 2025-07-20 | 建立多Agent协同框架 | Agent #1 |

## 🔗 重要链接

- [GitHub仓库](https://github.com/openpenpal/openpenpal)
- [API在线文档](http://localhost:8080/swagger)
- [项目看板](https://github.com/openpenpal/openpenpal/projects)
- [问题追踪](https://github.com/openpenpal/openpenpal/issues)

## 💡 文档使用提示

1. **新Agent入职**: 请先阅读[多Agent开发指南](./guides/multi-agent-guide.md)
2. **查找API**: 使用[统一API规范](./architecture/api-specification.md)
3. **遇到问题**: 查看[常见问题FAQ](./troubleshooting/faq.md)
4. **提交代码**: 遵循[开发规范](./development/coding-standards.md)

---

## 📋 文档维护

**文档维护原则**:
- 🎯 保持单一信息源 (Single Source of Truth)
- ⏰ 及时更新，避免过时信息
- 📁 使用清晰的目录结构
- 🔗 提供充分的交叉引用
- 📝 遵循[文档格式标准](./project/DOCUMENTATION_STANDARDS.md)

**质量保证**:
- 定期检查链接有效性
- 及时归档过时文档
- 统一文档格式和命名规范
- 建立文档更新责任制

---

**最后更新**: 2025-07-24  
**维护**: 文档优化系统